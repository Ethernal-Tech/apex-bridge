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
	validatorsKeys validatorobserver.ValidatorsPerChain,
	lastBatchID uint64,
	lastBatchType uint8,
) (bool, *core.GeneratedBatchTxData, error) {
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

		keys := make([]eth.ValidatorChainData, 0, len(validatorsKeys[chainID].Keys))

		for _, key := range validatorsKeys[chainID].Keys {
			keys = append(keys, key)
		}

		tx := eth.EVMValidatorSetChangeTx{
			BatchNonceID:        nextBatchID,
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

	if lastBatchType != uint8(ValidatorSet) {
		// vsc tx not sent, send it. It is not possible to get here otherwise.
		batch, err := createVSCTxFn()

		return false, batch, err
	}

	// vsc tx sent, check its status and resend if needed
	status, _, err := cco.bridgeSC.GetBatchStatusAndTransactions(ctx, chainID, lastBatchID)
	if err != nil {
		return false, nil, err
	}

	switch status {
	case 2:
		// vsc tx executed on evm chain, send final
		txRaw := []byte("0xdeadbeef")

		txsHashBytes, err := common.Keccak256(txRaw)
		if err != nil {
			return false, nil, err
		}

		txHash := hex.EncodeToString(txsHashBytes)

		return false, &core.GeneratedBatchTxData{
			BatchType: uint8(ValidatorSetFinal),
			TxRaw:     txRaw,
			TxHash:    txHash,
		}, nil
	case 3:
		// vsc tx failed on evm chain, resend
		batch, err := createVSCTxFn()

		return true, batch, err
	default:
		return false, nil, fmt.Errorf("unexpected status %d for batch with ID %d", status, nextBatchID-1)
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
		confirmedTransactions,
		common.DfmToWei(new(big.Int).SetUint64(cco.config.MinFeeForBridging)))

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
	batchNonceID uint64,
	ttl uint64,
	confirmedTransactions []eth.ConfirmedTransaction,
	minFeeForBridging *big.Int,
) eth.EVMSmartContractTransaction {
	sourceAddrTxMap := map[string]eth.EVMSmartContractTransactionReceiver{}
	feeAmount := big.NewInt(0)

	updateAmount := func(mp map[string]eth.EVMSmartContractTransactionReceiver, addr string, amount *big.Int) {
		val, exists := mp[addr]
		if !exists {
			val.Amount = amount
			val.Address = common.HexToAddress(addr)
		} else {
			val.Amount.Add(val.Amount, amount)
		}

		mp[addr] = val
	}

	for _, tx := range confirmedTransactions {
		for _, recv := range tx.Receivers {
			amount := common.DfmToWei(recv.Amount)
			// In case a transaction is of type refund, batcher should transfer minFeeForBridging
			// to fee payer address, and the rest is transferred to the user.
			// if else would be nicer but linter does not think the same way
			if tx.TransactionType == uint8(common.RefundConfirmedTxType) {
				feeAmount.Add(feeAmount, minFeeForBridging)
				updateAmount(sourceAddrTxMap, recv.DestinationAddress, amount.Sub(amount, minFeeForBridging))

				continue
			}

			if recv.DestinationAddress == common.EthZeroAddr {
				feeAmount.Add(feeAmount, amount)

				continue
			}

			updateAmount(sourceAddrTxMap, recv.DestinationAddress, amount)
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

func (cco *EVMChainOperations) GenerateMultisigAddress(
	validators *validatorobserver.ValidatorsPerChain, chainID string) error {
	return nil
}
