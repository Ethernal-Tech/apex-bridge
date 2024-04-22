package api

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Ethernal-Tech/apex-bridge/validatorcomponents/core"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/hashicorp/go-hclog"
)

type ApiImpl struct {
	ctx       context.Context
	apiConfig core.ApiConfig
	handler   http.Handler
	server    *http.Server
	logger    hclog.Logger
}

var _ core.Api = (*ApiImpl)(nil)

func NewApi(ctx context.Context, apiConfig core.ApiConfig, controllers []core.ApiController, logger hclog.Logger) (*ApiImpl, error) {
	headersOk := handlers.AllowedHeaders(apiConfig.AllowedHeaders)
	originsOk := handlers.AllowedOrigins(apiConfig.AllowedOrigins)
	methodsOk := handlers.AllowedMethods(apiConfig.AllowedMethods)

	router := mux.NewRouter().StrictSlash(true)

	for _, controller := range controllers {
		controllerPathPrefix := controller.GetPathPrefix()
		endpoints := controller.GetEndpoints()
		for _, endpoint := range endpoints {
			endpointPath := fmt.Sprintf("/%v/%v/%v", apiConfig.PathPrefix, controllerPathPrefix, endpoint.Path)
			router.HandleFunc(endpointPath, endpoint.Handler).Methods(endpoint.Method)
		}
	}

	handler := handlers.CORS(originsOk, headersOk, methodsOk)(router)

	return &ApiImpl{
		ctx:       ctx,
		apiConfig: apiConfig,
		handler:   handler,
		logger:    logger,
	}, nil
}

func (api *ApiImpl) Start() error {
	api.logger.Debug("Starting api")
	api.server = &http.Server{Addr: fmt.Sprintf(":%v", api.apiConfig.Port), Handler: api.handler}
	err := api.server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		api.logger.Error("error while trying to start api server", "err", err)
		return fmt.Errorf("error while trying to start api server. err: %w", err)
	}

	api.logger.Debug("Started api")

	return nil
}

func (api *ApiImpl) Dispose() error {
	err := api.server.Shutdown(context.Background())
	api.logger.Debug("Stopped api")
	if err != nil {
		api.logger.Error("error while trying to shutdown api server", "err", err)
		return fmt.Errorf("error while trying to shutdown api server. err %w", err)
	}

	return nil
}
