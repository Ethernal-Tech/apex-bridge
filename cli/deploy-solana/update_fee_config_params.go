package clideploysolana

import (
	"context"
	"fmt"

	"github.com/Ethernal-Tech/apex-bridge/common"
	solsendtx "github.com/Ethernal-Tech/solana-infrastructure/sendtx"
	solanawallet "github.com/Ethernal-Tech/solana-infrastructure/wallet"
	"github.com/spf13/cobra"
)

const (
	updateFeeConfigMinOperationFeeFlag   = "min-operation-fee"
	updateFeeConfigMinFeeForBridgingFlag = "min-fee-for-bridging"

	updateFeeConfigUpdateTreasuryFlag     = "update-treasury"
	updateFeeConfigNewTreasuryAddressFlag = "new-treasury-address"

	updateFeeConfigMinOperationFeeFlagDesc   = "minimal operation fee amount (lamports); sent on every update (use current on-chain values if unchanged)" //nolint:lll
	updateFeeConfigMinFeeForBridgingFlagDesc = "minimal fee for bridging (lamports); sent on every update (use current on-chain values if unchanged)"     //nolint:lll

	updateFeeConfigUpdateTreasuryFlagDesc     = "set new treasury wallet from --new-treasury-address"
	updateFeeConfigNewTreasuryAddressFlagDesc = "new treasury wallet address (required with --update-treasury)"
)

type updateFeeConfigParams struct {
	altCommonParams

	minOperationFee uint64
	bridgingFee     uint64

	updateTreasury     bool
	newTreasuryAddress string
}

func (p *updateFeeConfigParams) setFlags(cmd *cobra.Command) {
	p.setCommonFlags(cmd.Flags())

	cmd.Flags().Uint64Var(
		&p.minOperationFee,
		updateFeeConfigMinOperationFeeFlag,
		0,
		updateFeeConfigMinOperationFeeFlagDesc,
	)
	cmd.Flags().Uint64Var(
		&p.bridgingFee,
		updateFeeConfigMinFeeForBridgingFlag,
		0,
		updateFeeConfigMinFeeForBridgingFlagDesc,
	)
	cmd.Flags().BoolVar(
		&p.updateTreasury,
		updateFeeConfigUpdateTreasuryFlag,
		false,
		updateFeeConfigUpdateTreasuryFlagDesc,
	)
	cmd.Flags().StringVar(
		&p.newTreasuryAddress,
		updateFeeConfigNewTreasuryAddressFlag,
		"",
		updateFeeConfigNewTreasuryAddressFlagDesc,
	)
}

func (p *updateFeeConfigParams) validateFlags() error {
	if err := p.validateCommonFlags(); err != nil {
		return err
	}

	if p.minOperationFee == 0 {
		return fmt.Errorf("min operational fee must be greater than 0: --%s", updateFeeConfigMinOperationFeeFlag)
	}

	if p.bridgingFee == 0 {
		return fmt.Errorf("minimal fee for bridging must be greater than 0: --%s", updateFeeConfigMinFeeForBridgingFlag)
	}

	if p.updateTreasury && p.newTreasuryAddress == "" {
		return fmt.Errorf("new treasury address required when --%s is set: --%s",
			updateFeeConfigUpdateTreasuryFlag, updateFeeConfigNewTreasuryAddressFlag)
	}

	if !p.updateTreasury && p.newTreasuryAddress != "" {
		return fmt.Errorf("use --%s with --%s", updateFeeConfigNewTreasuryAddressFlag, updateFeeConfigUpdateTreasuryFlag)
	}

	if p.updateTreasury {
		if _, err := solanawallet.PublicKeyFromAddress(p.newTreasuryAddress); err != nil {
			return fmt.Errorf("invalid new treasury address: %w", err)
		}
	}

	return nil
}

func (p *updateFeeConfigParams) Execute(outputter common.OutputFormatter) (common.ICommandResult, error) {
	ctx := context.Background()

	provider, err := solanawallet.NewProvider(p.rpcURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create provider: %w", err)
	}

	recentBlockhash, err := provider.GetLatestBlockhash(ctx)
	if err != nil {
		return nil, fmt.Errorf("get latest blockhash: %w", err)
	}

	txSender := solsendtx.NewTxSender(provider, nil)

	txDto := solsendtx.UpdateFeeConfigDto{
		ProgramID:          p.programPublicKey,
		AuthorityAddr:      p.adminPrivateKey.PublicKey().String(),
		MinOperationFee:    p.minOperationFee,
		BridgingFee:        p.bridgingFee,
		UpdateTreasury:     p.updateTreasury,
		NewTreasuryAddress: p.newTreasuryAddress,
	}

	tx, err := txSender.CreateTx(
		ctx,
		p.adminPrivateKey.PublicKey(),
		solsendtx.InstructionTypeUpdateFeeConfig,
		recentBlockhash,
		txDto,
	)
	if err != nil {
		return nil, fmt.Errorf("create update fee config tx: %w", err)
	}

	_, _ = outputter.Write([]byte("Submitting update-fee-config transaction..."))
	outputter.WriteOutput()

	sig, err := sendSignedTransaction(
		ctx,
		provider,
		txSender,
		tx,
		p.confirmationTimeout,
		p.adminPrivateKey,
	)
	if err != nil {
		return nil, fmt.Errorf("send update fee config tx: %w", err)
	}

	return &deployProgramResult{
		Output:      "update fee config tx finalized",
		TxSignature: sig.String(),
	}, nil
}
