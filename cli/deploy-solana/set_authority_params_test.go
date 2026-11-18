package clideploysolana

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/Ethernal-Tech/apex-bridge/common"
	solsendtx "github.com/Ethernal-Tech/solana-infrastructure/sendtx"
	"github.com/Ethernal-Tech/solana-infrastructure/sendtx/skyline_program"
	solanawallet "github.com/Ethernal-Tech/solana-infrastructure/wallet"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func newValidSetAuthorityParams(t *testing.T) *setAuthorityParams {
	t.Helper()

	dir := t.TempDir()
	writeTestKeypair(t, filepath.Join(dir, "admin.json"))

	p := &setAuthorityParams{}
	p.rpcURL = "http://127.0.0.1:8899"
	p.programID = solana.NewWallet().PublicKey().String()
	p.adminKeyPath = filepath.Join(dir, "admin.json")
	p.confirmationTimeoutSeconds = 120
	p.newAuthorityKeyPath = filepath.Join(dir, "new-admin.json")
	p.newAuthority = writeTestKeypair(t, p.newAuthorityKeyPath).PublicKey().String()

	return p
}

func TestSetAuthorityValidateFlags(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		p := newValidSetAuthorityParams(t)
		require.NoError(t, p.validateFlags())
		require.Equal(t, p.newAuthority, p.newAuthorityKey.PublicKey().String())
	})

	t.Run("missing new authority", func(t *testing.T) {
		p := newValidSetAuthorityParams(t)
		p.newAuthority = ""
		require.ErrorContains(t, p.validateFlags(), "new-authority")
	})

	t.Run("missing new signer", func(t *testing.T) {
		p := newValidSetAuthorityParams(t)
		p.newAuthorityKeyPath = ""
		require.ErrorContains(t, p.validateFlags(), "--new-authority-key")
	})

	t.Run("unreadable new signer", func(t *testing.T) {
		p := newValidSetAuthorityParams(t)
		p.newAuthorityKeyPath = filepath.Join(t.TempDir(), "absent.json")
		require.ErrorContains(t, p.validateFlags(), "invalid --new-authority-key")
	})

	t.Run("new signer differs from requested address", func(t *testing.T) {
		p := newValidSetAuthorityParams(t)
		p.newAuthority = solana.NewWallet().PublicKey().String()
		require.ErrorContains(t, p.validateFlags(), "--new-authority-key does not match --new-authority")
	})

	t.Run("not base58", func(t *testing.T) {
		p := newValidSetAuthorityParams(t)
		p.newAuthority = "not-a-key-0OIl"
		require.ErrorContains(t, p.validateFlags(), "new-authority")
	})

	t.Run("default key", func(t *testing.T) {
		p := newValidSetAuthorityParams(t)
		p.newAuthority = solana.PublicKey{}.String()
		require.ErrorContains(t, p.validateFlags(), "default")
	})

	t.Run("off-curve key", func(t *testing.T) {
		p := newValidSetAuthorityParams(t)
		pda, _, err := solana.FindProgramAddress([][]byte{[]byte("fee_config")}, solana.NewWallet().PublicKey())
		require.NoError(t, err)
		require.False(t, pda.IsOnCurve())

		p.newAuthority = pda.String()
		require.ErrorContains(t, p.validateFlags(), "not on ed25519 curve")
	})

	t.Run("same as admin", func(t *testing.T) {
		p := newValidSetAuthorityParams(t)
		key := writeTestKeypair(t, p.adminKeyPath)
		p.newAuthority = key.PublicKey().String()
		require.ErrorContains(t, p.validateFlags(), "new authority equals current admin")
	})

	t.Run("neither flag", func(t *testing.T) {
		p := newValidSetAuthorityParams(t)
		p.adminKeyPath = ""
		err := p.validateFlags()
		require.ErrorContains(t, err, "--admin-key")
		require.ErrorContains(t, err, "--key-config")
	})
}

func TestSetAuthorityFlags(t *testing.T) {
	cmd := &cobra.Command{Use: "x"}
	(&setAuthorityParams{}).setFlags(cmd)

	for _, name := range []string{"url", "program-id", "admin-key", "key-config", "new-authority", "new-authority-key",
		"confirmation-timeout-seconds"} {
		require.NotNil(t, cmd.Flags().Lookup(name), name)
	}

	require.NoError(t, cmd.ParseFlags([]string{"--admin-key", "a", "--key-config", "b"}))
	require.ErrorContains(t, cmd.ValidateFlagGroups(), "admin-key")
}

