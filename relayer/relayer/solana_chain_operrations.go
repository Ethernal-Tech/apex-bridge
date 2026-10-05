package relayer

import (
	"context"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/eth"
	"github.com/Ethernal-Tech/apex-bridge/relayer/core"
	solanatx "github.com/Ethernal-Tech/apex-bridge/solana"
	"github.com/Ethernal-Tech/solana-infrastructure/sendtx"
	"github.com/Ethernal-Tech/solana-infrastructure/wallet"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/hashicorp/go-hclog"
)

const expiredBlockhashErrMsg = "Blockhash not found"

type SolanaChainOperations struct {
	chainID    string
	config     *solanatx.SolanaChainConfig
	privateKey *solana.PrivateKey
	txSender   *sendtx.TxSender
	txProvider wallet.ITxProvider
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

	txProvider, err := wallet.NewProvider(config.TxProviderEndpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create tx provider: %w", err)
	}

	txSender := sendtx.NewTxSender(txProvider, nil)

	return &SolanaChainOperations{
		chainID:    chainConfig.ChainID,
		config:     config,
		privateKey: privateKey,
		txSender:   txSender,
		txProvider: txProvider,
		logger:     logger,
	}, nil
}

func (sco *SolanaChainOperations) SendTx(
	ctx context.Context, sc eth.IBridgeSmartContract, smartContractData *eth.ConfirmedBatch,
) error {
	var payload sendtx.SolanaPayload

	err := payload.Unmarshal(smartContractData.RawTransaction)
	if err != nil {
		return fmt.Errorf("failed to unmarshal solana payload: %w", err)
	}

	sco.logger.Info("Received payload", "payload", payload)

	if payload.BatchID != smartContractData.ID {
		return fmt.Errorf("batch ID mismatch: %d != %d", payload.BatchID, smartContractData.ID)
	}

	signaturePairs, err := sco.getSignaturePairs(ctx, sc, smartContractData.Signatures, smartContractData.RawTransaction)
	if err != nil {
		return fmt.Errorf("failed to get signature pairs: %w", err)
	}

	if len(signaturePairs) == 0 {
		return fmt.Errorf("no signature pairs provided")
	}

	sco.logger.Info("Signature pairs", "len", len(signaturePairs), "signaturePairs", signaturePairs)

	programID, err := solana.PublicKeyFromBase58(sco.config.ProgramID)
	if err != nil {
		return fmt.Errorf("failed to convert program ID to solana.PublicKey: %w", err)
	}

	bridgingTxDto := sendtx.BridgeTransactionDto{
		Ctx:            ctx,
		ProgramID:      programID,
		SenderAddr:     sco.privateKey.PublicKey().String(),
		Receivers:      payload.Receivers,
		SignaturePairs: signaturePairs,
		PayloadBytes:   smartContractData.RawTransaction,
	}

	if len(payload.Blockhash) != 32 || payload.Blockhash == [32]byte{} {
		return fmt.Errorf("invalid blockhash: %s", payload.Blockhash)
	}

	blockhash := solana.HashFromBytes(payload.Blockhash[:])

	tx, err := sco.txSender.CreateTx(
		ctx, sco.privateKey.PublicKey(),
		sendtx.InstructionTypeBridgeTransaction,
		blockhash,
		bridgingTxDto,
	)
	if err != nil {
		return fmt.Errorf("failed to create tx: %w", err)
	}

	if version := tx.Message.GetVersion(); version != solana.MessageVersionV1 {
		return fmt.Errorf("expected a v1 batch transaction, got message version %s", messageVersionLabel(version))
	}

	binaryTx, err := wallet.MarshalTransaction(tx)
	if err != nil {
		return fmt.Errorf("failed to marshal tx: %w", err)
	}

	sco.logger.Info("Signing transaction", "tx", hex.EncodeToString(binaryTx))

	_, err = tx.Sign(func(key solana.PublicKey) *solana.PrivateKey {
		return sco.privateKey
	})
	if err != nil {
		return fmt.Errorf("failed to sign tx: %w", err)
	}

	// Size is logged after signing: the marshal above carries no signature
	// bytes yet, so it is not the size the node sees.
	signedBinaryTx, err := wallet.MarshalTransaction(tx)
	if err != nil {
		return fmt.Errorf("failed to marshal signed tx: %w", err)
	}

	sco.logger.Info("Signed transaction",
		"size", len(signedBinaryTx),
		"maxSize", solana.MaxTransactionSizeV1,
		"version", messageVersionLabel(tx.Message.GetVersion()),
		"receivers", len(payload.Receivers))

	txSignature, err := sco.txSender.SendTx(ctx, tx)
	if err != nil {
		if isExpiredBlockhashErr(err) {
			// unrecoverable, no point in retrying
			sco.logger.Warn("blockhash expired, treating tx as already submitted", "err", err)

			return nil
		}

		return fmt.Errorf("failed to send tx: %w", err)
	}

	err = sco.txProvider.WaitForSignature(ctx, *txSignature, rpc.CommitmentFinalized, sco.config.ConfirmationTimeout)
	if err != nil {
		return fmt.Errorf("failed to wait for signature: %w", err)
	}

	sco.logger.Info("Submitted the bridge transaction to skyline solana program",
		"signature", txSignature.String(),
		"bitmap", smartContractData.Bitmap,
		"rawTx", hex.EncodeToString(smartContractData.RawTransaction))

	return nil
}

