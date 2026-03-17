package relayer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Ethernal-Tech/apex-bridge/eth"
	"github.com/Ethernal-Tech/apex-bridge/relayer/core"
	solanatx "github.com/Ethernal-Tech/apex-bridge/solana"
	"github.com/Ethernal-Tech/cardano-infrastructure/secrets"
	secretsHelper "github.com/Ethernal-Tech/cardano-infrastructure/secrets/helper"
	"github.com/Ethernal-Tech/solana-infrastructure/sendtx"
	solanaWallet "github.com/Ethernal-Tech/solana-infrastructure/wallet"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockTxSubmiter struct {
	mock.Mock
}

var _ solanaWallet.ITxSubmiter = (*mockTxSubmiter)(nil)

func (m *mockTxSubmiter) SendTransaction(ctx context.Context, tx *solana.Transaction) (solana.Signature, error) {
	args := m.Called(ctx, tx)

	return args.Get(0).(solana.Signature), args.Error(1)
}

func (m *mockTxSubmiter) WaitForSignature(
	ctx context.Context, sig solana.Signature, commitment rpc.CommitmentType, maxWaitTime time.Duration,
) error {
	return m.Called(ctx, sig, commitment, maxWaitTime).Error(0)
}

func buildTestRawTx(t *testing.T, signerPubKey solana.PublicKey) []byte {
	t.Helper()

	txProvider, err := solanaWallet.NewProvider(rpc.LocalNet_WS)
	require.NoError(t, err)

	receiverWallet, err := solana.NewRandomPrivateKey()
	require.NoError(t, err)

	blockHash, err := solana.NewRandomPrivateKey()
	require.NoError(t, err)

	txSender := sendtx.NewTxSender(txProvider, nil)

	bridgingTxDto := sendtx.BridgeTransactionDto{
		SenderAddr: signerPubKey.String(),
		Receivers: []sendtx.BridgingTxReceiver{
			{
				Address: receiverWallet.PublicKey().String(),
				TokenAmount: solanaWallet.TokenAmount{
					Amount:    new(big.Int).SetUint64(1_000_000),
					TokenMint: solana.NewWallet().PublicKey().String(),
				},
			},
		},
		BatchID: 1,
	}

	tx, err := txSender.CreateTx(
		context.Background(),
		signerPubKey,
		sendtx.InstructionTypeBridgeTransaction,
		solana.Hash(blockHash.PublicKey()),
		bridgingTxDto,
	)
	require.NoError(t, err)

	raw, err := solanaWallet.MarshalTransaction(tx)
	require.NoError(t, err)

	return raw
}

func TestNewSolanaChainOperations(t *testing.T) {
	const chainID = "solana"

	testDir, err := os.MkdirTemp("", "relayer-solana")
	require.NoError(t, err)

	defer os.RemoveAll(testDir)

	secretsDir := filepath.Join(testDir, "stp")

	secretsMngr, err := secretsHelper.CreateSecretsManager(&secrets.SecretsManagerConfig{
		Path: secretsDir,
		Type: secrets.Local,
	})
	require.NoError(t, err)

	t.Run("invalid chain config JSON", func(t *testing.T) {
		chainConfig := core.ChainConfig{
			ChainID:       chainID,
			ChainSpecific: json.RawMessage([]byte(`{invalid`)),
		}

		ops, err := NewSolanaChainOperations(chainConfig, hclog.NewNullLogger())
		require.Error(t, err)
		require.Nil(t, ops)
	})

	t.Run("missing secrets manager dir", func(t *testing.T) {
		chainConfig := core.ChainConfig{
			ChainID:        chainID,
			ChainSpecific:  json.RawMessage([]byte(`{"txProviderEndpoint":"ws://localhost:8900"}`)),
			RelayerDataDir: "/nonexistent/path/that/does/not/exist",
		}

		ops, err := NewSolanaChainOperations(chainConfig, hclog.NewNullLogger())
		require.Error(t, err)
		require.Nil(t, ops)
	})

	t.Run("missing private key", func(t *testing.T) {
		emptySecretsDir := filepath.Join(testDir, "empty-secrets")
		require.NoError(t, os.MkdirAll(emptySecretsDir, 0700))

		chainConfig := core.ChainConfig{
			ChainID:        chainID,
			ChainSpecific:  json.RawMessage([]byte(`{"txProviderEndpoint":"ws://localhost:8900"}`)),
			RelayerDataDir: emptySecretsDir,
		}

		ops, err := NewSolanaChainOperations(chainConfig, hclog.NewNullLogger())
		require.Error(t, err)
		require.Nil(t, ops)
		require.Contains(t, err.Error(), "failed to load relayer solana private key")
	})

	t.Run("success", func(t *testing.T) {
		_, err := solanatx.GenerateAndStoreRelayerSolanaPrivateKey(secretsMngr, chainID, true)
		require.NoError(t, err)

		txProviderEndpoint := rpc.LocalNet_WS

		configRaw := json.RawMessage([]byte(fmt.Sprintf(`{
			"txProviderEndpoint": "%s"
		}`, txProviderEndpoint)))

		chainConfig := core.ChainConfig{
			ChainID:        chainID,
			ChainSpecific:  configRaw,
			RelayerDataDir: secretsDir,
		}

		ops, err := NewSolanaChainOperations(chainConfig, hclog.NewNullLogger())
		require.NoError(t, err)
		require.NotNil(t, ops)
		require.Equal(t, chainID, ops.chainID)
		require.NotNil(t, ops.privateKey)
		require.NotNil(t, ops.txSender)
		require.Equal(t, txProviderEndpoint, ops.config.TxProviderEndpoint)
	})
}

