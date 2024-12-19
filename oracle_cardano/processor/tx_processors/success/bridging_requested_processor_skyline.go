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

var _ core.CardanoTxSuccessProcessor = (*BridgingRequestedProcessorSkylineImpl)(nil)

type BridgingRequestedProcessorSkylineImpl struct {
	logger hclog.Logger
}

func NewNativeBridgingRequestedProcessor(logger hclog.Logger) *BridgingRequestedProcessorSkylineImpl {
	return &BridgingRequestedProcessorSkylineImpl{
		logger: logger.Named("bridging_requested_processor_skyline"),
	}
}

func (*BridgingRequestedProcessorSkylineImpl) GetType() common.BridgingTxType {
	return common.BridgingTxTypeBridgingRequest
}

func (*BridgingRequestedProcessorSkylineImpl) PreValidate(tx *core.CardanoTx, appConfig *cCore.AppConfig) error {
	return nil
}

func (p *BridgingRequestedProcessorSkylineImpl) ValidateAndAddClaim(
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
		return fmt.Errorf("validation failed for tx: %s, err: %w", tx.Hash, err)
	}

	return nil
}

func (p *BridgingRequestedProcessorSkylineImpl) addBridgingRequestClaim(
	claims *cCore.BridgeClaims, tx *core.CardanoTx,
	metadata *common.BridgingRequestMetadata, appConfig *cCore.AppConfig,
) {
	totalAmountSrc := uint64(0)
	totalAmountDest := uint64(0)
	totalAmountWrappedTokenSrc := uint64(0)
	totalAmountWrappedTokenDest := uint64(0)

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

		var (
			receiverAmount        uint64
			receiverAmountWrapped uint64
		)

		if receiver.IsNativeToken {
			receiverAmount = receiver.Amount + receiver.Additional.DestinationCurrencyAmount
			receiverAmountWrapped = uint64(0)

			totalAmountSrc += receiver.Additional.SourceCurrencyAmount
			totalAmountWrappedTokenSrc += receiver.Amount

			totalAmountDest += receiver.Amount + receiver.Additional.DestinationCurrencyAmount
			totalAmountWrappedTokenDest += uint64(0)
		} else {
			receiverAmount = receiver.Additional.DestinationCurrencyAmount
			receiverAmountWrapped = receiver.Amount

			totalAmountSrc += receiver.Amount + receiver.Additional.SourceCurrencyAmount
			totalAmountWrappedTokenSrc += uint64(0)

			totalAmountDest += receiver.Additional.DestinationCurrencyAmount
			totalAmountWrappedTokenDest += receiver.Amount
		}

		receivers = append(receivers, cCore.BridgingRequestReceiver{
			DestinationAddress: receiverAddr,
			Amount:             new(big.Int).SetUint64(receiverAmount),
			AmountWrapped:      new(big.Int).SetUint64(receiverAmountWrapped),
		})
	}

	feeAmount := metadata.FeeAmount.DestinationCurrencyAmount

	totalAmountDest += feeAmount

	receivers = append(receivers, cCore.BridgingRequestReceiver{
		DestinationAddress: feeAddress,
		Amount:             new(big.Int).SetUint64(feeAmount),
		AmountWrapped:      new(big.Int).SetUint64(0),
	})

	claim := cCore.BridgingRequestClaim{
		ObservedTransactionHash:         tx.Hash,
		SourceChainId:                   common.ToNumChainID(tx.OriginChainID),
		DestinationChainId:              common.ToNumChainID(metadata.DestinationChainID),
		Receivers:                       receivers,
		NativeCurrencyAmountDestination: new(big.Int).SetUint64(totalAmountDest),
		NativeCurrencyAmountSource:      new(big.Int).SetUint64(totalAmountSrc),
		WrappedTokenAmountSource:        new(big.Int).SetUint64(totalAmountWrappedTokenSrc),
		WrappedTokenAmountDestination:   new(big.Int).SetUint64(totalAmountWrappedTokenDest),
		RetryCounter:                    big.NewInt(int64(tx.BatchFailedCount)),
	}

	claims.BridgingRequestClaims = append(claims.BridgingRequestClaims, claim)

	p.logger.Info("Added BridgingRequestClaim",
		"txHash", tx.Hash, "metadata", metadata, "claim", cCore.BridgingRequestClaimString(claim))
}

