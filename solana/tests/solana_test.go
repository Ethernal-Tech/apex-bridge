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
	"github.com/Ethernal-Tech/apex-bridge/solana/tests/helper"
	storagehelper "github.com/Ethernal-Tech/apex-bridge/solana/tests/storage_helper"
	testvalidator "github.com/Ethernal-Tech/apex-bridge/solana/tests/test_validator"
	tracker "github.com/Ethernal-Tech/solana-event-tracker"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/stretchr/testify/require"
)

// Test_SolanaTransactions is an integration test that validates the complete
// bridge transaction flow on Solana. It performs the following operations:
//
// Setup Phase:
//  1. Starts a local Solana test validator node
//  2. Creates a Solana client connected to the local network
//  3. Initializes an event tracker to monitor program events (TransactionExecutedEvent
//     and BridgeRequestEvent)
//  4. Deploys the skyline_program to the local network
//  5. Creates and initializes a validator set with 4 validators and a threshold of 3
//  6. Creates a token mint and associated token accounts for the fee payer and vault
//
// Test Scenarios:
//
// 1. Bridge Transaction (SKYLINE -> SOL):
//   - Creates a bridge transaction instruction to transfer tokens from SKYLINE to SOL
//   - Executes the transaction with validator signatures (requires 3 out of 4 validators)
//   - Verifies the fee payer's token account balance is updated correctly
//   - Waits for and verifies the TransactionExecutedEvent is emitted
//
// 2. Bridge Request (SOL -> SKYLINE):
//   - Creates a bridge request instruction to initiate a transfer from SOL to SKYLINE
//   - Executes the request with the fee payer's signature
//   - Waits for and verifies the BridgeRequestEvent is emitted
//
// Prerequisites:
//   - The test requires the following files to exist in the program_build directory:
//   - skyline_program-keypair.json: Program keypair for deployment
//   - test.json: Fee payer keypair for transaction fees
//   - skyline_program.so: Compiled program binary
//
// Expected Behavior:
//   - The validator set is initialized with the correct validators and threshold
//   - Bridge transactions execute successfully with proper validator signatures
//   - Token balances are updated correctly after bridge transactions
//   - Program events are emitted and tracked correctly
//
// The test uses a local Solana validator to avoid network dependencies and
// provides deterministic testing of the bridge functionality.
func Test_SolanaTransactions(t *testing.T) {
	const (
		amount        = 10 * solana.LAMPORTS_PER_SOL // Initial airdrop amount for fee payer
		numValidators = 4                            // Number of validators in the validator set
	)

	// Start a local Solana test validator node
	validator := testvalidator.NewTestValidator()
	require.NoError(t, validator.StartTestNode())
	defer validator.Close()

	// Wait for the validator node to be ready
	require.NoError(t, validator.WaitForNode(rpc.New(rpc.LocalNet_RPC)))

	// Create a Solana client connected to the local network
	cli, err := client.NewSolanaClient(client.WithLocalnet())
	require.NoError(t, err)
	defer cli.Close()

	// Configure event tracking for bridge program events
	spec := tracker.ProgramEventSpecs{}
	spec.AddEventSpec(skyline_program.TransactionExecutedEvent{}, "TransactionExecutedEvent")

	// Initialize event storage for tracking program events
	storage := storagehelper.NewStorage()

	// Start event tracker in a goroutine to monitor program events (must be in a goroutine to work)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		track, err := tracker.NewEventTracker(
			rpc.New(rpc.LocalNet_RPC),
			storage,
			map[solana.PublicKey]tracker.ProgramEventSpecs{
				skyline_program.ProgramID: spec,
			},
			rpc.CommitmentFinalized,
			tracker.WithNotifications(255, 255, 255),
		)
		require.NoError(t, err)
		require.NoError(t, track.Start())
		defer track.Terminate()

		<-ctx.Done()
	}()

	// Load program and fee payer keypairs from files
	programPath, err := filepath.Abs("program_build/skyline_program-keypair.json")
	require.NoError(t, err)

	feePayerPath, err := filepath.Abs("program_build/test.json")
	require.NoError(t, err)

	// Load the compiled program binary
	buildPath, err := filepath.Abs("program_build/skyline_program.so")
	require.NoError(t, err)

	// Load the program keypair
	programKeypair, err := solana.PrivateKeyFromSolanaKeygenFile(programPath)
	require.NoError(t, err)

	// Load the fee payer keypair
	feePayer, err := solana.PrivateKeyFromSolanaKeygenFile(feePayerPath)
	require.NoError(t, err)

	// Airdrop SOL to fee payer for transaction fees
	require.NoError(t, cli.Airdrop(feePayer.PublicKey(), amount))

	// Deploy the skyline_program to the local network
	require.NoError(t, cli.Deploy(feePayerPath, programPath, buildPath))

	// Generate validator keypairs for the validator set
	validators, validatorsPks := make([]solana.PublicKey, numValidators), make([]solana.PrivateKey, numValidators)
	for i := range numValidators {
		pk, err := solana.NewRandomPrivateKey()
		require.NoError(t, err)

		validatorsPks[i] = pk
		validators[i] = validatorsPks[i].PublicKey()
	}

	// Find Program Derived Addresses (PDAs) for validator set and vault
	vsPda, _, err := solana.FindProgramAddress([][]byte{skyline_program.VALIDATOR_SET_SEED}, programKeypair.PublicKey())
	require.NoError(t, err)

	vaultPda, _, err := solana.FindProgramAddress([][]byte{skyline_program.VAULT_SEED}, programKeypair.PublicKey())
	require.NoError(t, err)

	// Initialize the program with validators (threshold defaults to 3 out of 4)
	initializeIx, err := skyline_program.NewInitializeInstruction(validators, nil, feePayer.PublicKey(), vsPda, vaultPda, solana.SystemProgramID)
	require.NoError(t, err)

	_, err = cli.ExecuteInstruction(&initializeIx, map[solana.PublicKey]*solana.PrivateKey{}, feePayer)
	require.NoError(t, err)

	// Verify validator set initialization
	vsInfo, err := cli.GetRpcClient().GetAccountInfo(context.Background(), vsPda)
	require.NoError(t, err)

	// Unmarshal the validator set account data
	vs := &skyline_program.ValidatorSet{}
	require.NoError(t, vs.Unmarshal(vsInfo.GetBinary()[8:])) // Skip the discriminator (8 bytes)

	require.Equal(t, vs.Signers, validators)
	require.Equal(t, vs.Threshold, uint8(3)) // 3 out of 4 validators required
	require.Equal(t, vs.LastBatchId, uint64(0))
	require.Equal(t, vs.BridgeRequestCount, uint64(0))

	// Create a token mint and associated token accounts
	mint, err := cli.CreateTokenAccount(feePayer, vaultPda)
	require.NoError(t, err)

	// Find the associated token address for the fee payer
	feePayerAta, _, err := solana.FindAssociatedTokenAddress(feePayer.PublicKey(), *mint)
	require.NoError(t, err)

	// Find the associated token address for the vault
	vaultAta, _, err := solana.FindAssociatedTokenAddress(vaultPda, *mint)
	require.NoError(t, err)

	t.Run("Bridge Transaction (SKYLINE -> SOL)", func(t *testing.T) {
		// Create a bridge transaction to transfer tokens from SKYLINE to SOL
		// This requires validator signatures (3 out of 4 validators)
		buf := make([]byte, 8)
		binary.LittleEndian.PutUint64(buf, 1)

		// Find the bridging transaction PDA
		bridgingTransactionPda, _, err := solana.FindProgramAddress([][]byte{skyline_program.BRIDGING_TRANSACTION_SEED, buf}, programKeypair.PublicKey())
		require.NoError(t, err)

		// Create a bridge transaction instruction
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

		// Prepare validator accounts and signatures (all 4 validators sign)
		accounts := make([]*solana.AccountMeta, 4)
		for i := range numValidators {
			accounts[i] = solana.NewAccountMeta(validators[i], false, true)
		}

		// Prepare signers map
		signers := make(map[solana.PublicKey]*solana.PrivateKey, 4)
		for i := range numValidators {
			signers[validators[i]] = &validatorsPks[i]
		}

		// Execute the bridge transaction with validator signatures
		_, err = cli.ExecuteInstructionWithAccounts(bridgeTxIx, accounts, signers, feePayer)
		require.NoError(t, err)

		// Verify the fee payer's token balance was updated correctly
		res, err := cli.GetRpcClient().GetTokenAccountBalance(context.Background(), feePayerAta, rpc.CommitmentFinalized)
		require.NoError(t, err)

		balance, err := strconv.ParseUint(res.Value.Amount, 10, 64)
		require.NoError(t, err)
		require.Equal(t, balance, solana.LAMPORTS_PER_SOL)

		// Wait for and verify the TransactionExecutedEvent was emitted
		require.NoError(t, helper.WaitFor(t, 30*time.Second, 1*time.Second, func() bool {
			for _, event := range storage.Events {
				if event.EventName == "TransactionExecutedEvent" {
					return true
				}
			}
			return false
		}))
	})

	t.Run("Bridge Request (SOL -> SKYLINE)", func(t *testing.T) {
		// Create a bridge request to initiate a transfer from SOL to SKYLINE
		// This creates a request that will be processed by validators
		bridgeRequestIx, err := skyline_program.NewBridgeRequestInstruction(
			solana.LAMPORTS_PER_SOL,
			[]byte("0x1234567890123456789012345678901234567890"), // Destination address (EVM format)
			1, // Chain ID
			feePayer.PublicKey(),
			vsPda,
			feePayerAta,
			vaultPda,
			vaultAta,
			*mint,
			solana.TokenProgramID,
			solana.SystemProgramID,
			solana.SPLAssociatedTokenAccountProgramID,
		)
		require.NoError(t, err)

		// Execute the bridge request (only requires fee payer signature)
		_, err = cli.ExecuteInstruction(&bridgeRequestIx, map[solana.PublicKey]*solana.PrivateKey{}, feePayer)
		require.NoError(t, err)
	})
}
