package clideploysolana

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gagliardetto/solana-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mockAuthorityKeyConfig(t *testing.T, current, next solana.PrivateKey) string {
	t.Helper()
	t.Setenv("AWS_ACCESS_KEY_ID", "test")
	t.Setenv("AWS_SECRET_ACCESS_KEY", "test")
	t.Setenv("AWS_PROFILE", "")
	t.Setenv("AWS_EC2_METADATA_DISABLED", "true")

	keys := map[string]solana.PrivateKey{
		"/bridge/node/other_solana_admin":     current,
		"/bridge/node/other_new_solana_admin": next,
	}

	ssm := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var request struct {
			Name           string
			WithDecryption bool
		}
		if !assert.NoError(t, json.NewDecoder(r.Body).Decode(&request)) {
			return
		}

		key, ok := keys[request.Name]
		if !assert.True(t, ok, "unexpected SSM parameter %s", request.Name) {
			return
		}

		assert.Equal(t, "AmazonSSM.GetParameter", r.Header.Get("X-Amz-Target"))
		assert.True(t, request.WithDecryption)

		value, err := json.Marshal(key)
		if assert.NoError(t, err) {
			assert.NoError(t, json.NewEncoder(w).Encode(map[string]any{"Parameter": map[string]string{
				"Name": request.Name, "Type": "SecureString", "Value": string(value),
			}}))
		}
	}))
	t.Cleanup(ssm.Close)

	cfg := filepath.Join(t.TempDir(), "aws.json")
	config, err := json.Marshal(map[string]any{
		"type": "aws-ssm", "name": "node", "server_url": ssm.URL,
		"extra": map[string]string{"region": "us-east-1", "ssm-parameter-path": "/bridge"},
	})
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(cfg, config, 0o600))

	return cfg
}

func TestSetUpgradeAuthorityFromSSM(t *testing.T) {
	current := solana.NewWallet().PrivateKey
	next := solana.NewWallet().PrivateKey
	cfg := mockAuthorityKeyConfig(t, current, next)
	program := solana.NewWallet().PublicKey()
	p := setUpgradeAuthorityParams{
		rpcURL: "http://localhost:8899", programID: program.String(), privateKeyConfig: cfg, commitment: "confirmed",
	}
	require.NoError(t, p.validateFlags())
	require.Equal(t, current.PublicKey(), p.currentPrivateKey.PublicKey())
	require.Equal(t, next.PublicKey(), p.newPrivateKey.PublicKey())

	for _, payer := range []solana.PrivateKey{current, solana.NewWallet().PrivateKey} {
		tx, err := signUpgradeAuthorityChange(program, solana.Hash{1}, p.currentPrivateKey, p.newPrivateKey, payer)
		require.NoError(t, err)
		require.NoError(t, tx.VerifySignatures())
		require.Equal(t, payer.PublicKey(), tx.Message.AccountKeys[0])
		require.Len(t, tx.Message.Instructions, 1)
		ix := tx.Message.Instructions[0]
		require.Equal(t, solana.BPFLoaderUpgradeableProgramID, tx.Message.AccountKeys[ix.ProgramIDIndex])
		require.Equal(t, []byte{7, 0, 0, 0}, []byte(ix.Data))
		require.Len(t, ix.Accounts, 3)

		programData, _, err := solana.FindProgramAddress([][]byte{program.Bytes()}, solana.BPFLoaderUpgradeableProgramID)
		require.NoError(t, err)
		require.Equal(t, programData, tx.Message.AccountKeys[ix.Accounts[0]])
		require.Equal(t, current.PublicKey(), tx.Message.AccountKeys[ix.Accounts[1]])
		require.Equal(t, next.PublicKey(), tx.Message.AccountKeys[ix.Accounts[2]])
		require.True(t, tx.Message.IsSigner(current.PublicKey()))
		require.True(t, tx.Message.IsSigner(next.PublicKey()))
	}

	blockhash := solana.Hash{2}
	wantTx, err := signUpgradeAuthorityChange(program, blockhash, current, next, current)
	require.NoError(t, err)

	wantSignature := wantTx.Signatures[0].String()

	rpc := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			ID     json.RawMessage
			Method string
			Params []json.RawMessage
		}
		if !assert.NoError(t, json.NewDecoder(r.Body).Decode(&req)) {
			return
		}

		var result any

		switch req.Method {
		case "getLatestBlockhash":
			result = map[string]any{"value": map[string]string{"blockhash": blockhash.String()}}
		case "sendTransaction":
			var encoded string
			if !assert.NotEmpty(t, req.Params) || !assert.NoError(t, json.Unmarshal(req.Params[0], &encoded)) {
				return
			}

			tx, err := solana.TransactionFromBase64(encoded)
			if !assert.NoError(t, err) {
				return
			}

			assert.NoError(t, tx.VerifySignatures())
			assert.Equal(t, wantTx.Signatures, tx.Signatures)

			result = wantSignature
		case "getSignatureStatuses":
			result = map[string]any{"value": []map[string]string{{"confirmationStatus": "confirmed"}}}
		default:
			t.Errorf("unexpected RPC method %s", req.Method)
			http.Error(w, "unknown method", http.StatusBadRequest)

			return
		}

		assert.NoError(t, json.NewEncoder(w).Encode(map[string]any{"jsonrpc": "2.0", "id": req.ID, "result": result}))
	}))
	defer rpc.Close()

	p.rpcURL = rpc.URL
	result, err := p.Execute(fakeOutputter{})
	require.NoError(t, err)
	require.Contains(t, result.GetOutput(), wantSignature)
}
