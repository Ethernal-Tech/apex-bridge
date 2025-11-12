package cardanotx

import (
	"context"
	"encoding/hex"
	"testing"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
	"github.com/Ethernal-Tech/cardano-infrastructure/wallet"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateTx(t *testing.T) {
	protocolParameters, _ := hex.DecodeString("7b22636f6c6c61746572616c50657263656e74616765223a3135302c22646563656e7472616c697a6174696f6e223a6e756c6c2c22657865637574696f6e556e6974507269636573223a7b2270726963654d656d6f7279223a302e303537372c2270726963655374657073223a302e303030303732317d2c2265787472615072616f73456e74726f7079223a6e756c6c2c226d6178426c6f636b426f647953697a65223a39303131322c226d6178426c6f636b457865637574696f6e556e697473223a7b226d656d6f7279223a36323030303030302c227374657073223a32303030303030303030307d2c226d6178426c6f636b48656164657253697a65223a313130302c226d6178436f6c6c61746572616c496e70757473223a332c226d61785478457865637574696f6e556e697473223a7b226d656d6f7279223a31343030303030302c227374657073223a31303030303030303030307d2c226d6178547853697a65223a31363338342c226d617856616c756553697a65223a353030302c226d696e506f6f6c436f7374223a3137303030303030302c226d696e5554784f56616c7565223a6e756c6c2c226d6f6e6574617279457870616e73696f6e223a302e3030332c22706f6f6c506c65646765496e666c75656e6365223a302e332c22706f6f6c5265746972654d617845706f6368223a31382c2270726f746f636f6c56657273696f6e223a7b226d616a6f72223a382c226d696e6f72223a307d2c227374616b65416464726573734465706f736974223a323030303030302c227374616b65506f6f6c4465706f736974223a3530303030303030302c227374616b65506f6f6c5461726765744e756d223a3530302c227472656173757279437574223a302e322c2274784665654669786564223a3135353338312c22747846656550657242797465223a34342c227574786f436f737450657242797465223a343331307d")
	outputAddr := "addr_test1vqjysa7p4mhu0l25qknwznvj0kghtr29ud7zp732ezwtzec0w8g3u"
	walletsKeyHashes := []string{
		"d6b67f93ffa4e2651271cc9bcdbdedb2539911266b534d9c163cba21",
		"cba89c7084bf0ce4bf404346b668a7e83c8c9c250d1cafd8d8996e41",
	}
	walletsFeeKeyHashes := []string{
		"f0f4837b3a306752a2b3e52394168bc7391de3dce11364b723cc55cf",
		"47344d5bd7b2fea56336ba789579705a944760032585ef64084c92db",
	}
	testnetMagic := uint(42)
	networkID := wallet.MainNetNetwork
	policyScriptMultiSig := wallet.NewPolicyScript(walletsKeyHashes, len(walletsKeyHashes))
	policyScriptFee := wallet.NewPolicyScript(walletsFeeKeyHashes, len(walletsFeeKeyHashes))
	cardanoCliBinary := wallet.ResolveCardanoCliBinary(networkID)
	cliUtils := wallet.NewCliUtils(cardanoCliBinary)
	multiSigAddr, err := cliUtils.GetPolicyScriptEnterpriseAddress(testnetMagic, policyScriptMultiSig)
	require.NoError(t, err)
	feeAddr, err := cliUtils.GetPolicyScriptEnterpriseAddress(testnetMagic, policyScriptFee)
	require.NoError(t, err)
	stakeAddress, err := cliUtils.GetPolicyScriptRewardAddress(testnetMagic, policyScriptMultiSig)
	require.NoError(t, err)
	txInputsInfos := TxInputInfos{
		MultiSig: []*TxInputInfo{
			{
				TxInputs:     wallet.TxInputs{},
				PolicyScript: policyScriptMultiSig,
				Address:      multiSigAddr,
			},
		},
		MultiSigFee: &TxInputInfo{
			TxInputs:     wallet.TxInputs{},
			PolicyScript: policyScriptFee,
			Address:      feeAddr,
		},
	}
	isInOutputs := func(ots []*indexer.TxOutput, addr string) bool {
		for _, x := range ots {
			if x.Address == addr {
				return true
			}
		}
		return false
	}
	getAmountFromOutputs := func(ots []*indexer.TxOutput, addr string) uint64 {
		for _, x := range ots {
			if x.Address == addr {
				return x.Amount
			}
		}
		return 0
	}
	addrAndAmounts := []common.AddressAndAmount{
		{
			Address:       multiSigAddr,
			TokensAmounts: map[string]uint64{wallet.AdaTokenName: common.MinUtxoAmountDefault},
			IncludeChange: common.MinUtxoAmountDefault,
		},
	}
	t.Run("empty multisig inputs", func(t *testing.T) {
		txInputsInfos.MultiSigFee.TxInputs = wallet.TxInputs{
			Inputs: []wallet.TxInput{
				{
					Hash:  "e99a5bde15aa05f24fcc04b7eabc1520d3397283b1ee720de9fe2653abbb0c9f",
					Index: 1,
				},
			},
			Sum: map[string]uint64{wallet.AdaTokenName: common.MinUtxoAmountDefault * 3},
		}
		txInputsInfos.MultiSig[0].TxInputs = wallet.TxInputs{}
		outputs := []wallet.TxOutput{
			{
				Addr:   outputAddr,
				Amount: common.MinUtxoAmountDefault,
			},
		}
		_, _, err := CreateTx(
			context.Background(), cardanoCliBinary, testnetMagic, protocolParameters, 1000, nil, &txInputsInfos, nil, outputs, nil, addrAndAmounts, nil, 0)
		require.ErrorContains(t, err, "no inputs found for either multisig (0) and refund (0) or fee multisig (1)")
	})
	t.Run("empty fee multisig inputs", func(t *testing.T) {
		txInputsInfos.MultiSig[0].TxInputs = wallet.TxInputs{
			Inputs: []wallet.TxInput{
				{
					Hash:  "e99a5bde15aa05f24fcc04b7eabc1520d3397283b1ee720de9fe2653abbb0c9f",
					Index: 1,
				},
			},
		}
		txInputsInfos.MultiSigFee.TxInputs = wallet.TxInputs{}
		outputs := []wallet.TxOutput{
			{
				Addr:   outputAddr,
				Amount: common.MinUtxoAmountDefault,
			},
		}
		_, _, err := CreateTx(
			context.Background(), cardanoCliBinary, testnetMagic, protocolParameters, 1000, nil, &txInputsInfos, nil, outputs, nil, addrAndAmounts, nil, 0)
		require.ErrorContains(t, err, "no inputs found for either multisig (1) and refund (0) or fee multisig (0)")
	})
	t.Run("not enough funds on multisig", func(t *testing.T) {
		txInputsInfos.MultiSig = []*TxInputInfo{
			{
				PolicyScript: policyScriptMultiSig,
				Address:      multiSigAddr,
				TxInputs: wallet.TxInputs{
					Inputs: []wallet.TxInput{
						{
							Hash: "e99a5bde15aa05f24fcc04b7eabc1520d3397283b1ee720de9fe2653abbb0c9f",
						},
					},
					Sum: map[string]uint64{wallet.AdaTokenName: common.MinUtxoAmountDefault * 3},
				},
			},
		}
		txInputsInfos.MultiSigFee.TxInputs = wallet.TxInputs{
			Inputs: []wallet.TxInput{
				{
					Hash:  "e99a5bde15aa05f24fcc04b7eabc1520d3397283b1ee720de9fe2653abbb0c9f",
					Index: 1,
				},
			},
			Sum: map[string]uint64{wallet.AdaTokenName: common.MinUtxoAmountDefault * 3},
		}
		outputs := []wallet.TxOutput{
			{
				Addr:   outputAddr,
				Amount: common.MinUtxoAmountDefault * 4,
			},
		}
		addrAndAmounts[0].TokensAmounts[wallet.AdaTokenName] = common.MinUtxoAmountDefault * 4
		_, _, err := CreateTx(
			context.Background(), cardanoCliBinary, testnetMagic, protocolParameters, 1000, nil, &txInputsInfos, nil, outputs, nil, addrAndAmounts, nil, 0)
		require.ErrorContains(t, err, "invalid amount: has = 3000000")
	})
	t.Run("not enough funds on fee multisig", func(t *testing.T) {
		txInputsInfos.MultiSig = []*TxInputInfo{
			{
				PolicyScript: policyScriptMultiSig,
				Address:      multiSigAddr,
				TxInputs: wallet.TxInputs{
					Inputs: []wallet.TxInput{
						{
							Hash: "e99a5bde15aa05f24fcc04b7eabc1520d3397283b1ee720de9fe2653abbb0c9f",
						},
					},
					Sum: map[string]uint64{wallet.AdaTokenName: common.MinUtxoAmountDefault * 3},
				},
			},
			{
				PolicyScript: policyScriptMultiSig,
				Address:      multiSigAddr,
				TxInputs: wallet.TxInputs{
					Inputs: []wallet.TxInput{
						{
							Hash:  "e99a5bde15aa05f24fcc04b7eabc1520d3397283b1ee720de9fe2653abbb0c9f",
							Index: 1,
						},
					},
					Sum: map[string]uint64{wallet.AdaTokenName: 20},
				},
			},
		}
		txInputsInfos.MultiSigFee.TxInputs = wallet.TxInputs{
			Inputs: []wallet.TxInput{
				{
					Hash:  "e99a5bde15aa05f24fcc04b7eabc1520d3397283b1ee720de9fe2653abbb0c9f",
					Index: 1,
				},
			},
			Sum: map[string]uint64{wallet.AdaTokenName: 20},
		}
		outputs := []wallet.TxOutput{
			{
				Addr:   outputAddr,
				Amount: common.MinUtxoAmountDefault,
			},
		}
		addrAndAmounts[0].TokensAmounts[wallet.AdaTokenName] = common.MinUtxoAmountDefault
		_, _, err := CreateTx(
			context.Background(), cardanoCliBinary, testnetMagic, protocolParameters, 1000, nil, &txInputsInfos, nil, outputs, nil, addrAndAmounts, nil, 0)
		require.ErrorContains(t, err, "invalid amount: has = 20")
	})
	t.Run("multisig and fee not in outputs with change for multisig", func(t *testing.T) {
		txInputsInfos.MultiSig = []*TxInputInfo{
			{
				TxInputs: wallet.TxInputs{
					Inputs: []wallet.TxInput{
						{
							Hash: "e99a5bde15aa05f24fcc04b7eabc1520d3397283b1ee720de9fe2653abbb0c9f",
						},
					},
					Sum: map[string]uint64{wallet.AdaTokenName: common.MinUtxoAmountDefault * 3},
				},
				PolicyScript: policyScriptMultiSig,
				Address:      multiSigAddr,
			},
		}
		txInputsInfos.MultiSigFee.TxInputs = wallet.TxInputs{
			Inputs: []wallet.TxInput{
				{
					Hash:  "e99a5bde15aa05f24fcc04b7eabc1520d3397283b1ee720de9fe2653abbb0c9f",
					Index: 1,
				},
			},
			Sum: map[string]uint64{wallet.AdaTokenName: common.MinUtxoAmountDefault * 3},
		}
		outputs := []wallet.TxOutput{
			{
				Addr:   outputAddr,
				Amount: common.MinUtxoAmountDefault,
			},
		}
		rawTx, hash, err := CreateTx(
			context.Background(), cardanoCliBinary, testnetMagic, protocolParameters, 1000, nil, &txInputsInfos, nil, outputs, nil, addrAndAmounts, nil, 0)
		require.NoError(t, err)
		require.NotEmpty(t, hash)
		info, err := common.ParseTxInfo(rawTx, true)
		require.NoError(t, err)
		require.Len(t, info.Outputs, 3)
		assert.Equal(t, getAmountFromOutputs(info.Outputs, outputAddr), common.MinUtxoAmountDefault)
		assert.Equal(t, getAmountFromOutputs(info.Outputs, multiSigAddr), common.MinUtxoAmountDefault*2)
		assert.True(t, isInOutputs(info.Outputs, feeAddr))
	})
	t.Run("multisig and fee not in outputs without change for multisig", func(t *testing.T) {
		txInputsInfos.MultiSig = []*TxInputInfo{
			{
				TxInputs: wallet.TxInputs{
					Inputs: []wallet.TxInput{
						{
							Hash: "e99a5bde15aa05f24fcc04b7eabc1520d3397283b1ee720de9fe2653abbb0c9f",
						},
					},
					Sum: map[string]uint64{wallet.AdaTokenName: common.MinUtxoAmountDefault},
				},
				PolicyScript: policyScriptMultiSig,
				Address:      multiSigAddr,
			},
		}
		txInputsInfos.MultiSigFee.TxInputs = wallet.TxInputs{
			Inputs: []wallet.TxInput{
				{
					Hash:  "e99a5bde15aa05f24fcc04b7eabc1520d3397283b1ee720de9fe2653abbb0c9f",
					Index: 1,
				},
			},
			Sum: map[string]uint64{wallet.AdaTokenName: common.MinUtxoAmountDefault * 3},
		}
		outputs := []wallet.TxOutput{
			{
				Addr:   outputAddr,
				Amount: common.MinUtxoAmountDefault,
			},
		}
		rawTx, hash, err := CreateTx(
			context.Background(), cardanoCliBinary, testnetMagic, protocolParameters, 1000, nil, &txInputsInfos, nil, outputs, nil, addrAndAmounts, nil, 0)
		require.NoError(t, err)
		require.NotEmpty(t, hash)
		info, err := common.ParseTxInfo(rawTx, true)
		require.NoError(t, err)
		require.Len(t, info.Outputs, 2)
		assert.Equal(t, getAmountFromOutputs(info.Outputs, outputAddr), common.MinUtxoAmountDefault)
		assert.False(t, isInOutputs(info.Outputs, multiSigAddr))
		assert.True(t, isInOutputs(info.Outputs, feeAddr))
	})
	t.Run("multisig in outputs, fee not in outputs with change for multisig", func(t *testing.T) {
		txInputsInfos.MultiSig = []*TxInputInfo{
			{
				TxInputs: wallet.TxInputs{
					Inputs: []wallet.TxInput{
						{
							Hash: "e99a5bde15aa05f24fcc04b7eabc1520d3397283b1ee720de9fe2653abbb0c9f",
						},
					},
					Sum: map[string]uint64{wallet.AdaTokenName: common.MinUtxoAmountDefault * 3},
				},
				PolicyScript: policyScriptMultiSig,
				Address:      multiSigAddr,
			},
		}
		txInputsInfos.MultiSigFee.TxInputs = wallet.TxInputs{
			Inputs: []wallet.TxInput{
				{
					Hash:  "e99a5bde15aa05f24fcc04b7eabc1520d3397283b1ee720de9fe2653abbb0c9f",
					Index: 1,
				},
			},
			Sum: map[string]uint64{wallet.AdaTokenName: common.MinUtxoAmountDefault * 3},
		}
		outputs := []wallet.TxOutput{
			{
				Addr:   multiSigAddr,
				Amount: 131,
			},
			{
				Addr:   outputAddr,
				Amount: common.MinUtxoAmountDefault,
			},
		}
		addrAndAmounts[0].TokensAmounts[wallet.AdaTokenName] = common.MinUtxoAmountDefault + 131
		rawTx, hash, err := CreateTx(
			context.Background(), cardanoCliBinary, testnetMagic, protocolParameters, 1000, nil, &txInputsInfos, nil, outputs, nil, addrAndAmounts, nil, 0)
		require.NoError(t, err)
		require.NotEmpty(t, hash)
		info, err := common.ParseTxInfo(rawTx, true)
		require.NoError(t, err)
		require.Len(t, info.Outputs, 3)
		assert.Equal(t, getAmountFromOutputs(info.Outputs, outputAddr), common.MinUtxoAmountDefault)
		assert.Equal(t, getAmountFromOutputs(info.Outputs, multiSigAddr), common.MinUtxoAmountDefault*2)
		assert.True(t, isInOutputs(info.Outputs, feeAddr))
	})
	t.Run("multisig in outputs, fee not in outputs without change for multisig", func(t *testing.T) {
		txInputsInfos.MultiSig = []*TxInputInfo{
			{
				TxInputs: wallet.TxInputs{
					Inputs: []wallet.TxInput{
						{
							Hash: "e99a5bde15aa05f24fcc04b7eabc1520d3397283b1ee720de9fe2653abbb0c9f",
						},
					},
					Sum: map[string]uint64{wallet.AdaTokenName: common.MinUtxoAmountDefault},
				},
				PolicyScript: policyScriptMultiSig,
				Address:      multiSigAddr,
			},
		}
		txInputsInfos.MultiSigFee.TxInputs = wallet.TxInputs{
			Inputs: []wallet.TxInput{
				{
					Hash:  "e99a5bde15aa05f24fcc04b7eabc1520d3397283b1ee720de9fe2653abbb0c9f",
					Index: 1,
				},
			},
			Sum: map[string]uint64{wallet.AdaTokenName: common.MinUtxoAmountDefault * 3},
		}
		outputs := []wallet.TxOutput{
			{
				Addr:   multiSigAddr,
				Amount: 131,
			},
			{
				Addr:   outputAddr,
				Amount: common.MinUtxoAmountDefault,
			},
		}
		addrAndAmounts[0].TokensAmounts[wallet.AdaTokenName] = common.MinUtxoAmountDefault + 131

		rawTx, hash, err := CreateTx(
			context.Background(), cardanoCliBinary, testnetMagic, protocolParameters, 1000, nil, &txInputsInfos, nil, outputs, nil, addrAndAmounts, nil, 0)
		require.NoError(t, err)
		require.NotEmpty(t, hash)
		info, err := common.ParseTxInfo(rawTx, true)
		require.NoError(t, err)
		require.Len(t, info.Outputs, 2)
		assert.Equal(t, getAmountFromOutputs(info.Outputs, outputAddr), common.MinUtxoAmountDefault)
		assert.False(t, isInOutputs(info.Outputs, multiSigAddr))
		assert.True(t, isInOutputs(info.Outputs, feeAddr))
	})
	t.Run("multisig not in outputs, fee in outputs with change for multisig", func(t *testing.T) {
		txInputsInfos.MultiSig = []*TxInputInfo{
			{
				TxInputs: wallet.TxInputs{
					Inputs: []wallet.TxInput{
						{
							Hash: "e99a5bde15aa05f24fcc04b7eabc1520d3397283b1ee720de9fe2653abbb0c9f",
						},
					},
					Sum: map[string]uint64{wallet.AdaTokenName: common.MinUtxoAmountDefault * 3},
				},
				PolicyScript: policyScriptMultiSig,
				Address:      multiSigAddr,
			},
		}
		txInputsInfos.MultiSigFee.TxInputs = wallet.TxInputs{
			Inputs: []wallet.TxInput{
				{
					Hash:  "e99a5bde15aa05f24fcc04b7eabc1520d3397283b1ee720de9fe2653abbb0c9f",
					Index: 1,
				},
			},
			Sum: map[string]uint64{wallet.AdaTokenName: common.MinUtxoAmountDefault * 3},
		}
		outputs := []wallet.TxOutput{
			{
				Addr:   feeAddr,
				Amount: 131,
			},
			{
				Addr:   outputAddr,
				Amount: common.MinUtxoAmountDefault,
			},
		}
		addrAndAmounts[0].TokensAmounts[wallet.AdaTokenName] = common.MinUtxoAmountDefault + 131
		rawTx, hash, err := CreateTx(
			context.Background(), cardanoCliBinary, testnetMagic, protocolParameters, 1000, nil, &txInputsInfos, nil, outputs, nil, addrAndAmounts, nil, 0)
		require.NoError(t, err)
		require.NotEmpty(t, hash)
		info, err := common.ParseTxInfo(rawTx, true)
		require.NoError(t, err)
		require.Len(t, info.Outputs, 3)
		assert.Equal(t, getAmountFromOutputs(info.Outputs, outputAddr), common.MinUtxoAmountDefault)
		assert.Equal(t, getAmountFromOutputs(info.Outputs, multiSigAddr), common.MinUtxoAmountDefault*2-131)
		assert.True(t, isInOutputs(info.Outputs, feeAddr))
	})
	t.Run("multisig not in outputs, fee in outputs without change for multisig", func(t *testing.T) {
		txInputsInfos.MultiSig = []*TxInputInfo{
			{
				TxInputs: wallet.TxInputs{
					Inputs: []wallet.TxInput{
						{
							Hash: "e99a5bde15aa05f24fcc04b7eabc1520d3397283b1ee720de9fe2653abbb0c9f",
						},
					},
					Sum: map[string]uint64{wallet.AdaTokenName: common.MinUtxoAmountDefault + 131},
				},
				PolicyScript: policyScriptMultiSig,
				Address:      multiSigAddr,
			},
		}
		txInputsInfos.MultiSigFee.TxInputs = wallet.TxInputs{
			Inputs: []wallet.TxInput{
				{
					Hash:  "e99a5bde15aa05f24fcc04b7eabc1520d3397283b1ee720de9fe2653abbb0c9f",
					Index: 1,
				},
			},
			Sum: map[string]uint64{wallet.AdaTokenName: common.MinUtxoAmountDefault * 3},
		}
		outputs := []wallet.TxOutput{
			{
				Addr:   feeAddr,
				Amount: 131,
			},
			{
				Addr:   outputAddr,
				Amount: common.MinUtxoAmountDefault,
			},
		}
		rawTx, hash, err := CreateTx(
			context.Background(), cardanoCliBinary, testnetMagic, protocolParameters, 1000, nil, &txInputsInfos, nil, outputs, nil, addrAndAmounts, nil, 0)
		require.NoError(t, err)
		require.NotEmpty(t, hash)
		info, err := common.ParseTxInfo(rawTx, true)
		require.NoError(t, err)
		require.Len(t, info.Outputs, 2)
		assert.Equal(t, getAmountFromOutputs(info.Outputs, outputAddr), common.MinUtxoAmountDefault)
		assert.False(t, isInOutputs(info.Outputs, multiSigAddr))
		assert.True(t, isInOutputs(info.Outputs, feeAddr))
	})
	t.Run("multisig in outputs, fee in outputs without change for fee", func(t *testing.T) {
		txInputsInfos.MultiSig = []*TxInputInfo{
			{
				TxInputs: wallet.TxInputs{
					Inputs: []wallet.TxInput{
						{
							Hash: "e99a5bde15aa05f24fcc04b7eabc1520d3397283b1ee720de9fe2653abbb0c9f",
						},
					},
					Sum: map[string]uint64{wallet.AdaTokenName: common.MinUtxoAmountDefault + 131},
				},
				PolicyScript: policyScriptMultiSig,
				Address:      multiSigAddr,
			},
		}
		txInputsInfos.MultiSigFee.TxInputs = wallet.TxInputs{
			Inputs: []wallet.TxInput{
				{
					Hash:  "e99a5bde15aa05f24fcc04b7eabc1520d3397283b1ee720de9fe2653abbb0c9f",
					Index: 1,
				},
			},
			Sum: map[string]uint64{wallet.AdaTokenName: 195377 - 131},
		}
		outputs := []wallet.TxOutput{
			{
				Addr:   multiSigAddr,
				Amount: 150,
			},
			{
				Addr:   feeAddr,
				Amount: 131,
			},
			{
				Addr:   outputAddr,
				Amount: common.MinUtxoAmountDefault,
			},
		}
		addrAndAmounts[0].TokensAmounts[wallet.AdaTokenName] = common.MinUtxoAmountDefault + 150 + 131
		rawTx, hash, err := CreateTx(
			context.Background(), cardanoCliBinary, testnetMagic, protocolParameters, 1000, nil, &txInputsInfos, nil, outputs, nil, addrAndAmounts, nil, 0)
		require.NoError(t, err)
		require.NotEmpty(t, hash)

		info, err := common.ParseTxInfo(rawTx, true)
		require.NoError(t, err)
		require.Len(t, info.Outputs, 1)
		assert.Equal(t, getAmountFromOutputs(info.Outputs, outputAddr), common.MinUtxoAmountDefault)
		assert.False(t, isInOutputs(info.Outputs, multiSigAddr))
		assert.False(t, isInOutputs(info.Outputs, feeAddr))

		t.Run("additional witnesses", func(t *testing.T) {
			_, _, err := CreateTx(
				context.Background(), cardanoCliBinary, testnetMagic, protocolParameters, 1000, nil, &txInputsInfos, nil, outputs, nil, addrAndAmounts, nil, 1)
			require.Error(t, err)

			txInputsInfos.MultiSigFee.TxInputs.Sum = map[string]uint64{wallet.AdaTokenName: 199821 - 131}

			rawTx, _, err = CreateTx(
				context.Background(), cardanoCliBinary, testnetMagic, protocolParameters, 1000, nil, &txInputsInfos, nil, outputs, nil, addrAndAmounts, nil, 1)
			require.NoError(t, err)

			info, err := common.ParseTxInfo(rawTx, true)
			require.NoError(t, err)
			require.Len(t, info.Outputs, 1)
			assert.Equal(t, getAmountFromOutputs(info.Outputs, outputAddr), common.MinUtxoAmountDefault)
			assert.False(t, isInOutputs(info.Outputs, multiSigAddr))
			assert.False(t, isInOutputs(info.Outputs, feeAddr))
		})
	})

	keyRegDepositAmount := uint64(2000000)
	registrationCert, err := cliUtils.CreateRegistrationCertificate(stakeAddress, keyRegDepositAmount)
	require.NoError(t, err)

	poolID := "pool1f0drqjkgfhqcdeyvfuvgv9hsss59hpfj5rrrk9hlg7tm29tmkjr"
	delegationCert, err := cliUtils.CreateDelegationCertificate(stakeAddress, poolID)
	require.NoError(t, err)

	certData := &CertificatesData{
		Certificates: []*CertificatesWithScript{
			{
				PolicyScript: policyScriptMultiSig,
				Certificates: []wallet.ICardanoArtifact{registrationCert, delegationCert},
			},
		},
		RegistrationFee: keyRegDepositAmount,
	}
	txInputsInfos = TxInputInfos{
		MultiSig: []*TxInputInfo{
			{
				TxInputs:     wallet.TxInputs{},
				PolicyScript: policyScriptMultiSig,
				Address:      multiSigAddr,
			},
		},
		MultiSigFee: &TxInputInfo{
			TxInputs:     wallet.TxInputs{},
			PolicyScript: policyScriptFee,
			Address:      feeAddr,
		},
	}

	t.Run("only certificates, no multisig inputs and not enough fee", func(t *testing.T) {
		txInputsInfos.MultiSigFee.TxInputs = wallet.TxInputs{
			Inputs: []wallet.TxInput{
				{
					Hash:  "e99a5bde15aa05f24fcc04b7eabc1520d3397283b1ee720de9fe2653abbb0c9f",
					Index: 1,
				},
			},
			Sum: map[string]uint64{wallet.AdaTokenName: keyRegDepositAmount},
		}
		outputs := []wallet.TxOutput{}
		_, _, err = CreateTx(
			context.Background(), cardanoCliBinary, testnetMagic, protocolParameters, 1000, nil, &txInputsInfos, nil, outputs, certData, addrAndAmounts, nil, 0)
		require.ErrorContains(t, err, "invalid amount")
	})
	t.Run("only certificates, no multisig inputs and exact fee amount", func(t *testing.T) {
		txInputsInfos.MultiSigFee.TxInputs = wallet.TxInputs{
			Inputs: []wallet.TxInput{
				{
					Hash:  "e99a5bde15aa05f24fcc04b7eabc1520d3397283b1ee720de9fe2653abbb0c9f",
					Index: 1,
				},
			},
			Sum: map[string]uint64{wallet.AdaTokenName: keyRegDepositAmount + 191901},
		}
		outputs := []wallet.TxOutput{}
		addrAndAmounts[0].TokensAmounts[wallet.AdaTokenName] = 0

		rawTx, hash, err := CreateTx(
			context.Background(), cardanoCliBinary, testnetMagic, protocolParameters, 1000, nil, &txInputsInfos, nil, outputs, certData, addrAndAmounts, nil, 0)
		require.NoError(t, err)
		require.NotEmpty(t, hash)

		info, err := common.ParseTxInfo(rawTx, true)

		require.NoError(t, err)
		assert.False(t, isInOutputs(info.Outputs, feeAddr))
		require.Len(t, info.Outputs, 0)
	})
	t.Run("only certificates, not enough fee", func(t *testing.T) {
		txInputsInfos.MultiSig = []*TxInputInfo{
			{
				TxInputs: wallet.TxInputs{
					Inputs: []wallet.TxInput{
						{
							Hash: "e99a5bde15aa05f24fcc04b7eabc1520d3397283b1ee720de9fe2653abbb0c9f",
						},
					},
					Sum: map[string]uint64{wallet.AdaTokenName: common.MinUtxoAmountDefault},
				},
				PolicyScript: policyScriptMultiSig,
				Address:      multiSigAddr,
			},
		}
		txInputsInfos.MultiSigFee.TxInputs = wallet.TxInputs{
			Inputs: []wallet.TxInput{
				{
					Hash:  "e99a5bde15aa05f24fcc04b7eabc1520d3397283b1ee720de9fe2653abbb0c9f",
					Index: 1,
				},
			},
			Sum: map[string]uint64{wallet.AdaTokenName: keyRegDepositAmount},
		}
		outputs := []wallet.TxOutput{}
		_, _, err = CreateTx(
			context.Background(), cardanoCliBinary, testnetMagic, protocolParameters, 1000, nil, &txInputsInfos, nil, outputs, certData, addrAndAmounts, nil, 0)
		require.ErrorContains(t, err, "invalid amount")
	})
	t.Run("only certificates, exact fee", func(t *testing.T) {
		txInputsInfos.MultiSig = []*TxInputInfo{
			{
				TxInputs: wallet.TxInputs{
					Inputs: []wallet.TxInput{
						{
							Hash: "e99a5bde15aa05f24fcc04b7eabc1520d3397283b1ee720de9fe2653abbb0c9f",
						},
					},
					Sum: map[string]uint64{wallet.AdaTokenName: common.MinUtxoAmountDefault},
				},
				PolicyScript: policyScriptMultiSig,
				Address:      multiSigAddr,
			},
		}
		txInputsInfos.MultiSigFee.TxInputs = wallet.TxInputs{
			Inputs: []wallet.TxInput{
				{
					Hash:  "e99a5bde15aa05f24fcc04b7eabc1520d3397283b1ee720de9fe2653abbb0c9f",
					Index: 1,
				},
			},
			Sum: map[string]uint64{wallet.AdaTokenName: keyRegDepositAmount + 208621},
		}
		outputs := []wallet.TxOutput{
			{
				Addr:   multiSigAddr,
				Amount: common.MinUtxoAmountDefault,
			},
		}
		rawTx, hash, err := CreateTx(
			context.Background(), cardanoCliBinary, testnetMagic, protocolParameters, 1000, nil, &txInputsInfos, nil, outputs, certData, addrAndAmounts, nil, 0)
		require.NoError(t, err)
		require.NotEmpty(t, hash)

		info, err := common.ParseTxInfo(rawTx, true)
		require.NoError(t, err)
		require.Len(t, info.Outputs, 1)
		assert.True(t, isInOutputs(info.Outputs, multiSigAddr))
		assert.False(t, isInOutputs(info.Outputs, feeAddr))
	})
	t.Run("with certificates, exact fee", func(t *testing.T) {
		txInputsInfos.MultiSig = []*TxInputInfo{
			{
				TxInputs: wallet.TxInputs{
					Inputs: []wallet.TxInput{
						{
							Hash: "e99a5bde15aa05f24fcc04b7eabc1520d3397283b1ee720de9fe2653abbb0c9f",
						},
					},
					Sum: map[string]uint64{wallet.AdaTokenName: common.MinUtxoAmountDefault + 150},
				},
				PolicyScript: policyScriptMultiSig,
				Address:      multiSigAddr,
			},
		}
		txInputsInfos.MultiSigFee.TxInputs = wallet.TxInputs{
			Inputs: []wallet.TxInput{
				{
					Hash:  "e99a5bde15aa05f24fcc04b7eabc1520d3397283b1ee720de9fe2653abbb0c9f",
					Index: 1,
				},
			},
			Sum: map[string]uint64{wallet.AdaTokenName: keyRegDepositAmount + 212977},
		}
		outputs := []wallet.TxOutput{
			{
				Addr:   multiSigAddr,
				Amount: 150,
			},
			{
				Addr:   outputAddr,
				Amount: common.MinUtxoAmountDefault,
			},
		}
		addrAndAmounts[0].TokensAmounts[wallet.AdaTokenName] = common.MinUtxoAmountDefault + 150
		rawTx, hash, err := CreateTx(
			context.Background(), cardanoCliBinary, testnetMagic, protocolParameters, 1000, nil, &txInputsInfos, nil, outputs, certData, addrAndAmounts, nil, 0)
		require.NoError(t, err)
		require.NotEmpty(t, hash)

		info, err := common.ParseTxInfo(rawTx, true)
		require.NoError(t, err)
		require.Len(t, info.Outputs, 2)
		assert.Equal(t, getAmountFromOutputs(info.Outputs, outputAddr), common.MinUtxoAmountDefault)
		assert.True(t, isInOutputs(info.Outputs, multiSigAddr))
		assert.True(t, isInOutputs(info.Outputs, outputAddr))
		assert.False(t, isInOutputs(info.Outputs, feeAddr))
	})
	t.Run("with certificates and all outputs", func(t *testing.T) {
		txInputsInfos.MultiSig = []*TxInputInfo{
			{
				TxInputs: wallet.TxInputs{
					Inputs: []wallet.TxInput{
						{
							Hash: "e99a5bde15aa05f24fcc04b7eabc1520d3397283b1ee720de9fe2653abbb0c9f",
						},
					},
					Sum: map[string]uint64{wallet.AdaTokenName: common.MinUtxoAmountDefault + 150},
				},
				PolicyScript: policyScriptMultiSig,
				Address:      multiSigAddr,
			},
		}
		txInputsInfos.MultiSigFee.TxInputs = wallet.TxInputs{
			Inputs: []wallet.TxInput{
				{
					Hash:  "e99a5bde15aa05f24fcc04b7eabc1520d3397283b1ee720de9fe2653abbb0c9f",
					Index: 1,
				},
			},
			Sum: map[string]uint64{wallet.AdaTokenName: keyRegDepositAmount + 217365 + 100},
		}
		outputs := []wallet.TxOutput{
			{
				Addr:   multiSigAddr,
				Amount: 150,
			},
			{
				Addr:   feeAddr,
				Amount: 100,
			},
			{
				Addr:   outputAddr,
				Amount: common.MinUtxoAmountDefault,
			},
		}
		rawTx, hash, err := CreateTx(
			context.Background(), cardanoCliBinary, testnetMagic, protocolParameters, 1000, nil, &txInputsInfos, nil, outputs, certData, addrAndAmounts, nil, 0)
		require.NoError(t, err)
		require.NotEmpty(t, hash)

		info, err := common.ParseTxInfo(rawTx, true)
		require.NoError(t, err)
		require.Len(t, info.Outputs, 3)
		assert.Equal(t, getAmountFromOutputs(info.Outputs, outputAddr), common.MinUtxoAmountDefault)
		assert.True(t, isInOutputs(info.Outputs, multiSigAddr))
		assert.True(t, isInOutputs(info.Outputs, outputAddr))
		assert.True(t, isInOutputs(info.Outputs, feeAddr))
	})
}
