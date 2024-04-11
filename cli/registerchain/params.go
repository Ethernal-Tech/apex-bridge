package cliregisterchain

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"path"

	cardanotx "github.com/Ethernal-Tech/apex-bridge/cardano"
	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/contractbinding"
	ethtxhelper "github.com/Ethernal-Tech/apex-bridge/eth/txhelper"
	cardanowallet "github.com/Ethernal-Tech/cardano-infrastructure/wallet"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/spf13/cobra"
)

const (
	directoryFlag           = "dir"
	validatorDirectoryFlag  = "validator-dir"
	validatorPrivateKeyFlag = "validator-pk"
	blockfrostUrlFlag       = "block-frost"    // https://cardano-preview.blockfrost.io/api/v0
	blockfrostProjectIDFlag = "block-frost-id" // preview7mGSjpyEKb24OxQ4cCxomxZ5axMs5PvE
	multisigAddrFlag        = "addr"           // addr_test1wrs0nrc0rvrfl7pxjl8vgqp5xuvt8j4n8a2lu8gef80wxhq4lmleh
	multisigFeeAddrFlag     = "addr-fee"       // addr_test1vqjysa7p4mhu0l25qknwznvj0kghtr29ud7zp732ezwtzec0w8g3u
	bridgeUrlFlag           = "bridge-url"     // https://polygon-mumbai-pokt.nodies.app
	bridgeSCAddrFlag        = "bridge-addr"    // 0x69b6eEAff0A5c5F80a242104B79F4aC5c40E5130
	chainIDFlag             = "chain"
	initialTokenSupplyFlag  = "token-supply"

	directoryFlagDesc           = "cardano wallet directory"
	validatorDirectoryFlagDesc  = "validator ECDSA wallet directory"
	chainIDFlagDesc             = "chain ID (prime, vector, etc)"
	blockfrostUrlFlagDesc       = "block-frost url"
	blockfrostProjectIDFlagDesc = "block-frost project id"
	multisigAddrFlagDesc        = "multisig address"
	multisigFeeAddrFlagDesc     = "fee payer address"
	bridgeUrlFlagDesc           = "bridge node url"
	bridgeSCAddrFlagDesc        = "bridge smart contract address"
	validatorPrivateKeyFlagDesc = "validator's private key for the bridge"
	initialTokenSupplyFlagDesc  = "initial token supply for the chain"
)

type registerChainParams struct {
	directory           string
	validatorDir        string
	validatorPrivateKey string
	blockfrostUrl       string
	blockfrostProjectID string
	multisigAddr        string
	multisigFeeAddr     string
	bridgeUrl           string
	bridgeSCAddr        string
	chainID             string
	initialTokenSupply  string

	ethTxHelper ethtxhelper.IEthTxHelper
}

func (ip *registerChainParams) validateFlags() error {
	if !common.IsValidURL(ip.bridgeUrl) {
		return fmt.Errorf("invalid bridge node url: %s", ip.bridgeUrl)
	}

	if !common.IsValidURL(ip.blockfrostUrl) {
		return fmt.Errorf("invalid block-frost url: %s", ip.blockfrostUrl)
	}

	if ip.blockfrostProjectID == "" {
		return errors.New("block-frost project ID not specified")
	}

	if ip.directory == "" {
		return fmt.Errorf("invalid directory: %s", ip.directory)
	}

	if ip.validatorDir == "" && ip.validatorPrivateKey == "" {
		return fmt.Errorf("neither a validator directory nor a private key is specified")
	}

	if ip.multisigAddr == "" {
		return fmt.Errorf("multisig address not specified")
	}

	if ip.multisigFeeAddr == "" {
		return fmt.Errorf("fee payer address not specified")
	}

	addrDecoded, err := common.DecodeHex(ip.bridgeSCAddr)
	if err != nil || len(addrDecoded) == 0 || len(addrDecoded) > 20 {
		return fmt.Errorf("invalid bridge smart contract address: %s", ip.bridgeSCAddr)
	}

	if ip.chainID == "" {
		return errors.New("--chain flag not specified")
	}

	ethTxHelper, err := ethtxhelper.NewEThTxHelper(ethtxhelper.WithNodeUrl(ip.bridgeUrl))
	if err != nil {
		return fmt.Errorf("failed to connect to the bridge node: %w", err)
	}

	if _, ok := new(big.Int).SetString(ip.initialTokenSupply, 0); !ok {
		return errors.New("invalid initial token supply")
	}

	ip.ethTxHelper = ethTxHelper

	return nil
}

