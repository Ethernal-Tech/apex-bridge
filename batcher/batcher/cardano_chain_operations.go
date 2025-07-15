package batcher

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/Ethernal-Tech/apex-bridge/batcher/core"
	cardano "github.com/Ethernal-Tech/apex-bridge/cardano"
	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/eth"
	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
	"github.com/Ethernal-Tech/cardano-infrastructure/secrets"
	cardanowallet "github.com/Ethernal-Tech/cardano-infrastructure/wallet"
	"github.com/hashicorp/go-hclog"
)

var (
	_ core.ChainOperations = (*CardanoChainOperations)(nil)

	errTxSizeTooBig = errors.New("batch tx size too big")
)

// Get real tx size from protocolParams/config
const (
	maxTxSize = 16000
)

type batchInitialData struct {
	BatchNonceID           uint64
	Metadata               []byte
	ProtocolParams         []byte
	MultisigPolicyScript   *cardanowallet.PolicyScript
	FeePolicyScript        *cardanowallet.PolicyScript
	MultisigAddr           string
	FeeAddr                string
	MultisigStakeKeyHashes []string
}

type CardanoChainOperations struct {
	config           *cardano.CardanoChainConfig
	wallet           *cardano.ApexCardanoWallet
	txProvider       cardanowallet.ITxDataRetriever
	db               indexer.Database
	gasLimiter       eth.GasLimitHolder
	cardanoCliBinary string
	logger           hclog.Logger
}

func NewCardanoChainOperations(
	jsonConfig json.RawMessage,
	db indexer.Database,
	secretsManager secrets.SecretsManager,
	chainID string,
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
	data, err := cco.createBatchInitialData(ctx, bridgeSmartContract, chainID, batchNonceID)
	if err != nil {
		return nil, err
	}

	txData, err := cco.generateBatchTransaction(data, confirmedTransactions)

	if cco.shouldConsolidate(err) {
		cco.logger.Warn("consolidation batch generation started", "err", err)

		txData, err = cco.generateConsolidationTransaction(data)
		if err != nil {
			err = fmt.Errorf("consolidation batch failed: %w", err)
		}
	}

	return txData, err
}

