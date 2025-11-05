package controllers

import (
	"errors"
	"net/http"

	"github.com/Ethernal-Tech/apex-bridge/eth"
	"github.com/Ethernal-Tech/apex-bridge/validatorcomponents/api/model/response"
	"github.com/Ethernal-Tech/apex-bridge/validatorcomponents/api/utils"
	"github.com/Ethernal-Tech/apex-bridge/validatorcomponents/core"
	"github.com/hashicorp/go-hclog"
)

type SettingsControllerImpl struct {
	appConfig           *core.AppConfig
	adminContractOracle eth.IOracleAdminSmartContract
	logger              hclog.Logger
}

var _ core.APIController = (*SettingsControllerImpl)(nil)

func NewSettingsController(
	appConfig *core.AppConfig,
	adminContractOracle eth.IOracleAdminSmartContract,
	logger hclog.Logger,
) *SettingsControllerImpl {
	return &SettingsControllerImpl{
		appConfig:           appConfig,
		adminContractOracle: adminContractOracle,
		logger:              logger,
	}
}

func (*SettingsControllerImpl) GetPathPrefix() string {
	return "Settings"
}

func (c *SettingsControllerImpl) GetEndpoints() []*core.APIEndpoint {
	return []*core.APIEndpoint{
		{Path: "Get", Method: http.MethodGet, Handler: c.getSettings, APIKeyAuth: true},
		{Path: "ValidatorChangeStatus", Method: http.MethodGet, Handler: c.getValidatorSetChangeStatus, APIKeyAuth: true},
		{Path: "GetMultiSigBridgingAddr", Method: http.MethodGet, Handler: c.getMultiSigBridgingAddr, APIKeyAuth: true},
	}
}

func (c *SettingsControllerImpl) getSettings(w http.ResponseWriter, r *http.Request) {
	utils.WriteResponse(w, r, http.StatusOK, response.NewSettingsResponse(c.appConfig), c.logger)
}

func (c *SettingsControllerImpl) getValidatorSetChangeStatus(w http.ResponseWriter, r *http.Request) {
	result, err := c.adminContractOracle.GetValidatorChangeStatus(r.Context())
	if err != nil {
		utils.WriteErrorResponse(
			w, r, http.StatusBadRequest,
			errors.New("can't get validator change status"), c.logger)

		return
	}

	utils.WriteResponse(w, r, http.StatusOK, response.NewValidatorChangeStatusResponse(result), c.logger)
}

func (c *SettingsControllerImpl) getMultiSigBridgingAddr(w http.ResponseWriter, r *http.Request) {
	utils.WriteResponse(w, r, http.StatusOK, response.NewMultiSigAddressesResponse(c.appConfig), c.logger)
}
