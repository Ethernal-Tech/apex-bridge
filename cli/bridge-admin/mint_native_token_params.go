package clibridgeadmin

import (
	"context"
	"fmt"

	cardanotx "github.com/Ethernal-Tech/apex-bridge/cardano"
	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/spf13/cobra"

	infracommon "github.com/Ethernal-Tech/cardano-infrastructure/common"
	cardanowallet "github.com/Ethernal-Tech/cardano-infrastructure/wallet"
)

const (
	stakePrivateKeyFlag = "stake-key"
	ogmiosURLFlag       = "ogmios"
	networkIDFlag       = "network-id"
	testnetMagicFlag    = "testnet-magic"
	tokenNameFlag       = "token-name"
	mintAmountFlag      = "amount"

	stakePrivateKeyFlagDesc = "wallet stake signing key"
	ogmiosURLFlagDesc       = "ogmios url"
	networkIDFlagDesc       = "network id"
	testnetMagicFlagDesc    = "testnet magic number. leave 0 for mainnet"
	tokenNameFlagDesc       = "name of the token to mint"
	mintAmountFlagDesc      = "amount to mint"

	maxInputs            = 40
	testNetProtocolMagic = uint(2)
)

type mintNativeTokenParams struct {
	privateKeyRaw      string
	stakePrivateKeyRaw string
	ogmiosURL          string
	networkID          uint
	testnetMagic       uint
	tokenName          string
	mintAmount         uint64

	wallet *cardanowallet.Wallet
}

// ValidateFlags implements common.CliCommandValidator.
func (m *mintNativeTokenParams) ValidateFlags() error {
	if m.privateKeyRaw == "" {
		return fmt.Errorf("flag --%s not specified", privateKeyFlag)
	}

	bytes, err := cardanotx.GetCardanoPrivateKeyBytes(m.privateKeyRaw)
	if err != nil {
		return fmt.Errorf("invalid --%s value %s", privateKeyFlag, m.privateKeyRaw)
	}

	var stakeBytes []byte
	if len(m.stakePrivateKeyRaw) > 0 {
		stakeBytes, err = cardanotx.GetCardanoPrivateKeyBytes(m.stakePrivateKeyRaw)
		if err != nil {
			return fmt.Errorf("invalid --%s value %s", stakePrivateKeyFlag, m.stakePrivateKeyRaw)
		}
	}

	m.wallet = cardanowallet.NewWallet(bytes, stakeBytes)

	if !common.IsValidHTTPURL(m.ogmiosURL) {
		return fmt.Errorf("invalid --%s: %s", ogmiosURLFlag, m.ogmiosURL)
	}

	if len(m.tokenName) == 0 {
		return fmt.Errorf("invalid --%s value %s", tokenNameFlag, m.tokenName)
	}

	if m.mintAmount <= 0 {
		return fmt.Errorf("invalid --%s value %d", mintAmountFlag, m.mintAmount)
	}

	return nil
}

func (m *mintNativeTokenParams) RegisterFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(
		&m.privateKeyRaw,
		privateKeyFlag,
		"",
		privateKeyFlagDesc,
	)

	cmd.Flags().StringVar(
		&m.stakePrivateKeyRaw,
		stakePrivateKeyFlag,
		"",
		stakePrivateKeyFlagDesc,
	)

	cmd.Flags().StringVar(
		&m.ogmiosURL,
		ogmiosURLFlag,
		"",
		ogmiosURLFlagDesc,
	)

	cmd.Flags().UintVar(
		&m.networkID,
		networkIDFlag,
		0,
		networkIDFlagDesc,
	)

	cmd.Flags().UintVar(
		&m.testnetMagic,
		testnetMagicFlag,
		0,
		testnetMagicFlagDesc,
	)

	cmd.Flags().StringVar(
		&m.tokenName,
		tokenNameFlag,
		"",
		tokenNameFlagDesc,
	)

	cmd.Flags().Uint64Var(
		&m.mintAmount,
		mintAmountFlag,
		0,
		mintAmountFlagDesc,
	)
}

// Execute implements common.CliCommandExecutor.
func (m *mintNativeTokenParams) Execute(outputter common.OutputFormatter) (common.ICommandResult, error) {
	txHash, err := mintTokenOnAddr(
		context.Background(),
		outputter,
		cardanowallet.CardanoNetworkType(m.networkID),
		m.testnetMagic,
		cardanowallet.NewTxProviderOgmios(m.ogmiosURL),
		m.wallet,
		m.tokenName,
		m.mintAmount,
	)

	if err != nil {
		return nil, err
	}

	_, _ = outputter.Write([]byte(fmt.Sprintf(
		"Done minting %s:%d. txHash:%s\n", m.tokenName, m.mintAmount, txHash)))

	return nil, nil
}

var (
	_ common.CliCommandExecutor = (*mintNativeTokenParams)(nil)
)

