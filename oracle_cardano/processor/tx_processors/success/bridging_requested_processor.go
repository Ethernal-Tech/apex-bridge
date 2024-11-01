package successtxprocessors

import (
	"fmt"
	"math/big"
	"strings"

	cardanotx "github.com/Ethernal-Tech/apex-bridge/cardano"
	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/oracle_cardano/core"
	"github.com/Ethernal-Tech/apex-bridge/oracle_cardano/utils"
	cCore "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	cUtils "github.com/Ethernal-Tech/apex-bridge/oracle_common/utils"
	goEthCommon "github.com/ethereum/go-ethereum/common"
	"github.com/hashicorp/go-hclog"
)

var _ core.CardanoTxSuccessProcessor = (*BridgingRequestedProcessorImpl)(nil)

type BridgingRequestedProcessorImpl struct {
	logger hclog.Logger
}

func NewBridgingRequestedProcessor(logger hclog.Logger) *BridgingRequestedProcessorImpl {
	return &BridgingRequestedProcessorImpl{
		logger: logger.Named("bridging_requested_processor"),
	}
}

func (*BridgingRequestedProcessorImpl) GetType() common.BridgingTxType {
	return common.BridgingTxTypeBridgingRequest
}

func (*BridgingRequestedProcessorImpl) PreValidate(tx *core.CardanoTx, appConfig *cCore.AppConfig) error {
	return nil
}

func (p *BridgingRequestedProcessorImpl) ValidateAndAddClaim(
	claims *cCore.BridgeClaims, tx *core.CardanoTx, appConfig *cCore.AppConfig,
) error {
	metadata, err := common.UnmarshalMetadata[common.BridgingRequestMetadata](common.MetadataEncodingTypeCbor, tx.Metadata)
	if err != nil {
		return fmt.Errorf("failed to unmarshal metadata: tx: %v, err: %w", tx, err)
	}

	if metadata.BridgingTxType != p.GetType() {
		return fmt.Errorf("ValidateAndAddClaim called for irrelevant tx: %v", tx)
	}

	p.logger.Debug("Validating relevant tx", "txHash", tx.Hash, "metadata", metadata)

	err = p.validate(tx, metadata, appConfig)
	if err == nil {
		p.addBridgingRequestClaim(claims, tx, metadata, appConfig)
	} else {
		//nolint:godox
		// TODO: Refund
		// p.addRefundRequestClaim(claims, tx, metadata)
		return fmt.Errorf("validation failed for tx: %v, err: %w", tx, err)
	}

	return nil
}

func (p *BridgingRequestedProcessorImpl) addBridgingRequestClaim(
	claims *cCore.BridgeClaims, tx *core.CardanoTx,
	metadata *common.BridgingRequestMetadata, appConfig *cCore.AppConfig,
) {
	totalAmount := big.NewInt(0)

	var feeAddress string

	cardanoDestConfig, ethDestConfig := cUtils.GetChainConfig(appConfig, metadata.DestinationChainID)

	switch {
	case cardanoDestConfig != nil:
		feeAddress = cardanoDestConfig.BridgingAddresses.FeeAddress
	case ethDestConfig != nil:
		feeAddress = common.EthZeroAddr
	default:
		p.logger.Warn("Added BridgingRequestClaim not supported chain", "chainId", metadata.DestinationChainID)

		return
	}

	receivers := make([]cCore.BridgingRequestReceiver, 0, len(metadata.Transactions))

	for _, receiver := range metadata.Transactions {
		receiverAddr := strings.Join(receiver.Address, "")

		if receiverAddr == feeAddress {
			// fee address will be added at the end
			continue
		}

		receiverAmount := new(big.Int).SetUint64(receiver.Amount)
		receivers = append(receivers, cCore.BridgingRequestReceiver{
			DestinationAddress: receiverAddr,
			Amount:             receiverAmount,
		})

		totalAmount.Add(totalAmount, receiverAmount)
	}

	feeAmount := new(big.Int).SetUint64(metadata.FeeAmount)

	totalAmount.Add(totalAmount, feeAmount)

	receivers = append(receivers, cCore.BridgingRequestReceiver{
		DestinationAddress: feeAddress,
		Amount:             feeAmount,
	})

	claim := cCore.BridgingRequestClaim{
		ObservedTransactionHash: tx.Hash,
		SourceChainId:           common.ToNumChainID(tx.OriginChainID),
		DestinationChainId:      common.ToNumChainID(metadata.DestinationChainID),
		Receivers:               receivers,
		TotalAmount:             totalAmount,
		RetryCounter:            big.NewInt(int64(tx.TryCount)),
	}

	claims.BridgingRequestClaims = append(claims.BridgingRequestClaims, claim)

	p.logger.Info("Added BridgingRequestClaim",
		"txHash", tx.Hash, "metadata", metadata, "claim", cCore.BridgingRequestClaimString(claim))
}

