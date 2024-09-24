package relayer

import (
	"context"
	"encoding/json"
	"fmt"

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

	evmSmartContract, err := eth.NewEVMGatewaySmartContractWithWallet(
		config.NodeURL, gatewayAddress, wallet, config.DynamicTx, logger)
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

	return cco.evmSmartContract.Deposit(ctx, signature, smartContractData.Bitmap, smartContractData.RawTransaction)
}
