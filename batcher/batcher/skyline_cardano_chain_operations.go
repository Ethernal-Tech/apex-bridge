package batcher

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"

	"github.com/Ethernal-Tech/apex-bridge/batcher/core"
	cardano "github.com/Ethernal-Tech/apex-bridge/cardano"
	"github.com/Ethernal-Tech/apex-bridge/eth"
	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
	"github.com/Ethernal-Tech/cardano-infrastructure/secrets"
	cardanowallet "github.com/Ethernal-Tech/cardano-infrastructure/wallet"
	"github.com/hashicorp/go-hclog"
)

var (
	_ core.ChainOperations = (*SkylineCardanoChainOperations)(nil)
)

type SkylineCardanoChainOperations struct {
	config           *cardano.CardanoChainConfig
	wallet           *cardano.CardanoWallet
	txProvider       cardanowallet.ITxDataRetriever
	db               indexer.Database
	gasLimiter       eth.GasLimitHolder
	cardanoCliBinary string
	logger           hclog.Logger
}

func NewSkylineCardanoChainOperations(
	jsonConfig json.RawMessage,
	db indexer.Database,
	secretsManager secrets.SecretsManager,
	chainID string,
	logger hclog.Logger,
) (*SkylineCardanoChainOperations, error) {
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

	return &SkylineCardanoChainOperations{
		wallet:           cardanoWallet,
		config:           cardanoConfig,
		txProvider:       txProvider,
		cardanoCliBinary: cardanowallet.ResolveCardanoCliBinary(cardanoConfig.NetworkID),
		gasLimiter:       eth.NewGasLimitHolder(submitBatchMinGasLimit, submitBatchMaxGasLimit, submitBatchStepsGasLimit),
		db:               db,
		logger:           logger,
	}, nil
}

// GenerateBatchTransaction implements core.ChainOperations.
func (scco *SkylineCardanoChainOperations) GenerateBatchTransaction(
	ctx context.Context,
	bridgeSmartContract eth.IBridgeSmartContract,
	chainID string,
	confirmedTransactions []eth.ConfirmedTransaction,
	batchNonceID uint64,
) (*core.GeneratedBatchTxData, error) {
	validatorsData, err := scco.getCardanoData(ctx, bridgeSmartContract, chainID)
	if err != nil {
		return nil, err
	}

	metadata, err := cardano.CreateBatchMetaData(batchNonceID)
	if err != nil {
		return nil, err
	}

	protocolParams, err := scco.txProvider.GetProtocolParameters(ctx)
	if err != nil {
		return nil, err
	}

	slotNumber, err := scco.getSlotNumber()
	if err != nil {
		return nil, err
	}

	multisigPolicyScript, multisigFeePolicyScript, err := cardano.GetPolicyScripts(validatorsData)
	if err != nil {
		return nil, err
	}

	multisigAddress, multisigFeeAddress, err := cardano.GetMultisigAddresses(
		scco.cardanoCliBinary, uint(scco.config.NetworkMagic), multisigPolicyScript, multisigFeePolicyScript)
	if err != nil {
		return nil, err
	}

	// this has to change
	txOutputs, err := getSkylineOutputs(confirmedTransactions, scco.config.Destinations, chainID, scco.config.NetworkID, scco.logger)
	if err != nil {
		return nil, errors.New("no valid tx outputs")
	}

	multisigUtxos, feeUtxos, err := scco.getUTXOs(
		multisigAddress, multisigFeeAddress, *txOutputs)
	if err != nil {
		return nil, err
	}

	scco.logger.Info("Creating batch tx", "batchID", batchNonceID,
		"magic", scco.config.NetworkMagic, "binary", scco.cardanoCliBinary,
		"slot", slotNumber, "multisig", len(multisigUtxos), "fee", len(feeUtxos), "outputs", len(txOutputs.Outputs))

	// Create Tx
	txRaw, txHash, err := cardano.CreateTx(
		scco.cardanoCliBinary,
		uint(scco.config.NetworkMagic),
		protocolParams,
		slotNumber+scco.config.TTLSlotNumberInc,
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
		return nil, errors.New("fatal error, tx size too big")
	}

	return &core.GeneratedBatchTxData{
		TxRaw:  txRaw,
		TxHash: txHash,
	}, nil
}

