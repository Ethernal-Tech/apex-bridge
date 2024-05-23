package core

import (
	"fmt"
	"strings"

	"github.com/Ethernal-Tech/apex-bridge/contractbinding"
	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
)

type BridgeClaimsBlockInfo struct {
	ChainID string
	Slot    uint64
	Hash    string
}

type ContractClaims = contractbinding.IBridgeStructsValidatorClaims
type BridgingRequestClaim = contractbinding.IBridgeStructsBridgingRequestClaim
type BatchExecutedClaim = contractbinding.IBridgeStructsBatchExecutedClaim
type BatchExecutionFailedClaim = contractbinding.IBridgeStructsBatchExecutionFailedClaim
type RefundRequestClaim = contractbinding.IBridgeStructsRefundRequestClaim
type RefundExecutedClaim = contractbinding.IBridgeStructsRefundExecutedClaim
type UTXO = contractbinding.IBridgeStructsUTXO
type UTXOs = contractbinding.IBridgeStructsUTXOs
type BridgingRequestReceiver = contractbinding.IBridgeStructsReceiver

type BridgeClaims struct {
	ContractClaims
}

func (bi *BridgeClaimsBlockInfo) EqualWithUnprocessed(tx *CardanoTx) bool {
	return bi.ChainID == tx.OriginChainID && bi.Slot == tx.BlockSlot && bi.Hash == tx.BlockHash
}

func (bi *BridgeClaimsBlockInfo) EqualWithProcessed(tx *ProcessedCardanoTx) bool {
	return bi.ChainID == tx.OriginChainID && bi.Slot == tx.BlockSlot && bi.Hash == tx.BlockHash
}

func (bi *BridgeClaimsBlockInfo) EqualWithExpected(tx *BridgeExpectedCardanoTx, block *indexer.CardanoBlock) bool {
	return bi.ChainID == tx.ChainID && bi.Slot == block.Slot && bi.Hash == block.Hash
}

func (bc *BridgeClaims) Count() int {
	return len(bc.BridgingRequestClaims) +
		len(bc.BatchExecutedClaims) +
		len(bc.BatchExecutionFailedClaims) /* + len(bc.RefundRequest) + len(bc.RefundExecuted)*/
}

func (bc *BridgeClaims) Any() bool {
	return bc.Count() > 0
}

func RefundExecutedClaimString(c RefundExecutedClaim) string {
	var sb strings.Builder

	sb.WriteString("ObservedTransactionHash = ")
	sb.WriteString(c.ObservedTransactionHash)
	sb.WriteString("\nChainID = ")
	sb.WriteString(c.ChainID)
	sb.WriteString("\nRefundTxHash = ")
	sb.WriteString(c.RefundTxHash)
	sb.WriteString("\nUtxo = ")
	sb.WriteString(fmt.Sprintf("{ Nonce = %v, TxHash = %s, TxIndex = %v, Amount = %v }",
		c.Utxo.Nonce, c.Utxo.TxHash, c.Utxo.TxIndex, c.Utxo.Amount))

	return sb.String()
}

func RefundRequestClaimString(c RefundRequestClaim) string {
	var sb strings.Builder

	sb.WriteString("ObservedTransactionHash = ")
	sb.WriteString(c.ObservedTransactionHash)
	sb.WriteString("\nPreviousRefundTxHash = ")
	sb.WriteString(c.PreviousRefundTxHash)
	sb.WriteString("\nChainID = ")
	sb.WriteString(c.ChainID)
	sb.WriteString("\nReceiver = ")
	sb.WriteString(c.Receiver)
	sb.WriteString("\nUtxo = ")
	sb.WriteString(fmt.Sprintf("{ Nonce = %v, TxHash = %s, TxIndex = %v, Amount = %v }",
		c.Utxo.Nonce, c.Utxo.TxHash, c.Utxo.TxIndex, c.Utxo.Amount))
	sb.WriteString("\nRawTransaction = ")
	sb.WriteString(c.RawTransaction)
	sb.WriteString("\nMultisigSignature = ")
	sb.WriteString(c.MultisigSignature)
	sb.WriteString("\nRetryCounter = ")
	sb.WriteString(c.RetryCounter.String())

	return sb.String()
}

func BatchExecutionFailedClaimString(c BatchExecutionFailedClaim) string {
	var sb strings.Builder

	sb.WriteString("ObservedTransactionHash = ")
	sb.WriteString(c.ObservedTransactionHash)
	sb.WriteString("\nChainID = ")
	sb.WriteString(c.ChainID)
	sb.WriteString("\nBatchNonceID = ")
	sb.WriteString(c.BatchNonceID.String())

	return sb.String()
}

