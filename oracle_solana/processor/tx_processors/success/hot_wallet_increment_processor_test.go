package successtxprocessors

import (
	"math/big"
	"testing"

	"github.com/Ethernal-Tech/apex-bridge/common"
	oCore "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	"github.com/Ethernal-Tech/apex-bridge/oracle_solana/core"
	"github.com/gagliardetto/solana-go"
	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/require"
)

func TestSolanaHotWalletIncrementProcessor(t *testing.T) {
	proc := NewSolanaHotWalletIncrementProcessor(hclog.NewNullLogger())
	appConfig := &oCore.AppConfig{
		SolanaChains: map[string]*oCore.SolanaChainConfig{
			common.ChainIDStrSolana: {},
		},
		ChainIDConverter: common.NewTestChainIDConverter(),
	}
	appConfig.FillOut()

	t.Run("GetType", func(t *testing.T) {
		require.Equal(t, common.TxTypeHotWalletFund, proc.GetType())
	})

	t.Run("PreValidate empty tx", func(t *testing.T) {
		err := proc.PreValidate(&core.SolanaTx{}, appConfig)
		require.Error(t, err)
	})

	t.Run("ValidateAndAddClaim empty tx value", func(t *testing.T) {
		claims := &oCore.BridgeClaims{}
		tx := &core.SolanaTx{
			OriginChainID: common.ChainIDStrSolana,
			Metadata:      []byte{},
		}

		err := proc.ValidateAndAddClaim(claims, tx, appConfig)
		require.ErrorContains(t, err, "tx value is zero or not set")
	})

	t.Run("ValidateAndAddClaim random metadata", func(t *testing.T) {
		err := proc.PreValidate(&core.SolanaTx{
			OriginChainID: common.ChainIDStrSolana,
			Metadata:      []byte{1, 2, 3},
			Value:         new(big.Int).SetUint64(1),
		}, appConfig)
		require.Error(t, err)
		require.ErrorContains(t, err, "validation failed for tx")
	})

	t.Run("ValidateAndAddClaim wrong origin chain", func(t *testing.T) {
		err := proc.PreValidate(&core.SolanaTx{
			OriginChainID: common.ChainIDStrPrime,
			Metadata:      []byte{},
			Value:         new(big.Int).SetUint64(1),
		}, appConfig)
		require.Error(t, err)
		require.ErrorContains(t, err, "validation failed for tx")
	})

	t.Run("ValidateAndAddClaim valid", func(t *testing.T) {
		claims := &oCore.BridgeClaims{}
		txSig := solana.Signature{1, 20}

		tx := &core.SolanaTx{
			TxSignature:   txSig,
			OriginChainID: common.ChainIDStrSolana,
			Metadata:      []byte{},
			Value:         new(big.Int).SetUint64(1),
		}

		err := proc.PreValidate(tx, appConfig)
		require.NoError(t, err)

		err = proc.ValidateAndAddClaim(claims, tx, appConfig)
		require.NoError(t, err)

		require.Len(t, claims.HotWalletIncrementClaims, 1)
		require.Equal(t, common.ChainIDIntSolana, claims.HotWalletIncrementClaims[0].ChainId)
		require.Equal(t, tx.Value, claims.HotWalletIncrementClaims[0].Amount)
		require.Equal(t, big.NewInt(0), claims.HotWalletIncrementClaims[0].AmountWrapped)
		require.Equal(t, txSig[:], claims.HotWalletIncrementClaims[0].TxHash)
	})
}
