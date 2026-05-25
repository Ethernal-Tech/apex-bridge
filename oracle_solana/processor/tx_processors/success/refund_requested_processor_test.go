package successtxprocessors

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/Ethernal-Tech/apex-bridge/common"
	oCore "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	solCore "github.com/Ethernal-Tech/apex-bridge/oracle_solana/core"
	solanatx "github.com/Ethernal-Tech/apex-bridge/solana"
	"github.com/Ethernal-Tech/cardano-infrastructure/wallet"
	"github.com/gagliardetto/solana-go"
	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/require"
)

func testRefundAppConfig() *oCore.AppConfig {
	const (
		solanaCurrencyID    = uint16(1)
		solanaWrappedID     = uint16(2)
		solanaRegularToken  = uint16(3)
		solanaUntrackedID   = uint16(4)
		primeCurrencyID     = uint16(11)
		primeWrappedID      = uint16(12)
		primeRegularTokenID = uint16(13)
	)

	cfg := &oCore.AppConfig{
		SolanaChains: map[string]*oCore.SolanaChainConfig{
			common.ChainIDStrSolana: {
				SolanaChainConfig: solanatx.SolanaChainConfig{
					DestinationChains: map[string]common.TokenPairs{
						common.ChainIDStrPrime: {
							{SourceTokenID: solanaCurrencyID, DestinationTokenID: primeCurrencyID, TrackSourceToken: true, TrackDestinationToken: true},
							{SourceTokenID: solanaWrappedID, DestinationTokenID: primeWrappedID, TrackSourceToken: true, TrackDestinationToken: true},
							{SourceTokenID: solanaRegularToken, DestinationTokenID: primeRegularTokenID, TrackSourceToken: false, TrackDestinationToken: true},
							{SourceTokenID: solanaUntrackedID, DestinationTokenID: primeRegularTokenID, TrackSourceToken: false, TrackDestinationToken: false},
						},
					},
					Tokens: map[uint16]common.Token{
						solanaCurrencyID:   {ChainSpecific: wallet.AdaTokenName, LockUnlock: true},
						solanaWrappedID:    {ChainSpecific: "wrapped-sol", LockUnlock: true, IsWrappedCurrency: true},
						solanaRegularToken: {ChainSpecific: "token-x", LockUnlock: false, IsWrappedCurrency: false},
						solanaUntrackedID:  {ChainSpecific: "token-y", LockUnlock: false, IsWrappedCurrency: false},
					},
					MinFeeForBridging: common.LamportToWei(big.NewInt(10)),
				},
			},
		},
		BridgingSettings: oCore.BridgingSettings{
			MaxReceiversPerBridgingRequest: 5,
			MaxAmountAllowedToBridge:       common.LamportToWei(big.NewInt(1_000_000_000)),
		},
		TryCountLimits: oCore.TryCountLimits{
			MaxBatchTryCount:  1,
			MaxSubmitTryCount: 1,
			MaxRefundTryCount: 3,
		},
		RefundEnabled:    true,
		ChainIDConverter: common.NewTestChainIDConverter(),
	}

	cfg.FillOut()

	return cfg
}

