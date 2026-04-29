package databaseaccess

import (
	"github.com/Ethernal-Tech/apex-bridge/common"
	cCore "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	cDatabaseaccess "github.com/Ethernal-Tech/apex-bridge/oracle_common/database_access"
	"github.com/Ethernal-Tech/apex-bridge/oracle_solana/core"
	"go.etcd.io/bbolt"
)

type BBoltDatabase struct {
	cDatabaseaccess.BBoltDBBase[
		*core.SolanaTx,
		*core.ProcessedSolanaTx,
		*core.BridgeExpectedSolanaTx,
	]
}

var _ core.Database = (*BBoltDatabase)(nil)

func (bd *BBoltDatabase) Init(db *bbolt.DB, appConfig *cCore.AppConfig, typeRegister common.TypeRegister) {
	bd.BBoltDBBase.DB = db
	bd.SupportedChains = make(map[string]bool, len(appConfig.SolanaChains))
	bd.TypeRegister = typeRegister

	for _, chain := range appConfig.SolanaChains {
		bd.SupportedChains[chain.ChainID] = true
	}
}
