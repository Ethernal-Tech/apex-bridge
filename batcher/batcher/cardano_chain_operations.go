package batcher

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"

	"github.com/Ethernal-Tech/apex-bridge/batcher/core"
	cardano "github.com/Ethernal-Tech/apex-bridge/cardano"
	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/eth"
	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
	"github.com/Ethernal-Tech/cardano-infrastructure/secrets"
	cardanowallet "github.com/Ethernal-Tech/cardano-infrastructure/wallet"
	"github.com/hashicorp/go-hclog"
)

var (
	_ core.ChainOperations = (*CardanoChainOperations)(nil)

	errTxSizeTooBig    = errors.New("batch tx size too big")
	errSkipConfirmedTx = errors.New("skip confirmed tx")
)

// Get real tx size from protocolParams/config
const (
	maxTxSize = 16000
)

type batchInitialData struct {
	BatchNonceID   uint64
	Metadata       []byte
	ProtocolParams []byte
	ChainID        uint8
}

type CardanoChainOperations struct {
	config                       *cardano.CardanoChainConfig
	wallet                       *cardano.ApexCardanoWallet
	txProvider                   cardanowallet.ITxDataRetriever
	db                           indexer.Database
	gasLimiter                   eth.GasLimitHolder
	cardanoCliBinary             string
	bridgingAddressesManager     common.BridgingAddressesManager
	bridgingAddressesCoordinator common.BridgingAddressesCoordinator
	logger                       hclog.Logger
}

func NewCardanoChainOperations(
	jsonConfig json.RawMessage,
	db indexer.Database,
	secretsManager secrets.SecretsManager,
	chainID string,
	bridgingAddressesManager common.BridgingAddressesManager,
	bridgingAddressesCoordinator common.BridgingAddressesCoordinator,
	logger hclog.Logger,
) (*CardanoChainOperations, error) {
	cardanoConfig, err := cardano.NewCardanoChainConfig(jsonConfig)
	if err != nil {
		return nil, err
	}

	txProvider, err := cardanoConfig.CreateTxProvider()
	if err != nil {
		return nil, fmt.Errorf("failed to create tx provider: %w", err)
	}

	cardanoWallet, err := cardano.LoadWallet(secretsManager, chainID)
	if err != nil {
		return nil, err
	}

	return &CardanoChainOperations{
		wallet:           cardanoWallet,
		config:           cardanoConfig,
		txProvider:       txProvider,
		cardanoCliBinary: cardanowallet.ResolveCardanoCliBinary(cardanoConfig.NetworkID),
		gasLimiter: eth.NewGasLimitHolder(submitBatchMinGasLimit,
			submitBatchMaxGasLimit, submitBatchStepsGasLimit),
		db:                           db,
		logger:                       logger,
		bridgingAddressesManager:     bridgingAddressesManager,
		bridgingAddressesCoordinator: bridgingAddressesCoordinator,
	}, nil
}

// GenerateBatchTransaction implements core.ChainOperations.
func (cco *CardanoChainOperations) GenerateBatchTransaction(
	ctx context.Context,
	chainID string,
	confirmedTransactions []eth.ConfirmedTransaction,
	batchNonceID uint64,
) (*core.GeneratedBatchTxData, error) {
	data, err := cco.createBatchInitialData(ctx, chainID, batchNonceID)
	if err != nil {
		return nil, err
	}

	txData, chosenMultisigAddresses, err := cco.generateBatchTransaction(data, confirmedTransactions)

	if cco.shouldConsolidate(err) {
		cco.logger.Warn("consolidation batch generation started", "err", err)

		txData, err = cco.generateConsolidationTransaction(data, chosenMultisigAddresses)
		if err != nil {
			err = fmt.Errorf("consolidation batch failed: %w", err)
		}
	}

	return txData, err
}

