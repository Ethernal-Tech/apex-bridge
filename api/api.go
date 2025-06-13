// @title Oracle API
// @version 1.0
// @BasePath /api
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name X-API-Key
package api

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/Ethernal-Tech/apex-bridge/api/core"
	"github.com/Ethernal-Tech/apex-bridge/api/utils"
	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/hashicorp/go-hclog"
)

const (
	apiStartDelay = 5 * time.Second
)

type APIImpl struct {
	ctx       context.Context
	apiConfig core.APIConfig
	handler   http.Handler
	server    *http.Server
	logger    hclog.Logger

	serverClosedCh chan bool
}

var _ core.API = (*APIImpl)(nil)

func NewAPI(
	ctx context.Context, apiConfig core.APIConfig,
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

			endpointHandler = endpointWrapper(endpoint.Path, endpointHandler, logger)

			router.HandleFunc(endpointPath, endpointHandler).Methods(endpoint.Method)

			logger.Debug("Registered api endpoint", "endpoint", endpointPath, "method", endpoint.Method)
		}
	}

	handler := handlers.CORS(originsOk, headersOk, methodsOk)(router)

	return &APIImpl{
		ctx:       ctx,
		apiConfig: apiConfig,
		handler:   handler,
		logger:    logger,
	}, nil
}

func (api *APIImpl) Start() {
	// delay api start a bit, in case OS has not released port yet from a previous run
	select {
	case <-api.ctx.Done():
		return
	case <-time.After(apiStartDelay):
	}

	api.logger.Debug("Checking process running on port",
		"port", api.apiConfig.Port, "process", utils.FormatProcessOnPort(api.apiConfig.Port))

	api.serverClosedCh = make(chan bool)

	err := common.RetryForever(api.ctx, apiStartDelay, func(ctx context.Context) error {
		api.logger.Debug("Trying to start api")

		srvCtx, cancelFunc := context.WithCancel(ctx)
		defer cancelFunc()

		api.server = &http.Server{
			Addr:              fmt.Sprintf(":%d", api.apiConfig.Port),
			Handler:           api.handler,
			ReadHeaderTimeout: 3 * time.Second,
			ConnContext:       func(ctx context.Context, c net.Conn) context.Context { return srvCtx },
			BaseContext:       func(l net.Listener) context.Context { return srvCtx },
		}

		err := api.server.ListenAndServe()
		if err == nil || err == http.ErrServerClosed {
			return nil
		}

		api.logger.Debug("Checking process running on port",
			"port", api.apiConfig.Port, "process", utils.FormatProcessOnPort(api.apiConfig.Port))

		api.logger.Error("Error while trying to start api. Retrying...", "err", err)

		api.server.Close()

		return err
	})
	if err != nil {
		api.logger.Error("error after api ListenAndServe", "err", err)
	}

	api.logger.Debug("Stopped api")
	api.serverClosedCh <- true
}

func (api *APIImpl) Dispose() error {
	var apiErrors []error

	if api.server == nil {
		return nil
	}

	err := api.server.Shutdown(context.Background())
	if err != nil {
		apiErrors = append(apiErrors, fmt.Errorf("error while trying to shutdown api server. err %w", err))
	}

	api.logger.Debug("Called api shutdown")

	select {
	case <-time.After(time.Second * 5):
		api.logger.Debug("api not closed after a timeout")

		if err := api.server.Close(); err != nil {
			apiErrors = append(apiErrors, fmt.Errorf("error while trying to close api server. err: %w", err))
		}

		api.logger.Debug("Called forceful Close")
	case <-api.serverClosedCh:
	}

	api.logger.Debug("Finished disposing")

	return errors.Join(apiErrors...)
}

func endpointWrapper(path string, handler core.APIEndpointHandler, logger hclog.Logger) core.APIEndpointHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Debug("endpoint called", "path", path, "url", r.URL)
		handler(w, r)
		logger.Debug("endpoint call finished", "path", path, "url", r.URL)
	}
}

func withAPIKeyAuth(
	apiConfig core.APIConfig, handler core.APIEndpointHandler, logger hclog.Logger,
) core.APIEndpointHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		apiKeyHeaderValue := r.Header.Get(apiConfig.APIKeyHeader)
		if apiKeyHeaderValue == "" {
			utils.WriteUnauthorizedResponse(w, r, logger)

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
			utils.WriteUnauthorizedResponse(w, r, logger)

			return
		}

		handler(w, r)
	}
}
