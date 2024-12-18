package core

import (
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/contractbinding"
)

type ClaimType = string

const (
	BRCClaimType ClaimType = "BRC"
)

type ContractClaims = contractbinding.IBridgeStructsValidatorClaims
type BridgingRequestClaim = contractbinding.IBridgeStructsBridgingRequestClaim
type BatchExecutedClaim = contractbinding.IBridgeStructsBatchExecutedClaim
type BatchExecutionFailedClaim = contractbinding.IBridgeStructsBatchExecutionFailedClaim
type RefundRequestClaim = contractbinding.IBridgeStructsRefundRequestClaim
type RefundExecutedClaim = contractbinding.IBridgeStructsRefundExecutedClaim
type BridgingRequestReceiver = contractbinding.IBridgeStructsReceiver
type HotWalletIncrementClaim = contractbinding.IBridgeStructsHotWalletIncrementClaim

type BridgeClaims struct {
	ContractClaims
}

func (bc *BridgeClaims) Count() int {
	return len(bc.BridgingRequestClaims) +
		len(bc.BatchExecutedClaims) +
		len(bc.BatchExecutionFailedClaims) +
		len(bc.HotWalletIncrementClaims) /* + len(bc.RefundRequest) + len(bc.RefundExecuted)*/
}

func (bc *BridgeClaims) Any() bool {
	return bc.Count() > 0
}

func (bc *BridgeClaims) CanAddMore(maxAmount int) bool {
	return bc.Count() < maxAmount
}

func RefundExecutedClaimString(c RefundExecutedClaim) string {
	var sb strings.Builder

	sb.WriteString("ObservedTransactionHash = ")
	sb.WriteString(hex.EncodeToString(c.ObservedTransactionHash[:]))
	sb.WriteString("\nChainID = ")
	sb.WriteString(common.ToStrChainID(c.ChainId))
	sb.WriteString("\nRefundTxHash = ")
	sb.WriteString(hex.EncodeToString(c.RefundTxHash[:]))

	return sb.String()
}

func RefundRequestClaimString(c RefundRequestClaim) string {
	var sb strings.Builder

	sb.WriteString("ObservedTransactionHash = ")
	sb.WriteString(hex.EncodeToString(c.ObservedTransactionHash[:]))
	sb.WriteString("\nPreviousRefundTxHash = ")
	sb.WriteString(hex.EncodeToString(c.PreviousRefundTxHash[:]))
	sb.WriteString("\nChainID = ")
	sb.WriteString(common.ToStrChainID(c.ChainId))
	sb.WriteString("\nReceiver = ")
	sb.WriteString(c.Receiver)
	sb.WriteString("\nRawTransaction = ")
	sb.WriteString(hex.EncodeToString(c.RawTransaction))
	sb.WriteString("\nSignature = ")
	sb.WriteString(hex.EncodeToString(c.Signature))
	sb.WriteString("\nRetryCounter = ")
	sb.WriteString(fmt.Sprint(c.RetryCounter))

	return sb.String()
}

func BatchExecutionFailedClaimString(c BatchExecutionFailedClaim) string {
	var sb strings.Builder

	sb.WriteString("ObservedTransactionHash = ")
	sb.WriteString(hex.EncodeToString(c.ObservedTransactionHash[:]))
	sb.WriteString("\nChainID = ")
	sb.WriteString(common.ToStrChainID(c.ChainId))
	sb.WriteString("\nBatchNonceID = ")
	sb.WriteString(fmt.Sprint(c.BatchNonceId))

	return sb.String()
}

func BatchExecutedClaimString(c BatchExecutedClaim) string {
	var sb strings.Builder

	sb.WriteString("ObservedTransactionHash = ")
	sb.WriteString(hex.EncodeToString(c.ObservedTransactionHash[:]))
	sb.WriteString("\nChainID = ")
	sb.WriteString(common.ToStrChainID(c.ChainId))
	sb.WriteString("\nBatchNonceID = ")
	sb.WriteString(fmt.Sprint(c.BatchNonceId))

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
	sb.WriteString(hex.EncodeToString(c.ObservedTransactionHash[:]))
	sb.WriteString("\nRetryCounter = ")

	if c.RetryCounter == nil {
		sb.WriteString("nil")
	} else {
		sb.WriteString(c.RetryCounter.String())
	}

	sb.WriteString("\nReceivers = [")
	sb.WriteString(sbReceivers.String())
	sb.WriteString("]")
	sb.WriteString("\nNativeCurrencyAmountDestination = ")
	sb.WriteString(c.NativeCurrencyAmountDestination.String())
	sb.WriteString("\nWrappedTokenAmountDestination = ")
	sb.WriteString(c.WrappedTokenAmountDestination.String())
	sb.WriteString("\nSourceChainID = ")
	sb.WriteString(common.ToStrChainID(c.SourceChainId))
	sb.WriteString("\nDestinationChainID = ")
	sb.WriteString(common.ToStrChainID(c.DestinationChainId))

	return sb.String()
}

func HotWalletIncrementClaimsString(c HotWalletIncrementClaim) string {
	if !c.IsIncrement {
		return fmt.Sprintf("(%s, -%s)", common.ToStrChainID(c.ChainId), c.Amount)
	}

	return fmt.Sprintf("(%s, %s)", common.ToStrChainID(c.ChainId), c.Amount)
}

func (bc BridgeClaims) String() string {
	var (
		sb     strings.Builder
		sbBRC  strings.Builder
		sbBEC  strings.Builder
		sbBEFC strings.Builder
		sbRRC  strings.Builder
		sbREC  strings.Builder
		sbHWIC strings.Builder
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

	for _, rec := range bc.HotWalletIncrementClaims {
		if sbHWIC.Len() > 0 {
			sbHWIC.WriteString(", ")
		}

		sbHWIC.WriteString(HotWalletIncrementClaimsString(rec))
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

	sb.WriteString("\nHotWalletIncrementClaims = [")
	sb.WriteString(sbHWIC.String())
	sb.WriteString("]")

	return sb.String()
}
