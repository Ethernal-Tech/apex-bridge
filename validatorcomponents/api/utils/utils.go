package utils

import (
	"encoding/json"
	"net/http"

	"github.com/Ethernal-Tech/apex-bridge/validatorcomponents/api/model/response"
)

func WriteErrorResponse(w http.ResponseWriter, status int, err string) {
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(response.ErrorResponse{Err: err})
}

func WriteUnauthorizedResponse(w http.ResponseWriter) {
	w.WriteHeader(http.StatusUnauthorized)
	json.NewEncoder(w).Encode(response.ErrorResponse{Err: "Unauthorized"})
}
