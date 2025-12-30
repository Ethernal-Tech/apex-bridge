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

type solanaClientOption func(*SolanaClient) error

func WithCommitment(commitment rpc.CommitmentType) solanaClientOption {
	return func(s *SolanaClient) error {
		s.commitment = commitment
		return nil
	}
}

func WithWSClient(wsCli *ws.Client) solanaClientOption {
	return func(s *SolanaClient) error {
		s.wsCli = wsCli
		return nil
	}
}

func WithRPCClient(cli *rpc.Client) solanaClientOption {
	return func(s *SolanaClient) error {
		s.cli = cli
		return nil
	}
}

func WithDevnet() solanaClientOption {
	return func(s *SolanaClient) error {
		s.cli = rpc.New(rpc.DevNet_RPC)
		wsCli, err := ws.Connect(context.Background(), rpc.DevNet_WS)
		if err != nil {
			return fmt.Errorf("failed to connect to devnet: %w", err)
		}

		s.wsCli = wsCli
		return nil
	}
}

func WithLocalnet() solanaClientOption {
	return func(s *SolanaClient) error {
		s.cli = rpc.New(rpc.LocalNet_RPC)
		wsCli, err := ws.Connect(context.Background(), rpc.LocalNet_WS)
		if err != nil {
			return fmt.Errorf("failed to connect to localnet: %w", err)
		}

		s.wsCli = wsCli
		return nil
	}
}

func WithMainnet() solanaClientOption {
	return func(s *SolanaClient) error {
		s.cli = rpc.New(rpc.MainNetBetaSerum_RPC)
		wsCli, err := ws.Connect(context.Background(), rpc.MainNetBetaSerum_WS)
		if err != nil {
			return fmt.Errorf("failed to connect to mainnet: %w", err)
		}

		s.wsCli = wsCli
		return nil
	}
}

func WithCustomRPC(rpcUrl string) solanaClientOption {
	return func(s *SolanaClient) error {
		s.cli = rpc.New(rpcUrl)
		return nil
	}
}

// NewSolanaClient creates a new SolanaClient instance with the provided RPC and WebSocket clients.
// The client is initialized with CommitmentFinalized as the default commitment level.
func NewSolanaClient(opts ...solanaClientOption) (*SolanaClient, error) {
	s := &SolanaClient{}

	for _, opt := range opts {
		if err := opt(s); err != nil {
			return nil, err
		}
	}

	return s, nil
}

func (s *SolanaClient) GetRpcClient() *rpc.Client {
	return s.cli
}

func (s *SolanaClient) GetWsClient() *ws.Client {
	return s.wsCli
}

func (s *SolanaClient) Close() {
	if s.wsCli != nil {
		s.wsCli.Close()
	}
	if s.cli != nil {
		if err := s.cli.Close(); err != nil {
			fmt.Println("Error while closing RPC", err)
		}
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

	if err = s.waitForSignature(sig, rpc.CommitmentFinalized); err != nil {
		return nil, err
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

	if err = s.waitForSignature(sig, rpc.CommitmentFinalized); err != nil {
		return nil, err
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

	if err = s.waitForSignature(sig, rpc.CommitmentFinalized); err != nil {
		return nil, err
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
	sig, err := s.cli.RequestAirdrop(context.TODO(), addr, amount, rpc.CommitmentFinalized)
	if err != nil {
		return fmt.Errorf("failed to request airdrop: %w", err)
	}

	if err = s.waitForSignature(sig, rpc.CommitmentFinalized); err != nil {
		return err
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

	err = s.waitForSignature(sigMint, rpc.CommitmentFinalized)
	if err != nil {
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
func (s *SolanaClient) MintToAccount(pk solana.PrivateKey, receiver solana.PublicKey, mint solana.PublicKey, amount uint64) (ata solana.PublicKey, err error) {
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

	mintToIx := token.NewMintToInstruction(amount, mint, ata, pk.PublicKey(), []solana.PublicKey{}).Build()
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

	err = s.waitForSignature(sig, rpc.CommitmentFinalized)
	return
}

func (s *SolanaClient) waitForSignature(sig solana.Signature, commitment rpc.CommitmentType) error {

	sub, err := s.wsCli.SignatureSubscribe(sig, commitment)
	if err != nil {
		return err
	}
	defer sub.Unsubscribe()

	rd := <-sub.Response()
	if rd.Value.Err != nil {
		return fmt.Errorf("transaction failed: %v", rd.Value.Err)
	}

	return nil
}
