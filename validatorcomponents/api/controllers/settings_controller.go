package controllers

import (
	"net/http"

	"github.com/Ethernal-Tech/apex-bridge/validatorcomponents/api/model/response"
	"github.com/Ethernal-Tech/apex-bridge/validatorcomponents/api/utils"
	"github.com/Ethernal-Tech/apex-bridge/validatorcomponents/core"
	"github.com/hashicorp/go-hclog"
)

type SettingsControllerImpl struct {
	appConfig *core.AppConfig
	logger    hclog.Logger
}

var _ core.APIController = (*SettingsControllerImpl)(nil)

func NewSettingsController(
	appConfig *core.AppConfig,
	logger hclog.Logger,
) *SettingsControllerImpl {
	return &SettingsControllerImpl{
		appConfig: appConfig,
		logger:    logger,
	}
}

func (*SettingsControllerImpl) GetPathPrefix() string {
	return "Settings"
}

func (c *SettingsControllerImpl) GetEndpoints() []*core.APIEndpoint {
	return []*core.APIEndpoint{
		{Path: "Get", Method: http.MethodGet, Handler: c.getSettings, APIKeyAuth: true},
	}
}

func (c *SettingsControllerImpl) getSettings(w http.ResponseWriter, r *http.Request) {
	utils.WriteResponse(w, r, http.StatusOK, response.NewSettingsResponse(c.appConfig), c.logger)
}
