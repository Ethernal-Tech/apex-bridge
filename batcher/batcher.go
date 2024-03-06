package batcher

import (
	"context"
	"math/big"
	"time"

	cardanotx "github.com/Ethernal-Tech/apex-bridge/cardano"
	"github.com/Ethernal-Tech/apex-bridge/contractbinding"
	ethtxhelper "github.com/Ethernal-Tech/apex-bridge/eth/txhelper"
	cardanowallet "github.com/Ethernal-Tech/cardano-infrastructure/wallet"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/hashicorp/go-hclog"
)

type Batcher struct {
	config *BatcherConfiguration
	logger hclog.Logger
}

func NewBatcher(config *BatcherConfiguration, logger hclog.Logger) *Batcher {
	return &Batcher{
		config: config,
		logger: logger,
	}
}

func (b *Batcher) Execute(ctx context.Context) {
	var (
		ethClient *ethclient.Client
		err       error
		timerTime = time.Millisecond * time.Duration(b.config.PullTimeMilis)
	)

	timer := time.NewTimer(timerTime)
	defer timer.Stop()

	for {
		select {
		case <-timer.C:
		case <-ctx.Done():
			return
		}

		if ethClient == nil {
			ethClient, err = ethclient.Dial(b.config.Bridge.NodeUrl)
			if err != nil {
				b.logger.Error("Failed to dial bridge", "err", err)

				continue
			}
		}

		ethTxHelper, _ := ethtxhelper.NewEThTxHelper(ethtxhelper.WithClient(ethClient)) // nolint

		// TODO: handle lost connection errors from ethClient ->
		// in the case of error ethClient should be set to nil in order to redial again next time

		// invoke smart contract(s)
		smartContractData, err := b.getSmartContractData(ctx, ethTxHelper)
		if err != nil {
			b.logger.Error("Failed to query bridge sc", "err", err)

			return // TODO: recoverable error handling?
		}

		if err := b.sendTx(ctx, smartContractData, ethTxHelper); err != nil {
			b.logger.Error("failed to send tx", "err", err)
		}

		timer.Reset(timerTime)
	}
}

func (b Batcher) sendTx(ctx context.Context, data *SmartContractData, ethTxHelper ethtxhelper.IEthTxHelper) error {
	contract, err := contractbinding.NewTestContract(
		common.HexToAddress(b.config.Bridge.SmartContractAddress), ethTxHelper.GetClient())
	if err != nil {
		return err
	}

	wallet, err := ethtxhelper.NewEthTxWallet(string(b.config.Bridge.SigningKey))
	if err != nil {
		return err
	}

	witnessMultiSig, witnessMultiSigFee, err := b.createCardanoTxWitness(ctx, data)
	if err != nil {
		return err
	}

	// first call is just for creating tx
	tx, err := ethTxHelper.SendTx(ctx, wallet, bind.TransactOpts{}, true, func(txOpts *bind.TransactOpts) (*types.Transaction, error) {
		return contract.SetValue(txOpts, new(big.Int).SetUint64(
			data.Dummy.Uint64()+uint64(len(witnessMultiSig)+len(witnessMultiSigFee))))
	})
	if err != nil {
		return err
	}

	b.logger.Info("tx has been sent", "tx hash", tx.Hash().String())

	receipt, err := ethTxHelper.WaitForReceipt(ctx, tx.Hash().String(), true)
	if err != nil {
		return err
	}

	b.logger.Info("tx has been executed", "block", receipt.BlockHash.String(), "tx hash", receipt.TxHash.String())

	return nil
}

func (b Batcher) createCardanoTxWitness(_ context.Context, data *SmartContractData) ([]byte, []byte, error) {
	sigKey := cardanotx.NewSigningKey(b.config.Cardano.SigningKeyMultiSig)
	sigKeyFee := cardanotx.NewSigningKey(b.config.Cardano.SigningKeyMultiSigFee)

	txProvider, err := cardanowallet.NewTxProviderBlockFrost(b.config.Cardano.BlockfrostUrl, b.config.Cardano.BlockfrostAPIKey)
	if err != nil {
		return nil, nil, err
	}

	defer txProvider.Dispose()

	metadata, err := cardanotx.CreateMetaData(data.Dummy)
	if err != nil {
		return nil, nil, err
	}

	// TODO: should retrieved from sc
	keyHashesMultiSig := data.KeyHashesMultiSig
	keyHashesMultiSigFee := data.KeyHashesMultiSigFee
	outputs := dummyOutputs

	txInfos, err := cardanotx.NewTxInputInfos(keyHashesMultiSig, keyHashesMultiSigFee, b.config.Cardano.TestNetMagic)
	if err != nil {
		return nil, nil, err
	}

	err = txInfos.CalculateWithRetriever(txProvider, cardanowallet.GetOutputsSum(outputs), b.config.Cardano.PotentialFee)
	if err != nil {
		return nil, nil, err
	}

	protocolParams, err := txProvider.GetProtocolParameters()
	if err != nil {
		return nil, nil, err
	}

	slotNumber, err := txProvider.GetSlot()
	if err != nil {
		return nil, nil, err
	}

	txRaw, err := cardanotx.CreateTx(b.config.Cardano.TestNetMagic, protocolParams, slotNumber+cardanotx.TTLSlotNumberInc,
		metadata, txInfos, outputs)
	if err != nil {
		return nil, nil, err
	}

	witnessMultiSig, err := cardanotx.AddTxWitness(sigKey, txRaw)
	if err != nil {
		return nil, nil, err
	}

	witnessMultiSigFee, err := cardanotx.AddTxWitness(sigKeyFee, txRaw)
	if err != nil {
		return nil, nil, err
	}

	return witnessMultiSig, witnessMultiSigFee, err
}
