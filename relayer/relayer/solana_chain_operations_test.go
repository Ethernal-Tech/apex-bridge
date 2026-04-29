package relayer

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Ethernal-Tech/apex-bridge/common"
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

func Test_DecodePayload(t *testing.T) {
	payloadString := "2b0000000000000036694e6b6d74484a455757644568344d384a43547964535059547048324c3738667a563979445162627677012c000000000000003735627673687851445034597966783763447346624765525455463650695341664a716f613931344c5944392b00000000000000536f31313131313131313131313131313131313131313131313131313131313131313131313131313131320a00000000000000323030303030303030303000000000000000"
	payloadBytes, err := hex.DecodeString(payloadString)
	require.NoError(t, err)

	var payload sendtx.SolanaPayload
	err = payload.Unmarshal(payloadBytes)
	require.NoError(t, err)

	// fmt.Println(payload)

	require.Equal(t, payload.BatchID, uint64(48))
	require.Equal(t, payload.Blockhash, "6iNkmtHJEWWdEh4M8JCTydSPYTpH2L78fzV9yDQbbvw")
	require.Equal(t, payload.Receivers[0].Address, "75bvshxQDP4Yyfx7cDsFbGeRTUF6PiSAfJqoa914LYD9")
	require.Equal(t, payload.Receivers[0].TokenAmount.Amount, big.NewInt(2000000000))
	require.Equal(t, payload.Receivers[0].TokenAmount.TokenMint, "So11111111111111111111111111111111111111112")
}

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

func buildTestRawTx(t *testing.T, _ solana.PublicKey) []byte {
	t.Helper()

	receiverWallet, err := solana.NewRandomPrivateKey()
	require.NoError(t, err)

	blockHash, err := solana.NewRandomPrivateKey()
	require.NoError(t, err)

	payload := sendtx.SolanaPayload{
		Blockhash: solana.Hash(blockHash.PublicKey()).String(),
		Receivers: []sendtx.BridgingTxReceiver{
			{
				Address: receiverWallet.PublicKey().String(),
				TokenAmount: solanaWallet.TokenAmount{
					Amount:    new(big.Int).SetUint64(1_000_000),
					TokenMint: solana.NewWallet().PublicKey().String(),
				},
			},
		},
		BatchID: 2,
	}

	raw, err := payload.Marshal()
	require.NoError(t, err)

	return raw
}

func buildBridgeMockWithSignatures(
	t *testing.T, ctx context.Context, chainID string, payloadBytes []byte, signatures [][]byte,
) *eth.BridgeSmartContractMock {
	t.Helper()

	validatorsCount := uint64(4)
	validatorsThreashold := common.GetRequiredSignaturesForConsensus(validatorsCount)

	if signatures == nil {
		signatures = make([][]byte, validatorsThreashold)
	}

	validators := make([]eth.ValidatorChainData, validatorsCount)
	for i := range validators {
		priv, err := solana.NewRandomPrivateKey()
		require.NoError(t, err)

		sig, err := priv.Sign(payloadBytes)
		require.NoError(t, err)

		if i < int(validatorsThreashold) {
			sigBytes := make([]byte, len(sig))
			copy(sigBytes, sig[:])
			signatures[i] = sigBytes
		}

		validators[i] = eth.ValidatorChainData{
			Key: [4]*big.Int{new(big.Int).SetBytes(priv.PublicKey().Bytes()), big.NewInt(0), big.NewInt(0), big.NewInt(0)},
		}
	}

	bridgeMock := &eth.BridgeSmartContractMock{}
	bridgeMock.On("GetValidatorsChainData", ctx, chainID).Return(validators, nil)

	return bridgeMock
}

