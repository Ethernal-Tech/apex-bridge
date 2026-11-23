package clideploysolana

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/solana-infrastructure/sendtx/skyline_program"
	solanawallet "github.com/Ethernal-Tech/solana-infrastructure/wallet"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestRenameMintBurnTokenFlags(t *testing.T) {
	cmd, _, err := GetDeploySolanaCommand().Find([]string{"rename-mint-burn-token"})
	require.NoError(t, err)
	require.Equal(t, renameMintBurnTokenCommandUse, cmd.Name())

	path := filepath.Join(t.TempDir(), "admin.json")
	admin := writeTestKeypair(t, path)
	require.NoError(t, cmd.ParseFlags([]string{
		"--url", "http://localhost:8899", "--program-id", solana.NewWallet().PublicKey().String(),
		"--admin-key", path, "--token-id", "258", "--token-name", "Renamed",
	}))
	require.ErrorContains(t, cmd.PreRunE(cmd, nil), "--token-symbol")
	require.NoError(t, cmd.Flags().Set("token-symbol", "NEW"))
	require.NoError(t, cmd.PreRunE(cmd, nil))
	require.Equal(t, admin, renameMintBurnTokenParamsData.adminPrivateKey)
	require.NoError(t, cmd.Flags().Set("token-id", "0"))
	require.ErrorContains(t, cmd.PreRunE(cmd, nil), "--token-id")
	require.NoError(t, cmd.Flags().Set("key-config", "secrets.json"))
	require.ErrorContains(t, cmd.ValidateFlagGroups(), "admin-key")
}

func TestRenameMintBurnTokenTransaction(t *testing.T) {
	ctx := context.Background()
	admin := solana.NewWallet().PrivateKey
	program := solana.NewWallet().PublicKey()
	mint := solana.NewWallet().PublicKey()
	p := &renameMintBurnTokenParams{
		programAdminParams: programAdminParams{
			rpcURL: "http://localhost:8899", programID: program.String(),
			privateKeyConfig: mockAuthorityKeyConfig(t, admin, nil), confirmationTimeoutSeconds: 15,
		},
		tokenID: 258, tokenName: "Renamed", tokenSymbol: "NEW",
	}
	require.NoError(t, p.validateFlags())

	registry, _, err := solana.FindProgramAddress([][]byte{skyline_program.TOKEN_REGISTRY_SEED, {2, 1}}, program)
	require.NoError(t, err)
	metadata, _, err := solana.FindTokenMetadataAddress(mint)
	require.NoError(t, err)
	feeConfig, _, err := solana.FindProgramAddress([][]byte{skyline_program.FEE_CONFIG_SEED}, program)
	require.NoError(t, err)
	vault, _, err := solana.FindProgramAddress([][]byte{skyline_program.VAULT_SEED}, program)
	require.NoError(t, err)

	mockRegistry := func(t *testing.T, lockUnlock bool) *solanawallet.MockTxProvider {
		t.Helper()

		provider := new(solanawallet.MockTxProvider)
		entry := skyline_program.TokenRegistry{TokenId: p.tokenID, Mint: mint, IsLockUnlock: lockUnlock}
		body, err := entry.Marshal()
		require.NoError(t, err)
		provider.On("GetAccountInfo", ctx, registry).Return(&rpc.GetAccountInfoResult{
			Value: &rpc.Account{Owner: program, Data: rpc.DataBytesOrJSONFromBytes(append(skyline_program.Account_TokenRegistry[:], body...))},
		}, nil).Once()

		return provider
	}

	t.Run("lock/unlock rejected before signing", func(t *testing.T) {
		provider := mockRegistry(t, true)
		_, err := p.renameToken(ctx, provider, common.InitializeOutputter(&cobra.Command{}))
		require.ErrorContains(t, err, "mint/burn")
		provider.AssertExpectations(t)
	})

	t.Run("mint/burn signed and finalized", func(t *testing.T) {
		provider := mockRegistry(t, false)
		blockhash := solana.Hash{2}
		sig := solana.Signature{3}

		provider.On("GetLatestBlockhash", ctx).Return(blockhash, nil).Once()
		provider.On("SendTransaction", ctx, mock.Anything).Run(func(args mock.Arguments) {
			tx := args.Get(1).(*solana.Transaction)
			require.NoError(t, tx.VerifySignatures())
			require.Equal(t, solana.PublicKeySlice{admin.PublicKey()}, tx.Message.Signers())
			require.Equal(t, admin.PublicKey(), tx.Message.AccountKeys[0])
			require.Equal(t, blockhash, tx.Message.RecentBlockhash)
			require.Len(t, tx.Message.Instructions, 1)
			ix := tx.Message.Instructions[0]
			require.Equal(t, program, tx.Message.AccountKeys[ix.ProgramIDIndex])

			wantAccounts := []solana.PublicKey{admin.PublicKey(), feeConfig, registry, mint, vault,
				metadata, solana.TokenMetadataProgramID, solana.SystemProgramID, solana.SysVarInstructionsPubkey}
			require.Len(t, ix.Accounts, len(wantAccounts))

			for i, account := range ix.Accounts {
				require.Equal(t, wantAccounts[i], tx.Message.AccountKeys[account])
			}
			// Borsh u16 token ID, followed by length-prefixed name and symbol.
			wantData := append(skyline_program.Instruction_RenameMintBurnToken[:], []byte("\x02\x01\x07\x00\x00\x00Renamed\x03\x00\x00\x00NEW")...)
			require.Equal(t, wantData, []byte(ix.Data))
		}).Return(sig, nil).Once()
		provider.On("WaitForSignature", ctx, sig, rpc.CommitmentFinalized, p.confirmationTimeout).Return(nil).Once()

		result, err := p.renameToken(ctx, provider, common.InitializeOutputter(&cobra.Command{}))
		require.NoError(t, err)
		require.Contains(t, result.GetOutput(), sig.String())
		require.Contains(t, result.GetOutput(), mint.String())
		provider.AssertExpectations(t)
	})
}
