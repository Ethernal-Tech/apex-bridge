package core

type APIController interface {
	GetPathPrefix() string
	GetEndpoints() []*APIEndpoint
}

type API interface {
	Start()
	Dispose() error
}
