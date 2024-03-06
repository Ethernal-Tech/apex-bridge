package relayer

import (
	"context"
	"time"

	cardanotx "github.com/Ethernal-Tech/apex-bridge/cardano"
	ethtxhelper "github.com/Ethernal-Tech/apex-bridge/eth/txhelper"
	cardanowallet "github.com/Ethernal-Tech/cardano-infrastructure/wallet"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/hashicorp/go-hclog"
)

type Relayer struct {
	config *RelayerConfiguration
	logger hclog.Logger
}

func NewRelayer(config *RelayerConfiguration, logger hclog.Logger) *Relayer {
	return &Relayer{
		config: config,
		logger: logger,
	}
}

func (r *Relayer) Execute(ctx context.Context) {
	var (
		ethClient *ethclient.Client
		err       error
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

		if ethClient == nil {
			ethClient, err = ethclient.Dial(r.config.Bridge.NodeUrl)
			if err != nil {
				r.logger.Error("Failed to dial bridge", "err", err)

				continue
			}
		}

		ethTxHelper, _ := ethtxhelper.NewEThTxHelper(ethtxhelper.WithClient(ethClient)) // nolint

		// TODO: handle lost connection errors from ethClient ->
		// in the case of error ethClient should be set to nil in order to redial again next time

		// invoke smart contract(s)
		smartContractData, err := r.getSmartContractData(ctx, ethTxHelper)
		if err != nil {
			r.logger.Error("Failed to query bridge sc", "err", err)

			return // TODO: recoverable error handling?
		}

		if err := r.SendTx(smartContractData); err != nil {
			r.logger.Error("failed to send tx", "err", err)
		}

		timer.Reset(timerTime)
	}
}

func (r *Relayer) SendTx(smartContractData *SmartContractData) error {
	txProvider, err := cardanowallet.NewTxProviderBlockFrost(r.config.Cardano.BlockfrostUrl, r.config.Cardano.BlockfrostAPIKey)
	if err != nil {
		return err
	}

	defer txProvider.Dispose()

	witnesses := make([][]byte, len(smartContractData.multisigSignatures)+len(smartContractData.feePayerMultisigSignatures))
	copy(witnesses, smartContractData.multisigSignatures)
	copy(witnesses[len(smartContractData.multisigSignatures):], smartContractData.feePayerMultisigSignatures)

	txSigned, txHash, err := cardanotx.AssemblyFinalTx(smartContractData.rawTransaction, witnesses)
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
