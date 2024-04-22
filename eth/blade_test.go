package eth

import (
	"context"
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestSC(t *testing.T) {
	ctx, cancelCtx := context.WithTimeout(context.Background(), time.Second*60)
	defer cancelCtx()

	bridgeSC := NewBridgeSmartContract("http://localhost:10002", "0x0400000000000000000000000000000000000000")

	t.Run("ShouldCreateBatch test", func(t *testing.T) {
		ret, err := bridgeSC.ShouldCreateBatch(ctx, "destChain")
		require.NoError(t, err)
		require.Equal(t, false, ret)
	})

	t.Run("GetConfirmedBatch test", func(t *testing.T) {
		ret, err := bridgeSC.GetConfirmedBatch(ctx, "destChain")
		require.NoError(t, err)
		require.Equal(t, &ConfirmedBatch{
			Id:                         "0",
			RawTransaction:             []byte{},
			MultisigSignatures:         [][]byte{},
			FeePayerMultisigSignatures: [][]byte{},
		}, ret)
	})

	t.Run("GetConfirmedTransactions test", func(t *testing.T) {
		_, err := bridgeSC.GetConfirmedTransactions(ctx, "destChain")
		// Probably reverts on: https://github.com/Ethernal-Tech/apex-bridge-smartcontracts/blob/upgrade/contracts/BridgeContract.sol#L169
		require.Error(t, err)
	})

	t.Run("GetAvailableUTXOs test", func(t *testing.T) {
		ret, err := bridgeSC.GetAvailableUTXOs(ctx, "destChain")
		require.NoError(t, err)
		require.Equal(t, &UTXOs{
			MultisigOwnedUTXOs: []UTXO{},
			FeePayerOwnedUTXOs: []UTXO{},
		}, ret)
	})

	t.Run("GetLastObservedBlock test", func(t *testing.T) {
		ret, err := bridgeSC.GetLastObservedBlock(ctx, "destChain")
		require.NoError(t, err)
		require.Equal(t, &CardanoBlock{
			BlockHash: "",
			BlockSlot: 0,
		}, ret)
	})

	t.Run("GetValidatorsCardanoData test", func(t *testing.T) {
		ret, err := bridgeSC.GetValidatorsCardanoData(ctx, "destChain")
		require.NoError(t, err)
		require.Equal(t, make([]ValidatorCardanoData, 0), ret)
	})

	t.Run("GetNextBatchId test", func(t *testing.T) {
		ret, err := bridgeSC.GetNextBatchId(ctx, "destChain")
		require.NoError(t, err)
		require.Equal(t, 0, big.NewInt(0).Cmp(ret))
	})

	t.Run("GetAllRegisteredChains test", func(t *testing.T) {
		ret, err := bridgeSC.GetAllRegisteredChains(ctx)
		require.NoError(t, err)
		require.Equal(t, make([]Chain, 0), ret)
	})
}
