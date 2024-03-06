package batcher

import (
	"context"
	"math/big"

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
}

// TODO: replace with real smart contract query
func (b Batcher) getSmartContractData(ctx context.Context, ethTxHelper ethtxhelper.IEthTxHelper) (*SmartContractData, error) {
	contract, err := contractbinding.NewTestContract(
		common.HexToAddress(b.config.Bridge.SmartContractAddress),
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

	return &SmartContractData{
		Dummy:                v,
		KeyHashesMultiSig:    dummyKeyHashes[:len(dummyKeyHashes)/2],
		KeyHashesMultiSigFee: dummyKeyHashes[len(dummyKeyHashes)/2:],
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
	dummyOutputs = []cardanowallet.TxOutput{
		{
			Addr:   "addr_test1vqjysa7p4mhu0l25qknwznvj0kghtr29ud7zp732ezwtzec0w8g3u",
			Amount: cardanowallet.MinUTxODefaultValue,
		},
	}
)
