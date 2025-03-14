package batcher

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/Ethernal-Tech/apex-bridge/batcher/core"
	cardano "github.com/Ethernal-Tech/apex-bridge/cardano"
	"github.com/Ethernal-Tech/apex-bridge/eth"
	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
	"github.com/Ethernal-Tech/cardano-infrastructure/secrets"
	cardanowallet "github.com/Ethernal-Tech/cardano-infrastructure/wallet"
	"github.com/hashicorp/go-hclog"
)

var (
	_ core.ChainOperations            = (*CardanoChainOperations)(nil)
	_ ICardanoChainOperationsStrategy = (*CardanoChainOperationReactorStrategy)(nil)

	errTxSizeTooBig = errors.New("batch tx size too big")
)

// Get real tx size from protocolParams/config
const (
	maxTxSize = 16000
)

type CardanoChainOperations struct {
	config           *cardano.CardanoChainConfig
	wallet           *cardano.CardanoWallet
	txProvider       cardanowallet.ITxDataRetriever
	db               indexer.Database
	gasLimiter       eth.GasLimitHolder
	cardanoCliBinary string
	strategy         ICardanoChainOperationsStrategy
	logger           hclog.Logger
}

func NewCardanoChainOperations(
	jsonConfig json.RawMessage,
	db indexer.Database,
	secretsManager secrets.SecretsManager,
	chainID string,
	strategy ICardanoChainOperationsStrategy,
	logger hclog.Logger,
) (*CardanoChainOperations, error) {
	cardanoConfig, err := cardano.NewCardanoChainConfig(jsonConfig)
	if err != nil {
		return nil, err
	}

	txProvider, err := cardanoConfig.CreateTxProvider()
	if err != nil {
		return nil, fmt.Errorf("failed to create tx provider: %w", err)
	}

	cardanoWallet, err := cardano.LoadWallet(secretsManager, chainID)
	if err != nil {
		return nil, err
	}

	return &CardanoChainOperations{
		wallet:           cardanoWallet,
		config:           cardanoConfig,
		txProvider:       txProvider,
		cardanoCliBinary: cardanowallet.ResolveCardanoCliBinary(cardanoConfig.NetworkID),
		gasLimiter:       eth.NewGasLimitHolder(submitBatchMinGasLimit, submitBatchMaxGasLimit, submitBatchStepsGasLimit),
		db:               db,
		strategy:         strategy,
		logger:           logger,
	}, nil
}

// GenerateBatchTransaction implements core.ChainOperations.
func (cco *CardanoChainOperations) GenerateBatchTransaction(
	ctx context.Context,
	bridgeSmartContract eth.IBridgeSmartContract,
	chainID string,
	confirmedTransactions []eth.ConfirmedTransaction,
	batchNonceID uint64,
) (*core.GeneratedBatchTxData, error) {
	txData, err := cco.generateBatchTransaction(
		ctx, bridgeSmartContract, chainID, confirmedTransactions, batchNonceID)

	if cco.shouldConsolidate(err) {
		cco.logger.Warn("consolidation batch generation started", "err", err)

		txData, err = cco.generateConsolidationTransaction(ctx, bridgeSmartContract, chainID, batchNonceID)
		if err != nil {
			err = fmt.Errorf("consolidation batch failed: %w", err)
		}
	}

	return txData, err
}

// SignBatchTransaction implements core.ChainOperations.
func (cco *CardanoChainOperations) SignBatchTransaction(
	generatedBatchData *core.GeneratedBatchTxData) ([]byte, []byte, error) {
	txBuilder, err := cardanowallet.NewTxBuilder(cco.cardanoCliBinary)
	if err != nil {
		return nil, nil, err
	}

	defer txBuilder.Dispose()

	witnessMultiSig, err := txBuilder.CreateTxWitness(generatedBatchData.TxRaw, cco.wallet.MultiSig)
	if err != nil {
		return nil, nil, err
	}

	witnessMultiSigFee, err := txBuilder.CreateTxWitness(generatedBatchData.TxRaw, cco.wallet.MultiSigFee)
	if err != nil {
		return nil, nil, err
	}

	return witnessMultiSig, witnessMultiSigFee, nil
}

