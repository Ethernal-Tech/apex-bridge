package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/Ethernal-Tech/apex-bridge/validatorcomponents/api/utils"
	"github.com/Ethernal-Tech/apex-bridge/validatorcomponents/core"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/hashicorp/go-hclog"
)

type APIImpl struct {
	apiConfig core.APIConfig
	handler   http.Handler
	server    *http.Server
	logger    hclog.Logger

	finishDispose chan error
}

var _ core.API = (*APIImpl)(nil)

func NewAPI(
	apiConfig core.APIConfig,
	controllers []core.APIController, logger hclog.Logger,
) (
	*APIImpl, error,
) {
	headersOk := handlers.AllowedHeaders(apiConfig.AllowedHeaders)
	originsOk := handlers.AllowedOrigins(apiConfig.AllowedOrigins)
	methodsOk := handlers.AllowedMethods(apiConfig.AllowedMethods)

	router := mux.NewRouter().StrictSlash(true)

	for _, controller := range controllers {
		controllerPathPrefix := controller.GetPathPrefix()
		endpoints := controller.GetEndpoints()

		for _, endpoint := range endpoints {
			endpointPath := fmt.Sprintf("/%s/%s/%s", apiConfig.PathPrefix, controllerPathPrefix, endpoint.Path)

			endpointHandler := endpoint.Handler
			if endpoint.APIKeyAuth {
				endpointHandler = withAPIKeyAuth(apiConfig, endpointHandler, logger)
			}

			router.HandleFunc(endpointPath, endpointHandler).Methods(endpoint.Method)

			logger.Debug("Registered api endpoint", "endpoint", endpointPath, "method", endpoint.Method)
		}
	}

	handler := handlers.CORS(originsOk, headersOk, methodsOk)(router)

	return &APIImpl{
		apiConfig: apiConfig,
		handler:   handler,
		logger:    logger,
	}, nil
}

func (api *APIImpl) Start() {
	api.logger.Debug("Starting api")
	api.server = &http.Server{
		Addr:              fmt.Sprintf(":%d", api.apiConfig.Port),
		Handler:           api.handler,
		ReadHeaderTimeout: 3 * time.Second,
	}

	api.finishDispose = make(chan error)

	err := api.server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		api.logger.Error("error after api ListenAndServe", "err", err)

		api.logger.Debug("api.finishDispose <- err", "err", err)
		api.finishDispose <- fmt.Errorf("error after api ListenAndServe. err: %w", err)

		return
	}

	api.logger.Debug("api.finishDispose <- nil", "err", err)
	api.finishDispose <- nil

	api.logger.Debug("Stopped api")
}

func (api *APIImpl) Dispose() error {
	var apiErrors []error

	api.logger.Debug("Calling api shutdown")

	err := api.server.Shutdown(context.Background())
	if err != nil {
		apiErrors = append(apiErrors, fmt.Errorf("error while trying to shutdown api server. err %w", err))
	}

	api.logger.Debug("Called api shutdown")

	select {
	case <-time.After(time.Second * 5):
		api.logger.Debug("api not closed after a timeout. Calling forceful Close")

		if err := api.server.Close(); err != nil {
			apiErrors = append(apiErrors, fmt.Errorf("error while trying to close api server. err: %w", err))
		}

		api.logger.Debug("Called forceful Close")
	case err := <-api.finishDispose:
		if err != nil {
			apiErrors = append(apiErrors, err)
		}

		api.logger.Debug("case <-api.finishDispose:")
	}

	if len(apiErrors) > 0 {
		return errors.Join(apiErrors...)
	}

	return nil
}

func withAPIKeyAuth(
	apiConfig core.APIConfig, handler core.APIEndpointHandler, logger hclog.Logger,
) core.APIEndpointHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		apiKeyHeaderValue := r.Header.Get(apiConfig.APIKeyHeader)
		if apiKeyHeaderValue == "" {
			err := utils.WriteUnauthorizedResponse(w)
			if err != nil {
				logger.Error("error while WriteUnauthorizedResponse", "err", err)
			}

			return
		}

		authorized := false

		for _, apiKey := range apiConfig.APIKeys {
			if apiKey == apiKeyHeaderValue {
				authorized = true

				break
			}
		}

		if !authorized {
			err := utils.WriteUnauthorizedResponse(w)
			if err != nil {
				logger.Error("error while WriteUnauthorizedResponse", "err", err)
			}

			return
		}

		handler(w, r)
	}
}
