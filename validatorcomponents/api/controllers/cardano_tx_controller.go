package controllers

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"

	batcherCore "github.com/Ethernal-Tech/apex-bridge/batcher/core"
	cardanotx "github.com/Ethernal-Tech/apex-bridge/cardano"
	"github.com/Ethernal-Tech/apex-bridge/common"
	oracleCore "github.com/Ethernal-Tech/apex-bridge/oracle/core"
	"github.com/Ethernal-Tech/apex-bridge/validatorcomponents/api/model/request"
	"github.com/Ethernal-Tech/apex-bridge/validatorcomponents/api/model/response"
	"github.com/Ethernal-Tech/apex-bridge/validatorcomponents/api/utils"
	"github.com/Ethernal-Tech/apex-bridge/validatorcomponents/core"
	"github.com/Ethernal-Tech/cardano-infrastructure/wallet"
	"github.com/hashicorp/go-hclog"
)

type CardanoTxControllerImpl struct {
	oracleConfig  *oracleCore.AppConfig
	batcherConfig *batcherCore.BatcherManagerConfiguration
	logger        hclog.Logger
}

var _ core.APIController = (*CardanoTxControllerImpl)(nil)

func NewCardanoTxController(
	oracleConfig *oracleCore.AppConfig,
	batcherConfig *batcherCore.BatcherManagerConfiguration,
	logger hclog.Logger,
) *CardanoTxControllerImpl {
	return &CardanoTxControllerImpl{
		oracleConfig:  oracleConfig,
		batcherConfig: batcherConfig,
		logger:        logger,
	}
}

func (*CardanoTxControllerImpl) GetPathPrefix() string {
	return "CardanoTx"
}

func (c *CardanoTxControllerImpl) GetEndpoints() []*core.APIEndpoint {
	return []*core.APIEndpoint{
		{Path: "CreateBridgingTx", Method: http.MethodPost, Handler: c.createBridgingTx, APIKeyAuth: true},
		{Path: "SignBridgingTx", Method: http.MethodPost, Handler: c.signBridgingTx, APIKeyAuth: true},
	}
}

func (c *CardanoTxControllerImpl) createBridgingTx(w http.ResponseWriter, r *http.Request) {
	c.logger.Debug("createBridgingTx called", "url", r.URL)

	var requestBody request.CreateBridgingTxRequest

	err := json.NewDecoder(r.Body).Decode(&requestBody)
	if err != nil {
		c.logger.Debug("createBridgingTx request", "err", err.Error(), "url", r.URL)

		rerr := utils.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
		if rerr != nil {
			c.logger.Error("error while WriteErrorResponse", "err", rerr)
		}

		return
	}

	c.logger.Debug("createBridgingTx request", "body", requestBody, "url", r.URL)

	err = c.validateAndFillOutCreateBridgingTxRequest(&requestBody)
	if err != nil {
		c.logger.Debug("createBridgingTx request", "err", err.Error(), "url", r.URL)

		rerr := utils.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
		if rerr != nil {
			c.logger.Error("error while WriteErrorResponse", "err", rerr)
		}

		return
	}

	txRaw, txHash, err := c.createTx(requestBody)
	if err != nil {
		c.logger.Debug("createBridgingTx request", "err", err.Error(), "url", r.URL)

		rerr := utils.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
		if rerr != nil {
			c.logger.Error("error while WriteErrorResponse", "err", rerr)
		}

		return
	}

	c.logger.Debug("createBridgingTx success", "url", r.URL)

	w.Header().Set("Content-Type", "application/json")

	err = json.NewEncoder(w).Encode(response.NewFullBridgingTxResponse(txRaw, txHash, requestBody.BridgingFee))
	if err != nil {
		c.logger.Error("error while writing response", "err", err)
	}
}

