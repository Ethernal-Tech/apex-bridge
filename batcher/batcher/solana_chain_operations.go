package batcher

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"sort"
	"strings"
	"time"

	"github.com/Ethernal-Tech/apex-bridge/batcher/core"
	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/eth"
	solanatx "github.com/Ethernal-Tech/apex-bridge/solana"
	"github.com/Ethernal-Tech/cardano-infrastructure/secrets"
	"github.com/Ethernal-Tech/solana-infrastructure/sendtx"
	solanaTrackerStore "github.com/Ethernal-Tech/solana-infrastructure/tracker/store"
	"github.com/Ethernal-Tech/solana-infrastructure/wallet"
	"github.com/gagliardetto/solana-go"
	"github.com/hashicorp/go-hclog"
)

var (
	_ core.ChainOperations = (*SolanaChainOperations)(nil)
)

type SolanaChainOperations struct {
	config           *solanatx.SolanaChainConfig
	privateKey       solana.PrivateKey
	db               solanaTrackerStore.StorageHandler
	gasLimiter       eth.GasLimitHolder
	secretsManager   secrets.SecretsManager
	chainIDConverter *common.ChainIDConverter
	logger           hclog.Logger
}

func NewSolanaChainOperations(
	jsonConfig json.RawMessage,
	chainIDConverter *common.ChainIDConverter,
	db solanaTrackerStore.StorageHandler,
	secretsManager secrets.SecretsManager,
	destChainID string,
	logger hclog.Logger,
) (*SolanaChainOperations, error) {
	solanaConfig, err := solanatx.NewSolanaChainConfig(jsonConfig)
	if err != nil {
		return nil, err
	}

	solanaPrivateKey, err := solanatx.LoadBatcherSolanaPrivateKey(secretsManager, destChainID)
	if err != nil {
		return nil, err
	}

	return &SolanaChainOperations{
		config:           solanaConfig,
		privateKey:       *solanaPrivateKey,
		chainIDConverter: chainIDConverter,
		db:               db,
		gasLimiter: eth.NewGasLimitHolder(submitBatchMinGasLimit,
			submitBatchMaxGasLimit, submitBatchStepsGasLimit),
		secretsManager: secretsManager,
		logger:         logger,
	}, nil
}

func (sco *SolanaChainOperations) GenerateBatchTransaction(
	ctx context.Context, destinationChain string, confirmedTransactions []eth.ConfirmedTransaction, batchNonceID uint64,
) (*core.GeneratedBatchTxData, error) {
	receivers, err := sco.newSolanaReceivers(
		sco.config,
		confirmedTransactions,
		sco.config.MinFeeForBridging,
	)
	if err != nil {
		return nil, err
	}

	// get the slot number and try to fetch the block hash for that slot
	slotNumber, err := sco.getSlotNumber()
	if err != nil {
		return nil, err
	}

	blockHash, err := sco.getBlockhashBySlot(slotNumber)
	if err != nil {
		return nil, err
	}

	sco.logger.Debug("blockHash chosen for batch", "blockHash", blockHash, "slot", slotNumber)

	payload := sendtx.SolanaPayload{
		Blockhash: blockHash.String(),
		Receivers: receivers,
		BatchID:   batchNonceID,
	}

	payloadBytes, err := payload.Marshal()
	if err != nil {
		return nil, err
	}

	hash := sha256.Sum256(payloadBytes)
	payloadHash := hex.EncodeToString(hash[:])

	sco.logger.Debug("Batch payload data has been generated",
		"id", batchNonceID, "payload", payload, "hash", payloadHash,
		"slot", slotNumber)

	return &core.GeneratedBatchTxData{
		TxRaw:  payloadBytes,
		TxHash: payloadHash,
	}, nil
}

func (sco *SolanaChainOperations) SignBatchTransaction(
	generatedBatchData *core.GeneratedBatchTxData) (*core.BatchSignatures, error) {
	signature, err := sco.privateKey.Sign(generatedBatchData.TxRaw)
	if err != nil {
		return nil, err
	}

	signatureBytes := signature[:]

	if sco.logger.IsDebug() {
		sco.logger.Debug("Signature has been created",
			"signature", signature.String(),
			"public", sco.privateKey.PublicKey().String())
	}

	return &core.BatchSignatures{
		Multisig: signatureBytes,
	}, nil
}

func (sco *SolanaChainOperations) IsSynchronized(
	ctx context.Context, bridgeSmartContract eth.IBridgeSmartContract, chainID string,
) (bool, error) {
	lastObservedBlockBridge, err := bridgeSmartContract.GetLastObservedBlock(ctx, chainID)
	if err != nil {
		return false, err
	}

	lastOracleBlockPoint, err := sco.db.GetLatestBlockPoint()
	if err != nil {
		return false, err
	}

	return lastOracleBlockPoint != nil &&
		lastOracleBlockPoint.BlockSlot >= lastObservedBlockBridge.BlockSlot.Uint64(), nil
}

