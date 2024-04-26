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
) (
	*BridgingRequestStateControllerImpl, error,
) {
	return &BridgingRequestStateControllerImpl{
		bridgingRequestStateManager: bridgingRequestStateManager,
		logger:                      logger,
	}, nil
}

func (*BridgingRequestStateControllerImpl) GetPathPrefix() string {
	return "BridgingRequestState"
}

func (c *BridgingRequestStateControllerImpl) GetEndpoints() []*core.APIEndpoint {
	return []*core.APIEndpoint{
		{Path: "Get", Method: http.MethodGet, Handler: c.get, APIKeyAuth: true},
		{Path: "GetAllForUser", Method: http.MethodGet, Handler: c.getAllForUser, APIKeyAuth: true},
	}
}

func (c *BridgingRequestStateControllerImpl) get(w http.ResponseWriter, r *http.Request) {
	c.logger.Debug("get called", "url", r.URL)

	queryValues := r.URL.Query()
	c.logger.Debug("query values", queryValues, "url", r.URL)

	chainIDArr, exists := queryValues["chainId"]
	if !exists || len(chainIDArr) == 0 {
		c.logger.Debug("chainId missing from query", "url", r.URL)

		rerr := utils.WriteErrorResponse(w, http.StatusBadRequest, "chainId missing from query")
		if rerr != nil {
			c.logger.Error("error while WriteErrorResponse", "err", rerr)
		}

		return
	}

	txHashArr, exists := queryValues["txHash"]
	if !exists || len(txHashArr) == 0 {
		c.logger.Debug("txHash missing from query", "url", r.URL)

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
		c.logger.Debug(err.Error(), "url", r.URL)

		rerr := utils.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
		if rerr != nil {
			c.logger.Error("error while WriteErrorResponse", "err", rerr)
		}

		return
	}

	if state == nil {
		c.logger.Debug("Not found", "url", r.URL)

		rerr := utils.WriteErrorResponse(w, http.StatusNotFound, "Not found")
		if rerr != nil {
			c.logger.Error("error while WriteErrorResponse", "err", rerr)
		}

		return
	}

	c.logger.Debug("get success", "url", r.URL)

	err = json.NewEncoder(w).Encode(response.NewBridgingRequestStateResponse(state))
	if err != nil {
		c.logger.Error("error while writing response", "err", err)
	}
}

func (c *BridgingRequestStateControllerImpl) getAllForUser(w http.ResponseWriter, r *http.Request) {
	c.logger.Debug("getAllForUser called", "url", r.URL)

	queryValues := r.URL.Query()
	c.logger.Debug("query values", queryValues, "url", r.URL)

	chainIDArr, exists := queryValues["chainId"]
	if !exists || len(chainIDArr) == 0 {
		c.logger.Debug("chainId missing from query", "url", r.URL)

		rerr := utils.WriteErrorResponse(w, http.StatusBadRequest, "chainId missing from query")
		if rerr != nil {
			c.logger.Error("error while WriteErrorResponse", "err", rerr)
		}

		return
	}

	userAddrArr, exists := queryValues["userAddr"]
	if !exists || len(userAddrArr) == 0 {
		c.logger.Debug("userAddr missing from query", "url", r.URL)

		rerr := utils.WriteErrorResponse(w, http.StatusBadRequest, "userAddr missing from query")
		if rerr != nil {
			c.logger.Error("error while WriteErrorResponse", "err", rerr)
		}

		return
	}

	chainID := chainIDArr[0]
	userAddr := userAddrArr[0]

	states, err := c.bridgingRequestStateManager.GetAllForUser(chainID, userAddr)
	if err != nil {
		c.logger.Debug(err.Error(), "url", r.URL)

		rerr := utils.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
		if rerr != nil {
			c.logger.Error("error while WriteErrorResponse", "err", rerr)
		}

		return
	}

	statesResponse := make([]*response.BridgingRequestStateResponse, 0, len(states))
	for _, state := range states {
		statesResponse = append(statesResponse, response.NewBridgingRequestStateResponse(state))
	}

	c.logger.Debug("getAllForUser success", "url", r.URL)

	err = json.NewEncoder(w).Encode(statesResponse)
	if err != nil {
		c.logger.Error("error while writing response", "err", err)
	}
}
