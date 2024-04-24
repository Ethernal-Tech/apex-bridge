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
	directoryFlag              = "dir"
	validatorAccountDirFlag    = "validator-dir"
	validatorAccountConfigFlag = "validator-config"
	blockfrostUrlFlag          = "block-frost"
	blockfrostProjectIDFlag    = "block-frost-id"
	socketPathFlag             = "socket-path"
	testnetMagicFlag           = "testnet"
	multisigAddrFlag           = "addr"
	multisigFeeAddrFlag        = "addr-fee"
	bridgeUrlFlag              = "bridge-url"
	bridgeSCAddrFlag           = "bridge-addr"
	chainIDFlag                = "chain"
	initialTokenSupplyFlag     = "token-supply"

	directoryFlagDesc              = "cardano wallet directory"
	validatorAccountDirFlagDesc    = "the directory for the Blade data if the local FS is used"
	validatorAccountConfigFlagDesc = "the path to the SecretsManager config file, if omitted, the local FS secrets manager is used"
	chainIDFlagDesc                = "chain ID (prime, vector, etc)"
	blockfrostUrlFlagDesc          = "block-frost url"
	blockfrostProjectIDFlagDesc    = "block-frost project id"
	socketPathFlagDesc             = "socket path for cardano node"
	testnetMagicFlagDesc           = "testnet magic number. leave 0 for mainnet"
	multisigAddrFlagDesc           = "multisig address"
	multisigFeeAddrFlagDesc        = "fee payer address"
	bridgeUrlFlagDesc              = "bridge node url"
	bridgeSCAddrFlagDesc           = "bridge smart contract address"
	initialTokenSupplyFlagDesc     = "initial token supply for the chain"
)

type registerChainParams struct {
	directory           string
	validatorDir        string
	validatorConfigDir  string
	blockfrostUrl       string
	blockfrostProjectID string
	socketPath          string
	testnetMagic        uint
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

	if ip.blockfrostUrl == "" && ip.socketPath == "" {
		return errors.New("neither a block frost nor a socket path is specified")
	}

	if ip.blockfrostUrl != "" && !common.IsValidURL(ip.blockfrostUrl) {
		return fmt.Errorf("invalid block-frost url: %s", ip.blockfrostUrl)
	}

	if ip.directory == "" {
		return fmt.Errorf("invalid directory: %s", ip.directory)
	}

	if ip.validatorDir == "" && ip.validatorConfigDir == "" {
		return fmt.Errorf("no config file or data directory passed in for validator secrets")
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
		validatorAccountDirFlag,
		"",
		validatorAccountDirFlagDesc,
	)

	cmd.Flags().StringVar(
		&ip.validatorConfigDir,
		validatorAccountConfigFlag,
		"",
		validatorAccountConfigFlagDesc,
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
		&ip.socketPath,
		socketPathFlag,
		"",
		socketPathFlagDesc,
	)

	cmd.Flags().UintVar(
		&ip.testnetMagic,
		testnetMagicFlag,
		0,
		testnetMagicFlagDesc,
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

	cmd.MarkFlagsMutuallyExclusive(validatorAccountDirFlag, validatorAccountConfigFlag)
	cmd.MarkFlagsMutuallyExclusive(blockfrostUrlFlag, socketPathFlag)
}

func (ip *registerChainParams) Execute() (common.ICommandResult, error) {
	walletCardano, err := cardanotx.LoadWallet(path.Join(path.Clean(ip.directory), ip.chainID), false)
	if err != nil {
		return nil, fmt.Errorf("failed to load cardano wallet: %w", err)
	}

	secretsManager, err := common.GetSecretsManager(ip.validatorDir, ip.validatorConfigDir, true)
	if err != nil {
		return nil, fmt.Errorf("failed to load blade wallet: %w", err)
	}

	walletEth, err := ethtxhelper.NewEthTxWalletFromSecretManager(secretsManager)
	if err != nil {
		return nil, fmt.Errorf("failed to load blade wallet: %w", err)
	}

	contract, err := contractbinding.NewBridgeContract(common.HexToAddress(ip.bridgeSCAddr), ip.ethTxHelper.GetClient())
	if err != nil {
		return nil, fmt.Errorf("failed to connect to bridge smart contract: %w", err)
	}

	initialTokenSupply, _ := new(big.Int).SetString(ip.initialTokenSupply, 0)

	cardanoProvider, err := ip.getCardanoProvider()
	if err != nil {
		return nil, fmt.Errorf("failed to create cardano tx provider: %w", err)
	}

	utxos, err := cardanoProvider.GetUtxos(context.Background(), ip.multisigAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve utxos for multisig address: %w", err)
	}

	utxosFee, err := cardanoProvider.GetUtxos(context.Background(), ip.multisigFeeAddr)
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
					VerifyingKey:    hex.EncodeToString(walletCardano.MultiSig.GetVerificationKey()),
					VerifyingKeyFee: hex.EncodeToString(walletCardano.MultiSigFee.GetVerificationKey()),
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

func (ip *registerChainParams) getCardanoProvider() (cardanowallet.IUTxORetriever, error) {
	if ip.socketPath != "" {
		return cardanowallet.NewTxProviderCli(ip.testnetMagic, ip.socketPath)
	}

	return cardanowallet.NewTxProviderBlockFrost(ip.blockfrostUrl, ip.blockfrostProjectID)
}
