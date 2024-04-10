package core

type ValidatorComponents interface {
	Start() error
	Stop() error
	ErrorCh() <-chan error
}