func TestNewSolanaChainOperations(t *testing.T) {
	const chainID = common.ChainIDStrSolana

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
			chainID:    common.ChainIDStrSolana,
			config:     &solanatx.SolanaChainConfig{TxProviderEndpoint: rpc.LocalNet_WS},
			privateKey: &privateKey,
			txSender:   sendtx.NewTxSender(submiterMock, nil),
			logger:     hclog.NewNullLogger(),
		}

		rawTx := buildTestRawTx(t, privateKey.PublicKey())
		bridgeMock := buildBridgeMockWithSignatures(t, ctx, common.ChainIDStrSolana, rawTx, nil)

		batch := &eth.ConfirmedBatch{
			ID:             2,
			RawTransaction: rawTx,
			Bitmap:         new(big.Int),
			Signatures:     [][]byte{make([]byte, 64), make([]byte, 64), make([]byte, 64)},
		}

		err := ops.SendTx(ctx, bridgeMock, batch)
		require.Error(t, err)
		require.Contains(t, err.Error(), "no signature pairs provided")
	})

	t.Run("invalid RawTransaction", func(t *testing.T) {
		submiterMock := &mockTxSubmiter{}

		ops := &SolanaChainOperations{
			chainID:    common.ChainIDStrSolana,
			config:     &solanatx.SolanaChainConfig{TxProviderEndpoint: rpc.LocalNet_WS},
			privateKey: &privateKey,
			txSender:   sendtx.NewTxSender(submiterMock, nil),
			logger:     hclog.NewNullLogger(),
		}

		validSigs := make([][]byte, 3)
		for i := range validSigs {
			validSigs[i] = make([]byte, 64)
			validSigs[i][0] = 1
		}

		batch := &eth.ConfirmedBatch{
			RawTransaction: []byte("garbage"),
			Bitmap:         new(big.Int),
			Signatures:     validSigs,
		}

		err := ops.SendTx(ctx, &eth.BridgeSmartContractMock{}, batch)
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to unmarshal payload")
	})

	t.Run("send returns error", func(t *testing.T) {
		submiterMock := &mockTxSubmiter{}

		ops := &SolanaChainOperations{
			chainID:    common.ChainIDStrSolana,
			config:     &solanatx.SolanaChainConfig{TxProviderEndpoint: rpc.LocalNet_WS},
			privateKey: &privateKey,
			txSender:   sendtx.NewTxSender(submiterMock, nil),
			logger:     hclog.NewNullLogger(),
		}

		rawTx := buildTestRawTx(t, privateKey.PublicKey())
		validSigs := make([][]byte, 3)
		bridgeMock := buildBridgeMockWithSignatures(t, ctx, common.ChainIDStrSolana, rawTx, validSigs)

		batch := &eth.ConfirmedBatch{
			ID:             2,
			RawTransaction: rawTx,
			Bitmap:         new(big.Int),
			Signatures:     validSigs,
		}

		submiterMock.On("SendTransaction", ctx, mock.AnythingOfType("*solana.Transaction")).
			Return(solana.Signature{}, errors.New("rpc unavailable")).Once()

		err := ops.SendTx(ctx, bridgeMock, batch)
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to send tx")

		submiterMock.AssertExpectations(t)
	})

	t.Run("success", func(t *testing.T) {
		submiterMock := &mockTxSubmiter{}

		ops := &SolanaChainOperations{
			chainID:    common.ChainIDStrSolana,
			config:     &solanatx.SolanaChainConfig{TxProviderEndpoint: rpc.LocalNet_WS},
			privateKey: &privateKey,
			txSender:   sendtx.NewTxSender(submiterMock, nil),
			logger:     hclog.NewNullLogger(),
		}

		rawTx := buildTestRawTx(t, privateKey.PublicKey())
		validSigs := make([][]byte, 3)
		bridgeMock := buildBridgeMockWithSignatures(t, ctx, common.ChainIDStrSolana, rawTx, validSigs)

		batch := &eth.ConfirmedBatch{
			ID:             2,
			RawTransaction: rawTx,
			Bitmap:         new(big.Int),
			Signatures:     validSigs,
		}

		expectedSig := solana.Signature{1, 2, 3}

		submiterMock.On("SendTransaction", ctx, mock.AnythingOfType("*solana.Transaction")).
			Return(expectedSig, nil).Once()

		err := ops.SendTx(ctx, bridgeMock, batch)
		require.NoError(t, err)

		submiterMock.AssertExpectations(t)
	})
}