func (c *CardanoTxControllerImpl) signBridgingTx(w http.ResponseWriter, r *http.Request) {
	c.logger.Debug("signBridgingTx called", "url", r.URL)

	var requestBody request.SignBridgingTxRequest

	err := json.NewDecoder(r.Body).Decode(&requestBody)
	if err != nil {
		c.logger.Debug("signBridgingTx request", "err", err.Error(), "url", r.URL)

		rerr := utils.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
		if rerr != nil {
			c.logger.Error("error while WriteErrorResponse", "err", rerr)
		}

		return
	}

	if requestBody.TxRaw == "" || requestBody.SigningKeyHex == "" || requestBody.TxHash == "" {
		c.logger.Debug("signBridgingTx request", "err", err.Error(), "url", r.URL)

		rerr := utils.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
		if rerr != nil {
			c.logger.Error("error while WriteErrorResponse", "err", rerr)
		}

		return
	}

	c.logger.Debug("signBridgingTx request", "body", requestBody, "url", r.URL)

	signedTx, err := c.signTx(requestBody)
	if err != nil {
		c.logger.Debug("signBridgingTx request", "err", err.Error(), "url", r.URL)

		rerr := utils.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
		if rerr != nil {
			c.logger.Error("error while WriteErrorResponse", "err", rerr)
		}

		return
	}

	c.logger.Debug("signBridgingTx success", "url", r.URL)

	w.Header().Set("Content-Type", "application/json")

	err = json.NewEncoder(w).Encode(response.NewBridgingTxResponse(signedTx, requestBody.TxHash))
	if err != nil {
		c.logger.Error("error while writing response", "err", err)
	}
}

func (c *CardanoTxControllerImpl) validateAndFillOutCreateBridgingTxRequest(
	requestBody *request.CreateBridgingTxRequest,
) error {
	_, exists := c.oracleConfig.CardanoChains[requestBody.SourceChainID]
	if !exists {
		return fmt.Errorf("source chain not registered: %v", requestBody.SourceChainID)
	}

	destinationChainConfig := c.oracleConfig.CardanoChains[requestBody.DestinationChainID]
	if destinationChainConfig == nil {
		return fmt.Errorf("destination chain not registered: %v", requestBody.DestinationChainID)
	}

	if len(requestBody.Transactions) > c.oracleConfig.BridgingSettings.MaxReceiversPerBridgingRequest {
		return fmt.Errorf("number of receivers in metadata greater than maximum allowed - no: %v, max: %v, requestBody: %v",
			len(requestBody.Transactions), c.oracleConfig.BridgingSettings.MaxReceiversPerBridgingRequest, requestBody)
	}

	var (
		receiverAmountSum uint64 = 0
		feeSum            uint64 = 0
	)

	foundAUtxoValueBelowMinimumValue := false
	foundAnInvalidReceiverAddr := false
	foundTheFeeReceiverAddr := false

	for _, receiver := range requestBody.Transactions {
		if receiver.Amount < c.oracleConfig.BridgingSettings.UtxoMinValue {
			foundAUtxoValueBelowMinimumValue = true

			break
		}

		_, err := wallet.GetAddressInfo(receiver.Addr)
		if err != nil {
			foundAnInvalidReceiverAddr = true

			break
		}

		if receiver.Addr == destinationChainConfig.BridgingAddresses.FeeAddress {
			foundTheFeeReceiverAddr = true
			feeSum += receiver.Amount
		}

		receiverAmountSum += receiver.Amount
	}

	if foundAUtxoValueBelowMinimumValue {
		return fmt.Errorf("found a utxo value below minimum value in request body receivers: %v", requestBody)
	}

	if foundAnInvalidReceiverAddr {
		return fmt.Errorf("found an invalid receiver addr in request body: %v", requestBody)
	}

	if !foundTheFeeReceiverAddr {
		if requestBody.BridgingFee == 0 {
			requestBody.BridgingFee = c.oracleConfig.BridgingSettings.MinFeeForBridging
		}

		requestBody.Transactions = append(requestBody.Transactions,
			request.CreateBridgingTxTransactionRequest{
				Addr:   destinationChainConfig.BridgingAddresses.FeeAddress,
				Amount: requestBody.BridgingFee,
			},
		)
	} else {
		requestBody.BridgingFee = feeSum
	}

	if requestBody.BridgingFee < c.oracleConfig.BridgingSettings.MinFeeForBridging {
		return fmt.Errorf("bridging fee in request body is less than minimum: %v", requestBody)
	}

	return nil
}

