package utils

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os/exec"
	"regexp"

	"github.com/Ethernal-Tech/apex-bridge/validatorcomponents/api/model/response"
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
	logger.Error("error happened", "url", r.URL, "status", status, "err", err)

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
	cmd := exec.Command("sh", "-c", fmt.Sprintf("ss -tulpn | grep %d", port))

	// Run the command and capture the output
	var out bytes.Buffer
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("cmd failed or no results found: %w", err)
	}

	result := out.String()
	re := regexp.MustCompile(`users:\(\((.*?)\)\)`)
	match := re.FindStringSubmatch(result)

	if len(match) > 1 {
		return match[1], nil
	}

	return "", nil
}
