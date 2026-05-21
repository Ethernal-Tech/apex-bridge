package clideploysolana

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/Ethernal-Tech/apex-bridge/common"
	solsendtx "github.com/Ethernal-Tech/solana-infrastructure/sendtx"
	solanawallet "github.com/Ethernal-Tech/solana-infrastructure/wallet"
	"github.com/gagliardetto/solana-go"
	"github.com/spf13/cobra"
)

const (
	extendALTAddressFlag            = "alt-address"
	extendALTTokenIDAndMintFlag     = "token-id-and-mint"
	extendALTAddressFlagDesc        = "ALT public key to extend"
	extendALTTokenIDAndMintFlagDesc = "token_id:token mint public key (repeat for multiple mints)" //nolint:gosec
)

type extendALTParams struct {
	altCommonParams

	altAddress string
	tokenMints []string

	altPublicKey     solana.PublicKey
	tokenMintsParsed map[uint16]solana.PublicKey
}

func (p *extendALTParams) setFlags(cmd *cobra.Command) {
	p.setCommonFlags(cmd.Flags())

	cmd.Flags().StringVar(
		&p.altAddress,
		extendALTAddressFlag,
		"",
		extendALTAddressFlagDesc,
	)
	cmd.Flags().StringArrayVar(
		&p.tokenMints,
		extendALTTokenIDAndMintFlag,
		nil,
		extendALTTokenIDAndMintFlagDesc,
	)
}

func (p *extendALTParams) validateFlags() error {
	if err := p.validateCommonFlags(); err != nil {
		return err
	}

	if p.altAddress == "" {
		return fmt.Errorf("ALT address not specified: --%s", extendALTAddressFlag)
	}

	altPublicKey, err := solanawallet.PublicKeyFromAddress(p.altAddress)
	if err != nil {
		return fmt.Errorf("invalid ALT address: %w", err)
	}

	p.altPublicKey = altPublicKey

	tokenMintsParsed := make(map[uint16]solana.PublicKey)

	for i, tokenMint := range p.tokenMints {
		tokenMint = strings.TrimSpace(tokenMint)
		if tokenMint == "" {
			return fmt.Errorf("token mint %d is empty", i)
		}

		tokenIDAndMint := strings.Split(tokenMint, ":")
		if len(tokenIDAndMint) != 2 {
			return fmt.Errorf("invalid token ID and mint %s", tokenMint)
		}

		tokenID, err := strconv.ParseUint(tokenIDAndMint[0], 10, 16)
		if err != nil {
			return fmt.Errorf("invalid token ID %s: %w", tokenIDAndMint[0], err)
		}

		tokenMintPubKey, err := solanawallet.PublicKeyFromAddress(tokenIDAndMint[1])
		if err != nil {
			return fmt.Errorf("invalid token mint %s: %w", tokenMint, err)
		}

		tokenMintsParsed[uint16(tokenID)] = tokenMintPubKey
	}

	p.tokenMintsParsed = tokenMintsParsed

	return nil
}

func (p *extendALTParams) Execute(outputter common.OutputFormatter) (common.ICommandResult, error) {
	ctx := context.Background()

	provider, err := solanawallet.NewProvider(p.rpcURL)
	if err != nil {
		return nil, fmt.Errorf("create provider: %w", err)
	}

	txSender := solsendtx.NewTxSender(provider, nil)
	altAdmin := solanawallet.NewALTAdmin(provider)

	altPubKeys, err := txSender.BridgeTransactionALTAddresses(p.programPublicKey, p.tokenMintsParsed)
	if err != nil {
		return nil, fmt.Errorf("get bridge transaction ALT addresses: %w", err)
	}

	ixs, err := altAdmin.NewExtendInstructions(
		ctx,
		p.altPublicKey,
		p.adminPrivateKey.PublicKey(),
		p.adminPrivateKey.PublicKey(),
		altPubKeys,
	)
	if err != nil {
		return nil, fmt.Errorf("new extend instructions: %w", err)
	}

	if len(ixs) == 0 {
		return &deployProgramResult{
			Output: fmt.Sprintf("ALT already up to date: %s", p.altPublicKey.String()),
		}, nil
	}

	_, _ = outputter.Write([]byte(fmt.Sprintf("Submitting extend-alt transaction (%d instructions)...", len(ixs))))
	outputter.WriteOutput()

	sig, err := sendALTInstructions(
		ctx,
		provider,
		txSender,
		p.adminPrivateKey,
		ixs,
		p.confirmationTimeout,
	)
	if err != nil {
		return nil, fmt.Errorf("send extend ALT tx: %w", err)
	}

	state, err := provider.GetAddressLookupTable(ctx, p.altPublicKey)
	if err != nil {
		return nil, fmt.Errorf("get address lookup table: %w", err)
	}

	return &deployProgramResult{
		Output: "ALT extended: " + p.altPublicKey.String() +
			" with " + strconv.Itoa(len(state.Addresses)) + " addresses",
		TxSignature: sig.String(),
	}, nil
}
