package databaseaccess

import (
	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/oracle_cardano/core"
	cCore "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	cDatabaseaccess "github.com/Ethernal-Tech/apex-bridge/oracle_common/database_access"
	"go.etcd.io/bbolt"
)

type BBoltDatabase struct {
	cDatabaseaccess.BBoltDBBase[
		*core.CardanoTx,
		*core.ProcessedCardanoTx,
		*core.BridgeExpectedCardanoTx,
	]
}

var _ core.Database = (*BBoltDatabase)(nil)

func (bd *BBoltDatabase) Init(db *bbolt.DB, appConfig *cCore.AppConfig, typeRegister common.TypeRegister) {
	bd.BBoltDBBase.DB = db
	bd.SupportedChains = make(map[string]bool, len(appConfig.CardanoChains))
	bd.TypeRegister = typeRegister

	for _, chain := range appConfig.CardanoChains {
		bd.SupportedChains[chain.ChainID] = true
	}
}
