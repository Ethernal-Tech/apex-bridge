package client

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/gagliardetto/solana-go"
	associatedtokenaccount "github.com/gagliardetto/solana-go/programs/associated-token-account"
	"github.com/gagliardetto/solana-go/programs/system"
	"github.com/gagliardetto/solana-go/programs/token"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/gagliardetto/solana-go/rpc/ws"
)

// SolanaClient wraps RPC and WebSocket clients for Solana blockchain interactions.
// It provides methods to execute instructions and manage transactions with
// automatic signature subscription and confirmation.
type SolanaClient struct {
	cli        *rpc.Client
	wsCli      *ws.Client
	commitment rpc.CommitmentType
}

// NewSolanaClient creates a new SolanaClient instance with the provided RPC and WebSocket clients.
// The client is initialized with CommitmentFinalized as the default commitment level.
func NewSolanaClient(cli *rpc.Client, wsCli *ws.Client) *SolanaClient {
	return &SolanaClient{
		cli:        cli,
		wsCli:      wsCli,
		commitment: rpc.CommitmentFinalized,
	}
}

// Close closes the underlying RPC and WebSocket clients.
// It safely handles nil clients and should be called when the SolanaClient is no longer needed
// to properly release network resources.
func (s *SolanaClient) Close() {
	if s.wsCli != nil {
		s.wsCli.Close()
	}
	if s.cli != nil {
		s.cli.Close()
	}
}

// ExecuteInstruction builds, signs, and sends a transaction containing a single instruction.
// It waits for the transaction to be confirmed using WebSocket subscription.
//
// Parameters:
//   - ix: The instruction to execute
//   - signers: Map of public keys to their corresponding private keys for signing
//   - feePayer: The public key of the account that will pay for the transaction fees
//
// Returns the transaction signature on success, or an error if any step fails.
func (s *SolanaClient) ExecuteInstruction(
	ix *solana.Instruction,
	signers map[solana.PublicKey]*solana.PrivateKey,
	feePayer solana.PrivateKey,
) (*solana.Signature, error) {
	blockhash, err := s.cli.GetLatestBlockhash(context.TODO(), rpc.CommitmentFinalized)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest blockhash: %w", err)
	}

	tx, err := solana.NewTransactionBuilder().
		SetRecentBlockHash(blockhash.Value.Blockhash).
		SetFeePayer(feePayer.PublicKey()).
		AddInstruction(*ix).
		Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build transaction: %w", err)
	}

	signers[feePayer.PublicKey()] = &feePayer

	_, err = tx.Sign(func(key solana.PublicKey) *solana.PrivateKey {
		return signers[key]
	})
	if err != nil {
		return nil, fmt.Errorf("failed to sign transaction: %w", err)
	}

	sig, err := s.cli.SendTransaction(context.TODO(), tx)
	if err != nil {
		return nil, fmt.Errorf("failed to send transaction: %w", err)
	}

	sub, err := s.wsCli.SignatureSubscribe(sig, s.commitment)
	if err != nil {
		return nil, fmt.Errorf("failed to subscribe to signature: %w", err)
	}
	defer sub.Unsubscribe()

	result := <-sub.Response()
	if result.Value.Err != nil {
		return nil, fmt.Errorf("send tx failed: %v", result.Value.Err)
	}

	return &sig, nil
}

// ExecuteInstructionWithAccounts executes an instruction with additional account metadata.
// This method merges the instruction's existing accounts with the provided additional accounts
// before building and sending the transaction.
//
// Parameters:
//   - ix: The base instruction to execute
//   - accounts: Additional account metadata to append to the instruction's accounts
//   - signers: Map of public keys to their corresponding private keys for signing
//   - feePayer: The public key of the account that will pay for the transaction fees
//
// Returns the transaction signature on success, or an error if any step fails.
func (s *SolanaClient) ExecuteInstructionWithAccounts(
	ix solana.Instruction,
	accounts []*solana.AccountMeta,
	signers map[solana.PublicKey]*solana.PrivateKey,
	feePayer solana.PrivateKey,
) (*solana.Signature, error) {
	blockhash, err := s.cli.GetLatestBlockhash(context.TODO(), rpc.CommitmentFinalized)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest blockhash: %w", err)
	}

	ixAccounts := append(ix.Accounts(), accounts...)

	data, err := ix.Data()
	if err != nil {
		return nil, fmt.Errorf("failed to get instruction data: %w", err)
	}

	tx, err := solana.NewTransactionBuilder().
		SetRecentBlockHash(blockhash.Value.Blockhash).
		SetFeePayer(feePayer.PublicKey()).
		AddInstruction(solana.NewInstruction(ix.ProgramID(), ixAccounts, data)).
		Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build transaction: %w", err)
	}

	signers[feePayer.PublicKey()] = &feePayer

	_, err = tx.Sign(func(key solana.PublicKey) *solana.PrivateKey {
		return signers[key]
	})
	if err != nil {
		return nil, fmt.Errorf("failed to sign transaction: %w", err)
	}

	sig, err := s.cli.SendTransaction(context.TODO(), tx)
	if err != nil {
		return nil, fmt.Errorf("failed to send transaction: %w", err)
	}

	sub, err := s.wsCli.SignatureSubscribe(sig, s.commitment)
	if err != nil {
		return nil, fmt.Errorf("failed to subscribe to signature: %w", err)
	}
	defer sub.Unsubscribe()

	result := <-sub.Response()
	if result.Value.Err != nil {
		return nil, fmt.Errorf("send tx failed: %v", result.Value.Err)
	}

	return &sig, nil
}

