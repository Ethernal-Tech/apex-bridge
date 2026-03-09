package batcher

import (
	"context"
	"encoding/json"
	"math/big"
	"sort"

	"github.com/Ethernal-Tech/apex-bridge/batcher/core"
	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/eth"
	solana "github.com/Ethernal-Tech/apex-bridge/solana"
	"github.com/Ethernal-Tech/cardano-infrastructure/secrets"
	"github.com/Ethernal-Tech/solana-infrastructure/sendtx"
	solanaTrackerStore "github.com/Ethernal-Tech/solana-infrastructure/tracker/store"
	"github.com/Ethernal-Tech/solana-infrastructure/wallet"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/hashicorp/go-hclog"
	"github.com/mr-tron/base58"
)

var (
	_ core.ChainOperations = (*SolanaChainOperations)(nil)
)

type SolanaChainOperations struct {
	config           *solana.SolanaChainConfig
	wallet           solana.ApexSolanaWallet
	solanaProgramKey *wallet.Wallet
	db               solanaTrackerStore.BoltStorageHandler
	gasLimiter       eth.GasLimitHolder
	secretsManager   secrets.SecretsManager
	chainIDConverter *common.ChainIDConverter
	logger           hclog.Logger
}

func NewSolanaChainOperations(
	jsonConfig json.RawMessage,
	chainIDConverter *common.ChainIDConverter,
	db solanaTrackerStore.BoltStorageHandler,
	secretsManager secrets.SecretsManager,
	destChainID string,
	logger hclog.Logger,
) (*SolanaChainOperations, error) {
	solanaConfig, err := solana.NewSolanaChainConfig(jsonConfig)
	if err != nil {
		return nil, err
	}

	solanaWallet, err := solana.LoadWallet(secretsManager, destChainID)
	if err != nil {
		return nil, err
	}

	solanaProgramKey, err := solana.LoadSolanaProgramKeyPair(secretsManager)
	if err != nil {
		return nil, err
	}

	return &SolanaChainOperations{
		config:           solanaConfig,
		wallet:           *solanaWallet,
		solanaProgramKey: solanaProgramKey,
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
	txProvider, err := wallet.NewProvider(rpc.LocalNet_WS)
	if err != nil {
		return nil, err
	}

	chainConfig := sendtx.ChainConfig{
		MinAmountToBridge: sco.config.MinFeeForBridging.Uint64(),
	}

	txSender, err := sendtx.NewTxSender(txProvider, chainConfig, sco.solanaProgramKey.PrivateKey)
	if err != nil {
		return nil, err
	}

	bridgingTxs, err := sco.newSolanaSmartContractTransaction(
		sco.config,
		destinationChain,
		batchNonceID,
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

	blockHash, err := sco.db.GetBlockhashBySlot(slotNumber)
	if err != nil {
		return nil, err
	}

	tx, err := txSender.CreateTx(
		ctx, *sco.wallet.BridgeWallet, sendtx.InstructionTypeBridgeTransaction, blockHash, bridgingTxs)
	if err != nil {
		return nil, err
	}

	txHash := tx.Signatures[0].String()

	rawTx, err := tx.MarshalBinary()
	if err != nil {
		return nil, err
	}

	sco.logger.Debug("Batch transaction data has been generated",
		"id", batchNonceID, "tx", bridgingTxs, "hash", txHash,
		"slot", slotNumber)

	return &core.GeneratedBatchTxData{
		TxRaw:  rawTx,
		TxHash: txHash,
	}, nil
}

func (sco *SolanaChainOperations) SignBatchTransaction(
	generatedBatchData *core.GeneratedBatchTxData) (*core.BatchSignatures, error) {
	bridgeWallet := sco.wallet.BridgeWallet

	hashBytes, err := base58.Decode(generatedBatchData.TxHash)
	if err != nil {
		return nil, err
	}

	signature, err := bridgeWallet.Sign(hashBytes)
	if err != nil {
		return nil, err
	}

	signatureBytes, err := bridgeWallet.MarshalSignature(*signature)
	if err != nil {
		return nil, err
	}

	if sco.logger.IsDebug() {
		sco.logger.Debug("Signature has been created",
			"signature", signature.String(),
			"public", bridgeWallet.PublicKey.String())
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
	err := bridgeSmartContract.SubmitSignedBatch(ctx, batch, sco.gasLimiter.GetGasLimit())

	sco.gasLimiter.Update(err)

	return err
}

func (sco *SolanaChainOperations) getSlotNumber() (uint64, error) {
	slot, err := sco.db.ReadSlot()
	if err != nil {
		return 0, err
	}

	newSlot, err := getNumberWithRoundingThreshold(
		slot, sco.config.SlotRoundingThreshold, sco.config.NoBatchPeriodPercent)
	if err != nil {
		return 0, err
	}

	sco.logger.Debug("calculate slotNumber with rounding", "slot", slot, "newSlot", newSlot)

	return newSlot, nil
}

func (sco *SolanaChainOperations) newSolanaSmartContractTransaction(
	config *solana.SolanaChainConfig,
	destChainID string,
	batchNonceID uint64,
	confirmedTransactions []eth.ConfirmedTransaction,
	minFeeForBridging *big.Int,
) (*sendtx.BridgeTransactionDto, error) {
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
			amount := new(big.Int).Set(recv.Amount)
			tokenAmount := new(big.Int).Set(recv.AmountWrapped)

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

	return &sendtx.BridgeTransactionDto{
		SrcChainID: 0,
		DstChainID: sco.chainIDConverter.StrToInt[destChainID],
		SenderAddr: sco.wallet.BridgeWallet.PrivateKey.String(),
		Receivers:  receivers,
		BatchID:    batchNonceID,
	}, nil
}