func mintTokenOnAddr(
	ctx context.Context,
	outputter common.OutputFormatter,
	networkType cardanowallet.CardanoNetworkType,
	networkMagic uint,
	txProvider cardanowallet.ITxProvider,
	minterWallet *cardanowallet.Wallet, tokenName string,
	mintAmount uint64,
) (string, error) {
	keyHash, err := cardanowallet.GetKeyHash(minterWallet.VerificationKey)
	if err != nil {
		return "", err
	}

	policy := cardanowallet.PolicyScript{
		Type:    cardanowallet.PolicyScriptSigType,
		KeyHash: keyHash,
	}

	cardanoCliBinary := cardanowallet.ResolveCardanoCliBinary(networkType)

	pid, err := cardanowallet.NewCliUtils(cardanoCliBinary).GetPolicyID(policy)
	if err != nil {
		return "", err
	}

	mintToken := cardanowallet.NewTokenAmount(
		cardanowallet.NewToken(pid, tokenName), mintAmount)

	walletAddr, err := cardanotx.GetAddress(networkType, minterWallet)
	if err != nil {
		return "", err
	}

	txRaw, txHash, err := createMintTx(
		ctx, networkType, networkMagic, txProvider, minterWallet,
		mintToken, policy,
	)
	if err != nil {
		return "", err
	}

	err = submitTokenTx(ctx, outputter, txProvider, txRaw, txHash, walletAddr.String())
	if err != nil {
		return "", err
	}

	return txHash, nil
}

func createMintTx(
	ctx context.Context,
	networkType cardanowallet.CardanoNetworkType,
	networkMagic uint,
	txProvider cardanowallet.ITxProvider,
	wallet *cardanowallet.Wallet,
	token cardanowallet.TokenAmount,
	tokenPolicyScript cardanowallet.IPolicyScript,
) ([]byte, string, error) {
	walletAddr, err := cardanotx.GetAddress(networkType, wallet)
	if err != nil {
		return nil, "", err
	}

	senderAddr := walletAddr.String()

	builder, err := cardanowallet.NewTxBuilder(cardanowallet.ResolveCardanoCliBinary(networkType))
	if err != nil {
		return nil, "", err
	}

	defer builder.Dispose()

	builder.SetTestNetMagic(networkMagic)

	_, err = infracommon.ExecuteWithRetry(ctx, func(ctx context.Context) (bool, error) {
		return true, builder.SetProtocolParametersAndTTL(ctx, txProvider, 0)
	})
	if err != nil {
		return nil, "", err
	}

	allUtxos, err := infracommon.ExecuteWithRetry(ctx, func(ctx context.Context) ([]cardanowallet.Utxo, error) {
		return txProvider.GetUtxos(ctx, senderAddr)
	})
	if err != nil {
		return nil, "", err
	}

	potentialMinUtxo, err := cardanowallet.GetMinUtxoForSumMap(
		builder,
		senderAddr,
		cardanowallet.AddSumMaps(
			cardanowallet.GetUtxosSum(allUtxos),
			cardanowallet.GetTokensSumMap(token)))
	if err != nil {
		return nil, "", err
	}

	desiredLovelaceAmount := common.PotentialFeeDefault + max(potentialMinUtxo, common.MinUtxoAmountDefault)

	inputs, err := cardanowallet.GetUTXOsForAmount(
		allUtxos, cardanowallet.AdaTokenName, desiredLovelaceAmount, maxInputs)
	if err != nil {
		return nil, "", err
	}

	senderTokens, err := cardanowallet.GetTokensFromSumMap(inputs.Sum)
	if err != nil {
		return nil, "", err
	}

	builder.AddInputs(inputs.Inputs...).AddTokenMints(
		[]cardanowallet.IPolicyScript{tokenPolicyScript},
		[]cardanowallet.TokenAmount{token},
	)
	builder.AddOutputs(cardanowallet.TxOutput{
		Addr:   senderAddr,
		Tokens: append(senderTokens, token),
	})

	fee, err := builder.CalculateFee(1)
	if err != nil {
		return nil, "", err
	}

	lovelaceInputAmount := inputs.Sum[cardanowallet.AdaTokenName]

	change := lovelaceInputAmount - fee
	// handle overflow or insufficient amount
	if change > lovelaceInputAmount || change < potentialMinUtxo {
		return []byte{}, "", fmt.Errorf("insufficient amount: %d", change)
	}

	if change > 0 {
		builder.UpdateOutputAmount(-1, change)
	} else {
		builder.RemoveOutput(-1)
	}

	builder.SetFee(fee)

	txRaw, txHash, err := builder.Build()
	if err != nil {
		return nil, "", err
	}

	txSigned, err := builder.SignTx(txRaw, []cardanowallet.ITxSigner{wallet})
	if err != nil {
		return nil, "", err
	}

	return txSigned, txHash, nil
}

func submitTokenTx(
	ctx context.Context,
	outputter common.OutputFormatter,
	txProvider cardanowallet.ITxProvider,
	txRaw []byte,
	txHash string,
	receiverAddr string,
) error {
	if err := txProvider.SubmitTx(ctx, txRaw); err != nil {
		return err
	}

	_, _ = outputter.Write([]byte(fmt.Sprintf(
		"transaction has been submitted. hash = %s\n", txHash)))

	newAmounts, err := infracommon.ExecuteWithRetry(ctx, func(ctx context.Context) (map[string]uint64, error) {
		utxos, err := txProvider.GetUtxos(ctx, receiverAddr)
		if err != nil {
			return nil, err
		}

		for _, x := range utxos {
			if x.Hash == txHash {
				return cardanowallet.GetUtxosSum(utxos), nil
			}
		}

		return nil, infracommon.ErrRetryTryAgain
	}, infracommon.WithRetryCount(60))
	if err != nil {
		return err
	}

	_, _ = outputter.Write([]byte(fmt.Sprintf(
		"transaction has been included in block. hash = %s, balance = %v\n", txHash, newAmounts)))

	return nil
}