func TestSolanaChainOperations_SendTx(t *testing.T) {
	ctx := context.Background()

	privateKey, err := solana.NewRandomPrivateKey()
	require.NoError(t, err)

	t.Run("invalid signature (all zeros)", func(t *testing.T) {
		submiterMock := &mockTxSubmiter{}

		ops := &SolanaChainOperations{
			privateKey: &privateKey,
			txSender:   sendtx.NewTxSender(submiterMock, nil),
			logger:     hclog.NewNullLogger(),
		}

		batch := &eth.ConfirmedBatch{
			RawTransaction: buildTestRawTx(t, privateKey.PublicKey()),
			Bitmap:         new(big.Int),
			Signatures:     [][]byte{make([]byte, 64)},
		}

		err := ops.SendTx(ctx, nil, batch)
		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid signature")
	})

	t.Run("invalid RawTransaction", func(t *testing.T) {
		submiterMock := &mockTxSubmiter{}

		ops := &SolanaChainOperations{
			privateKey: &privateKey,
			txSender:   sendtx.NewTxSender(submiterMock, nil),
			logger:     hclog.NewNullLogger(),
		}

		validSig := make([]byte, 64)
		validSig[0] = 1

		batch := &eth.ConfirmedBatch{
			RawTransaction: []byte("garbage"),
			Bitmap:         new(big.Int),
			Signatures:     [][]byte{validSig},
		}

		err := ops.SendTx(ctx, nil, batch)
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to unmarshal tx")
	})

	t.Run("send returns error", func(t *testing.T) {
		submiterMock := &mockTxSubmiter{}

		ops := &SolanaChainOperations{
			privateKey: &privateKey,
			txSender:   sendtx.NewTxSender(submiterMock, nil),
			logger:     hclog.NewNullLogger(),
		}

		rawTx := buildTestRawTx(t, privateKey.PublicKey())

		validSig := make([]byte, 64)
		validSig[0] = 1

		batch := &eth.ConfirmedBatch{
			RawTransaction: rawTx,
			Bitmap:         new(big.Int),
			Signatures:     [][]byte{validSig},
		}

		submiterMock.On("SendTransaction", ctx, mock.AnythingOfType("*solana.Transaction")).
			Return(solana.Signature{}, errors.New("rpc unavailable")).Once()

		err := ops.SendTx(ctx, nil, batch)
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to send tx")

		submiterMock.AssertExpectations(t)
	})

	t.Run("success", func(t *testing.T) {
		submiterMock := &mockTxSubmiter{}

		ops := &SolanaChainOperations{
			privateKey: &privateKey,
			txSender:   sendtx.NewTxSender(submiterMock, nil),
			logger:     hclog.NewNullLogger(),
		}

		rawTx := buildTestRawTx(t, privateKey.PublicKey())

		validSig := make([]byte, 64)
		validSig[0] = 1

		batch := &eth.ConfirmedBatch{
			RawTransaction: rawTx,
			Bitmap:         new(big.Int),
			Signatures:     [][]byte{validSig},
		}

		expectedSig := solana.Signature{1, 2, 3}

		submiterMock.On("SendTransaction", ctx, mock.AnythingOfType("*solana.Transaction")).
			Return(expectedSig, nil).Once()

		err := ops.SendTx(ctx, nil, batch)
		require.NoError(t, err)

		submiterMock.AssertExpectations(t)
	})
}
