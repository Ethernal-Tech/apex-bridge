package clideploysolana

import (
	"context"
	"fmt"
	"strconv"

	"github.com/Ethernal-Tech/apex-bridge/common"
	solsendtx "github.com/Ethernal-Tech/solana-infrastructure/sendtx"
	solanawallet "github.com/Ethernal-Tech/solana-infrastructure/wallet"
	"github.com/gagliardetto/solana-go"
	"github.com/spf13/cobra"
)

const (
	registerMintBurnTokenIDFlag                = "token-id"
	registerMintBurnTokenNameFlag              = "token-name"
	registerMintBurnTokenSymbolFlag            = "token-symbol"
	registerMintBurnTokenURIFlag               = "token-uri"
	registerMintBurnTokenDecimalsFlag          = "token-decimals"
	registerMintBurnTokenMinBridgingAmountFlag = "min-bridging-amount" //nolint:gosec

	registerMintBurnTokenIDFlagDesc     = "token ID"
	registerMintBurnTokenNameFlagDesc   = "token name"
	registerMintBurnTokenSymbolFlagDesc = "token symbol (defaults to token name when empty)" //nolint:gosec

	/*
		The URI field is a Metaplex Token Metadata URI — a URL pointing to an off-chain JSON file that describes the token.
		It's passed into the create_metadata_accounts_v3 CPI call in your Rust program, as seen here:

		register_mint_burn_token.rs
		Lines 164-176
		        create_metadata_accounts_v3(
		            CpiContext::new_with_signer(
		                ctx.accounts.metadata_program.to_account_info(),
		                CreateMetadataAccountsV3 {
		                    metadata: ctx.accounts.metadata.to_account_info(),
		                    mint: mint.to_account_info(),
		                    mint_authority: vault.to_account_info(),
		                    payer: ctx.accounts.authority.to_account_info(),
		                    update_authority: vault.to_account_info(),
		                    system_program: ctx.accounts.system_program.to_account_info(),
		                    rent: ctx.accounts.rent.to_account_info(),
		                },
		                signer_seeds,
		            ),
		The JSON file at that URI should follow the Metaplex Token Metadata Standard and look something like this:

		{
		  "name": "My Token",
		  "symbol": "MTK",
		  "description": "A bridgeable SPL token.",
		  "image": "https://arweave.net/<hash>/image.png",
		  "external_url": "https://yourproject.com",
		  "properties": {
		    "category": "currency"
		  }
		}

		The most important fields are name, symbol, image, and description.
		The name and symbol in the JSON should match what you pass as the Name and Symbol parameters in the DTO.
		Where to host it: The JSON file is typically uploaded to a permanent/immutable storage provider:
		 - Arweave — most common for Solana tokens (e.g. https://arweave.net/<tx-id>)
		 - IPFS — via Pinata, NFT.Storage, etc. (e.g. https://ipfs.io/ipfs/<cid>)
		 - Shadow Drive — Solana-native decentralized storage.
		If you don't need off-chain metadata (no image, no description), you can pass an empty string ""
		the on-chain name and symbol fields will still be stored in the Metaplex metadata account regardless.
		The URI is only for clients/explorers that want to display richer info like an icon or description.
	*/
	registerMintBurnTokenURIFlagDesc               = "token metadata URI" //nolint:gosec
	registerMintBurnTokenDecimalsFlagDesc          = "token decimals"
	registerMintBurnTokenMinBridgingAmountFlagDesc = "minimal token amount to bridge (lamports)"
)

type registerMintBurnTokenParams struct {
	registerTokenCommonParams

	tokenID           uint16
	tokenName         string
	tokenSymbol       string
	tokenURI          string
	tokenDecimals     uint64
	minBridgingAmount uint64
}

