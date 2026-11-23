package clideploysolana

import (
	"context"
	"encoding/binary"
	"fmt"

	"github.com/Ethernal-Tech/apex-bridge/common"
	solsendtx "github.com/Ethernal-Tech/solana-infrastructure/sendtx"
	"github.com/Ethernal-Tech/solana-infrastructure/sendtx/skyline_program"
	solanawallet "github.com/Ethernal-Tech/solana-infrastructure/wallet"
	"github.com/gagliardetto/solana-go"
	"github.com/spf13/cobra"
)

type renameMintBurnTokenParams struct {
	programAdminParams

	tokenID     uint16
	tokenName   string
	tokenSymbol string
}

func (p *renameMintBurnTokenParams) setFlags(cmd *cobra.Command) {
	p.setCommonFlagsOnCommand(cmd)
	cmd.Flags().Uint16Var(&p.tokenID, registerMintBurnTokenIDFlag, 0, registerMintBurnTokenIDFlagDesc)
	cmd.Flags().StringVar(&p.tokenName, registerMintBurnTokenNameFlag, "", "new on-chain token name")
	cmd.Flags().StringVar(&p.tokenSymbol, registerMintBurnTokenSymbolFlag, "", "new on-chain token symbol")
}

func (p *renameMintBurnTokenParams) validateFlags() error {
	if err := p.validateCommonFlags(); err != nil {
		return err
	}

	if p.tokenID == 0 {
		return fmt.Errorf("token ID must be greater than 0: --%s", registerMintBurnTokenIDFlag)
	}

	return validateTokenMetadata(
		p.tokenName, registerMintBurnTokenNameFlag, p.tokenSymbol, registerMintBurnTokenSymbolFlag, "", "",
	)
}

func (p *renameMintBurnTokenParams) Execute(outputter common.OutputFormatter) (common.ICommandResult, error) {
	provider, err := solanawallet.NewProvider(p.rpcURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create provider: %w", err)
	}

	return p.renameToken(context.Background(), provider, outputter)
}

func (p *renameMintBurnTokenParams) renameToken(
	ctx context.Context, provider solanawallet.ITxProvider, outputter common.OutputFormatter,
) (common.ICommandResult, error) {
	tokenIDBytes := binary.LittleEndian.AppendUint16(nil, p.tokenID)

	registry, _, err := solana.FindProgramAddress(
		[][]byte{skyline_program.TOKEN_REGISTRY_SEED, tokenIDBytes}, p.programPublicKey,
	)
	if err != nil {
		return nil, fmt.Errorf("derive token registry: %w", err)
	}

	info, err := provider.GetAccountInfo(ctx, registry)
	if err != nil {
		return nil, fmt.Errorf("get token registry for ID %d: %w", p.tokenID, err)
	}

	if info == nil || info.Value == nil || info.Value.Data == nil {
		return nil, fmt.Errorf("token registry for ID %d is missing", p.tokenID)
	}

	if info.Value.Owner != p.programPublicKey {
		return nil, fmt.Errorf("token registry %s is not owned by program %s", registry, p.programPublicKey)
	}

	token, err := skyline_program.ParseAccount_TokenRegistry(info.Value.Data.GetBinary())
	if err != nil {
		return nil, fmt.Errorf("decode token registry: %w", err)
	}

	if token.TokenId != p.tokenID || token.IsLockUnlock {
		return nil, fmt.Errorf("token ID %d does not identify a registered mint/burn token", p.tokenID)
	}

	ix, err := p.renameInstruction(registry, token.Mint)
	if err != nil {
		return nil, err
	}

	blockhash, err := provider.GetLatestBlockhash(ctx)
	if err != nil {
		return nil, fmt.Errorf("get latest blockhash: %w", err)
	}

	tx, err := solana.NewTransaction(
		[]solana.Instruction{ix}, blockhash, solana.TransactionPayer(p.adminPrivateKey.PublicKey()),
	)
	if err != nil {
		return nil, fmt.Errorf("create rename mint/burn token tx: %w", err)
	}

	_, _ = outputter.Write([]byte("Submitting rename-mint-burn-token transaction..."))
	outputter.WriteOutput()

	txSender := solsendtx.NewTxSender(provider, nil)

	sig, err := sendSignedTransaction(ctx, provider, txSender, tx, p.confirmationTimeout, p.adminPrivateKey)
	if err != nil {
		return nil, fmt.Errorf("send rename mint/burn token tx: %w", err)
	}

	return &deployProgramResult{
		Output:      fmt.Sprintf("renamed mint/burn token ID %d to %q (%s)", p.tokenID, p.tokenName, p.tokenSymbol),
		MintAddress: token.Mint.String(),
		TxSignature: sig.String(),
	}, nil
}

func (p *renameMintBurnTokenParams) renameInstruction(registry, mint solana.PublicKey) (solana.Instruction, error) {
	feeConfig, _, err := solana.FindProgramAddress([][]byte{skyline_program.FEE_CONFIG_SEED}, p.programPublicKey)
	if err != nil {
		return nil, fmt.Errorf("derive fee config: %w", err)
	}

	vault, _, err := solana.FindProgramAddress([][]byte{skyline_program.VAULT_SEED}, p.programPublicKey)
	if err != nil {
		return nil, fmt.Errorf("derive vault: %w", err)
	}

	metadata, _, err := solana.FindTokenMetadataAddress(mint)
	if err != nil {
		return nil, fmt.Errorf("derive token metadata: %w", err)
	}

	ix, err := skyline_program.NewRenameMintBurnTokenInstruction(
		p.tokenID, p.tokenName, p.tokenSymbol, p.adminPrivateKey.PublicKey(), feeConfig, registry, mint,
		vault, metadata, solana.TokenMetadataProgramID, solana.SystemProgramID, solana.SysVarInstructionsPubkey,
	)
	if err != nil {
		return nil, fmt.Errorf("build rename mint/burn token instruction: %w", err)
	}

	// The generated binding targets its default deployment. Honor --program-id.
	data, err := ix.Data()
	if err != nil {
		return nil, fmt.Errorf("encode rename mint/burn token instruction: %w", err)
	}

	return solana.NewInstruction(p.programPublicKey, ix.Accounts(), data), nil
}