// SignBatchTransaction implements core.ChainOperations.
func (cco *CardanoChainOperations) SignBatchTransaction(
	generatedBatchData *core.GeneratedBatchTxData) (*core.BatchSignatures, error) {
	txBuilder, err := cardanowallet.NewTxBuilder(cco.cardanoCliBinary)
	if err != nil {
		return nil, err
	}

	defer txBuilder.Dispose()

	var (
		multisigWitness      []byte
		stakeMultisigWitness []byte
	)

	if generatedBatchData.IsPaymentSignNeeded {
		multisigWitness, err = txBuilder.CreateTxWitness(generatedBatchData.TxRaw, cco.wallet.MultiSig)
		if err != nil {
			return nil, err
		}
	}

	if generatedBatchData.IsStakeSignNeeded {
		stakeMultisigWitness, err = txBuilder.CreateTxWitness(
			generatedBatchData.TxRaw, cardanowallet.NewStakeSigner(cco.wallet.MultiSig))
		if err != nil {
			return nil, err
		}
	}

	feeWitness, err := txBuilder.CreateTxWitness(generatedBatchData.TxRaw, cco.wallet.Fee)
	if err != nil {
		return nil, err
	}

	return &core.BatchSignatures{
		Multisig:     multisigWitness,
		MultsigStake: stakeMultisigWitness,
		Fee:          feeWitness,
	}, nil
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
	data *batchInitialData,
	confirmedTransactions []eth.ConfirmedTransaction,
) (*core.GeneratedBatchTxData, error) {
	certificates := make([]*cardano.CertificatesWithScript, 0)
	keyRegistrationFee := uint64(0)
	hasStakeDelegationTx := false
	hasNonStakeDelegationTx := false

	for _, tx := range confirmedTransactions {
		if tx.TransactionType == uint8(common.StakeDelConfirmedTxType) {
			certificate, depositAmount, err := cco.getStakingDelegateCertificate(data, &tx)
			if err != nil {
				return nil, err
			}

			hasStakeDelegationTx = true
			keyRegistrationFee += depositAmount

			certificates = append(certificates, certificate)
		} else {
			hasNonStakeDelegationTx = true
		}
	}

	var certificatedData *cardano.CertificatesData = nil

	if len(certificates) > 0 {
		certificatedData = &cardano.CertificatesData{
			Certificates:    certificates,
			RegistrationFee: keyRegistrationFee,
		}
	}

	// dirty hack. do not take any multisig utxo if there are no bridging/refund/nonStake txs
	if !hasNonStakeDelegationTx {
		defer func(oldTakeAtLeastUtxoCount uint) {
			cco.config.TakeAtLeastUtxoCount = oldTakeAtLeastUtxoCount
		}(cco.config.TakeAtLeastUtxoCount)

		cco.config.TakeAtLeastUtxoCount = 0
	}

	txOutputs, err := getOutputs(confirmedTransactions, cco.config, cco.logger)
	if err != nil {
		return nil, err
	}

	multisigUtxos, feeUtxos, err := cco.getUTXOsForNormalBatch(
		data.MultisigAddr, data.FeeAddr, data.ProtocolParams, txOutputs)
	if err != nil {
		return nil, err
	}

	slotNumber, err := cco.getSlotNumber()
	if err != nil {
		return nil, err
	}

	cco.logger.Info("Creating batch tx", "batchID", data.BatchNonceID,
		"magic", cco.config.NetworkMagic, "binary", cco.cardanoCliBinary,
		"slot", slotNumber, "multisig", len(multisigUtxos), "fee", len(feeUtxos), "outputs", len(txOutputs.Outputs))

	// Create Tx
	txRaw, txHash, err := cardano.CreateTx(
		cco.cardanoCliBinary,
		uint(cco.config.NetworkMagic),
		data.ProtocolParams,
		slotNumber+cco.config.TTLSlotNumberInc,
		data.Metadata,
		cardano.TxInputInfos{
			MultiSig: &cardano.TxInputInfo{
				PolicyScript: data.MultisigPolicyScript,
				Address:      data.MultisigAddr,
				TxInputs:     convertUTXOsToTxInputs(multisigUtxos),
			},
			MultiSigFee: &cardano.TxInputInfo{
				PolicyScript: data.FeePolicyScript,
				Address:      data.FeeAddr,
				TxInputs:     convertUTXOsToTxInputs(feeUtxos),
			},
		},
		txOutputs.Outputs,
		certificatedData,
	)
	if err != nil {
		return nil, err
	}

	if len(txRaw) > maxTxSize {
		return nil, fmt.Errorf("%w: (size, max) = (%d, %d)",
			errTxSizeTooBig, len(txRaw), maxTxSize)
	}

	return &core.GeneratedBatchTxData{
		TxRaw:               txRaw,
		TxHash:              txHash,
		IsStakeSignNeeded:   hasStakeDelegationTx,
		IsPaymentSignNeeded: hasNonStakeDelegationTx,
	}, nil
}

func (cco *CardanoChainOperations) shouldConsolidate(err error) bool {
	return errors.Is(err, cardanowallet.ErrUTXOsLimitReached) || errors.Is(err, errTxSizeTooBig)
}

