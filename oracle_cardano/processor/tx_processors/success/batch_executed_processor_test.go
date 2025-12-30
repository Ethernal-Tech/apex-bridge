package successtxprocessors

import (
	"testing"

	brAddrManager "github.com/Ethernal-Tech/apex-bridge/bridging_addresses_manager"
	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/oracle_cardano/core"
	cCore "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestBatchExecutedProcessor(t *testing.T) {
	proc := NewBatchExecutedProcessor(hclog.NewNullLogger())

	brAddrManagerMock := &brAddrManager.BridgingAddressesManagerMock{}
	brAddrManagerMock.On("GetAllPaymentAddresses", mock.Anything).Return([]string{"addr_bridging"}, nil)
	brAddrManagerMock.On("GetFeeMultisigAddress", mock.Anything).Return("addr_fee")

	appConfig := cCore.AppConfig{
		BridgingAddressesManager: brAddrManagerMock,
		CardanoChains:            map[string]*cCore.CardanoChainConfig{common.ChainIDStrPrime: {}},
		ChainIDConverter:         common.NewChainIDConverterForTest(),
	}
	appConfig.FillOut()

	txInputs := append(make([]*indexer.TxInputOutput, 0), &indexer.TxInputOutput{
		Input: indexer.TxInput{},
		Output: indexer.TxOutput{
			Address: "addr_fee",
		},
	})

	t.Run("ValidateAndAddClaim empty tx", func(t *testing.T) {
		claims := &cCore.BridgeClaims{}

		err := proc.ValidateAndAddClaim(claims, &core.CardanoTx{}, &appConfig)
		require.Error(t, err)
	})

	t.Run("ValidateAndAddClaim irrelevant metadata", func(t *testing.T) {
		irrelevantMetadata, err := common.SimulateRealMetadata(common.MetadataEncodingTypeCbor, common.BaseMetadata{
			BridgingTxType: common.BridgingTxTypeBridgingRequest,
		})
		require.NoError(t, err)
		require.NotNil(t, irrelevantMetadata)

		claims := &cCore.BridgeClaims{}
		err = proc.ValidateAndAddClaim(claims, &core.CardanoTx{
			Tx: indexer.Tx{
				Metadata: irrelevantMetadata,
			},
		}, &appConfig)
		require.Error(t, err)
	})

	t.Run("ValidateAndAddClaim valid but metadata not full", func(t *testing.T) {
		relevantButNotFullMetadata, err := common.SimulateRealMetadata(common.MetadataEncodingTypeCbor, common.BaseMetadata{
			BridgingTxType: common.BridgingTxTypeBatchExecution,
		})
		require.NoError(t, err)
		require.NotNil(t, relevantButNotFullMetadata)

		claims := &cCore.BridgeClaims{}
		err = proc.ValidateAndAddClaim(claims, &core.CardanoTx{
			OriginChainID: common.ChainIDStrPrime,
			Tx: indexer.Tx{
				Metadata: relevantButNotFullMetadata,
				Inputs:   txInputs,
			},
		}, &appConfig)
		require.NoError(t, err)
		require.True(t, claims.Count() == 1)
		require.Len(t, claims.BatchExecutedClaims, 1)
		require.Equal(t, [32]byte{}, claims.BatchExecutedClaims[0].ObservedTransactionHash)
	})

	t.Run("ValidateAndAddClaim valid full metadata", func(t *testing.T) {
		batchNonceID := uint64(1)
		relevantFullMetadata, err := common.SimulateRealMetadata(common.MetadataEncodingTypeCbor, common.BatchExecutedMetadata{
			BridgingTxType: common.BridgingTxTypeBatchExecution,
			BatchNonceID:   batchNonceID,
		})
		require.NoError(t, err)
		require.NotNil(t, relevantFullMetadata)

		claims := &cCore.BridgeClaims{}
		txHash := indexer.Hash{1, 20}

		txOutputs := []*indexer.TxOutput{
			{Address: "addr_bridging", Amount: 1},
			{Address: "addr_fee", Amount: 2},
		}
		err = proc.ValidateAndAddClaim(claims, &core.CardanoTx{
			OriginChainID: common.ChainIDStrPrime,
			Tx: indexer.Tx{
				Hash:     txHash,
				Metadata: relevantFullMetadata,
				Outputs:  txOutputs,
				Inputs:   txInputs,
			},
		}, &appConfig)
		require.NoError(t, err)
		require.True(t, claims.Count() == 1)
		require.Len(t, claims.BatchExecutedClaims, 1)
		require.Equal(t, txHash[:], claims.BatchExecutedClaims[0].ObservedTransactionHash[:])
		require.Equal(t, batchNonceID, claims.BatchExecutedClaims[0].BatchNonceId)
	})

	t.Run("validate method fail", func(t *testing.T) {
		cardanoChains := make(map[string]*cCore.CardanoChainConfig)

		brAddrManMock := &brAddrManager.BridgingAddressesManagerMock{}
		brAddrManMock.On("GetAllPaymentAddresses", mock.Anything).Return([]string{"addr1"}, nil)
		brAddrManMock.On("GetFeeMultisigAddress", mock.Anything).Return("addr2")

		cardanoChains[common.ChainIDStrPrime] = &cCore.CardanoChainConfig{}

		config := &cCore.AppConfig{
			BridgingAddressesManager: brAddrManMock,
			CardanoChains:            cardanoChains,
			Bridge:                   cCore.BridgeConfig{},
			Settings:                 cCore.AppSettings{},
			BridgingSettings:         cCore.BridgingSettings{},
			ChainIDConverter:         common.NewChainIDConverterForTest(),
		}

		config.FillOut()

		inputs := []*indexer.TxInputOutput{
			{
				Output: indexer.TxOutput{
					Address: "addr3",
					IsUsed:  true,
				},
			},
		}
		tx := core.CardanoTx{
			OriginChainID: common.ChainIDStrPrime,
			Tx: indexer.Tx{
				Inputs: inputs,
			},
		}

		tx.Inputs[0].Output.Address = "addr1"
		err := proc.validate(&tx, &common.BatchExecutedMetadata{}, config)
		require.Error(t, err)
		require.ErrorContains(t, err, "fee address not found in tx inputs")
	})

	t.Run("validate method origin chain not registered", func(t *testing.T) {
		cardanoChains := make(map[string]*cCore.CardanoChainConfig)

		config := &cCore.AppConfig{
			CardanoChains:    cardanoChains,
			Bridge:           cCore.BridgeConfig{},
			Settings:         cCore.AppSettings{},
			BridgingSettings: cCore.BridgingSettings{},
		}

		config.FillOut()

		inputs := []*indexer.TxInputOutput{
			{
				Output: indexer.TxOutput{
					Address: "addr3",
					IsUsed:  true,
				},
			},
		}
		tx := core.CardanoTx{
			OriginChainID: common.ChainIDStrPrime,
			Tx: indexer.Tx{
				Inputs: inputs,
			},
		}

		err := proc.validate(&tx, &common.BatchExecutedMetadata{}, config)
		require.Error(t, err)
		require.ErrorContains(t, err, "unsupported chain id found in tx")
	})

	t.Run("validate method pass", func(t *testing.T) {
		cardanoChains := make(map[string]*cCore.CardanoChainConfig)

		brAddrManMock := &brAddrManager.BridgingAddressesManagerMock{}
		brAddrManMock.On("GetAllPaymentAddresses", mock.Anything).Return([]string{"addr1"}, nil)
		brAddrManMock.On("GetFeeMultisigAddress", mock.Anything).Return("addr2")

		cardanoChains[common.ChainIDStrPrime] = &cCore.CardanoChainConfig{}

		config := &cCore.AppConfig{
			BridgingAddressesManager: brAddrManMock,
			CardanoChains:            cardanoChains,
			Bridge:                   cCore.BridgeConfig{},
			Settings:                 cCore.AppSettings{},
			BridgingSettings:         cCore.BridgingSettings{},
			ChainIDConverter:         common.NewChainIDConverterForTest(),
		}

		config.FillOut()

		inputs := []*indexer.TxInputOutput{
			{
				Output: indexer.TxOutput{
					Address: "addr1",
					IsUsed:  true,
				},
			},
			{
				Output: indexer.TxOutput{
					Address: "addr2",
					IsUsed:  true,
				},
			},
		}
		tx := core.CardanoTx{
			OriginChainID: common.ChainIDStrPrime,
			Tx: indexer.Tx{
				Inputs: inputs,
			},
		}

		err := proc.validate(&tx, &common.BatchExecutedMetadata{}, config)
		require.NoError(t, err)

		tx.Tx.Inputs = append(tx.Tx.Inputs, &indexer.TxInputOutput{
			Output: indexer.TxOutput{
				Address: "addr1",
				IsUsed:  true,
			},
		})
		err = proc.validate(&tx, &common.BatchExecutedMetadata{}, config)
		require.NoError(t, err)

		tx.Tx.Inputs = append(tx.Tx.Inputs, &indexer.TxInputOutput{
			Output: indexer.TxOutput{
				Address: "addr2",
				IsUsed:  true,
			},
		})
		err = proc.validate(&tx, &common.BatchExecutedMetadata{}, config)
		require.NoError(t, err)
	})
}
