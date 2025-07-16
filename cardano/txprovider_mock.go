package cardanotx

import (
	"context"

	cardanowallet "github.com/Ethernal-Tech/cardano-infrastructure/wallet"
	"github.com/stretchr/testify/mock"
)

type TxProviderTestMock struct {
	mock.Mock
	ReturnDefaultParameters bool
}

// GetTip implements wallet.ITxProvider.
func (m *TxProviderTestMock) GetTip(ctx context.Context) (cardanowallet.QueryTipData, error) {
	args := m.Called(ctx)

	arg0, _ := args.Get(0).(cardanowallet.QueryTipData)

	return arg0, args.Error(1)
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
		return []byte(`{"costModels":{"PlutusV1":[197209,0,1,1,396231,621,0,1,150000,1000,0,1,150000,32,2477736,29175,4,29773,100,29773,100,29773,100,29773,100,29773,100,29773,100,100,100,29773,100,150000,32,150000,32,150000,32,150000,1000,0,1,150000,32,150000,1000,0,8,148000,425507,118,0,1,1,150000,1000,0,8,150000,112536,247,1,150000,10000,1,136542,1326,1,1000,150000,1000,1,150000,32,150000,32,150000,32,1,1,150000,1,150000,4,103599,248,1,103599,248,1,145276,1366,1,179690,497,1,150000,32,150000,32,150000,32,150000,32,150000,32,150000,32,148000,425507,118,0,1,1,61516,11218,0,1,150000,32,148000,425507,118,0,1,1,148000,425507,118,0,1,1,2477736,29175,4,0,82363,4,150000,5000,0,1,150000,32,197209,0,1,1,150000,32,150000,32,150000,32,150000,32,150000,32,150000,32,150000,32,3345831,1,1],"PlutusV2":[205665,812,1,1,1000,571,0,1,1000,24177,4,1,1000,32,117366,10475,4,23000,100,23000,100,23000,100,23000,100,23000,100,23000,100,100,100,23000,100,19537,32,175354,32,46417,4,221973,511,0,1,89141,32,497525,14068,4,2,196500,453240,220,0,1,1,1000,28662,4,2,245000,216773,62,1,1060367,12586,1,208512,421,1,187000,1000,52998,1,80436,32,43249,32,1000,32,80556,1,57667,4,1000,10,197145,156,1,197145,156,1,204924,473,1,208896,511,1,52467,32,64832,32,65493,32,22558,32,16563,32,76511,32,196500,453240,220,0,1,1,69522,11687,0,1,60091,32,196500,453240,220,0,1,1,196500,453240,220,0,1,1,1159724,392670,0,2,806990,30482,4,1927926,82523,4,265318,0,4,0,85931,32,205665,812,1,1,41182,32,212342,32,31220,32,32696,32,43357,32,32247,32,38314,32,35892428,10,9462713,1021,10,38887044,32947,10]},"protocolVersion":{"major":7,"minor":0},"maxBlockHeaderSize":1100,"maxBlockBodySize":65536,"maxTxSize":16384,"txFeeFixed":155381,"txFeePerByte":44,"stakeAddressDeposit":2000000,"stakePoolDeposit":0,"minPoolCost":0,"poolRetireMaxEpoch":18,"stakePoolTargetNum":100,"poolPledgeInfluence":0,"monetaryExpansion":0.1,"treasuryCut":0.1,"collateralPercentage":150,"executionUnitPrices":{"priceMemory":0.0577,"priceSteps":0.0000721},"utxoCostPerByte":4310,"maxTxExecutionUnits":{"memory":16000000,"steps":10000000000},"maxBlockExecutionUnits":{"memory":80000000,"steps":40000000000},"maxCollateralInputs":3,"maxValueSize":5000,"extraPraosEntropy":null,"decentralization":null,"minUTxOValue":null}`), nil
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

func (m *TxProviderTestMock) GetStakeAddressInfo(
	ctx context.Context,
	stakeAddress string,
) (cardanowallet.QueryStakeAddressInfo, error) {
	args := m.Called(ctx, stakeAddress)

	arg0, _ := args.Get(0).(cardanowallet.QueryStakeAddressInfo)

	return arg0, args.Error(1)
}

func (m *TxProviderTestMock) GetStakePools(ctx context.Context) ([]string, error) {
	args := m.Called(ctx)

	arg0, _ := args.Get(0).([]string)

	return arg0, args.Error(1)
}

func (m *TxProviderTestMock) Dispose() {
	m.Called()
}
