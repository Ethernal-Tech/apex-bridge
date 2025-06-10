package databaseaccess

import (
	oCore "github.com/Ethernal-Tech/apex-bridge/oracle_cardano/core"
	cDatabaseaccess "github.com/Ethernal-Tech/apex-bridge/oracle_common/database_access"
	"github.com/Ethernal-Tech/apex-bridge/staking/core"
	"go.etcd.io/bbolt"
)

type BBoltDatabase struct {
	cDatabaseaccess.BBoltDBBase[
		*oCore.CardanoTx,
		*oCore.ProcessedCardanoTx,
		*oCore.BridgeExpectedCardanoTx,
	]
}

var _ core.Database = (*BBoltDatabase)(nil)

func (bd *BBoltDatabase) Init(db *bbolt.DB, smConfig *core.StakingManagerConfiguration) {
	bd.BBoltDBBase.DB = db
	bd.SupportedChains = make(map[string]bool, len(smConfig.Chains))

	for _, chain := range smConfig.Chains {
		bd.SupportedChains[chain.ChainID] = true
	}
}
