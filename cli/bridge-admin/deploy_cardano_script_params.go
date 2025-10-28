package clibridgeadmin

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"regexp"

	cardanotx "github.com/Ethernal-Tech/apex-bridge/cardano"
	"github.com/Ethernal-Tech/apex-bridge/common"
	infracommon "github.com/Ethernal-Tech/cardano-infrastructure/common"
	cardanowallet "github.com/Ethernal-Tech/cardano-infrastructure/wallet"
	"github.com/spf13/cobra"
)

const (
	plutusScriptDirFlag     = "plutus-script-dir"
	plutusScriptDirFlagDesc = "directory containing Plutus script files"

	nftPolicyIDFlag     = "nft-policy-id"
	nftPolicyIDFlagDesc = "the policy ID of the NFT used in the minting policy"

	nftNameHexFlag     = "nft-name-hex"
	nftNameHexFlagDesc = "the name of the NFT in hex format"

	plutusScriptBuildFile = "build.js"
)

type deploymentTxResult struct {
	txSigned   []byte
	txHash     string
	plutusAddr string
}

type deployCardanoScriptParams struct {
	privateKeyRaw      string
	stakePrivateKeyRaw string
	networkID          uint
	testnetMagic       uint
	ogmiosURL          string
	plutusScriptDir    string
	nftPolicyID        string
	nftNameHex         string

	wallet *cardanowallet.Wallet
}

// ValidateFlags implements common.CliCommandValidator.
func (d *deployCardanoScriptParams) ValidateFlags() error {
	if d.privateKeyRaw == "" {
		return fmt.Errorf("flag --%s not specified", privateKeyFlag)
	}

	bytes, err := cardanotx.GetCardanoPrivateKeyBytes(d.privateKeyRaw)
	if err != nil {
		return fmt.Errorf("invalid --%s value %s", privateKeyFlag, d.privateKeyRaw)
	}

	var stakeBytes []byte
	if len(d.stakePrivateKeyRaw) > 0 {
		stakeBytes, err = cardanotx.GetCardanoPrivateKeyBytes(d.stakePrivateKeyRaw)
		if err != nil {
			return fmt.Errorf("invalid --%s value %s", stakePrivateKeyFlag, d.stakePrivateKeyRaw)
		}
	}

	d.wallet = cardanowallet.NewWallet(bytes, stakeBytes)

	if !common.IsValidHTTPURL(d.ogmiosURL) {
		return fmt.Errorf("invalid --%s: %s", ogmiosURLFlag, d.ogmiosURL)
	}

	hexRe := regexp.MustCompile(`^[0-9a-fA-F]+$`)

	if !hexRe.MatchString(d.nftPolicyID) {
		return fmt.Errorf("invalid --%s value %s", nftPolicyIDFlag, d.nftPolicyID)
	}

	if !hexRe.MatchString(d.nftNameHex) {
		return fmt.Errorf("invalid --%s value %s", nftNameHexFlag, d.nftNameHex)
	}

	return nil
}

func (d *deployCardanoScriptParams) RegisterFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(
		&d.privateKeyRaw,
		privateKeyFlag,
		"",
		privateKeyFlagDesc,
	)

	cmd.Flags().StringVar(
		&d.stakePrivateKeyRaw,
		stakePrivateKeyFlag,
		"",
		stakePrivateKeyFlagDesc,
	)

	cmd.Flags().UintVar(
		&d.networkID,
		networkIDFlag,
		0,
		networkIDFlagDesc,
	)

	cmd.Flags().UintVar(
		&d.testnetMagic,
		testnetMagicFlag,
		0,
		testnetMagicFlagDesc,
	)

	cmd.Flags().StringVar(
		&d.ogmiosURL,
		ogmiosURLFlag,
		"",
		ogmiosURLFlagDesc,
	)

	cmd.Flags().StringVar(
		&d.plutusScriptDir,
		plutusScriptDirFlag,
		"",
		plutusScriptDirFlagDesc,
	)

	cmd.Flags().StringVar(
		&d.nftPolicyID,
		nftPolicyIDFlag,
		"",
		nftPolicyIDFlagDesc,
	)

	cmd.Flags().StringVar(
		&d.nftNameHex,
		nftNameHexFlag,
		"",
		nftNameHexFlagDesc,
	)
}

// Execute implements common.CliCommandExecutor.
func (d *deployCardanoScriptParams) Execute(outputter common.OutputFormatter) (common.ICommandResult, error) {
	ctx := context.Background()
	plutusScriptPath := filepath.Join(filepath.Clean(d.plutusScriptDir), plutusScriptBuildFile)
	txProvider := cardanowallet.NewTxProviderOgmios(d.ogmiosURL)

	result, err := deployCardanoScript(
		ctx,
		outputter,
		plutusScriptPath,
		d,
		cardanowallet.CardanoNetworkType(d.networkID),
		txProvider,
	)
	if err != nil {
		return nil, err
	}

	_, _ = outputter.Write(fmt.Appendf(nil,
		"Cardano script deployment on address %s done. txHash:%s\n", result.PlutusAddr, result.TxHash))

	return result, nil
}

