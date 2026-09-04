package relayer

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"

	cardanotx "github.com/Ethernal-Tech/apex-bridge/cardano"
	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/eth"
	ethtxhelper "github.com/Ethernal-Tech/apex-bridge/eth/txhelper"
	"github.com/Ethernal-Tech/apex-bridge/relayer/core"
	"github.com/Ethernal-Tech/bn256"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/hashicorp/go-hclog"
)

var (
	_ core.ChainOperations = (*CardanoChainOperations)(nil)
	_ core.BalanceReporter = (*EVMChainOperations)(nil)
)

type EVMChainOperations struct {
	config           *cardanotx.RelayerEVMChainConfig
	evmSmartContract eth.IEVMGatewaySmartContract
	txHelper         *eth.EthHelperWrapper
	walletAddr       ethcommon.Address
	chainID          string
	logger           hclog.Logger
}

func NewEVMChainOperations(
	jsonConfig json.RawMessage,
	chainID string,
	gatewayAddress string,
	logger hclog.Logger,
) (*EVMChainOperations, error) {
	config, err := cardanotx.NewRelayerEVMChainConfig(jsonConfig)
	if err != nil {
		return nil, err
	}

	secretsManager, err := common.GetSecretsManager(
		config.DataDir, config.ConfigPath, true)
	if err != nil {
		return nil, fmt.Errorf("failed to create secrets manager: %w", err)
	}

	wallet, err := eth.GetRelayerEVMPrivateKey(secretsManager, chainID)
	if err != nil {
		return nil, fmt.Errorf("failed to load wallet for relayer: %w", err)
	}

	var gasPrice, gasFeeCap, gasTipCap *big.Int

	if config.GasPrice > 0 {
		gasPrice = new(big.Int).SetUint64(config.GasPrice)
	}

	if config.GasFeeCap > 0 {
		gasFeeCap = new(big.Int).SetUint64(config.GasFeeCap)
	}

	if config.GasTipCap > 0 {
		gasTipCap = new(big.Int).SetUint64(config.GasTipCap)
	}

	txHelper := eth.NewEthHelperWrapperWithWallet(wallet, logger.Named("tx_helper_wrapper"),
		ethtxhelper.WithNodeURL(config.NodeURL),
		ethtxhelper.WithInitClientAndChainIDFn(context.Background()),
		ethtxhelper.WithDynamicTx(config.DynamicTx),
		ethtxhelper.WithGasFeeMultiplier(config.GasFeeMultiplier),
		ethtxhelper.WithLogger(logger.Named("tx_helper")))

	evmSmartContract, err := eth.NewEVMGatewaySmartContract(
		gatewayAddress, txHelper, config.DepositGasLimit,
		gasPrice, gasFeeCap, gasTipCap, logger)
	if err != nil {
		return nil, err
	}

	return &EVMChainOperations{
		config:           config,
		chainID:          chainID,
		evmSmartContract: evmSmartContract,
		txHelper:         txHelper,
		walletAddr:       wallet.GetAddress(),
		logger:           logger,
	}, nil
}

// SendTx implements core.ChainOperations.
func (cco *EVMChainOperations) SendTx(
	ctx context.Context, _ eth.IBridgeSmartContract, smartContractData *eth.ConfirmedBatch,
) (err error) {
	signatures := make(bn256.Signatures, len(smartContractData.Signatures))
	for i, bytes := range smartContractData.Signatures {
		signatures[i], err = bn256.UnmarshalSignature(bytes)
		if err != nil {
			return fmt.Errorf("invalid signature: %w", err)
		}
	}

	signature, _ := signatures.Aggregate().Marshal() // error is always nil

	cco.logger.Info("Submitting deposit transaction",
		"signature", hex.EncodeToString(signature),
		"bitmap", smartContractData.Bitmap,
		"rawTx", hex.EncodeToString(smartContractData.RawTransaction))

	return cco.evmSmartContract.Deposit(ctx, signature, smartContractData.Bitmap, smartContractData.RawTransaction)
}

// GetRelayerAddress implements core.BalanceReporter.
func (cco *EVMChainOperations) GetRelayerAddress() string {
	return cco.walletAddr.String()
}

// GetRelayerBalance implements core.BalanceReporter.
// It returns the balance in wei.
func (cco *EVMChainOperations) GetRelayerBalance(ctx context.Context) (*big.Int, error) {
	ethTxHelper, err := cco.txHelper.GetEthHelper()
	if err != nil {
		return nil, fmt.Errorf("error while GetEthHelper: %w", err)
	}

	balance, err := ethTxHelper.GetClient().BalanceAt(ctx, cco.walletAddr, nil)
	if err != nil {
		return nil, cco.txHelper.ProcessError(err)
	}

	return balance, nil
}
