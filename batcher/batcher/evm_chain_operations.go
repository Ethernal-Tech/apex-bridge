package batcher

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"sort"

	"github.com/Ethernal-Tech/apex-bridge/batcher/core"
	cardano "github.com/Ethernal-Tech/apex-bridge/cardano"
	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/eth"
	"github.com/Ethernal-Tech/apex-bridge/testenv"
	"github.com/Ethernal-Tech/apex-bridge/validatorobserver"
	eventTrackerStore "github.com/Ethernal-Tech/blockchain-event-tracker/store"
	"github.com/Ethernal-Tech/bn256"
	"github.com/Ethernal-Tech/cardano-infrastructure/secrets"
	"github.com/hashicorp/go-hclog"
)

var (
	_ core.ChainOperations = (*EVMChainOperations)(nil)
)

type EVMChainOperations struct {
	config       *cardano.BatcherEVMChainConfig
	privateKey   *bn256.PrivateKey
	db           eventTrackerStore.EventTrackerStore
	ttlFormatter testenv.TTLFormatterFunc
	gasLimiter   eth.GasLimitHolder
	logger       hclog.Logger
	bridgeSC     eth.IBridgeSmartContract

	// vsc is an internal variable (field) used only during the validator set change
	// process. When false, a standard validator set change tx/batch should be sent.
	// When true, a finalize validator set change tx/batch should be sent. Note: Even
	// when true, a standard validator set change tx/batch may still be sent if the
	// previous validator set change tx/batch failed and requires a retry. This flag
	// is reset to false after a finalize validator set change tx/batch is created.
	vsc bool
}

func NewEVMChainOperations(
	jsonConfig json.RawMessage,
	secretsManager secrets.SecretsManager,
	db eventTrackerStore.EventTrackerStore,
	chainID string,
	logger hclog.Logger,
	bridgeSC eth.IBridgeSmartContract,
) (*EVMChainOperations, error) {
	config, err := cardano.NewBatcherEVMChainConfig(jsonConfig)
	if err != nil {
		return nil, err
	}

	privateKey, err := eth.GetBatcherEVMPrivateKey(secretsManager, chainID)
	if err != nil {
		return nil, err
	}

	return &EVMChainOperations{
		config:       config,
		privateKey:   privateKey,
		db:           db,
		ttlFormatter: testenv.GetTTLFormatter(config.TestMode),
		gasLimiter:   eth.NewGasLimitHolder(submitBatchMinGasLimit, submitBatchMaxGasLimit, submitBatchStepsGasLimit),
		logger:       logger,
		bridgeSC:     bridgeSC,
	}, nil
}

// CreateValidatorSetChangeTx implements core.ChainOperations.
func (cco *EVMChainOperations) CreateValidatorSetChangeTx(
	ctx context.Context,
	chainID string,
	nextBatchID uint64,
	bridgeSmartContract eth.IBridgeSmartContract,
	validatorsKeys *validatorobserver.Validators,
) (*core.GeneratedBatchTxData, error) {
	createVSCTxFn := func() (*core.GeneratedBatchTxData, error) {
		lastProcessedBlock, err := cco.db.GetLastProcessedBlock()
		if err != nil {
			return nil, err
		}

		blockRounded, err := getNumberWithRoundingThreshold(
			lastProcessedBlock, cco.config.BlockRoundingThreshold, cco.config.NoBatchPeriodPercent)
		if err != nil {
			return nil, err
		}

		currentValidatorSetNumber, err := cco.bridgeSC.GetCurrentValidatorSetID(ctx)
		if err != nil {
			return nil, err
		}

		validatorSetNumber := big.NewInt(0).Add(currentValidatorSetNumber, big.NewInt(1))

		ttl := big.NewInt(0).SetUint64((cco.ttlFormatter(blockRounded+cco.config.TTLBlockNumberInc, nextBatchID)))

		keys := make([]eth.ValidatorChainData, 0, len(validatorsKeys.Data[chainID].Keys))

		for _, key := range validatorsKeys.Data[chainID].Keys {
			keys = append(keys, key)
		}

		tx := eth.EVMValidatorSetChangeTx{
			ValidatorsSetNumber: validatorSetNumber,
			TTL:                 ttl,
			ValidatorsChainData: keys,
		}

		txsBytes, err := tx.Pack()
		if err != nil {
			return nil, err
		}

		txsHashBytes, err := common.Keccak256(txsBytes)
		if err != nil {
			return nil, err
		}

		txHash := hex.EncodeToString(txsHashBytes)

		return &core.GeneratedBatchTxData{
			BatchType: uint8(ValidatorSet),
			TxRaw:     txsBytes,
			TxHash:    txHash,
		}, nil
	}

	// The main logic operates as follows: if the "vsc" flag is set to false, it indicates
	// that we have not yet sent a validator set change tx/batch, and therefore we need to
	// create one. Otherwise, we proceed with additional logic. There are two valid and one
	// invalid scenario. The selected path depends on the status of the previously sent tx.
	// Since "vsc" is set to true, the previous batch/tx is guaranteed to be a validator set
	// change tx/batch. If the given batch was successfully executed (status 2), we need to
	// create a finalize validator set change tx and reset the "vsc" flag to false. If the
	// status is 3 (failed), it is necessary to create and resend the validator set change
	// tx/batch (retry). If the status is neither of these two, we return an error indicating
	// an unexpected state.
	//
	// Note: For the above logic to function correctly, it is assumed that the caller invokes
	// this method at the appropriate moment. For example, this method should not be called
	// if the previous batch has not yet been processed (which would return an error due to
	// status 1), or if a finalize tx/batch has already been created but a new validator set
	// change cycle has not yet started (in this case a validator set change tx would be again
	// created, since "vsc" has been reset). See (*BatcherImpl).execute for an example of a
	// correctly implemented caller.
	if !cco.vsc {
		return createVSCTxFn()
	}

	status, _, err := cco.bridgeSC.GetBatchStatusAndTransactions(ctx, chainID, nextBatchID-1)
	if err != nil {
		return nil, err
	}

	switch status {
	case 2:
		cco.vsc = false

		return &core.GeneratedBatchTxData{
			BatchType: uint8(ValidatorSetFinal),
			TxRaw:     []byte{},
			TxHash:    "",
		}, nil
	case 3:
		return createVSCTxFn()
	default:
		return nil, fmt.Errorf("unexpected status %d for batch with ID %d", status, nextBatchID-1)
	}
}

