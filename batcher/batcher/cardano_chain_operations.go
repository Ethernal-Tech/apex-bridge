package batcher

import (
	"context"
	"math/big"

	"github.com/Ethernal-Tech/apex-bridge/batcher/bridge"
	"github.com/Ethernal-Tech/apex-bridge/batcher/core"
	cardanotx "github.com/Ethernal-Tech/apex-bridge/cardano"
	"github.com/Ethernal-Tech/apex-bridge/contractbinding"
	ethtxhelper "github.com/Ethernal-Tech/apex-bridge/eth/txhelper"
	cardanowallet "github.com/Ethernal-Tech/cardano-infrastructure/wallet"
)

var _ core.ChainOperations = (*CardanoChainOperations)(nil)

type CardanoChainOperations struct {
	txProvider *cardanowallet.TxProviderBlockFrost
	config     core.CardanoChainConfig
}

func NewCardanoChainOperations(txProvider *cardanowallet.TxProviderBlockFrost, config core.CardanoChainConfig) *CardanoChainOperations {
	return &CardanoChainOperations{
		txProvider: txProvider,
		config:     config,
	}
}

// GenerateBatchTransaction implements core.ChainOperations.
func (cco *CardanoChainOperations) GenerateBatchTransaction(
	ctx context.Context,
	ethTxHelper ethtxhelper.IEthTxHelper,
	smartContractAddress string,
	destinationChain string,
	confirmedTransactions []contractbinding.TestContractConfirmedTransaction) ([]byte, string, *contractbinding.TestContractUTXOs, error) {

	var outputs []cardanowallet.TxOutput
	var txCost *big.Int = big.NewInt(0)
	for _, transaction := range confirmedTransactions {
		for _, receiver := range transaction.Receivers {
			outputs = append(outputs, cardanowallet.TxOutput{
				Addr:   receiver.DestinationAddress,
				Amount: receiver.Amount.Uint64(),
			})
			txCost.Add(txCost, receiver.Amount)
		}
	}

	inputUtxos, err := bridge.GetAvailableUTXOs(ctx, ethTxHelper, smartContractAddress, destinationChain, txCost)
	if err != nil {
		return nil, "", nil, err
	}

	// TODO: Create correct metadata
	metadata, err := cardanotx.CreateMetaData(big.NewInt(1))
	if err != nil {
		return nil, "", nil, err
	}

	protocolParams, err := cco.txProvider.GetProtocolParameters(ctx)
	if err != nil {
		return nil, "", nil, err
	}

	// TODO: Get slot from smart contract
	slotNumber := uint64(44102853 + 5*24*60*60)

	// TODO: Get keyhashes and atLeast from contract
	// TODO: Create PolicyScript from keyhashes and atLeast
	// TODO: Generate multisig addresses from keyhashes and atLeast
	atLeast := 3*2/3 + 1
	multisigPolicyScript, err := cardanowallet.NewPolicyScript(dummyKeyHashes[0:3], atLeast)
	if err != nil {
		return nil, "", nil, err
	}
	multisigFeePolicyScript, err := cardanowallet.NewPolicyScript(dummyKeyHashes[3:], atLeast)
	if err != nil {
		return nil, "", nil, err
	}

	var multisigInputs []cardanowallet.TxInput
	var multisigInputsSum uint64 = 0
	for _, utxo := range inputUtxos.MultisigOwnedUTXOs {
		multisigInputs = append(multisigInputs, cardanowallet.TxInput{
			Hash:  utxo.TxHash,
			Index: uint32(utxo.TxIndex.Uint64()),
		})
		multisigInputsSum += utxo.Amount.Uint64()
	}

	var multisigFeeInputs []cardanowallet.TxInput
	var multisigFeeInputsSum uint64 = 0
	for _, utxo := range inputUtxos.FeePayerOwnedUTXOs {
		multisigFeeInputs = append(multisigFeeInputs, cardanowallet.TxInput{
			Hash:  utxo.TxHash,
			Index: uint32(utxo.TxIndex.Uint64()),
		})
		multisigFeeInputsSum += utxo.Amount.Uint64()
	}

	multisigAddress, err := multisigPolicyScript.CreateMultiSigAddress(cco.config.TestNetMagic)
	if err != nil {
		return nil, "", nil, err
	}
	multisigFeeAddress, err := multisigFeePolicyScript.CreateMultiSigAddress(cco.config.TestNetMagic)
	if err != nil {
		return nil, "", nil, err
	}

	txInfos := &cardanotx.TxInputInfos{
		TestNetMagic: cco.config.TestNetMagic,
		MultiSig: &cardanotx.TxInputInfo{
			PolicyScript: multisigPolicyScript,
			Inputs:       multisigInputs,
			InputsSum:    multisigInputsSum,
			Address:      multisigAddress,
		},
		MultiSigFee: &cardanotx.TxInputInfo{
			PolicyScript: multisigFeePolicyScript,
			Inputs:       multisigFeeInputs,
			InputsSum:    multisigFeeInputsSum,
			Address:      multisigFeeAddress,
		},
	}

	rawTx, txHash, err := cardanotx.CreateTx(cco.config.TestNetMagic, protocolParams, slotNumber+cardanotx.TTLSlotNumberInc,
		metadata, txInfos, outputs)
	if err != nil {
		return nil, "", nil, err
	}

	return rawTx, txHash, inputUtxos, nil
}

// SignBatchTransaction implements core.ChainOperations.
func (*CardanoChainOperations) SignBatchTransaction(txHash string, signingKey string, signingKeyFee string) ([]byte, []byte, error) {
	sigKey := cardanotx.NewSigningKey(signingKey)
	sigKeyFee := cardanotx.NewSigningKey(signingKeyFee)

	witnessMultiSig, err := cardanotx.CreateTxWitness(txHash, sigKey)
	if err != nil {
		return nil, nil, err
	}

	witnessMultiSigFee, err := cardanotx.CreateTxWitness(txHash, sigKeyFee)
	if err != nil {
		return nil, nil, err
	}

	return witnessMultiSig, witnessMultiSigFee, nil
}

var (
	dummyKeyHashes = []string{
		"eff5e22355217ec6d770c3668010c2761fa0863afa12e96cff8a2205",
		"ad8e0ab92e1febfcaf44889d68c3ae78b59dc9c5fa9e05a272214c13",
		"bfd1c0eb0a453a7b7d668166ce5ca779c655e09e11487a6fac72dd6f",
		"b4689f2e8f37b406c5eb41b1fe2c9e9f4eec2597c3cc31b8dfee8f56",
		"39c196d28f804f70704b6dec5991fbb1112e648e067d17ca7abe614b",
		"adea661341df075349cbb2ad02905ce1828f8cf3e66f5012d48c3168",
	}
	dummySigningKeys = []string{
		"58201825bce09711e1563fc1702587da6892d1d869894386323bd4378ea5e3d6cba0",
		"5820ccdae0d1cd3fa9be16a497941acff33b9aa20bdbf2f9aa5715942d152988e083",
		"582094bfc7d65a5d936e7b527c93ea6bf75de51029290b1ef8c8877bffe070398b40",
		"58204cd84bf321e70ab223fbdbfe5eba249a5249bd9becbeb82109d45e56c9c610a9",
		"58208fcc8cac6b7fedf4c30aed170633df487642cb22f7e8615684e2b98e367fcaa3",
		"582058fb35da120c65855ad691dadf5681a2e4fc62e9dcda0d0774ff6fdc463a679a",
	}
)
