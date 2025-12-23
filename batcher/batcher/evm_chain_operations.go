package batcher

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"math/big"
	"sort"

	"github.com/Ethernal-Tech/apex-bridge/batcher/core"
	cardano "github.com/Ethernal-Tech/apex-bridge/cardano"
	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/eth"
	"github.com/Ethernal-Tech/apex-bridge/testenv"
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
		ttlFormatter: testenv.GetTTLFormatter(config.TestMode),
		gasLimiter:   eth.NewGasLimitHolder(submitBatchMinGasLimit, submitBatchMaxGasLimit, submitBatchStepsGasLimit),
		logger:       logger,
	}, nil
}

// GenerateBatchTransaction implements core.ChainOperations.
func (cco *EVMChainOperations) GenerateBatchTransaction(
	ctx context.Context,
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

	txs, err := newEVMSmartContractTransaction(
		cco.config,
		batchNonceID,
		cco.ttlFormatter(blockRounded+cco.config.TTLBlockNumberInc, batchNonceID),
		confirmedTransactions,
		common.DfmToWei(new(big.Int).SetUint64(cco.config.MinFeeForBridging)))
	if err != nil {
		return nil, err
	}

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
	generatedBatchData *core.GeneratedBatchTxData) (*core.BatchSignatures, error) {
	txsHashBytes, err := common.DecodeHex(generatedBatchData.TxHash)
	if err != nil {
		return nil, err
	}

	signature, err := cco.privateKey.Sign(txsHashBytes, eth.BN256Domain)
	if err != nil {
		return nil, err
	}

	signatureBytes, err := signature.Marshal()
	if err != nil {
		return nil, err
	}

	if cco.logger.IsDebug() {
		cco.logger.Debug("Signature has been created",
			"signature", hex.EncodeToString(signatureBytes),
			"public", hex.EncodeToString(cco.privateKey.PublicKey().Marshal()))
	}

	return &core.BatchSignatures{
		Multisig: signatureBytes,
	}, nil
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
	config *cardano.BatcherEVMChainConfig,
	batchNonceID uint64,
	ttl uint64,
	confirmedTransactions []eth.ConfirmedTransaction,
	minFeeForBridging *big.Int,
) (*eth.EVMSmartContractTransaction, error) {
	sourceAddrTxMap := map[string][]eth.EVMSmartContractTransactionReceiver{}
	feeAmount := big.NewInt(0)

	currencyID, err := config.GetCurrencyID()
	if err != nil {
		return nil, err
	}

	updateAmount := func(
		mp map[string][]eth.EVMSmartContractTransactionReceiver,
		addr string,
		tokenID uint16,
		amount *big.Int,
	) {
		var newEntry eth.EVMSmartContractTransactionReceiver
		val, exists := mp[addr]

		if !exists || len(val) == 0 {
			newEntry.Amount = amount
			newEntry.Address = common.HexToAddress(addr)
			newEntry.TokenID = tokenID

			val = append(val, newEntry)
		} else {
			// check if there is a same token id first
			found := false

			for i, entry := range val {
				if entry.TokenID == tokenID {
					val[i].Amount.Add(val[i].Amount, amount)

					found = true

					break
				}
			}

			if !found {
				newEntry.Amount = amount
				newEntry.Address = common.HexToAddress(addr)
				newEntry.TokenID = tokenID

				val = append(val, newEntry)
			}
		}

		mp[addr] = val
	}

	for _, tx := range confirmedTransactions {
		for _, recv := range tx.Receivers {
			amount := common.DfmToWei(recv.Amount)
			tokenAmount := common.DfmToWei(recv.AmountWrapped)

			if recv.DestinationAddress == common.EthZeroAddr {
				feeAmount.Add(feeAmount, amount)

				continue
			}

			if amount.Cmp(big.NewInt(0)) == 1 {
				// In case a transaction is of type refund, batcher should transfer minFeeForBridging
				// to fee payer address, and the rest is transferred to the user.
				if tx.TransactionType == uint8(common.RefundConfirmedTxType) {
					feeAmount.Add(feeAmount, minFeeForBridging)
					updateAmount(sourceAddrTxMap, recv.DestinationAddress, currencyID, amount.Sub(amount, minFeeForBridging))
				} else {
					updateAmount(sourceAddrTxMap, recv.DestinationAddress, currencyID, amount)
				}
			}

			if tokenAmount.Cmp(big.NewInt(0)) == 1 {
				var realTokenID = recv.TokenId

				// when defunding, sc doesn't know the correct tokenId of the wrapped token on this chain
				// also for backward compatibility during the process of syncing -
				// rebuilding confirmedTx.Receivers from confirmedTx.receivers
				if recv.TokenId == 0 {
					wrappedTokenID, err := config.GetWrappedTokenID()
					if err != nil {
						return nil, err
					}

					realTokenID = wrappedTokenID
				}

				updateAmount(sourceAddrTxMap, recv.DestinationAddress, realTokenID, tokenAmount)
			}
		}
	}

	receivers := make([]eth.EVMSmartContractTransactionReceiver, 0, len(sourceAddrTxMap))

	for _, v := range sourceAddrTxMap {
		receivers = append(receivers, v...)
	}

	// every batcher should have same order
	sort.SliceStable(receivers, func(i, j int) bool {
		if receivers[i].Address.Cmp(receivers[j].Address) != 0 {
			return receivers[i].Address.Cmp(receivers[j].Address) < 0
		}

		return receivers[i].TokenID < receivers[j].TokenID
	})

	return &eth.EVMSmartContractTransaction{
		BatchNonceID: batchNonceID,
		TTL:          ttl,
		FeeAmount:    feeAmount,
		Receivers:    receivers,
	}, nil
}
