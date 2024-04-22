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

var _ core.ApiController = (*BridgingRequestStateControllerImpl)(nil)

func NewBridgingRequestStateController(bridgingRequestStateManager core.BridgingRequestStateManager, logger hclog.Logger) (*BridgingRequestStateControllerImpl, error) {
	return &BridgingRequestStateControllerImpl{
		bridgingRequestStateManager: bridgingRequestStateManager,
		logger:                      logger,
	}, nil
}

func (*BridgingRequestStateControllerImpl) GetPathPrefix() string {
	return "BridgingRequestState"
}

func (c *BridgingRequestStateControllerImpl) GetEndpoints() []*core.ApiEndpoint {
	return []*core.ApiEndpoint{
		{Path: "Get", Method: http.MethodGet, Handler: c.get, ApiKeyAuth: true},
		{Path: "GetAllForUser", Method: http.MethodGet, Handler: c.getAllForUser, ApiKeyAuth: true},
	}
}

func (c *BridgingRequestStateControllerImpl) get(w http.ResponseWriter, r *http.Request) {
	c.logger.Debug("get called", "url", r.URL)

	queryValues := r.URL.Query()
	c.logger.Debug("query values", queryValues, "url", r.URL)

	chainIdArr, exists := queryValues["chainId"]
	if !exists || len(chainIdArr) == 0 {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "chainId missing from query")
		c.logger.Debug("chainId missing from query", "url", r.URL)
		return
	}
	txHashArr, exists := queryValues["txHash"]
	if !exists || len(txHashArr) == 0 {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "txHash missing from query")
		c.logger.Debug("txHash missing from query", "url", r.URL)
		return
	}

	chainId := chainIdArr[0]
	txHash := txHashArr[0]

	state, err := c.bridgingRequestStateManager.Get(chainId, txHash)
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
		c.logger.Debug(err.Error(), "url", r.URL)
		return
	}

	if state == nil {
		utils.WriteErrorResponse(w, http.StatusNotFound, "Not found")
		c.logger.Debug("Not found", "url", r.URL)
		return
	}

	c.logger.Debug("get success", "url", r.URL)
	json.NewEncoder(w).Encode(response.NewBridgingRequestStateResponse(state))
}

func (c *BridgingRequestStateControllerImpl) getAllForUser(w http.ResponseWriter, r *http.Request) {
	c.logger.Debug("getAllForUser called", "url", r.URL)

	queryValues := r.URL.Query()
	c.logger.Debug("query values", queryValues, "url", r.URL)

	chainIdArr, exists := queryValues["chainId"]
	if !exists || len(chainIdArr) == 0 {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "chainId missing from query")
		c.logger.Debug("chainId missing from query", "url", r.URL)
		return
	}
	userAddrArr, exists := queryValues["userAddr"]
	if !exists || len(userAddrArr) == 0 {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "userAddr missing from query")
		c.logger.Debug("userAddr missing from query", "url", r.URL)
		return
	}

	chainId := chainIdArr[0]
	userAddr := userAddrArr[0]

	states, err := c.bridgingRequestStateManager.GetAllForUser(chainId, userAddr)
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
		c.logger.Debug(err.Error(), "url", r.URL)
		return
	}

	statesResponse := make([]*response.BridgingRequestStateResponse, 0, len(states))
	for _, state := range states {
		statesResponse = append(statesResponse, response.NewBridgingRequestStateResponse(state))
	}

	c.logger.Debug("getAllForUser success", "url", r.URL)
	json.NewEncoder(w).Encode(statesResponse)
}