// GenerateBatchTransaction implements core.ChainOperations.
func (cco *EVMChainOperations) GenerateBatchTransaction(
	ctx context.Context,
	bridgeSmartContract eth.IBridgeSmartContract,
	chainID string,
	confirmedTransactions []eth.ConfirmedTransaction,
	batchNonceID uint64,
) (*core.GeneratedBatchTxData, error) {
	lastProcessedBlock, err := cco.db.GetLastProcessedBlock()
	if err != nil {
		return nil, err
	}

	blockRounded, err := getNumberWithRoundingThreshold(
		lastProcessedBlock, cco.config.BlockRoundingThreshold, cco.config.NoBatchPeriodPercent)
	if err != nil {
		return nil, err
	}

	txs := newEVMSmartContractTransaction(
		batchNonceID,
		cco.ttlFormatter(blockRounded+cco.config.TTLBlockNumberInc, batchNonceID),
		confirmedTransactions)

	txsBytes, err := txs.Pack()
	if err != nil {
		return nil, err
	}

	txsHashBytes, err := common.Keccak256(txsBytes)
	if err != nil {
		return nil, err
	}

	txHash := hex.EncodeToString(txsHashBytes)

	cco.logger.Debug("Batch transaction data has been generated",
		"id", batchNonceID, "tx", txs, "hash", txHash,
		"lastBlock", lastProcessedBlock,
		"rounding", cco.config.BlockRoundingThreshold,
		"noBatchPercent", cco.config.NoBatchPeriodPercent)

	return &core.GeneratedBatchTxData{
		TxRaw:  txsBytes,
		TxHash: txHash,
	}, nil
}

// SignBatchTransaction implements core.ChainOperations.
func (cco *EVMChainOperations) SignBatchTransaction(
	generatedBatchData *core.GeneratedBatchTxData) ([]byte, []byte, error) {
	txsHashBytes, err := common.DecodeHex(generatedBatchData.TxHash)
	if err != nil {
		return nil, nil, err
	}

	signature, err := cco.privateKey.Sign(txsHashBytes, eth.BN256Domain)
	if err != nil {
		return nil, nil, err
	}

	signatureBytes, err := signature.Marshal()
	if err != nil {
		return nil, nil, err
	}

	if cco.logger.IsDebug() {
		cco.logger.Debug("Signature has been created",
			"signature", hex.EncodeToString(signatureBytes),
			"public", hex.EncodeToString(cco.privateKey.PublicKey().Marshal()))
	}

	return signatureBytes, nil, nil
}

func (cco *EVMChainOperations) IsSynchronized(
	ctx context.Context, bridgeSmartContract eth.IBridgeSmartContract, chainID string,
) (bool, error) {
	lastObservedBlockBridge, err := bridgeSmartContract.GetLastObservedBlock(ctx, chainID)
	if err != nil {
		return false, err
	}

	latestBlock, err := cco.db.GetLastProcessedBlock()
	if err != nil {
		return false, err
	}

	return latestBlock >= lastObservedBlockBridge.BlockSlot.Uint64(), nil
}

func (cco *EVMChainOperations) Submit(
	ctx context.Context, bridgeSmartContract eth.IBridgeSmartContract, batch eth.SignedBatch,
) error {
	err := bridgeSmartContract.SubmitSignedBatchEVM(ctx, batch, cco.gasLimiter.GetGasLimit())

	cco.gasLimiter.Update(err)

	return err
}

func newEVMSmartContractTransaction(
	batchNonceID uint64, ttl uint64, confirmedTransactions []eth.ConfirmedTransaction,
) eth.EVMSmartContractTransaction {
	sourceAddrTxMap := map[string]eth.EVMSmartContractTransactionReceiver{}
	feeAmount := big.NewInt(0)

	for _, tx := range confirmedTransactions {
		for _, recv := range tx.Receivers {
			weiAmount := common.DfmToWei(recv.Amount)

			if recv.DestinationAddress == common.EthZeroAddr {
				feeAmount.Add(feeAmount, weiAmount)

				continue
			}

			val, exists := sourceAddrTxMap[recv.DestinationAddress]
			if !exists {
				val.Amount = weiAmount
				val.Address = common.HexToAddress(recv.DestinationAddress)
			} else {
				val.Amount.Add(val.Amount, weiAmount)
			}

			sourceAddrTxMap[recv.DestinationAddress] = val
		}
	}

	receivers := make([]eth.EVMSmartContractTransactionReceiver, 0, len(sourceAddrTxMap))

	for _, v := range sourceAddrTxMap {
		receivers = append(receivers, v)
	}

	// every batcher should have same order
	sort.Slice(receivers, func(i, j int) bool {
		return receivers[i].Address.Cmp(receivers[j].Address) < 0
	})

	return eth.EVMSmartContractTransaction{
		BatchNonceID: batchNonceID,
		TTL:          ttl,
		FeeAmount:    feeAmount,
		Receivers:    receivers,
	}
}