func (cco *CardanoChainOperations) generateConsolidationTransaction(
	data *batchInitialData,
) (*core.GeneratedBatchTxData, error) {
	multisigUtxos, feeUtxos, err := cco.getUTXOsForConsolidation(data.MultisigAddr, data.FeeAddr)
	if err != nil {
		return nil, err
	}

	multisigTxOutput, err := getTxOutputFromSumMap(data.MultisigAddr, getSumMapFromTxInputOutput(multisigUtxos))
	if err != nil {
		return nil, err
	}

	slotNumber, err := cco.getSlotNumber()
	if err != nil {
		return nil, err
	}

	cco.logger.Info("Creating consolidation tx", "consolidationTxID", data.BatchNonceID,
		"magic", cco.config.NetworkMagic, "binary", cco.cardanoCliBinary,
		"slot", slotNumber, "multisig", len(multisigUtxos), "fee", len(feeUtxos))

	// Create Tx
	txRaw, txHash, err := cardano.CreateTx(
		cco.cardanoCliBinary,
		uint(cco.config.NetworkMagic),
		data.ProtocolParams,
		slotNumber+cco.config.TTLSlotNumberInc,
		data.Metadata,
		cardano.TxInputInfos{
			MultiSig: &cardano.TxInputInfo{
				PolicyScript: data.MultisigPolicyScript,
				Address:      data.MultisigAddr,
				TxInputs:     convertUTXOsToTxInputs(multisigUtxos),
			},
			MultiSigFee: &cardano.TxInputInfo{
				PolicyScript: data.FeePolicyScript,
				Address:      data.FeeAddr,
				TxInputs:     convertUTXOsToTxInputs(feeUtxos),
			},
		},
		[]cardanowallet.TxOutput{multisigTxOutput},
		nil,
	)
	if err != nil {
		return nil, err
	}

	if len(txRaw) > maxTxSize {
		return nil, fmt.Errorf("%w: (size, max) = (%d, %d)", errTxSizeTooBig, len(txRaw), maxTxSize)
	}

	return &core.GeneratedBatchTxData{
		IsConsolidation:     true,
		IsPaymentSignNeeded: true,
		TxRaw:               txRaw,
		TxHash:              txHash,
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

	knownTokens, err := cardano.GetKnownTokens(cco.config)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get known tokens: %w", err)
	}

	multisigUtxos = filterOutUtxosWithUnknownTokens(multisigUtxos, knownTokens...)
	feeUtxos = filterOutUtxosWithUnknownTokens(feeUtxos)

	if len(feeUtxos) == 0 {
		return nil, nil, fmt.Errorf("fee multisig does not have any utxo: %s", multisigFeeAddress)
	}

	cco.logger.Debug("UTXOs retrieved",
		"multisig", multisigAddress, "utxos", multisigUtxos, "fee", multisigFeeAddress, "utxos", feeUtxos)

	// do not take more than maxFeeUtxoCount
	feeUtxos = feeUtxos[:min(int(cco.config.MaxFeeUtxoCount), len(feeUtxos))] //nolint:gosec
	// do not take more than maxUtxoCount - length of chosen fee utxos
	maxUtxosCnt := min(getMaxUtxoCount(cco.config, len(feeUtxos)), len(multisigUtxos))
	multisigUtxos = multisigUtxos[:maxUtxosCnt]

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

	knownTokens, err := cardano.GetKnownTokens(cco.config)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get known tokens: %w", err)
	}

	multisigUtxos = filterOutUtxosWithUnknownTokens(multisigUtxos, knownTokens...)
	feeUtxos = filterOutUtxosWithUnknownTokens(feeUtxos)

	cco.logger.Debug("UTXOs retrieved",
		"multisig", multisigAddress, "utxos", multisigUtxos, "fee", multisigFeeAddress, "utxos", feeUtxos)

	if len(feeUtxos) == 0 {
		return nil, nil, fmt.Errorf("fee multisig does not have any utxo: %s", multisigFeeAddress)
	}

	feeUtxos = feeUtxos[:min(cco.config.MaxFeeUtxoCount, uint(len(feeUtxos)))] // do not take more than MaxFeeUtxoCount

	minUtxoLovelaceAmount, err := calculateMinUtxoLovelaceAmount(
		cco.cardanoCliBinary, protocolParams, multisigAddress, multisigUtxos, txOutputs.Outputs)
	if err != nil {
		return nil, nil, err
	}

	multisigUtxos, err = getNeededUtxos(
		multisigUtxos,
		txOutputs.Sum,
		minUtxoLovelaceAmount,
		getMaxUtxoCount(cco.config, len(feeUtxos)),
		int(cco.config.TakeAtLeastUtxoCount), //nolint:gosec
	)
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

