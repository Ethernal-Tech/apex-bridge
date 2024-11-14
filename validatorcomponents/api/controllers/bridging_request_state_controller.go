package controllers

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/validatorcomponents/api/model/response"
	"github.com/Ethernal-Tech/apex-bridge/validatorcomponents/api/utils"
	"github.com/Ethernal-Tech/apex-bridge/validatorcomponents/core"
	"github.com/hashicorp/go-hclog"
)

type BridgingRequestStateControllerImpl struct {
	bridgingRequestStateManager core.BridgingRequestStateManager
	logger                      hclog.Logger
}

var _ core.APIController = (*BridgingRequestStateControllerImpl)(nil)

func NewBridgingRequestStateController(
	bridgingRequestStateManager core.BridgingRequestStateManager, logger hclog.Logger,
) *BridgingRequestStateControllerImpl {
	return &BridgingRequestStateControllerImpl{
		bridgingRequestStateManager: bridgingRequestStateManager,
		logger:                      logger,
	}
}

func (*BridgingRequestStateControllerImpl) GetPathPrefix() string {
	return "BridgingRequestState"
}

func (c *BridgingRequestStateControllerImpl) GetEndpoints() []*core.APIEndpoint {
	return []*core.APIEndpoint{
		{Path: "Get", Method: http.MethodGet, Handler: c.get, APIKeyAuth: true},
		{Path: "GetMultiple", Method: http.MethodGet, Handler: c.getMultiple, APIKeyAuth: true},
	}
}

func (c *BridgingRequestStateControllerImpl) get(w http.ResponseWriter, r *http.Request) {
	queryValues := r.URL.Query()
	c.logger.Debug("get request", "query values", queryValues, "url", r.URL)

	chainIDArr, exists := queryValues["chainId"]
	if !exists || len(chainIDArr) == 0 {
		utils.WriteErrorResponse(
			w, r, http.StatusBadRequest,
			errors.New("chainId missing from query"), c.logger)

		return
	}

	txHashArr, exists := queryValues["txHash"]
	if !exists || len(txHashArr) == 0 {
		utils.WriteErrorResponse(
			w, r, http.StatusBadRequest,
			errors.New("txHash missing from query"), c.logger)

		return
	}

	chainID := chainIDArr[0]
	txHash := common.NewHashFromHexString(txHashArr[0])

	state, err := c.bridgingRequestStateManager.Get(chainID, txHash)
	if err != nil {
		utils.WriteErrorResponse(
			w, r, http.StatusBadRequest,
			fmt.Errorf("failed to get bridging request state: %w", err), c.logger)

		return
	}

	if state == nil {
		utils.WriteErrorResponse(
			w, r, http.StatusNotFound,
			errors.New("not found"), c.logger)

		return
	}

	utils.WriteResponse(w, r, http.StatusOK, response.NewBridgingRequestStateResponse(state), c.logger)
}

func (c *BridgingRequestStateControllerImpl) getMultiple(w http.ResponseWriter, r *http.Request) {
	queryValues := r.URL.Query()
	c.logger.Debug("query values", queryValues, "url", r.URL)

	chainIDArr, exists := queryValues["chainId"]
	if !exists || len(chainIDArr) == 0 {
		utils.WriteErrorResponse(
			w, r, http.StatusBadRequest,
			errors.New("chainId missing from query"), c.logger)

		return
	}

	chainID := chainIDArr[0]

	txHashesStrs := queryValues["txHash"]
	txHashes := make([]common.Hash, len(txHashesStrs))

	for i, x := range txHashesStrs {
		txHashes[i] = common.NewHashFromHexString(x)
	}

	states, err := c.bridgingRequestStateManager.GetMultiple(chainID, txHashes)
	if err != nil {
		utils.WriteErrorResponse(
			w, r, http.StatusBadRequest,
			fmt.Errorf("failed to get bridging request states: %w", err), c.logger)

		return
	}

	statesResponse := make(map[string]*response.BridgingRequestStateResponse, len(states))
	for _, state := range states {
		statesResponse[state.SourceTxHash.String()] = response.NewBridgingRequestStateResponse(state)
	}

	utils.WriteResponse(w, r, http.StatusOK, statesResponse, c.logger)
}
