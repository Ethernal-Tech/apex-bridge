package controllers

import (
	"context"
	"encoding/hex"
	"errors"
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
	"github.com/Ethernal-Tech/cardano-infrastructure/wallet"
	goEthCommon "github.com/ethereum/go-ethereum/common"
	"github.com/hashicorp/go-hclog"
)

type CardanoTxControllerImpl struct {
	oracleConfig  *oCore.AppConfig
	batcherConfig *batcherCore.BatcherManagerConfiguration
	logger        hclog.Logger
}

var _ core.APIController = (*CardanoTxControllerImpl)(nil)

func NewCardanoTxController(
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
	requestBody, ok := utils.DecodeModel[request.CreateBridgingTxRequest](w, r, c.logger)
	if !ok {
		return
	}

	c.logger.Debug("createBridgingTx request", "body", requestBody, "url", r.URL)

	err := c.validateAndFillOutCreateBridgingTxRequest(&requestBody)
	if err != nil {
		utils.WriteErrorResponse(
			w, r, http.StatusBadRequest,
			fmt.Errorf("validation error. err: %w", err), c.logger)

		return
	}

	txRaw, txHash, err := c.createTx(requestBody)
	if err != nil {
		utils.WriteErrorResponse(w, r, http.StatusInternalServerError, err, c.logger)

		return
	}

	utils.WriteResponse(
		w, r, http.StatusOK,
		response.NewFullBridgingTxResponse(txRaw, txHash, requestBody.BridgingFee), c.logger)
}

func (c *CardanoTxControllerImpl) signBridgingTx(w http.ResponseWriter, r *http.Request) {
	requestBody, ok := utils.DecodeModel[request.SignBridgingTxRequest](w, r, c.logger)
	if !ok {
		return
	}

	if requestBody.TxRaw == "" || requestBody.SigningKeyHex == "" || requestBody.TxHash == "" {
		utils.WriteErrorResponse(
			w, r, http.StatusBadRequest, errors.New("invalid input data"), c.logger)

		return
	}

	c.logger.Debug("signBridgingTx request", "body", requestBody, "url", r.URL)

	signedTx, err := c.signTx(requestBody)
	if err != nil {
		utils.WriteErrorResponse(
			w, r, http.StatusBadRequest, errors.New("validation error"), c.logger)

		return
	}

	utils.WriteResponse(
		w, r, http.StatusOK,
		response.NewBridgingTxResponse(signedTx, requestBody.TxHash), c.logger)
}

