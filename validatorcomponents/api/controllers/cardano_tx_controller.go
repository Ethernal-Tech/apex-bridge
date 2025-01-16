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
		if cardanoDestConfig != nil {
			requestBody.BridgingFee = cardanoDestConfig.MinFeeForBridging
		} else if ethDestConfig != nil {
			requestBody.BridgingFee = ethDestConfig.MinFeeForBridging
		}
	}

	receiverAmountSum.Add(receiverAmountSum, new(big.Int).SetUint64(requestBody.BridgingFee))

	if (cardanoDestConfig != nil && requestBody.BridgingFee < cardanoDestConfig.MinFeeForBridging) ||
		(ethDestConfig != nil && requestBody.BridgingFee < ethDestConfig.MinFeeForBridging) {
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

	sourceChainConfig := c.oracleConfig.CardanoChains[requestBody.SourceChainID]
	minAmountToBridge := uint64(0)

	destCardanoChainConfig, exists := c.oracleConfig.CardanoChains[requestBody.DestinationChainID]
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
			Addr:         tx.Addr,
			Amount:       tx.Amount,
			BridgingType: sendtx.BridgingTypeNormal,
		}
	}

	txRawBytes, txHash, _, err := txSender.CreateBridgingTx(
		context.Background(),
		requestBody.SourceChainID, requestBody.DestinationChainID,
		requestBody.SenderAddr, receivers, sendtx.NewExchangeRate())
	if err != nil {
		return "", "", fmt.Errorf("failed to build tx: %w", err)
	}

	txRaw := hex.EncodeToString(txRawBytes)

	return txRaw, txHash, nil
}