/*
func (*BridgingRequestedProcessorSkylineImpl) addRefundRequestClaim(
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

func (p *BridgingRequestedProcessorSkylineImpl) validate(
	tx *core.CardanoTx, metadata *common.BridgingRequestMetadata, appConfig *cCore.AppConfig,
) error {
	chainConfig := appConfig.CardanoChains[tx.OriginChainID]
	if chainConfig == nil {
		return fmt.Errorf("unsupported chain id found in tx. chain id: %v", tx.OriginChainID)
	}

	if err := utils.ValidateOutputsHaveTokens(tx, appConfig); err != nil {
		return err
	}

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

	exchangeRate, err := GetExchangeRate(metadata.DestinationChainID)
	if err != nil {
		return err
	}

	for _, receiver := range metadata.Transactions {
		receiverAddr := strings.Join(receiver.Address, "")

		var totalAmountToBridge uint64

		if cardanoDestConfig != nil {
			if receiver.Additional == nil {
				if !cardanotx.IsValidOutputAddress(receiverAddr, cardanoDestConfig.NetworkID) {
					foundAnInvalidReceiverAddr = true

					break
				}

				if receiverAddr == cardanoDestConfig.BridgingAddresses.FeeAddress {
					feeSum += receiver.Amount

					continue
				} else {
					foundAUtxoValueBelowMinimumValue = true

					break
				}
			}

			if receiver.IsNativeToken {
				// amount_to_bridge must be >= minUtxoAmount on destination
				if receiver.Amount < cardanoDestConfig.UtxoMinAmount {
					foundAUtxoValueBelowMinimumValue = true

					break
				}

				totalAmountToBridge = receiver.Amount
			} else {
				// amount_to_bridge must be >= minUtxoAmount on source
				if receiver.Amount < cardanoSrcConfig.UtxoMinAmount {
					foundAUtxoValueBelowMinimumValue = true

					break
				}

				if receiver.Additional.DestinationCurrencyAmount < cardanoDestConfig.UtxoMinAmount {
					foundAUtxoValueBelowMinimumValue = true

					break
				}

				totalAmountToBridge = receiver.Amount
			}

			if uint64(float64(receiver.Additional.DestinationCurrencyAmount)*exchangeRate) !=
				receiver.Additional.SourceCurrencyAmount {
				return fmt.Errorf("found an exchange rate error in metadata: %v", metadata)
			}

			if !cardanotx.IsValidOutputAddress(receiverAddr, cardanoDestConfig.NetworkID) {
				foundAnInvalidReceiverAddr = true

				break
			}

			receiverAmountSum.Add(receiverAmountSum, new(big.Int).SetUint64(totalAmountToBridge))
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
	feeAmount := metadata.FeeAmount.DestinationCurrencyAmount + feeSum
	receiverAmountSum.Add(receiverAmountSum, new(big.Int).SetUint64(feeAmount))

	if cardanoDestConfig != nil && feeAmount < cardanoDestConfig.MinFeeForBridging {
		return fmt.Errorf("bridging fee in metadata receivers is less than minimum: %v", metadata)
	}

	if receiverAmountSum.Cmp(new(big.Int).SetUint64(multisigUtxo.Amount)) != 0 {
		return fmt.Errorf("multisig amount is not equal to sum of receiver amounts + fee: expected %v but got %v",
			multisigUtxo.Amount, receiverAmountSum)
	}

	if appConfig.BridgingSettings.MaxAmountAllowedToBridge != nil &&
		appConfig.BridgingSettings.MaxAmountAllowedToBridge.Sign() > 0 &&
		receiverAmountSum.Cmp(appConfig.BridgingSettings.MaxAmountAllowedToBridge) == 1 {
		return fmt.Errorf("sum of receiver amounts + fee: %v greater than maximum allowed: %v",
			receiverAmountSum, appConfig.BridgingSettings.MaxAmountAllowedToBridge)
	}

	return nil
}

func GetExchangeRate(destinationChainID string) (float64, error) {
	if destinationChainID == "prime" {
		return 0.5, nil
	}

	return 2, nil
}