func TestSolanaChainOperations_getSignaturePairs(t *testing.T) {
	ctx := context.Background()

	privateKey, err := solana.NewRandomPrivateKey()
	require.NoError(t, err)

	signatures := []string{
		"1d2029ec9be00fa914308ac316971ac6a6ea396bcc88604c05b7f2508523a82b27a656ffe71428438d75e043ce34c210dd09a4cd26f6a2f883e14b658a7a5009",
		"19bd584c978f466256ce99f0d91a8c800658021936f5900e125ff57d0bdfede378df99718b80f10a75440ddea1137584e89402499964274cadc87f90ed55250f",
		"d3fd84e5deb00a20bb18b793b6be5e2a41435a7806db5af40b4ae519c4406672f1e2df0bda8d6093ad08aeb1a774053adac585dbc6c8b96c0e36d01fbc35ff05",
	}

	signatureBytes := make([][]byte, len(signatures))

	for i, signature := range signatures {
		signatureBytes[i], err = hex.DecodeString(signature)
		require.NoError(t, err)
	}

	validatorKeys := []string{
		"Ggc6De36VRJrgDQG8KFkGPTDbQBKQ1PSpyP9JLappixz",
		"GcSam1r9z7ixqPanySqrPgV3inz3j7aCexLuosnJry9Y",
		"f7o6dvzJRQcBspfiNRLzyyFtLSdxgKyvNc3rZCUJg8T",
		"F3rcR9BR4hnyHyPWKh3PotJrtLa7AiAo4YVtki9LC2xC",
	}

	validatorPublicKeys := make([]solana.PublicKey, len(validatorKeys))

	for i, validatorKey := range validatorKeys {
		pubKey, err := solanaWallet.PublicKeyFromAddress(validatorKey)
		require.NoError(t, err)

		validatorPublicKeys[i] = pubKey
	}

	bridgeSmartContractMock := &eth.BridgeSmartContractMock{}
	bridgeSmartContractMock.On("GetValidatorsChainData", ctx, common.ChainIDStrSolana).Return([]eth.ValidatorChainData{
		{
			Key: [4]*big.Int{new(big.Int).SetBytes(validatorPublicKeys[0].Bytes()), big.NewInt(0), big.NewInt(0), big.NewInt(0)},
		},
		{
			Key: [4]*big.Int{new(big.Int).SetBytes(validatorPublicKeys[1].Bytes()), big.NewInt(0), big.NewInt(0), big.NewInt(0)},
		},
		{
			Key: [4]*big.Int{new(big.Int).SetBytes(validatorPublicKeys[2].Bytes()), big.NewInt(0), big.NewInt(0), big.NewInt(0)},
		},
		{
			Key: [4]*big.Int{new(big.Int).SetBytes(validatorPublicKeys[3].Bytes()), big.NewInt(0), big.NewInt(0), big.NewInt(0)},
		},
	}, nil)

	payloadBytes, err := hex.DecodeString("e331b7f63f87b29b4fb474900bb417d6ab113f4e5a89f6c459338bea486d3a89012c000000000000003779445271506433444a4e6f4b54556a57654a384672795a484a68344d6641707554343232557a765a5878552b00000000000000536f31313131313131313131313131313131313131313131313131313131313131313131313131313131320100000000000000")
	require.NoError(t, err)

	ops := &SolanaChainOperations{
		chainID:    common.ChainIDStrSolana,
		privateKey: &privateKey,
		txSender:   sendtx.NewTxSender(nil, nil),
		logger:     hclog.NewNullLogger(),
	}

	signaturePairs, err := ops.getSignaturePairs(ctx, bridgeSmartContractMock, signatureBytes, payloadBytes)
	require.NoError(t, err)
	require.Equal(t, len(signaturePairs), len(validatorPublicKeys)-1)
}

func TestPayload_Unmarshal(t *testing.T) {
	var payload = sendtx.SolanaPayload{
		Receivers: []sendtx.BridgingTxReceiver{
			{
				Address: "Ggc6De36VRJrgDQG8KFkGPTDbQBKQ1PSpyP9JLappixz",
				TokenAmount: solanaWallet.TokenAmount{
					Amount:    big.NewInt(1000000000000000000),
					TokenMint: solana.NewWallet().PublicKey().String(),
				},
			},
		},
		BatchID:   1,
		Blockhash: "fee448eb51b7b7dc767700ad1246eb2c57e683c644838d99934ee3619d23fefe",
	}

	payloadBytes, err := payload.Marshal()
	require.NoError(t, err)
	fmt.Println(hex.EncodeToString(payloadBytes))

	var payload2 sendtx.SolanaPayload
	err = payload2.Unmarshal(payloadBytes)
	require.NoError(t, err)
	fmt.Println(payload2)
}
