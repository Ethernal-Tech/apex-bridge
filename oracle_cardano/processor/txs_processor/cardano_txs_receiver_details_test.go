package processor

import (
	"math/big"
	"testing"

	cardanotx "github.com/Ethernal-Tech/apex-bridge/cardano"
	"github.com/Ethernal-Tech/apex-bridge/common"
	cCore "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	"github.com/Ethernal-Tech/cardano-infrastructure/sendtx"
	cardanowallet "github.com/Ethernal-Tech/cardano-infrastructure/wallet"
	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/require"
)

func TestCardanoGetBridgingRequestStateDetails(t *testing.T) {
	const (
		originChainID = common.ChainIDStrPrime
		currencyID    = uint16(1)
		wrappedID     = uint16(2)
	)

	newReceiver := func(chains map[string]*cCore.CardanoChainConfig) *CardanoTxsReceiverImpl {
		return NewCardanoTxsReceiverImpl(
			&cCore.AppConfig{CardanoChains: chains}, nil, nil, nil, hclog.NewNullLogger())
	}

	chains := map[string]*cCore.CardanoChainConfig{
		originChainID: {
			CardanoChainConfig: cardanotx.CardanoChainConfig{
				Tokens: map[uint16]common.Token{
					currencyID: {ChainSpecific: cardanowallet.AdaTokenName},
					wrappedID:  {ChainSpecific: "wrapped", IsWrappedCurrency: true},
				},
			},
		},
	}

	t.Run("reads amounts, sender and destination out of the metadata", func(t *testing.T) {
		metadata, err := common.SimulateRealMetadata(
			common.MetadataEncodingTypeCbor, common.BridgingRequestMetadata{
				BridgingTxType:     sendtx.BridgingRequestType(common.BridgingTxTypeBridgingRequest),
				DestinationChainID: common.ChainIDStrVector,
				SenderAddr:         []string{"addr_part1", "addr_part2"},
				Transactions: []sendtx.BridgingRequestMetadataTransaction{
					{Address: []string{"receiver1"}, Amount: 100, TokenID: currencyID},
					{Address: []string{"receiver2"}, Amount: 70, TokenID: wrappedID},
				},
				BridgingFee:  10,
				OperationFee: 5,
			})
		require.NoError(t, err)

		dstChainID, details := newReceiver(chains).getBridgingRequestStateDetails(originChainID, metadata)

		require.Equal(t, common.ChainIDStrVector, dstChainID)
		require.NotNil(t, details)
		require.Equal(t, "addr_part1addr_part2", details.SenderAddr)
		require.Equal(t, common.DfmToWei(big.NewInt(100)), details.Amount)
		require.Equal(t, common.DfmToWei(big.NewInt(70)), details.TokenAmount)
		require.Equal(t, wrappedID, details.TokenID)
		require.Equal(t, common.DfmToWei(big.NewInt(10)), details.BridgingFee)
		require.Equal(t, common.DfmToWei(big.NewInt(5)), details.OperationFee)

		require.Len(t, details.Receivers, 2)
		require.Equal(t, "receiver1", details.Receivers[0].Address)
		require.Equal(t, common.DfmToWei(big.NewInt(100)), details.Receivers[0].Amount)
		require.Equal(t, currencyID, details.Receivers[0].TokenID)
	})

	t.Run("malformed metadata yields no details instead of an error", func(t *testing.T) {
		dstChainID, details := newReceiver(chains).getBridgingRequestStateDetails(
			originChainID, []byte{0xFF, 0xFF})

		require.Empty(t, dstChainID)
		require.Nil(t, details)
	})

	t.Run("unknown chain yields no details instead of an error", func(t *testing.T) {
		dstChainID, details := newReceiver(nil).getBridgingRequestStateDetails(originChainID, nil)

		require.Empty(t, dstChainID)
		require.Nil(t, details)
	})
}
