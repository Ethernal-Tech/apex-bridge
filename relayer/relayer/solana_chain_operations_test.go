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

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/eth"
	"github.com/Ethernal-Tech/apex-bridge/relayer/core"
	solanatx "github.com/Ethernal-Tech/apex-bridge/solana"
	"github.com/Ethernal-Tech/cardano-infrastructure/secrets"
	secretsHelper "github.com/Ethernal-Tech/cardano-infrastructure/secrets/helper"
	"github.com/Ethernal-Tech/solana-infrastructure/sendtx"
	"github.com/Ethernal-Tech/solana-infrastructure/sendtx/skyline_program"
	solanaWallet "github.com/Ethernal-Tech/solana-infrastructure/wallet"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func Test_DecodePayload(t *testing.T) {
	payloadBytes, err := hex.DecodeString(
		"bc69d80eedd69641ee4109ea79868425a3b4946afa2e5683ee946b262b659729" +
			"0168c61393628c69f2dd67c2a6754c1b3a706deae20f659f55432ef9cbc24d" +
			"0198190000ca9a3b0000000000ca9a3b000000000100000000000000",
	)
	require.NoError(t, err)

	var payload sendtx.SolanaPayload
	err = payload.Unmarshal(payloadBytes)
	require.NoError(t, err)

	require.Equal(t, uint64(1), payload.BatchID)
	require.Equal(t, uint64(1_000_000_000), payload.FeeAmount)
	require.Len(t, payload.Receivers, 1)
	require.NotEqual(t, [32]byte{}, payload.Blockhash)
}

