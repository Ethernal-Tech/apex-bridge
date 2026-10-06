package core

import (
	"github.com/Ethernal-Tech/apex-bridge/common"
	oracleCC "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
)

type BridgingRequestStateDB interface {
	AddBridgingRequestState(state *common.BridgingRequestState) error
	UpdateBridgingRequestState(state *common.BridgingRequestState) error
	GetBridgingRequestState(sourceChainID string, sourceTxHash []byte) (*common.BridgingRequestState, error)
	GetBridgingRequestStatesPage(
		fromSyncIndex uint64, limit int,
	) (states []*common.BridgingRequestState, nextSyncIndex uint64, err error)
	GetSyncInstanceID() (string, error)
}

type Database interface {
	BridgingRequestStateDB
	oracleCC.ProtocolParamsDB
	Init(filePath string) error
	Close() error
}

type BridgingRequestStateManager interface {
	common.BridgingRequestStateUpdater

	Get(sourceChainID string, sourceTxHash []byte) (*common.BridgingRequestState, error)
	GetMultiple(sourceChainID string, sourceTxHashes [][]byte) ([]*common.BridgingRequestState, error)
	GetPage(
		fromSyncIndex uint64, limit int,
	) (states []*common.BridgingRequestState, nextSyncIndex uint64, instanceID string, err error)
}

type RelayerImitator interface {
	common.IStartable
}

type ValidatorComponents interface {
	Start() error
	Dispose() error
}
