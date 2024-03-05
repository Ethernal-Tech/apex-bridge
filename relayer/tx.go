package relayer

import (
	"encoding/hex"
	"encoding/json"
	"math/big"

	cardanowallet "github.com/Ethernal-Tech/cardano-infrastructure/wallet"
	"github.com/fxamacker/cbor/v2"
)

func CreateMetaData(v *big.Int) ([]byte, error) {
	metadata := map[string]interface{}{
		"0": map[string]interface{}{
			"type":  "multi",
			"value": v.String(),
		},
	}

	return json.Marshal(metadata)
}

func CreateTx(config *RelayerConfiguration,
	metadataBytes []byte,
	keyHashesMultiSig []string,
	keyHashesMultiSigFee []string) ([]byte, error) {
	txProvider, err := cardanowallet.NewTxProviderBlockFrost(config.Cardano.BlockfrostUrl, config.Cardano.BlockfrostAPIKey)
	if err != nil {
		return nil, err
	}

	defer txProvider.Dispose()

	policyScriptMultiSig, err := cardanowallet.NewPolicyScript(keyHashesMultiSig, len(keyHashesMultiSig)*2/3+1)
	if err != nil {
		return nil, err
	}

	policyScriptFeeMultiSig, err := cardanowallet.NewPolicyScript(keyHashesMultiSigFee, len(keyHashesMultiSigFee)*2/3+1)
	if err != nil {
		return nil, err
	}

	multiSigAddr, err := policyScriptMultiSig.CreateMultiSigAddress(config.Cardano.TestNetMagic)
	if err != nil {
		return nil, err
	}

	multiSigFeeAddr, err := policyScriptFeeMultiSig.CreateMultiSigAddress(config.Cardano.TestNetMagic)
	if err != nil {
		return nil, err
	}

	outputs := []cardanowallet.TxOutput{
		{
			Addr:   "addr_test1vqjysa7p4mhu0l25qknwznvj0kghtr29ud7zp732ezwtzec0w8g3u",
			Amount: cardanowallet.MinUTxODefaultValue,
		},
	}
	outputsSum := cardanowallet.GetOutputsSum(outputs)

	builder, err := cardanowallet.NewTxBuilder()
	if err != nil {
		return nil, err
	}

	defer builder.Dispose()

	if err := builder.SetProtocolParametersAndTTL(txProvider, 0); err != nil {
		return nil, err
	}

	multiSigInputs, err := cardanowallet.GetUTXOsForAmount(txProvider, multiSigAddr, cardanowallet.MinUTxODefaultValue)
	if err != nil {
		return nil, err
	}

	multiSigFeeInputs, err := cardanowallet.GetUTXOsForAmount(txProvider, multiSigFeeAddr, config.Cardano.PotentialFee)
	if err != nil {
		return nil, err
	}

	builder.SetMetaData(metadataBytes).SetTestNetMagic(config.Cardano.TestNetMagic)
	builder.AddOutputs(outputs...).AddOutputs(cardanowallet.TxOutput{
		Addr: multiSigAddr,
	}).AddOutputs(cardanowallet.TxOutput{
		Addr: multiSigFeeAddr,
	})
	builder.AddInputsWithScript(policyScriptMultiSig, multiSigInputs.Inputs...)
	builder.AddInputsWithScript(policyScriptFeeMultiSig, multiSigFeeInputs.Inputs...)

	fee, err := builder.CalculateFee(0)
	if err != nil {
		return nil, err
	}

	builder.SetFee(fee)

	builder.UpdateOutputAmount(-2, multiSigInputs.Sum-outputsSum)
	builder.UpdateOutputAmount(-1, multiSigFeeInputs.Sum-fee)

	return builder.Build()
}

func AddTxWitness(key SigningKey, txRaw []byte) ([]byte, error) {
	builder, err := cardanowallet.NewTxBuilder()
	if err != nil {
		return nil, err
	}

	defer builder.Dispose()

	return builder.AddWitness(txRaw, key)
}

func AssemblyFinalTx(txRaw []byte, witnesses [][]byte) ([]byte, string, error) {
	builder, err := cardanowallet.NewTxBuilder()
	if err != nil {
		return nil, "", err
	}

	defer builder.Dispose()

	txSigned, err := builder.AssembleWitnesses(txRaw, witnesses)
	if err != nil {
		return nil, "", err
	}

	hash, err := builder.GetTxHash(txRaw)
	if err != nil {
		return nil, "", err
	}

	return txSigned, hash, nil
}

type SigningKey []byte

func NewSigningKey(s string) SigningKey {
	return SigningKey(decodeCbor(s))
}

func (sk SigningKey) GetSigningKey() []byte {
	return []byte(sk)
}

func decodeCbor(s string) (r []byte) {
	b, _ := hex.DecodeString(s)
	_ = cbor.Unmarshal(b, &r)

	return r
}
