package batcher

import (
	"bytes"
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
	eventTrackerStore "github.com/Ethernal-Tech/blockchain-event-tracker/store"
	"github.com/Ethernal-Tech/bn256"
	"github.com/Ethernal-Tech/cardano-infrastructure/secrets"
	"github.com/hashicorp/go-hclog"
)

var (
	_ core.ChainOperations = (*EVMChainOperationsTestMode)(nil)

	count = 0
)

type EVMChainOperationsTestMode struct {
	config     *cardano.BatcherEVMChainConfig
	privateKey *bn256.PrivateKey
	db         eventTrackerStore.EventTrackerStore
	logger     hclog.Logger
	testMode   uint8
}

func NewEVMChainOperationsTestMode(
	jsonConfig json.RawMessage,
	secretsManager secrets.SecretsManager,
	db eventTrackerStore.EventTrackerStore,
	chainID string,
	logger hclog.Logger,
	testMode uint8,
) (*EVMChainOperationsTestMode, error) {
	config, err := cardano.NewBatcherEVMChainConfig(jsonConfig)
	if err != nil {
		return nil, err
	}

	privateKey, err := eth.GetBatcherEVMPrivateKey(secretsManager, chainID)
	if err != nil {
		return nil, err
	}

	return &EVMChainOperationsTestMode{
		config:     config,
		privateKey: privateKey,
		db:         db,
		logger:     logger,
		testMode:   testMode,
	}, nil
}

// GenerateBatchTransaction implements core.ChainOperations.
func (cco *EVMChainOperationsTestMode) GenerateBatchTransaction(
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

	txs := newEVMSmartContractTransactionTestMode(
		batchNonceID, blockRounded+cco.config.TTLBlockNumberInc, confirmedTransactions, cco.testMode)

	txsBytes, err := txs.Pack()
	if err != nil {
		return nil, err
	}

	txsHashBytes, err := common.Keccak256(txsBytes)
	if err != nil {
		return nil, err
	}

	return &core.GeneratedBatchTxData{
		TxRaw:  txsBytes,
		TxHash: hex.EncodeToString(txsHashBytes),
	}, nil
}

// SignBatchTransaction implements core.ChainOperations.
func (cco *EVMChainOperationsTestMode) SignBatchTransaction(txHash string) ([]byte, []byte, error) {
	txsHashBytes, err := common.DecodeHex(txHash)
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

	return signatureBytes, nil, nil
}

func (cco *EVMChainOperationsTestMode) IsSynchronized(
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

func (cco *EVMChainOperationsTestMode) Submit(
	ctx context.Context, bridgeSmartContract eth.IBridgeSmartContract, batch eth.SignedBatch,
) error {
	return bridgeSmartContract.SubmitSignedBatchEVM(ctx, batch)
}

func newEVMSmartContractTransactionTestMode(
	batchNonceID uint64, ttl uint64, confirmedTransactions []eth.ConfirmedTransaction, testMode uint8,
) eth.EVMSmartContractTransaction {
	sourceAddrTxMap := map[string]eth.EVMSmartContractTransactionReceiver{}
	feeAmount := new(big.Int).SetUint64(0)

	for _, tx := range confirmedTransactions {
		for _, recv := range tx.Receivers {
			if recv.DestinationAddress == common.EthZeroAddr {
				feeAmount.Add(feeAmount, common.DfmToWei(recv.Amount))

				continue
			}

			val, exists := sourceAddrTxMap[recv.DestinationAddress]
			if !exists {
				val.Amount = common.DfmToWei(new(big.Int).Set(recv.Amount))
				val.Address = common.HexToAddress(recv.DestinationAddress)
			} else {
				val.Amount.Add(val.Amount, common.DfmToWei(recv.Amount))
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
		return bytes.Compare(receivers[i].Address[:], receivers[j].Address[:]) < 0
	})

	testTTL := testTTLSetup(ttl, testMode)

	return eth.EVMSmartContractTransaction{
		BatchNonceID: batchNonceID,
		TTL:          testTTL,
		FeeAmount:    feeAmount,
		Receivers:    receivers,
	}
}

// Test modes
// 1 - single batch fail
// 2 - 5 batches fail in a raw
// 3 - 5 bathces fail in "random" predetermined sequence
// 4 - random failure on different validators - disabled for now
func testTTLSetup(ttl uint64, testMode uint8) uint64 {
	fmt.Printf("Test mode: %d\n", testMode)

	switch testMode {
	case 1:
		if count > 0 {
			return ttl
		}

		count++

		return 0
	case 2:
		if count > 4 {
			return ttl
		}

		count++

		return 0
	case 3:
		if count > 4 {
			return ttl
		}

		count++

		return 0
	case 4:
		// // disabled
		// if count > 4 {
		// 	return ttl
		// }
		// // set the failure rat
		// failureRate := 50
		// if rand.Intn(100) > failureRate {
		// 	count++
		// 	return 0
		// }
		return ttl
	}

	return ttl
}
