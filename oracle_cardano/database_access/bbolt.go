package databaseaccess

import (
	"github.com/Ethernal-Tech/apex-bridge/oracle_cardano/core"
	cDatabaseaccess "github.com/Ethernal-Tech/apex-bridge/oracle_common/database_access"
	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
)

type BBoltDatabase struct {
	cDatabaseaccess.BBoltDBBase[
		core.CardanoTx,
		core.ProcessedCardanoTx,
		core.BridgeExpectedCardanoTx,
		core.BridgeExpectedCardanoDBTx,
	]
}

var _ core.Database = (*BBoltDatabase)(nil)

func (bd *BBoltDatabase) Init(filePath string) error {
	return bd.BBoltDBBase.Init(filePath, nil)
}

func (bd *BBoltDatabase) GetProcessedTx(
	chainID string, txHash indexer.Hash,
) (result *core.ProcessedCardanoTx, err error) {
	return bd.BBoltDBBase.GetProcessedTx(chainID, core.ToCardanoTxKey(chainID, txHash))
}

func (bd *BBoltDatabase) MarkTxs(
	expectedInvalid []*core.BridgeExpectedCardanoTx, expectedProcessed []*core.BridgeExpectedCardanoTx,
	allProcessed []*core.ProcessedCardanoTx) error {
	return bd.BBoltDBBase.MarkTxs(
		expectedInvalid, expectedProcessed, allProcessed, nil,
	)
}