func TestRefundRequestProcessorSkyline(t *testing.T) {
	proc := NewRefundRequestProcessorSkyline(hclog.NewNullLogger())
	disabledProc := NewRefundDisabledProcessor()

	t.Run("GetType", func(t *testing.T) {
		require.Equal(t, common.TxTypeRefundRequest, proc.GetType())
	})

	t.Run("PreValidate", func(t *testing.T) {
		require.NoError(t, proc.PreValidate(&solCore.SolanaTx{}, &oCore.AppConfig{}))
	})

	t.Run("Refund disabled processor paths", func(t *testing.T) {
		cfg := testRefundAppConfig()

		require.Equal(t, common.TxTypeRefundRequest, disabledProc.GetType())
		require.NoError(t, disabledProc.PreValidate(&solCore.SolanaTx{}, cfg))

		require.NoError(t, disabledProc.HandleBridgingProcessorPreValidate(&solCore.SolanaTx{}, cfg))

		err := disabledProc.HandleBridgingProcessorError(
			&oCore.BridgeClaims{},
			&solCore.SolanaTx{},
			cfg,
			fmt.Errorf("test err"),
			"context msg",
		)
		require.ErrorContains(t, err, "test err")
		require.ErrorContains(t, err, "context msg")

		err = disabledProc.ValidateAndAddClaim(&oCore.BridgeClaims{}, &solCore.SolanaTx{}, cfg)
		require.ErrorContains(t, err, "refund is not enabled")
	})

	t.Run("HandleBridgingProcessorPreValidate try count exceeded", func(t *testing.T) {
		cfg := testRefundAppConfig()
		require.NoError(t, proc.HandleBridgingProcessorPreValidate(&solCore.SolanaTx{}, cfg))

		err := proc.HandleBridgingProcessorPreValidate(&solCore.SolanaTx{
			BatchTryCount: 2,
		}, cfg)
		require.ErrorContains(t, err, "try count exceeded")

		err = proc.HandleBridgingProcessorPreValidate(&solCore.SolanaTx{
			SubmitTryCount: 2,
		}, cfg)
		require.ErrorContains(t, err, "try count exceeded")
	})

	t.Run("HandleBridgingProcessorError passes through ValidateAndAddClaim", func(t *testing.T) {
		cfg := testRefundAppConfig()
		claims := &oCore.BridgeClaims{}

		err := proc.HandleBridgingProcessorError(claims, &solCore.SolanaTx{}, cfg, nil, "ctx")
		require.ErrorContains(t, err, "failed to unmarshal metadata")
		require.Equal(t, 0, claims.Count())
	})

	t.Run("ValidateAndAddClaim empty tx", func(t *testing.T) {
		cfg := testRefundAppConfig()

		err := proc.ValidateAndAddClaim(&oCore.BridgeClaims{}, &solCore.SolanaTx{}, cfg)
		require.ErrorContains(t, err, "failed to unmarshal metadata")
	})

	t.Run("ValidateAndAddClaim origin chain not registered", func(t *testing.T) {
		cfg := testRefundAppConfig()
		metadata, err := solCore.MarshalSolMetadata(solCore.RefundBridgingRequestSolMetadata{
			BridgingTxType:     common.TxTypeRefundRequest,
			DestinationChainID: common.ChainIDStrPrime,
			SenderAddr:         solana.NewWallet().PublicKey().String(),
		})
		require.NoError(t, err)

		err = proc.ValidateAndAddClaim(&oCore.BridgeClaims{}, &solCore.SolanaTx{
			OriginChainID: "invalid",
			Metadata:      metadata,
		}, cfg)
		require.ErrorContains(t, err, "unsupported chain id found in tx")
	})

	t.Run("ValidateAndAddClaim invalid sender", func(t *testing.T) {
		cfg := testRefundAppConfig()
		metadata, err := solCore.MarshalSolMetadata(solCore.RefundBridgingRequestSolMetadata{
			BridgingTxType:     common.TxTypeRefundRequest,
			DestinationChainID: common.ChainIDStrPrime,
			SenderAddr:         "invalid_sender",
		})
		require.NoError(t, err)

		err = proc.ValidateAndAddClaim(&oCore.BridgeClaims{}, &solCore.SolanaTx{
			OriginChainID: common.ChainIDStrSolana,
			Metadata:      metadata,
		}, cfg)
		require.ErrorContains(t, err, "invalid sender addr")
	})

	t.Run("ValidateAndAddClaim tx value below min fee", func(t *testing.T) {
		cfg := testRefundAppConfig()
		metadata, err := solCore.MarshalSolMetadata(solCore.RefundBridgingRequestSolMetadata{
			BridgingTxType:     common.TxTypeRefundRequest,
			DestinationChainID: common.ChainIDStrPrime,
			SenderAddr:         solana.NewWallet().PublicKey().String(),
		})
		require.NoError(t, err)

		err = proc.ValidateAndAddClaim(&oCore.BridgeClaims{}, &solCore.SolanaTx{
			OriginChainID: common.ChainIDStrSolana,
			Metadata:      metadata,
			Value:         common.LamportToWei(big.NewInt(9)),
		}, cfg)
		require.ErrorContains(t, err, "less than the minimum required for refund")
	})

	t.Run("ValidateAndAddClaim refund try count exceeded", func(t *testing.T) {
		cfg := testRefundAppConfig()
		metadata, err := solCore.MarshalSolMetadata(solCore.RefundBridgingRequestSolMetadata{
			BridgingTxType:     common.TxTypeRefundRequest,
			DestinationChainID: common.ChainIDStrPrime,
			SenderAddr:         solana.NewWallet().PublicKey().String(),
		})
		require.NoError(t, err)

		err = proc.ValidateAndAddClaim(&oCore.BridgeClaims{}, &solCore.SolanaTx{
			OriginChainID:  common.ChainIDStrSolana,
			Metadata:       metadata,
			Value:          common.LamportToWei(big.NewInt(10)),
			RefundTryCount: 4,
		}, cfg)
		require.ErrorContains(t, err, "try count exceeded")
	})

	t.Run("ValidateAndAddClaim unregistered token in metadata", func(t *testing.T) {
		cfg := testRefundAppConfig()
		metadata, err := solCore.MarshalSolMetadata(solCore.RefundBridgingRequestSolMetadata{
			BridgingTxType:     common.TxTypeRefundRequest,
			DestinationChainID: common.ChainIDStrPrime,
			SenderAddr:         solana.NewWallet().PublicKey().String(),
			Transactions: []solCore.BridgingRequestSolMetadataTransaction{
				{TokenID: 999, Amount: common.LamportToWei(big.NewInt(1))},
			},
		})
		require.NoError(t, err)

		err = proc.ValidateAndAddClaim(&oCore.BridgeClaims{}, &solCore.SolanaTx{
			OriginChainID: common.ChainIDStrSolana,
			Metadata:      metadata,
			Value:         common.LamportToWei(big.NewInt(10)),
		}, cfg)
		require.ErrorContains(t, err, "token with ID 999 is not registered")
	})

	t.Run("ValidateAndAddClaim valid with mixed token types", func(t *testing.T) {
		cfg := testRefundAppConfig()
		sender := solana.NewWallet().PublicKey().String()
		txSig := solana.Signature{1, 2, 3}
		metadata, err := solCore.MarshalSolMetadata(solCore.RefundBridgingRequestSolMetadata{
			BridgingTxType:     common.TxTypeRefundRequest,
			DestinationChainID: common.ChainIDStrPrime,
			SenderAddr:         sender,
			Transactions: []solCore.BridgingRequestSolMetadataTransaction{
				{TokenID: 1, Amount: common.LamportToWei(big.NewInt(100))}, // currency tracked
				{TokenID: 2, Amount: common.LamportToWei(big.NewInt(50))},  // wrapped tracked
				{TokenID: 3, Amount: common.LamportToWei(big.NewInt(25))},  // non-wrapped token
				{TokenID: 4, Amount: common.LamportToWei(big.NewInt(10))},  // untracked pair
			},
		})
		require.NoError(t, err)

		claims := &oCore.BridgeClaims{}
		err = proc.ValidateAndAddClaim(claims, &solCore.SolanaTx{
			OriginChainID:  common.ChainIDStrSolana,
			Metadata:       metadata,
			TxSignature:    txSig,
			BatchTryCount:  1,
			RefundTryCount: 2,
			Value:          common.LamportToWei(big.NewInt(999)),
		}, cfg)
		require.NoError(t, err)

		require.Equal(t, 1, claims.Count())
		require.Len(t, claims.RefundRequestClaims, 1)
		claim := claims.RefundRequestClaims[0]
		require.Equal(t, cfg.ChainIDConverter.ToChainIDNum(common.ChainIDStrSolana), claim.OriginChainId)
		require.Equal(t, cfg.ChainIDConverter.ToChainIDNum(common.ChainIDStrPrime), claim.DestinationChainId)
		require.Equal(t, sender, claim.OriginSenderAddress)
		require.Equal(t, txSig[:], claim.OriginTransactionHash)
		require.Equal(t, common.LamportToWei(big.NewInt(100)), claim.OriginAmount)
		require.Equal(t, common.LamportToWei(big.NewInt(50)), claim.OriginWrappedAmount)
		require.True(t, claim.ShouldDecrementHotWallet)
		require.Equal(t, uint64(2), claim.RetryCounter)
		require.Len(t, claim.TokenAmounts, 4)
		require.Equal(t, uint16(1), claim.TokenAmounts[0].TokenId)
		require.Equal(t, common.LamportToWei(big.NewInt(999)), claim.TokenAmounts[0].AmountCurrency)
		require.Equal(t, uint64(0), claim.TokenAmounts[0].AmountTokens.Uint64())
		require.Equal(t, uint16(2), claim.TokenAmounts[1].TokenId)
		require.Equal(t, common.LamportToWei(big.NewInt(50)), claim.TokenAmounts[1].AmountTokens)
	})

	t.Run("ValidateAndAddClaim unsupported destination chain gives zero destination id", func(t *testing.T) {
		cfg := testRefundAppConfig()
		metadata, err := solCore.MarshalSolMetadata(solCore.RefundBridgingRequestSolMetadata{
			BridgingTxType:     common.TxTypeRefundRequest,
			DestinationChainID: "unknown",
			SenderAddr:         solana.NewWallet().PublicKey().String(),
			Transactions: []solCore.BridgingRequestSolMetadataTransaction{
				{TokenID: 1, Amount: common.LamportToWei(big.NewInt(20))},
			},
		})
		require.NoError(t, err)

		claims := &oCore.BridgeClaims{}
		err = proc.ValidateAndAddClaim(claims, &solCore.SolanaTx{
			OriginChainID: common.ChainIDStrSolana,
			Metadata:      metadata,
			Value:         common.LamportToWei(big.NewInt(20)),
		}, cfg)
		require.NoError(t, err)
		require.Len(t, claims.RefundRequestClaims, 1)
		require.Equal(t, uint8(0), claims.RefundRequestClaims[0].DestinationChainId)
		require.Equal(t, uint64(0), claims.RefundRequestClaims[0].OriginAmount.Uint64())
		require.Equal(t, uint64(0), claims.RefundRequestClaims[0].OriginWrappedAmount.Uint64())
	})

	t.Run("buildRefundTokenAmounts track flags", func(t *testing.T) {
		cfg := testRefundAppConfig()
		chainCfg := cfg.SolanaChains[common.ChainIDStrSolana]
		currencyID, err := chainCfg.GetCurrencyID()
		require.NoError(t, err)

		txValue := common.LamportToWei(big.NewInt(200))
		metadata := &solCore.RefundBridgingRequestSolMetadata{
			DestinationChainID: common.ChainIDStrPrime,
			Transactions: []solCore.BridgingRequestSolMetadataTransaction{
				{TokenID: 1, Amount: common.LamportToWei(big.NewInt(30))},
				{TokenID: 2, Amount: common.LamportToWei(big.NewInt(40))},
				{TokenID: 4, Amount: common.LamportToWei(big.NewInt(50))},
			},
		}

		tokenAmounts, totalCurrency, totalWrapped := buildRefundTokenAmounts(chainCfg, txValue, metadata, currencyID)
		require.Len(t, tokenAmounts, 3)
		require.Equal(t, common.LamportToWei(big.NewInt(30)), totalCurrency)
		require.Equal(t, common.LamportToWei(big.NewInt(40)), totalWrapped)
		require.Equal(t, txValue, tokenAmounts[0].AmountCurrency)
		require.Equal(t, uint64(0), tokenAmounts[1].AmountCurrency.Uint64())
	})
}
