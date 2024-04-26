package txprocessors

import (
	"fmt"
	"math/big"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/oracle/core"
	"github.com/Ethernal-Tech/apex-bridge/oracle/utils"
	wallet "github.com/Ethernal-Tech/cardano-infrastructure/wallet"
)

var _ core.CardanoTxProcessor = (*BridgingRequestedProcessorImpl)(nil)

type BridgingRequestedProcessorImpl struct {
}

func NewBridgingRequestedProcessor() *BridgingRequestedProcessorImpl {
	return &BridgingRequestedProcessorImpl{}
}

func (*BridgingRequestedProcessorImpl) GetType() core.TxProcessorType {
	return core.TxProcessorTypeBridgingRequest
}

func (*BridgingRequestedProcessorImpl) IsTxRelevant(tx *core.CardanoTx) (bool, error) {
	metadata, err := common.UnmarshalMetadata[common.BaseMetadata](common.MetadataEncodingTypeCbor, tx.Metadata)

	if err == nil && metadata != nil {
		return metadata.BridgingTxType == common.BridgingTxTypeBridgingRequest, err
	}

	return false, err
}

func (p *BridgingRequestedProcessorImpl) ValidateAndAddClaim(
	claims *core.BridgeClaims, tx *core.CardanoTx, appConfig *core.AppConfig,
) error {
	relevant, err := p.IsTxRelevant(tx)
	if err != nil {
		return err
	}

	if !relevant {
		return fmt.Errorf("ValidateAndAddClaim called for irrelevant tx: %v", tx)
	}

	metadata, err := common.UnmarshalMetadata[common.BridgingRequestMetadata](common.MetadataEncodingTypeCbor, tx.Metadata)
	if err != nil {
		return fmt.Errorf("failed to unmarshal metadata: tx: %v, err: %w", tx, err)
	}

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

func (*BridgingRequestedProcessorImpl) addBridgingRequestClaim(
	claims *core.BridgeClaims, tx *core.CardanoTx, metadata *common.BridgingRequestMetadata, appConfig *core.AppConfig,
) {
	receivers := make([]core.BridgingRequestReceiver, 0, len(metadata.Transactions))
	for _, receiver := range metadata.Transactions {
		receivers = append(receivers, core.BridgingRequestReceiver{
			DestinationAddress: receiver.Address,
			Amount:             new(big.Int).SetUint64(receiver.Amount),
		})
	}

	var outputUtxo core.UTXO

	for _, utxo := range tx.Outputs {
		if utxo.Address == appConfig.CardanoChains[tx.OriginChainID].BridgingAddresses.BridgingAddress {
			outputUtxo = core.UTXO{
				TxHash:  tx.Hash,
				TxIndex: new(big.Int).SetUint64(uint64(tx.Indx)),
				Amount:  new(big.Int).SetUint64(utxo.Amount),
			}
		}
	}

	claim := core.BridgingRequestClaim{
		ObservedTransactionHash: tx.Hash,
		SourceChainID:           tx.OriginChainID,
		DestinationChainID:      metadata.DestinationChainID,
		OutputUTXO:              outputUtxo,
		Receivers:               receivers,
	}

	claims.BridgingRequestClaims = append(claims.BridgingRequestClaims, claim)
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

}
*/

func (*BridgingRequestedProcessorImpl) validate(
	tx *core.CardanoTx, metadata *common.BridgingRequestMetadata, appConfig *core.AppConfig,
) error {
	multisigUtxo, err := utils.ValidateTxOutputs(tx, appConfig)
	if err != nil {
		return err
	}

	destinationChainConfig := appConfig.CardanoChains[metadata.DestinationChainID]
	if destinationChainConfig == nil {
		return fmt.Errorf("destination chain not registered: %v", metadata.DestinationChainID)
	}

	if len(metadata.Transactions) > appConfig.BridgingSettings.MaxReceiversPerBridgingRequest {
		return fmt.Errorf("number of receivers in metadata greater than maximum allowed - no: %v, max: %v, metadata: %v",
			len(metadata.Transactions), appConfig.BridgingSettings.MaxReceiversPerBridgingRequest, metadata)
	}

	var (
		receiverAmountSum uint64 = 0
		feeSum            uint64 = 0
	)

	foundAUtxoValueBelowMinimumValue := false
	foundAnInvalidReceiverAddr := false
	foundTheFeeReceiverAddr := false

	for _, receiver := range metadata.Transactions {
		if receiver.Amount < appConfig.BridgingSettings.UtxoMinValue {
			foundAUtxoValueBelowMinimumValue = true

			break
		}

		_, err := wallet.GetAddressInfo(receiver.Address)
		if err != nil {
			foundAnInvalidReceiverAddr = true

			break
		}

		if receiver.Address == destinationChainConfig.BridgingAddresses.FeeAddress {
			foundTheFeeReceiverAddr = true
			feeSum += receiver.Amount
		}

		receiverAmountSum += receiver.Amount
	}

	if receiverAmountSum != multisigUtxo.Amount {
		return fmt.Errorf("receivers amounts and multisig amount missmatch: expected %v but got %v",
			receiverAmountSum, multisigUtxo.Amount)
	}

	if foundAUtxoValueBelowMinimumValue {
		return fmt.Errorf("found a utxo value below minimum value in metadata receivers: %v", metadata)
	}

	if foundAnInvalidReceiverAddr {
		return fmt.Errorf("found an invalid receiver addr in metadata: %v", metadata)
	}

	if !foundTheFeeReceiverAddr {
		return fmt.Errorf("destination chain fee address not found in receiver addrs in metadata: %v", metadata)
	}

	if feeSum < appConfig.BridgingSettings.MinFeeForBridging {
		return fmt.Errorf("bridging fee in metadata receivers is less than minimum: %v", metadata)
	}

	return nil
}