// SignBatchTransaction implements core.ChainOperations.
func (scco *SkylineCardanoChainOperations) SignBatchTransaction(txHash string) ([]byte, []byte, error) {
	witnessMultiSig, err := cardanowallet.CreateTxWitness(txHash, scco.wallet.MultiSig)
	if err != nil {
		return nil, nil, err
	}

	witnessMultiSigFee, err := cardanowallet.CreateTxWitness(txHash, scco.wallet.MultiSigFee)
	if err != nil {
		return nil, nil, err
	}

	return witnessMultiSig, witnessMultiSigFee, nil
}

// IsSynchronized implements core.IsSynchronized.
func (scco *SkylineCardanoChainOperations) IsSynchronized(
	ctx context.Context, bridgeSmartContract eth.IBridgeSmartContract, chainID string,
) (bool, error) {
	lastObservedBlockBridge, err := bridgeSmartContract.GetLastObservedBlock(ctx, chainID)
	if err != nil {
		return false, err
	}

	lastOracleBlockPoint, err := scco.db.GetLatestBlockPoint()
	if err != nil {
		return false, err
	}

	return lastOracleBlockPoint != nil &&
		lastOracleBlockPoint.BlockSlot >= lastObservedBlockBridge.BlockSlot.Uint64(), nil
}

// Submit implements core.Submit.
func (scco *SkylineCardanoChainOperations) Submit(
	ctx context.Context, bridgeSmartContract eth.IBridgeSmartContract, batch eth.SignedBatch,
) error {
	err := bridgeSmartContract.SubmitSignedBatch(ctx, batch, scco.gasLimiter.GetGasLimit())

	scco.gasLimiter.Update(err)

	return err
}

func (scco *SkylineCardanoChainOperations) getSlotNumber() (uint64, error) {
	data, err := scco.db.GetLatestBlockPoint()
	if err != nil {
		return 0, err
	}

	slot := uint64(0)
	if data != nil {
		slot = data.BlockSlot
	}

	newSlot, err := getNumberWithRoundingThreshold(
		slot, scco.config.SlotRoundingThreshold, scco.config.NoBatchPeriodPercent)
	if err != nil {
		return 0, err
	}

	scco.logger.Debug("calculate slotNumber with rounding", "slot", slot, "newSlot", newSlot)

	return newSlot, nil
}

