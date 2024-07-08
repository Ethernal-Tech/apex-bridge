package relayer

import (
	"context"
	"encoding/json"
	"fmt"

	cardanotx "github.com/Ethernal-Tech/apex-bridge/cardano"
	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/eth"
	"github.com/Ethernal-Tech/apex-bridge/relayer/core"
	"github.com/hashicorp/go-hclog"
)

var _ core.ChainOperations = (*CardanoChainOperations)(nil)

type EVMChainOperations struct {
	config  *cardanotx.EVMChainConfig
	chainID string
	logger  hclog.Logger
}

func NewEVMChainOperations(
	jsonConfig json.RawMessage,
	chainID string,
	logger hclog.Logger,
) (*EVMChainOperations, error) {
	config, err := cardanotx.NewEVMChainConfig(jsonConfig)
	if err != nil {
		return nil, err
	}

	return &EVMChainOperations{
		config:  config,
		chainID: chainID,
		logger:  logger,
	}, nil
}

// SendTx implements core.ChainOperations.
func (cco *EVMChainOperations) SendTx(
	ctx context.Context, bridgeSmartContract eth.IBridgeSmartContract, smartContractData *eth.ConfirmedBatch,
) error {
	validatorsDatas, err := bridgeSmartContract.GetValidatorsChainData(ctx, cco.chainID)
	if err != nil {
		return nil
	}

	bitmap := common.NewBitmap(smartContractData.Bitmap)

	// TODO: aggregate bls public keys
	for i, x := range validatorsDatas {
		if bitmap.IsSet(uint64(i)) {
			fmt.Println(i, x.Key)
		}
	}

	// TODO: send actual tx to nexus/evm chain
	return nil
}
