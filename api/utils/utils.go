package utils

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os/exec"
	"path/filepath"

	"github.com/Ethernal-Tech/apex-bridge/api/model/response"
	"github.com/Ethernal-Tech/apex-bridge/validatorcomponents/core"
	loggerInfra "github.com/Ethernal-Tech/cardano-infrastructure/logger"
	"github.com/hashicorp/go-hclog"
)

func WriteResponse(w http.ResponseWriter, r *http.Request, status int, response any, logger hclog.Logger) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Error("write response error", "url", r.URL, "status", status, "err", err)
	}
}

func WriteErrorResponse(w http.ResponseWriter, r *http.Request, status int, err error, logger hclog.Logger) {
	logger.Info("error happened", "url", r.URL, "status", status, "err", err)

	WriteResponse(w, r, status, response.ErrorResponse{Err: err.Error()}, logger)
}

func WriteUnauthorizedResponse(w http.ResponseWriter, r *http.Request, logger hclog.Logger) {
	WriteErrorResponse(w, r, http.StatusUnauthorized, errors.New("Unauthorized"), logger)
}

func DecodeModel[T any](w http.ResponseWriter, r *http.Request, logger hclog.Logger) (T, bool) {
	var requestBody T

	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		WriteErrorResponse(w, r, http.StatusBadRequest, fmt.Errorf("bad request: %w", err), logger)

		return requestBody, false
	}

	return requestBody, true
}

func FormatProcessOnPort(port uint32) string {
	process, err := ProcessOnPort(port)
	if err != nil {
		return err.Error()
	}

	return process
}

func ProcessOnPort(port uint32) (string, error) {
	cmd := exec.Command("sh", "-c", fmt.Sprintf("lsof -i tcp:%d | grep LISTEN | awk '{print $2}'", port)) //nolint:gosec

	// Run the command and capture the output
	var out bytes.Buffer
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("cmd failed: %w", err)
	}

	return out.String(), nil
}

func NewAPILogger(appConfig *core.AppConfig) (hclog.Logger, error) {
	logDir := filepath.Dir(appConfig.Settings.Logger.LogFilePath)

	apiLoggerConfig := appConfig.Settings.Logger
	apiLoggerConfig.LogFilePath = filepath.Join(logDir, "api.log")

	apiLogger, err := loggerInfra.NewLogger(apiLoggerConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create apiLogger. err: %w", err)
	}

	return apiLogger, nil
}
