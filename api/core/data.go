package core

import "net/http"

type APIEndpointHandler = func(w http.ResponseWriter, r *http.Request)

type APIEndpoint struct {
	Path       string
	Method     string
	Handler    APIEndpointHandler
	APIKeyAuth bool
}
