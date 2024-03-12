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
	config    *BatcherConfiguration
	logger    hclog.Logger
	ethClient *ethclient.Client
}

func NewBatcher(config *BatcherConfiguration, logger hclog.Logger) *Batcher {
	return &Batcher{
		config:    config,
		logger:    logger,
		ethClient: nil,
	}
}

func (b *Batcher) Start(ctx context.Context) {
	var (
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

		b.execute(ctx)

		timer.Reset(timerTime)
	}
}

func (b *Batcher) execute(ctx context.Context) {
	var (
		err error
	)

	if b.ethClient == nil {
		b.ethClient, err = ethclient.Dial(b.config.Bridge.NodeUrl)
		if err != nil {
			b.logger.Error("Failed to dial bridge", "err", err, "chainId", b.config.CardanoChain.ChainId)
		}
	}

	ethTxHelper, err := ethtxhelper.NewEThTxHelper(ethtxhelper.WithClient(b.ethClient)) // nolint
	if err != nil {
		// In case of error, reset ethClient to nil to try again in the next iteration.
		b.ethClient = nil
		return
	}

	// TODO: Update smart contract calls depeding of configuration
	// invoke smart contract(s)
	smartContractData, err := b.getSmartContractData(ctx, ethTxHelper)
	if err != nil {
		b.logger.Error("Failed to query bridge sc", "err", err, "chainId", b.config.CardanoChain.ChainId)

		b.ethClient = nil
		return
	}

	if err := b.sendTx(ctx, smartContractData, ethTxHelper); err != nil {
		b.logger.Error("failed to send tx", "err", err, "chainId", b.config.CardanoChain.ChainId)
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

	b.logger.Info("tx has been sent", "tx hash", tx.Hash().String(), "chainId", b.config.CardanoChain.ChainId)

	receipt, err := ethTxHelper.WaitForReceipt(ctx, tx.Hash().String(), true)
	if err != nil {
		return err
	}

	b.logger.Info("tx has been executed", "block", receipt.BlockHash.String(), "tx hash", receipt.TxHash.String(), "chainId", b.config.CardanoChain.ChainId)

	return nil
}

func (b Batcher) createCardanoTxWitness(_ context.Context, data *SmartContractData) ([]byte, []byte, error) {
	sigKey := cardanotx.NewSigningKey(b.config.CardanoChain.SigningKeyMultiSig)
	sigKeyFee := cardanotx.NewSigningKey(b.config.CardanoChain.SigningKeyMultiSigFee)

	txProvider, err := cardanowallet.NewTxProviderBlockFrost(b.config.CardanoChain.BlockfrostUrl, b.config.CardanoChain.BlockfrostAPIKey)
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

	txInfos, err := cardanotx.NewTxInputInfos(keyHashesMultiSig, keyHashesMultiSigFee, b.config.CardanoChain.TestNetMagic)
	if err != nil {
		return nil, nil, err
	}

	err = txInfos.CalculateWithRetriever(txProvider, cardanowallet.GetOutputsSum(outputs), b.config.CardanoChain.PotentialFee)
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

	_, txHash, err := cardanotx.CreateTx(b.config.CardanoChain.TestNetMagic, protocolParams, slotNumber+cardanotx.TTLSlotNumberInc,
		metadata, txInfos, outputs)
	if err != nil {
		return nil, nil, err
	}

	witnessMultiSig, err := cardanotx.CreateTxWitness(txHash, sigKey)
	if err != nil {
		return nil, nil, err
	}

	witnessMultiSigFee, err := cardanotx.CreateTxWitness(txHash, sigKeyFee)
	if err != nil {
		return nil, nil, err
	}

	return witnessMultiSig, witnessMultiSigFee, err
}
