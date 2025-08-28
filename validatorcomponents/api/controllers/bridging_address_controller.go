package controllers

import (
	"errors"
	"fmt"
	"net/http"

	apiCore "github.com/Ethernal-Tech/apex-bridge/api/core"
	apiUtils "github.com/Ethernal-Tech/apex-bridge/api/utils"
	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/validatorcomponents/api/model/response"
	"github.com/hashicorp/go-hclog"
)

type BridgingAddressControllerImpl struct {
	bridgingAddressesCoordinator common.BridgingAddressesCoordinator
	bridgingAddressManager       common.BridgingAddressesManager
	logger                       hclog.Logger
}

var _ apiCore.APIController = (*BridgingAddressControllerImpl)(nil)

func NewBridgingAddressController(
	bridgingAddressesCoordinator common.BridgingAddressesCoordinator,
	bridgingAddressManager common.BridgingAddressesManager,
	logger hclog.Logger,
) *BridgingAddressControllerImpl {
	return &BridgingAddressControllerImpl{
		bridgingAddressesCoordinator: bridgingAddressesCoordinator,
		bridgingAddressManager:       bridgingAddressManager,
		logger:                       logger,
	}
}

func (*BridgingAddressControllerImpl) GetPathPrefix() string {
	return "BridgingAddress"
}

func (c *BridgingAddressControllerImpl) GetEndpoints() []*apiCore.APIEndpoint {
	return []*apiCore.APIEndpoint{
		{Path: "Get", Method: http.MethodGet, Handler: c.getBridgingAddress, APIKeyAuth: true},
		{Path: "GetAllAddresses", Method: http.MethodGet, Handler: c.getAllBridgingAddresses, APIKeyAuth: true},
	}
}

func (c *BridgingAddressControllerImpl) getBridgingAddress(w http.ResponseWriter, r *http.Request) {
	queryValues := r.URL.Query()
	c.logger.Debug("getBridgingAddress request", "query values", queryValues, "url", r.URL)

	chainIDArr, exists := queryValues["chainId"]
	if !exists || len(chainIDArr) == 0 {
		apiUtils.WriteErrorResponse(
			w, r, http.StatusBadRequest,
			errors.New("chainId missing from query"), c.logger)

		return
	}

	chainIDStr := chainIDArr[0]
	chainID := common.ToNumChainID(chainIDStr)

	bridgingAddress, err := c.bridgingAddressesCoordinator.GetAddressToBridgeTo(chainID)
	if err != nil {
		apiUtils.WriteErrorResponse(
			w, r, http.StatusBadRequest,
			fmt.Errorf("get address from bridging address coordinator: %w", err), c.logger)

		return
	}

	apiUtils.WriteResponse(w, r, http.StatusOK, response.NewBridgingAddressResponse(
		chainIDStr, bridgingAddress), c.logger)
}

func (c *BridgingAddressControllerImpl) getAllBridgingAddresses(w http.ResponseWriter, r *http.Request) {
	queryValues := r.URL.Query()
	c.logger.Debug("getAllBridgingAddresses request", "query values", queryValues, "url", r.URL)

	chainIDArr, exists := queryValues["chainId"]
	if !exists || len(chainIDArr) == 0 {
		apiUtils.WriteErrorResponse(
			w, r, http.StatusBadRequest,
			errors.New("chainId missing from query"), c.logger)

		return
	}

	chainIDStr := chainIDArr[0]
	chainID := common.ToNumChainID(chainIDStr)

	bridgingAddresses := c.bridgingAddressManager.GetAllPaymentAddresses(chainID)

	apiUtils.WriteResponse(w, r, http.StatusOK, response.NewAllBridgingAddressesResponse(
		bridgingAddresses), c.logger)
}
