package controllers

import (
	"errors"
	"fmt"
	"net/http"

	apiCore "github.com/Ethernal-Tech/apex-bridge/api/core"
	apiUtils "github.com/Ethernal-Tech/apex-bridge/api/utils"
	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/validatorcomponents/api/model/response"
	"github.com/Ethernal-Tech/apex-bridge/validatorcomponents/core"
	"github.com/hashicorp/go-hclog"
)

type BridgingRequestStateControllerImpl struct {
	bridgingRequestStateManager core.BridgingRequestStateManager
	logger                      hclog.Logger
}

var _ apiCore.APIController = (*BridgingRequestStateControllerImpl)(nil)

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

func (c *BridgingRequestStateControllerImpl) GetEndpoints() []*apiCore.APIEndpoint {
	return []*apiCore.APIEndpoint{
		{Path: "Get", Method: http.MethodGet, Handler: c.get, APIKeyAuth: true},
		{Path: "GetMultiple", Method: http.MethodGet, Handler: c.getMultiple, APIKeyAuth: true},
	}
}

// @Summary Get state of bridging request
// @Description Returns the current status of a bridging request, along with the destination chain ID and transaction hash, based on the given source chain ID and transaction hash.
// @Tags BridgingRequestState
// @Produce json
// @Param chainId query string true "Source chain ID"
// @Param txHash query string true "Source transaction hash"
// @Success 200 {object} response.BridgingRequestStateResponse
// @Failure 400 {object} response.ErrorResponse "Bad Request – One or more query parameters are missing, or the bridging request state could not be retrieved."
// @Failure 401 {object} response.ErrorResponse "Unauthorized – API key missing or invalid."
// @Failure 404 {object} response.ErrorResponse "Not Found - No bridging request found for the given parameters."
// @Security ApiKeyAuth
// @Router /BridgingRequestState/Get [get]
func (c *BridgingRequestStateControllerImpl) get(w http.ResponseWriter, r *http.Request) {
	queryValues := r.URL.Query()
	c.logger.Debug("get request", "query values", queryValues, "url", r.URL)

	chainIDArr, exists := queryValues["chainId"]
	if !exists || len(chainIDArr) == 0 {
		apiUtils.WriteErrorResponse(
			w, r, http.StatusBadRequest,
			errors.New("chainId missing from query"), c.logger)

		return
	}

	txHashArr, exists := queryValues["txHash"]
	if !exists || len(txHashArr) == 0 {
		apiUtils.WriteErrorResponse(
			w, r, http.StatusBadRequest,
			errors.New("txHash missing from query"), c.logger)

		return
	}

	chainID := chainIDArr[0]
	txHash := common.NewHashFromHexString(txHashArr[0])

	state, err := c.bridgingRequestStateManager.Get(chainID, txHash)
	if err != nil {
		apiUtils.WriteErrorResponse(
			w, r, http.StatusBadRequest,
			fmt.Errorf("failed to get bridging request state: %w", err), c.logger)

		return
	}

	if state == nil {
		apiUtils.WriteErrorResponse(
			w, r, http.StatusNotFound,
			errors.New("not found"), c.logger)

		return
	}

	apiUtils.WriteResponse(w, r, http.StatusOK, response.NewBridgingRequestStateResponse(state), c.logger)
}

// @Summary Get states of multiple bridging requests
// @Description Returns statuses and related data for one or more bridging requests, based on the given source chain ID and transaction hash(es).
// @Tags BridgingRequestState
// @Produce json
// @Param chainId query string true "Source chain ID"
// @Param txHash query []string true "Source transaction hashes"
// @Success 200 {object} map[string]response.BridgingRequestStateResponse "OK – Returns a map with source transaction hashes as keys and the associated bridging request data as values."
// @Failure 400 {object} response.ErrorResponse "Bad Request – chainId is missing from the query or the bridging request states could not be retrieved."
// @Failure 401 {object} response.ErrorResponse "Unauthorized – API key missing or invalid."
// @Security ApiKeyAuth
// @Router /BridgingRequestState/GetMultiple [get]
func (c *BridgingRequestStateControllerImpl) getMultiple(w http.ResponseWriter, r *http.Request) {
	queryValues := r.URL.Query()
	c.logger.Debug("query values", queryValues, "url", r.URL)

	chainIDArr, exists := queryValues["chainId"]
	if !exists || len(chainIDArr) == 0 {
		apiUtils.WriteErrorResponse(
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
		apiUtils.WriteErrorResponse(
			w, r, http.StatusBadRequest,
			fmt.Errorf("failed to get bridging request states: %w", err), c.logger)

		return
	}

	statesResponse := make(map[string]*response.BridgingRequestStateResponse, len(states))
	for _, state := range states {
		statesResponse[state.SourceTxHash.String()] = response.NewBridgingRequestStateResponse(state)
	}

	apiUtils.WriteResponse(w, r, http.StatusOK, statesResponse, c.logger)
}
