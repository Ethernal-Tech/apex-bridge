package controllers

import (
	"context"
	"fmt"
	"math/big"
	"net/http"

	apiCore "github.com/Ethernal-Tech/apex-bridge/api/core"
	apiUtils "github.com/Ethernal-Tech/apex-bridge/api/utils"
	batcherCore "github.com/Ethernal-Tech/apex-bridge/batcher/core"
	"github.com/Ethernal-Tech/apex-bridge/common"
	oCore "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	oUtils "github.com/Ethernal-Tech/apex-bridge/oracle_common/utils"
	"github.com/Ethernal-Tech/apex-bridge/validatorcomponents/api/model/request"
	"github.com/Ethernal-Tech/apex-bridge/validatorcomponents/api/model/response"
	"github.com/Ethernal-Tech/cardano-infrastructure/sendtx"
	"github.com/Ethernal-Tech/cardano-infrastructure/wallet"
	goEthCommon "github.com/ethereum/go-ethereum/common"
	"github.com/hashicorp/go-hclog"
)

type ReactorTxControllerImpl struct {
	oracleConfig  *oCore.AppConfig
	batcherConfig *batcherCore.BatcherManagerConfiguration
	logger        hclog.Logger
}

var _ apiCore.APIController = (*ReactorTxControllerImpl)(nil)

func NewReactorTxController(
	oracleConfig *oCore.AppConfig,
	batcherConfig *batcherCore.BatcherManagerConfiguration,
	logger hclog.Logger,
) *ReactorTxControllerImpl {
	return &ReactorTxControllerImpl{
		oracleConfig:  oracleConfig,
		batcherConfig: batcherConfig,
		logger:        logger,
	}
}

func (*ReactorTxControllerImpl) GetPathPrefix() string {
	return "CardanoTx"
}

func (c *ReactorTxControllerImpl) GetEndpoints() []*apiCore.APIEndpoint {
	return []*apiCore.APIEndpoint{
		{Path: "CreateBridgingTx", Method: http.MethodPost, Handler: c.createBridgingTx, APIKeyAuth: true},
	}
}

func (c *ReactorTxControllerImpl) createBridgingTx(w http.ResponseWriter, r *http.Request) {
	requestBody, ok := apiUtils.DecodeModel[request.CreateBridgingTxRequest](w, r, c.logger)
	if !ok {
		return
	}

	c.logger.Debug("createBridgingTx request", "body", requestBody, "url", r.URL)

	err := c.validateAndFillOutCreateBridgingTxRequest(&requestBody)
	if err != nil {
		apiUtils.WriteErrorResponse(
			w, r, http.StatusBadRequest,
			fmt.Errorf("validation error. err: %w", err), c.logger)

		return
	}

	txInfo, err := c.createTx(r.Context(), requestBody)
	if err != nil {
		apiUtils.WriteErrorResponse(w, r, http.StatusInternalServerError, err, c.logger)

		return
	}

	var amount uint64
	for _, transaction := range requestBody.Transactions {
		amount += transaction.Amount
	}

	apiUtils.WriteResponse(
		w, r, http.StatusOK,
		response.NewBridgingTxResponse(txInfo.TxRaw, txInfo.TxHash, requestBody.BridgingFee, amount, 0), c.logger)
}

func (c *ReactorTxControllerImpl) validateAndFillOutCreateBridgingTxRequest(
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

	cardanoDestChainFeeAddress := c.oracleConfig.GetFeeMultisigAddress(requestBody.DestinationChainID)

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
			if receiver.Addr == cardanoDestChainFeeAddress {
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
		requestBody.BridgingFee = cardanoSrcConfig.MinFeeForBridging
	}

	receiverAmountSum.Add(receiverAmountSum, new(big.Int).SetUint64(requestBody.BridgingFee))

	if requestBody.BridgingFee < cardanoSrcConfig.MinFeeForBridging {
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

func (c *ReactorTxControllerImpl) createTx(ctx context.Context, requestBody request.CreateBridgingTxRequest) (
	*sendtx.TxInfo, error,
) {
	txSenderChainsConfig, err := c.oracleConfig.ToSendTxChainConfigs()
	if err != nil {
		return nil, fmt.Errorf("failed to create configuration")
	}

	txSender := sendtx.NewTxSender(txSenderChainsConfig)

	receivers := make([]sendtx.BridgingTxReceiver, len(requestBody.Transactions))
	for i, tx := range requestBody.Transactions {
		receivers[i] = sendtx.BridgingTxReceiver{
			Addr:         tx.Addr,
			Amount:       tx.Amount,
			BridgingType: sendtx.BridgingTypeNormal,
		}
	}

	txInfo, _, err := txSender.CreateBridgingTx(
		ctx,
		sendtx.BridgingTxDto{
			SrcChainID:             requestBody.SourceChainID,
			DstChainID:             requestBody.DestinationChainID,
			SenderAddr:             requestBody.SenderAddr,
			SenderAddrPolicyScript: requestBody.SenderAddrPolicyScript,
			Receivers:              receivers,
			BridgingAddress:        requestBody.BridgingAddress,
			BridgingFee:            requestBody.BridgingFee,
			OperationFee:           0,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to build tx: %w", err)
	}

	return txInfo, nil
}