// IsSynchronized implements core.IsSynchronized.
func (cco *CardanoChainOperations) IsSynchronized(
	ctx context.Context, bridgeSmartContract eth.IBridgeSmartContract, chainID string,
) (bool, error) {
	lastObservedBlockBridge, err := bridgeSmartContract.GetLastObservedBlock(ctx, chainID)
	if err != nil {
		return false, err
	}

	lastOracleBlockPoint, err := cco.db.GetLatestBlockPoint()
	if err != nil {
		return false, err
	}

	return lastOracleBlockPoint != nil &&
		lastOracleBlockPoint.BlockSlot >= lastObservedBlockBridge.BlockSlot.Uint64(), nil
}

// Submit implements core.Submit.
func (cco *CardanoChainOperations) Submit(
	ctx context.Context, bridgeSmartContract eth.IBridgeSmartContract, batch eth.SignedBatch,
) error {
	err := bridgeSmartContract.SubmitSignedBatch(ctx, batch, cco.gasLimiter.GetGasLimit())

	cco.gasLimiter.Update(err)

	return err
}

func (cco *CardanoChainOperations) generateBatchTransaction(
	ctx context.Context,
	bridgeSmartContract eth.IBridgeSmartContract,
	chainID string,
	confirmedTransactions []eth.ConfirmedTransaction,
	batchNonceID uint64,
) (*core.GeneratedBatchTxData, error) {
	validatorsData, err := cco.getCardanoData(ctx, bridgeSmartContract, chainID)
	if err != nil {
		return nil, err
	}

	metadata, err := cardano.CreateBatchMetaData(batchNonceID)
	if err != nil {
		return nil, err
	}

	protocolParams, err := cco.txProvider.GetProtocolParameters(ctx)
	if err != nil {
		return nil, err
	}

	multisigPolicyScript, multisigFeePolicyScript, err := cardano.GetPolicyScripts(validatorsData)
	if err != nil {
		return nil, err
	}

	multisigAddress, multisigFeeAddress, err := cardano.GetMultisigAddresses(
		cco.cardanoCliBinary, uint(cco.config.NetworkMagic), multisigPolicyScript, multisigFeePolicyScript)
	if err != nil {
		return nil, err
	}

	txOutputs, err := cco.strategy.GetOutputs(confirmedTransactions, cco.config, cco.logger)
	if err != nil {
		return nil, err
	}

	multisigUtxos, feeUtxos, err := cco.getUTXOsForNormalBatch(
		multisigAddress, multisigFeeAddress, protocolParams, txOutputs)
	if err != nil {
		return nil, err
	}

	slotNumber, err := cco.getSlotNumber()
	if err != nil {
		return nil, err
	}

	cco.logger.Info("Creating batch tx", "batchID", batchNonceID,
		"magic", cco.config.NetworkMagic, "binary", cco.cardanoCliBinary,
		"slot", slotNumber, "multisig", len(multisigUtxos), "fee", len(feeUtxos), "outputs", len(txOutputs.Outputs))

	// Create Tx
	txRaw, txHash, err := cardano.CreateTx(
		cco.cardanoCliBinary,
		uint(cco.config.NetworkMagic),
		protocolParams,
		slotNumber+cco.config.TTLSlotNumberInc,
		metadata,
		cardano.TxInputInfos{
			MultiSig: &cardano.TxInputInfo{
				PolicyScript: multisigPolicyScript,
				Address:      multisigAddress,
				TxInputs:     convertUTXOsToTxInputs(multisigUtxos),
			},
			MultiSigFee: &cardano.TxInputInfo{
				PolicyScript: multisigFeePolicyScript,
				Address:      multisigFeeAddress,
				TxInputs:     convertUTXOsToTxInputs(feeUtxos),
			},
		},
		txOutputs.Outputs,
	)
	if err != nil {
		return nil, err
	}

	if len(txRaw) > maxTxSize {
		return nil, fmt.Errorf("%w: (size, max) = (%d, %d)",
			errTxSizeTooBig, len(txRaw), maxTxSize)
	}

	return &core.GeneratedBatchTxData{
		TxRaw:  txRaw,
		TxHash: txHash,
	}, nil
}