func (p *registerMintBurnTokenParams) setFlags(cmd *cobra.Command) {
	p.setRegisterTokenCommonFlags(cmd.Flags())

	cmd.Flags().Uint16Var(
		&p.tokenID,
		registerMintBurnTokenIDFlag,
		0,
		registerMintBurnTokenIDFlagDesc,
	)
	cmd.Flags().StringVar(
		&p.tokenName,
		registerMintBurnTokenNameFlag,
		"",
		registerMintBurnTokenNameFlagDesc,
	)
	cmd.Flags().StringVar(
		&p.tokenSymbol,
		registerMintBurnTokenSymbolFlag,
		"",
		registerMintBurnTokenSymbolFlagDesc,
	)
	cmd.Flags().StringVar(
		&p.tokenURI,
		registerMintBurnTokenURIFlag,
		"",
		registerMintBurnTokenURIFlagDesc,
	)
	cmd.Flags().Uint64Var(
		&p.tokenDecimals,
		registerMintBurnTokenDecimalsFlag,
		uint64(solana.SolDecimals),
		registerMintBurnTokenDecimalsFlagDesc,
	)
	cmd.Flags().Uint64Var(
		&p.minBridgingAmount,
		registerMintBurnTokenMinBridgingAmountFlag,
		0,
		registerMintBurnTokenMinBridgingAmountFlagDesc,
	)
}

func (p *registerMintBurnTokenParams) validateFlags() error {
	if err := p.validateRegisterTokenCommonFlags(); err != nil {
		return err
	}

	if p.tokenName == "" {
		return fmt.Errorf("token name not specified: --%s", registerMintBurnTokenNameFlag)
	}

	if p.tokenSymbol == "" {
		p.tokenSymbol = p.tokenName
	}

	if p.tokenDecimals > uint64(^uint8(0)) {
		return fmt.Errorf("token decimals must be <= 255: --%s", registerMintBurnTokenDecimalsFlag)
	}

	return nil
}

func (p *registerMintBurnTokenParams) Execute(outputter common.OutputFormatter) (common.ICommandResult, error) {
	ctx := context.Background()

	provider, err := solanawallet.NewProvider(p.rpcURL)
	if err != nil {
		return nil, fmt.Errorf("create provider: %w", err)
	}

	recentBlockhash, err := provider.GetLatestBlockhash(ctx)
	if err != nil {
		return nil, fmt.Errorf("get latest blockhash: %w", err)
	}

	txSender := solsendtx.NewTxSender(provider, &solsendtx.ChainConfig{
		TreasuryAddress: p.treasuryPublicKey,
	})

	tokenKeyPair, err := solanawallet.NewWallet()
	if err != nil {
		return nil, fmt.Errorf("create token keypair: %w", err)
	}

	txDto := solsendtx.RegisterTokenMintBurnDto{
		ProgramID:         p.programPublicKey,
		AuthorityAddr:     p.adminPrivateKey.PublicKey().String(),
		TokenMint:         tokenKeyPair.PublicKey.String(),
		TokenID:           p.tokenID,
		MinBridgingAmount: p.minBridgingAmount,
		Name:              p.tokenName,
		Symbol:            p.tokenSymbol,
		URI:               p.tokenURI,
		Decimals:          uint8(p.tokenDecimals), //nolint:gosec // validated in validateFlags
	}

	tx, err := txSender.CreateTx(
		ctx,
		p.adminPrivateKey.PublicKey(),
		solsendtx.InstructionTypeRegisterTokensMintBurn,
		recentBlockhash,
		txDto,
	)
	if err != nil {
		return nil, fmt.Errorf("create register token mint tx: %w", err)
	}

	_, _ = outputter.Write([]byte("Submitting register-mint-burn-token transaction..."))
	outputter.WriteOutput()

	sig, err := sendSignedTransaction(
		ctx,
		provider,
		txSender,
		tx,
		p.confirmationTimeout,
		p.adminPrivateKey,
		tokenKeyPair.PrivateKey,
	)
	if err != nil {
		return nil, fmt.Errorf("send register token mint tx: %w", err)
	}

	return &deployProgramResult{
		Output:      "registered mint/burn token id " + strconv.Itoa(int(p.tokenID)),
		MintAddress: tokenKeyPair.PublicKey.String(),
		TxSignature: sig.String(),
	}, nil
}
