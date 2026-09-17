package processor

import (
	"math/big"
	"testing"

	"github.com/Ethernal-Tech/apex-bridge/common"
	oCore "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	"github.com/Ethernal-Tech/apex-bridge/oracle_solana/core"
	solana "github.com/Ethernal-Tech/apex-bridge/solana"
	cardanowallet "github.com/Ethernal-Tech/cardano-infrastructure/wallet"
	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/require"
)

func TestSolGetBridgingRequestStateDetails(t *testing.T) {
	const (
		originChainID = common.ChainIDStrSolana
		currencyID    = uint16(1)
		wrappedID     = uint16(2)
	)

	newReceiver := func(chains map[string]*oCore.SolanaChainConfig) *SolEventReceiverImpl {
		return NewSolTxsReceiver(
			&oCore.AppConfig{SolanaChains: chains}, nil, nil, nil, hclog.NewNullLogger())
	}

	chains := map[string]*oCore.SolanaChainConfig{
		originChainID: {
			SolanaChainConfig: solana.SolanaChainConfig{
				Tokens: map[uint16]common.Token{
					currencyID: {ChainSpecific: cardanowallet.AdaTokenName},
					wrappedID:  {ChainSpecific: "wrapped", IsWrappedCurrency: true},
				},
			},
		},
	}

	t.Run("reads amounts, sender and destination out of the metadata", func(t *testing.T) {
		metadata, err := core.MarshalSolMetadata(core.BridgingRequestSolMetadata{
			BridgingTxType:     common.BridgingTxTypeBridgingRequest,
			DestinationChainID: common.ChainIDStrPrime,
			SenderAddr:         "solsender",
			Transactions: []core.BridgingRequestSolMetadataTransaction{
				{Address: "receiver1", Amount: big.NewInt(100), TokenID: currencyID},
				{Address: "receiver2", Amount: big.NewInt(70), TokenID: wrappedID},
			},
			BridgingFee:  big.NewInt(10),
			OperationFee: big.NewInt(5),
		})
		require.NoError(t, err)

		dstChainID, details := newReceiver(chains).getBridgingRequestStateDetails(originChainID, metadata)

		require.Equal(t, common.ChainIDStrPrime, dstChainID)
		require.NotNil(t, details)
		require.Equal(t, "solsender", details.SenderAddr)
		// solana metadata amounts are already wei, they must not be converted
		require.Equal(t, big.NewInt(100), details.Amount)
		require.Equal(t, big.NewInt(70), details.TokenAmount)
		require.Equal(t, wrappedID, details.TokenID)
		require.Equal(t, big.NewInt(10), details.BridgingFee)
		require.Equal(t, big.NewInt(5), details.OperationFee)
		require.Len(t, details.Receivers, 2)
	})

	t.Run("malformed metadata yields no details instead of an error", func(t *testing.T) {
		dstChainID, details := newReceiver(chains).getBridgingRequestStateDetails(
			originChainID, []byte("not json"))

		require.Empty(t, dstChainID)
		require.Nil(t, details)
	})

	t.Run("unknown chain yields no details instead of an error", func(t *testing.T) {
		dstChainID, details := newReceiver(nil).getBridgingRequestStateDetails(originChainID, nil)

		require.Empty(t, dstChainID)
		require.Nil(t, details)
	})
}
