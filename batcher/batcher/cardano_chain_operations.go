package batcher

import (
	"context"
	"encoding/hex"
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

type utxoSelectionResult struct {
	multisigUtxos            map[uint8][]*indexer.TxInputOutput
	feeUtxos                 []*indexer.TxInputOutput
	chosenMultisigUtxosCount int
}

type refundUtxoWithReceiverAddr struct {
	Utxo *indexer.TxInputOutput
	Addr string
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

	txData, chosenMultisigAddresses, err := cco.generateBatchTransaction(ctx, data, confirmedTransactions)

	if cco.shouldConsolidate(err) {
		consolidationType := cco.getConsolidationType(err)
		cco.logger.Warn("consolidation batch generation started", "err", err, "consolidationType", consolidationType.String())

		txData, err = cco.generateConsolidationTransaction(ctx, data, chosenMultisigAddresses, consolidationType)
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
	ctx context.Context,
	data *batchInitialData,
	confirmedTransactions []eth.ConfirmedTransaction,
) (
	*core.GeneratedBatchTxData,
	[]common.AddressAndAmount,
	error,
) {
	certificateData, hasBridgingTx, err := cco.getCertificateData(data, confirmedTransactions)
	if err != nil {
		return nil, nil, err
	}

	feeMultisigAddress := cco.bridgingAddressesManager.GetFeeMultisigAddress(data.ChainID)

	refundUtxosPerConfirmedTx, err :=
		cco.getUtxosFromRefundTransactions(confirmedTransactions)
	if err != nil {
		return nil, nil, err
	}

	getOutputsData, err := getOutputs(
		confirmedTransactions, cco.config, feeMultisigAddress, refundUtxosPerConfirmedTx, cco.logger)
	if err != nil {
		return nil, nil, err
	}

	txPlutusMintData, err := cco.getPlutusMintData(ctx, data, getOutputsData.MintTokens)
	if err != nil {
		return nil, nil, err
	}

	cco.logger.Debug("Mint tokens data", txPlutusMintData)

	cco.logger.Debug("Getting addresses and amounts", "chain", common.ToStrChainID(data.ChainID),
		"outputs", getOutputsData.TxOutputs.Outputs, "redistribution", getOutputsData.IsRedistribution)

	var mintTokens []cardanowallet.MintTokenAmount
	if txPlutusMintData != nil {
		mintTokens = txPlutusMintData.Tokens
	}

	multisigAddresses, isRedistribution, err := cco.bridgingAddressesCoordinator.GetAddressesAndAmountsForBatch(
		data.ChainID,
		cco.cardanoCliBinary,
		getOutputsData.IsRedistribution,
		data.ProtocolParams,
		getOutputsData.TxOutputs,
		mintTokens,
	)
	if err != nil {
		return nil, multisigAddresses, err
	}

	cco.logger.Debug("Chosen multisig addresses to pay from",
		"chain", common.ToStrChainID(data.ChainID), "addresses", multisigAddresses)

	alreadyUsed := map[string]struct{}{}
	refundUtxos := make([]refundUtxoWithReceiverAddr, 0, len(refundUtxosPerConfirmedTx))

	for i, refundUtxosForConfirmedTx := range refundUtxosPerConfirmedTx {
		for _, utxo := range refundUtxosForConfirmedTx {
			key := utxo.Input.String()

			if _, exists := alreadyUsed[key]; exists {
				cco.logger.Warn("refund utxo already",
					"chainID", data.ChainID, "batchID", data.BatchNonceID, "input", key)

				continue
			}

			alreadyUsed[key] = struct{}{}

			refundUtxos = append(refundUtxos, refundUtxoWithReceiverAddr{
				Utxo: utxo,
				Addr: confirmedTransactions[i].Receivers[0].DestinationAddress,
			})
		}
	}

	utxoSelectionResult, err := cco.getUTXOsForNormalBatch(
		multisigAddresses, feeMultisigAddress, isRedistribution, len(refundUtxos))
	if err != nil {
		return nil, multisigAddresses, err
	}

	refundUtxoInputs, refundUtxoOutputs, err :=
		cco.prepareRefundInputsOutputs(data.ChainID, refundUtxos)
	if err != nil {
		return nil, multisigAddresses, fmt.Errorf("failed to prepare refund inputs/outputs. err: %w", err)
	}

	getOutputsData.TxOutputs.Outputs = append(getOutputsData.TxOutputs.Outputs, refundUtxoOutputs...)

	slotNumber, err := cco.getSlotNumber()
	if err != nil {
		return nil, multisigAddresses, err
	}

	cco.logger.Info("Creating batch tx", "batchID", data.BatchNonceID,
		"magic", cco.config.NetworkMagic, "binary", cco.cardanoCliBinary,
		"slot", slotNumber, "multisig", utxoSelectionResult.chosenMultisigUtxosCount+len(refundUtxoInputs),
		"fee", len(utxoSelectionResult.feeUtxos), "outputs", len(getOutputsData.TxOutputs.Outputs))

	refundTxInputs, err := cco.prepareMultisigInputsForNormalBatch(data.ChainID, refundUtxoInputs)
	if err != nil {
		return nil, multisigAddresses, fmt.Errorf("failed to prepare refund inputs for multisig: %w", err)
	}

	txInputs, err := cco.prepareMultisigInputsForNormalBatch(
		data.ChainID, utxoSelectionResult.multisigUtxos)
	if err != nil {
		return nil, multisigAddresses, fmt.Errorf("failed to prepare inputs for multisig: %w", err)
	}

	if len(mintTokens) > 0 {
		err := cco.prepareCustodialInputsForNormalBatch(data.ChainID, txInputs, getOutputsData)
		if err != nil {
			return nil, multisigAddresses, fmt.Errorf("failed to prepare custodial inputs and outputs: %w", err)
		}
	}

	feePolicyScript, ok := cco.bridgingAddressesManager.GetFeeMultisigPolicyScript(data.ChainID)
	if !ok {
		return nil, multisigAddresses, fmt.Errorf("failed to get fee policy script for chain: %d", data.ChainID)
	}

	txInputs.MultiSigFee = &cardano.TxInputInfo{
		PolicyScript: feePolicyScript,
		Address:      feeMultisigAddress,
		TxInputs:     convertUTXOsToTxInputs(utxoSelectionResult.feeUtxos),
	}

	var addrAndAmountToDeduct []common.AddressAndAmount

	if isRedistribution {
		getOutputsData.TxOutputs.Outputs, err = addRedistributionOutputs(getOutputsData.TxOutputs.Outputs, multisigAddresses)
		if err != nil {
			return nil, multisigAddresses, err
		}
	} else {
		addrAndAmountToDeduct = multisigAddresses
	}

	additionalWitnessCount := 0

	if cco.config.RelayerAddress != "" {
		additionalWitnessCount = 1
	}

	cco.logger.Debug("TX INPUTS", "batchID", data.BatchNonceID,
		"chain", common.ToStrChainID(data.ChainID), "txInputs", txInputs)
	cco.logger.Debug("REFUND TX INPUTS", "batchID", data.BatchNonceID,
		"chain", common.ToStrChainID(data.ChainID), "refundTxInputs", refundTxInputs.MultiSig)
	cco.logger.Debug("TX OUTPUTS", "batchID", data.BatchNonceID,
		"chain", common.ToStrChainID(data.ChainID), "txOutputs.Outputs", getOutputsData.TxOutputs.Outputs)

	// Create Tx
	txRaw, txHash, err := cardano.CreateTx(
		ctx,
		cco.cardanoCliBinary,
		uint(cco.config.NetworkMagic),
		data.ProtocolParams,
		slotNumber+cco.config.TTLSlotNumberInc,
		data.Metadata,
		txInputs,
		refundTxInputs.MultiSig,
		getOutputsData.TxOutputs.Outputs,
		certificateData,
		addrAndAmountToDeduct,
		txPlutusMintData,
		additionalWitnessCount,
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
		IsStakeSignNeeded:   certificateData != nil,
		IsPaymentSignNeeded: hasBridgingTx,
		BatchType:           eth.BatchTypeNormal,
	}, multisigAddresses, nil
}

func (cco *CardanoChainOperations) getPlutusMintData(
	ctx context.Context,
	data *batchInitialData, mintTokens []cardanowallet.MintTokenAmount,
) (*cardano.PlutusMintData, error) {
	if len(mintTokens) == 0 {
		return nil, nil
	}

	tokens := make([]cardanowallet.MintTokenAmount, 0, len(mintTokens))
	tokensPolicyID := ""

	// Populate map of locked tokens from the addr0
	addr0Address, ok := cco.bridgingAddressesManager.GetPaymentAddressFromIndex(data.ChainID, 0)
	if !ok {
		return nil, fmt.Errorf("failed to get address 0 for chain %s", common.ToStrChainID(data.ChainID))
	}

	addr0Utxos, err := cco.db.GetAllTxOutputs(addr0Address, true)
	if err != nil {
		return nil, fmt.Errorf("failed to get utxos from address %s", addr0Address)
	}

	availableLockedTokens := cardano.GetSumMapFromTxInputOutput(addr0Utxos)

	for _, mintToken := range mintTokens {
		tokensPolicyID = mintToken.PolicyID

		if mintToken.Amount >= availableLockedTokens[mintToken.String()] {
			tokens = append(tokens, cardanowallet.NewMintTokenAmount(
				mintToken.Token, mintToken.Amount-availableLockedTokens[mintToken.String()], false))
		} else {
			tokens = append(tokens, cardanowallet.NewMintTokenAmount(
				mintToken.Token, availableLockedTokens[mintToken.String()]-mintToken.Amount, true))
		}
	}

	relayerAddr := cco.config.RelayerAddress

	relayerUtxos, err := cco.txProvider.GetUtxos(ctx, relayerAddr) // cco.db.GetAllTxOutputs(relayerAddr, true)
	if err != nil {
		return nil, fmt.Errorf("failed to get utxos from address %s", relayerAddr)
	}

	// filter out unknown tokens
	filteredUtxos := make([]cardanowallet.Utxo, 0, len(relayerUtxos))

	for _, utxo := range relayerUtxos {
		if len(utxo.Tokens) == 0 {
			filteredUtxos = append(filteredUtxos, utxo)
		}
	}

	sort.SliceStable(filteredUtxos, func(i, j int) bool {
		return relayerUtxos[i].Amount > relayerUtxos[j].Amount
	})

	relayerTxInputs := make([]cardanowallet.TxInput, 1)

	// take largest utxo as collateral
	relayerTxInputs[0] = cardanowallet.TxInput{
		Hash:  filteredUtxos[0].Hash,
		Index: filteredUtxos[0].Index,
	}

	sum := map[string]uint64{
		cardanowallet.AdaTokenName: filteredUtxos[0].Amount,
	}

	cco.logger.Debug("Relayer tx inputs", "txInputs", relayerTxInputs, "sum", sum)

	executionUnitData, err := extractPlutusExecutionParams(data.ProtocolParams)
	if err != nil {
		return nil, fmt.Errorf("failed to extract plutus execution params: %w", err)
	}

	return &cardano.PlutusMintData{
		Tokens:        tokens,
		TxInReference: *cco.config.MintingScriptTxInput,
		Collateral: cardanowallet.TxInputs{
			Inputs: relayerTxInputs,
			Sum:    sum,
		},
		CollateralAddress: relayerAddr,
		TokensPolicyID:    tokensPolicyID,
		TxProvider:        cco.txProvider,
		ExecutionUnitData: executionUnitData,
	}, nil
}

func (cco *CardanoChainOperations) prepareRefundInputsOutputs(
	chainID uint8,
	refundUtxos []refundUtxoWithReceiverAddr,
) (
	map[uint8][]*indexer.TxInputOutput,
	[]cardanowallet.TxOutput,
	error,
) {
	refundUtxoInputs := make(map[uint8][]*indexer.TxInputOutput)
	refundUtxoOutputs := make([]cardanowallet.TxOutput, 0, 2*len(refundUtxos))

	// process refund utxos by adding them into inputs, and creating outputs
	for _, obj := range refundUtxos {
		index, ok := cco.bridgingAddressesManager.GetPaymentAddressIndex(
			chainID, obj.Utxo.Output.Address)
		if !ok {
			return nil, nil, fmt.Errorf("failed to get index for address %s on chain %s",
				obj.Utxo.Output.Address, common.ToStrChainID(chainID))
		}

		refundUtxoInputs[index] = append(refundUtxoInputs[index], obj.Utxo)

		tokens := make([]cardanowallet.TokenAmount, len(obj.Utxo.Output.Tokens))
		for i, t := range obj.Utxo.Output.Tokens {
			tokens[i] = cardanowallet.NewTokenAmount(
				cardanowallet.NewToken(t.PolicyID, t.Name),
				t.Amount,
			)
		}

		// using cco.Config.MinFeeForBridgingTokens here without checking if there are native tokens
		// because this is a special refund case that only happens when someone sends unknown tokens
		// to the bridging address. So here we assume that there are tokens to be returned
		refundUtxoOutputs = append(refundUtxoOutputs, cardanowallet.TxOutput{
			Addr:   obj.Addr,
			Amount: obj.Utxo.Output.Amount - cco.config.MinFeeForBridgingTokens,
			Tokens: tokens,
		})

		refundUtxoOutputs = append(refundUtxoOutputs, cardanowallet.TxOutput{
			Addr:   cco.bridgingAddressesManager.GetFeeMultisigAddress(chainID),
			Amount: cco.config.MinFeeForBridgingTokens,
		})
	}

	return refundUtxoInputs, refundUtxoOutputs, nil
}

func (cco *CardanoChainOperations) prepareMultisigInputsForNormalBatch(
	chainID uint8, multisigUtxos map[uint8][]*indexer.TxInputOutput,
) (*cardano.TxInputInfos, error) {
	txInputs := &cardano.TxInputInfos{}

	for addressIndex, utxos := range multisigUtxos {
		policyScript, ok := cco.bridgingAddressesManager.GetPaymentPolicyScript(chainID, addressIndex)
		if !ok {
			return nil, fmt.Errorf("failed to get payment policy script for address index: %d", addressIndex)
		}

		if addr, ok := cco.bridgingAddressesManager.GetPaymentAddressFromIndex(chainID, addressIndex); !ok {
			return nil, fmt.Errorf("failed to get payment address for address index: %d", addressIndex)
		} else {
			txInputs.MultiSig = append(txInputs.MultiSig, &cardano.TxInputInfo{
				PolicyScript: policyScript,
				Address:      addr,
				TxInputs:     convertUTXOsToTxInputs(utxos),
			})
		}
	}

	return txInputs, nil
}

func (cco *CardanoChainOperations) findCustodialTxOutput(custodialAddr string) (*indexer.TxInputOutput, error) {
	txOutputs, err := cco.db.GetAllTxOutputs(custodialAddr, true)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve UTXOs for address %s: %w", custodialAddr, err)
	}

	for _, out := range txOutputs {
		tokens := out.Output.Tokens
		// nft token has to be the only token in utxo by design
		if len(tokens) == 1 &&
			tokens[0].Equals(indexer.TokenAmount{
				PolicyID: cco.config.CustodialNft.PolicyID,
				Name:     cco.config.CustodialNft.Name,
				Amount:   1,
			}) {
			return out, nil
		}
	}

	return nil, fmt.Errorf("no custodial UTXO containing the expected NFT found at address %s", custodialAddr)
}

func (cco *CardanoChainOperations) prepareCustodialInputsForNormalBatch(
	chainID uint8, txInputs *cardano.TxInputInfos, getOutputsData *getOutputsData,
) error {
	custodialAddr, ok := cco.bridgingAddressesManager.GetCustodialAddress(chainID)
	if !ok {
		return fmt.Errorf("failed to get custodial address for chain: %s", common.ToStrChainID(chainID))
	}

	custodialPolicyScript, ok := cco.bridgingAddressesManager.GetCustodialPolicyScript(chainID)
	if !ok {
		return fmt.Errorf("failed to get custodial policy script for chain: %s", common.ToStrChainID(chainID))
	}

	custodialTxOutput, err := cco.findCustodialTxOutput(custodialAddr)
	if err != nil {
		return err
	}

	txInputs.Custodial = &cardano.TxInputInfo{
		TxInputs:     convertUTXOsToTxInputs([]*indexer.TxInputOutput{custodialTxOutput}),
		PolicyScript: custodialPolicyScript,
		Address:      custodialAddr,
	}

	custodialNft := cardanowallet.NewTokenAmount(*cco.config.CustodialNft, 1)

	// custodial address must receive back it's nft token in standalone output
	getOutputsData.TxOutputs.Outputs = append(getOutputsData.TxOutputs.Outputs,
		cardanowallet.NewTxOutput(custodialAddr, custodialTxOutput.Output.Amount, custodialNft))

	return nil
}

func (cco *CardanoChainOperations) shouldConsolidate(err error) bool {
	return errors.Is(err, cardanowallet.ErrUTXOsLimitReached) ||
		errors.Is(err, errTxSizeTooBig) ||
		errors.Is(err, cardano.ErrInsufficientChange)
}

func (cco *CardanoChainOperations) getConsolidationType(err error) core.ConsolidationType {
	if errors.Is(err, cardano.ErrInsufficientChange) {
		return core.ConsolidationTypeToZeroAddress
	}

	return core.ConsolidationTypeSameAddress
}

func (cco *CardanoChainOperations) generateConsolidationTransaction(
	ctx context.Context,
	data *batchInitialData,
	chosenMultisigAddresses []common.AddressAndAmount,
	consolidationType core.ConsolidationType,
) (*core.GeneratedBatchTxData, error) {
	cco.logger.Info("Preparing consolidation tx", "consolidationType", consolidationType.String(),
		"consolidationTxID", data.BatchNonceID, "chain id", data.ChainID,
		"chosenMultisigAddresses", chosenMultisigAddresses)

	feeMultisigAddress := cco.bridgingAddressesManager.GetFeeMultisigAddress(data.ChainID)

	utxosForConsolidationRet, err := cco.getUTXOsForConsolidation(
		chosenMultisigAddresses, feeMultisigAddress, consolidationType)
	if err != nil {
		return nil, err
	}

	multisigTxOutputs := make([]cardanowallet.TxOutput, 0, len(chosenMultisigAddresses))

	var (
		addr string
		ok   bool
	)

	// If not all chosen addresses are used for consolidation remove the unused
	for i, addressAndAmount := range utxosForConsolidationRet.chosenMultisigAddresses {
		sum := cardano.GetSumMapFromTxInputOutput(utxosForConsolidationRet.multisigUtxos[addressAndAmount.AddressIndex])

		addr = addressAndAmount.Address
		if consolidationType == core.ConsolidationTypeToZeroAddress {
			// Consolidate everything to addr 0
			addr, ok = cco.bridgingAddressesManager.GetPaymentAddressFromIndex(data.ChainID, 0)
			if !ok {
				return nil, fmt.Errorf("failed to get first bridging address for chain: %d", data.ChainID)
			}
		}

		multisigTxOutput, err := getTxOutputFromSumMap(
			addr,
			sum,
		)
		if err != nil {
			return nil, err
		}

		multisigTxOutputs = append(multisigTxOutputs, multisigTxOutput)
		utxosForConsolidationRet.chosenMultisigAddresses[i].TokensAmounts = sum
	}

	slotNumber, err := cco.getSlotNumber()
	if err != nil {
		return nil, err
	}

	cco.logger.Info("Creating consolidation tx", "consolidationTxID", data.BatchNonceID,
		"magic", cco.config.NetworkMagic, "binary", cco.cardanoCliBinary,
		"slot", slotNumber, "multisig", len(utxosForConsolidationRet.multisigUtxos),
		"fee", len(utxosForConsolidationRet.feeUtxos))

	// Generate tx inputs
	feePolicyScript, ok := cco.bridgingAddressesManager.GetFeeMultisigPolicyScript(data.ChainID)
	if !ok {
		return nil, fmt.Errorf("failed to get fee policy script for chain: %d", data.ChainID)
	}

	txInputs := cardano.TxInputInfos{
		MultiSig: make([]*cardano.TxInputInfo, 0, len(utxosForConsolidationRet.chosenMultisigAddresses)),
		MultiSigFee: &cardano.TxInputInfo{
			PolicyScript: feePolicyScript,
			Address:      feeMultisigAddress,
			TxInputs:     convertUTXOsToTxInputs(utxosForConsolidationRet.feeUtxos),
		},
	}

	for _, addressAndAmount := range utxosForConsolidationRet.chosenMultisigAddresses {
		policyScript, ok := cco.bridgingAddressesManager.GetPaymentPolicyScript(data.ChainID, addressAndAmount.AddressIndex)
		if !ok {
			return nil, fmt.Errorf("failed to get payment policy script for address: %d", addressAndAmount.AddressIndex)
		}

		txInputs.MultiSig = append(txInputs.MultiSig, &cardano.TxInputInfo{
			PolicyScript: policyScript,
			Address:      addressAndAmount.Address,
			TxInputs:     convertUTXOsToTxInputs(utxosForConsolidationRet.multisigUtxos[addressAndAmount.AddressIndex]),
		})
	}

	// Create Tx
	txRaw, txHash, err := cardano.CreateTx(
		ctx,
		cco.cardanoCliBinary,
		uint(cco.config.NetworkMagic),
		data.ProtocolParams,
		slotNumber+cco.config.TTLSlotNumberInc,
		data.Metadata,
		&txInputs,
		nil,
		multisigTxOutputs,
		nil,
		utxosForConsolidationRet.chosenMultisigAddresses,
		nil,
		0,
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

type utxosForConsolidation struct {
	chosenMultisigAddresses []common.AddressAndAmount
	multisigUtxos           map[uint8][]*indexer.TxInputOutput
	feeUtxos                []*indexer.TxInputOutput
}

func (cco *CardanoChainOperations) getUTXOsForConsolidation(
	chosenMultisigAddresses []common.AddressAndAmount,
	multisigFeeAddress string,
	consolidationType core.ConsolidationType,
) (*utxosForConsolidation, error) {
	feeUtxos, err := cco.db.GetAllTxOutputs(multisigFeeAddress, true)
	if err != nil {
		return nil, err
	}

	feeUtxos = cardano.FilterOutUtxosWithUnknownTokens(feeUtxos)

	if len(feeUtxos) == 0 {
		return nil, fmt.Errorf("fee multisig does not have any utxo: %s", multisigFeeAddress)
	}

	// do not take more than maxFeeUtxoCount
	feeUtxos = feeUtxos[:min(int(cco.config.MaxFeeUtxoCount), len(feeUtxos))] //nolint:gosec
	cco.logger.Debug("UTXOs retrieved fee", "address", multisigFeeAddress, "utxos", feeUtxos)

	maxUtxoCount := int(max(cco.config.MaxUtxoCount-uint(len(feeUtxos)), 1)) //nolint:gosec

	// In case we have more or equal to MaxUtxoCount addresses, we need to reduce the number of addresses
	// to the MaxUtxoCount / 2, otherwise we will not be able to create a meaningful transaction

	if len(chosenMultisigAddresses) >= maxUtxoCount/2 {
		sort.SliceStable(chosenMultisigAddresses, func(i, j int) bool {
			return chosenMultisigAddresses[i].UtxoCount > chosenMultisigAddresses[j].UtxoCount
		})

		cco.logger.Debug("Number of chosen addresses greather or equal to MaxUtxoCount / 2",
			"Num of chosen addresses", len(chosenMultisigAddresses), "MaxUtxoCount", maxUtxoCount/2)

		chosenMultisigAddresses = chosenMultisigAddresses[:maxUtxoCount/2]
	}

	consolidationInputs := make([]AddressConsolidationData, 0)

	knownTokens, err := cardano.GetKnownTokens(cco.config)
	if err != nil {
		return nil, fmt.Errorf("failed to get known tokens: %w", err)
	}

	totalNumberOfUtxos := 0

	for _, addressAndAmount := range chosenMultisigAddresses {
		multisigUtxos, err := cco.db.GetAllTxOutputs(addressAndAmount.Address, true)
		if err != nil {
			return nil, err
		}

		if addressAndAmount.AddressIndex == 0 {
			multisigUtxos = cardano.FilterOutUtxosWithUnknownTokens(multisigUtxos, knownTokens...)
		} else {
			multisigUtxos = cardano.FilterOutUtxosWithUnknownTokens(multisigUtxos)
		}

		if len(multisigUtxos) > 0 {
			totalNumberOfUtxos += len(multisigUtxos)

			consolidationInputs = append(consolidationInputs, AddressConsolidationData{
				Address:      addressAndAmount.Address,
				AddressIndex: addressAndAmount.AddressIndex,
				UtxoCount:    len(multisigUtxos),
				Utxos:        multisigUtxos,
			})

			cco.logger.Debug("UTXOs retrieved multisig", "address", addressAndAmount.Address, "utxos", multisigUtxos)
		}
	}

	consolidationChosenInputs, err := allocateInputsForConsolidation(
		consolidationInputs, maxUtxoCount, totalNumberOfUtxos, consolidationType)
	if err != nil {
		return nil, err
	}

	cco.logger.Debug("Consolidation chosen inputs", "max", cco.config.MaxUtxoCount,
		"totalNumberOfUtxos", totalNumberOfUtxos,
		"inputs", consolidationInputs, "chosen", consolidationChosenInputs)

	multisigUtxos := make(map[uint8][]*indexer.TxInputOutput)

	for _, input := range consolidationChosenInputs {
		multisigUtxos[input.AddressIndex] = input.Utxos
	}

	cco.logger.Debug("UTXOs chosen", "chosenMultisigAddresses", chosenMultisigAddresses,
		"multisig", multisigUtxos, "fee", feeUtxos)

	return &utxosForConsolidation{
		chosenMultisigAddresses: chosenMultisigAddresses,
		multisigUtxos:           multisigUtxos,
		feeUtxos:                feeUtxos,
	}, nil
}

func (cco *CardanoChainOperations) getUTXOsForNormalBatch(
	multisigAddresses []common.AddressAndAmount, multisigFeeAddress string, isRedistribution bool,
	refundUtxosCnt int,
) (*utxoSelectionResult, error) {
	feeUtxos, err := cco.getFeeUTXOsForNormalBatch(multisigFeeAddress)
	if err != nil {
		return nil, err
	}

	knownTokens, err := cardano.GetKnownTokens(cco.config)
	if err != nil {
		return nil, fmt.Errorf("failed to get known tokens: %w", err)
	}

	chosenMultisigUtxos := make(map[uint8][]*indexer.TxInputOutput)
	chosenMultisigUtxosSoFar := 0

	for _, addressAndAmount := range multisigAddresses {
		multisigUtxos, err := cco.db.GetAllTxOutputs(addressAndAmount.Address, true)
		if err != nil {
			return nil, err
		}

		if addressAndAmount.AddressIndex == 0 {
			multisigUtxos = cardano.FilterOutUtxosWithUnknownTokens(multisigUtxos, knownTokens...)
		} else {
			multisigUtxos = cardano.FilterOutUtxosWithUnknownTokens(multisigUtxos)
		}

		cco.logger.Debug("UTXOs retrieved",
			"multisig", addressAndAmount.Address, "utxos", multisigUtxos)

		if isRedistribution {
			if len(multisigUtxos) > getMaxUtxoCount(cco.config, refundUtxosCnt+len(feeUtxos)+chosenMultisigUtxosSoFar) {
				cco.logger.Debug("REDISTRIBUTION ErrUTXOsLimitReached", "multisigUtxos count", len(multisigUtxos))

				return nil, fmt.Errorf(
					"UTXO limit reached during redistribution. "+
						"multisigUtxos count: %d. Err: %w",
					len(multisigUtxos), cardanowallet.ErrUTXOsLimitReached,
				)
			}
		} else {
			cco.logger.Debug("Change included in utxo selection", addressAndAmount.IncludeChange)

			multisigUtxos, err = getNeededUtxos(
				multisigUtxos,
				addressAndAmount.TokensAmounts,
				addressAndAmount.IncludeChange,
				getMaxUtxoCount(cco.config, refundUtxosCnt+len(feeUtxos)+chosenMultisigUtxosSoFar),
				int(cco.config.TakeAtLeastUtxoCount), //nolint:gosec
			)
			if err != nil {
				return nil, err
			}
		}

		cco.logger.Debug("UTXOs chosen", addressAndAmount.Address, multisigUtxos)
		chosenMultisigUtxos[addressAndAmount.AddressIndex] = multisigUtxos
		chosenMultisigUtxosSoFar += len(multisigUtxos)
	}

	return &utxoSelectionResult{
		multisigUtxos:            chosenMultisigUtxos,
		feeUtxos:                 feeUtxos,
		chosenMultisigUtxosCount: chosenMultisigUtxosSoFar,
	}, nil
}

func (cco *CardanoChainOperations) getFeeUTXOsForNormalBatch(
	multisigFeeAddress string,
) ([]*indexer.TxInputOutput, error) {
	feeUtxos, err := cco.db.GetAllTxOutputs(multisigFeeAddress, true)
	if err != nil {
		return nil, err
	}

	feeUtxos = cardano.FilterOutUtxosWithUnknownTokens(feeUtxos)

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

// this only handles special type of refund - when utxo contains unknown native tokens
func (cco *CardanoChainOperations) getUtxosFromRefundTransactions(
	confirmedTxs []eth.ConfirmedTransaction,
) ([][]*indexer.TxInputOutput, error) {
	utxosPerConfirmedTxs := make([][]*indexer.TxInputOutput, len(confirmedTxs))

	for i, ct := range confirmedTxs {
		if len(ct.OutputIndexes) == 0 {
			continue
		}

		indexes, err := common.UnpackNumbersToBytes[[]common.TxOutputIndex](ct.OutputIndexes)
		if err != nil {
			// this error could happen only if there is a bug in the smart contract (or oracle sent wrong values)
			cco.logger.Warn("failed to unpack output indexes",
				"err", err, "indxs", hex.EncodeToString(ct.OutputIndexes))

			continue
		}

		utxosPerConfirmedTxs[i] = make([]*indexer.TxInputOutput, len(indexes))

		for j, indx := range indexes {
			txInput := indexer.TxInput{
				Hash:  ct.ObservedTransactionHash,
				Index: uint32(indx),
			}

			// for now return error
			txOutput, err := cco.db.GetTxOutput(txInput)
			if err != nil {
				return nil, fmt.Errorf("failed to get tx output for %v: %w", txInput, err)
			}

			utxosPerConfirmedTxs[i][j] = &indexer.TxInputOutput{
				Input:  txInput,
				Output: txOutput,
			}
		}
	}

	return utxosPerConfirmedTxs, nil
}

func (cco *CardanoChainOperations) getCertificateData(
	data *batchInitialData, confirmedTransactions []eth.ConfirmedTransaction,
) (*cardano.CertificatesData, bool, error) {
	var (
		certificates         []*cardano.CertificatesWithScript
		keyRegistrationFee   uint64
		keyDeregistrationFee uint64
		hasBridgingTx        = false
	)

	for _, tx := range confirmedTransactions {
		if tx.TransactionType == uint8(common.StakeConfirmedTxType) {
			policyScript, ok := cco.bridgingAddressesManager.GetStakePolicyScript(data.ChainID, tx.BridgeAddrIndex)
			if !ok {
				return nil, false, fmt.Errorf("failed to get stake policy script for address: %d", tx.BridgeAddrIndex)
			}

			multisigStakeAddress, ok := cco.bridgingAddressesManager.GetStakeAddressFromIndex(data.ChainID, tx.BridgeAddrIndex)
			if !ok {
				return nil, false, fmt.Errorf("failed to get stake address from index: %d", tx.BridgeAddrIndex)
			}

			certificate, depositAmount, err := getStakingCertificates(
				cco.cardanoCliBinary, data, &tx, policyScript, multisigStakeAddress)

			if errors.Is(err, errSkipConfirmedTx) {
				cco.logger.Error("Staking delegation transaction skipped",
					"tx", eth.ConfirmedTransactionsWrapper{Txs: []eth.ConfirmedTransaction{tx}}, "err", err)

				continue
			} else if err != nil {
				return nil, false, err
			}

			certificates = append(certificates, certificate)

			if tx.TransactionSubType == uint8(common.StakeRegDelConfirmedTxSubType) {
				keyRegistrationFee += depositAmount
			}

			if tx.TransactionSubType == uint8(common.StakeDeregConfirmedTxSubType) {
				keyDeregistrationFee += depositAmount
			}
		} else {
			hasBridgingTx = true
		}
	}

	if len(certificates) > 0 {
		return &cardano.CertificatesData{
			Certificates:      certificates,
			RegistrationFee:   keyRegistrationFee,
			DeregistrationFee: keyDeregistrationFee,
		}, hasBridgingTx, nil
	}

	return nil, hasBridgingTx, nil
}

func getMaxUtxoCount(config *cardano.CardanoChainConfig, prevUtxosCnt int) int {
	return max(int(config.MaxUtxoCount)-prevUtxosCnt, 0) //nolint:gosec
}
