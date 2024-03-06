package cardanotx

import (
	"encoding/hex"
	"encoding/json"
	"math/big"

	cardanowallet "github.com/Ethernal-Tech/cardano-infrastructure/wallet"
	"github.com/fxamacker/cbor/v2"
)

const TTLSlotNumberInc = 200

func CreateTx(testNetMagic uint,
	protocolParams []byte,
	timeToLive uint64,
	metadataBytes []byte,
	txInputInfos *TxInputInfos,
	outputs []cardanowallet.TxOutput) ([]byte, error) {
	outputsSum := cardanowallet.GetOutputsSum(outputs)

	builder, err := cardanowallet.NewTxBuilder()
	if err != nil {
		return nil, err
	}

	defer builder.Dispose()

	builder.SetProtocolParameters(protocolParams).SetTimeToLive(timeToLive)
	builder.SetMetaData(metadataBytes).SetTestNetMagic(testNetMagic)
	builder.AddOutputs(outputs...).AddOutputs(cardanowallet.TxOutput{
		Addr: txInputInfos.MultiSig.Address,
	}).AddOutputs(cardanowallet.TxOutput{
		Addr: txInputInfos.MultiSigFee.Address,
	})
	builder.AddInputsWithScript(txInputInfos.MultiSig.PolicyScript, txInputInfos.MultiSig.Inputs...)
	builder.AddInputsWithScript(txInputInfos.MultiSigFee.PolicyScript, txInputInfos.MultiSigFee.Inputs...)

	fee, err := builder.CalculateFee(0)
	if err != nil {
		return nil, err
	}

	builder.SetFee(fee)

	builder.UpdateOutputAmount(-2, txInputInfos.MultiSig.InputsSum-outputsSum)
	builder.UpdateOutputAmount(-1, txInputInfos.MultiSigFee.InputsSum-fee)

	return builder.Build()
}

func AddTxWitness(key cardanowallet.ISigningKeyRetriver, txRaw []byte) ([]byte, error) {
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

func CreateMetaData(v *big.Int) ([]byte, error) {
	metadata := map[string]interface{}{
		"0": map[string]interface{}{
			"type":  "multi",
			"value": v.String(),
		},
	}

	return json.Marshal(metadata)
}
