package cardanotx

import (
	"encoding/hex"
	"encoding/json"
	"math/big"

	"github.com/Ethernal-Tech/apex-bridge/common"
	cardanowallet "github.com/Ethernal-Tech/cardano-infrastructure/wallet"
	"github.com/fxamacker/cbor/v2"
)

const TTLSlotNumberInc = 200

// CreateTx creates tx and returns cbor of raw transaction data, tx hash and error
func CreateTx(testNetMagic uint,
	protocolParams []byte,
	timeToLive uint64,
	metadataBytes []byte,
	txInputInfos *TxInputInfos,
	outputs []cardanowallet.TxOutput) ([]byte, string, error) {
	outputsSum := cardanowallet.GetOutputsSum(outputs)

	builder, err := cardanowallet.NewTxBuilder()
	if err != nil {
		return nil, "", err
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
		return nil, "", err
	}

	builder.SetFee(fee)

	builder.UpdateOutputAmount(-2, txInputInfos.MultiSig.InputsSum-outputsSum)
	builder.UpdateOutputAmount(-1, txInputInfos.MultiSigFee.InputsSum-fee)

	return builder.Build()
}

// CreateTxWitness creates cbor of vkey+signature pair of tx hash
func CreateTxWitness(txHash string, key cardanowallet.ISigner) ([]byte, error) {
	return cardanowallet.CreateTxWitness(txHash, key)
}

// AssembleTxWitnesses assembles all witnesses in final cbor of signed tx
func AssembleTxWitnesses(txRaw []byte, witnesses [][]byte) ([]byte, error) {
	return cardanowallet.AssembleTxWitnesses(txRaw, witnesses)
}

type SigningKey struct {
	private []byte
	public  []byte
}

func NewSigningKey(s string) SigningKey {
	private := decodeCbor(s)

	return SigningKey{
		private: private,
		public:  cardanowallet.GetVerificationKeyFromSigningKey(private),
	}
}

func (sk SigningKey) GetSigningKey() []byte {
	return sk.private
}

func (sk SigningKey) GetVerificationKey() []byte {
	return sk.public
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

func CreateBatchMetaData(v *big.Int) ([]byte, error) {
	return common.MarshalMetadata(common.MetadataEncodingTypeJson, common.BatchExecutedMetadata{
		BridgingTxType: common.BridgingTxTypeBatchExecution,
		BatchNonceId:   v.Uint64(),
	})
}
