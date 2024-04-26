package utils

import (
	"encoding/json"
	"net/http"

	"github.com/Ethernal-Tech/apex-bridge/validatorcomponents/api/model/response"
)

func WriteErrorResponse(w http.ResponseWriter, status int, err string) error {
	w.WriteHeader(status)

	return json.NewEncoder(w).Encode(response.ErrorResponse{Err: err})
}

func WriteUnauthorizedResponse(w http.ResponseWriter) error {
	w.WriteHeader(http.StatusUnauthorized)

	return json.NewEncoder(w).Encode(response.ErrorResponse{Err: "Unauthorized"})
}
