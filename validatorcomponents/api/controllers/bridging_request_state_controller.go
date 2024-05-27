package controllers

import (
	"encoding/json"
	"net/http"

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
	c.logger.Debug("get called", "url", r.URL)

	queryValues := r.URL.Query()
	c.logger.Debug("get request", "query values", queryValues, "url", r.URL)

	chainIDArr, exists := queryValues["chainId"]
	if !exists || len(chainIDArr) == 0 {
		c.logger.Debug("get request", "err", "chainId missing from query", "url", r.URL)

		rerr := utils.WriteErrorResponse(w, http.StatusBadRequest, "chainId missing from query")
		if rerr != nil {
			c.logger.Error("error while WriteErrorResponse", "err", rerr)
		}

		return
	}

	txHashArr, exists := queryValues["txHash"]
	if !exists || len(txHashArr) == 0 {
		c.logger.Debug("get request", "err", "txHash missing from query", "url", r.URL)

		rerr := utils.WriteErrorResponse(w, http.StatusBadRequest, "txHash missing from query")
		if rerr != nil {
			c.logger.Error("error while WriteErrorResponse", "err", rerr)
		}

		return
	}

	chainID := chainIDArr[0]
	txHash := txHashArr[0]

	state, err := c.bridgingRequestStateManager.Get(chainID, txHash)
	if err != nil {
		c.logger.Debug("get request", "err", err.Error(), "url", r.URL)

		rerr := utils.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
		if rerr != nil {
			c.logger.Error("error while WriteErrorResponse", "err", rerr)
		}

		return
	}

	if state == nil {
		c.logger.Debug("get request - Not found", "url", r.URL)

		rerr := utils.WriteErrorResponse(w, http.StatusNotFound, "Not found")
		if rerr != nil {
			c.logger.Error("error while WriteErrorResponse", "err", rerr)
		}

		return
	}

	c.logger.Debug("get success", "url", r.URL)

	w.Header().Set("Content-Type", "application/json")

	err = json.NewEncoder(w).Encode(response.NewBridgingRequestStateResponse(state))
	if err != nil {
		c.logger.Error("error while writing response", "err", err)
	}
}

func (c *BridgingRequestStateControllerImpl) getMultiple(w http.ResponseWriter, r *http.Request) {
	c.logger.Debug("getMultiple called", "url", r.URL)

	queryValues := r.URL.Query()
	c.logger.Debug("query values", queryValues, "url", r.URL)

	chainIDArr, exists := queryValues["chainId"]
	if !exists || len(chainIDArr) == 0 {
		c.logger.Debug("getMultiple request", "err", "chainId missing from query", "url", r.URL)

		rerr := utils.WriteErrorResponse(w, http.StatusBadRequest, "chainId missing from query")
		if rerr != nil {
			c.logger.Error("error while WriteErrorResponse", "err", rerr)
		}

		return
	}

	chainID := chainIDArr[0]
	txHashes := queryValues["txHash"]

	states, err := c.bridgingRequestStateManager.GetMultiple(chainID, txHashes)
	if err != nil {
		c.logger.Debug("getMultiple request", "err", err.Error(), "url", r.URL)

		rerr := utils.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
		if rerr != nil {
			c.logger.Error("error while WriteErrorResponse", "err", rerr)
		}

		return
	}

	statesResponse := make(map[string]*response.BridgingRequestStateResponse, len(states))
	for _, state := range states {
		statesResponse[state.SourceTxHash] = response.NewBridgingRequestStateResponse(state)
	}

	c.logger.Debug("getMultiple success", "url", r.URL)

	w.Header().Set("Content-Type", "application/json")

	err = json.NewEncoder(w).Encode(statesResponse)
	if err != nil {
		c.logger.Error("error while writing response", "err", err)
	}
}