// SignBatchTransaction implements core.ChainOperations.
func (cco *CardanoChainOperations) SignBatchTransaction(
	generatedBatchData *core.GeneratedBatchTxData) (*core.BatchSignatures, error) {
	txBuilder, err := cardanowallet.NewTxBuilder(cco.cardanoCliBinary)
	if err != nil {
		return nil, err
	}

	defer txBuilder.Dispose()

	var stakeMultisigWitness []byte

	var paymentMultisigWitness []byte

	if generatedBatchData.IsPaymentSignNeeded {
		paymentMultisigWitness, err = txBuilder.CreateTxWitness(generatedBatchData.TxRaw, cco.wallet.MultiSig)
		if err != nil {
			return nil, err
		}
	}

	if generatedBatchData.IsStakeSignNeeded {
		stakeMultisigWitness, err = txBuilder.CreateTxWitness(
			generatedBatchData.TxRaw, cardanowallet.NewStakeSigner(cco.wallet.MultiSig))
		if err != nil {
			return nil, err
		}
	}

	feeWitness, err := txBuilder.CreateTxWitness(generatedBatchData.TxRaw, cco.wallet.Fee)
	if err != nil {
		return nil, err
	}

	return &core.BatchSignatures{
		Multisig:     paymentMultisigWitness,
		MultsigStake: stakeMultisigWitness,
		Fee:          feeWitness,
	}, nil
}

// IsSynchronized implements core.IsSynchronized.
func (cco *CardanoChainOperations) IsSynchronized(
	ctx context.Context, bridgeSmartContract eth.IBridgeSmartContract, chainID string,
) (bool, error) {
	lastObservedBlockBridge, err := bridgeSmartContract.GetLastObservedBlock(ctx, chainID)
	if err != nil {
		return false, err
	}

	lastOracleBlockPoint, err := cco.db.GetLatestBlockPoint()
	if err != nil {
		return false, err
	}

	return lastOracleBlockPoint != nil &&
		lastOracleBlockPoint.BlockSlot >= lastObservedBlockBridge.BlockSlot.Uint64(), nil
}

// Submit implements core.Submit.
func (cco *CardanoChainOperations) Submit(
	ctx context.Context, bridgeSmartContract eth.IBridgeSmartContract, batch eth.SignedBatch,
) error {
	err := bridgeSmartContract.SubmitSignedBatch(ctx, batch, cco.gasLimiter.GetGasLimit())

	cco.gasLimiter.Update(err)

	return err
}

