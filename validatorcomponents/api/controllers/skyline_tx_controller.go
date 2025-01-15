package controllers

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"
	"net/http"

	batcherCore "github.com/Ethernal-Tech/apex-bridge/batcher/core"
	cardanotx "github.com/Ethernal-Tech/apex-bridge/cardano"
	"github.com/Ethernal-Tech/apex-bridge/common"
	oCore "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	oUtils "github.com/Ethernal-Tech/apex-bridge/oracle_common/utils"
	"github.com/Ethernal-Tech/apex-bridge/validatorcomponents/api/model/request"
	"github.com/Ethernal-Tech/apex-bridge/validatorcomponents/api/model/response"
	"github.com/Ethernal-Tech/apex-bridge/validatorcomponents/api/utils"
	"github.com/Ethernal-Tech/apex-bridge/validatorcomponents/core"
	"github.com/Ethernal-Tech/cardano-infrastructure/sendtx"
	"github.com/Ethernal-Tech/cardano-infrastructure/wallet"
	"github.com/hashicorp/go-hclog"
)

type SkylineTxControllerImpl struct {
	oracleConfig  *oCore.AppConfig
	batcherConfig *batcherCore.BatcherManagerConfiguration
	logger        hclog.Logger
}

var _ core.APIController = (*SkylineTxControllerImpl)(nil)

func NewSkylineTxController(
	oracleConfig *oCore.AppConfig,
	batcherConfig *batcherCore.BatcherManagerConfiguration,
	logger hclog.Logger,
) *CardanoTxControllerImpl {
	return &CardanoTxControllerImpl{
		oracleConfig:  oracleConfig,
		batcherConfig: batcherConfig,
		logger:        logger,
	}
}

func (*SkylineTxControllerImpl) GetPathPrefix() string {
	return "CardanoTx"
}

func (sc *SkylineTxControllerImpl) GetEndpoints() []*core.APIEndpoint {
	return []*core.APIEndpoint{
		{Path: "CreateBridgingTx", Method: http.MethodPost, Handler: sc.createBridgingTx, APIKeyAuth: true},
	}
}

func (sc *SkylineTxControllerImpl) createBridgingTx(w http.ResponseWriter, r *http.Request) {
	requestBody, ok := utils.DecodeModel[request.CreateBridgingTxRequest](w, r, sc.logger)
	if !ok {
		return
	}

	sc.logger.Debug("createBridgingTx request", "body", requestBody, "url", r.URL)

	err := sc.validateAndFillOutCreateBridgingTxRequest(&requestBody)
	if err != nil {
		utils.WriteErrorResponse(
			w, r, http.StatusBadRequest,
			fmt.Errorf("validation error. err: %w", err), sc.logger,
		)

		return
	}

	txRaw, txHash, bridgingRequestMetadata, err := sc.createTx(requestBody)
	if err != nil {
		utils.WriteErrorResponse(w, r, http.StatusInternalServerError, err, sc.logger)

		return
	}

	currencyOutput, tokenOutput, bridgingFee := getOutputAmounts(bridgingRequestMetadata)

	utils.WriteResponse(
		w, r, http.StatusOK,
		response.NewFullSkylineBridgingTxResponse(txRaw, txHash, bridgingFee, currencyOutput, tokenOutput), sc.logger,
	)
}

func (sc *SkylineTxControllerImpl) validateAndFillOutCreateBridgingTxRequest(
	requestBody *request.CreateBridgingTxRequest,
) error {
	cardanoSrcConfig, _ := oUtils.GetChainConfig(sc.oracleConfig, requestBody.SourceChainID)
	if cardanoSrcConfig == nil {
		return fmt.Errorf("origin chain not registered: %v", requestBody.SourceChainID)
	}

	cardanoDestConfig, _ := oUtils.GetChainConfig(sc.oracleConfig, requestBody.DestinationChainID)
	if cardanoDestConfig == nil {
		return fmt.Errorf("destination chain not registered: %v", requestBody.DestinationChainID)
	}

	if len(requestBody.Transactions) > sc.oracleConfig.BridgingSettings.MaxReceiversPerBridgingRequest {
		return fmt.Errorf("number of receivers in metadata greater than maximum allowed - no: %v, max: %v, requestBody: %v",
			len(requestBody.Transactions), sc.oracleConfig.BridgingSettings.MaxReceiversPerBridgingRequest, requestBody)
	}

	receiverAmountSum := big.NewInt(0)
	feeSum := uint64(0)
	foundAUtxoValueBelowMinimumValue := false
	foundAnInvalidReceiverAddr := false
	transactions := make([]request.CreateBridgingTxTransactionRequest, 0, len(requestBody.Transactions))

	for _, receiver := range requestBody.Transactions {
		if receiver.Amount < cardanoDestConfig.UtxoMinAmount {
			foundAUtxoValueBelowMinimumValue = true

			break
		}

		addr, err := wallet.NewCardanoAddressFromString(receiver.Addr)
		if err != nil || addr.GetInfo().Network != cardanoDestConfig.NetworkID {
			foundAnInvalidReceiverAddr = true

			break
		}

		// if fee address is specified in transactions just add amount to the fee sum
		// otherwise keep this transaction
		if receiver.Addr == cardanoDestConfig.BridgingAddresses.FeeAddress {
			feeSum += receiver.Amount
		} else {
			transactions = append(transactions, receiver)
			receiverAmountSum.Add(receiverAmountSum, new(big.Int).SetUint64(receiver.Amount))
		}
	}

	if foundAUtxoValueBelowMinimumValue {
		return fmt.Errorf("found a utxo value below minimum value in request body receivers: %v", requestBody)
	}

	if foundAnInvalidReceiverAddr {
		return fmt.Errorf("found an invalid receiver addr in request body: %v", requestBody)
	}

	requestBody.BridgingFee += feeSum
	requestBody.Transactions = transactions

	// this is just convinient way to setup default min fee
	if requestBody.BridgingFee == 0 {
		requestBody.BridgingFee = cardanoDestConfig.MinFeeForBridging
	}

	receiverAmountSum.Add(receiverAmountSum, new(big.Int).SetUint64(requestBody.BridgingFee))

	if requestBody.BridgingFee < cardanoDestConfig.MinFeeForBridging {
		return fmt.Errorf("bridging fee in request body is less than minimum: %v", requestBody)
	}

	if sc.oracleConfig.BridgingSettings.MaxAmountAllowedToBridge != nil &&
		sc.oracleConfig.BridgingSettings.MaxAmountAllowedToBridge.Sign() > 0 &&
		receiverAmountSum.Cmp(sc.oracleConfig.BridgingSettings.MaxAmountAllowedToBridge) == 1 {
		return fmt.Errorf("sum of receiver amounts + fee greater than maximum allowed: %v, for request: %v",
			sc.oracleConfig.BridgingSettings.MaxAmountAllowedToBridge, requestBody)
	}

	return nil
}

