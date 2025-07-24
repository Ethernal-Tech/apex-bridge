package bridgingaddressscoordinator

import (
	"testing"

	bam "github.com/Ethernal-Tech/apex-bridge/bridging_addresses_manager"
	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
	cardanowallet "github.com/Ethernal-Tech/cardano-infrastructure/wallet"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestBridgingAddressesCoordinator(t *testing.T) {
	chainID := uint8(common.ChainIDIntPrime)

	t.Run("GetAddressesAndAmountsToPayFrom 1 address currency", func(t *testing.T) {
		bridgingAddressesManagerMock := &bam.BridgingAddressesManagerMock{}
		bridgingAddressesManagerMock.On("GetAllPaymentAddresses", mock.Anything).Return([]string{"addr1"}, nil)
		dbMock := &indexer.DatabaseMock{}
		dbMock.On("GetAllTxOutputs", mock.Anything, true).
			Return([]*indexer.TxInputOutput{
				{
					Input: indexer.TxInput{
						Hash: indexer.NewHashFromHexString("0x0012"),
					},
					Output: indexer.TxOutput{
						Amount: 1_000_000_000,
					},
				},
			}, error(nil)).Once()

		coordinator := NewBridgingAddressesCoordinator(bridgingAddressesManagerMock, map[string]indexer.Database{
			"prime": dbMock,
		})

		amounts, err := coordinator.GetAddressesAndAmountsToPayFrom(chainID, "", []byte{}, []cardanowallet.TxOutput{
			{
				Addr:   "addr1",
				Amount: 10_000_000,
			},
		})
		require.NoError(t, err)
		require.Equal(t, "addr1", amounts[0].Address)
		require.Equal(t, uint64(10_000_000), amounts[0].TokensAmounts[cardanowallet.AdaTokenName])
		require.Equal(t, uint8(0), amounts[0].AddressIndex)
	})

	t.Run("GetAddressesAndAmountsToPayFrom 1 address currency not enough funds", func(t *testing.T) {
		bridgingAddressesManagerMock := &bam.BridgingAddressesManagerMock{}
		bridgingAddressesManagerMock.On("GetAllPaymentAddresses", mock.Anything).Return([]string{"addr1"}, nil)
		dbMock := &indexer.DatabaseMock{}
		dbMock.On("GetAllTxOutputs", mock.Anything, true).
			Return([]*indexer.TxInputOutput{
				{
					Input: indexer.TxInput{
						Hash: indexer.NewHashFromHexString("0x0012"),
					},
					Output: indexer.TxOutput{
						Amount: 1_000_000,
					},
				},
			}, error(nil)).Once()

		coordinator := NewBridgingAddressesCoordinator(bridgingAddressesManagerMock, map[string]indexer.Database{
			"prime": dbMock,
		})

		amounts, err := coordinator.GetAddressesAndAmountsToPayFrom(chainID, "", []byte{}, []cardanowallet.TxOutput{
			{
				Addr:   "addr1",
				Amount: 10_000_000,
			},
		})
		require.NoError(t, err)
		require.Equal(t, "addr1", amounts[0].Address)
		require.Equal(t, uint64(10_000_000), amounts[0].TokensAmounts[cardanowallet.AdaTokenName])
		require.Equal(t, uint8(0), amounts[0].AddressIndex)
	})

	t.Run("GetAddressesAndAmountsToPayFrom 1 address native", func(t *testing.T) {
		bridgingAddressesManagerMock := &bam.BridgingAddressesManagerMock{}
		bridgingAddressesManagerMock.On("GetAllPaymentAddresses", mock.Anything).Return([]string{"addr1"}, nil)
		dbMock := &indexer.DatabaseMock{}
		dbMock.On("GetAllTxOutputs", mock.Anything, true).
			Return([]*indexer.TxInputOutput{
				{
					Input: indexer.TxInput{
						Hash: indexer.NewHashFromHexString("0x0012"),
					},
					Output: indexer.TxOutput{
						Amount: 1_000_000_000,
						Tokens: []indexer.TokenAmount{
							{
								PolicyID: "policy1",
								Name:     "MyToken",
								Amount:   1000000,
							},
						},
					},
				},
			}, error(nil)).Once()

		coordinator := NewBridgingAddressesCoordinator(bridgingAddressesManagerMock, map[string]indexer.Database{
			"prime": dbMock,
		})

		amounts, err := coordinator.GetAddressesAndAmountsToPayFrom(chainID, "", []byte{}, []cardanowallet.TxOutput{
			{
				Addr:   "addr1",
				Amount: 10_000_000,
				Tokens: []cardanowallet.TokenAmount{
					cardanowallet.NewTokenAmount(cardanowallet.NewToken("policy1", "MyToken"), 1000000),
				},
			},
		})
		require.NoError(t, err)
		require.Equal(t, "addr1", amounts[0].Address)
		require.Equal(t, uint64(10_000_000), amounts[0].TokensAmounts[cardanowallet.AdaTokenName])
		require.Equal(t, uint64(1000000), amounts[0].TokensAmounts["policy1"])
		require.Equal(t, uint8(0), amounts[0].AddressIndex)
	})

}
