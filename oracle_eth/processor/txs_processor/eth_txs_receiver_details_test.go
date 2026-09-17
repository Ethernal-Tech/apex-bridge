package processor

import (
	"math/big"
	"testing"

	"github.com/Ethernal-Tech/apex-bridge/common"
	oCore "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	"github.com/Ethernal-Tech/apex-bridge/oracle_eth/core"
	cardanowallet "github.com/Ethernal-Tech/cardano-infrastructure/wallet"
	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/require"
)

func TestEthGetBridgingRequestStateDetails(t *testing.T) {
	const (
		originChainID = common.ChainIDStrNexus
		currencyID    = uint16(1)
		wrappedID     = uint16(2)
	)

	newReceiver := func(chains map[string]*oCore.EthChainConfig) *EthTxsReceiverImpl {
		return NewEthTxsReceiverImpl(
			&oCore.AppConfig{EthChains: chains}, nil, nil, nil, hclog.NewNullLogger())
	}

	chains := map[string]*oCore.EthChainConfig{
		originChainID: {
			Tokens: map[uint16]common.Token{
				currencyID: {ChainSpecific: cardanowallet.AdaTokenName},
				wrappedID:  {ChainSpecific: "wrapped", IsWrappedCurrency: true},
			},
		},
	}

	t.Run("reads amounts, sender and destination out of the metadata", func(t *testing.T) {
		metadata, err := core.MarshalEthMetadata(core.BridgingRequestEthMetadata{
			BridgingTxType:     common.BridgingTxTypeBridgingRequest,
			DestinationChainID: common.ChainIDStrPrime,
			SenderAddr:         "0xsender",
			Transactions: []core.BridgingRequestEthMetadataTransaction{
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
		require.Equal(t, "0xsender", details.SenderAddr)
		// eth metadata amounts are already wei, they must not be converted
		require.Equal(t, big.NewInt(100), details.Amount)
		require.Equal(t, big.NewInt(70), details.TokenAmount)
		require.Equal(t, wrappedID, details.TokenID)
		require.Equal(t, big.NewInt(10), details.BridgingFee)
		require.Equal(t, big.NewInt(5), details.OperationFee)
		require.Len(t, details.Receivers, 2)
	})

	// nexus configs carry colored coins that are neither the currency nor the wrapped currency;
	// they still have to reach the bridging history
	t.Run("counts a colored coin as the token amount", func(t *testing.T) {
		const coloredID = uint16(3)

		chains[originChainID].Tokens[coloredID] = common.Token{ChainSpecific: "0x11"}

		metadata, err := core.MarshalEthMetadata(core.BridgingRequestEthMetadata{
			BridgingTxType:     common.BridgingTxTypeBridgingRequest,
			DestinationChainID: common.ChainIDStrPrime,
			SenderAddr:         "0xsender",
			Transactions: []core.BridgingRequestEthMetadataTransaction{
				{Address: "receiver1", Amount: big.NewInt(70), TokenID: coloredID},
			},
			BridgingFee:  big.NewInt(10),
			OperationFee: big.NewInt(5),
		})
		require.NoError(t, err)

		_, details := newReceiver(chains).getBridgingRequestStateDetails(originChainID, metadata)

		require.NotNil(t, details)
		require.Equal(t, big.NewInt(0), details.Amount)
		require.Equal(t, big.NewInt(70), details.TokenAmount)
		require.Equal(t, coloredID, details.TokenID)
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
