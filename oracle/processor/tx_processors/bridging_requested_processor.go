package txprocessors

import (
	"fmt"
	"math/big"
	"strings"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/oracle/core"
	"github.com/Ethernal-Tech/apex-bridge/oracle/utils"
	wallet "github.com/Ethernal-Tech/cardano-infrastructure/wallet"
	"github.com/hashicorp/go-hclog"
)

var _ core.CardanoTxProcessor = (*BridgingRequestedProcessorImpl)(nil)

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

func (p *BridgingRequestedProcessorImpl) ValidateAndAddClaim(
	claims *core.BridgeClaims, tx *core.CardanoTx, appConfig *core.AppConfig,
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
	claims *core.BridgeClaims, tx *core.CardanoTx,
	metadata *common.BridgingRequestMetadata, appConfig *core.AppConfig,
) {
	destConfig := appConfig.CardanoChains[metadata.DestinationChainID]
	totalAmount := big.NewInt(0)

	receivers := make([]core.BridgingRequestReceiver, 0, len(metadata.Transactions))
	for _, receiver := range metadata.Transactions {
		receiverAddr := strings.Join(receiver.Address, "")

		if receiverAddr == destConfig.BridgingAddresses.FeeAddress {
			// fee address will be added at the end
			continue
		}

		receivers = append(receivers, core.BridgingRequestReceiver{
			DestinationAddress: receiverAddr,
			Amount:             new(big.Int).SetUint64(receiver.Amount),
		})

		totalAmount.Add(totalAmount, new(big.Int).SetUint64(receiver.Amount))
	}

	totalAmount.Add(totalAmount, new(big.Int).SetUint64(metadata.FeeAmount))

	receivers = append(receivers, core.BridgingRequestReceiver{
		DestinationAddress: destConfig.BridgingAddresses.FeeAddress,
		Amount:             new(big.Int).SetUint64(metadata.FeeAmount),
	})

	claim := core.BridgingRequestClaim{
		ObservedTransactionHash: tx.Hash,
		SourceChainId:           common.ToNumChainID(tx.OriginChainID),
		DestinationChainId:      common.ToNumChainID(metadata.DestinationChainID),
		Receivers:               receivers,
		TotalAmount:             totalAmount,
	}

	claims.BridgingRequestClaims = append(claims.BridgingRequestClaims, claim)

	p.logger.Info("Added BridgingRequestClaim",
		"txHash", tx.Hash, "metadata", metadata, "claim", core.BridgingRequestClaimString(claim))
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
	tx *core.CardanoTx, metadata *common.BridgingRequestMetadata, appConfig *core.AppConfig,
) error {
	multisigUtxo, err := utils.ValidateTxOutputs(tx, appConfig)
	if err != nil {
		return err
	}

	cardanoDestConfig, ethDestConfig := utils.GetChainConfig(appConfig, metadata.DestinationChainID)
	if cardanoDestConfig == nil && ethDestConfig == nil {
		return fmt.Errorf("destination chain not registered: %v", metadata.DestinationChainID)
	}

	if len(metadata.Transactions) > appConfig.BridgingSettings.MaxReceiversPerBridgingRequest {
		return fmt.Errorf("number of receivers in metadata greater than maximum allowed - no: %v, max: %v, metadata: %v",
			len(metadata.Transactions), appConfig.BridgingSettings.MaxReceiversPerBridgingRequest, metadata)
	}

	receiverAmountSum := uint64(0)
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
			// BridgingRequestedProcessorImpl must know for which chain it operates

			addr, err := wallet.NewAddress(receiverAddr)
			if err != nil || addr.GetNetwork() != cardanoDestConfig.NetworkID {
				foundAnInvalidReceiverAddr = true

				break
			}

			if receiverAddr == cardanoDestConfig.BridgingAddresses.FeeAddress {
				feeSum += receiver.Amount
			} else {
				receiverAmountSum += receiver.Amount
			}
		} else if ethDestConfig != nil {
			// a TODO: validate eth addresses
			if receiverAddr == ethDestConfig.BridgingAddresses.FeeAddress {
				feeSum += receiver.Amount
			} else {
				receiverAmountSum += receiver.Amount
			}
		}
	}

	if cardanoDestConfig != nil && foundAUtxoValueBelowMinimumValue {
		return fmt.Errorf("found a utxo value below minimum value in metadata receivers: %v", metadata)
	}

	if foundAnInvalidReceiverAddr {
		return fmt.Errorf("found an invalid receiver addr in metadata: %v", metadata)
	}

	// update fee amount if needed with sum of fee address receivers
	metadata.FeeAmount += feeSum
	receiverAmountSum += metadata.FeeAmount

	if metadata.FeeAmount < appConfig.BridgingSettings.MinFeeForBridging {
		return fmt.Errorf("bridging fee in metadata receivers is less than minimum: %v", metadata)
	}

	if receiverAmountSum != multisigUtxo.Amount {
		return fmt.Errorf("receivers amounts and multisig amount missmatch: expected %v but got %v",
			receiverAmountSum, multisigUtxo.Amount)
	}

	return nil
}
