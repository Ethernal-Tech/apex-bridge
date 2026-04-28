package processor

import (
	"math/big"
	"testing"

	"github.com/Ethernal-Tech/apex-bridge/common"
	oCore "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	"github.com/Ethernal-Tech/apex-bridge/oracle_solana/core"
	solanatx "github.com/Ethernal-Tech/apex-bridge/solana"
	skyline "github.com/Ethernal-Tech/solana-infrastructure/sendtx/skyline_program"
	"github.com/Ethernal-Tech/solana-infrastructure/tracker"
	"github.com/gagliardetto/solana-go"
	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/require"
)

func TestParseHotWalletIncrementEvent(t *testing.T) {
	validMint := solana.NewWallet().PublicKey()

	appConfig := &oCore.AppConfig{
		SolanaChains: map[string]*oCore.SolanaChainConfig{
			common.ChainIDStrSolana: {
				SolanaChainConfig: solanatx.SolanaChainConfig{
					Tokens: map[uint16]common.Token{
						1: {ChainSpecific: validMint.String()},
					},
				},
			},
		},
		ChainIDConverter: common.NewTestChainIDConverter(),
	}
	appConfig.FillOut()

	receiver := NewSolTxsReceiver(
		appConfig,
		nil,
		nil,
		nil,
		hclog.NewNullLogger(),
	)

	t.Run("invalid event type", func(t *testing.T) {
		_, err := receiver.parseHotWalletIncrementEvent(common.ChainIDStrSolana, tracker.EventNotification{
			EventName: core.HotWalletIncrementEvent,
			EventData: "invalid",
		})
		require.ErrorContains(t, err, "failed to parse hot wallet increment event")
	})

	t.Run("unknown token mint", func(t *testing.T) {
		_, err := receiver.parseHotWalletIncrementEvent(common.ChainIDStrSolana, tracker.EventNotification{
			EventName: core.HotWalletIncrementEvent,
			EventData: &skyline.HotWalletIncrementEvent{
				Sender: solana.NewWallet().PublicKey(),
				Mint:   solana.NewWallet().PublicKey(),
				Amount: 10,
			},
		})
		require.ErrorContains(t, err, "failed to get token id by mint")
	})

	t.Run("valid", func(t *testing.T) {
		txSig := solana.Signature{1, 2, 3}
		slot := uint64(123)
		amountLamports := uint64(10)

		tx, err := receiver.parseHotWalletIncrementEvent(common.ChainIDStrSolana, tracker.EventNotification{
			EventName:   core.HotWalletIncrementEvent,
			SlotNumber:  slot,
			TxSignature: txSig,
			EventData: &skyline.HotWalletIncrementEvent{
				Sender: solana.NewWallet().PublicKey(),
				Mint:   validMint,
				Amount: amountLamports,
			},
		})
		require.NoError(t, err)
		require.Equal(t, common.ChainIDStrSolana, tx.OriginChainID)
		require.Equal(t, slot, tx.SlotNumber)
		require.Equal(t, txSig, tx.TxSignature)
		require.Equal(t, common.LamportToWei(new(big.Int).SetUint64(amountLamports)), tx.Value)
		require.Empty(t, tx.Metadata)
	})
}
