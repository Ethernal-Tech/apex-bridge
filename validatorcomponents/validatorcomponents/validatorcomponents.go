package validatorcomponents

import (
	"errors"
	"fmt"

	"github.com/Ethernal-Tech/apex-bridge/validatorcomponents/core"

	"github.com/Ethernal-Tech/apex-bridge/batcher/batcher_manager"
	batcherCore "github.com/Ethernal-Tech/apex-bridge/batcher/core"
	oracleCore "github.com/Ethernal-Tech/apex-bridge/oracle/core"
	"github.com/Ethernal-Tech/apex-bridge/oracle/oracle"
)

type ValidatorComponentsImpl struct {
	oracle         oracleCore.Oracle
	batcherManager batcherCore.BatcherManager
}

var _ core.ValidatorComponents = (*ValidatorComponentsImpl)(nil)

func NewValidatorComponents(appConfig *core.AppConfig) (*ValidatorComponentsImpl, error) {
	oracleConfig, batcherConfig := appConfig.SeparateConfigs()

	oracle := oracle.NewOracle(oracleConfig)
	if oracle == nil {
		return nil, fmt.Errorf("failed to create oracle")
	}

	batcherManager := batcher_manager.NewBatcherManager(batcherConfig, make(map[string]batcherCore.ChainOperations))
	if batcherManager == nil {
		return nil, fmt.Errorf("failed to create batcher manager")
	}

	return &ValidatorComponentsImpl{
		oracle:         oracle,
		batcherManager: batcherManager,
	}, nil
}

func (v *ValidatorComponentsImpl) Start() error {
	err := v.oracle.Start()
	if err != nil {
		return fmt.Errorf("failed to start oracle. error: %v", err)
	}

	err = v.batcherManager.Start()
	if err != nil {
		return fmt.Errorf("failed to start batchers. error: %v", err)
	}

	return nil
}

func (v *ValidatorComponentsImpl) Stop() error {
	errb := v.batcherManager.Stop()
	erro := v.oracle.Stop()

	return errors.Join(errb, erro)
}

func (v *ValidatorComponentsImpl) ErrorCh() <-chan error {
	return v.oracle.ErrorCh()
}
