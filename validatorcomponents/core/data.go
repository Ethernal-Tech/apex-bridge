package core

import "net/http"

type ApiEndpoint struct {
	Path    string
	Method  string
	Handler func(w http.ResponseWriter, r *http.Request)
}
