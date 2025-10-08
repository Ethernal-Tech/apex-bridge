package relayer

import (
	"context"
	"encoding/json"
	"fmt"

	cardanotx "github.com/Ethernal-Tech/apex-bridge/cardano"
	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/eth"
	"github.com/Ethernal-Tech/apex-bridge/relayer/core"
	infracommon "github.com/Ethernal-Tech/cardano-infrastructure/common"
	cardanowallet "github.com/Ethernal-Tech/cardano-infrastructure/wallet"
	"github.com/hashicorp/go-hclog"
)

var _ core.ChainOperations = (*CardanoChainOperations)(nil)

type CardanoChainOperations struct {
	txProvider       cardanowallet.ITxProvider
	cardanoCliBinary string
	logger           hclog.Logger
}

func NewCardanoChainOperations(
	jsonConfig json.RawMessage,
	logger hclog.Logger,
) (*CardanoChainOperations, error) {
	config, err := cardanotx.NewCardanoChainConfig(jsonConfig)
	if err != nil {
		return nil, err
	}

	txProvider, err := config.CreateTxProvider()
	if err != nil {
		return nil, fmt.Errorf("failed to create tx provider: %w", err)
	}

	return &CardanoChainOperations{
		txProvider:       txProvider,
		cardanoCliBinary: cardanowallet.ResolveCardanoCliBinary(config.NetworkID),
		logger:           logger,
	}, nil
}

// SendTx implements core.ChainOperations.
func (cco *CardanoChainOperations) SendTx(
	ctx context.Context, _ eth.IBridgeSmartContract, smartContractData *eth.ConfirmedBatch,
) error {
	cco.logger.Debug("confirmed batch - sending tx", "batchID", smartContractData.ID, "binary", cco.cardanoCliBinary)

	signaturesLength := len(smartContractData.Signatures)
	feeSignaturesLength := len(smartContractData.FeeSignatures)
	stakeSignaturesLength := len(smartContractData.StakeSignatures)

	if signaturesLength != feeSignaturesLength || feeSignaturesLength != stakeSignaturesLength {
		return fmt.Errorf("wrong number of signatures: %d, %d, %d",
			signaturesLength, feeSignaturesLength, stakeSignaturesLength)
	}

	// Combine all signatures into a single witnesses slice
	witnesses := make([][]byte, 0, signaturesLength*3)

	for i := range signaturesLength {
		for _, sig := range [][]byte{
			smartContractData.Signatures[i], smartContractData.FeeSignatures[i], smartContractData.StakeSignatures[i],
		} {
			if len(sig) > 0 {
				witnesses = append(witnesses, sig)
			}
		}
	}

	txBuilder, err := cardanowallet.NewTxBuilder(cco.cardanoCliBinary)
	if err != nil {
		return err
	}

	defer txBuilder.Dispose()

	txSigned, err := txBuilder.AssembleTxWitnesses(smartContractData.RawTransaction, witnesses)
	if err != nil {
		return err
	}

	tip, err := infracommon.ExecuteWithRetry(ctx, func(ctx context.Context) (cardanowallet.QueryTipData, error) {
		return cco.txProvider.GetTip(ctx)
	})
	if err == nil {
		cco.logger.Info("confirmed batch - sending tx current tip",
			"block", tip.Block, "slot", tip.Slot, "hash", tip.Hash)
	}

	info, err := common.ParseTxInfo(txSigned, false)
	if err == nil {
		cco.logger.Info("confirmed batch - sending tx",
			"hash", info.Hash, "ttl", info.TTL, "fee", info.Fee, "metadata", info.MetaData)
	} else {
		cco.logger.Error("confirmed batch - sending tx info error", "err", err)
	}

	return cco.txProvider.SubmitTx(ctx, txSigned)
}
