package oracle

import (
	"context"

	"github.com/Ethernal-Tech/apex-bridge/oracle/core"
	"github.com/hashicorp/go-hclog"
)

const (
	MainComponentName = "eth_oracle"
)

type OracleImpl struct {
	ctx       context.Context
	appConfig *core.AppConfig
	logger    hclog.Logger

	errorCh chan error
}

var _ core.Oracle = (*OracleImpl)(nil)

func NewOracle(
	ctx context.Context,
	appConfig *core.AppConfig,
	logger hclog.Logger,
) (*OracleImpl, error) {
	return &OracleImpl{
		ctx:       ctx,
		appConfig: appConfig,
		logger:    logger,
	}, nil
}

func (o *OracleImpl) Start() error {
	o.logger.Debug("Starting EthOracle")

	o.errorCh = make(chan error, 1)
	go o.errorHandler()

	o.logger.Debug("Started EthOracle")

	return nil
}

func (o *OracleImpl) Dispose() error {
	close(o.errorCh)

	return nil
}

func (o *OracleImpl) ErrorCh() <-chan error {
	return o.errorCh
}

type ErrorOrigin struct {
	err    error
	origin string
}

func (o *OracleImpl) errorHandler() {
	agg := make(chan ErrorOrigin)
	defer close(agg)

	select {
	case errorOrigin := <-agg:
		o.logger.Error("critical error", "origin", errorOrigin.origin, "err", errorOrigin.err)
		o.errorCh <- errorOrigin.err
	case <-o.ctx.Done():
	}
	o.logger.Debug("Exiting eth_oracle error handler")
}