func (cco *CardanoChainOperations) generateBatchTransaction(
	data *batchInitialData,
	confirmedTransactions []eth.ConfirmedTransaction,
) (
	*core.GeneratedBatchTxData,
	[]common.AddressAndAmount,
	error,
) {
	certificates := make([]*cardano.CertificatesWithScript, 0)
	keyRegistrationFee := uint64(0)
	hasStakeDelegationTx := false
	hasBridgingTx := false

	for _, tx := range confirmedTransactions {
		if tx.TransactionType == uint8(common.StakeDelConfirmedTxType) {
			certificate, depositAmount, err := cco.getStakingDelegateCertificate(data, &tx)
			if err != nil {
				return nil, nil, err
			}

			hasStakeDelegationTx = true
			keyRegistrationFee += depositAmount

			certificates = append(certificates, certificate)
		}

		if tx.TransactionType != uint8(common.StakeDelConfirmedTxType) {
			hasBridgingTx = true
		}
	}

	var certificateData *cardano.CertificatesData = nil

	if len(certificates) > 0 {
		certificateData = &cardano.CertificatesData{
			Certificates:    certificates,
			RegistrationFee: keyRegistrationFee,
		}
	}

	txOutputs, err := getOutputs(confirmedTransactions, cco.config, cco.logger)
	if err != nil {
		return nil, nil, err
	}

	cco.logger.Debug("Getting addresses and amounts to pay from", "outputs", txOutputs.Outputs)

	multisigAddresses, err := cco.bridgingAddressesCoordinator.GetAddressesAndAmountsToPayFrom(
		data.ChainID,
		cco.cardanoCliBinary,
		data.ProtocolParams,
		txOutputs.Outputs,
	)
	if err != nil {
		return nil, nil, err
	}

	cco.logger.Debug("Chosen multisig addresses", "addresses", multisigAddresses)

	feeMultisigAddress := cco.bridgingAddressesManager.GetFeeMultisigAddress(data.ChainID)

	multisigUtxos, feeUtxos, err := cco.getUTXOsForNormalBatch(
		multisigAddresses, feeMultisigAddress, data)
	if err != nil {
		return nil, multisigAddresses, err
	}

	slotNumber, err := cco.getSlotNumber()
	if err != nil {
		return nil, multisigAddresses, err
	}

	cco.logger.Info("Creating batch tx", "batchID", data.BatchNonceID,
		"magic", cco.config.NetworkMagic, "binary", cco.cardanoCliBinary,
		"slot", slotNumber, "multisig", len(multisigUtxos), "fee", len(feeUtxos), "outputs", len(txOutputs.Outputs))

	txInputs := cardano.TxInputInfos{}

	for addressIndex, utxos := range multisigUtxos {
		policyScript, ok := cco.bridgingAddressesManager.GetPaymentPolicyScript(data.ChainID, addressIndex)
		if !ok {
			return nil, multisigAddresses, fmt.Errorf("failed to get payment policy script for address: %d", addressIndex)
		}

		txInputs.MultiSig = append(txInputs.MultiSig, &cardano.TxInputInfo{
			PolicyScript: policyScript,
			Address:      multisigAddresses[addressIndex].Address,
			TxInputs:     convertUTXOsToTxInputs(utxos),
		})
	}

	feePolicyScript, ok := cco.bridgingAddressesManager.GetFeeMultisigPolicyScript(data.ChainID)
	if !ok {
		return nil, multisigAddresses, fmt.Errorf("failed to get fee policy script for chain: %d", data.ChainID)
	}

	txInputs.MultiSigFee = &cardano.TxInputInfo{
		PolicyScript: feePolicyScript,
		Address:      feeMultisigAddress,
		TxInputs:     convertUTXOsToTxInputs(feeUtxos),
	}

	// Create Tx
	txRaw, txHash, err := cardano.CreateTx(
		cco.cardanoCliBinary,
		uint(cco.config.NetworkMagic),
		data.ProtocolParams,
		slotNumber+cco.config.TTLSlotNumberInc,
		data.Metadata,
		txInputs,
		txOutputs.Outputs,
		certificateData,
	)
	if err != nil {
		return nil, multisigAddresses, err
	}

	if len(txRaw) > maxTxSize {
		return nil, multisigAddresses, fmt.Errorf("%w: (size, max) = (%d, %d)",
			errTxSizeTooBig, len(txRaw), maxTxSize)
	}

	return &core.GeneratedBatchTxData{
		TxRaw:               txRaw,
		TxHash:              txHash,
		IsStakeSignNeeded:   hasStakeDelegationTx,
		IsPaymentSignNeeded: hasBridgingTx,
		BatchType:           eth.BatchTypeNormal,
	}, multisigAddresses, nil
}

func (cco *CardanoChainOperations) shouldConsolidate(err error) bool {
	return errors.Is(err, cardanowallet.ErrUTXOsLimitReached) || errors.Is(err, errTxSizeTooBig)
}

