package successtxprocessors

import (
	"fmt"
	"math/big"

	"github.com/Ethernal-Tech/apex-bridge/common"
	oCore "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	oUtils "github.com/Ethernal-Tech/apex-bridge/oracle_common/utils"
	"github.com/Ethernal-Tech/apex-bridge/oracle_eth/core"
	wallet "github.com/Ethernal-Tech/cardano-infrastructure/wallet"
	"github.com/hashicorp/go-hclog"
)

var _ core.EthTxSuccessProcessor = (*BridgingRequestedProcessorImpl)(nil)

type BridgingRequestedProcessorImpl struct {
	logger hclog.Logger
}

func NewEthBridgingRequestedProcessor(logger hclog.Logger) *BridgingRequestedProcessorImpl {
	return &BridgingRequestedProcessorImpl{
		logger: logger.Named("eth_bridging_requested_processor"),
	}
}

func (*BridgingRequestedProcessorImpl) GetType() common.BridgingTxType {
	return common.BridgingTxTypeBridgingRequest
}

func (p *BridgingRequestedProcessorImpl) ValidateAndAddClaim(
	claims *oCore.BridgeClaims, tx *core.EthTx, appConfig *oCore.AppConfig,
) error {
	metadata, err := core.UnmarshalEthMetadata[core.BridgingRequestEthMetadata](
		tx.Metadata)
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
	claims *oCore.BridgeClaims, tx *core.EthTx,
	metadata *core.BridgingRequestEthMetadata, appConfig *oCore.AppConfig,
) {
	totalAmount := big.NewInt(0)

	cardanoDestConfig, _ := oUtils.GetChainConfig(appConfig, metadata.DestinationChainID)

	receivers := make([]oCore.BridgingRequestReceiver, 0, len(metadata.Transactions))

	for _, receiver := range metadata.Transactions {
		if receiver.Address == cardanoDestConfig.BridgingAddresses.FeeAddress {
			// fee address will be added at the end
			continue
		}

		receiverAmountDfm := common.WeiToDfm(receiver.Amount)

		receivers = append(receivers, oCore.BridgingRequestReceiver{
			DestinationAddress: receiver.Address,
			Amount:             receiverAmountDfm,
		})

		totalAmount.Add(totalAmount, receiverAmountDfm)
	}

	feeAmountDfm := common.WeiToDfm(metadata.FeeAmount)
	totalAmount.Add(totalAmount, feeAmountDfm)

	receivers = append(receivers, oCore.BridgingRequestReceiver{
		DestinationAddress: cardanoDestConfig.BridgingAddresses.FeeAddress,
		Amount:             feeAmountDfm,
	})

	claim := oCore.BridgingRequestClaim{
		ObservedTransactionHash: tx.Hash,
		SourceChainId:           common.ToNumChainID(tx.OriginChainID),
		DestinationChainId:      common.ToNumChainID(metadata.DestinationChainID),
		Receivers:               receivers,
		TotalAmount:             totalAmount,
	}

	claims.BridgingRequestClaims = append(claims.BridgingRequestClaims, claim)

	p.logger.Info("Added BridgingRequestClaim",
		"txHash", tx.Hash, "metadata", metadata, "claim", oCore.BridgingRequestClaimString(claim))
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
	tx *core.EthTx, metadata *core.BridgingRequestEthMetadata, appConfig *oCore.AppConfig,
) error {
	_, ethSrcConfig := oUtils.GetChainConfig(appConfig, tx.OriginChainID)
	if ethSrcConfig == nil {
		return fmt.Errorf("origin chain not registered: %v", tx.OriginChainID)
	}

	cardanoDestConfig, _ := oUtils.GetChainConfig(appConfig, metadata.DestinationChainID)
	if cardanoDestConfig == nil {
		return fmt.Errorf("destination chain not registered: %v", metadata.DestinationChainID)
	}

	if len(metadata.Transactions) > appConfig.BridgingSettings.MaxReceiversPerBridgingRequest {
		return fmt.Errorf("number of receivers in metadata greater than maximum allowed - no: %v, max: %v, metadata: %v",
			len(metadata.Transactions), appConfig.BridgingSettings.MaxReceiversPerBridgingRequest, metadata)
	}

	receiverAmountSum := big.NewInt(0)
	feeSum := big.NewInt(0)
	foundAUtxoValueBelowMinimumValue := false
	foundAnInvalidReceiverAddr := false

	for _, receiver := range metadata.Transactions {
		receiverAmountDfm := common.WeiToDfm(receiver.Amount)
		if receiverAmountDfm.Uint64() < appConfig.BridgingSettings.UtxoMinValue {
			foundAUtxoValueBelowMinimumValue = true

			break
		}

		addr, err := wallet.NewAddress(receiver.Address)
		if err != nil || addr.GetNetwork() != cardanoDestConfig.NetworkID {
			foundAnInvalidReceiverAddr = true

			break
		}

		if receiver.Address == cardanoDestConfig.BridgingAddresses.FeeAddress {
			feeSum.Add(feeSum, receiver.Amount)
		} else {
			receiverAmountSum.Add(receiverAmountSum, receiver.Amount)
		}
	}

	if foundAUtxoValueBelowMinimumValue {
		return fmt.Errorf("found a utxo value below minimum value in metadata receivers: %v", metadata)
	}

	if foundAnInvalidReceiverAddr {
		return fmt.Errorf("found an invalid receiver addr in metadata: %v", metadata)
	}

	// update fee amount if needed with sum of fee address receivers
	metadata.FeeAmount.Add(metadata.FeeAmount, feeSum)
	receiverAmountSum.Add(receiverAmountSum, metadata.FeeAmount)

	feeAmountDfm := common.WeiToDfm(metadata.FeeAmount)
	if feeAmountDfm.Uint64() < appConfig.BridgingSettings.MinFeeForBridging {
		return fmt.Errorf("bridging fee in metadata receivers is less than minimum: %v", metadata)
	}

	if tx.Value == nil || tx.Value.Cmp(receiverAmountSum) != 0 {
		return fmt.Errorf("tx value is not equal to sum of receiver amounts + fee: expected %v but got %v",
			receiverAmountSum, tx.Value)
	}

	return nil
}
