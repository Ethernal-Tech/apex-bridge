package cliregisterchain

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"strings"

	cardanotx "github.com/Ethernal-Tech/apex-bridge/cardano"
	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/contractbinding"
	"github.com/Ethernal-Tech/apex-bridge/eth"
	ethtxhelper "github.com/Ethernal-Tech/apex-bridge/eth/txhelper"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/spf13/cobra"
)

const (
	defaultGasLimit        = 5_242_880
	validatorDataDirFlag   = "validator-data-dir"
	validatorConfigFlag    = "validator-config"
	multisigAddrFlag       = "addr"
	multisigFeeAddrFlag    = "addr-fee"
	bridgeURLFlag          = "bridge-url"
	bridgeSCAddrFlag       = "bridge-addr"
	chainIDFlag            = "chain"
	chainTypeFlag          = "type"
	initialTokenSupplyFlag = "token-supply"

	validatorDataDirFlagDesc   = "(mandatory validator-config not specified) Path to bridge chain data directory when using local secrets manager" //nolint:lll
	validatorConfigFlagDesc    = "(mandatory validator-data not specified) Path to to bridge chain secrets manager config file"                    //nolint:lll
	chainIDFlagDesc            = "chain ID (prime, vector, etc)"
	chainTypeFlagDesc          = "chain type (0 is Cardano, 1 is EVM, etc)"
	socketPathFlagDesc         = "socket path for cardano node"
	multisigAddrFlagDesc       = "multisig address"
	multisigFeeAddrFlagDesc    = "fee payer address"
	bridgeURLFlagDesc          = "bridge node url"
	bridgeSCAddrFlagDesc       = "bridge smart contract address"
	initialTokenSupplyFlagDesc = "initial token supply for the chain"
)

type registerChainParams struct {
	validatorDataDir   string
	validatorConfig    string
	multisigAddr       string
	multisigFeeAddr    string
	bridgeURL          string
	bridgeSCAddr       string
	chainID            string
	chainType          uint8
	initialTokenSupply string

	ethTxHelper ethtxhelper.IEthTxHelper
}

func (ip *registerChainParams) validateFlags() error {
	if ip.chainID == "" {
		return fmt.Errorf("--%s flag not specified", chainIDFlag)
	}

	ip.chainID = strings.ToLower(ip.chainID)

	// for known chain IDs, chainType is already known
	switch ip.chainID {
	case common.ChainIDStrPrime, common.ChainIDStrVector:
		ip.chainType = common.ChainTypeCardano
	case common.ChainIDStrNexus:
		ip.chainType = common.ChainTypeEVM
	}

	if !common.IsValidHTTPURL(ip.bridgeURL) {
		return fmt.Errorf("invalid bridge node url: %s", ip.bridgeURL)
	}

	if ip.validatorDataDir == "" && ip.validatorConfig == "" {
		return fmt.Errorf("specify at least one of: %s, %s", validatorDataDirFlag, validatorConfigFlag)
	}

	if ip.multisigAddr == "" {
		return fmt.Errorf("multisig address not specified")
	}

	if ip.multisigFeeAddr == "" && ip.chainType == common.ChainTypeCardano {
		return fmt.Errorf("fee payer address not specified")
	}

	addrDecoded, err := common.DecodeHex(ip.bridgeSCAddr)
	if err != nil || len(addrDecoded) == 0 || len(addrDecoded) > 20 {
		return fmt.Errorf("invalid bridge smart contract address: %s", ip.bridgeSCAddr)
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
		&ip.validatorDataDir,
		validatorDataDirFlag,
		"",
		validatorDataDirFlagDesc,
	)

	cmd.Flags().StringVar(
		&ip.validatorConfig,
		validatorConfigFlag,
		"",
		validatorConfigFlagDesc,
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

	cmd.Flags().Uint8Var(
		&ip.chainType,
		chainTypeFlag,
		0,
		chainTypeFlagDesc,
	)

	cmd.MarkFlagsMutuallyExclusive(validatorDataDirFlag, validatorConfigFlag)
}

func (ip *registerChainParams) Execute(outputter common.OutputFormatter) (common.ICommandResult, error) {
	var validatorChainData eth.ValidatorChainData

	secretsManager, err := common.GetSecretsManager(ip.validatorDataDir, ip.validatorConfig, true)
	if err != nil {
		return nil, fmt.Errorf("failed to create secrets manager: %w", err)
	}

	switch ip.chainType {
	case common.ChainTypeCardano:
		walletCardano, err := cardanotx.LoadWallet(secretsManager, ip.chainID)
		if err != nil {
			return nil, fmt.Errorf("failed to load cardano wallet: %w", err)
		}

		validatorChainData.Key[0] = new(big.Int).SetBytes(walletCardano.MultiSig.VerificationKey)
		validatorChainData.Key[1] = new(big.Int).SetBytes(walletCardano.MultiSigFee.VerificationKey)
		validatorChainData.Key[2] = new(big.Int)
		validatorChainData.Key[3] = new(big.Int)
	case common.ChainTypeEVM:
		privateKey, err := eth.GetBatcherEVMPrivateKey(secretsManager, ip.chainID)
		if err != nil {
			return nil, fmt.Errorf("failed to load eth wallet: %w", err)
		}

		bigInts := privateKey.PublicKey().ToBigInt()
		validatorChainData.Key[0] = bigInts[0]
		validatorChainData.Key[1] = bigInts[1]
		validatorChainData.Key[2] = bigInts[2]
		validatorChainData.Key[3] = bigInts[3]
	default:
		return nil, fmt.Errorf("chain type does not exist: %d", ip.chainType)
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
				eth.Chain{
					Id:              common.ToNumChainID(ip.chainID),
					ChainType:       ip.chainType,
					AddressMultisig: ip.multisigAddr,
					AddressFeePayer: ip.multisigFeeAddr,
				},
				initialTokenSupply,
				validatorChainData)
		})
	if err != nil {
		return nil, err
	}

	_, _ = outputter.Write([]byte(fmt.Sprintf("transaction has been submitted: %s", tx.Hash())))
	outputter.WriteOutput()

	receipt, err := ip.ethTxHelper.WaitForReceipt(context.Background(), tx.Hash().String(), true)
	if err != nil {
		return nil, err
	}

	return &CmdResult{
		chainID:   ip.chainID,
		blockHash: receipt.BlockHash.String(),
	}, nil
}