func BatchExecutedClaimString(c BatchExecutedClaim) string {
	var sb strings.Builder

	sb.WriteString("ObservedTransactionHash = ")
	sb.WriteString(c.ObservedTransactionHash)
	sb.WriteString("\nChainID = ")
	sb.WriteString(c.ChainID)
	sb.WriteString("\nBatchNonceID = ")
	sb.WriteString(c.BatchNonceID.String())
	sb.WriteString("\nMultisigOwnedUTXOs = [")

	for _, utxo := range c.OutputUTXOs.MultisigOwnedUTXOs {
		sb.WriteString(fmt.Sprintf("{ Nonce = %v, TxHash = %s, TxIndex = %v, Amount = %v }",
			utxo.Nonce, utxo.TxHash, utxo.TxIndex, utxo.Amount))
	}

	sb.WriteString("]")

	sb.WriteString("\nFeePayerOwnedUTXOs = [")

	for _, utxo := range c.OutputUTXOs.FeePayerOwnedUTXOs {
		sb.WriteString(fmt.Sprintf("{ Nonce = %v, TxHash = %s, TxIndex = %v, Amount = %v }",
			utxo.Nonce, utxo.TxHash, utxo.TxIndex, utxo.Amount))
	}

	sb.WriteString("]")

	return sb.String()
}

func BridgingRequestClaimString(c BridgingRequestClaim) string {
	var (
		sb          strings.Builder
		sbReceivers strings.Builder
	)

	for _, r := range c.Receivers {
		if sbReceivers.Len() > 0 {
			sbReceivers.WriteString(", ")
		}

		sbReceivers.WriteString(fmt.Sprintf("{ DestinationAddress = %s, Amount = %v }",
			r.DestinationAddress, r.Amount))
	}

	sb.WriteString("ObservedTransactionHash = ")
	sb.WriteString(c.ObservedTransactionHash)
	sb.WriteString("\nReceivers = [")
	sb.WriteString(sbReceivers.String())
	sb.WriteString("]")
	sb.WriteString("\nOutputUTXO = ")
	sb.WriteString(fmt.Sprintf("{ Nonce = %v, TxHash = %s, TxIndex = %v, Amount = %v }",
		c.OutputUTXO.Nonce, c.OutputUTXO.TxHash, c.OutputUTXO.TxIndex, c.OutputUTXO.Amount))
	sb.WriteString("\nSourceChainID = ")
	sb.WriteString(c.SourceChainID)
	sb.WriteString("\nDestinationChainID = ")
	sb.WriteString(c.DestinationChainID)

	return sb.String()
}

func (bc BridgeClaims) String() string {
	var (
		sb     strings.Builder
		sbBRC  strings.Builder
		sbBEC  strings.Builder
		sbBEFC strings.Builder
		sbRRC  strings.Builder
		sbREC  strings.Builder
	)

	for _, brc := range bc.BridgingRequestClaims {
		if sbBRC.Len() > 0 {
			sbBRC.WriteString(",\n")
		}

		sbBRC.WriteString("{ ")
		sbBRC.WriteString(BridgingRequestClaimString(brc))
		sbBRC.WriteString(" }")
	}

	for _, bec := range bc.BatchExecutedClaims {
		if sbBEC.Len() > 0 {
			sbBEC.WriteString(",\n")
		}

		sbBEC.WriteString("{ ")
		sbBEC.WriteString(BatchExecutedClaimString(bec))
		sbBEC.WriteString(" }")
	}

	for _, befc := range bc.BatchExecutionFailedClaims {
		if sbBEFC.Len() > 0 {
			sbBEFC.WriteString(",\n")
		}

		sbBEFC.WriteString("{ ")
		sbBEFC.WriteString(BatchExecutionFailedClaimString(befc))
		sbBEFC.WriteString(" }")
	}

	for _, rrc := range bc.RefundRequestClaims {
		if sbRRC.Len() > 0 {
			sbRRC.WriteString(",\n")
		}

		sbRRC.WriteString("{ ")
		sbRRC.WriteString(RefundRequestClaimString(rrc))
		sbRRC.WriteString(" }")
	}

	for _, rec := range bc.RefundExecutedClaims {
		if sbREC.Len() > 0 {
			sbREC.WriteString(",\n")
		}

		sbREC.WriteString("{ ")
		sbREC.WriteString(RefundExecutedClaimString(rec))
		sbREC.WriteString(" }")
	}

	sb.WriteString("BridgingRequestClaims = \n[")
	sb.WriteString(sbBRC.String())
	sb.WriteString("]")

	sb.WriteString("\nBatchExecutedClaims = \n[")
	sb.WriteString(sbBEC.String())
	sb.WriteString("]")

	sb.WriteString("\nBatchExecutionFailedClaims = \n[")
	sb.WriteString(sbBEFC.String())
	sb.WriteString("]")

	sb.WriteString("\nRefundRequestClaims = \n[")
	sb.WriteString(sbRRC.String())
	sb.WriteString("]")

	sb.WriteString("\nRefundExecutedClaims = \n[")
	sb.WriteString(sbREC.String())
	sb.WriteString("]")

	return sb.String()
}
