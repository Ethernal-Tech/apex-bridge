package relayer

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/Ethernal-Tech/apex-bridge/eth"
	"github.com/Ethernal-Tech/apex-bridge/relayer/core"
	"github.com/Ethernal-Tech/cardano-infrastructure/logger"
	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestRelayerExecute(t *testing.T) {
	relayerConfig := &core.RelayerConfiguration{
		Bridge: core.BridgeConfig{
			NodeUrl:              "https://polygon-mumbai-pokt.nodies.app", // will be our node,
			SmartContractAddress: "0x816402271eE6D9078Fc8Cb537aDBDD58219485BA",
		},
		Base: core.BaseConfig{
			ChainId: "prime",
		},
		PullTimeMilis: 1000,
		Logger: logger.LoggerConfig{
			LogFilePath:   "./relayer_logs",
			LogLevel:      hclog.Debug,
			JSONLogFormat: false,
			AppendFile:    true,
		},
	}

	ctx, cancelCtx := context.WithTimeout(context.Background(), time.Second*60)
	defer cancelCtx()

	confirmedBatchRet := &eth.ConfirmedBatch{
		Id:                         "1",
		RawTransaction:             []byte{},
		MultisigSignatures:         [][]byte{},
		FeePayerMultisigSignatures: [][]byte{},
	}

	testError := errors.New("test err")

	t.Run("execute test fail to retrieve", func(t *testing.T) {
		bridgeSmartContractMock := &bridgeSmartContractMock{}
		operationsMock := &cardanoChainOperationsMock{}
		bridgeSmartContractMock.On("GetConfirmedBatch", ctx, "prime").Return(nil, testError)

		r := NewRelayer(relayerConfig, bridgeSmartContractMock, hclog.Default(), operationsMock)
		err := r.execute(ctx)
		require.Error(t, err)
		require.ErrorContains(t, err, "failed to retrieve confirmed batch")
	})

	t.Run("execute test fail to send", func(t *testing.T) {
		bridgeSmartContractMock := &bridgeSmartContractMock{}
		operationsMock := &cardanoChainOperationsMock{}
		bridgeSmartContractMock.On("GetConfirmedBatch", ctx, "prime").Return(confirmedBatchRet, nil)
		operationsMock.On("SendTx", confirmedBatchRet).Return(testError)

		r := NewRelayer(relayerConfig, bridgeSmartContractMock, hclog.Default(), operationsMock)
		err := r.execute(ctx)
		require.Error(t, err)
		require.ErrorContains(t, err, "failed to send confirmed batch")
	})

	t.Run("execute test valid", func(t *testing.T) {
		bridgeSmartContractMock := &bridgeSmartContractMock{}
		operationsMock := &cardanoChainOperationsMock{}
		bridgeSmartContractMock.On("GetConfirmedBatch", ctx, "prime").Return(confirmedBatchRet, nil)
		operationsMock.On("SendTx", confirmedBatchRet).Return(nil)

		r := NewRelayer(relayerConfig, bridgeSmartContractMock, hclog.Default(), operationsMock)
		require.NoError(t, r.execute(ctx))
	})
}

func TestRelayerGetChainSpecificOperations(t *testing.T) {
	jsonData := []byte(`{
		"testnetMagic": 2,
		"blockfrostUrl": "https://cardano-preview.blockfrost.io/api/v0",
		"blockfrostApiKey": "preview7mGSjpyEKb24OxQ4cCxomxZ5axMs5PvE",
		"atLeastValidators": 0.6666666666666666,
		"potentialFee": 300000
		}`)

	t.Run("invalid chain type", func(t *testing.T) {
		chainSpecificConfig := core.ChainSpecific{
			ChainType: "Invalid",
			Config:    json.RawMessage(""),
		}

		chainOp, err := GetChainSpecificOperations(chainSpecificConfig)
		require.Nil(t, chainOp)
		require.Error(t, err)
		require.ErrorContains(t, err, "unknown chain type")
	})

	t.Run("invalid cardano json config", func(t *testing.T) {
		chainSpecificConfig := core.ChainSpecific{
			ChainType: "Cardano",
			Config:    json.RawMessage(""),
		}

		chainOp, err := GetChainSpecificOperations(chainSpecificConfig)
		require.Nil(t, chainOp)
		require.Error(t, err)
		require.ErrorContains(t, err, "failed to unmarshal Cardano configuration")
	})

	t.Run("valid cardano config", func(t *testing.T) {
		chainSpecificConfig := core.ChainSpecific{
			ChainType: "Cardano",
			Config:    json.RawMessage(jsonData),
		}

		chainOp, err := GetChainSpecificOperations(chainSpecificConfig)
		require.NotNil(t, chainOp)
		require.NoError(t, err)
	})

	t.Run("valid cardano config check case sensitivity", func(t *testing.T) {
		chainSpecificConfig := core.ChainSpecific{
			ChainType: "CaRdAnO",
			Config:    json.RawMessage(jsonData),
		}

		chainOp, err := GetChainSpecificOperations(chainSpecificConfig)
		require.NotNil(t, chainOp)
		require.NoError(t, err)
	})
}

type bridgeSmartContractMock struct {
	mock.Mock
}

func (m *bridgeSmartContractMock) GetConfirmedBatch(
	ctx context.Context, destinationChain string) (*eth.ConfirmedBatch, error) {
	args := m.Called(ctx, destinationChain)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*eth.ConfirmedBatch), args.Error(1)
}

type cardanoChainOperationsMock struct {
	mock.Mock
}

func (m *cardanoChainOperationsMock) SendTx(smartContractData *eth.ConfirmedBatch) error {
	args := m.Called(smartContractData)

	return args.Error(0)
}
