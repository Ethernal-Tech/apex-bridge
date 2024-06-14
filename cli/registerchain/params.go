package cliregisterchain

import (
	"context"
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
	defaultGasLimit             = 5_242_880
	keysDirectoryFlag           = "keys-dir"
	bridgeValidatorDataDirFlag  = "bridge-validator-data-dir"
	bridgeValidatorConfigFlag   = "bridge-validator-config"
	ogmiosURLFlag               = "ogmios"
	blockfrostURLFlag           = "blockfrost"
	blockfrostProjectAPIKeyFlag = "blockfrost-api-key"
	socketPathFlag              = "socket-path"
	networkMagicFlag            = "network-magic"
	multisigAddrFlag            = "addr"
	multisigFeeAddrFlag         = "addr-fee"
	bridgeURLFlag               = "bridge-url"
	bridgeSCAddrFlag            = "bridge-addr"
	chainIDFlag                 = "chain"
	initialTokenSupplyFlag      = "token-supply"

	keysDirectoryFlagDesc           = "cardano wallet directory"
	bridgeValidatorDataDirFlagDesc  = "(mandatory if bridge-validator-config not specified) Path to bridge chain data directory when using local secrets manager" //nolint:lll
	bridgeValidatorConfigFlagDesc   = "(mandatory if bridge-validator-data not specified) Path to to bridge chain secrets manager config file"                    //nolint:lll
	chainIDFlagDesc                 = "chain ID (prime, vector, etc)"
	ogmiosURLFlagDesc               = "ogmios url"
	blockfrostURLFlagDesc           = "blockfrost url"
	blockfrostProjectAPIKeyFlagDesc = "blockfrost API key for prime network" //nolint:gosec
	socketPathFlagDesc              = "socket path for cardano node"
	networkMagicFlagDesc            = "network magic of a chain (default 0 and it means Mainnet)"
	multisigAddrFlagDesc            = "multisig address"
	multisigFeeAddrFlagDesc         = "fee payer address"
	bridgeURLFlagDesc               = "bridge node url"
	bridgeSCAddrFlagDesc            = "bridge smart contract address"
	initialTokenSupplyFlagDesc      = "initial token supply for the chain"
)

type registerChainParams struct {
	keysDirectory           string
	bridgeValidatorDataDir  string
	bridgeValidatorConfig   string
	blockfrostURL           string
	blockfrostProjectAPIKey string
	ogmiosURL               string
	socketPath              string
	networkMagic            uint
	multisigAddr            string
	multisigFeeAddr         string
	bridgeURL               string
	bridgeSCAddr            string
	chainID                 string
	initialTokenSupply      string

	ethTxHelper ethtxhelper.IEthTxHelper
}

