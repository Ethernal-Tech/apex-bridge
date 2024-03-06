package relayer

import (
	"context"
	"math/big"

	cardanotx "github.com/Ethernal-Tech/apex-bridge/cardano"
	"github.com/Ethernal-Tech/apex-bridge/contractbinding"
	ethtxhelper "github.com/Ethernal-Tech/apex-bridge/eth/txhelper"
	cardanowallet "github.com/Ethernal-Tech/cardano-infrastructure/wallet"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
)

// TODO: real sc data
type SmartContractData struct {
	Dummy                *big.Int
	KeyHashesMultiSig    []string
	KeyHashesMultiSigFee []string
	WitnessesMultiSig    [][]byte
	WitnessesMultiSigFee [][]byte
}

// TODO: replace with real smart contract query
func (r Relayer) getSmartContractData(ctx context.Context, ethTxHelper ethtxhelper.IEthTxHelper) (*SmartContractData, error) {
	contract, err := contractbinding.NewTestContract(
		common.HexToAddress(r.config.Bridge.SmartContractAddress),
		ethTxHelper.GetClient())
	if err != nil {
		return nil, err // TODO: recoverable error?
	}

	v, err := contract.GetValue(&bind.CallOpts{
		Context: ctx,
	})
	if err != nil {
		return nil, err
	}

	txProvider, err := cardanowallet.NewTxProviderBlockFrost(r.config.Cardano.BlockfrostUrl, r.config.Cardano.BlockfrostAPIKey)
	if err != nil {
		return nil, err
	}

	defer txProvider.Dispose()

	metadata, err := cardanotx.CreateMetaData(v)
	if err != nil {
		return nil, err
	}

	// TODO: should retrieved from sc
	keyHashesMultiSig := dummyKeyHashes[:len(dummyKeyHashes)/2]
	keyHashesMultiSigFee := dummyKeyHashes[len(dummyKeyHashes)/2:]
	outputs := dummyOutputs

	txInfos, err := cardanotx.NewTxInputInfos(keyHashesMultiSig, keyHashesMultiSigFee, r.config.Cardano.TestNetMagic)
	if err != nil {
		return nil, err
	}

	err = txInfos.CalculateWithRetriever(txProvider, cardanowallet.GetOutputsSum(outputs), r.config.Cardano.PotentialFee)
	if err != nil {
		return nil, err
	}

	protocolParams, err := txProvider.GetProtocolParameters()
	if err != nil {
		return nil, err
	}

	slotNumber, err := txProvider.GetSlot()
	if err != nil {
		return nil, err
	}

	txRaw, err := cardanotx.CreateTx(r.config.Cardano.TestNetMagic, protocolParams, slotNumber+cardanotx.TTLSlotNumberInc,
		metadata, txInfos, outputs)
	if err != nil {
		return nil, err
	}

	witnessesMultiSig := make([][]byte, len(dummySigningKeys)/2)
	witnessesMultiSigFee := make([][]byte, len(dummySigningKeys)/2)
	for i := range witnessesMultiSig {
		sigKey := cardanotx.NewSigningKey(dummySigningKeys[i])
		sigKeyFee := cardanotx.NewSigningKey(dummySigningKeys[i+len(dummySigningKeys)/2])

		witnessesMultiSig[i], err = cardanotx.AddTxWitness(sigKey, txRaw)
		if err != nil {
			return nil, err
		}

		witnessesMultiSigFee[i], err = cardanotx.AddTxWitness(sigKeyFee, txRaw)
		if err != nil {
			return nil, err
		}
	}

	return &SmartContractData{
		Dummy:                v,
		KeyHashesMultiSig:    keyHashesMultiSig,
		KeyHashesMultiSigFee: keyHashesMultiSigFee,
		WitnessesMultiSig:    witnessesMultiSig,
		WitnessesMultiSigFee: witnessesMultiSigFee,
	}, nil
}

var (
	dummyKeyHashes = []string{
		"089732e4f6fc248b599c6b24b75187c39842f515733c833e0f09795b",
		"474187985a19732d1abbe1114c1af4cf084d58511884800ddfca3a82",
		"d92df0aff3bf46f084c5744ef25ef33f34318621027a66790b66da31",
		"cd0f2d9b43edb2cfa501f4d7c64413ed57c9147ce0c3aac520bfc565",
		"f8dd5736c4bc7b0d07bff7f018948838f87c703c01b368a38f2cf234",
		"004ee443c6b1a1aa59699545b7bfdf25db64c4d3a64fd1fe10d20829",
	}
	dummySigningKeys = []string{
		"58201217236ac24d8ac12684b308cf9468f68ef5283096896dc1c5c3caf8351e2847",
		"58207e62090b7c574dd71423d4d1d089675bcde049fb2c677fea7add2d94120f01de",
		"582060d76923536885313a7a9dc5a8ed68a22a5e0edee88ca5eb8b10f1e162c57530",
		"5820f2c3b9527ec2f0d70e6ee2db5752e27066fe63f5c84d1aa5bf20a5fc4d2411e6",
		"58202bf1bed17d19f44f53ac64fa4621c879f8295af52080cffb2a8d9d10117ae772",
		"58202cdf4d3b56f3d9ea7b7c9424d841273e2adb1bd11a98a4370ad22f3bac9104e2",
	}
	dummyOutputs = []cardanowallet.TxOutput{
		{
			Addr:   "addr_test1vqjysa7p4mhu0l25qknwznvj0kghtr29ud7zp732ezwtzec0w8g3u",
			Amount: cardanowallet.MinUTxODefaultValue,
		},
	}
)