func (sco *SolanaChainOperations) Submit(
	ctx context.Context, bridgeSmartContract eth.IBridgeSmartContract, batch eth.SignedBatch) error {
	err := bridgeSmartContract.SubmitSignedBatchSolana(ctx, batch, sco.gasLimiter.GetGasLimit())

	sco.gasLimiter.Update(err)

	return err
}

func (sco *SolanaChainOperations) getBlockhashBySlot(slot uint64) (solana.Hash, error) {
	shouldWaitError := func(slot uint64, err error) bool {
		return strings.Contains(err.Error(), fmt.Sprintf("slot %d has not been processed yet", slot))
	}

	for {
		blockhash, err := sco.db.GetBlockhashBySlot(slot)
		if err != nil {
			if shouldWaitError(slot, err) {
				time.Sleep(400 * time.Millisecond)

				sco.logger.Debug("retrying to get blockhash", "slot", slot)

				continue
			}

			return solana.Hash{}, err
		}

		return blockhash, nil
	}
}

func (sco *SolanaChainOperations) getSlotNumber() (uint64, error) {
	latestBlockPoint, err := sco.db.GetLatestBlockPoint()
	if err != nil {
		return 0, err
	}

	newSlot, err := getNumberWithRoundingThresholdRoundDown(
		latestBlockPoint.BlockSlot, sco.config.SlotRoundingThreshold, sco.config.NoBatchPeriodPercent)
	if err != nil {
		return 0, err
	}

	sco.logger.Debug("calculate slotNumber with rounding", "slot", latestBlockPoint.BlockSlot, "newSlot", newSlot)

	return newSlot, nil
}

func (sco *SolanaChainOperations) newSolanaReceivers(
	config *solanatx.SolanaChainConfig,
	confirmedTransactions []eth.ConfirmedTransaction,
	minFeeForBridging *big.Int,
) ([]sendtx.BridgingTxReceiver, error) {
	sourceAddrTxMap := map[string][]sendtx.BridgingTxReceiver{}
	updateAmount := func(
		mp map[string][]sendtx.BridgingTxReceiver,
		addr string,
		tokenMint string,
		amount *big.Int,
	) error {
		var newEntry sendtx.BridgingTxReceiver

		val, exists := mp[addr]

		if !exists || len(val) == 0 {
			newEntry.Address = addr
			newEntry.TokenAmount = wallet.TokenAmount{
				Amount:    amount,
				TokenMint: tokenMint,
			}

			val = append(val, newEntry)
		} else {
			// check if there is a same token id first
			found := false

			for i, entry := range val {
				if entry.TokenAmount.TokenMint == tokenMint {
					val[i].TokenAmount.Amount.Add(val[i].TokenAmount.Amount, amount)

					found = true

					break
				}
			}

			if !found {
				newEntry.Address = addr

				newEntry.TokenAmount = wallet.TokenAmount{
					Amount:    amount,
					TokenMint: tokenMint,
				}

				val = append(val, newEntry)
			}
		}

		mp[addr] = val

		return nil
	}

	for _, tx := range confirmedTransactions {
		for _, recv := range tx.Receivers {
			amount := common.WeiToLamport(recv.Amount)
			tokenAmount := common.WeiToLamport(recv.AmountWrapped)

			if amount.Cmp(big.NewInt(0)) == 1 {
				tokenMint, err := config.GetTokenMint(recv.TokenId)
				if err != nil {
					return nil, err
				}

				err = updateAmount(sourceAddrTxMap, recv.DestinationAddress, tokenMint, amount.Sub(amount, minFeeForBridging))
				if err != nil {
					return nil, err
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

				tokenMint, err := config.GetTokenMint(realTokenID)
				if err != nil {
					return nil, err
				}

				err = updateAmount(sourceAddrTxMap, recv.DestinationAddress, tokenMint, tokenAmount)
				if err != nil {
					return nil, err
				}
			}
		}
	}

	receivers := make([]sendtx.BridgingTxReceiver, 0, len(sourceAddrTxMap))

	for _, v := range sourceAddrTxMap {
		receivers = append(receivers, v...)
	}

	// every batcher should have same order
	sort.SliceStable(receivers, func(i, j int) bool {
		if receivers[i].Address != receivers[j].Address {
			return receivers[i].Address < receivers[j].Address
		}

		return receivers[i].TokenAmount.TokenMint < receivers[j].TokenAmount.TokenMint
	})

	return receivers, nil
}
