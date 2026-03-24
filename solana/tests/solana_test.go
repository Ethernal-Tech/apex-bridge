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

// bridgeTransactionRemainingAccounts builds Anchor `remaining_accounts` for `bridge_transaction`
// (see programs/skyline-program/src/instructions/bridge_transaction.rs).
func bridgeTransactionRemainingAccounts(
	vaultPda solana.PublicKey,
	validators []solana.PublicKey,
	mints []solana.PublicKey,
	transfers []skyline_program.TransferItem,
) ([]*solana.AccountMeta, error) {
	nVal, nMint, nXfer := len(validators), len(mints), len(transfers)
	out := make([]*solana.AccountMeta, 0, nVal+nMint*3+nXfer*2)

	for _, pk := range validators {
		out = append(out, solana.NewAccountMeta(pk, false, true))
	}

	for _, m := range mints {
		out = append(out, solana.NewAccountMeta(m, false, false))
	}

	for _, tr := range transfers {
		out = append(out, solana.NewAccountMeta(tr.Recipient, false, false))
	}

	for _, m := range mints {
		reg, _, err := solana.FindProgramAddress([][]byte{skyline_program.TOKEN_REGISTRY_SEED, m[:]},
			skyline_program.ProgramID)
		if err != nil {
			return nil, err
		}
		out = append(out, solana.NewAccountMeta(reg, false, false))
	}

	for _, tr := range transfers {
		mintPk := mints[tr.MintIndex]
		ata, _, err := solana.FindAssociatedTokenAddress(tr.Recipient, mintPk)
		if err != nil {
			return nil, err
		}
		out = append(out, solana.NewAccountMeta(ata, true, false))
	}

	for _, m := range mints {
		vaultAta, _, err := solana.FindAssociatedTokenAddress(vaultPda, m)
		if err != nil {
			return nil, err
		}
		out = append(out, solana.NewAccountMeta(vaultAta, true, false))
	}
	
	return out, nil
}

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
//  6. Creates an SPL mint (fee payer as mint authority), registers lock/unlock, funds vault ATA
//
// Test Scenarios:
//
// 1. Bridge Transaction (SKYLINE -> SOL):
//   - Creates a bridge transaction instruction to transfer tokens from SKYLINE to SOL
//   - Executes the transaction with validator signatures (requires 3 out of 4 validators)
//   - Verifies the recipient's token account balance is updated correctly
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
		amount            = 10 * solana.LAMPORTS_PER_SOL // Initial airdrop amount for fee payer
		numValidators     = 4                            // Number of validators in the validator set
		minBridgingFee    = 1 * solana.LAMPORTS_PER_SOL
		minOperationFee   = 1 * solana.LAMPORTS_PER_SOL
		minAmountToBridge = uint64(1)
		tokenID           = uint16(17)
		numReceivers      = 1
	)

	ctx := context.Background()

	// Start a local Solana test validator node
	validator := testvalidator.NewTestValidator()
	require.NoError(t, validator.StartTestNode())
	defer validator.Close()

	// Wait for the validator node to be ready
	require.NoError(t, validator.WaitForNode(rpc.New(rpc.LocalNet_RPC)))

	// Create a Solana client connected to the local network
	cli, err := client.NewSolanaClient(client.WithLocalnet(ctx))
	require.NoError(t, err)
	defer cli.Close()

	// Configure event tracking for bridge program events
	spec := tracker.ProgramEventSpecs{}
	spec.AddEventSpec(skyline_program.TransactionExecutedEvent{}, "TransactionExecutedEvent")
	spec.AddEventSpec(skyline_program.BridgeRequestEvent{}, "BridgeRequestEvent")

	// Initialize event storage for tracking program events
	storage := storagehelper.NewStorage()

	// Start event tracker in a goroutine to monitor program events (must be in a goroutine to work)
	trackerCtx, cancel := context.WithCancel(ctx)
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

		<-trackerCtx.Done()
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

	feeConfigPda, _, err := solana.FindProgramAddress([][]byte{skyline_program.FEE_CONFIG_SEED}, programKeypair.PublicKey())
	require.NoError(t, err)

	treasuryAccountPK, err := solana.NewRandomPrivateKey()
	require.NoError(t, err)

	relayerAccountPK, err := solana.NewRandomPrivateKey()
	require.NoError(t, err)

	// Airdrop SOL to fee payer for transaction fees
	require.NoError(t, cli.Airdrop(ctx, feePayer.PublicKey(), amount))

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

	lastId := uint64(0)

	// Initialize the program with validators (threshold defaults to 3 out of 4)
	initializeIx, err := skyline_program.NewInitializeInstruction(validators, &lastId, minOperationFee, minBridgingFee,
		feePayer.PublicKey(), vsPda, vaultPda, feeConfigPda,
		treasuryAccountPK.PublicKey(), relayerAccountPK.PublicKey(), solana.SystemProgramID)
	require.NoError(t, err)

	_, err = cli.ExecuteInstruction(ctx, &initializeIx, map[solana.PublicKey]*solana.PrivateKey{}, feePayer)
	require.NoError(t, err)

	// Verify validator set initialization
	vsInfo, err := cli.GetRPCClient().GetAccountInfo(ctx, vsPda)
	require.NoError(t, err)

	// Unmarshal the validator set account data
	vs := &skyline_program.ValidatorSet{}
	require.NoError(t, vs.Unmarshal(vsInfo.GetBinary()[8:])) // Skip the discriminator (8 bytes)

	require.Equal(t, vs.Signers, validators)
	require.Equal(t, vs.Threshold, uint8(3)) // 3 out of 4 validators required
	require.Equal(t, vs.LastBatchId, uint64(0))
	require.Equal(t, vs.BridgeRequestCount, uint64(0))

	// Lock/unlock path: fee payer mint authority, register token, pre-fund vault ATA (bridge_transaction transfers out).
	mint, err := cli.CreateTokenAccount(ctx, feePayer, feePayer.PublicKey())
	require.NoError(t, err)

	tokenRegistryPda, _, err := solana.FindProgramAddress(
		[][]byte{skyline_program.TOKEN_REGISTRY_SEED, (*mint)[:]}, skyline_program.ProgramID)
	require.NoError(t, err)

	tokenIDSeed := make([]byte, 2)
	binary.LittleEndian.PutUint16(tokenIDSeed, tokenID)
	tokenIdGuardPda, _, err := solana.FindProgramAddress(
		[][]byte{skyline_program.TOKEN_ID_GUARD_SEED, tokenIDSeed}, skyline_program.ProgramID)
	require.NoError(t, err)

	registerIx, err := skyline_program.NewRegisterLockUnlockTokenInstruction(
		tokenID,
		minAmountToBridge,
		feePayer.PublicKey(),
		feeConfigPda,
		*mint,
		tokenRegistryPda,
		tokenIdGuardPda,
		solana.SystemProgramID,
	)
	require.NoError(t, err)

	_, err = cli.ExecuteInstruction(ctx, &registerIx, map[solana.PublicKey]*solana.PrivateKey{}, feePayer)
	require.NoError(t, err)

	_, err = cli.MintToAccount(ctx, feePayer, vaultPda, *mint, solana.LAMPORTS_PER_SOL)
	require.NoError(t, err)

	feePayerAta, _, err := solana.FindAssociatedTokenAddress(feePayer.PublicKey(), *mint)
	require.NoError(t, err)
	vaultAta, _, err := solana.FindAssociatedTokenAddress(vaultPda, *mint)
	require.NoError(t, err)

	recipient, err := solana.NewRandomPrivateKey()
	require.NoError(t, err)

	transferItems := make([]skyline_program.TransferItem, 0, numReceivers)
	mints := make([]solana.PublicKey, 0, numReceivers)

	bridgeAmount := solana.LAMPORTS_PER_SOL
	transferItem := skyline_program.TransferItem{
		Recipient: recipient.PublicKey(),
		MintIndex: 0,
		Amount:    bridgeAmount,
	}

	transferItems = append(transferItems, transferItem)
	mints = append(mints, *mint)

	t.Run("Bridge Transaction (SKYLINE -> SOL)", func(t *testing.T) {
		// Create a bridge transaction instruction
		bridgeTxIx, err := skyline_program.NewBridgeTransactionInstruction(
			transferItems,
			mints,
			1,
			feePayer.PublicKey(),
			vsPda,
			vaultPda,
			solana.TokenProgramID,
			solana.SystemProgramID,
			solana.SPLAssociatedTokenAccountProgramID,
		)
		require.NoError(t, err)

		remaining, err := bridgeTransactionRemainingAccounts(vaultPda, validators, mints, transferItems)
		require.NoError(t, err)

		// Prepare signers map (validators + fee payer added inside client)
		signers := make(map[solana.PublicKey]*solana.PrivateKey, numValidators)
		for i := range numValidators {
			signers[validators[i]] = &validatorsPks[i]
		}

		// Execute the bridge transaction with validator signatures
		_, err = cli.ExecuteInstructionWithAccounts(ctx, bridgeTxIx, remaining, signers, feePayer)
		require.NoError(t, err)

		recipientAta, _, err := solana.FindAssociatedTokenAddress(recipient.PublicKey(), *mint)
		require.NoError(t, err)

		res, err := cli.GetRPCClient().GetTokenAccountBalance(ctx, recipientAta, rpc.CommitmentFinalized)
		require.NoError(t, err)

		balance, err := strconv.ParseUint(res.Value.Amount, 10, 64)
		require.NoError(t, err)
		require.Equal(t, balance, solana.LAMPORTS_PER_SOL)

		// Wait for and verify the TransactionExecutedEvent was emitted
		require.NoError(t, helper.WaitFor(t, 60*time.Second, 1*time.Second, func() bool {
			for _, event := range storage.Events {
				if event.EventName == "TransactionExecutedEvent" {
					return true
				}
			}
			return false
		}))
	})

	t.Run("Bridge Request (SOL -> SKYLINE)", func(t *testing.T) {
		sendAmount := solana.LAMPORTS_PER_SOL
		// bridge_request charges SOL: relayer gets fee_config.bridge_fee, treasury gets fee - bridge_fee (see on-chain split).
		requestFees := minOperationFee + minBridgingFee

		_, err := cli.MintToAccount(ctx, feePayer, feePayer.PublicKey(), *mint, sendAmount)
		require.NoError(t, err)

		balanceBeforeBridging, err := cli.GetRPCClient().GetTokenAccountBalance(ctx, feePayerAta, rpc.CommitmentFinalized)
		require.NoError(t, err)

		initialBalance, err := strconv.ParseUint(balanceBeforeBridging.Value.Amount, 10, 64)
		require.NoError(t, err)
		require.GreaterOrEqual(t, initialBalance, sendAmount)

		bridgeRequestIx, err := skyline_program.NewBridgeRequestInstruction(
			sendAmount,
			"0x1234567890123456789012345678901234567890",
			"1",
			requestFees,
			feePayer.PublicKey(),
			vsPda,
			feePayerAta,
			vaultPda,
			vaultAta,
			*mint,
			tokenRegistryPda,
			solana.TokenProgramID,
			solana.SystemProgramID,
			solana.SPLAssociatedTokenAccountProgramID,
			feeConfigPda,
			treasuryAccountPK.PublicKey(),
			relayerAccountPK.PublicKey(),
		)
		require.NoError(t, err)

		_, err = cli.ExecuteInstruction(ctx, &bridgeRequestIx, map[solana.PublicKey]*solana.PrivateKey{}, feePayer)
		require.NoError(t, err)

		require.NoError(t, helper.WaitFor(t, 60*time.Second, 1*time.Second, func() bool {
			for _, event := range storage.Events {
				if event.EventName == "BridgeRequestEvent" {
					return true
				}
			}
			return false
		}))

		balanceAfterBridging, err := cli.GetRPCClient().GetTokenAccountBalance(ctx, feePayerAta, rpc.CommitmentFinalized)
		require.NoError(t, err)

		finalTokenBal, err := strconv.ParseUint(balanceAfterBridging.Value.Amount, 10, 64)
		require.NoError(t, err)
		require.Equal(t, initialBalance-sendAmount, finalTokenBal)

		treasuryBalance, err := cli.GetRPCClient().
			GetBalance(ctx, treasuryAccountPK.PublicKey(), rpc.CommitmentFinalized)
		require.NoError(t, err)
		require.Equal(t, uint64(minOperationFee), treasuryBalance.Value)

		relayerBalance, err := cli.GetRPCClient().
			GetBalance(ctx, relayerAccountPK.PublicKey(), rpc.CommitmentFinalized)
		require.NoError(t, err)
		require.Equal(t, uint64(minBridgingFee), relayerBalance.Value)
	})
}
