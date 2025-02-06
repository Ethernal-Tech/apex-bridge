package controllers

import (
	"net/http"

	apiCore "github.com/Ethernal-Tech/apex-bridge/api/core"
	apiUtils "github.com/Ethernal-Tech/apex-bridge/api/utils"
	"github.com/Ethernal-Tech/apex-bridge/validatorcomponents/api/model/response"
	"github.com/Ethernal-Tech/apex-bridge/validatorcomponents/core"
	"github.com/hashicorp/go-hclog"
)

type SettingsControllerImpl struct {
	appConfig *core.AppConfig
	logger    hclog.Logger
}

var _ apiCore.APIController = (*SettingsControllerImpl)(nil)

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

func (c *SettingsControllerImpl) GetEndpoints() []*apiCore.APIEndpoint {
	return []*apiCore.APIEndpoint{
		{Path: "Get", Method: http.MethodGet, Handler: c.getSettings, APIKeyAuth: true},
	}
}

func (c *SettingsControllerImpl) getSettings(w http.ResponseWriter, r *http.Request) {
	apiUtils.WriteResponse(w, r, http.StatusOK, response.NewSettingsResponse(c.appConfig), c.logger)
}
