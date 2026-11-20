package cardanotx

import (
	"context"
	"testing"

	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
	"github.com/Ethernal-Tech/cardano-infrastructure/wallet"
	"github.com/stretchr/testify/require"
)

func TestCreateUnknownTokensRefundTxOutput(t *testing.T) {
	const (
		receiverAddr = "addr_test1vqjysa7p4mhu0l25qknwznvj0kghtr29ud7zp732ezwtzec0w8g3u"
		policyID     = "29f8873beb52e126f207a2dfd50f7cff556806b5b4cba9834a7b26a8"
		fee          = uint64(1_000_000)
	)

	protocolParams, err := (&TxProviderTestMock{ReturnDefaultParameters: true}).GetProtocolParameters(context.Background())
	require.NoError(t, err)

	txBuilder, err := wallet.NewTxBuilder(wallet.ResolveCardanoCliBinary())
	require.NoError(t, err)

	defer txBuilder.Dispose()

	txBuilder.SetProtocolParameters(protocolParams)

	utxoWithAmount := func(amount uint64) *indexer.TxOutput {
		return &indexer.TxOutput{
			Address: "bridging_addr",
			Amount:  amount,
			Tokens:  []indexer.TokenAmount{{PolicyID: policyID, Name: "tok", Amount: 100}},
		}
	}

	t.Run("amount covers fee and min utxo", func(t *testing.T) {
		output, err := CreateUnknownTokensRefundTxOutput(txBuilder, receiverAddr, utxoWithAmount(5_000_000), fee)
		require.NoError(t, err)
		require.Equal(t, receiverAddr, output.Addr)
		require.Equal(t, uint64(4_000_000), output.Amount)
		require.Equal(t, []wallet.TokenAmount{
			wallet.NewTokenAmount(wallet.NewToken(policyID, "tok"), 100),
		}, output.Tokens)
	})

	t.Run("amount below fee", func(t *testing.T) {
		_, err := CreateUnknownTokensRefundTxOutput(txBuilder, receiverAddr, utxoWithAmount(fee-1), fee)
		require.ErrorIs(t, err, ErrUnpayableUnknownTokensRefund)
	})

	t.Run("amount equal to fee", func(t *testing.T) {
		_, err := CreateUnknownTokensRefundTxOutput(txBuilder, receiverAddr, utxoWithAmount(fee), fee)
		require.ErrorIs(t, err, ErrUnpayableUnknownTokensRefund)
	})

	t.Run("amount after fee below min utxo", func(t *testing.T) {
		_, err := CreateUnknownTokensRefundTxOutput(txBuilder, receiverAddr, utxoWithAmount(fee+500_000), fee)
		require.ErrorIs(t, err, ErrUnpayableUnknownTokensRefund)
	})
}
