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
	"github.com/Ethernal-Tech/apex-bridge/relayer/core"
	"github.com/Ethernal-Tech/bn256"
	"github.com/hashicorp/go-hclog"
)

var _ core.ChainOperations = (*CardanoChainOperations)(nil)

type EVMChainOperations struct {
	config           *cardanotx.RelayerEVMChainConfig
	evmSmartContract eth.IEVMGatewaySmartContract
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

	bigIntZero := new(big.Int).SetUint64(0)

	gasPrice := new(big.Int).SetUint64(config.GasPrice)
	if gasPrice.Cmp(bigIntZero) <= 0 {
		gasPrice = nil
	}

	gasFeeCap := new(big.Int).SetUint64(config.GasFeeCap)
	if gasFeeCap.Cmp(bigIntZero) <= 0 {
		gasFeeCap = nil
	}

	gasTipCap := new(big.Int).SetUint64(config.GasTipCap)
	if gasTipCap.Cmp(bigIntZero) <= 0 {
		gasTipCap = nil
	}

	if config.DynamicTx && gasPrice != nil {
		return nil, fmt.Errorf("gasPrice cannot be set while dynamicTx is true %w", err)
	} else if !config.DynamicTx && (gasTipCap != nil || gasFeeCap != nil) {
		return nil, fmt.Errorf("gasFeeCap and gasTipCap cannot be set while dynamicTx is false %w", err)
	}

	evmSmartContract, err := eth.NewEVMGatewaySmartContractWithWallet(
		config.NodeURL, gatewayAddress, wallet, config.DynamicTx, config.DepositGasLimit,
		gasPrice, gasFeeCap, gasTipCap, logger)
	if err != nil {
		return nil, err
	}

	return &EVMChainOperations{
		config:           config,
		chainID:          chainID,
		evmSmartContract: evmSmartContract,
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