func (cco *CardanoChainOperations) shouldConsolidate(err error) bool {
	return errors.Is(err, cardanowallet.ErrUTXOsLimitReached) || errors.Is(err, errTxSizeTooBig)
}

func (cco *CardanoChainOperations) generateConsolidationTransaction(
	ctx context.Context,
	bridgeSmartContract eth.IBridgeSmartContract,
	chainID string,
	batchNonceID uint64,
) (*core.GeneratedBatchTxData, error) {
	validatorsData, err := cco.getCardanoData(ctx, bridgeSmartContract, chainID)
	if err != nil {
		return nil, err
	}

	metadata, err := cardano.CreateBatchMetaData(batchNonceID)
	if err != nil {
		return nil, err
	}

	protocolParams, err := cco.txProvider.GetProtocolParameters(ctx)
	if err != nil {
		return nil, err
	}

	multisigPolicyScript, multisigFeePolicyScript, err := cardano.GetPolicyScripts(validatorsData)
	if err != nil {
		return nil, err
	}

	multisigAddress, multisigFeeAddress, err := cardano.GetMultisigAddresses(
		cco.cardanoCliBinary, uint(cco.config.NetworkMagic), multisigPolicyScript, multisigFeePolicyScript)
	if err != nil {
		return nil, err
	}

	multisigUtxos, feeUtxos, err := cco.getUTXOsForConsolidation(multisigAddress, multisigFeeAddress)
	if err != nil {
		return nil, err
	}

	txMultisigOutput, err := getTxOutputFromUtxos(multisigUtxos, multisigAddress)
	if err != nil {
		return nil, err
	}

	slotNumber, err := cco.getSlotNumber()
	if err != nil {
		return nil, err
	}

	cco.logger.Info("Creating consolidation tx", "consolidationTxID", batchNonceID,
		"magic", cco.config.NetworkMagic, "binary", cco.cardanoCliBinary,
		"slot", slotNumber, "multisig", len(multisigUtxos), "fee", len(feeUtxos))

	// Create Tx
	txRaw, txHash, err := cardano.CreateTx(
		cco.cardanoCliBinary,
		uint(cco.config.NetworkMagic),
		protocolParams,
		slotNumber+cco.config.TTLSlotNumberInc,
		metadata,
		cardano.TxInputInfos{
			MultiSig: &cardano.TxInputInfo{
				PolicyScript: multisigPolicyScript,
				Address:      multisigAddress,
				TxInputs:     convertUTXOsToTxInputs(multisigUtxos),
			},
			MultiSigFee: &cardano.TxInputInfo{
				PolicyScript: multisigFeePolicyScript,
				Address:      multisigFeeAddress,
				TxInputs:     convertUTXOsToTxInputs(feeUtxos),
			},
		},
		[]cardanowallet.TxOutput{
			txMultisigOutput,
		},
	)
	if err != nil {
		return nil, err
	}

	if len(txRaw) > maxTxSize {
		return nil, fmt.Errorf("%w: (size, max) = (%d, %d)", errTxSizeTooBig, len(txRaw), maxTxSize)
	}

	return &core.GeneratedBatchTxData{
		IsConsolidation: true,
		TxRaw:           txRaw,
		TxHash:          txHash,
	}, nil
}