func (cco *CardanoChainOperations) generateConsolidationTransaction(
	data *batchInitialData,
	chosenMultisigAddresses []common.AddressAndAmount,
) (*core.GeneratedBatchTxData, error) {
	feeMultisigAddress := cco.bridgingAddressesManager.GetFeeMultisigAddress(data.ChainID)
	multisigUtxos, feeUtxos, err := cco.getUTXOsForConsolidation(chosenMultisigAddresses, feeMultisigAddress)
	if err != nil {
		return nil, err
	}

	multisigTxOutputs := make([]cardanowallet.TxOutput, 0, len(chosenMultisigAddresses))

	for _, addressAndAmount := range chosenMultisigAddresses {
		multisigTxOutput, err := getTxOutputFromSumMap(
			addressAndAmount.Address,
			getSumMapFromTxInputOutput(multisigUtxos[addressAndAmount.AddressIndex]),
		)
		if err != nil {
			return nil, err
		}

		multisigTxOutputs = append(multisigTxOutputs, multisigTxOutput)
	}

	slotNumber, err := cco.getSlotNumber()
	if err != nil {
		return nil, err
	}

	cco.logger.Info("Creating consolidation tx", "consolidationTxID", data.BatchNonceID,
		"magic", cco.config.NetworkMagic, "binary", cco.cardanoCliBinary,
		"slot", slotNumber, "multisig", len(multisigUtxos), "fee", len(feeUtxos))

	// Generate tx inputs
	feePolicyScript, ok := cco.bridgingAddressesManager.GetFeeMultisigPolicyScript(data.ChainID)
	if !ok {
		return nil, fmt.Errorf("failed to get fee policy script for chain: %d", data.ChainID)
	}

	txInputs := cardano.TxInputInfos{
		MultiSig: make([]*cardano.TxInputInfo, 0, len(chosenMultisigAddresses)),
		MultiSigFee: &cardano.TxInputInfo{
			PolicyScript: feePolicyScript,
			Address:      feeMultisigAddress,
			TxInputs:     convertUTXOsToTxInputs(feeUtxos),
		},
	}

	for _, addressAndAmount := range chosenMultisigAddresses {
		policyScript, ok := cco.bridgingAddressesManager.GetPaymentPolicyScript(data.ChainID, addressAndAmount.AddressIndex)
		if !ok {
			return nil, fmt.Errorf("failed to get payment policy script for address: %d", addressAndAmount.AddressIndex)
		}

		txInputs.MultiSig = append(txInputs.MultiSig, &cardano.TxInputInfo{
			PolicyScript: policyScript,
			Address:      addressAndAmount.Address,
			TxInputs:     convertUTXOsToTxInputs(multisigUtxos[addressAndAmount.AddressIndex]),
		})
	}

	// Create Tx
	txRaw, txHash, err := cardano.CreateTx(
		cco.cardanoCliBinary,
		uint(cco.config.NetworkMagic),
		data.ProtocolParams,
		slotNumber+cco.config.TTLSlotNumberInc,
		data.Metadata,
		txInputs,
		multisigTxOutputs,
		nil,
	)
	if err != nil {
		return nil, err
	}

	if len(txRaw) > maxTxSize {
		return nil, fmt.Errorf("%w: (size, max) = (%d, %d)", errTxSizeTooBig, len(txRaw), maxTxSize)
	}

	return &core.GeneratedBatchTxData{
		BatchType:           eth.BatchTypeConsolidation,
		TxRaw:               txRaw,
		TxHash:              txHash,
		IsPaymentSignNeeded: true,
	}, nil
}

func (cco *CardanoChainOperations) getUTXOsForConsolidation(
	chosenMultisigAddresses []common.AddressAndAmount,
	multisigFeeAddress string,
) (map[uint8][]*indexer.TxInputOutput, []*indexer.TxInputOutput, error) {
	feeUtxos, err := cco.db.GetAllTxOutputs(multisigFeeAddress, true)
	if err != nil {
		return nil, nil, err
	}

	feeUtxos = filterOutUtxosWithUnknownTokens(feeUtxos)

	sort.Slice(feeUtxos, func(i, j int) bool {
		return feeUtxos[i].Output.Amount < feeUtxos[j].Output.Amount
	})

	if len(feeUtxos) == 0 {
		return nil, nil, fmt.Errorf("fee multisig does not have any utxo: %s", multisigFeeAddress)
	}

	consolidationInputs := make([]AddressConsolidationData, len(chosenMultisigAddresses)+1)
	consolidationInputs[0] = AddressConsolidationData{
		Address:      multisigFeeAddress,
		AddressIndex: 0,
		UtxoCount:    len(feeUtxos),
		IsFee:        true,
		Utxos:        feeUtxos,
	}

	knownTokens, err := cardano.GetKnownTokens(cco.config)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get known tokens: %w", err)
	}

	queriedMultisigUtxos := make(map[uint8][]*indexer.TxInputOutput)

	for i, addressAndAmount := range chosenMultisigAddresses {
		multisigUtxos, err := cco.db.GetAllTxOutputs(addressAndAmount.Address, true)
		if err != nil {
			return nil, nil, err
		}

		multisigUtxos = filterOutUtxosWithUnknownTokens(multisigUtxos, knownTokens...)

		sort.Slice(multisigUtxos, func(i, j int) bool {
			return multisigUtxos[i].Output.Amount < multisigUtxos[j].Output.Amount
		})

		queriedMultisigUtxos[addressAndAmount.AddressIndex] = multisigUtxos

		consolidationInputs[i+1] = AddressConsolidationData{
			Address:      addressAndAmount.Address,
			AddressIndex: addressAndAmount.AddressIndex,
			UtxoCount:    len(multisigUtxos),
		}
	}

	cco.logger.Debug("UTXOs retrieved",
		"multisig", queriedMultisigUtxos, "fee", multisigFeeAddress, "utxos", feeUtxos)

	consolidationChosenInputs := allocateInputsForConsolidation(
		consolidationInputs, int(cco.config.MaxUtxoCount)) //nolint:gosec
	cco.logger.Debug("Consolidation chosen inputs", "max", cco.config.MaxUtxoCount,
		"inputs", consolidationInputs, "chosen", consolidationChosenInputs)

	multisigUtxos := make(map[uint8][]*indexer.TxInputOutput)

	for _, input := range consolidationChosenInputs {
		if input.Address == multisigFeeAddress {
			feeUtxos = feeUtxos[:input.UtxoCount]
		} else {
			multisigUtxos[input.AddressIndex] = queriedMultisigUtxos[input.AddressIndex][:input.UtxoCount]
		}
	}

	cco.logger.Debug("UTXOs chosen", "multisig", multisigUtxos, "fee", feeUtxos)

	return multisigUtxos, feeUtxos, nil
}