/*
func (*BridgingRequestedProcessorImpl) addRefundRequestClaim(
	claims *core.BridgeClaims, tx *core.CardanoTx, metadata *common.BridgingRequestMetadata,
) {

		var outputUtxos []core.Utxo
		for _, output := range tx.Outputs {
			outputUtxos = append(outputUtxos, core.Utxo{
				Address: output.Address,
				Amount:  output.Amount,
			})
		}

		// what goes into UtxoTransaction
		claim := core.RefundRequestClaim{
			TxHash:             tx.Hash,
			RetryCounter:       0,
			RefundToAddress:    metadata.SenderAddr,
			DestinationChainId: metadata.DestinationChainId,
			OutputUtxos:        outputUtxos,
			UtxoTransaction:    core.UtxoTransaction{},
		}

		claims.RefundRequest = append(claims.RefundRequest, claim)

		p.logger.Info("Added RefundRequestClaim",
		"txHash", tx.Hash, "metadata", metadata, "claim", core.RefundRequestClaimString(claim))
}
*/

func (p *BridgingRequestedProcessorImpl) validate(
	tx *core.CardanoTx, metadata *common.BridgingRequestMetadata, appConfig *cCore.AppConfig,
) error {
	multisigUtxo, err := utils.ValidateTxOutputs(tx, appConfig, false)
	if err != nil {
		return err
	}

	cardanoSrcConfig, _ := cUtils.GetChainConfig(appConfig, tx.OriginChainID)
	if cardanoSrcConfig == nil {
		return fmt.Errorf("origin chain not registered: %v", tx.OriginChainID)
	}

	cardanoDestConfig, ethDestConfig := cUtils.GetChainConfig(appConfig, metadata.DestinationChainID)
	if cardanoDestConfig == nil && ethDestConfig == nil {
		return fmt.Errorf("destination chain not registered: %v", metadata.DestinationChainID)
	}

	if len(metadata.Transactions) > appConfig.BridgingSettings.MaxReceiversPerBridgingRequest {
		return fmt.Errorf("number of receivers in metadata greater than maximum allowed - no: %v, max: %v, metadata: %v",
			len(metadata.Transactions), appConfig.BridgingSettings.MaxReceiversPerBridgingRequest, metadata)
	}

	receiverAmountSum := big.NewInt(0)
	feeSum := uint64(0)
	foundAUtxoValueBelowMinimumValue := false
	foundAnInvalidReceiverAddr := false

	for _, receiver := range metadata.Transactions {
		receiverAddr := strings.Join(receiver.Address, "")

		if cardanoDestConfig != nil {
			if receiver.Amount < appConfig.BridgingSettings.UtxoMinValue {
				foundAUtxoValueBelowMinimumValue = true

				break
			}

			if !cardanotx.IsValidOutputAddress(receiverAddr, cardanoDestConfig.NetworkID) {
				foundAnInvalidReceiverAddr = true

				break
			}

			if receiverAddr == cardanoDestConfig.BridgingAddresses.FeeAddress {
				feeSum += receiver.Amount
			} else {
				receiverAmountSum.Add(receiverAmountSum, new(big.Int).SetUint64(receiver.Amount))
			}
		} else if ethDestConfig != nil {
			if !goEthCommon.IsHexAddress(receiverAddr) {
				foundAnInvalidReceiverAddr = true

				break
			}

			if receiverAddr == common.EthZeroAddr {
				feeSum += receiver.Amount
			} else {
				receiverAmountSum.Add(receiverAmountSum, new(big.Int).SetUint64(receiver.Amount))
			}
		}
	}

	if foundAUtxoValueBelowMinimumValue {
		return fmt.Errorf("found a utxo value below minimum value in metadata receivers: %v", metadata)
	}

	if foundAnInvalidReceiverAddr {
		return fmt.Errorf("found an invalid receiver addr in metadata: %v", metadata)
	}

	// update fee amount if needed with sum of fee address receivers
	metadata.FeeAmount += feeSum
	receiverAmountSum.Add(receiverAmountSum, new(big.Int).SetUint64(metadata.FeeAmount))

	if metadata.FeeAmount < appConfig.BridgingSettings.MinFeeForBridging {
		return fmt.Errorf("bridging fee in metadata receivers is less than minimum: %v", metadata)
	}

	if receiverAmountSum.Cmp(new(big.Int).SetUint64(multisigUtxo.Amount)) != 0 {
		return fmt.Errorf("multisig amount is not equal to sum of receiver amounts + fee: expected %v but got %v",
			multisigUtxo.Amount, receiverAmountSum)
	}

	return nil
}