// CreateInstructionWithAccounts creates a new instruction by merging the provided instruction's
// accounts with additional account metadata. This is a utility method that does not execute
// the instruction, only constructs it.
//
// Parameters:
//   - ix: The base instruction
//   - accounts: Additional account metadata to append to the instruction's accounts
//
// Returns a new instruction with merged accounts, or an error if the instruction data cannot be retrieved.
func (s *SolanaClient) CreateInstructionWithAccounts(
	ix solana.Instruction,
	accounts []*solana.AccountMeta,
) (solana.Instruction, error) {
	data, err := ix.Data()
	if err != nil {
		return nil, fmt.Errorf("failed to get instruction data: %w", err)
	}

	return solana.NewInstruction(ix.ProgramID(), append(ix.Accounts(), accounts...), data), nil
}

// ExecuteMultipleInstructions builds, signs, and sends a transaction containing multiple instructions.
// All instructions are included in a single transaction and executed atomically.
// It waits for the transaction to be confirmed using WebSocket subscription.
//
// Parameters:
//   - ixs: Slice of instructions to execute in the transaction
//   - accounts: Additional account metadata (currently unused but kept for API consistency)
//   - signers: Map of public keys to their corresponding private keys for signing
//   - feePayer: The public key of the account that will pay for the transaction fees
//
// Returns the transaction signature on success, or an error if any step fails.
func (s *SolanaClient) ExecuteMultipleInstructions(
	ixs []solana.Instruction,
	accounts []*solana.AccountMeta,
	signers map[solana.PublicKey]*solana.PrivateKey,
	feePayer solana.PrivateKey,
) (*solana.Signature, error) {
	blockhash, err := s.cli.GetLatestBlockhash(context.TODO(), rpc.CommitmentFinalized)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest blockhash: %w", err)
	}

	builder := solana.NewTransactionBuilder().
		SetRecentBlockHash(blockhash.Value.Blockhash).
		SetFeePayer(feePayer.PublicKey())

	for _, ix := range ixs {
		builder.AddInstruction(ix)
	}

	tx, err := builder.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build transaction: %w", err)
	}

	signers[feePayer.PublicKey()] = &feePayer

	_, err = tx.Sign(func(key solana.PublicKey) *solana.PrivateKey {
		return signers[key]
	})
	if err != nil {
		return nil, fmt.Errorf("failed to sign transaction: %w", err)
	}

	sig, err := s.cli.SendTransaction(context.TODO(), tx)
	if err != nil {
		return nil, fmt.Errorf("failed to send transaction: %w", err)
	}

	sub, err := s.wsCli.SignatureSubscribe(sig, s.commitment)
	if err != nil {
		return nil, fmt.Errorf("failed to subscribe to signature: %w", err)
	}
	defer sub.Unsubscribe()

	result := <-sub.Response()
	if result.Value.Err != nil {
		return nil, fmt.Errorf("send tx failed: %v", result.Value.Err)
	}

	return &sig, nil
}

// Airdrop requests an airdrop of SOL tokens to the specified address.
// This is typically used for testing and development on local networks.
//
// Parameters:
//   - addr: The public key of the account to receive the airdrop
//   - amount: The amount of lamports to airdrop (1 SOL = 1,000,000,000 lamports)
//
// Returns an error if the airdrop request fails.
func (s *SolanaClient) Airdrop(addr solana.PublicKey, amount uint64) error {
	_, err := s.cli.RequestAirdrop(context.TODO(), addr, amount, rpc.CommitmentFinalized)
	if err != nil {
		return fmt.Errorf("failed to request airdrop: %w", err)
	}

	return nil
}

// Deploy deploys a Solana program to the localhost network using the solana CLI.
// This method executes the "solana program deploy" command and waits for the deployment
// to complete, including a 20-second delay to ensure the program is fully available.
//
// Parameters:
//   - feePayer: Path to the fee payer keypair file
//   - programKey: Path to the program keypair file
//   - buildPath: Path to the compiled program (.so file) to deploy
//
// Returns an error if the deployment command fails to start or complete.
// Note: This method requires the solana CLI to be installed and available in PATH.
func (s *SolanaClient) Deploy(feePayer string, programKey string, buildPath string) error {
	cmd := exec.Command("solana",
		"program", "deploy",
		"-u", "localhost",
		"--fee-payer", feePayer,
		"-k", programKey,
		buildPath,
		"--commitment", string(rpc.CommitmentFinalized),
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to deploy: %w", err)
	}

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("failed to deploy: %w", err)
	}

	time.Sleep(20 * time.Second)
	return nil
}

