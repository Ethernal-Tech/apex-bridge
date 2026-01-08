package successtxprocessors

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"testing"

	brAddrManager "github.com/Ethernal-Tech/apex-bridge/bridging_addresses_manager"
	cardanotx "github.com/Ethernal-Tech/apex-bridge/cardano"
	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/oracle_cardano/core"
	cChain "github.com/Ethernal-Tech/apex-bridge/oracle_common/chain"
	cCore "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
	"github.com/Ethernal-Tech/cardano-infrastructure/sendtx"
	"github.com/Ethernal-Tech/cardano-infrastructure/wallet"
	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/require"
)

func TestSkylineRefundRequestedProcessor(t *testing.T) {
	const (
		utxoMinValue             = 1000000
		minFeeForBridging        = 1000010
		defaultMinFeeForBridging = 2000010
		minFeeForBridgingTokens  = 1000010
		minOperationFee          = 1000010
		primeBridgingAddr        = "addr_test1vq6xsx99frfepnsjuhzac48vl9s2lc9awkvfknkgs89srqqslj660"
		primeBridgingFeeAddr     = "addr_test1vqqj5apwf5npsmudw0ranypkj9jw98t25wk4h83jy5mwypswekttt"
		cardanoBridgingAddr      = "addr_test1wrz24vv4tvfqsywkxn36rv5zagys2d7euafcgt50gmpgqpq4ju9uv"
		cardanoBridgingFeeAddr   = "addr_test1wq5dw0g9mpmjy0xd6g58kncapdf6vgcka9el4llhzwy5vhqz80tcq"
		validPrimeTestAddress    = "addr_test1wrz24vv4tvfqsywkxn36rv5zagys2d7euafcgt50gmpgqpq4ju9uv"
		validCardanoTestAddress  = "addr_test1wrz24vv4tvfqsywkxn36rv5zagys2d7euafcgt50gmpgqpq4ju9uv"

		policyID = "29f8873beb52e126f207a2dfd50f7cff556806b5b4cba9834a7b26a8"
	)

	maxAmountAllowedToBridge := new(big.Int).SetUint64(100000000)

	primeCurrencyID := uint16(1)
	cardanoCurrencyID := uint16(2)
	wrappedTokenPrimeID := uint16(3)
	wrappedTokenCardanoID := uint16(4)

	wrappedTokenPrime, err := wallet.NewTokenWithFullName(
		fmt.Sprintf("%s.%s",
			policyID,
			hex.EncodeToString([]byte("wrappedApex"))), true,
	)
	require.NoError(t, err)

	wrappedTokenAmountPrime := wallet.NewTokenAmount(wrappedTokenPrime, 2_000_000)

	wrappedTokenCardano, err := wallet.NewTokenWithFullName(
		fmt.Sprintf("%s.%s",
			policyID,
			hex.EncodeToString([]byte("wrappedCardano"))), true,
	)
	require.NoError(t, err)

	brAddrManagerMock := &brAddrManager.BridgingAddressesManagerMock{}
	brAddrManagerMock.On("GetAllPaymentAddresses", common.ChainIDIntPrime).Return([]string{primeBridgingAddr}, nil)
	brAddrManagerMock.On("GetPaymentAddressFromIndex", common.ChainIDIntPrime, uint8(0)).Return(primeBridgingAddr, true)
	brAddrManagerMock.On("GetFeeMultisigAddress", common.ChainIDIntPrime).Return(primeBridgingFeeAddr)
	brAddrManagerMock.On("GetAllPaymentAddresses", common.ChainIDIntCardano).Return([]string{cardanoBridgingAddr}, nil)
	brAddrManagerMock.On("GetFeeMultisigAddress", common.ChainIDIntCardano).Return(cardanoBridgingFeeAddr)

	getAppConfig := func(refundEnabled bool) *cCore.AppConfig {
		appConfig := &cCore.AppConfig{
			BridgingAddressesManager: brAddrManagerMock,
			CardanoChains: map[string]*cCore.CardanoChainConfig{
				common.ChainIDStrPrime: {
					CardanoChainConfig: cardanotx.CardanoChainConfig{
						NetworkID:     wallet.TestNetNetwork,
						UtxoMinAmount: utxoMinValue,
						DestinationChains: map[string]common.TokenPairs{
							common.ChainIDStrCardano: []common.TokenPair{
								{SourceTokenID: primeCurrencyID, DestinationTokenID: wrappedTokenCardanoID, TrackSourceToken: true, TrackDestinationToken: true},
								{SourceTokenID: wrappedTokenPrimeID, DestinationTokenID: cardanoCurrencyID, TrackSourceToken: true, TrackDestinationToken: true},
							},
						},
						Tokens: map[uint16]common.Token{
							primeCurrencyID:     {ChainSpecific: wallet.AdaTokenName, LockUnlock: true},
							wrappedTokenPrimeID: {ChainSpecific: wrappedTokenPrime.String(), LockUnlock: true, IsWrappedCurrency: true},
						},
						DefaultMinFeeForBridging: defaultMinFeeForBridging,
						MinFeeForBridgingTokens:  minFeeForBridgingTokens,
					},
					MinOperationFee: minOperationFee,
				},
				common.ChainIDStrCardano: {
					CardanoChainConfig: cardanotx.CardanoChainConfig{
						NetworkID:     wallet.TestNetNetwork,
						UtxoMinAmount: utxoMinValue,
						Tokens: map[uint16]common.Token{
							cardanoCurrencyID:     {ChainSpecific: wallet.AdaTokenName, LockUnlock: true},
							wrappedTokenCardanoID: {ChainSpecific: wrappedTokenCardano.String(), LockUnlock: true, IsWrappedCurrency: true},
						},
						DefaultMinFeeForBridging: defaultMinFeeForBridging,
						MinFeeForBridgingTokens:  minFeeForBridgingTokens,
					},
					MinOperationFee: minOperationFee,
				},
			},
			BridgingSettings: cCore.BridgingSettings{
				MaxReceiversPerBridgingRequest: 3,
				MaxAmountAllowedToBridge:       maxAmountAllowedToBridge,
			},
			TryCountLimits: cCore.TryCountLimits{
				MaxRefundTryCount: 3,
			},
			RefundEnabled:    refundEnabled,
			ChainIDConverter: common.NewTestChainIDConverter(),
		}
		appConfig.FillOut()

		return appConfig
	}

	getChainInfos := func() map[string]*cChain.CardanoChainInfo {
		appConfig := getAppConfig(true)
		chainInfos := make(map[string]*cChain.CardanoChainInfo, len(appConfig.CardanoChains))

		for _, cc := range appConfig.CardanoChains {
			info := cChain.NewCardanoChainInfo(cc)

			info.ProtocolParams = protocolParameters

			chainInfos[cc.ChainID] = info
		}

		return chainInfos
	}

	proc := NewRefundRequestProcessorSkyline(hclog.NewNullLogger(), getChainInfos())

	t.Run("ValidateAndAddClaim empty tx", func(t *testing.T) {
		claims := &cCore.BridgeClaims{}

		appConfig := getAppConfig(true)

		err := proc.ValidateAndAddClaim(claims, &core.CardanoTx{}, appConfig)
		require.ErrorContains(t, err, "failed to unmarshal metadata")
	})

	t.Run("HandleBridgingProcessorPreValidate - batchTryCount over", func(t *testing.T) {
		appConfig := getAppConfig(false)

		err := proc.HandleBridgingProcessorPreValidate(&core.CardanoTx{BatchTryCount: 1}, appConfig)
		require.ErrorContains(t, err, "try count exceeded")
	})

	t.Run("HandleBridgingProcessorPreValidate - submitTryCount over", func(t *testing.T) {
		appConfig := getAppConfig(false)

		err := proc.HandleBridgingProcessorPreValidate(&core.CardanoTx{SubmitTryCount: 1}, appConfig)
		require.ErrorContains(t, err, "try count exceeded")
	})

	t.Run("HandleBridgingProcessorError - empty ty", func(t *testing.T) {
		appConfig := getAppConfig(false)

		err := proc.HandleBridgingProcessorError(
			&cCore.BridgeClaims{}, &core.CardanoTx{}, appConfig, nil, "")
		require.ErrorContains(t, err, "failed to unmarshal metadata, err: EOF")
	})

	t.Run("ValidateAndAddClaim invalid sender address", func(t *testing.T) {
		relevantButNotFullMetadata, err := common.SimulateRealMetadata(common.MetadataEncodingTypeCbor, common.BaseMetadata{
			BridgingTxType: common.BridgingTxTypeBridgingRequest,
		})
		require.NoError(t, err)
		require.NotNil(t, relevantButNotFullMetadata)

		claims := &cCore.BridgeClaims{}

		appConfig := getAppConfig(true)

		err = proc.ValidateAndAddClaim(claims, &core.CardanoTx{
			Tx: indexer.Tx{
				Metadata: relevantButNotFullMetadata,
			},
			OriginChainID: common.ChainIDStrPrime,
		}, appConfig)
		require.ErrorContains(t, err, "invalid sender addr")
	})

	t.Run("ValidateAndAddClaim insufficient metadata", func(t *testing.T) {
		relevantButNotFullMetadata, err := common.SimulateRealMetadata(common.MetadataEncodingTypeCbor, common.BridgingRequestMetadata{
			BridgingTxType:     sendtx.BridgingRequestType(common.BridgingTxTypeBridgingRequest),
			DestinationChainID: "invalid",
			SenderAddr:         []string{validPrimeTestAddress},
			Transactions:       []sendtx.BridgingRequestMetadataTransaction{},
		})

		require.NoError(t, err)
		require.NotNil(t, relevantButNotFullMetadata)

		claims := &cCore.BridgeClaims{}

		appConfig := getAppConfig(true)

		txOutputs := []*indexer.TxOutput{
			{
				Address: primeBridgingAddr,
				Amount:  10_000_000,
				Tokens: []indexer.TokenAmount{
					{
						PolicyID: wrappedTokenAmountPrime.PolicyID,
						Name:     wrappedTokenAmountPrime.Name,
						Amount:   2_000_000,
					},
				},
			},
		}

		err = proc.ValidateAndAddClaim(claims, &core.CardanoTx{
			Tx: indexer.Tx{
				Metadata: relevantButNotFullMetadata,
				Outputs:  txOutputs,
			},
			OriginChainID: common.ChainIDStrPrime,
		}, appConfig)
		require.NoError(t, err)

		require.Len(t, claims.RefundRequestClaims, 1)
		require.Equal(t, appConfig.ChainIDConverter.ToNumChainID(common.ChainIDStrPrime), claims.RefundRequestClaims[0].OriginChainId)
		require.Equal(t, uint8(0), claims.RefundRequestClaims[0].DestinationChainId)
		require.Equal(t, uint64(10_000_000), claims.RefundRequestClaims[0].TokenAmounts[0].AmountCurrency.Uint64())
		require.Equal(t, uint64(2_000_000), claims.RefundRequestClaims[0].TokenAmounts[0].AmountTokens.Uint64())
		require.Equal(t, wrappedTokenPrimeID, claims.RefundRequestClaims[0].TokenAmounts[0].TokenId)
		require.Equal(t, uint64(0), claims.RefundRequestClaims[0].OriginAmount.Uint64())
		require.Equal(t, uint64(0), claims.RefundRequestClaims[0].OriginWrappedAmount.Uint64())
		require.Empty(t, claims.RefundRequestClaims[0].OutputIndexes)
	})

	//nolint:dupl
	t.Run("ValidateAndAddClaim unsuported sender chainID", func(t *testing.T) {
		metadata, err := common.SimulateRealMetadata(common.MetadataEncodingTypeCbor, common.BridgingRequestMetadata{
			BridgingTxType:     sendtx.BridgingRequestType(common.BridgingTxTypeBridgingRequest),
			DestinationChainID: common.ChainIDStrCardano,
			SenderAddr:         []string{validPrimeTestAddress},
			Transactions:       []sendtx.BridgingRequestMetadataTransaction{},
		})
		require.NoError(t, err)
		require.NotNil(t, metadata)

		claims := &cCore.BridgeClaims{}

		appConfig := getAppConfig(true)

		txOutputs := []*indexer.TxOutput{
			{Address: "addr1", Amount: 1},
			{Address: "addr2", Amount: 2},
			{Address: primeBridgingAddr, Amount: 3},
			{Address: primeBridgingFeeAddr, Amount: 4},
		}

		tx := indexer.Tx{
			Metadata: metadata,
			Outputs:  txOutputs,
		}

		err = proc.ValidateAndAddClaim(claims, &core.CardanoTx{
			Tx:            tx,
			OriginChainID: "invalid",
		}, appConfig)
		require.ErrorContains(t, err, "unsupported chain id found in tx")
	})

	t.Run("ValidateAndAddClaim invalid sender address", func(t *testing.T) {
		metadata, err := common.SimulateRealMetadata(common.MetadataEncodingTypeCbor, common.BridgingRequestMetadata{
			BridgingTxType:     sendtx.BridgingRequestType(common.BridgingTxTypeBridgingRequest),
			DestinationChainID: common.ChainIDStrCardano,
			SenderAddr:         []string{"invalid_address"},
			Transactions:       []sendtx.BridgingRequestMetadataTransaction{},
		})
		require.NoError(t, err)
		require.NotNil(t, metadata)

		claims := &cCore.BridgeClaims{}

		appConfig := getAppConfig(true)

		txOutputs := []*indexer.TxOutput{
			{Address: "addr1", Amount: 1},
			{Address: "addr2", Amount: 2},
			{Address: primeBridgingAddr, Amount: 3},
			{Address: primeBridgingFeeAddr, Amount: 4},
		}

		tx := indexer.Tx{
			Metadata: metadata,
			Outputs:  txOutputs,
		}

		err = proc.ValidateAndAddClaim(claims, &core.CardanoTx{
			Tx:            tx,
			OriginChainID: common.ChainIDStrPrime,
		}, appConfig)
		require.ErrorContains(t, err, "invalid sender addr")
	})

	t.Run("ValidateAndAddClaim outputs contains more unknown tokens than allowed", func(t *testing.T) {
		metadata, err := common.SimulateRealMetadata(common.MetadataEncodingTypeCbor, common.BridgingRequestMetadata{
			BridgingTxType:     sendtx.BridgingRequestType(common.BridgingTxTypeBridgingRequest),
			DestinationChainID: common.ChainIDStrCardano,
			SenderAddr:         []string{validPrimeTestAddress},
			Transactions:       []sendtx.BridgingRequestMetadataTransaction{},
		})
		require.NoError(t, err)
		require.NotNil(t, metadata)

		claims := &cCore.BridgeClaims{}

		appConfig := getAppConfig(true)

		txOutputs := []*indexer.TxOutput{
			{
				Address: primeBridgingAddr,
				Amount:  1,
				Tokens: []indexer.TokenAmount{
					{
						PolicyID: "1",
						Name:     "1",
						Amount:   1_000_000,
					},
				},
			},
			{
				Address: primeBridgingAddr,
				Amount:  1_000_000,
				Tokens: []indexer.TokenAmount{
					{
						PolicyID: "3",
						Name:     "1",
						Amount:   2_000_000,
					},
				},
			},
			{
				Address: primeBridgingAddr,
				Amount:  1_000_000,
				Tokens: []indexer.TokenAmount{
					{
						PolicyID: "3",
						Name:     "3",
						Amount:   3_000_000,
					},
				},
			},
			{
				Address: primeBridgingAddr,
				Amount:  1_000_000,
				Tokens: []indexer.TokenAmount{
					{
						PolicyID: "3",
						Name:     "1",
						Amount:   2_000_000,
					},
				},
			},
			{Address: "addr2", Amount: 2_000_000},
			{
				Address: primeBridgingAddr,
				Amount:  3_000_000,
				Tokens: []indexer.TokenAmount{
					{
						PolicyID: "1",
						Name:     "3",
						Amount:   100_000,
					},
				},
			},
			{Address: primeBridgingFeeAddr, Amount: 4},
		}

		tx := indexer.Tx{
			Metadata: metadata,
			Outputs:  txOutputs,
		}

		err = proc.ValidateAndAddClaim(claims, &core.CardanoTx{
			Tx:            tx,
			OriginChainID: common.ChainIDStrPrime,
		}, appConfig)
		require.ErrorContains(t, err, "more UTxOs with unknown tokens than allowed")
	})

	t.Run("ValidateAndAddClaim sum of amounts less than the minimum required - currency", func(t *testing.T) {
		metadata, err := common.SimulateRealMetadata(common.MetadataEncodingTypeCbor, common.BridgingRequestMetadata{
			BridgingTxType:     sendtx.BridgingRequestType(common.BridgingTxTypeBridgingRequest),
			DestinationChainID: common.ChainIDStrCardano,
			SenderAddr:         []string{validPrimeTestAddress},
			Transactions:       []sendtx.BridgingRequestMetadataTransaction{},
		})
		require.NoError(t, err)
		require.NotNil(t, metadata)

		claims := &cCore.BridgeClaims{}

		appConfig := getAppConfig(true)

		txOutputs := []*indexer.TxOutput{
			{Address: "addr1", Amount: 500_000},
			{Address: "addr2", Amount: 500_000},
			{
				Address: primeBridgingAddr,
				Amount:  minFeeForBridgingTokens + 1_500_000,
			},
			{Address: primeBridgingFeeAddr, Amount: 600_000},
		}

		tx := indexer.Tx{
			Metadata: metadata,
			Outputs:  txOutputs,
		}

		err = proc.ValidateAndAddClaim(claims, &core.CardanoTx{
			Tx:            tx,
			OriginChainID: common.ChainIDStrPrime,
		}, appConfig)
		require.ErrorContains(t, err, "less than the minimum required for refund")
	})

	t.Run("ValidateAndAddClaim sum of amounts less than the minimum required - token", func(t *testing.T) {
		metadata, err := common.SimulateRealMetadata(common.MetadataEncodingTypeCbor, common.BridgingRequestMetadata{
			BridgingTxType:     sendtx.BridgingRequestType(common.BridgingTxTypeBridgingRequest),
			DestinationChainID: common.ChainIDStrCardano,
			SenderAddr:         []string{validPrimeTestAddress},
			Transactions:       []sendtx.BridgingRequestMetadataTransaction{},
		})
		require.NoError(t, err)
		require.NotNil(t, metadata)

		claims := &cCore.BridgeClaims{}

		appConfig := getAppConfig(true)

		txOutputs := []*indexer.TxOutput{
			{Address: "addr1", Amount: 500_000},
			{Address: "addr2", Amount: 500_000},
			{
				Address: primeBridgingAddr,
				Amount:  minFeeForBridgingTokens + 1_000_000,
				Tokens: []indexer.TokenAmount{
					{
						PolicyID: wrappedTokenPrime.PolicyID,
						Name:     wrappedTokenPrime.Name,
						Amount:   wrappedTokenAmountPrime.Amount,
					},
				},
			},
			{Address: primeBridgingFeeAddr, Amount: 600_000},
		}

		tx := indexer.Tx{
			Metadata: metadata,
			Outputs:  txOutputs,
		}

		err = proc.ValidateAndAddClaim(claims, &core.CardanoTx{
			Tx:            tx,
			OriginChainID: common.ChainIDStrPrime,
		}, appConfig)
		require.ErrorContains(t, err, "less than the minimum required for refund")
	})

	t.Run("ValidateAndAddClaim try count exceeded", func(t *testing.T) {
		validMetadata, err := common.SimulateRealMetadata(common.MetadataEncodingTypeCbor, common.BridgingRequestMetadata{
			BridgingTxType:     sendtx.BridgingRequestType(common.BridgingTxTypeBridgingRequest),
			DestinationChainID: common.ChainIDStrCardano,
			SenderAddr:         []string{validPrimeTestAddress},
			Transactions:       []sendtx.BridgingRequestMetadataTransaction{},
		})
		require.NoError(t, err)
		require.NotNil(t, validMetadata)

		claims := &cCore.BridgeClaims{}

		appConfig := getAppConfig(true)

		txOutputs := []*indexer.TxOutput{
			{Address: "addr1", Amount: 500_000},
			{Address: "addr2", Amount: 500_000},
			{
				Address: primeBridgingAddr,
				Amount:  minFeeForBridgingTokens + 1_500_000,
				Tokens: []indexer.TokenAmount{
					{
						PolicyID: wrappedTokenPrime.PolicyID,
						Name:     wrappedTokenPrime.Name,
						Amount:   wrappedTokenAmountPrime.Amount,
					},
				},
			},
			{Address: primeBridgingFeeAddr, Amount: 1_000_000},
		}

		tx := indexer.Tx{
			Metadata: validMetadata,
			Outputs:  txOutputs,
		}

		err = proc.ValidateAndAddClaim(claims, &core.CardanoTx{
			Tx:             tx,
			OriginChainID:  common.ChainIDStrPrime,
			RefundTryCount: 4,
		}, appConfig)
		require.ErrorContains(t, err, "try count exceeded")
	})

	t.Run("ValidateAndAddClaim outputs contains both valid and invalid UTXOs", func(t *testing.T) {
		metadata, err := common.SimulateRealMetadata(common.MetadataEncodingTypeCbor, common.BridgingRequestMetadata{
			BridgingTxType:     sendtx.BridgingRequestType(common.BridgingTxTypeBridgingRequest),
			DestinationChainID: common.ChainIDStrCardano,
			SenderAddr:         []string{validPrimeTestAddress},
			Transactions:       []sendtx.BridgingRequestMetadataTransaction{},
		})
		require.NoError(t, err)
		require.NotNil(t, metadata)

		claims := &cCore.BridgeClaims{}

		appConfig := getAppConfig(true)
		chainIDConverter := appConfig.ChainIDConverter

		txOutputs := []*indexer.TxOutput{
			{
				Address: primeBridgingAddr,
				Amount:  1,
				Tokens: []indexer.TokenAmount{
					{
						PolicyID: wrappedTokenAmountPrime.PolicyID,
						Name:     wrappedTokenAmountPrime.Name,
						Amount:   1_000_000,
					},
				},
			},
			{
				Address: primeBridgingAddr,
				Amount:  1_000_000,
				Tokens: []indexer.TokenAmount{
					{
						PolicyID: wrappedTokenAmountPrime.PolicyID,
						Name:     wrappedTokenAmountPrime.Name,
						Amount:   2_000_000,
					},
				},
			},
			{
				Address: primeBridgingAddr,
				Amount:  1_000_000,
				Tokens: []indexer.TokenAmount{
					{
						PolicyID: wrappedTokenAmountPrime.PolicyID,
						Name:     wrappedTokenAmountPrime.Name,
						Amount:   3_000_000,
					},
				},
			},
			{
				Address: primeBridgingAddr,
				Amount:  1_000_000,
				Tokens: []indexer.TokenAmount{
					{
						PolicyID: wrappedTokenAmountPrime.PolicyID,
						Name:     wrappedTokenAmountPrime.Name,
						Amount:   2_000_000,
					},
				},
			},
			{Address: "addr2", Amount: 2_000_000},
			{
				Address: primeBridgingAddr,
				Amount:  3_000_000,
				Tokens: []indexer.TokenAmount{
					{
						PolicyID: wrappedTokenCardano.PolicyID,
						Name:     wrappedTokenCardano.Name, // invalid for sender chain
						Amount:   100_000,
					},
				},
			},
			{Address: primeBridgingFeeAddr, Amount: 4},
		}

		tx := indexer.Tx{
			Metadata: metadata,
			Outputs:  txOutputs,
		}

		err = proc.ValidateAndAddClaim(claims, &core.CardanoTx{
			Tx:            tx,
			OriginChainID: common.ChainIDStrPrime,
		}, appConfig)
		require.NoError(t, err)

		require.Len(t, claims.RefundRequestClaims, 1)
		require.Equal(t, chainIDConverter.ToNumChainID(common.ChainIDStrPrime), claims.RefundRequestClaims[0].OriginChainId)
		require.Equal(t, chainIDConverter.ToNumChainID(common.ChainIDStrCardano), claims.RefundRequestClaims[0].DestinationChainId)
		require.Equal(t, uint64(1+1_000_000+1_000_000+1_000_000+3_000_000), claims.RefundRequestClaims[0].TokenAmounts[0].AmountCurrency.Uint64())
		require.Equal(t, uint64(1_000_000+2_000_000+3_000_000+2_000_000), claims.RefundRequestClaims[0].TokenAmounts[0].AmountTokens.Uint64())
		require.Equal(t, wrappedTokenPrimeID, claims.RefundRequestClaims[0].TokenAmounts[0].TokenId)
		require.Equal(t, uint64(0), claims.RefundRequestClaims[0].OriginAmount.Uint64())
		require.Equal(t, uint64(0), claims.RefundRequestClaims[0].OriginWrappedAmount.Uint64())
		require.Equal(t, common.PackNumbersToBytes([]uint16{5}), claims.RefundRequestClaims[0].OutputIndexes)
	})

	t.Run("ValidateAndAddClaim valid", func(t *testing.T) {
		validMetadata, err := common.SimulateRealMetadata(common.MetadataEncodingTypeCbor, common.BridgingRequestMetadata{
			BridgingTxType:     sendtx.BridgingRequestType(common.BridgingTxTypeBridgingRequest),
			DestinationChainID: common.ChainIDStrCardano,
			SenderAddr:         []string{validPrimeTestAddress},
			Transactions:       []sendtx.BridgingRequestMetadataTransaction{},
		})
		require.NoError(t, err)
		require.NotNil(t, validMetadata)

		claims := &cCore.BridgeClaims{}

		appConfig := getAppConfig(true)
		chainIDConverter := appConfig.ChainIDConverter

		txOutputs := []*indexer.TxOutput{
			{Address: "addr1", Amount: 500_000},
			{Address: "addr2", Amount: 500_000},
			{
				Address: primeBridgingAddr,
				Amount:  minFeeForBridgingTokens + 1_500_000,
				Tokens: []indexer.TokenAmount{
					{
						PolicyID: wrappedTokenPrime.PolicyID,
						Name:     wrappedTokenPrime.Name,
						Amount:   wrappedTokenAmountPrime.Amount,
					},
				},
			},
			{Address: primeBridgingFeeAddr, Amount: 1_000_000},
		}

		tx := indexer.Tx{
			Metadata: validMetadata,
			Outputs:  txOutputs,
		}

		err = proc.ValidateAndAddClaim(claims, &core.CardanoTx{
			Tx:            tx,
			OriginChainID: common.ChainIDStrPrime,
		}, appConfig)
		require.NoError(t, err)

		require.Len(t, claims.RefundRequestClaims, 1)
		require.Equal(t, chainIDConverter.ToNumChainID(common.ChainIDStrPrime), claims.RefundRequestClaims[0].OriginChainId)
		require.Equal(t, chainIDConverter.ToNumChainID(common.ChainIDStrCardano), claims.RefundRequestClaims[0].DestinationChainId)
		require.Equal(t, uint64(minFeeForBridgingTokens+1_500_000), claims.RefundRequestClaims[0].TokenAmounts[0].AmountCurrency.Uint64())
		require.Equal(t, wrappedTokenAmountPrime.Amount, claims.RefundRequestClaims[0].TokenAmounts[0].AmountTokens.Uint64())
		require.Equal(t, wrappedTokenPrimeID, claims.RefundRequestClaims[0].TokenAmounts[0].TokenId)
		require.Equal(t, uint64(minFeeForBridgingTokens+1_500_000), claims.RefundRequestClaims[0].OriginAmount.Uint64())
		require.Equal(t, wrappedTokenAmountPrime.Amount, claims.RefundRequestClaims[0].OriginWrappedAmount.Uint64())
		require.Empty(t, claims.RefundRequestClaims[0].OutputIndexes)
	})

	t.Run("ValidateAndAddClaim valid - currency only", func(t *testing.T) {
		validMetadata, err := common.SimulateRealMetadata(common.MetadataEncodingTypeCbor, common.BridgingRequestMetadata{
			BridgingTxType:     sendtx.BridgingRequestType(common.BridgingTxTypeBridgingRequest),
			DestinationChainID: common.ChainIDStrCardano,
			SenderAddr:         []string{validPrimeTestAddress},
			Transactions:       []sendtx.BridgingRequestMetadataTransaction{},
		})
		require.NoError(t, err)
		require.NotNil(t, validMetadata)

		claims := &cCore.BridgeClaims{}

		appConfig := getAppConfig(true)
		chainIDConverter := appConfig.ChainIDConverter

		txOutputs := []*indexer.TxOutput{
			{Address: "addr1", Amount: 500_000},
			{Address: "addr2", Amount: 500_000},
			{
				Address: primeBridgingAddr,
				Amount:  minFeeForBridgingTokens + 2_500_000,
			},
			{Address: primeBridgingFeeAddr, Amount: 1_000_000},
		}

		tx := indexer.Tx{
			Metadata: validMetadata,
			Outputs:  txOutputs,
		}

		err = proc.ValidateAndAddClaim(claims, &core.CardanoTx{
			Tx:            tx,
			OriginChainID: common.ChainIDStrPrime,
		}, appConfig)
		require.NoError(t, err)

		require.Len(t, claims.RefundRequestClaims, 1)
		require.Equal(t, chainIDConverter.ToNumChainID(common.ChainIDStrPrime), claims.RefundRequestClaims[0].OriginChainId)
		require.Equal(t, chainIDConverter.ToNumChainID(common.ChainIDStrCardano), claims.RefundRequestClaims[0].DestinationChainId)
		require.Equal(t, uint64(minFeeForBridgingTokens+2_500_000), claims.RefundRequestClaims[0].TokenAmounts[0].AmountCurrency.Uint64())
		require.Equal(t, uint64(0), claims.RefundRequestClaims[0].TokenAmounts[0].AmountTokens.Uint64())
		require.Equal(t, primeCurrencyID, claims.RefundRequestClaims[0].TokenAmounts[0].TokenId)
		require.Equal(t, uint64(minFeeForBridgingTokens+2_500_000), claims.RefundRequestClaims[0].OriginAmount.Uint64())
		require.Equal(t, uint64(0), claims.RefundRequestClaims[0].OriginWrappedAmount.Uint64())
		require.Empty(t, claims.RefundRequestClaims[0].OutputIndexes)
	})
}
