package tests

import (
	"context"
	"encoding/binary"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/Ethernal-Tech/apex-bridge/solana/client"
	"github.com/Ethernal-Tech/apex-bridge/solana/skyline_program"
	testvalidator "github.com/Ethernal-Tech/apex-bridge/solana/tests/test_validator"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/stretchr/testify/require"
)

func Test_SolanaTest(t *testing.T) {
	const amount = 10 * solana.LAMPORTS_PER_SOL
	const numValidators = 4

	validator := testvalidator.NewTestValidator()
	err := validator.StartTestNode()
	require.NoError(t, err)
	defer validator.Close()

	err = validator.WaitForNode(rpc.New(rpc.LocalNet_RPC))
	require.NoError(t, err)

	cli, err := client.NewSolanaClient(client.WithLocalnet())
	require.NoError(t, err)
	defer cli.Close()

	programPath, err := filepath.Abs("program_build/skyline_program-keypair.json")
	require.NoError(t, err)

	feePayerPath, err := filepath.Abs("program_build/test.json")
	require.NoError(t, err)

	buildPath, err := filepath.Abs("program_build/skyline_program.so")
	require.NoError(t, err)

	programKeypair, err := solana.PrivateKeyFromSolanaKeygenFile(programPath)
	require.NoError(t, err)

	feePayer, err := solana.PrivateKeyFromSolanaKeygenFile(feePayerPath)
	require.NoError(t, err)

	require.NoError(t, cli.Airdrop(feePayer.PublicKey(), amount))

	require.NoError(t, cli.Deploy(feePayerPath, programPath, buildPath))

	validators, validatorsPks := make([]solana.PublicKey, numValidators), make([]solana.PrivateKey, numValidators)
	for i := range numValidators {
		pk, err := solana.NewRandomPrivateKey()
		require.NoError(t, err)

		validatorsPks[i] = pk
		validators[i] = validatorsPks[i].PublicKey()
	}

	vsPda, _, err := solana.FindProgramAddress([][]byte{skyline_program.VALIDATOR_SET_SEED}, programKeypair.PublicKey())
	require.NoError(t, err)

	vaultPda, _, err := solana.FindProgramAddress([][]byte{skyline_program.VAULT_SEED}, programKeypair.PublicKey())
	require.NoError(t, err)

	initializeIx, err := skyline_program.NewInitializeInstruction(validators, nil, feePayer.PublicKey(), vsPda, vaultPda, solana.SystemProgramID)
	require.NoError(t, err)

	_, err = cli.ExecuteInstruction(&initializeIx, map[solana.PublicKey]*solana.PrivateKey{}, feePayer)
	require.NoError(t, err)

	vsInfo, err := cli.GetRpcClient().GetAccountInfo(context.Background(), vsPda)
	require.NoError(t, err)

	vs := &skyline_program.ValidatorSet{}
	err = vs.Unmarshal(vsInfo.GetBinary()[8:])
	require.NoError(t, err)

	require.Equal(t, vs.Signers, validators)
	require.Equal(t, vs.Threshold, uint8(3))
	require.Equal(t, vs.LastBatchId, uint64(0))
	require.Equal(t, vs.BridgeRequestCount, uint64(0))

	mint, err := cli.CreateTokenAccount(feePayer, vaultPda)
	require.NoError(t, err)

	feePayerAta, _, err := solana.FindAssociatedTokenAddress(feePayer.PublicKey(), *mint)
	require.NoError(t, err)

	vaultAta, _, err := solana.FindAssociatedTokenAddress(vaultPda, *mint)
	require.NoError(t, err)

	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, 1)

	bridgingTransactionPda, _, err := solana.FindProgramAddress([][]byte{skyline_program.BRIDGING_TRANSACTION_SEED, buf}, programKeypair.PublicKey())
	require.NoError(t, err)

	bridgeTxIx, err := skyline_program.NewBridgeTransactionInstruction(
		solana.LAMPORTS_PER_SOL,
		1,
		feePayer.PublicKey(),
		vsPda,
		bridgingTransactionPda,
		*mint,
		feePayer.PublicKey(),
		feePayerAta,
		vaultPda,
		vaultAta,
		solana.TokenProgramID,
		solana.SystemProgramID,
		solana.SPLAssociatedTokenAccountProgramID,
	)
	require.NoError(t, err)

	accounts := make([]*solana.AccountMeta, 4)
	for i := range numValidators {
		accounts[i] = solana.NewAccountMeta(validators[i], false, true)
	}

	signers := make(map[solana.PublicKey]*solana.PrivateKey, 4)
	for i := range numValidators {
		signers[validators[i]] = &validatorsPks[i]
	}

	_, err = cli.ExecuteInstructionWithAccounts(bridgeTxIx, accounts, signers, feePayer)
	require.NoError(t, err)

	res, err := cli.GetRpcClient().GetTokenAccountBalance(context.Background(), feePayerAta, rpc.CommitmentFinalized)
	require.NoError(t, err)

	balance, err := strconv.ParseUint(res.Value.Amount, 10, 64)
	require.NoError(t, err)
	require.Equal(t, balance, solana.LAMPORTS_PER_SOL)

	time.Sleep(10 * time.Second)
}
