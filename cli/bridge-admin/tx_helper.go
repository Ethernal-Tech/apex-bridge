package clibridgeadmin

import (
	"context"
	"fmt"

	cardanotx "github.com/Ethernal-Tech/apex-bridge/cardano"
	infracommon "github.com/Ethernal-Tech/cardano-infrastructure/common"
	cardanowallet "github.com/Ethernal-Tech/cardano-infrastructure/wallet"
)

type TxBuildContext struct {
	Builder    *cardanowallet.TxBuilder
	SenderAddr string
	AllUtxos   []cardanowallet.Utxo
}

func prepareCardanoTxBuilder(
	ctx context.Context,
	networkType cardanowallet.CardanoNetworkType,
	networkMagic uint,
	txProvider cardanowallet.ITxProvider,
	wallet *cardanowallet.Wallet,
) (*TxBuildContext, error) {
	walletAddr, err := cardanotx.GetAddress(networkType, wallet)
	if err != nil {
		return nil, err
	}

	builder, err := cardanowallet.NewTxBuilder(cardanowallet.ResolveCardanoCliBinary(networkType))
	if err != nil {
		return nil, err
	}

	builder.SetTestNetMagic(networkMagic)

	_, err = infracommon.ExecuteWithRetry(ctx, func(ctx context.Context) (bool, error) {
		return true, builder.SetProtocolParametersAndTTL(ctx, txProvider, 0)
	})
	if err != nil {
		builder.Dispose()

		return nil, err
	}

	allUtxos, err := infracommon.ExecuteWithRetry(ctx, func(ctx context.Context) ([]cardanowallet.Utxo, error) {
		return txProvider.GetUtxos(ctx, walletAddr.String())
	})
	if err != nil {
		builder.Dispose()

		return nil, err
	}

	return &TxBuildContext{
		Builder:    builder,
		SenderAddr: walletAddr.String(),
		AllUtxos:   allUtxos,
	}, nil
}

func finalizeAndSignTx(
	builder *cardanowallet.TxBuilder,
	wallet *cardanowallet.Wallet,
	lovelaceInputAmount uint64,
	potentialMinUtxo uint64,
	spentAmount uint64,
) ([]byte, string, error) {
	fee, err := builder.CalculateFee(1)
	if err != nil {
		return nil, "", err
	}

	change := lovelaceInputAmount - fee - spentAmount
	if change > lovelaceInputAmount || change < potentialMinUtxo {
		return nil, "", fmt.Errorf("insufficient amount: %d", change)
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
