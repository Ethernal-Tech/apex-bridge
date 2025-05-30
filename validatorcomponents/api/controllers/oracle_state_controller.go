package controllers

import (
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"strings"

	"github.com/Ethernal-Tech/apex-bridge/common"
	oCore "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	"github.com/Ethernal-Tech/apex-bridge/validatorcomponents/api/model/response"
	"github.com/Ethernal-Tech/apex-bridge/validatorcomponents/api/utils"
	"github.com/Ethernal-Tech/apex-bridge/validatorcomponents/core"
	vcUtils "github.com/Ethernal-Tech/apex-bridge/validatorcomponents/utils"
	eventTrackerStore "github.com/Ethernal-Tech/blockchain-event-tracker/store"
	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
	"github.com/hashicorp/go-hclog"
)

type OracleStateControllerImpl struct {
	appConfig                   *core.AppConfig
	bridgingRequestStateManager core.BridgingRequestStateManager
	cardanoIndexerDBs           map[string]indexer.Database
	ethIndexerDBs               map[string]eventTrackerStore.EventTrackerStore
	adressesMap                 map[string][]string
	logger                      hclog.Logger
}

var _ core.APIController = (*OracleStateControllerImpl)(nil)

func NewOracleStateController(
	appConfig *core.AppConfig,
	bridgingRequestStateManager core.BridgingRequestStateManager,
	cardanoIndexerDBs map[string]indexer.Database,
	ethIndexerDBs map[string]eventTrackerStore.EventTrackerStore,
	adressesMap map[string][]string,
	logger hclog.Logger,
) *OracleStateControllerImpl {
	return &OracleStateControllerImpl{
		appConfig:                   appConfig,
		bridgingRequestStateManager: bridgingRequestStateManager,
		cardanoIndexerDBs:           cardanoIndexerDBs,
		ethIndexerDBs:               ethIndexerDBs,
		adressesMap:                 adressesMap,
		logger:                      logger,
	}
}

func (*OracleStateControllerImpl) GetPathPrefix() string {
	return "OracleState"
}

func (c *OracleStateControllerImpl) GetEndpoints() []*core.APIEndpoint {
	return []*core.APIEndpoint{
		{Path: "Get", Method: http.MethodGet, Handler: c.getState, APIKeyAuth: true},
		{Path: "GetHasTxFailed", Method: http.MethodGet, Handler: c.getHasTxFailed, APIKeyAuth: true},
	}
}

func (c *OracleStateControllerImpl) getState(w http.ResponseWriter, r *http.Request) {
	queryValues := r.URL.Query()

	chainIDArr, exists := queryValues["chainId"]
	if !exists || len(chainIDArr) == 0 {
		utils.WriteErrorResponse(
			w, r, http.StatusBadRequest,
			errors.New("chainId missing from query"), c.logger)

		return
	}

	chainID := chainIDArr[0]

	db, existsDB := c.cardanoIndexerDBs[chainID]
	addresses, existsAddrs := c.adressesMap[chainID]

	if !existsDB || !existsAddrs {
		utils.WriteErrorResponse(
			w, r, http.StatusBadRequest,
			fmt.Errorf("invalid chainID: %s", chainID), c.logger)

		return
	}

	latestBlockPoint, err := db.GetLatestBlockPoint()
	if err != nil {
		utils.WriteErrorResponse(
			w, r, http.StatusBadRequest,
			fmt.Errorf("get latest point: %w", err), c.logger)

		return
	}

	addressesUtxos := make([][]*indexer.TxInputOutput, len(addresses))
	count := 0

	for i, addr := range addresses {
		utxos, err := db.GetAllTxOutputs(addr, true)
		if err != nil {
			utils.WriteErrorResponse(
				w, r, http.StatusBadRequest,
				fmt.Errorf("get all tx outputs: %w", err), c.logger)

			return
		}

		addressesUtxos[i] = utxos
		count += len(utxos)
	}

	outputUtxos := make([]oCore.CardanoChainConfigUtxo, 0, count)

	for _, utxos := range addressesUtxos {
		for _, inp := range utxos {
			outputUtxos = append(outputUtxos, oCore.CardanoChainConfigUtxo{
				Hash:    inp.Input.Hash,
				Index:   inp.Input.Index,
				Address: inp.Output.Address,
				Amount:  inp.Output.Amount,
				Slot:    inp.Output.Slot,
			})
		}
	}

	utils.WriteResponse(w, r, http.StatusOK, response.NewOracleStateResponse(
		chainID, outputUtxos, latestBlockPoint.BlockSlot, latestBlockPoint.BlockHash), c.logger)
}