func (cco *CardanoChainOperations) getUTXOsForConsolidation(
	multisigAddress, multisigFeeAddress string,
) ([]*indexer.TxInputOutput, []*indexer.TxInputOutput, error) {
	multisigUtxos, err := cco.db.GetAllTxOutputs(multisigAddress, true)
	if err != nil {
		return nil, nil, err
	}

	feeUtxos, err := cco.db.GetAllTxOutputs(multisigFeeAddress, true)
	if err != nil {
		return nil, nil, err
	}

	multisigUtxos, feeUtxos, err = cco.strategy.FilterUtxos(multisigUtxos, feeUtxos, cco.config)
	if err != nil {
		return nil, nil, err
	}

	if len(feeUtxos) == 0 {
		return nil, nil, fmt.Errorf("fee multisig does not have any utxo: %s", multisigFeeAddress)
	}

	cco.logger.Debug("UTXOs retrieved",
		"multisig", multisigAddress, "utxos", multisigUtxos, "fee", multisigFeeAddress, "utxos", feeUtxos)

	// do not take more than maxFeeUtxoCount
	feeUtxos = feeUtxos[:min(cco.config.MaxFeeUtxoCount, len(feeUtxos))]
	// do not take more than maxUtxoCount - length of chosen fee utxos
	multisigUtxos = multisigUtxos[:min(cco.config.MaxUtxoCount-len(feeUtxos), len(multisigUtxos))]

	cco.logger.Debug("UTXOs chosen", "multisig", multisigUtxos, "fee", feeUtxos)

	return multisigUtxos, feeUtxos, nil
}

func (cco *CardanoChainOperations) getUTXOsForNormalBatch(
	multisigAddress, multisigFeeAddress string, protocolParams []byte, txOutputs cardano.TxOutputs,
) ([]*indexer.TxInputOutput, []*indexer.TxInputOutput, error) {
	multisigUtxos, err := cco.db.GetAllTxOutputs(multisigAddress, true)
	if err != nil {
		return nil, nil, err
	}

	feeUtxos, err := cco.db.GetAllTxOutputs(multisigFeeAddress, true)
	if err != nil {
		return nil, nil, err
	}

	multisigUtxos, feeUtxos, err = cco.strategy.FilterUtxos(multisigUtxos, feeUtxos, cco.config)
	if err != nil {
		return nil, nil, err
	}

	cco.logger.Debug("UTXOs retrieved",
		"multisig", multisigAddress, "utxos", multisigUtxos, "fee", multisigFeeAddress, "utxos", feeUtxos)

	lovelaceAmount, err := cco.calculateMinUtxoLovelaceAmount(
		multisigAddress, multisigUtxos, protocolParams, txOutputs.Outputs)
	if err != nil {
		return nil, nil, err
	}

	multisigUtxos, feeUtxos, err = getUTXOsForAmounts(
		cco.config, multisigFeeAddress, multisigUtxos, feeUtxos, txOutputs.Sum, lovelaceAmount)
	if err != nil {
		return nil, nil, err
	}

	cco.logger.Debug("UTXOs chosen", "multisig", multisigUtxos, "fee", feeUtxos)

	return multisigUtxos, feeUtxos, nil
}

func (cco *CardanoChainOperations) getSlotNumber() (uint64, error) {
	data, err := cco.db.GetLatestBlockPoint()
	if err != nil {
		return 0, err
	}

	slot := uint64(0)
	if data != nil {
		slot = data.BlockSlot
	}

	newSlot, err := getNumberWithRoundingThreshold(
		slot, cco.config.SlotRoundingThreshold, cco.config.NoBatchPeriodPercent)
	if err != nil {
		return 0, err
	}

	cco.logger.Debug("calculate slotNumber with rounding", "slot", slot, "newSlot", newSlot)

	return newSlot, nil
}

func (cco *CardanoChainOperations) getCardanoData(
	ctx context.Context, bridgeSmartContract eth.IBridgeSmartContract, chainID string,
) ([]eth.ValidatorChainData, error) {
	validatorsData, err := bridgeSmartContract.GetValidatorsChainData(ctx, chainID)
	if err != nil {
		return nil, err
	}

	hasVerificationKey, hasFeeVerificationKey := false, false

	for _, validator := range validatorsData {
		hasVerificationKey = hasVerificationKey || bytes.Equal(cco.wallet.MultiSig.VerificationKey,
			cardanowallet.PadKeyToSize(validator.Key[0].Bytes()))
		hasFeeVerificationKey = hasFeeVerificationKey || bytes.Equal(cco.wallet.MultiSigFee.VerificationKey,
			cardanowallet.PadKeyToSize(validator.Key[1].Bytes()))
	}

	if !hasVerificationKey {
		return nil, fmt.Errorf(
			"verifying key of current batcher wasn't found in validators data queried from smart contract")
	}

	if !hasFeeVerificationKey {
		return nil, fmt.Errorf(
			"verifying fee key of current batcher wasn't found in validators data queried from smart contract")
	}

	return validatorsData, nil
}