func deployCardanoScript(
	ctx context.Context, outputter common.OutputFormatter,
	plutusScriptPath string, d *deployCardanoScriptParams,
	networkType cardanowallet.CardanoNetworkType, txProvider cardanowallet.ITxProvider,
) (*deployCardanoScriptResult, error) {
	cmd := exec.Command("node", plutusScriptPath, d.nftPolicyID, d.nftNameHex) //nolint:gosec

	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to run build.js: %w", err)
	}

	var plutusScript cardanowallet.PlutusScript
	if err := json.Unmarshal(out, &plutusScript); err != nil {
		return nil, fmt.Errorf("failed to unmarshal Plutus script JSON: %w", err)
	}

	deploymentTxRes, err := createDeployCardanoScriptTx(
		ctx, networkType, d.testnetMagic, txProvider, d.wallet,
		plutusScript,
	)
	if err != nil {
		return nil, err
	}

	refScriptUtxo, err := submitDeploymentTx(ctx, outputter, txProvider, deploymentTxRes)
	if err != nil {
		return nil, err
	}

	cardanoCliBinary := cardanowallet.ResolveCardanoCliBinary(networkType)

	policyID, err := cardanowallet.NewCliUtils(cardanoCliBinary).GetPolicyID(plutusScript)
	if err != nil {
		return nil, err
	}

	return &deployCardanoScriptResult{
		TxHash:           deploymentTxRes.txHash,
		PlutusAddr:       deploymentTxRes.plutusAddr,
		RefScriptUtxoIdx: refScriptUtxo.Index,
		PolicyID:         policyID,
	}, nil
}

func createDeployCardanoScriptTx(
	ctx context.Context,
	networkType cardanowallet.CardanoNetworkType,
	networkMagic uint,
	txProvider cardanowallet.ITxProvider,
	wallet *cardanowallet.Wallet,
	plutusScript cardanowallet.ICardanoArtifact,
) (*deploymentTxResult, error) {
	txCtx, err := prepareCardanoTxBuilder(ctx, networkType, networkMagic, txProvider, wallet)
	if err != nil {
		return nil, err
	}
	defer txCtx.Builder.Dispose()

	potentialMinUtxo, err := cardanowallet.GetMinUtxoForSumMap(
		txCtx.Builder,
		txCtx.SenderAddr,
		cardanowallet.GetUtxosSum(txCtx.AllUtxos),
		plutusScript)
	if err != nil {
		return nil, err
	}

	minUtxo := max(potentialMinUtxo, common.MinUtxoAmountDefault)
	desiredLovelaceAmount := common.PotentialFeeDefault + minUtxo

	inputs, err := cardanowallet.GetUTXOsForAmount(
		txCtx.AllUtxos, cardanowallet.AdaTokenName, desiredLovelaceAmount, maxInputs)
	if err != nil {
		return nil, err
	}

	senderTokens, err := cardanowallet.GetTokensFromSumMap(inputs.Sum)
	if err != nil {
		return nil, err
	}

	_, plutusAddr, err := txCtx.Builder.AddInputs(inputs.Inputs...).AddOutputWithPlutusScript(
		plutusScript,
		minUtxo,
	)
	if err != nil {
		return nil, err
	}

	txCtx.Builder.AddOutputs(cardanowallet.TxOutput{
		Addr:   txCtx.SenderAddr,
		Amount: inputs.Sum[cardanowallet.AdaTokenName] - desiredLovelaceAmount,
		Tokens: senderTokens,
	})

	txSigned, txHash, err := finalizeAndSignTx(
		txCtx.Builder,
		wallet,
		inputs.Sum[cardanowallet.AdaTokenName],
		potentialMinUtxo,
		minUtxo,
	)
	if err != nil {
		return nil, err
	}

	return &deploymentTxResult{
		txSigned:   txSigned,
		txHash:     txHash,
		plutusAddr: plutusAddr,
	}, nil
}

func submitDeploymentTx(
	ctx context.Context,
	outputter common.OutputFormatter,
	txProvider cardanowallet.ITxProvider,
	deploymentTxRes *deploymentTxResult,
) (*cardanowallet.Utxo, error) {
	if err := txProvider.SubmitTx(ctx, deploymentTxRes.txSigned); err != nil {
		return nil, err
	}

	_, _ = outputter.Write(fmt.Appendf(nil,
		"transaction has been submitted. hash = %s\n", deploymentTxRes.txHash))

	refScriptUtxo, err := infracommon.ExecuteWithRetry(ctx, func(ctx context.Context) (*cardanowallet.Utxo, error) {
		utxos, err := txProvider.GetUtxos(ctx, deploymentTxRes.plutusAddr)
		if err != nil {
			return nil, err
		}

		for _, x := range utxos {
			if x.Hash == deploymentTxRes.txHash {
				return &x, nil
			}
		}

		return nil, infracommon.ErrRetryTryAgain
	}, infracommon.WithRetryCount(60))
	if err != nil {
		return nil, err
	}

	_, _ = outputter.Write(fmt.Appendf(nil,
		"transaction has been included in block. Utxo hash = %s, index = %d\n", refScriptUtxo.Hash, refScriptUtxo.Index))

	return refScriptUtxo, nil
}
