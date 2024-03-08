package relayer

import (
	"context"
	"encoding/json"
	"os"
	"time"

	cardanotx "github.com/Ethernal-Tech/apex-bridge/cardano"
	cardanowallet "github.com/Ethernal-Tech/cardano-infrastructure/wallet"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/hashicorp/go-hclog"
)

type Relayer struct {
	config     *RelayerConfiguration
	logger     hclog.Logger
	operations IChainOperations
}

func NewRelayer(config *RelayerConfiguration, logger hclog.Logger, operations IChainOperations) *Relayer {
	return &Relayer{
		config:     config,
		logger:     logger,
		operations: operations,
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

		// invoke smart contract(s)
		smartContractData, err := r.operations.GetConfirmedBatch(ctx, ethClient, r.config.Bridge.SmartContractAddress)
		if err != nil {
			r.logger.Error("Failed to query bridge sc", "err", err)

			continue
		}

		if err := r.SendTx(smartContractData); err != nil {
			r.logger.Error("failed to send tx", "err", err)
		}

		timer.Reset(timerTime)
	}
}

func (r *Relayer) SendTx(smartContractData *ConfirmedBatch) error {
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

func LoadConfig() (*RelayerConfiguration, error) {
	f, err := os.Open("config.json")
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var appConfig RelayerConfiguration
	decoder := json.NewDecoder(f)
	err = decoder.Decode(&appConfig)
	if err != nil {
		return nil, err
	}

	return &appConfig, nil
}