func (ip *registerChainParams) setFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(
		&ip.directory,
		directoryFlag,
		"",
		directoryFlagDesc,
	)

	cmd.Flags().StringVar(
		&ip.validatorDir,
		validatorDirectoryFlag,
		"",
		validatorDirectoryFlagDesc,
	)

	cmd.Flags().StringVar(
		&ip.validatorPrivateKey,
		validatorPrivateKeyFlag,
		"",
		validatorPrivateKeyFlagDesc,
	)

	cmd.Flags().StringVar(
		&ip.multisigAddr,
		multisigAddrFlag,
		"",
		multisigAddrFlagDesc,
	)

	cmd.Flags().StringVar(
		&ip.multisigFeeAddr,
		multisigFeeAddrFlag,
		"",
		multisigFeeAddrFlagDesc,
	)

	cmd.Flags().StringVar(
		&ip.blockfrostUrl,
		blockfrostUrlFlag,
		"",
		blockfrostUrlFlagDesc,
	)

	cmd.Flags().StringVar(
		&ip.blockfrostProjectID,
		blockfrostProjectIDFlag,
		"",
		blockfrostProjectIDFlagDesc,
	)

	cmd.Flags().StringVar(
		&ip.initialTokenSupply,
		initialTokenSupplyFlag,
		"",
		initialTokenSupplyFlagDesc,
	)

	cmd.Flags().StringVar(
		&ip.bridgeUrl,
		bridgeUrlFlag,
		"",
		bridgeUrlFlagDesc,
	)

	cmd.Flags().StringVar(
		&ip.bridgeSCAddr,
		bridgeSCAddrFlag,
		"",
		bridgeSCAddrFlagDesc,
	)

	cmd.Flags().StringVar(
		&ip.chainID,
		chainIDFlag,
		"",
		chainIDFlagDesc,
	)

	cmd.MarkFlagsMutuallyExclusive(validatorDirectoryFlag, validatorPrivateKeyFlag)
}

func (ip *registerChainParams) Execute() (common.ICommandResult, error) {
	wallet, err := cardanotx.LoadWallet(path.Join(path.Clean(ip.directory), ip.chainID), false)
	if err != nil {
		return nil, err
	}

	var walletEth *ethtxhelper.EthTxWallet

	if ip.validatorPrivateKey != "" {
		walletEth, err = ethtxhelper.NewEthTxWallet(ip.validatorPrivateKey)
	} else {
		walletEth, err = ethtxhelper.NewEthTxWalletFromBladeFile(path.Clean(ip.validatorDir))
	}

	if err != nil {
		return nil, fmt.Errorf("invalid validator ECDSA key: %w", err)
	}

	contract, err := contractbinding.NewBridgeContract(common.HexToAddress(ip.bridgeSCAddr), ip.ethTxHelper.GetClient())
	if err != nil {
		return nil, fmt.Errorf("failed to connect to bridge smart contract: %w", err)
	}

	initialTokenSupply, _ := new(big.Int).SetString(ip.initialTokenSupply, 0)

	blockfrost, err := cardanowallet.NewTxProviderBlockFrost(ip.blockfrostUrl, ip.blockfrostProjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to block frost: %w", err)
	}

	utxos, err := blockfrost.GetUtxos(context.Background(), ip.multisigAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve utxos for multisig address: %w", err)
	}

	utxosFee, err := blockfrost.GetUtxos(context.Background(), ip.multisigFeeAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve utxos for fee payer address: %w", err)
	}

	convertUtxos := func(utxos []cardanowallet.Utxo) []contractbinding.IBridgeContractStructsUTXO {
		result := make([]contractbinding.IBridgeContractStructsUTXO, len(utxos))
		for i, x := range utxos {
			result[i] = contractbinding.IBridgeContractStructsUTXO{
				TxHash:  x.Hash,
				TxIndex: new(big.Int).SetUint64(uint64(x.Index)),
				Amount:  new(big.Int).SetUint64(x.Amount),
			}
		}

		return result
	}

	multisigUtxos := convertUtxos(utxos)
	feePayerUtxos := convertUtxos(utxosFee)

	// create and send register chain tx
	tx, err := ip.ethTxHelper.SendTx(
		context.Background(),
		walletEth,
		bind.TransactOpts{},
		true, func(txOpts *bind.TransactOpts) (*types.Transaction, error) {
			return contract.RegisterChainGovernance(
				txOpts,
				ip.chainID,
				contractbinding.IBridgeContractStructsUTXOs{
					MultisigOwnedUTXOs: multisigUtxos,
					FeePayerOwnedUTXOs: feePayerUtxos,
				},
				ip.multisigAddr, ip.multisigFeeAddr,
				contractbinding.IBridgeContractStructsValidatorCardanoData{
					VerifyingKey:    hex.EncodeToString(wallet.MultiSig.GetVerificationKey()),
					VerifyingKeyFee: hex.EncodeToString(wallet.MultiSigFee.GetVerificationKey()),
				},
				initialTokenSupply)
		})
	if err != nil {
		return nil, err
	}

	receipt, err := ip.ethTxHelper.WaitForReceipt(context.Background(), tx.Hash().String(), true)
	if err != nil {
		return nil, err
	}

	return &CmdResult{
		chainID:   ip.chainID,
		blockHash: receipt.BlockHash.String(),
	}, nil
}