func TestSetAuthorityPrecheck(t *testing.T) {
	admin := solana.NewWallet().PublicKey()
	other := solana.NewWallet().PublicKey()
	newAuth := solana.NewWallet().PublicKey()

	t.Run("ok", func(t *testing.T) {
		require.NoError(t, checkCurrentAuthority(admin, admin, newAuth))
	})

	t.Run("admin mismatch", func(t *testing.T) {
		err := checkCurrentAuthority(other, admin, newAuth)
		require.ErrorContains(t, err, "admin key "+admin.String()+" does not match program config authority "+other.String())
	})

	t.Run("new equals on-chain", func(t *testing.T) {
		err := checkCurrentAuthority(admin, admin, admin)
		require.ErrorContains(t, err, "already")
	})
}

func TestSetAuthorityDto(t *testing.T) {
	admin := solana.NewWallet().PrivateKey
	newAuth := solana.NewWallet().PublicKey()
	programID := solana.NewWallet().PublicKey()

	dto := buildSetAuthorityDto(programID, admin.PublicKey(), newAuth)
	require.Equal(t, solsendtx.SetAuthorityDto{
		ProgramID:        programID,
		AuthorityAddr:    admin.PublicKey().String(),
		NewAuthorityAddr: newAuth.String(),
	}, dto)

	tx, err := solsendtx.NewTxSender(nil, nil).CreateTx(
		context.Background(), admin.PublicKey(), solsendtx.InstructionTypeSetAuthority, solana.Hash{}, dto)
	require.NoError(t, err)
	require.Equal(t, admin.PublicKey(), tx.Message.AccountKeys[0])

	_, err = tx.Sign(func(k solana.PublicKey) *solana.PrivateKey {
		if k == admin.PublicKey() {
			return &admin
		}

		return nil
	})
	require.ErrorContains(t, err, newAuth.String())
}

func TestSetAuthorityReadBack(t *testing.T) {
	want := solana.NewWallet().PublicKey()

	require.NoError(t, checkAuthorityChanged("sig1", want, want))

	err := checkAuthorityChanged("sigABC", solana.NewWallet().PublicKey(), want)
	require.ErrorContains(t, err, "sigABC")
}

func TestSetAuthoritySubmitsCosignedTransaction(t *testing.T) {
	p := newValidSetAuthorityParams(t)
	require.NoError(t, p.validateFlags())
	admin, recipient := p.adminPrivateKey.PublicKey(), p.newAuthorityKey.PublicKey()

	ctx := context.Background()
	provider := new(solanawallet.MockTxProvider)
	pda, _, err := solana.FindProgramAddress([][]byte{skyline_program.PROGRAM_CONFIG_SEED}, p.programPublicKey)
	require.NoError(t, err)

	for _, authority := range []solana.PublicKey{admin, recipient} {
		cfg := skyline_program.ProgramConfig{Authority: authority}
		body, err := cfg.Marshal()
		require.NoError(t, err)

		data := append(skyline_program.Account_ProgramConfig[:], body...)
		provider.On("GetAccountInfo", ctx, pda).Return(&rpc.GetAccountInfoResult{
			Value: &rpc.Account{Data: rpc.DataBytesOrJSONFromBytes(data)},
		}, nil).Once()
	}

	provider.On("GetLatestBlockhash", ctx).Return(solana.Hash{1}, nil).Once()

	sig := solana.Signature{1}
	provider.On("SendTransaction", ctx, mock.Anything).Run(func(args mock.Arguments) {
		tx := args.Get(1).(*solana.Transaction)
		require.Equal(t, admin, tx.Message.AccountKeys[0])
		require.ElementsMatch(t, []solana.PublicKey{admin, recipient}, tx.Message.Signers())
		require.NoError(t, tx.VerifySignatures())
	}).Return(sig, nil).Once()
	provider.On("WaitForSignature", ctx, sig, rpc.CommitmentFinalized, p.confirmationTimeout).Return(nil).Once()

	result, err := p.rotateAuthority(ctx, provider, common.InitializeOutputter(&cobra.Command{}))
	require.NoError(t, err)
	require.Contains(t, result.GetOutput(), sig.String())
	provider.AssertExpectations(t)
}
