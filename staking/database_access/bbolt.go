package databaseaccess

import (
	"encoding/binary"
	"fmt"
	"math"

	oCore "github.com/Ethernal-Tech/apex-bridge/oracle_cardano/core"
	cDatabaseaccess "github.com/Ethernal-Tech/apex-bridge/oracle_common/database_access"
	"github.com/Ethernal-Tech/apex-bridge/staking/core"
	"go.etcd.io/bbolt"
)

var exchangeRateBucket = []byte(ExchangeRateBucket)

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

func (bd *BBoltDatabase) UpdateExchangeRate(chainID string, exchangeRate float64) error {
	return bd.BBoltDBBase.DB.Update(func(tx *bbolt.Tx) error {
		var bytes [8]byte

		binary.BigEndian.PutUint64(bytes[:], math.Float64bits(exchangeRate))

		if err := tx.Bucket(exchangeRateBucket).Put([]byte(chainID), bytes[:]); err != nil {
			return fmt.Errorf("last exchange rate write error: %w", err)
		}

		return nil
	})
}

func (bd *BBoltDatabase) GetLastExchangeRate(chainID string) (float64, error) {
	var result float64

	err := bd.BBoltDBBase.DB.View(func(tx *bbolt.Tx) error {
		bytes := tx.Bucket(exchangeRateBucket).Get([]byte(chainID))
		if bytes == nil {
			return fmt.Errorf("could not get exchange rate for chainID %s: key not found", chainID)
		}

		bits := binary.BigEndian.Uint64(bytes)
		result = math.Float64frombits(bits)

		return nil
	})

	if err != nil {
		return 0, err
	}

	return result, nil
}
