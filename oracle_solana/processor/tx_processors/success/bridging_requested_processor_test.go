package successtxprocessors

import (
	"math/big"
	"testing"

	brAddrManager "github.com/Ethernal-Tech/apex-bridge/bridging_addresses_manager"
	cardanotx "github.com/Ethernal-Tech/apex-bridge/cardano"
	"github.com/Ethernal-Tech/apex-bridge/common"
	oCore "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	"github.com/Ethernal-Tech/apex-bridge/oracle_solana/core"
	solanatx "github.com/Ethernal-Tech/apex-bridge/solana"
	"github.com/Ethernal-Tech/cardano-infrastructure/wallet"
	"github.com/gagliardetto/solana-go"
	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/require"
)

func TestBridgingRequestedProcessor(t *testing.T) {
	const (
		utxoMinValue         = uint64(1000000)
		minColCoinsToBridge  = uint64(100000)
		feeAddrBridgingAmt   = uint64(1000005)
		validCardanoTestAddr = "addr_test1vq6xsx99frfepnsjuhzac48vl9s2lc9awkvfknkgs89srqqslj660"
		validEthTestAddr     = "0xA4d1233A67776575425Ab185f6a9251aa00fEA25"

		solanaCurrencyID = uint16(1)
		primeCurrencyID  = uint16(2)
		nexusCurrencyID  = uint16(3)
	)

	minFeeForBridging := common.LamportToWei(new(big.Int).SetUint64(1000010))
	minOperationFee := common.LamportToWei(new(big.Int).SetUint64(500))

	maxAmountAllowedToBridge := new(big.Int).SetUint64(100_000_000_000)

	primeFeeAddr := "addr_test1vqqj5apwf5npsmudw0ranypkj9jw98t25wk4h83jy5mwypswekttt"

	brAddrManagerMock := &brAddrManager.BridgingAddressesManagerMock{}
	brAddrManagerMock.On("GetAllPaymentAddresses", common.ChainIDIntPrime).Return([]string{validCardanoTestAddr}, nil)
	brAddrManagerMock.On("GetFeeMultisigAddress", common.ChainIDIntPrime).Return(primeFeeAddr)

	getAppConfig := func() *oCore.AppConfig {
		config := &oCore.AppConfig{
			BridgingAddressesManager: brAddrManagerMock,
			CardanoChains: map[string]*oCore.CardanoChainConfig{
				common.ChainIDStrPrime: {
					CardanoChainConfig: cardanotx.CardanoChainConfig{
						NetworkID:     wallet.TestNetNetwork,
						UtxoMinAmount: utxoMinValue,
						Tokens: map[uint16]common.Token{
							primeCurrencyID: {ChainSpecific: wallet.AdaTokenName, LockUnlock: true},
						},
					},
				},
			},
			EthChains: map[string]*oCore.EthChainConfig{
				common.ChainIDStrNexus: {
					BridgingAddresses: oCore.EthBridgingAddresses{
						BridgingAddress: validEthTestAddr,
					},
					Tokens: map[uint16]common.Token{
						nexusCurrencyID: {ChainSpecific: wallet.AdaTokenName, LockUnlock: true},
					},
					MinColCoinsAllowedToBridge: common.LamportToWei(new(big.Int).SetUint64(minColCoinsToBridge)),
					FeeAddrBridgingAmount:      common.LamportToWei(new(big.Int).SetUint64(feeAddrBridgingAmt)),
					MinOperationFee:            minOperationFee,
					MinFeeForBridging:          minFeeForBridging,
				},
			},
			SolanaChains: map[string]*oCore.SolanaChainConfig{
				common.ChainIDStrSolana: {
					SolanaChainConfig: solanatx.SolanaChainConfig{
						DestinationChains: map[string]common.TokenPairs{
							common.ChainIDStrPrime: {
								{SourceTokenID: solanaCurrencyID, DestinationTokenID: primeCurrencyID, TrackSourceToken: true, TrackDestinationToken: true},
							},
							common.ChainIDStrNexus: {
								{SourceTokenID: solanaCurrencyID, DestinationTokenID: nexusCurrencyID, TrackSourceToken: true, TrackDestinationToken: true},
							},
						},
						Tokens: map[uint16]common.Token{
							solanaCurrencyID: {ChainSpecific: wallet.AdaTokenName, LockUnlock: true},
						},
						MinFeeForBridging: common.WeiToLamport(minFeeForBridging),
					},
					FeeAddrBridgingAmount:      feeAddrBridgingAmt,
					MinColCoinsAllowedToBridge: minColCoinsToBridge,
					MinOperationFee:            common.WeiToLamport(minOperationFee).Uint64(),
				},
			},
			BridgingSettings: oCore.BridgingSettings{
				MaxReceiversPerBridgingRequest: 3,
				MaxAmountAllowedToBridge:       common.LamportToWei(maxAmountAllowedToBridge),
			},
			ChainIDConverter: common.NewTestChainIDConverter(),
		}
		config.FillOut()

		return config
	}

	proc := NewSolanaBridgingRequestedProcessor(NewRefundDisabledProcessor(), hclog.NewNullLogger(), nil)

	t.Run("GetType", func(t *testing.T) {
		require.Equal(t, common.BridgingTxTypeBridgingRequest, proc.GetType())
	})

	t.Run("PreValidate", func(t *testing.T) {
		err := proc.PreValidate(&core.SolanaTx{}, nil)
		require.NoError(t, err)
	})

	t.Run("ValidateAndAddClaim empty tx", func(t *testing.T) {
		claims := &oCore.BridgeClaims{}
		appConfig := getAppConfig()

		err := proc.ValidateAndAddClaim(claims, &core.SolanaTx{}, appConfig)
		require.Error(t, err)
		require.ErrorContains(t, err, "failed to unmarshal sol metadata")
	})

	t.Run("ValidateAndAddClaim irrelevant metadata", func(t *testing.T) {
		irrelevantMetadata, err := core.MarshalSolMetadata(core.BaseSolMetadata{
			BridgingTxType: common.BridgingTxTypeBatchExecution,
		})
		require.NoError(t, err)

		claims := &oCore.BridgeClaims{}
		appConfig := getAppConfig()

		err = proc.ValidateAndAddClaim(claims, &core.SolanaTx{
			Metadata: irrelevantMetadata,
		}, appConfig)
		require.Error(t, err)
		require.ErrorContains(t, err, "irrelevant tx")
	})

	t.Run("ValidateAndAddClaim origin chain not registered", func(t *testing.T) {
		metadata, err := core.MarshalSolMetadata(core.BridgingRequestSolMetadata{
			BridgingTxType:     common.BridgingTxTypeBridgingRequest,
			DestinationChainID: common.ChainIDStrPrime,
			SenderAddr:         "addr1",
			Transactions:       []core.BridgingRequestSolMetadataTransaction{},
			OperationFee:       minOperationFee,
		})
		require.NoError(t, err)

		claims := &oCore.BridgeClaims{}
		appConfig := getAppConfig()

		err = proc.ValidateAndAddClaim(claims, &core.SolanaTx{
			Metadata:      metadata,
			OriginChainID: "unregistered",
		}, appConfig)
		require.Error(t, err)
		require.ErrorContains(t, err, "unsupported chain id found in tx")
	})

	t.Run("ValidateAndAddClaim destination chain not registered", func(t *testing.T) {
		metadata, err := core.MarshalSolMetadata(core.BridgingRequestSolMetadata{
			BridgingTxType:     common.BridgingTxTypeBridgingRequest,
			DestinationChainID: "unregistered",
			SenderAddr:         "addr1",
			Transactions:       []core.BridgingRequestSolMetadataTransaction{},
			OperationFee:       minOperationFee,
		})
		require.NoError(t, err)

		claims := &oCore.BridgeClaims{}
		appConfig := getAppConfig()

		err = proc.ValidateAndAddClaim(claims, &core.SolanaTx{
			Metadata:      metadata,
			OriginChainID: common.ChainIDStrSolana,
		}, appConfig)
		require.Error(t, err)
		require.ErrorContains(t, err, "unsupported chain id found in tx")
	})

	t.Run("ValidateAndAddClaim operation fee below minimum", func(t *testing.T) {
		metadata, err := core.MarshalSolMetadata(core.BridgingRequestSolMetadata{
			BridgingTxType:     common.BridgingTxTypeBridgingRequest,
			DestinationChainID: common.ChainIDStrPrime,
			SenderAddr:         "addr1",
			Transactions: []core.BridgingRequestSolMetadataTransaction{
				{Address: validCardanoTestAddr, Amount: common.LamportToWei(new(big.Int).SetUint64(utxoMinValue)), TokenID: solanaCurrencyID},
			},
			OperationFee: new(big.Int).Sub(minOperationFee, big.NewInt(1)),
			BridgingFee:  minFeeForBridging,
		})
		require.NoError(t, err)

		claims := &oCore.BridgeClaims{}
		appConfig := getAppConfig()

		err = proc.ValidateAndAddClaim(claims, &core.SolanaTx{
			Metadata:      metadata,
			OriginChainID: common.ChainIDStrSolana,
		}, appConfig)
		require.Error(t, err)
		require.ErrorContains(t, err, "operation fee in metadata is less than minimum")
	})

	t.Run("ValidateAndAddClaim more than max receivers", func(t *testing.T) {
		metadata, err := core.MarshalSolMetadata(core.BridgingRequestSolMetadata{
			BridgingTxType:     common.BridgingTxTypeBridgingRequest,
			DestinationChainID: common.ChainIDStrPrime,
			SenderAddr:         "addr1",
			Transactions: []core.BridgingRequestSolMetadataTransaction{
				{Address: validCardanoTestAddr, Amount: big.NewInt(1), TokenID: solanaCurrencyID},
				{Address: validCardanoTestAddr, Amount: big.NewInt(1), TokenID: solanaCurrencyID},
				{Address: validCardanoTestAddr, Amount: big.NewInt(1), TokenID: solanaCurrencyID},
				{Address: validCardanoTestAddr, Amount: big.NewInt(1), TokenID: solanaCurrencyID},
			},
			OperationFee: minOperationFee,
			BridgingFee:  minFeeForBridging,
		})
		require.NoError(t, err)

		claims := &oCore.BridgeClaims{}
		appConfig := getAppConfig()

		err = proc.ValidateAndAddClaim(claims, &core.SolanaTx{
			Metadata:      metadata,
			OriginChainID: common.ChainIDStrSolana,
		}, appConfig)
		require.Error(t, err)
		require.ErrorContains(t, err, "number of receivers in metadata greater than maximum allowed")
	})

	t.Run("ValidateAndAddClaim transaction direction not allowed", func(t *testing.T) {
		metadata, err := core.MarshalSolMetadata(core.BridgingRequestSolMetadata{
			BridgingTxType:     common.BridgingTxTypeBridgingRequest,
			DestinationChainID: common.ChainIDStrPrime,
			SenderAddr:         "addr1",
			Transactions: []core.BridgingRequestSolMetadataTransaction{
				{Address: validCardanoTestAddr, Amount: common.LamportToWei(new(big.Int).SetUint64(utxoMinValue)), TokenID: 99},
			},
			OperationFee: minOperationFee,
			BridgingFee:  minFeeForBridging,
		})
		require.NoError(t, err)

		claims := &oCore.BridgeClaims{}
		appConfig := getAppConfig()

		err = proc.ValidateAndAddClaim(claims, &core.SolanaTx{
			Metadata:      metadata,
			OriginChainID: common.ChainIDStrSolana,
		}, appConfig)
		require.Error(t, err)
		require.ErrorContains(t, err, "invalid receiver")
	})

	t.Run("ValidateAndAddClaim invalid cardano receiver address", func(t *testing.T) {
		metadata, err := core.MarshalSolMetadata(core.BridgingRequestSolMetadata{
			BridgingTxType:     common.BridgingTxTypeBridgingRequest,
			DestinationChainID: common.ChainIDStrPrime,
			SenderAddr:         "addr1",
			Transactions: []core.BridgingRequestSolMetadataTransaction{
				{Address: "invalid_addr", Amount: common.LamportToWei(new(big.Int).SetUint64(utxoMinValue)), TokenID: solanaCurrencyID},
			},
			OperationFee: minOperationFee,
			BridgingFee:  minFeeForBridging,
		})
		require.NoError(t, err)

		claims := &oCore.BridgeClaims{}
		appConfig := getAppConfig()

		err = proc.ValidateAndAddClaim(claims, &core.SolanaTx{
			Metadata:      metadata,
			OriginChainID: common.ChainIDStrSolana,
		}, appConfig)
		require.Error(t, err)
		require.ErrorContains(t, err, "found an invalid receiver addr in metadata")
	})

	t.Run("ValidateAndAddClaim cardano utxo below minimum", func(t *testing.T) {
		metadata, err := core.MarshalSolMetadata(core.BridgingRequestSolMetadata{
			BridgingTxType:     common.BridgingTxTypeBridgingRequest,
			DestinationChainID: common.ChainIDStrPrime,
			SenderAddr:         "addr1",
			Transactions: []core.BridgingRequestSolMetadataTransaction{
				{Address: validCardanoTestAddr, Amount: big.NewInt(1), TokenID: solanaCurrencyID},
			},
			OperationFee: minOperationFee,
			BridgingFee:  minFeeForBridging,
		})
		require.NoError(t, err)

		claims := &oCore.BridgeClaims{}
		appConfig := getAppConfig()

		err = proc.ValidateAndAddClaim(claims, &core.SolanaTx{
			Metadata:      metadata,
			OriginChainID: common.ChainIDStrSolana,
		}, appConfig)
		require.Error(t, err)
		require.ErrorContains(t, err, "found an utxo value below minimum value in metadata receivers")
	})

	t.Run("ValidateAndAddClaim invalid eth receiver address", func(t *testing.T) {
		metadata, err := core.MarshalSolMetadata(core.BridgingRequestSolMetadata{
			BridgingTxType:     common.BridgingTxTypeBridgingRequest,
			DestinationChainID: common.ChainIDStrNexus,
			SenderAddr:         "addr1",
			Transactions: []core.BridgingRequestSolMetadataTransaction{
				{Address: "not_hex_addr", Amount: common.LamportToWei(new(big.Int).SetUint64(utxoMinValue)), TokenID: solanaCurrencyID},
			},
			OperationFee: minOperationFee,
			BridgingFee:  minFeeForBridging,
		})
		require.NoError(t, err)

		claims := &oCore.BridgeClaims{}
		appConfig := getAppConfig()

		err = proc.ValidateAndAddClaim(claims, &core.SolanaTx{
			Metadata:      metadata,
			OriginChainID: common.ChainIDStrSolana,
		}, appConfig)
		require.Error(t, err)
		require.ErrorContains(t, err, "found an invalid eth receiver addr in metadata")
	})

	t.Run("ValidateAndAddClaim eth receiver amount below minimum", func(t *testing.T) {
		metadata, err := core.MarshalSolMetadata(core.BridgingRequestSolMetadata{
			BridgingTxType:     common.BridgingTxTypeBridgingRequest,
			DestinationChainID: common.ChainIDStrNexus,
			SenderAddr:         "addr1",
			Transactions: []core.BridgingRequestSolMetadataTransaction{
				{Address: validEthTestAddr, Amount: big.NewInt(1), TokenID: solanaCurrencyID},
			},
			OperationFee: minOperationFee,
			BridgingFee:  minFeeForBridging,
		})
		require.NoError(t, err)

		claims := &oCore.BridgeClaims{}
		appConfig := getAppConfig()

		err = proc.ValidateAndAddClaim(claims, &core.SolanaTx{
			Metadata:      metadata,
			OriginChainID: common.ChainIDStrSolana,
		}, appConfig)
		require.Error(t, err)
		require.ErrorContains(t, err, "token amount below minimum allowed")
	})

	t.Run("ValidateAndAddClaim bridging fee below minimum", func(t *testing.T) {
		cardanoMinWei := common.DfmToWei(new(big.Int).SetUint64(utxoMinValue))
		bridgingFeeWei := new(big.Int).Sub(minFeeForBridging, big.NewInt(1))
		txValue := new(big.Int).Add(cardanoMinWei, bridgingFeeWei)

		metadata, err := core.MarshalSolMetadata(core.BridgingRequestSolMetadata{
			BridgingTxType:     common.BridgingTxTypeBridgingRequest,
			DestinationChainID: common.ChainIDStrPrime,
			SenderAddr:         "addr1",
			Transactions: []core.BridgingRequestSolMetadataTransaction{
				{Address: validCardanoTestAddr, Amount: cardanoMinWei, TokenID: solanaCurrencyID},
			},
			OperationFee: minOperationFee,
			BridgingFee:  bridgingFeeWei,
		})
		require.NoError(t, err)

		claims := &oCore.BridgeClaims{}
		appConfig := getAppConfig()

		err = proc.ValidateAndAddClaim(claims, &core.SolanaTx{
			Metadata:      metadata,
			OriginChainID: common.ChainIDStrSolana,
			Value:         txValue,
		}, appConfig)
		require.Error(t, err)
		require.ErrorContains(t, err, "bridging fee in metadata is less than minimum")
	})

	t.Run("ValidateAndAddClaim operation fee below minimum", func(t *testing.T) {
		cardanoMinWei := common.DfmToWei(new(big.Int).SetUint64(utxoMinValue))
		metadata, err := core.MarshalSolMetadata(core.BridgingRequestSolMetadata{
			BridgingTxType:     common.BridgingTxTypeBridgingRequest,
			DestinationChainID: common.ChainIDStrPrime,
			SenderAddr:         "addr1",
			Transactions: []core.BridgingRequestSolMetadataTransaction{
				{Address: validCardanoTestAddr, Amount: cardanoMinWei, TokenID: solanaCurrencyID},
			},
			OperationFee: new(big.Int).Sub(minOperationFee, big.NewInt(1)),
			BridgingFee:  minFeeForBridging,
		})
		require.NoError(t, err)

		claims := &oCore.BridgeClaims{}
		appConfig := getAppConfig()

		err = proc.ValidateAndAddClaim(claims, &core.SolanaTx{
			Metadata:      metadata,
			OriginChainID: common.ChainIDStrSolana,
			Value:         cardanoMinWei,
		}, appConfig)
		require.Error(t, err)
		require.ErrorContains(t, err, "operation fee in metadata is less than minimum")
	})

	t.Run("ValidateAndAddClaim amount above max", func(t *testing.T) {
		bigAmountWei := new(big.Int).Add(common.LamportToWei(maxAmountAllowedToBridge), big.NewInt(1))
		txValue := new(big.Int).Add(bigAmountWei, minFeeForBridging)

		metadata, err := core.MarshalSolMetadata(core.BridgingRequestSolMetadata{
			BridgingTxType:     common.BridgingTxTypeBridgingRequest,
			DestinationChainID: common.ChainIDStrPrime,
			SenderAddr:         "addr1",
			Transactions: []core.BridgingRequestSolMetadataTransaction{
				{Address: validCardanoTestAddr, Amount: bigAmountWei, TokenID: solanaCurrencyID},
			},
			OperationFee: minOperationFee,
			BridgingFee:  minFeeForBridging,
		})
		require.NoError(t, err)

		claims := &oCore.BridgeClaims{}
		appConfig := getAppConfig()

		err = proc.ValidateAndAddClaim(claims, &core.SolanaTx{
			Metadata:      metadata,
			OriginChainID: common.ChainIDStrSolana,
			Value:         txValue,
		}, appConfig)
		require.Error(t, err)
		require.ErrorContains(t, err, "greater than maximum allowed")
	})

	t.Run("ValidateAndAddClaim valid to Cardano", func(t *testing.T) {
		amount := common.DfmToWei(new(big.Int).SetUint64(utxoMinValue))
		txValue := new(big.Int).Add(amount, minFeeForBridging)

		var txSig solana.Signature

		copy(txSig[:], []byte("test_tx_signature_for_bridging_req_00001"))

		metadata, err := core.MarshalSolMetadata(core.BridgingRequestSolMetadata{
			BridgingTxType:     common.BridgingTxTypeBridgingRequest,
			DestinationChainID: common.ChainIDStrPrime,
			SenderAddr:         "addr1",
			Transactions: []core.BridgingRequestSolMetadataTransaction{
				{Address: validCardanoTestAddr, Amount: amount, TokenID: solanaCurrencyID},
			},
			OperationFee: minOperationFee,
			BridgingFee:  minFeeForBridging,
		})
		require.NoError(t, err)

		claims := &oCore.BridgeClaims{}
		appConfig := getAppConfig()

		err = proc.ValidateAndAddClaim(claims, &core.SolanaTx{
			Metadata:      metadata,
			OriginChainID: common.ChainIDStrSolana,
			TxSignature:   txSig,
			Value:         txValue,
		}, appConfig)
		require.NoError(t, err)
		require.Equal(t, 1, claims.Count())
		require.Len(t, claims.BridgingRequestClaims, 1)
		require.Equal(t, txSig[:], claims.BridgingRequestClaims[0].ObservedTransactionHash)
		require.Equal(t,
			common.ChainIDStrPrime,
			appConfig.ChainIDConverter.ToChainIDStr(claims.BridgingRequestClaims[0].DestinationChainId))
		require.Len(t, claims.BridgingRequestClaims[0].Receivers, 2)
		require.Equal(t, primeFeeAddr, claims.BridgingRequestClaims[0].Receivers[1].DestinationAddress)
	})

	t.Run("ValidateAndAddClaim valid to ETH", func(t *testing.T) {
		amount := common.LamportToWei(new(big.Int).SetUint64(utxoMinValue))
		txValue := new(big.Int).Add(amount, minFeeForBridging)

		var txSig solana.Signature

		copy(txSig[:], []byte("test_tx_sig_eth_dest_000002"))

		metadata, err := core.MarshalSolMetadata(core.BridgingRequestSolMetadata{
			BridgingTxType:     common.BridgingTxTypeBridgingRequest,
			DestinationChainID: common.ChainIDStrNexus,
			SenderAddr:         "addr1",
			Transactions: []core.BridgingRequestSolMetadataTransaction{
				{Address: validEthTestAddr, Amount: amount, TokenID: solanaCurrencyID},
			},
			OperationFee: minOperationFee,
			BridgingFee:  minFeeForBridging,
		})
		require.NoError(t, err)

		claims := &oCore.BridgeClaims{}
		appConfig := getAppConfig()

		err = proc.ValidateAndAddClaim(claims, &core.SolanaTx{
			Metadata:      metadata,
			OriginChainID: common.ChainIDStrSolana,
			TxSignature:   txSig,
			Value:         txValue,
		}, appConfig)
		require.NoError(t, err)
		require.Equal(t, 1, claims.Count())
		require.Len(t, claims.BridgingRequestClaims, 1)
		require.Equal(t, txSig[:], claims.BridgingRequestClaims[0].ObservedTransactionHash)
		require.Equal(t,
			common.ChainIDStrNexus,
			appConfig.ChainIDConverter.ToChainIDStr(claims.BridgingRequestClaims[0].DestinationChainId))
	})

	t.Run("ValidateAndAddClaim multiple receivers valid", func(t *testing.T) {
		amount := common.DfmToWei(new(big.Int).SetUint64(utxoMinValue))

		var txSig solana.Signature

		copy(txSig[:], []byte("test_tx_sig_multi_recv_000003"))

		metadata, err := core.MarshalSolMetadata(core.BridgingRequestSolMetadata{
			BridgingTxType:     common.BridgingTxTypeBridgingRequest,
			DestinationChainID: common.ChainIDStrPrime,
			SenderAddr:         "addr1",
			Transactions: []core.BridgingRequestSolMetadataTransaction{
				{Address: validCardanoTestAddr, Amount: amount, TokenID: solanaCurrencyID},
				{Address: validCardanoTestAddr, Amount: amount, TokenID: solanaCurrencyID},
			},
			OperationFee: minOperationFee,
			BridgingFee:  minFeeForBridging,
		})
		require.NoError(t, err)

		totalValue := new(big.Int).Mul(amount, big.NewInt(2))
		totalValue.Add(totalValue, minFeeForBridging)

		claims := &oCore.BridgeClaims{}
		appConfig := getAppConfig()

		err = proc.ValidateAndAddClaim(claims, &core.SolanaTx{
			Metadata:      metadata,
			OriginChainID: common.ChainIDStrSolana,
			TxSignature:   txSig,
			Value:         totalValue,
		}, appConfig)
		require.NoError(t, err)
		require.Equal(t, 1, claims.Count())
		require.Len(t, claims.BridgingRequestClaims[0].Receivers, 3)
		require.Equal(t, primeFeeAddr, claims.BridgingRequestClaims[0].Receivers[2].DestinationAddress)
	})
}
