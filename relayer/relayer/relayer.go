package relayer

import (
	"context"
	"time"

	cardanotx "github.com/Ethernal-Tech/apex-bridge/cardano"
	ethtxhelper "github.com/Ethernal-Tech/apex-bridge/eth/txhelper"
	"github.com/Ethernal-Tech/apex-bridge/relayer/bridge"
	"github.com/Ethernal-Tech/apex-bridge/relayer/core"
	cardanowallet "github.com/Ethernal-Tech/cardano-infrastructure/wallet"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/hashicorp/go-hclog"
)

type RelayerImpl struct {
	config    *core.RelayerConfiguration
	logger    hclog.Logger
	ethClient *ethclient.Client
}

var _ core.Relayer = (*RelayerImpl)(nil)

func NewRelayer(config *core.RelayerConfiguration, logger hclog.Logger) *RelayerImpl {
	return &RelayerImpl{
		config:    config,
		logger:    logger,
		ethClient: nil,
	}
}

func (r *RelayerImpl) Start(ctx context.Context) {
	var (
		timerTime = time.Millisecond * time.Duration(r.config.PullTimeMilis)
	)

	timer := time.NewTimer(timerTime)
	defer timer.Stop()

	for {
		select {
		case <-timer.C:
		case <-ctx.Done():
			return
		}

		r.execute(ctx)

		timer.Reset(timerTime)
	}
}

func (r *RelayerImpl) execute(ctx context.Context) {
	var (
		err error
	)

	if r.ethClient == nil {
		r.ethClient, err = ethclient.Dial(r.config.Bridge.NodeUrl)
		if err != nil {
			r.logger.Error("Failed to dial bridge", "err", err)
			return
		}
	}

	ethTxHelper, err := ethtxhelper.NewEThTxHelper(ethtxhelper.WithClient(r.ethClient))
	if err != nil {
		// In case of error, reset ethClient to nil to try again in the next iteration.
		r.ethClient = nil
		return
	}

	// invoke smart contract(s)
	smartContractData, err := bridge.GetSmartContractData(ctx, ethTxHelper, r.config.CardanoChain, r.config.Bridge)
	if err != nil {
		r.logger.Error("Failed to query bridge sc", "err", err)

		r.ethClient = nil
		return
	}

	if err := r.SendTx(smartContractData); err != nil {
		r.logger.Error("failed to send tx", "err", err)
	}
}

func (r *RelayerImpl) SendTx(smartContractData *bridge.SmartContractData) error {
	txProvider, err := cardanowallet.NewTxProviderBlockFrost(r.config.CardanoChain.BlockfrostUrl, r.config.CardanoChain.BlockfrostAPIKey)
	if err != nil {
		return err
	}

	defer txProvider.Dispose()

	// TODO: some things here are hardcoded and contains dummy values
	outputs := dummyOutputs

	txInfos, err := cardanotx.NewTxInputInfos(
		smartContractData.KeyHashesMultiSig, smartContractData.KeyHashesMultiSigFee, r.config.CardanoChain.TestNetMagic)
	if err != nil {
		return err
	}

	err = txInfos.CalculateWithRetriever(txProvider, cardanowallet.GetOutputsSum(outputs), r.config.CardanoChain.PotentialFee)
	if err != nil {
		return err
	}

	metadata, err := cardanotx.CreateMetaData(smartContractData.Dummy)
	if err != nil {
		return err
	}

	protocolParams, err := txProvider.GetProtocolParameters()
	if err != nil {
		return err
	}

	slotNumber, err := txProvider.GetSlot()
	if err != nil {
		return err
	}

	txRaw, txHash, err := cardanotx.CreateTx(r.config.CardanoChain.TestNetMagic, protocolParams, slotNumber+cardanotx.TTLSlotNumberInc,
		metadata, txInfos, outputs)
	if err != nil {
		return err
	}

	witnesses := make([][]byte, len(smartContractData.KeyHashesMultiSig)+len(smartContractData.KeyHashesMultiSigFee))
	copy(witnesses, smartContractData.WitnessesMultiSig)
	copy(witnesses[len(smartContractData.WitnessesMultiSig):], smartContractData.WitnessesMultiSigFee)

	txSigned, err := cardanotx.AssembleTxWitnesses(txRaw, witnesses)
	if err != nil {
		return err
	}

	if err := txProvider.SubmitTx(txSigned); err != nil {
		return err
	}

	r.logger.Info("transaction has been sent", "hash", txHash)

	// TODO: relayer should not wait for confirmation of block including
	// that is the job for oracle
	txData, err := cardanowallet.WaitForTransaction(context.Background(), txProvider, txHash, 100, time.Second*2)
	if err != nil {
		return err
	}

	r.logger.Info("transaction has been included in block", "hash", txHash, "block", txData["block"])

	return nil
}

var (
	dummyOutputs = []cardanowallet.TxOutput{
		{
			Addr:   "addr_test1vqjysa7p4mhu0l25qknwznvj0kghtr29ud7zp732ezwtzec0w8g3u",
			Amount: cardanowallet.MinUTxODefaultValue,
		},
	}
)