func buildTestRawTx(t *testing.T, _ solana.PublicKey) []byte {
	t.Helper()

	receiverWallet, err := solana.NewRandomPrivateKey()
	require.NoError(t, err)

	blockHash, err := solana.NewRandomPrivateKey()
	require.NoError(t, err)

	payload := sendtx.SolanaPayload{
		Blockhash: [32]byte(blockHash.PublicKey()),
		Receivers: []sendtx.PayloadReceiver{
			{
				Address: [32]byte(receiverWallet.PublicKey()),
				TokenAmount: solanaWallet.TokenAmount{
					TokenID: 1,
					Amount:  1_000_000,
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

func mockBridgeTokenRegistry(
	t *testing.T,
	mockProvider *solanaWallet.MockTxProvider,
	regs ...skyline_program.TokenRegistry,
) {
	t.Helper()

	results := make(rpc.GetProgramAccountsResult, 0, len(regs))

	for _, reg := range regs {
		body, err := reg.Marshal()
		require.NoError(t, err)

		raw := append(append([]byte(nil), skyline_program.Account_TokenRegistry[:]...), body...)
		results = append(results, &rpc.KeyedAccount{
			Pubkey:  solana.NewWallet().PublicKey(),
			Account: &rpc.Account{Data: rpc.DataBytesOrJSONFromBytes(raw)},
		})
	}

	mockProvider.On(
		"GetProgramAccounts",
		mock.Anything,
		skyline_program.ProgramID,
		mock.AnythingOfType("*rpc.GetProgramAccountsOpts"),
	).Return(results, nil)
}

func defaultBridgeTokenRegistry(tokenID uint16) skyline_program.TokenRegistry {
	return skyline_program.TokenRegistry{
		TokenId:           tokenID,
		Mint:              solana.NewWallet().PublicKey(),
		IsLockUnlock:      true,
		MinBridgingAmount: 10,
		Bump:              250,
	}
}

func newSendTxTestOps(t *testing.T, mockProvider *solanaWallet.MockTxProvider, privateKey solana.PrivateKey) *SolanaChainOperations {
	t.Helper()

	mockBridgeTokenRegistry(t, mockProvider, defaultBridgeTokenRegistry(1))

	return &SolanaChainOperations{
		chainID: common.ChainIDStrSolana,
		config: &solanatx.SolanaChainConfig{
			TxProviderEndpoint: rpc.LocalNet_WS,
			ProgramID:          skyline_program.ProgramID.String(),
		},
		privateKey: &privateKey,
		txSender:   sendtx.NewTxSender(mockProvider, nil),
		logger:     hclog.NewNullLogger(),
	}
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
		submiterMock := &solanaWallet.MockTxProvider{}

		ops := &SolanaChainOperations{
			chainID:    common.ChainIDStrSolana,
			config:     &solanatx.SolanaChainConfig{TxProviderEndpoint: rpc.LocalNet_WS, ProgramID: "CkTNcuk9EELmuR65eCfzKfz8XpDvJ27FPFHauGHVD1E9"},
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
		submiterMock := &solanaWallet.MockTxProvider{}

		ops := &SolanaChainOperations{
			chainID:    common.ChainIDStrSolana,
			config:     &solanatx.SolanaChainConfig{TxProviderEndpoint: rpc.LocalNet_WS, ProgramID: "CkTNcuk9EELmuR65eCfzKfz8XpDvJ27FPFHauGHVD1E9"},
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
		require.Contains(t, err.Error(), "failed to unmarshal solana payload")
	})

	t.Run("send returns error", func(t *testing.T) {
		submiterMock := &solanaWallet.MockTxProvider{}
		ops := newSendTxTestOps(t, submiterMock, privateKey)

		rawTx := buildTestRawTx(t, privateKey.PublicKey())
		validSigs := make([][]byte, 3)
		bridgeMock := buildBridgeMockWithSignatures(t, ctx, common.ChainIDStrSolana, rawTx, validSigs)

		batch := &eth.ConfirmedBatch{
			ID:             2,
			RawTransaction: rawTx,
			Bitmap:         new(big.Int),
			Signatures:     validSigs,
		}

		submiterMock.On("SendTransaction", mock.Anything, mock.AnythingOfType("*solana.Transaction")).
			Return(solana.Signature{}, errors.New("rpc unavailable")).Once()

		err := ops.SendTx(ctx, bridgeMock, batch)
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to send tx")

		submiterMock.AssertExpectations(t)
	})

	t.Run("create tx returns error when token registry lookup fails", func(t *testing.T) {
		submiterMock := &solanaWallet.MockTxProvider{}
		submiterMock.On(
			"GetProgramAccounts",
			mock.Anything,
			skyline_program.ProgramID,
			mock.AnythingOfType("*rpc.GetProgramAccountsOpts"),
		).Return(rpc.GetProgramAccountsResult(nil), errors.New("rpc unavailable")).Once()

		ops := &SolanaChainOperations{
			chainID: common.ChainIDStrSolana,
			config: &solanatx.SolanaChainConfig{
				TxProviderEndpoint: rpc.LocalNet_WS,
				ProgramID:          skyline_program.ProgramID.String(),
			},
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

		err := ops.SendTx(ctx, bridgeMock, batch)
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to create tx")
		require.Contains(t, err.Error(), "get token registry accounts")

		submiterMock.AssertExpectations(t)
		submiterMock.AssertNotCalled(t, "SendTransaction", mock.Anything, mock.Anything)
	})

	t.Run("success", func(t *testing.T) {
		submiterMock := &solanaWallet.MockTxProvider{}
		ops := newSendTxTestOps(t, submiterMock, privateKey)

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

		submiterMock.On("SendTransaction", mock.Anything, mock.AnythingOfType("*solana.Transaction")).
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
	receiverWallet, err := solana.NewRandomPrivateKey()
	require.NoError(t, err)

	blockHashWallet, err := solana.NewRandomPrivateKey()
	require.NoError(t, err)

	payload := sendtx.SolanaPayload{
		Receivers: []sendtx.PayloadReceiver{
			{
				Address: [32]byte(receiverWallet.PublicKey()),
				TokenAmount: solanaWallet.TokenAmount{
					TokenID: 1,
					Amount:  1_000_000_000,
				},
			},
		},
		BatchID:   1,
		Blockhash: [32]byte(blockHashWallet.PublicKey()),
	}

	payloadBytes, err := payload.Marshal()
	require.NoError(t, err)

	var payload2 sendtx.SolanaPayload
	err = payload2.Unmarshal(payloadBytes)
	require.NoError(t, err)
	require.Equal(t, payload, payload2)
}