func (sc *SkylineTxControllerImpl) createTx(requestBody request.CreateBridgingTxRequest) (
	string, string, *sendtx.BridgingRequestMetadata, error,
) {
	var batcherChainConfig batcherCore.ChainConfig

	for _, batcherChain := range sc.batcherConfig.Chains {
		if batcherChain.ChainID == requestBody.SourceChainID {
			batcherChainConfig = batcherChain

			break
		}
	}

	cardanoConfig, err := cardanotx.NewCardanoChainConfig(batcherChainConfig.ChainSpecific)
	if err != nil {
		return "", "", nil, err
	}

	txProvider, err := cardanoConfig.CreateTxProvider()
	if err != nil {
		return "", "", nil, fmt.Errorf("failed to create tx provider: %w", err)
	}

	sourceChainConfig := sc.oracleConfig.CardanoChains[requestBody.SourceChainID]
	minAmountToBridge := uint64(0)

	destCardanoChainConfig, exists := sc.oracleConfig.CardanoChains[requestBody.DestinationChainID]
	if exists {
		minAmountToBridge = destCardanoChainConfig.UtxoMinAmount
	}

	txSender := sendtx.NewTxSender(
		requestBody.BridgingFee,
		minAmountToBridge,
		cardanoConfig.PotentialFee,
		common.MaxInputsPerBridgingTxDefault,
		map[string]sendtx.ChainConfig{
			requestBody.SourceChainID: {
				CardanoCliBinary: wallet.ResolveCardanoCliBinary(sourceChainConfig.NetworkID),
				TxProvider:       txProvider,
				MultiSigAddr:     sourceChainConfig.BridgingAddresses.BridgingAddress,
				TestNetMagic:     uint(sourceChainConfig.NetworkMagic),
				TTLSlotNumberInc: cardanoConfig.TTLSlotNumberInc,
				MinUtxoValue:     sourceChainConfig.UtxoMinAmount,
				ExchangeRate:     make(map[string]float64),
			},
			requestBody.DestinationChainID: {},
		},
	)

	receivers := make([]sendtx.BridgingTxReceiver, len(requestBody.Transactions))
	for i, tx := range requestBody.Transactions {
		receivers[i] = sendtx.BridgingTxReceiver{
			Addr:   tx.Addr,
			Amount: tx.Amount,
		}
		if tx.IsNativeToken {
			receivers[i].BridgingType = sendtx.BridgingTypeNativeTokenOnSource
		} else {
			receivers[i].BridgingType = sendtx.BridgingTypeCurrencyOnSource
		}
	}

	txRawBytes, txHash, metadata, err := txSender.CreateBridgingTx(
		context.Background(),
		requestBody.SourceChainID, requestBody.DestinationChainID,
		requestBody.SenderAddr, receivers,
	)
	if err != nil {
		return "", "", nil, fmt.Errorf("failed to build tx: %w", err)
	}

	txRaw := hex.EncodeToString(txRawBytes)

	return txRaw, txHash, metadata, nil
}

func getOutputAmounts(metadata *sendtx.BridgingRequestMetadata) (outputCurrencyLovelace uint64, outputNativeToken uint64, bridgingFee uint64) {
	bridgingFee = metadata.FeeAmount.SrcAmount

	for _, x := range metadata.Transactions {
		if x.IsNativeTokenOnSource() {
			// WADA/WAPEX to ADA/APEX
			outputNativeToken += x.Amount
		} else {
			// ADA/APEX to WADA/WAPEX or reactor
			outputCurrencyLovelace += x.Amount
		}

		if x.Additional != nil {
			bridgingFee += x.Additional.SrcAmount
		}
	}

	return outputCurrencyLovelace, outputNativeToken, bridgingFee
}
