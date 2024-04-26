package cardanotx

import (
	"context"
	"encoding/hex"

	cardanowallet "github.com/Ethernal-Tech/cardano-infrastructure/wallet"
	"github.com/stretchr/testify/mock"
)

type TxProviderTestMock struct {
	mock.Mock
	ReturnDefaultParameters bool
}

var _ cardanowallet.ITxProvider = (*TxProviderTestMock)(nil)

func (m *TxProviderTestMock) SubmitTx(ctx context.Context, txSigned []byte) error {
	args := m.Called(ctx, txSigned)

	return args.Error(0)
}

func (m *TxProviderTestMock) GetTxByHash(ctx context.Context, hash string) (map[string]interface{}, error) {
	args := m.Called(ctx, hash)

	arg0, _ := args.Get(0).(map[string]interface{})

	return arg0, args.Error(1)
}

func (m *TxProviderTestMock) GetSlot(ctx context.Context) (uint64, error) {
	args := m.Called(ctx)

	arg0, _ := args.Get(0).(uint64)

	return arg0, args.Error(1)
}

func (m *TxProviderTestMock) GetProtocolParameters(ctx context.Context) ([]byte, error) {
	if m.ReturnDefaultParameters {
		//nolint:lll
		protocolParameters, _ := hex.DecodeString("7b22636f6c6c61746572616c50657263656e74616765223a3135302c22646563656e7472616c697a6174696f6e223a6e756c6c2c22657865637574696f6e556e6974507269636573223a7b2270726963654d656d6f7279223a302e303537372c2270726963655374657073223a302e303030303732317d2c2265787472615072616f73456e74726f7079223a6e756c6c2c226d6178426c6f636b426f647953697a65223a39303131322c226d6178426c6f636b457865637574696f6e556e697473223a7b226d656d6f7279223a36323030303030302c227374657073223a32303030303030303030307d2c226d6178426c6f636b48656164657253697a65223a313130302c226d6178436f6c6c61746572616c496e70757473223a332c226d61785478457865637574696f6e556e697473223a7b226d656d6f7279223a31343030303030302c227374657073223a31303030303030303030307d2c226d6178547853697a65223a31363338342c226d617856616c756553697a65223a353030302c226d696e506f6f6c436f7374223a3137303030303030302c226d696e5554784f56616c7565223a6e756c6c2c226d6f6e6574617279457870616e73696f6e223a302e3030332c22706f6f6c506c65646765496e666c75656e6365223a302e332c22706f6f6c5265746972654d617845706f6368223a31382c2270726f746f636f6c56657273696f6e223a7b226d616a6f72223a382c226d696e6f72223a307d2c227374616b65416464726573734465706f736974223a323030303030302c227374616b65506f6f6c4465706f736974223a3530303030303030302c227374616b65506f6f6c5461726765744e756d223a3530302c227472656173757279437574223a302e322c2274784665654669786564223a3135353338312c22747846656550657242797465223a34342c227574786f436f737450657242797465223a343331307d")

		return protocolParameters, nil
	}

	args := m.Called(ctx)

	arg0, _ := args.Get(0).([]byte)

	return arg0, args.Error(1)
}

func (m *TxProviderTestMock) GetUtxos(ctx context.Context, addr string) ([]cardanowallet.Utxo, error) {
	args := m.Called(ctx, addr)

	arg0, _ := args.Get(0).([]cardanowallet.Utxo)

	return arg0, args.Error(1)
}

func (m *TxProviderTestMock) Dispose() {
	m.Called()
}
