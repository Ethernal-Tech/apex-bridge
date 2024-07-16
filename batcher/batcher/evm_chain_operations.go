package batcher

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"math/big"
	"sort"

	"github.com/Ethernal-Tech/apex-bridge/batcher/core"
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

type EVMChainOperations struct {
	privateKey *bn256.PrivateKey
	db         eventTrackerStore.EventTrackerStore
	logger     hclog.Logger
}

func NewEVMChainOperations(
	secretsManager secrets.SecretsManager,
	db eventTrackerStore.EventTrackerStore,
	chainID string,
	logger hclog.Logger,
) (*EVMChainOperations, error) {
	privateKey, err := eth.GetValidatorBLSPrivateKey(secretsManager, chainID)
	if err != nil {
		return nil, err
	}

	return &EVMChainOperations{
		privateKey: privateKey,
		db:         db,
		logger:     logger,
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
	txs := newEVMSmartContractTransaction(batchNonceID, confirmedTransactions)

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
	txsHashBytes, err := hex.DecodeString(txHash)
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
	batchNonceID uint64, confirmedTransactions []eth.ConfirmedTransaction,
) eth.EVMSmartContractTransaction {
	sourceAddrTxMap := map[string]eth.EVMSmartContractTransactionReceiver{}

	for _, tx := range confirmedTransactions {
		for _, recv := range tx.Receivers {
			key := fmt.Sprintf("%d_%s", tx.SourceChainId, recv.DestinationAddress)

			val, exists := sourceAddrTxMap[key]
			if !exists {
				val.Amount = new(big.Int).Set(recv.Amount)
				val.Address = common.HexToAddress(recv.DestinationAddress)
				val.SourceID = tx.SourceChainId
			} else {
				val.Amount.Add(val.Amount, recv.Amount)
			}

			sourceAddrTxMap[key] = val
		}
	}

	receivers := make([]eth.EVMSmartContractTransactionReceiver, 0, len(sourceAddrTxMap))

	for _, v := range sourceAddrTxMap {
		receivers = append(receivers, v)
	}

	// every batcher should have same order
	sort.Slice(receivers, func(i, j int) bool {
		if receivers[i].SourceID == receivers[j].SourceID {
			return bytes.Compare(receivers[i].Address[:], receivers[j].Address[:]) < 0
		}

		return receivers[i].SourceID < receivers[j].SourceID
	})

	return eth.EVMSmartContractTransaction{
		BatchNonceID: batchNonceID,
		Receivers:    receivers,
	}
}
