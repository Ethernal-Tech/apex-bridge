package core

import "net/http"

type ApiEndpointHandler = func(w http.ResponseWriter, r *http.Request)

type ApiEndpoint struct {
	Path       string
	Method     string
	Handler    ApiEndpointHandler
	ApiKeyAuth bool
}
