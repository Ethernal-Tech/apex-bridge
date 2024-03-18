package contractbinding

import (
	"fmt"
	"math/big"

	cardanotx "github.com/Ethernal-Tech/apex-bridge/cardano"
	cardanowallet "github.com/Ethernal-Tech/cardano-infrastructure/wallet"
)

type BatcherTestContractMock struct{}

type UTXO struct {
	TxHash  string
	TxIndex *big.Int
	Amount  *big.Int
}

type UTXOs struct {
	MultisigOwnedUTXOs []UTXO
	FeePayerOwnedUTXOs []UTXO
}

type ConfirmedTransaction struct {
	Nonce     *big.Int
	Receivers map[string]*big.Int
}

type SignedBatch struct {
	ID                        string
	DestinationChainID        string
	RawTransaction            string
	MultisigSignature         string
	FeePayerMultisigSignature string
	IncludedTransactions      []ConfirmedTransaction
	UsedUTXOs                 UTXOs
}

func NewBatcherTestContractMock() *BatcherTestContractMock {
	return &BatcherTestContractMock{}
}

// ShouldCreateBatch mocks the contract function
func (m *BatcherTestContractMock) ShouldCreateBatch(destinationChain string) (bool, error) {
	// Mocked implementation for testing
	return true, nil
}

// GetConfirmedTransactions mocks the contract function
func (m *BatcherTestContractMock) GetConfirmedTransactions(destinationChain string) ([]ConfirmedTransaction, error) {
	// Mocked implementation for testing

	var retVal []ConfirmedTransaction

	for id, output := range dummyOutputs {
		retVal = append(retVal, ConfirmedTransaction{Nonce: big.NewInt(int64(id)), Receivers: map[string]*big.Int{output.Addr: big.NewInt(int64(output.Amount))}})
	}

	return retVal, nil
}

// GetAvailableUTXOs mocks the contract function
func (m *BatcherTestContractMock) GetAvailableUTXOs(destinationChain string, txCost *big.Int) (UTXOs, error) {
	// Mocked implementation for testing
	var retVal UTXOs

	// Query UTXOs from chain
	txInfos, err := cardanotx.NewTxInputInfos(
		dummyKeyHashes[0:3], dummyKeyHashes[3:], uint(2))
	if err != nil {
		return retVal, err
	}

	txProvider, err := cardanowallet.NewTxProviderBlockFrost("https://cardano-preview.blockfrost.io/api/v0", "preview7mGSjpyEKb24OxQ4cCxomxZ5axMs5PvE")
	if err != nil {
		return retVal, err
	}

	err = txInfos.CalculateWithRetriever(txProvider, txCost.Uint64(), uint64(300000))
	if err != nil {
		return retVal, err
	}

	// Get multisig utxos for amount and place them into map
	var utxoAmount map[string]*big.Int = make(map[string]*big.Int)
	multisigUtxos, err := txProvider.GetUtxos(txInfos.MultiSig.Address)
	if err != nil {
		return retVal, err
	}
	for _, utxo := range multisigUtxos {
		utxoAmount[utxo.Hash] = big.NewInt(int64(utxo.Amount))
	}

	for _, input := range txInfos.MultiSig.Inputs {
		retVal.MultisigOwnedUTXOs = append(retVal.MultisigOwnedUTXOs, UTXO{
			TxHash:  input.Hash,
			TxIndex: big.NewInt(int64(input.Index)),
			Amount:  utxoAmount[input.Hash],
		})
	}

	// Get multisig fee utxos for amount and place them into map
	utxoAmount = make(map[string]*big.Int)
	multisigUtxos, err = txProvider.GetUtxos(txInfos.MultiSigFee.Address)
	if err != nil {
		return retVal, err
	}
	for _, utxo := range multisigUtxos {
		utxoAmount[utxo.Hash] = big.NewInt(int64(utxo.Amount))
	}

	for _, input := range txInfos.MultiSigFee.Inputs {
		retVal.FeePayerOwnedUTXOs = append(retVal.FeePayerOwnedUTXOs, UTXO{
			TxHash:  input.Hash,
			TxIndex: big.NewInt(int64(input.Index)),
			Amount:  utxoAmount[input.Hash],
		})
	}

	return retVal, nil
}

// SubmitSignedBatch mocks the contract function
func (m *BatcherTestContractMock) SubmitSignedBatch(signedBatch SignedBatch) error {
	// Mocked implementation for testing
	// Here you can add assertions or custom logic for testing
	fmt.Println(signedBatch)
	return nil
}

var (
	dummyMumbaiAccPk = "3761f6deeb2e0b2aa8b843e804d880afa6e5fecf1631f411e267641a72d0ca20"
	dummyKeyHashes   = []string{
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
	dummyOutputs = []cardanowallet.TxOutput{
		{
			Addr:   "addr_test1vqjysa7p4mhu0l25qknwznvj0kghtr29ud7zp732ezwtzec0w8g3u",
			Amount: cardanowallet.MinUTxODefaultValue,
		},
	}
)