func (sco *SolanaChainOperations) getSignaturePairs(
	ctx context.Context,
	sc eth.IBridgeSmartContract,
	signatures [][]byte,
	payloadBytes []byte,
) (map[solana.PublicKey]solana.Signature, error) {
	validatorsData, err := sc.GetValidatorsChainData(ctx, sco.chainID)
	if err != nil {
		return nil, fmt.Errorf("failed to get validators data: %w", err)
	}

	validatorPublicKeys := make([]solana.PublicKey, len(validatorsData))

	for i, validatorData := range validatorsData {
		rawPubKey := make([]byte, solana.PublicKeyLength)
		validatorData.Key[0].FillBytes(rawPubKey)

		validatorPublicKeys[i], err = wallet.PublicKeyFromBytes(rawPubKey)
		if err != nil {
			return nil, fmt.Errorf("failed to convert validator key to public key: %w", err)
		}
	}

	signaturePairs := make(map[solana.PublicKey]solana.Signature)

	signatureThreashold := common.GetRequiredSignaturesForConsensus(uint64(len(validatorPublicKeys)))
	if len(signatures) < int(signatureThreashold) { //nolint:gosec
		return nil, fmt.Errorf("not enough signatures: %d < %d", len(signatures), signatureThreashold)
	}

	for _, signatureBytes := range signatures {
		if len(signaturePairs) >= int(signatureThreashold) { //nolint:gosec
			break
		}

		signature := solana.SignatureFromBytes(signatureBytes)

		for _, validatorPublicKey := range validatorPublicKeys {
			if signature.Verify(validatorPublicKey, payloadBytes) {
				signaturePairs[validatorPublicKey] = signature

				break
			}
		}
	}

	return signaturePairs, nil
}

// messageVersionLabel names a message version the way the wire format and the
// docs do. solana-go's enum is offset by one - legacy is 0, v0 is 1, v1 is 2 -
// so logging the raw value reads as the version below the one in hand.
func messageVersionLabel(version solana.MessageVersion) string {
	switch version {
	case solana.MessageVersionLegacy:
		return "legacy"
	case solana.MessageVersionV0:
		return "v0"
	case solana.MessageVersionV1:
		return "v1"
	default:
		return fmt.Sprintf("unknown(%d)", version)
	}
}

func isExpiredBlockhashErr(err error) bool {
	return err != nil && strings.Contains(err.Error(), expiredBlockhashErrMsg)
}
