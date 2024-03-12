package tx_processors

import (
	"fmt"

	"github.com/Ethernal-Tech/apex-bridge/oracle/core"
	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
	wallet "github.com/Ethernal-Tech/cardano-infrastructure/wallet"
)

var _ core.CardanoTxProcessor = (*BridgingRequestedProcessorImpl)(nil)

type BridgingRequestedProcessorImpl struct {
}

func NewBridgingRequestedProcessor() *BridgingRequestedProcessorImpl {
	return &BridgingRequestedProcessorImpl{}
}

func (*BridgingRequestedProcessorImpl) IsTxRelevant(tx *core.CardanoTx, appConfig *core.AppConfig) (bool, error) {
	metadata, err := core.UnmarshalBaseMetadata(tx.Metadata)

	if err == nil && metadata != nil {
		return metadata.BridgingTxType == core.BridgingTxTypeBridgingRequest, err
	}

	return false, err
}

func (p *BridgingRequestedProcessorImpl) ValidateAndAddClaim(claims *core.BridgeClaims, tx *core.CardanoTx, appConfig *core.AppConfig) error {
	relevant, err := p.IsTxRelevant(tx, appConfig)
	if err != nil {
		return err
	}

	if !relevant {
		return fmt.Errorf("ValidateAndAddClaim called for irrelevant tx: %v", tx)
	}

	metadata, err := core.UnmarshalBridgingRequestMetadata(tx.Metadata)
	if err != nil {
		return fmt.Errorf("failed to unmarshal metadata: tx: %v,\n err: %v", tx, err)
	}

	err = p.validate(tx, metadata, appConfig)
	if err == nil {
		p.addBridgingRequestClaim(claims, tx, metadata)
	} else {
		return fmt.Errorf("validation failed for tx: %v", tx)
		// p.addRefundRequestClaim(claims, tx, metadata)
	}

	return nil
}

func (*BridgingRequestedProcessorImpl) addBridgingRequestClaim(claims *core.BridgeClaims, tx *core.CardanoTx, metadata *core.BridgingRequestMetadata) {
	var receivers []core.BridgingRequestReceiver
	for _, receiver := range metadata.Transactions {
		receivers = append(receivers, core.BridgingRequestReceiver{
			Address: receiver.Address,
			Amount:  receiver.Amount,
		})
	}

	var utxos []core.Utxo
	for _, utxo := range tx.Outputs {
		utxos = append(utxos, core.Utxo{
			Address: utxo.Address,
			Amount:  utxo.Amount,
		})
	}
	claim := core.BridgingRequestClaim{
		TxHash:             tx.Hash,
		DestinationChainId: metadata.DestinationChainId,
		Receivers:          receivers,
		OutputUtxos:        utxos,
	}

	claims.BridgingRequest = append(claims.BridgingRequest, claim)
}

/*
func (*BridgingRequestedProcessorImpl) addRefundRequestClaim(claims *core.BridgeClaims, tx *core.CardanoTx, metadata *core.BridgingRequestMetadata) {

		var outputUtxos []core.Utxo
		for _, output := range tx.Outputs {
			outputUtxos = append(outputUtxos, core.Utxo{
				Address: output.Address,
				Amount:  output.Amount,
			})
		}

		// TODO: what goes into UtxoTransaction
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

func (*BridgingRequestedProcessorImpl) validate(tx *core.CardanoTx, metadata *core.BridgingRequestMetadata, appConfig *core.AppConfig) error {
	foundDestinationChainConfig := false
	var bridgingAddressOnOrigin *core.BridgingAddress
	var utxoToBridgingAddressOnOrigin *indexer.TxOutput = nil
	foundBridgingAddressOnOrigin := false
	foundMultipleUtxosToBridgingAddressOnOrigin := false
	for _, chainConfig := range appConfig.CardanoChains {
		if chainConfig.ChainId == tx.OriginChainId {
			for _, bridgingAddress := range chainConfig.BridgingAddresses {
				if bridgingAddress.ChainId == metadata.DestinationChainId {
					for _, utxo := range tx.Outputs {
						if utxo.Address == bridgingAddress.Address {
							if utxoToBridgingAddressOnOrigin != nil {
								foundMultipleUtxosToBridgingAddressOnOrigin = true
							} else {
								bridgingAddressOnOrigin = &bridgingAddress
								utxoToBridgingAddressOnOrigin = utxo
								foundBridgingAddressOnOrigin = true
							}
						}
					}
				}
			}
		} else if metadata.DestinationChainId == chainConfig.ChainId {
			foundDestinationChainConfig = true
		}
	}

	if !foundDestinationChainConfig || !foundBridgingAddressOnOrigin {
		return fmt.Errorf("destination chain not registered: %v", metadata.DestinationChainId)
	}

	if foundMultipleUtxosToBridgingAddressOnOrigin {
		return fmt.Errorf("found multiple utxos to the bridging address on origin: %v", tx)
	}

	var receiverAmountSum uint64 = 0
	var feeSum uint64 = 0
	foundAUtxoValueBelowMinimumValue := false
	foundAnInvalidReceiverAddr := false
	foundTheFeeReceiverAddr := false
	for _, receiver := range metadata.Transactions {
		if receiver.Amount < appConfig.BridgingSettings.UtxoMinValue {
			foundAUtxoValueBelowMinimumValue = true
			break
		}

		addrInfo := wallet.GetAddressInfo(receiver.Address, wallet.AddressTypeAny)
		if !addrInfo.IsValid {
			foundAnInvalidReceiverAddr = true
			break
		}

		if receiver.Address == bridgingAddressOnOrigin.FeeAddress {
			foundTheFeeReceiverAddr = true
			feeSum += receiver.Amount
		}

		receiverAmountSum += receiver.Amount
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