func (cco *CardanoChainOperations) getUTXOsForNormalBatch(
	multisigAddresses []common.AddressAndAmount, multisigFeeAddress string, data *batchInitialData,
) (map[uint8][]*indexer.TxInputOutput, []*indexer.TxInputOutput, error) {
	feeUtxos, err := cco.getFeeUTXOsForNormalBatch(multisigFeeAddress)
	if err != nil {
		return nil, nil, err
	}

	knownTokens, err := cardano.GetKnownTokens(cco.config)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get known tokens: %w", err)
	}

	chosenMultisigUtxos := make(map[uint8][]*indexer.TxInputOutput)

	for _, addressAndAmount := range multisigAddresses {
		multisigUtxos, err := cco.db.GetAllTxOutputs(addressAndAmount.Address, true)
		if err != nil {
			return nil, nil, err
		}

		multisigUtxos = filterOutUtxosWithUnknownTokens(multisigUtxos, knownTokens...)

		cco.logger.Debug("UTXOs retrieved",
			"multisig", addressAndAmount.Address, "utxos", multisigUtxos)

		// Create output matching the address and amount
		output := []cardanowallet.TxOutput{
			{
				Addr:   addressAndAmount.Address,
				Amount: addressAndAmount.TokensAmounts[cardanowallet.AdaTokenName],
			},
		}

		for policyID, token := range addressAndAmount.TokensAmounts {
			tokenName := ""

			for _, token := range knownTokens {
				if token.PolicyID == policyID {
					tokenName = token.Name

					break
				}
			}

			output[0].Tokens = append(output[0].Tokens, cardanowallet.TokenAmount{
				Token:  cardanowallet.NewToken(policyID, tokenName),
				Amount: token,
			})
		}

		cco.logger.Debug("Output for calculateMinUtxoLovelaceAmount", "output", output)

		minUtxoLovelaceAmount := uint64(0)
		if !addressAndAmount.FullAmount {
			minUtxoLovelaceAmount, err = calculateMinUtxoLovelaceAmount(
				cco.cardanoCliBinary, data.ProtocolParams, addressAndAmount.Address, multisigUtxos, output)
			if err != nil {
				return nil, nil, err
			}
		}

		cco.logger.Debug("Min Utxo Lovelace Amount", "minUtxoLovelaceAmount", minUtxoLovelaceAmount)

		multisigUtxos, err = getNeededUtxos(
			multisigUtxos,
			addressAndAmount.TokensAmounts,
			minUtxoLovelaceAmount,
			getMaxUtxoCount(cco.config, len(feeUtxos)),
			int(cco.config.TakeAtLeastUtxoCount), //nolint:gosec
		)
		if err != nil {
			return nil, nil, err
		}

		cco.logger.Debug("UTXOs chosen", "multisig", multisigUtxos)
		chosenMultisigUtxos[addressAndAmount.AddressIndex] = multisigUtxos
	}

	return chosenMultisigUtxos, feeUtxos, nil
}