func (ip *registerChainParams) validateFlags() error {
	if !common.IsValidURL(ip.bridgeURL) {
		return fmt.Errorf("invalid bridge node url: %s", ip.bridgeURL)
	}

	if ip.blockfrostURL == "" && ip.socketPath == "" && ip.ogmiosURL == "" {
		return errors.New("neither a blockfrost nor a ogmios nor a socket path is specified")
	}

	if ip.blockfrostURL != "" && !common.IsValidURL(ip.blockfrostURL) {
		return fmt.Errorf("invalid blockfrost url: %s", ip.blockfrostURL)
	}

	if ip.ogmiosURL != "" && !common.IsValidURL(ip.ogmiosURL) {
		return fmt.Errorf("invalid ogmios url: %s", ip.ogmiosURL)
	}

	if ip.keysDirectory == "" {
		return fmt.Errorf("invalid directory for Cardano keys: %s", ip.keysDirectory)
	}

	if ip.bridgeValidatorDataDir == "" && ip.bridgeValidatorConfig == "" {
		return fmt.Errorf("specify at least one of: %s, %s", bridgeValidatorDataDirFlag, bridgeValidatorConfigFlag)
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
		return fmt.Errorf("--%s flag not specified", chainIDFlag)
	}

	ethTxHelper, err := ethtxhelper.NewEThTxHelper(ethtxhelper.WithNodeURL(ip.bridgeURL))
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
		&ip.keysDirectory,
		keysDirectoryFlag,
		"",
		keysDirectoryFlagDesc,
	)

	cmd.Flags().StringVar(
		&ip.bridgeValidatorDataDir,
		bridgeValidatorDataDirFlag,
		"",
		bridgeValidatorDataDirFlagDesc,
	)

	cmd.Flags().StringVar(
		&ip.bridgeValidatorConfig,
		bridgeValidatorConfigFlag,
		"",
		bridgeValidatorConfigFlagDesc,
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
		&ip.ogmiosURL,
		ogmiosURLFlag,
		"",
		ogmiosURLFlagDesc,
	)

	cmd.Flags().StringVar(
		&ip.blockfrostURL,
		blockfrostURLFlag,
		"",
		blockfrostURLFlagDesc,
	)

	cmd.Flags().StringVar(
		&ip.blockfrostProjectAPIKey,
		blockfrostProjectAPIKeyFlag,
		"",
		blockfrostProjectAPIKeyFlagDesc,
	)

	cmd.Flags().StringVar(
		&ip.socketPath,
		socketPathFlag,
		"",
		socketPathFlagDesc,
	)

	cmd.Flags().UintVar(
		&ip.networkMagic,
		networkMagicFlag,
		0,
		networkMagicFlagDesc,
	)

	cmd.Flags().StringVar(
		&ip.initialTokenSupply,
		initialTokenSupplyFlag,
		"",
		initialTokenSupplyFlagDesc,
	)

	cmd.Flags().StringVar(
		&ip.bridgeURL,
		bridgeURLFlag,
		"",
		bridgeURLFlagDesc,
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

	cmd.MarkFlagsMutuallyExclusive(bridgeValidatorDataDirFlag, bridgeValidatorConfigFlag)
	cmd.MarkFlagsMutuallyExclusive(blockfrostURLFlag, socketPathFlag, ogmiosURLFlag)
}

func (ip *registerChainParams) Execute() (common.ICommandResult, error) {
	walletCardano, err := cardanotx.LoadWallet(path.Join(path.Clean(ip.keysDirectory), ip.chainID), false)
	if err != nil {
		return nil, fmt.Errorf("failed to load cardano wallet: %w", err)
	}

	secretsManager, err := common.GetSecretsManager(ip.bridgeValidatorDataDir, ip.bridgeValidatorConfig, true)
	if err != nil {
		return nil, fmt.Errorf("failed to create secrets manager: %w", err)
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

	convertUtxos := func(utxos []cardanowallet.Utxo) []contractbinding.IBridgeStructsUTXO {
		result := make([]contractbinding.IBridgeStructsUTXO, len(utxos))
		for i, x := range utxos {
			result[i] = contractbinding.IBridgeStructsUTXO{
				TxHash:  common.MustHashToBytes32(x.Hash),
				TxIndex: uint64(x.Index),
				Amount:  x.Amount,
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
		bind.TransactOpts{
			GasLimit: defaultGasLimit,
		},
		func(txOpts *bind.TransactOpts) (*types.Transaction, error) {
			return contract.RegisterChainGovernance(
				txOpts,
				contractbinding.IBridgeStructsChain{
					Id:              common.ToNumChainID(ip.chainID),
					AddressMultisig: ip.multisigAddr,
					AddressFeePayer: ip.multisigFeeAddr,
				},
				contractbinding.IBridgeStructsUTXOs{
					MultisigOwnedUTXOs: multisigUtxos,
					FeePayerOwnedUTXOs: feePayerUtxos,
				},
				initialTokenSupply,
				contractbinding.IBridgeStructsValidatorCardanoData{
					VerifyingKey:    [32]byte(walletCardano.MultiSig.GetVerificationKey()),
					VerifyingKeyFee: [32]byte(walletCardano.MultiSigFee.GetVerificationKey()),
				})
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
	if ip.ogmiosURL != "" {
		return cardanowallet.NewTxProviderOgmios(ip.ogmiosURL), nil
	} else if ip.socketPath != "" {
		return cardanowallet.NewTxProviderCli(ip.networkMagic, ip.socketPath)
	}

	return cardanowallet.NewTxProviderBlockFrost(ip.blockfrostURL, ip.blockfrostProjectAPIKey), nil
}
