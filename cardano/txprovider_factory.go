package cardanotx

import cardanowallet "github.com/Ethernal-Tech/cardano-infrastructure/wallet"

func GetTxProvider(url, apiKey string) (cardanowallet.ITxProvider, error) {
	if url == "" {
		return &TxProviderTestMock{
			ReturnDefaultParameters: true,
		}, nil
	}

	return cardanowallet.NewTxProviderBlockFrost(url, apiKey)
}