func (cco *CardanoChainOperations) calculateMinUtxoLovelaceAmount(
	multisigAddr string, multisigUtxos []*indexer.TxInputOutput,
	protocolParams []byte, txOutputs []cardanowallet.TxOutput,
) (uint64, error) {
	sumMap := subtractTxOutputsFromSumMap(getSumMapFromTxInputOutput(multisigUtxos), txOutputs)

	tokens, err := cardanowallet.GetTokensFromSumMap(sumMap)
	if err != nil {
		return 0, err
	}

	txBuilder, err := cardanowallet.NewTxBuilder(cco.cardanoCliBinary)
	if err != nil {
		return 0, err
	}

	defer txBuilder.Dispose()

	// calculate final multisig output change
	minUtxo, err := txBuilder.SetProtocolParameters(protocolParams).CalculateMinUtxo(cardanowallet.TxOutput{
		Addr:   multisigAddr,
		Amount: sumMap[cardanowallet.AdaTokenName],
		Tokens: tokens,
	})
	if err != nil {
		return 0, err
	}

	return minUtxo, nil
}

func convertUTXOsToTxInputs(utxos []*indexer.TxInputOutput) (result cardanowallet.TxInputs) {
	// For now we are taking all available UTXOs as fee (should always be 1-2 of them)
	result.Inputs = make([]cardanowallet.TxInput, len(utxos))
	result.Sum = make(map[string]uint64)

	for i, utxo := range utxos {
		result.Inputs[i] = cardanowallet.TxInput{
			Hash:  utxo.Input.Hash.String(),
			Index: utxo.Input.Index,
		}

		result.Sum[cardanowallet.AdaTokenName] += utxo.Output.Amount

		for _, token := range utxo.Output.Tokens {
			result.Sum[token.TokenName()] += token.Amount
		}
	}

	return result
}

func getSumMapFromTxInputOutput(utxos []*indexer.TxInputOutput) map[string]uint64 {
	totalSum := map[string]uint64{}

	for _, utxo := range utxos {
		totalSum[cardanowallet.AdaTokenName] += utxo.Output.Amount

		for _, token := range utxo.Output.Tokens {
			totalSum[token.TokenName()] += token.Amount
		}
	}

	return totalSum
}

func getTxOutputFromUtxos(utxos []*indexer.TxInputOutput, addr string) (cardanowallet.TxOutput, error) {
	totalSum := getSumMapFromTxInputOutput(utxos)
	tokens := make([]cardanowallet.TokenAmount, 0, len(totalSum)-1)

	for tokenName, amount := range totalSum {
		if tokenName != cardanowallet.AdaTokenName {
			newToken, err := cardanowallet.NewTokenWithFullName(tokenName, true)
			if err != nil {
				return cardanowallet.TxOutput{}, err
			}

			tokens = append(tokens, cardanowallet.NewTokenAmount(newToken, amount))
		}
	}

	return cardanowallet.NewTxOutput(addr, totalSum[cardanowallet.AdaTokenName], tokens...), nil
}

func subtractTxOutputsFromSumMap(
	sumMap map[string]uint64, txOutputs []cardanowallet.TxOutput,
) map[string]uint64 {
	for _, out := range txOutputs {
		if value, exists := sumMap[cardanowallet.AdaTokenName]; exists {
			if value > out.Amount {
				sumMap[cardanowallet.AdaTokenName] = value - out.Amount
			} else {
				delete(sumMap, cardanowallet.AdaTokenName)
			}
		}

		for _, token := range out.Tokens {
			tokenName := token.TokenName()
			if value, exists := sumMap[tokenName]; exists {
				if value > token.Amount {
					sumMap[tokenName] = value - token.Amount
				} else {
					delete(sumMap, tokenName)
				}
			}
		}
	}

	return sumMap
}
