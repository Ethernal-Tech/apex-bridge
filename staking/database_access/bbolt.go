package databaseaccess

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"

	oCore "github.com/Ethernal-Tech/apex-bridge/oracle_cardano/core"
	cDatabaseaccess "github.com/Ethernal-Tech/apex-bridge/oracle_common/database_access"
	"github.com/Ethernal-Tech/apex-bridge/staking/core"
	"go.etcd.io/bbolt"
)

var exchangeRateBucket = []byte(ExchangeRateBucket)

type StakingAddressDecoder func([]byte) (core.StakingAddress, error)

type BBoltDatabase struct {
	cDatabaseaccess.BBoltDBBase[
		*oCore.CardanoTx,
		*oCore.ProcessedCardanoTx,
		*oCore.BridgeExpectedCardanoTx,
	]

	DecodeStakingAddress StakingAddressDecoder
}

var _ core.Database = (*BBoltDatabase)(nil)

func NewBBoltDatabase(decodeStakingAddress StakingAddressDecoder) *BBoltDatabase {
	return &BBoltDatabase{
		DecodeStakingAddress: decodeStakingAddress,
	}
}

func (bd *BBoltDatabase) Init(db *bbolt.DB, smConfig *core.StakingManagerConfiguration) {
	bd.BBoltDBBase.DB = db
	bd.SupportedChains = make(map[string]bool, len(smConfig.Chains))

	for _, chain := range smConfig.Chains {
		bd.SupportedChains[chain.ChainID] = true
	}
}

func (bd *BBoltDatabase) UpdateExchangeRate(chainID string, exchangeRate float64) error {
	return bd.UpdateStakingAddressAndExRate(chainID, nil, &exchangeRate)
}

func (bd *BBoltDatabase) GetLastExchangeRate(chainID string) (result float64, err error) {
	err = bd.BBoltDBBase.DB.View(func(tx *bbolt.Tx) error {
		bytes := tx.Bucket(exchangeRateBucket).Get([]byte(chainID))
		if bytes == nil {
			return fmt.Errorf("could not get exchange rate for chainID %s: key not found", chainID)
		}

		bits := binary.BigEndian.Uint64(bytes)
		result = math.Float64frombits(bits)

		return nil
	})

	return result, err
}

func (bd *BBoltDatabase) UpdateStakingAddress(chainID string, stakingAddress core.StakingAddress) error {
	return bd.UpdateStakingAddressAndExRate(chainID, stakingAddress, nil)
}

func (bd *BBoltDatabase) GetStakingAddress(chainID string, address string) (result core.StakingAddress, err error) {
	if supported := bd.SupportedChains[chainID]; !supported {
		return result, fmt.Errorf("unsupported chain: %s", chainID)
	}

	err = bd.BBoltDBBase.DB.View(func(tx *bbolt.Tx) error {
		data := tx.Bucket(ChainBucket(StakingAddressesBucket, chainID)).Get([]byte(address))
		if data == nil {
			return fmt.Errorf("staking address not found: %s", address)
		}

		result, err = bd.DecodeStakingAddress(data)
		if err != nil {
			return fmt.Errorf("decode error: %w", err)
		}

		return nil
	})

	return result, err
}

func (bd *BBoltDatabase) GetAllStakingAddresses(chainID string) (result []core.StakingAddress, err error) {
	if supported := bd.SupportedChains[chainID]; !supported {
		return nil, fmt.Errorf("unsupported chain: %s", chainID)
	}

	err = bd.DB.View(func(tx *bbolt.Tx) error {
		cursor := tx.Bucket(ChainBucket(StakingAddressesBucket, chainID)).Cursor()

		for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
			stakingAddress, err := bd.DecodeStakingAddress(v)
			if err != nil {
				return fmt.Errorf("decode error: %w", err)
			}

			result = append(result, stakingAddress)
		}

		return nil
	})

	return result, err
}

func (bd *BBoltDatabase) UpdateStakingAddressAndExRate(
	chainID string,
	stakingAddress core.StakingAddress,
	exchangeRate *float64,
) error {
	return bd.BBoltDBBase.DB.Update(func(tx *bbolt.Tx) error {
		if supported := bd.SupportedChains[chainID]; !supported {
			return fmt.Errorf("unsupported chain: %s", chainID)
		}

		if stakingAddress != nil {
			bytes, err := json.Marshal(stakingAddress)
			if err != nil {
				return fmt.Errorf("could not marshal staking address %s: %w", stakingAddress.GetAddress(), err)
			}

			stakingAddressesBucket := tx.Bucket(ChainBucket(StakingAddressesBucket, chainID))
			if err = stakingAddressesBucket.Put([]byte(stakingAddress.GetAddress()), bytes); err != nil {
				return fmt.Errorf("staking address %s write error: %w", stakingAddress.GetAddress(), err)
			}
		}

		if exchangeRate != nil {
			var bytes [8]byte

			binary.BigEndian.PutUint64(bytes[:], math.Float64bits(*exchangeRate))

			if err := tx.Bucket(exchangeRateBucket).Put([]byte(chainID), bytes[:]); err != nil {
				return fmt.Errorf("last exchange rate write error: %w", err)
			}
		}

		return nil
	})
}