// CreateTokenAccount creates a new SPL token mint account and initializes it.
// This method generates a new keypair for the mint, creates the account, and initializes
// it with the specified mint authority. The mint is initialized with 9 decimals.
//
// Parameters:
//   - pk: The private key of the account that will pay for the account creation and initialization
//   - mintAuthority: The public key that will have authority over the mint (can mint tokens)
//
// Returns the public key of the newly created mint account, or an error if any step fails.
func (s *SolanaClient) CreateTokenAccount(pk solana.PrivateKey, mintAuthority solana.PublicKey) (*solana.PublicKey, error) {
	tokenPk, err := solana.NewRandomPrivateKey()
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	block, err := s.cli.GetLatestBlockhash(context.TODO(), rpc.CommitmentFinalized)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	rent, err := s.cli.GetMinimumBalanceForRentExemption(context.Background(), token.MINT_SIZE, rpc.CommitmentFinalized)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	caIx := system.NewCreateAccountInstruction(rent+1, uint64(token.MINT_SIZE), token.ProgramID, pk.PublicKey(), tokenPk.PublicKey()).Build()

	mintIx := token.NewInitializeMint2Instruction(9, mintAuthority, mintAuthority, tokenPk.PublicKey())
	mintTx, err := solana.NewTransactionBuilder().AddInstruction(caIx).AddInstruction(mintIx.Build()).SetFeePayer(pk.PublicKey()).SetRecentBlockHash(block.Value.Blockhash).Build()
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	_, err = mintTx.Sign(func(key solana.PublicKey) *solana.PrivateKey {
		if key.Equals(pk.PublicKey()) {
			return &pk
		}
		if key.Equals(tokenPk.PublicKey()) {
			return &tokenPk
		}
		return nil
	})
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	sigMint, err := s.cli.SendTransaction(context.TODO(), mintTx)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	subMint, err := s.wsCli.SignatureSubscribe(sigMint, rpc.CommitmentFinalized)
	if err != nil {
		fmt.Println("Subscription error:", err)
		return nil, err
	}

	rd := <-subMint.Response()
	if rd.Value.Err != nil {
		fmt.Println("Transaction failed:", rd.Value.Err)
		return nil, err
	}

	ret := tokenPk.PublicKey()
	return &ret, nil
}

// MintToAccount mints tokens to a receiver's associated token account (ATA).
// If the receiver's ATA does not exist, it will be created automatically before minting.
// The method mints 1,000,000,000 base units (1 token with 9 decimals) to the receiver.
//
// Parameters:
//   - pk: The private key of the mint authority (must have permission to mint tokens)
//   - receiver: The public key of the account that will receive the minted tokens
//   - mint: The public key of the token mint
//
// Returns the public key of the receiver's associated token account and an error if any step fails.
func (s *SolanaClient) MintToAccount(pk solana.PrivateKey, receiver solana.PublicKey, mint solana.PublicKey) (ata solana.PublicKey, err error) {
	ata, _, err = solana.FindAssociatedTokenAddress(receiver, mint)
	if err != nil {
		return
	}

	var instructions []solana.Instruction

	ataInfo, err := s.cli.GetAccountInfo(context.Background(), ata)
	if err != nil || ataInfo.Value == nil {
		ataIx := associatedtokenaccount.NewCreateInstruction(
			pk.PublicKey(),
			receiver,
			mint,
		).Build()
		instructions = append(instructions, ataIx)
	}

	mintToIx := token.NewMintToInstruction(1_000_000_000, mint, ata, pk.PublicKey(), []solana.PublicKey{}).Build()
	instructions = append(instructions, mintToIx)

	blockhash, err := s.cli.GetLatestBlockhash(context.TODO(), rpc.CommitmentFinalized)
	if err != nil {
		return
	}

	builder := solana.NewTransactionBuilder().SetRecentBlockHash(blockhash.Value.Blockhash).SetFeePayer(pk.PublicKey())
	for _, ix := range instructions {
		builder.AddInstruction(ix)
	}

	tx, err := builder.Build()
	if err != nil {
		return
	}

	_, err = tx.Sign(func(key solana.PublicKey) *solana.PrivateKey {
		if key.Equals(pk.PublicKey()) {
			return &pk
		}
		return nil
	})
	if err != nil {
		return
	}

	sig, err := s.cli.SendTransaction(context.TODO(), tx)
	if err != nil {
		return
	}

	sub, err := s.wsCli.SignatureSubscribe(sig, rpc.CommitmentFinalized)
	if err != nil {
		return
	}

	result := <-sub.Response()
	if result.Value.Err != nil {
		err = fmt.Errorf("mint to account transaction failed: %v", result.Value.Err)
		return
	}

	return
}
