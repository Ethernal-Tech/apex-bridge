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
	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
	"github.com/Ethernal-Tech/cardano-infrastructure/secrets"
	"github.com/hashicorp/go-hclog"
)

var (
	_ core.ChainOperations = (*EVMChainOperations)(nil)
)

type EVMChainOperations struct {
	config *cardano.EVMChainConfig
	db     indexer.Database
	logger hclog.Logger
}

func NewEVMChainOperations(
	jsonConfig json.RawMessage,
	db indexer.Database,
	secretsManager secrets.SecretsManager,
	chainID string,
	logger hclog.Logger,
) (*EVMChainOperations, error) {
	config, err := cardano.NewEVMChainConfig(jsonConfig)
	if err != nil {
		return nil, err
	}

	return &EVMChainOperations{
		config: config,
		db:     db,
		logger: logger,
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
	txs := newEVMSmartContractTransaction(chainID, confirmedTransactions)

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

	// TODO: calculate bls signature of txHash
	blsSignature := txsHashBytes

	return blsSignature, nil, nil
}

func (cco *EVMChainOperations) IsSynchronized(
	ctx context.Context, bridgeSmartContract eth.IBridgeSmartContract, chainID string,
) (bool, error) {
	return true, nil
}

func newEVMSmartContractTransaction(
	chainID string, confirmedTransactions []eth.ConfirmedTransaction,
) (result eth.EVMSmartContractTransaction) {
	mp := map[string]*big.Int{}

	for _, tx := range confirmedTransactions {
		for _, recv := range tx.Receivers {
			if val, exists := mp[recv.DestinationAddress]; exists {
				val.Add(val, recv.Amount)
			} else {
				mp[recv.DestinationAddress] = new(big.Int).Set(recv.Amount)
			}
		}
	}

	result.ChainID = common.ToNumChainID(chainID)
	result.Receivers = make([]eth.EVMSmartContractTransactionReceiver, 0, len(mp))

	for k, v := range mp {
		result.Receivers = append(result.Receivers, eth.EVMSmartContractTransactionReceiver{
			Address: common.HexToAddress(k),
			Amount:  v,
		})
	}

	// every batcher should have same order
	sort.Slice(result.Receivers, func(i, j int) bool {
		return bytes.Compare(result.Receivers[i].Address[:], result.Receivers[j].Address[:]) < 0
	})

	return result
}
