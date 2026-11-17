package batcher

import (
	"context"
	"encoding/json"
	"path/filepath"
	"testing"

	bac "github.com/Ethernal-Tech/apex-bridge/bridging_addresses_coordinator"
	bam "github.com/Ethernal-Tech/apex-bridge/bridging_addresses_manager"
	cardano "github.com/Ethernal-Tech/apex-bridge/cardano"
	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
	"github.com/Ethernal-Tech/cardano-infrastructure/secrets"
	secretsHelper "github.com/Ethernal-Tech/cardano-infrastructure/secrets/helper"
	"github.com/Ethernal-Tech/cardano-infrastructure/wallet"
	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/require"
)

func Test_prepareRefundInputsOutputs(t *testing.T) {
	const (
		bridgingAddr = "addr_test1vq6xsx99frfepnsjuhzac48vl9s2lc9awkvfknkgs89srqqslj660"
		feeAddr      = "addr_test1vqqj5apwf5npsmudw0ranypkj9jw98t25wk4h83jy5mwypswekttt"
		receiverAddr = "addr_test1vqjysa7p4mhu0l25qknwznvj0kghtr29ud7zp732ezwtzec0w8g3u"
		policyID     = "29f8873beb52e126f207a2dfd50f7cff556806b5b4cba9834a7b26a8"
		fee          = uint64(1_000_000)
	)

	secretsMngr, err := secretsHelper.CreateSecretsManager(&secrets.SecretsManagerConfig{
		Path: filepath.Join(t.TempDir(), "stp"),
		Type: secrets.Local,
	})
	require.NoError(t, err)

	_, err = cardano.GenerateWallet(secretsMngr, "prime", true, false)
	require.NoError(t, err)

	configRaw := json.RawMessage(`{
		"socketPath": "./socket",
		"testnetMagic": 42,
		"networkID": 0,
		"minUtxoAmount": 1000000,
		"minFeeForBridgingTokens": 1000000
	}`)

	brAddrManagerMock := &bam.BridgingAddressesManagerMock{}
	brAddrManagerMock.On("GetPaymentAddressIndex", common.ChainIDIntPrime, bridgingAddr).Return(uint8(0), true)
	brAddrManagerMock.On("GetFeeMultisigAddress", common.ChainIDIntPrime).Return(feeAddr)

	cco, err := NewCardanoChainOperations(configRaw, common.NewTestChainIDConverter(), &indexer.DatabaseMock{},
		secretsMngr, "prime", brAddrManagerMock, &bac.BridgingAddressesCoordinatorMock{}, hclog.NewNullLogger())
	require.NoError(t, err)
	require.Equal(t, wallet.TestNetNetwork, cco.config.NetworkID)

	protocolParams, err := (&cardano.TxProviderTestMock{ReturnDefaultParameters: true}).
		GetProtocolParameters(context.Background())
	require.NoError(t, err)

	data := &batchInitialData{ChainID: common.ChainIDIntPrime, ProtocolParams: protocolParams}

	refundUtxo := func(index uint32, amount uint64, receiver string) refundUtxoWithReceiverAddr {
		return refundUtxoWithReceiverAddr{
			Utxo: &indexer.TxInputOutput{
				Input: indexer.TxInput{Hash: indexer.Hash{1}, Index: index},
				Output: indexer.TxOutput{
					Address: bridgingAddr,
					Amount:  amount,
					Tokens:  []indexer.TokenAmount{{PolicyID: policyID, Name: "tok", Amount: 100}},
				},
			},
			Addr: receiver,
		}
	}

	t.Run("no refunds", func(t *testing.T) {
		inputs, outputs, err := cco.prepareRefundInputsOutputs(data, nil)
		require.NoError(t, err)
		require.Empty(t, inputs)
		require.Empty(t, outputs)
	})

	t.Run("unpayable refunds are skipped, payable one is kept", func(t *testing.T) {
		inputs, outputs, err := cco.prepareRefundInputsOutputs(data, []refundUtxoWithReceiverAddr{
			refundUtxo(0, 5_000_000, receiverAddr),     // payable
			refundUtxo(1, fee-1, receiverAddr),         // below fee
			refundUtxo(2, fee, receiverAddr),           // zero after fee
			refundUtxo(3, fee+500_000, receiverAddr),   // below min utxo after fee
			refundUtxo(4, 5_000_000, "not_an_address"), // invalid receiver
		})
		require.NoError(t, err)

		require.Len(t, inputs, 1)
		require.Len(t, inputs[0], 1)
		require.Equal(t, uint32(0), inputs[0][0].Input.Index)

		require.Len(t, outputs, 2)
		require.Equal(t, receiverAddr, outputs[0].Addr)
		require.Equal(t, uint64(5_000_000)-fee, outputs[0].Amount)
		require.Len(t, outputs[0].Tokens, 1)
		require.Equal(t, feeAddr, outputs[1].Addr)
		require.Equal(t, fee, outputs[1].Amount)
	})
}
