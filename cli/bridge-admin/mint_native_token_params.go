package clibridgeadmin

import (
	"context"
	"encoding/hex"
	"fmt"

	cardanotx "github.com/Ethernal-Tech/apex-bridge/cardano"
	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/spf13/cobra"

	infracommon "github.com/Ethernal-Tech/cardano-infrastructure/common"
	cardanowallet "github.com/Ethernal-Tech/cardano-infrastructure/wallet"
)

const (
	// privateKeyFlag      = "key"
	stakePrivateKeyFlag = "stake-key"
	ogmiosURLFlag       = "ogmios"
	networkIDFlag       = "network-id"
	testnetMagicFlag    = "testnet-magic"
	tokenNameFlag       = "token-name"
	mintAmountFlag      = "amount"

	// privateKeyFlagDesc      = "wallet payment signing key"
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

	bytes, err := getCardanoPrivateKeyBytes(m.privateKeyRaw)
	if err != nil {
		return fmt.Errorf("invalid --%s value %s", privateKeyFlag, m.privateKeyRaw)
	}

	var stakeBytes []byte
	if len(m.stakePrivateKeyRaw) > 0 {
		stakeBytes, err = getCardanoPrivateKeyBytes(m.stakePrivateKeyRaw)
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
		return fmt.Errorf("invalid --%s value %s", mintAmountFlag, m.mintAmount)
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
	txHash, err := fundAddressWithToken(
		context.Background(),
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

	_, _ = outputter.Write([]byte(fmt.Sprintf("Done minting %s:%d. txHash:%s\n", m.tokenName, m.mintAmount, txHash)))

	return nil, nil
}

var (
	_ common.CliCommandExecutor = (*mintNativeTokenParams)(nil)
)

func fundAddressWithToken(ctx context.Context,
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

	txHash, err := mintTokens(
		ctx, networkType, networkMagic, txProvider, minterWallet,
		mintToken, policy,
	)
	if err != nil {
		return "", err
	}

	return txHash, nil
}

func mintTokens(
	ctx context.Context,
	networkType cardanowallet.CardanoNetworkType,
	networkMagic uint,
	txProvider cardanowallet.ITxProvider,
	wallet *cardanowallet.Wallet,
	token cardanowallet.TokenAmount,
	tokenPolicyScript cardanowallet.IPolicyScript,
) (string, error) {
	walletAddr, err := cardanotx.GetAddress(networkType, wallet)
	if err != nil {
		return "", err
	}

	txRaw, txHash, err := createMintTx(
		ctx, networkType, networkMagic, txProvider, wallet,
		token, tokenPolicyScript,
	)
	if err != nil {
		return "", err
	}

	err = submitTokenTx(ctx, txProvider, txRaw, txHash, walletAddr.String())
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

	if err := builder.SetProtocolParametersAndTTL(ctx, txProvider, 0); err != nil {
		return nil, "", err
	}

	allUtxos, err := txProvider.GetUtxos(ctx, senderAddr)
	if err != nil {
		return nil, "", err
	}

	changeOutputMinUtxo, err := cardanowallet.GetTokenCostSum(builder, senderAddr, allUtxos)
	if err != nil {
		return nil, "", err
	}

	mintOutputMinUtxo, err := cardanowallet.GetTokenCostSum(
		builder, senderAddr, []cardanowallet.Utxo{
			{
				Amount: 0,
				Tokens: []cardanowallet.TokenAmount{token},
			},
		},
	)
	if err != nil {
		return nil, "", err
	}

	minUtxoAmount := max(mintOutputMinUtxo, changeOutputMinUtxo, common.MinUtxoAmountDefault)

	desiredLovelaceAmount := common.PotentialFeeDefault + 2*minUtxoAmount

	inputs, err := cardanowallet.GetUTXOsForAmount(
		allUtxos, cardanowallet.AdaTokenName, desiredLovelaceAmount, maxInputs)
	if err != nil {
		return nil, "", err
	}

	senderTokens, err := cardanowallet.GetTokensFromSumMap(inputs.Sum)
	if err != nil {
		return nil, "", err
	}

	txOutput := cardanowallet.TxOutput{
		Addr:   senderAddr,
		Amount: minUtxoAmount,
		Tokens: append(senderTokens, token),
	}

	builder.AddInputs(inputs.Inputs...).AddTokenMints(
		[]cardanowallet.IPolicyScript{tokenPolicyScript},
		[]cardanowallet.TokenAmount{token},
	)
	builder.AddOutputs(txOutput, cardanowallet.TxOutput{
		Addr: senderAddr,
	})

	fee, err := builder.CalculateFee(1)
	if err != nil {
		return nil, "", err
	}

	outputsSumMap := cardanowallet.GetOutputsSum([]cardanowallet.TxOutput{txOutput})
	outputsSumMap[cardanowallet.AdaTokenName] += fee

	lovelaceInputAmount := inputs.Sum[cardanowallet.AdaTokenName]

	change := lovelaceInputAmount - minUtxoAmount - fee
	// handle overflow or insufficient amount
	if change > lovelaceInputAmount || change < minUtxoAmount {
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
	txProvider cardanowallet.ITxProvider,
	txRaw []byte,
	txHash string,
	receiverAddr string,
) error {
	if err := txProvider.SubmitTx(ctx, txRaw); err != nil {
		return err
	}

	fmt.Println("transaction has been submitted. hash =", txHash)

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

	fmt.Printf("transaction has been included in block. hash = %s, balance = %v\n", txHash, newAmounts)

	return nil
}

func getCardanoPrivateKeyBytes(str string) ([]byte, error) {
	bytes, err := cardanowallet.GetKeyBytes(str)
	if err != nil {
		bytes, err = hex.DecodeString(str)
		if err != nil {
			return nil, err
		}

		bytes = cardanowallet.PadKeyToSize(bytes)
	}

	return bytes, nil
}
