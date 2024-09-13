package batcher

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
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
	_ core.ChainOperations = (*EVMChainOperations)(nil)
)

type TTLFormatterFunc func(ttl uint64, batchID uint64) uint64

type EVMChainOperations struct {
	config       *cardano.BatcherEVMChainConfig
	privateKey   *bn256.PrivateKey
	db           eventTrackerStore.EventTrackerStore
	ttlFormatter TTLFormatterFunc
	logger       hclog.Logger
}

func NewEVMChainOperations(
	jsonConfig json.RawMessage,
	secretsManager secrets.SecretsManager,
	db eventTrackerStore.EventTrackerStore,
	chainID string,
	logger hclog.Logger,
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
		ttlFormatter: getTTLFormatter(config.TestMode),
		logger:       logger,
	}, nil
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

	return &core.GeneratedBatchTxData{
		TxRaw:  txsBytes,
		TxHash: hex.EncodeToString(txsHashBytes),
	}, nil
}

// SignBatchTransaction implements core.ChainOperations.
func (cco *EVMChainOperations) SignBatchTransaction(txHash string) ([]byte, []byte, error) {
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
	return bridgeSmartContract.SubmitSignedBatchEVM(ctx, batch)
}

func newEVMSmartContractTransaction(
	batchNonceID uint64, ttl uint64, confirmedTransactions []eth.ConfirmedTransaction,
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

	return eth.EVMSmartContractTransaction{
		BatchNonceID: batchNonceID,
		TTL:          ttl,
		FeeAmount:    feeAmount,
		Receivers:    receivers,
	}
}

// getTTLFormatter returns formater for a test mode. By default it is just identity function
// 1 - first batch will fail
// 2 - first five batches will fail
// 3 - First batch 5 bathces fail in "random" predetermined sequence
func getTTLFormatter(testMode uint8) TTLFormatterFunc {
	switch testMode {
	default:
		return func(ttl, batchID uint64) uint64 {
			return ttl
		}
	case 1:
		return func(ttl, batchID uint64) uint64 {
			if batchID > 1 {
				return ttl
			}

			return 0
		}
	case 2:
		return func(ttl, batchID uint64) uint64 {
			if batchID > 5 {
				return ttl
			}

			return 0
		}
	case 3:
		return func(ttl, batchID uint64) uint64 {
			if batchID%2 == 0 && batchID <= 10 {
				return 0
			}

			return ttl
		}
	}
}