func (c *CardanoTxControllerImpl) validateAndFillOutCreateBridgingTxRequest(
	requestBody *request.CreateBridgingTxRequest,
) error {
	cardanoSrcConfig, _ := oUtils.GetChainConfig(c.oracleConfig, requestBody.SourceChainID)
	if cardanoSrcConfig == nil {
		return fmt.Errorf("origin chain not registered: %v", requestBody.SourceChainID)
	}

	cardanoDestConfig, ethDestConfig := oUtils.GetChainConfig(c.oracleConfig, requestBody.DestinationChainID)
	if cardanoDestConfig == nil && ethDestConfig == nil {
		return fmt.Errorf("destination chain not registered: %v", requestBody.DestinationChainID)
	}

	if len(requestBody.Transactions) > c.oracleConfig.BridgingSettings.MaxReceiversPerBridgingRequest {
		return fmt.Errorf("number of receivers in metadata greater than maximum allowed - no: %v, max: %v, requestBody: %v",
			len(requestBody.Transactions), c.oracleConfig.BridgingSettings.MaxReceiversPerBridgingRequest, requestBody)
	}

	receiverAmountSum := big.NewInt(0)
	feeSum := uint64(0)
	foundAUtxoValueBelowMinimumValue := false
	foundAnInvalidReceiverAddr := false
	transactions := make([]request.CreateBridgingTxTransactionRequest, 0, len(requestBody.Transactions))

	for _, receiver := range requestBody.Transactions {
		if cardanoDestConfig != nil {
			if receiver.Amount < c.oracleConfig.BridgingSettings.UtxoMinValue {
				foundAUtxoValueBelowMinimumValue = true

				break
			}

			addr, err := wallet.NewAddress(receiver.Addr)
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
		} else if ethDestConfig != nil {
			if !goEthCommon.IsHexAddress(receiver.Addr) {
				foundAnInvalidReceiverAddr = true

				break
			}

			if receiver.Addr == common.EthZeroAddr {
				feeSum += receiver.Amount
			} else {
				transactions = append(transactions, receiver)
				receiverAmountSum.Add(receiverAmountSum, new(big.Int).SetUint64(receiver.Amount))
			}
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
		requestBody.BridgingFee = c.oracleConfig.BridgingSettings.MinFeeForBridging
	}

	receiverAmountSum.Add(receiverAmountSum, new(big.Int).SetUint64(requestBody.BridgingFee))

	if requestBody.BridgingFee < c.oracleConfig.BridgingSettings.MinFeeForBridging {
		return fmt.Errorf("bridging fee in request body is less than minimum: %v", requestBody)
	}

	if c.oracleConfig.BridgingSettings.MaxAmountAllowedToBridge != nil &&
		c.oracleConfig.BridgingSettings.MaxAmountAllowedToBridge.Sign() > 0 &&
		receiverAmountSum.Cmp(c.oracleConfig.BridgingSettings.MaxAmountAllowedToBridge) == 1 {
		return fmt.Errorf("sum of receiver amounts + fee greater than maximum allowed: %v, for request: %v",
			c.oracleConfig.BridgingSettings.MaxAmountAllowedToBridge, requestBody)
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

	bridgingTxSender := cardanotx.NewBridgingTxSender(
		wallet.ResolveCardanoCliBinary(sourceChainConfig.NetworkID),
		txProvider, nil, uint(sourceChainConfig.NetworkMagic),
		sourceChainConfig.BridgingAddresses.BridgingAddress,
		cardanoConfig.TTLSlotNumberInc, cardanoConfig.PotentialFee,
	)

	receivers := make([]wallet.TxOutput, len(requestBody.Transactions))
	for i, tx := range requestBody.Transactions {
		receivers[i] = wallet.TxOutput{
			Addr:   tx.Addr,
			Amount: tx.Amount,
		}
	}

	txRawBytes, txHash, err := bridgingTxSender.CreateTx(
		context.Background(), requestBody.DestinationChainID,
		requestBody.SenderAddr, receivers, requestBody.BridgingFee,
		c.oracleConfig.BridgingSettings.UtxoMinValue,
	)
	if err != nil {
		return "", "", fmt.Errorf("failed to build tx: %w", err)
	}

	txRaw := hex.EncodeToString(txRawBytes)

	return txRaw, txHash, nil
}

func (c *CardanoTxControllerImpl) signTx(requestBody request.SignBridgingTxRequest) (
	string, error,
) {
	signingKeyBytes, err := common.DecodeHex(requestBody.SigningKeyHex)
	if err != nil {
		return "", fmt.Errorf("failed to decode singing key hex: %w", err)
	}

	txRawBytes, err := common.DecodeHex(requestBody.TxRaw)
	if err != nil {
		return "", fmt.Errorf("failed to decode raw tx: %w", err)
	}

	cardanoCliBinary := wallet.ResolveCardanoCliBinary(wallet.CardanoNetworkType(requestBody.NetworkID))
	senderWallet := wallet.NewWallet(
		wallet.GetVerificationKeyFromSigningKey(signingKeyBytes), signingKeyBytes)

	txBuilder, err := wallet.NewTxBuilder(cardanoCliBinary)
	if err != nil {
		return "", fmt.Errorf("failed to create tx builder: %w", err)
	}

	defer txBuilder.Dispose()

	signedTxBytes, err := txBuilder.SignTx(txRawBytes, []wallet.ITxSigner{senderWallet})
	if err != nil {
		return "", fmt.Errorf("failed to sign tx: %w", err)
	}

	return hex.EncodeToString(signedTxBytes), nil
}