func (c *CardanoTxControllerImpl) createTx(requestBody request.CreateBridgingTxRequest) (
	string, string, error,
) {
	sourceChainConfig := c.oracleConfig.CardanoChains[requestBody.SourceChainID]

	var batcherChainConfig batcherCore.ChainConfig

	for _, batcherChain := range c.batcherConfig.Chains {
		if batcherChain.ChainID == requestBody.SourceChainID {
			batcherChainConfig = batcherChain

			break
		}
	}

	cardanoConfig, err := cardanotx.NewCardanoChainConfig(batcherChainConfig.ChainSpecific)
	if err != nil {
		return "", "", err
	}

	txProvider, err := cardanoConfig.CreateTxProvider()
	if err != nil {
		return "", "", fmt.Errorf("failed to create tx provider: %w", err)
	}

	amountSum := uint64(0)
	txs := make([]common.BridgingRequestMetadataTransaction, len(requestBody.Transactions))

	for i, tx := range requestBody.Transactions {
		txs[i] = common.BridgingRequestMetadataTransaction{
			Address: common.SplitString(tx.Addr, 40),
			Amount:  tx.Amount,
		}

		amountSum += tx.Amount
	}

	metadata, err := common.MarshalMetadata(common.MetadataEncodingTypeJSON, common.BridgingRequestMetadata{
		BridgingTxType:     common.BridgingTxTypeBridgingRequest,
		DestinationChainID: requestBody.DestinationChainID,
		SenderAddr:         common.SplitString(requestBody.SenderAddr, 40),
		Transactions:       txs,
	})
	if err != nil {
		return "", "", fmt.Errorf("failed to marshal metadata: %w", err)
	}

	protocolParams, err := txProvider.GetProtocolParameters(context.Background())
	if err != nil {
		return "", "", fmt.Errorf("failed to get protocol parameters: %w", err)
	}

	qtd, err := txProvider.GetTip(context.Background())
	if err != nil {
		return "", "", fmt.Errorf("failed to get tip: %w", err)
	}

	outputs := []wallet.TxOutput{
		{
			Addr:   sourceChainConfig.BridgingAddresses.BridgingAddress,
			Amount: amountSum,
		},
	}

	inputs, err := wallet.GetUTXOsForAmount(
		context.Background(), txProvider, requestBody.SenderAddr,
		amountSum+cardanoConfig.PotentialFee, wallet.MinUTxODefaultValue)
	if err != nil {
		return "", "", fmt.Errorf("failed to get utxos: %w", err)
	}

	outputsSum := wallet.GetOutputsSum(outputs)

	builder, err := wallet.NewTxBuilder()
	if err != nil {
		return "", "", fmt.Errorf("failed to create tx builder: %w", err)
	}

	defer builder.Dispose()

	if len(metadata) != 0 {
		builder.SetMetaData(metadata)
	}

	builder.SetProtocolParameters(protocolParams).SetTimeToLive(qtd.Slot + cardanoConfig.TTLSlotNumberInc).
		SetTestNetMagic(uint(sourceChainConfig.NetworkMagic)).
		AddInputs(inputs.Inputs...).
		AddOutputs(outputs...).AddOutputs(wallet.TxOutput{Addr: requestBody.SenderAddr})

	fee, err := builder.CalculateFee(0)
	if err != nil {
		return "", "", fmt.Errorf("failed to calculate fee: %w", err)
	}

	change := inputs.Sum - outputsSum - fee
	// handle overflow or insufficient amount
	if change > inputs.Sum || (change > 0 && change < wallet.MinUTxODefaultValue) {
		return "", "", fmt.Errorf("insufficient amount %d for %d or min utxo not satisfied",
			inputs.Sum, outputsSum+fee)
	}

	if change == 0 {
		builder.RemoveOutput(-1)
	} else {
		builder.UpdateOutputAmount(-1, change)
	}

	builder.SetFee(fee)

	txRawBytes, txHash, err := builder.Build()
	if err != nil {
		return "", "", fmt.Errorf("failed to build tx: %w", err)
	}

	txRaw := hex.EncodeToString(txRawBytes)

	return txRaw, txHash, nil
}

func (c *CardanoTxControllerImpl) signTx(requestBody request.SignBridgingTxRequest) (
	string, error,
) {
	signingKeyBytes, err := hex.DecodeString(requestBody.SigningKeyHex)
	if err != nil {
		return "", fmt.Errorf("failed to decode singing key hex: %w", err)
	}

	txRawBytes, err := hex.DecodeString(requestBody.TxRaw)
	if err != nil {
		return "", fmt.Errorf("failed to decode raw tx: %w", err)
	}

	verificationKeyBytes := wallet.GetVerificationKeyFromSigningKey(signingKeyBytes)

	keyHash, err := wallet.GetKeyHash(verificationKeyBytes)
	if err != nil {
		return "", fmt.Errorf("failed to get key hash: %w", err)
	}

	senderWallet := wallet.NewWallet(
		verificationKeyBytes,
		signingKeyBytes,
		keyHash,
	)

	signedTxBytes, err := wallet.SignTx(txRawBytes, requestBody.TxHash, senderWallet)
	if err != nil {
		return "", fmt.Errorf("failed to sign tx: %w", err)
	}

	signedTx := hex.EncodeToString(signedTxBytes)

	return signedTx, nil
}