func (scco *SkylineCardanoChainOperations) getCardanoData(
	ctx context.Context, bridgeSmartContract eth.IBridgeSmartContract, chainID string,
) ([]eth.ValidatorChainData, error) {
	validatorsData, err := bridgeSmartContract.GetValidatorsChainData(ctx, chainID)
	if err != nil {
		return nil, err
	}

	hasVerificationKey, hasFeeVerificationKey := false, false

	for _, validator := range validatorsData {
		hasVerificationKey = hasVerificationKey || bytes.Equal(scco.wallet.MultiSig.VerificationKey,
			cardanowallet.PadKeyToSize(validator.Key[0].Bytes()))
		hasFeeVerificationKey = hasFeeVerificationKey || bytes.Equal(scco.wallet.MultiSigFee.VerificationKey,
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

// this has to chabge
func (scco *SkylineCardanoChainOperations) getUTXOs(
	multisigAddress, multisigFeeAddress string, txOutputs cardano.TxOutputs,
) (multisigUtxos []*indexer.TxInputOutput, feeUtxos []*indexer.TxInputOutput, err error) {
	multisigUtxos, err = scco.db.GetAllTxOutputs(multisigAddress, true)
	if err != nil {
		return
	}

	feeUtxos, err = scco.db.GetAllTxOutputs(multisigFeeAddress, true)
	if err != nil {
		return
	}

	multisigUtxos = filterOutTokenUtxos(multisigUtxos)
	feeUtxos = filterOutTokenUtxos(feeUtxos)

	if len(feeUtxos) == 0 {
		return nil, nil, fmt.Errorf("fee multisig does not have any utxo: %s", multisigFeeAddress)
	}

	scco.logger.Debug("UTXOs retrieved",
		"multisig", multisigAddress, "utxos", multisigUtxos, "fee", multisigFeeAddress, "utxos", feeUtxos)

	feeUtxos = feeUtxos[:min(maxFeeUtxoCount, len(feeUtxos))] // do not take more than maxFeeUtxoCount

	multisigUtxos, err = getNeededUtxos(
		multisigUtxos,
		txOutputs.Sum[cardanowallet.AdaTokenName],
		scco.config.UtxoMinAmount,
		len(feeUtxos)+len(txOutputs.Outputs),
		maxUtxoCount,
		scco.config.TakeAtLeastUtxoCount,
	)
	if err != nil {
		return
	}

	scco.logger.Debug("UTXOs chosen", "multisig", multisigUtxos, "fee", feeUtxos)

	return
}

func getConfigTokenExchange(destChainID string, isDestNativeToken bool, dests []cardano.CardanoConfigTokenExchange) (result cardano.CardanoConfigTokenExchange) {
	for _, x := range dests {
		if x.Chain != destChainID {
			continue
		}
		if isDestNativeToken && x.SrcTokenName == cardanowallet.AdaTokenName ||
			!isDestNativeToken && x.DstTokenName == cardanowallet.AdaTokenName {
			return x
		}
	}
	return result
}

func getSkylineOutputs(
	txs []eth.ConfirmedTransaction, destinations []cardano.CardanoConfigTokenExchange, destinationChainID string, networkID cardanowallet.CardanoNetworkType, logger hclog.Logger,
) (*cardano.TxOutputs, error) {
	receiversMap := map[string]cardanowallet.TxOutput{}

	for _, transaction := range txs {
		for _, receiver := range transaction.Receivers {
			data := receiversMap[receiver.DestinationAddress]
			data.Amount += receiver.Amount.Uint64()

			if receiver.AmountWrapped != nil {
				if len(data.Tokens) == 0 {
					tconf := getConfigTokenExchange(destinationChainID, true, destinations)
					token, err := cardanowallet.NewTokenAmountWithFullName(tconf.DstTokenName, 0, true)
					if err != nil {
						return nil, fmt.Errorf("failed to create new token amount")
					}
					data.Tokens = []cardanowallet.TokenAmount{token}
				}

				data.Tokens[0].Amount += receiver.AmountWrapped.Uint64()
			}

			receiversMap[receiver.DestinationAddress] = data
		}
	}

	result := cardano.TxOutputs{
		Outputs: make([]cardanowallet.TxOutput, 0, len(receiversMap)),
		Sum:     map[string]uint64{},
	}

	for addr, txOut := range receiversMap {
		if txOut.Amount == 0 {
			logger.Warn("skipped output with zero amount", "addr", addr)

			continue
		} else if !cardano.IsValidOutputAddress(addr, networkID) {
			logger.Warn("skipped output because it is invalid", "addr", addr)

			continue
		}

		txOut.Addr = addr

		result.Outputs = append(result.Outputs, txOut)
		result.Sum[cardanowallet.AdaTokenName] += txOut.Amount
		for _, token := range txOut.Tokens {
			result.Sum[token.TokenName()] += token.Amount
		}

	}

	// sort outputs because all batchers should have same order of outputs
	sort.Slice(result.Outputs, func(i, j int) bool {
		return result.Outputs[i].Addr < result.Outputs[j].Addr
	})

	return &result, nil
}
