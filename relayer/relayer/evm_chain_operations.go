package relayer

import (
	"context"
	"encoding/json"
	"fmt"

	cardanotx "github.com/Ethernal-Tech/apex-bridge/cardano"
	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/eth"
	ethtxhelper "github.com/Ethernal-Tech/apex-bridge/eth/txhelper"
	"github.com/Ethernal-Tech/apex-bridge/relayer/core"
	"github.com/Ethernal-Tech/bn256"
	"github.com/hashicorp/go-hclog"
)

var _ core.ChainOperations = (*CardanoChainOperations)(nil)

type EVMChainOperations struct {
	config           *cardanotx.RelayerEVMChainConfig
	evmSmartContract eth.IBridgeSmartContract // TODO: replace with correct smart contract interface
	chainID          string
	logger           hclog.Logger
}

func NewEVMChainOperations(
	jsonConfig json.RawMessage,
	chainID string,
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

	wallet, err := ethtxhelper.NewEthTxWalletFromSecretManager(secretsManager)
	if err != nil {
		return nil, fmt.Errorf("failed to load wallet for relayer: %w", err)
	}

	evmSmartContract, err := eth.NewBridgeSmartContractWithWallet(
		config.NodeURL, config.SmartContractAddr, wallet, config.DynamicTx, logger)
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
	ctx context.Context, bridgeSmartContract eth.IBridgeSmartContract, smartContractData *eth.ConfirmedBatch,
) (err error) {
	signatures := make(bn256.Signatures, len(smartContractData.Signatures))
	for i, bytes := range smartContractData.Signatures {
		signatures[i], err = bn256.UnmarshalSignature(bytes)
		if err != nil {
			return fmt.Errorf("invalid signature: %w", err)
		}
	}

	bitmap := common.NewBitmap(smartContractData.Bitmap)
	signature, _ := signatures.Aggregate().Marshal() // error is always nil

	fmt.Println(bitmap, signature)
	// a TODO: send actual tx to nexus/evm chain
	return nil
}
