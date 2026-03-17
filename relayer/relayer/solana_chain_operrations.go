package relayer

import (
	"context"
	"encoding/hex"
	"fmt"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/eth"
	"github.com/Ethernal-Tech/apex-bridge/relayer/core"
	solanatx "github.com/Ethernal-Tech/apex-bridge/solana"
	"github.com/Ethernal-Tech/solana-infrastructure/sendtx"
	"github.com/Ethernal-Tech/solana-infrastructure/wallet"
	"github.com/gagliardetto/solana-go"
	"github.com/hashicorp/go-hclog"
)

type SolanaChainOperations struct {
	chainID    string
	config     *solanatx.SolanaChainConfig
	privateKey *solana.PrivateKey
	txSender   *sendtx.TxSender
	logger     hclog.Logger
}

func NewSolanaChainOperations(
	chainConfig core.ChainConfig,
	logger hclog.Logger,
) (*SolanaChainOperations, error) {
	config, err := solanatx.NewSolanaChainConfig(chainConfig.ChainSpecific)
	if err != nil {
		return nil, err
	}

	secretsManager, err := common.GetSecretsManager(
		chainConfig.RelayerDataDir, chainConfig.RelayerConfigPath, true)
	if err != nil {
		return nil, fmt.Errorf("failed to create secrets manager: %w", err)
	}

	privateKey, err := solanatx.LoadRelayerSolanaPrivateKey(secretsManager, chainConfig.ChainID)
	if err != nil {
		return nil, fmt.Errorf("failed to load relayer solana private key: %w", err)
	}

	txProvider, err := wallet.NewProvider(config.TxProviderEndpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to create tx provider: %w", err)
	}

	txSender := sendtx.NewTxSender(txProvider, nil)

	return &SolanaChainOperations{
		chainID:    chainConfig.ChainID,
		config:     config,
		privateKey: privateKey,
		txSender:   txSender,
		logger:     logger,
	}, nil
}

func (sco *SolanaChainOperations) SendTx(
	ctx context.Context, _ eth.IBridgeSmartContract, smartContractData *eth.ConfirmedBatch,
) error {
	var sig solana.Signature

	signatures := make([]solana.Signature, len(smartContractData.Signatures))

	for i, bytes := range smartContractData.Signatures {
		sig = solana.SignatureFromBytes(bytes)
		if sig.Equals(solana.Signature{}) {
			return fmt.Errorf("invalid signature: %s", hex.EncodeToString(bytes))
		}

		signatures[i] = sig
	}

	tx, err := wallet.UnmarshalTransaction(smartContractData.RawTransaction)
	if err != nil {
		return fmt.Errorf("failed to unmarshal tx: %w", err)
	}

	tx.Signatures = signatures

	_, err = tx.Sign(func(_ solana.PublicKey) *solana.PrivateKey {
		return sco.privateKey
	})
	if err != nil {
		return fmt.Errorf("failed to sign tx: %w", err)
	}

	signature, err := sco.txSender.SendTx(ctx, tx)
	if err != nil {
		return fmt.Errorf("failed to send tx: %w", err)
	}

	sco.logger.Info("Submitted the bridge transaction to skyline solana program",
		"signature", signature.String(),
		"bitmap", smartContractData.Bitmap,
		"rawTx", hex.EncodeToString(smartContractData.RawTransaction))

	return nil
}