func (c *OracleStateControllerImpl) getHasTxFailed(w http.ResponseWriter, r *http.Request) {
	queryValues := r.URL.Query()

	chainIDArr, exists := queryValues["chainId"]
	if !exists || len(chainIDArr) == 0 {
		utils.WriteErrorResponse(
			w, r, http.StatusBadRequest,
			errors.New("chainId missing from query"), c.logger)

		return
	}

	chainID := chainIDArr[0]

	txHashArr, exists := queryValues["txHash"]
	if !exists || len(txHashArr) == 0 {
		utils.WriteErrorResponse(
			w, r, http.StatusBadRequest,
			errors.New("txHash missing from query"), c.logger)

		return
	}

	txHash := strings.TrimPrefix(txHashArr[0], "0x")

	ttlArr, exists := queryValues["ttl"]
	if !exists || len(ttlArr) == 0 {
		utils.WriteErrorResponse(
			w, r, http.StatusBadRequest,
			errors.New("ttl missing from query"), c.logger)

		return
	}

	ttl, ok := new(big.Int).SetString(ttlArr[0], 10)
	if !ok {
		utils.WriteErrorResponse(
			w, r, http.StatusBadRequest,
			errors.New("ttl invalid"), c.logger)

		return
	}

	hasFailed, err := c.hasTxFailed(chainID, txHash, ttl)
	if err != nil {
		utils.WriteErrorResponse(
			w, r, http.StatusBadRequest,
			fmt.Errorf("hasTxFailed err: %w", err), c.logger)

		return
	}

	utils.WriteResponse(w, r, http.StatusOK, response.HasTxFailedResponse{Failed: hasFailed}, c.logger)
}

func (c *OracleStateControllerImpl) hasTxFailed(
	chainID string, txHash string, ttl *big.Int,
) (bool, error) {
	cardanoConfig, ethConfig := vcUtils.GetChainConfig(c.appConfig, chainID)
	if cardanoConfig == nil && ethConfig == nil {
		return false, fmt.Errorf("invalid chainID: %s", chainID)
	}

	findTxFunc := c.findCardanoTx
	passedTTLFunc := c.passedCardanoTTL

	if ethConfig != nil {
		findTxFunc = c.findEthTx
		passedTTLFunc = c.passedEthTTL
	}

	var (
		foundTx, passedTTL bool
		err                error
	)

	foundTx, err = findTxFunc(chainID, txHash)
	if err != nil {
		return false, err
	}

	if !foundTx {
		passedTTL, err = passedTTLFunc(chainID, ttl)
		if err != nil {
			return false, err
		}
	}

	return !foundTx && passedTTL, nil
}

func (c *OracleStateControllerImpl) passedEthTTL(chainID string, ttl *big.Int) (bool, error) {
	db, existsDB := c.ethIndexerDBs[chainID]
	if !existsDB {
		return false, fmt.Errorf("couldn't find indexer db")
	}

	block, err := db.GetLastProcessedBlock()
	if err != nil {
		return false, fmt.Errorf("couldn't fetch indexer latest block point. err: %w", err)
	}

	return new(big.Int).SetUint64(block).Cmp(ttl) == 1, nil
}

func (c *OracleStateControllerImpl) findEthTx(chainID string, txHash string) (bool, error) {
	state, err := c.findBridgingRequestState(chainID, txHash)
	if err != nil {
		return false, fmt.Errorf("failed to find bridging request state. err: %w", err)
	}

	return state != nil, nil
}

func (c *OracleStateControllerImpl) passedCardanoTTL(chainID string, ttl *big.Int) (bool, error) {
	db, existsDB := c.cardanoIndexerDBs[chainID]
	if !existsDB {
		return false, fmt.Errorf("couldn't find indexer db")
	}

	block, err := db.GetLatestBlockPoint()
	if err != nil {
		return false, fmt.Errorf("couldn't fetch indexer latest block point. err: %w", err)
	}

	return new(big.Int).SetUint64(block.BlockSlot).Cmp(ttl) == 1, nil
}

func (c *OracleStateControllerImpl) findCardanoTx(chainID string, txHash string) (bool, error) {
	db, existsDB := c.cardanoIndexerDBs[chainID]
	if !existsDB {
		return false, fmt.Errorf("couldn't find indexer db")
	}

	indexerTxs, err := db.GetUnprocessedConfirmedTxs(0)
	if err != nil {
		return false, fmt.Errorf("couldn't fetch indexer txs. err: %w", err)
	}

	for _, tx := range indexerTxs {
		hashStr := hex.EncodeToString(tx.Hash[:])
		if txHash == hashStr {
			return true, nil
		}
	}

	state, err := c.findBridgingRequestState(chainID, txHash)
	if err != nil {
		return false, fmt.Errorf("failed to find bridging request state. err: %w", err)
	}

	return state != nil, nil
}

func (c *OracleStateControllerImpl) findBridgingRequestState(
	chainID string, txHash string,
) (*common.BridgingRequestState, error) {
	hashBytes, err := hex.DecodeString(txHash)
	if err != nil {
		return nil, fmt.Errorf("failed to decode txHash string. err: %w", err)
	}

	if len(hashBytes) != common.HashSize {
		return nil, fmt.Errorf("txHash invalid length. len: %d", len(hashBytes))
	}

	state, err := c.bridgingRequestStateManager.Get(chainID, common.Hash(hashBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to get bridging request state. err: %w", err)
	}

	return state, nil
}