func (cco *CardanoChainOperations) getValidatorsChainData(
	ctx context.Context, bridgeSmartContract eth.IBridgeSmartContract, chainID string,
) ([]eth.ValidatorChainData, error) {
	validatorsData, err := bridgeSmartContract.GetValidatorsChainData(ctx, chainID)
	if err != nil {
		return nil, err
	}

	for _, data := range validatorsData {
		if cardano.AreVerifyingKeysTheSame(cco.wallet, data) {
			return validatorsData, nil
		}
	}

	return nil, fmt.Errorf(
		"verifying keys of current batcher wasn't found in validators data queried from smart contract")
}

func (cco *CardanoChainOperations) createBatchInitialData(
	ctx context.Context,
	bridgeSmartContract eth.IBridgeSmartContract,
	chainID string,
	batchNonceID uint64,
) (*batchInitialData, error) {
	validatorsData, err := cco.getValidatorsChainData(ctx, bridgeSmartContract, chainID)
	if err != nil {
		return nil, err
	}

	metadata, err := common.MarshalMetadata(common.MetadataEncodingTypeJSON, common.BatchExecutedMetadata{
		BridgingTxType: common.BridgingTxTypeBatchExecution,
		BatchNonceID:   batchNonceID,
	})
	if err != nil {
		return nil, err
	}

	protocolParams, err := cco.txProvider.GetProtocolParameters(ctx)
	if err != nil {
		return nil, err
	}

	keyHashes, err := cardano.NewApexKeyHashes(validatorsData)
	if err != nil {
		return nil, err
	}

	policyScripts := cardano.NewApexPolicyScripts(keyHashes)

	addresses, err := cardano.NewApexAddresses(cco.cardanoCliBinary, uint(cco.config.NetworkMagic), policyScripts)
	if err != nil {
		return nil, err
	}

	return &batchInitialData{
		BatchNonceID:           batchNonceID,
		Metadata:               metadata,
		ProtocolParams:         protocolParams,
		MultisigPolicyScript:   policyScripts.Multisig.Payment,
		FeePolicyScript:        policyScripts.Fee.Payment,
		MultisigAddr:           addresses.Multisig.Payment,
		FeeAddr:                addresses.Fee.Payment,
		MultisigStakeKeyHashes: keyHashes.Multisig.Stake,
	}, nil
}

func (cco *CardanoChainOperations) getStakingDelegateCertificate(
	data *batchInitialData, tx *eth.ConfirmedTransaction,
) (*cardano.CertificatesWithScript, uint64, error) {
	// Generate policy script
	quorumCount := int(common.GetRequiredSignaturesForConsensus(uint64(len(data.MultisigStakeKeyHashes)))) //nolint:gosec
	policyScript := cardanowallet.NewPolicyScript(data.MultisigStakeKeyHashes, quorumCount,
		cardanowallet.WithAfter(uint64(tx.BridgeAddrIndex)))

	// Generate certificates
	keyRegDepositAmount, err := extractStakeKeyDepositAmount(data.ProtocolParams)
	if err != nil {
		return nil, 0, err
	}

	cliUtils := cardanowallet.NewCliUtils(cco.cardanoCliBinary)

	multisigStakeAddress, err := cliUtils.GetPolicyScriptRewardAddress(uint(cco.config.NetworkMagic), policyScript)
	if err != nil {
		return nil, 0, err
	}

	registrationCert, err := cliUtils.CreateRegistrationCertificate(multisigStakeAddress, keyRegDepositAmount)
	if err != nil {
		return nil, 0, err
	}

	delegationCert, err := cliUtils.CreateDelegationCertificate(multisigStakeAddress, tx.StakePoolId)
	if err != nil {
		return nil, 0, err
	}

	return &cardano.CertificatesWithScript{
		PolicyScript: policyScript,
		Certificates: []cardanowallet.ICertificate{registrationCert, delegationCert},
	}, keyRegDepositAmount, nil
}

func getMaxUtxoCount(config *cardano.CardanoChainConfig, prevUtxosCnt int) int {
	return max(int(config.MaxUtxoCount)-prevUtxosCnt, 0) //nolint:gosec
}