func (cco *CardanoChainOperations) getFeeUTXOsForNormalBatch(
	multisigFeeAddress string,
) ([]*indexer.TxInputOutput, error) {
	feeUtxos, err := cco.db.GetAllTxOutputs(multisigFeeAddress, true)
	if err != nil {
		return nil, err
	}

	feeUtxos = filterOutUtxosWithUnknownTokens(feeUtxos)

	if len(feeUtxos) == 0 {
		return nil, fmt.Errorf("fee multisig does not have any utxo: %s", multisigFeeAddress)
	}

	cco.logger.Debug("Fee UTXOs retrieved",
		"fee address", multisigFeeAddress, "utxos", feeUtxos)

	feeUtxos = feeUtxos[:min(cco.config.MaxFeeUtxoCount, uint(len(feeUtxos)))] // do not take more than MaxFeeUtxoCount

	return feeUtxos, nil
}

func (cco *CardanoChainOperations) getSlotNumber() (uint64, error) {
	data, err := cco.db.GetLatestBlockPoint()
	if err != nil {
		return 0, err
	}

	slot := uint64(0)
	if data != nil {
		slot = data.BlockSlot
	}

	newSlot, err := getNumberWithRoundingThreshold(
		slot, cco.config.SlotRoundingThreshold, cco.config.NoBatchPeriodPercent)
	if err != nil {
		return 0, err
	}

	cco.logger.Debug("calculate slotNumber with rounding", "slot", slot, "newSlot", newSlot)

	return newSlot, nil
}

func (cco *CardanoChainOperations) createBatchInitialData(
	ctx context.Context,
	chainID string,
	batchNonceID uint64,
) (*batchInitialData, error) {
	metadata, err := common.MarshalMetadata(common.MetadataEncodingTypeJSON, common.BatchExecutedMetadata{
		BridgingTxType: common.BridgingTxTypeBatchExecution,
		BatchNonceID:   batchNonceID,
	})
	if err != nil {
		return nil, err
	}

	protocolParams, err := cco.txProvider.GetProtocolParameters(ctx)
	if err != nil {
		return nil, err
	}

	return &batchInitialData{
		BatchNonceID:   batchNonceID,
		Metadata:       metadata,
		ProtocolParams: protocolParams,
		ChainID:        common.ToNumChainID(chainID),
	}, nil
}

func (cco *CardanoChainOperations) getStakingDelegateCertificate(
	data *batchInitialData, tx *eth.ConfirmedTransaction,
) (*cardano.CertificatesWithScript, uint64, error) {
	// Generate certificates
	keyRegDepositAmount, err := extractStakeKeyDepositAmount(data.ProtocolParams)
	if err != nil {
		return nil, 0, err
	}

	cliUtils := cardanowallet.NewCliUtils(cco.cardanoCliBinary)

	policyScript, ok := cco.bridgingAddressesManager.GetStakePolicyScript(data.ChainID, tx.BridgeAddrIndex)
	if !ok {
		return nil, 0, fmt.Errorf("failed to get stake policy script for address: %d", tx.BridgeAddrIndex)
	}

	multisigStakeAddress, ok := cco.bridgingAddressesManager.GetStakeAddressFromIndex(data.ChainID, tx.BridgeAddrIndex)
	if !ok {
		return nil, 0, fmt.Errorf("failed to get stake address from index: %d", tx.BridgeAddrIndex)
	}

	registrationCert, err := cliUtils.CreateRegistrationCertificate(multisigStakeAddress, keyRegDepositAmount)
	if err != nil {
		return nil, 0, err
	}

	delegationCert, err := cliUtils.CreateDelegationCertificate(multisigStakeAddress, tx.StakePoolId)
	if err != nil {
		return nil, 0, err
	}

	return &cardano.CertificatesWithScript{
		PolicyScript: policyScript,
		Certificates: []cardanowallet.ICertificate{registrationCert, delegationCert},
	}, keyRegDepositAmount, nil
}

func getMaxUtxoCount(config *cardano.CardanoChainConfig, prevUtxosCnt int) int {
	return max(int(config.MaxUtxoCount)-prevUtxosCnt, 0) //nolint:gosec
}
