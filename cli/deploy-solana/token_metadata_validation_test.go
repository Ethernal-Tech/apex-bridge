package clideploysolana

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/gagliardetto/solana-go"
	"github.com/stretchr/testify/require"
)

func newValidRegisterMintBurnTokenParams(t *testing.T) *registerMintBurnTokenParams {
	t.Helper()

	dir := t.TempDir()
	writeTestKeypair(t, filepath.Join(dir, "admin.json"))

	p := &registerMintBurnTokenParams{}
	p.rpcURL = "http://127.0.0.1:8899"
	p.programID = solana.NewWallet().PublicKey().String()
	p.adminKeyPath = filepath.Join(dir, "admin.json")
	p.confirmationTimeoutSeconds = 120
	p.treasuryAddress = solana.NewWallet().PublicKey().String()
	p.tokenID = 1
	p.tokenName = "Wrapped Token"
	p.tokenSymbol = "WTK"
	p.tokenURI = "https://example.com/token.json"
	p.tokenDecimals = 6

	return p
}

func TestValidateTokenMetadata(t *testing.T) {
	validate := func(name, symbol, uri string) error {
		return validateTokenMetadata(name, "name", symbol, "symbol", uri, "uri")
	}

	t.Run("at the limits", func(t *testing.T) {
		require.NoError(t, validate(
			strings.Repeat("n", metaplexMaxNameLength),
			strings.Repeat("s", metaplexMaxSymbolLength),
			strings.Repeat("u", metaplexMaxURILength),
		))
	})

	t.Run("empty uri", func(t *testing.T) {
		require.NoError(t, validate("Token", "TKN", ""))
	})

	t.Run("empty name", func(t *testing.T) {
		require.ErrorContains(t, validate("", "TKN", ""), "--name")
	})

	t.Run("empty symbol", func(t *testing.T) {
		require.ErrorContains(t, validate("Token", "", ""), "--symbol")
	})

	t.Run("name too long", func(t *testing.T) {
		err := validate(strings.Repeat("n", metaplexMaxNameLength+1), "TKN", "")
		require.ErrorContains(t, err, "at most 32")
		require.ErrorContains(t, err, "--name")
	})

	t.Run("symbol too long", func(t *testing.T) {
		err := validate("Token", strings.Repeat("s", metaplexMaxSymbolLength+1), "")
		require.ErrorContains(t, err, "at most 10")
		require.ErrorContains(t, err, "--symbol")
	})

	t.Run("uri too long", func(t *testing.T) {
		err := validate("Token", "TKN", strings.Repeat("u", metaplexMaxURILength+1))
		require.ErrorContains(t, err, "at most 200")
		require.ErrorContains(t, err, "--uri")
	})

	t.Run("limits count bytes not runes", func(t *testing.T) {
		// 6 runes, 12 bytes: Metaplex measures bytes.
		require.ErrorContains(t, validate("Token", "ЖЖЖЖЖЖ", ""), "12 bytes")
	})
}

func TestRegisterMintBurnTokenValidateFlags(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		require.NoError(t, newValidRegisterMintBurnTokenParams(t).validateFlags())
	})

	t.Run("missing name", func(t *testing.T) {
		p := newValidRegisterMintBurnTokenParams(t)
		p.tokenName = ""
		require.ErrorContains(t, p.validateFlags(), "--"+registerMintBurnTokenNameFlag)
	})

	t.Run("short name is the default symbol", func(t *testing.T) {
		p := newValidRegisterMintBurnTokenParams(t)
		p.tokenName = "Token"
		p.tokenSymbol = ""
		require.NoError(t, p.validateFlags())
		require.Equal(t, "Token", p.tokenSymbol)
	})

	t.Run("long name cannot be the default symbol", func(t *testing.T) {
		p := newValidRegisterMintBurnTokenParams(t)
		p.tokenName = "Wrapped Token" // 13 bytes, over the symbol limit
		p.tokenSymbol = ""
		require.ErrorContains(t, p.validateFlags(), "--"+registerMintBurnTokenSymbolFlag)
	})

	t.Run("name too long", func(t *testing.T) {
		p := newValidRegisterMintBurnTokenParams(t)
		p.tokenName = strings.Repeat("n", metaplexMaxNameLength+1)
		require.ErrorContains(t, p.validateFlags(), "--"+registerMintBurnTokenNameFlag)
	})

	t.Run("uri too long", func(t *testing.T) {
		p := newValidRegisterMintBurnTokenParams(t)
		p.tokenURI = strings.Repeat("u", metaplexMaxURILength+1)
		require.ErrorContains(t, p.validateFlags(), "--"+registerMintBurnTokenURIFlag)
	})
}
