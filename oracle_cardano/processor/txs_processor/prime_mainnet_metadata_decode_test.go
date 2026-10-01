package processor

import (
	"encoding/hex"
	"testing"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/oracle_cardano/core"
	failedtxprocessors "github.com/Ethernal-Tech/apex-bridge/oracle_cardano/processor/tx_processors/failed"
	successtxprocessors "github.com/Ethernal-Tech/apex-bridge/oracle_cardano/processor/tx_processors/success"
	cCore "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
	"github.com/fxamacker/cbor/v2"
	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/require"
)

// Regression test for the metadata decoding failure seen on prime mainnet for tx
// f1fb470da307161d5de475213e65e81ece2d9b90d5478c22bab90d7e889829e0, which carries a
// shelley-era auxiliary_data envelope instead of the alonzo one cardano-cli emits.
// Both that tx and one with the alonzo envelope have to reach the same processor
// with the same metadata.
//
// The two blobs below are the exact auxiliary_data CBOR the oracle sees in production:
// they were pulled once with gouroboros BlockFetch over node-to-node, the same protocol
// the indexer uses, and are the bytes indexer/gouroboros/utils.go createTx puts in
// tx.Metadata. They are inlined so this stays an offline unit test. Do not refresh them
// from blockfrost or an explorer: db-sync stores only the inner metadata map and drops
// the auxiliary_data envelope, which is the part that breaks.
const (
	// bridging request with a shelley auxiliary_data envelope, built outside our tooling:
	// tx f1fb470da307161d5de475213e65e81ece2d9b90d5478c22bab90d7e889829e0 in block
	// 2665984 (slot 73346512, babbage era)
	shelleyEnvelopeAuxDataHex = "a101a661746662726964676561646763617264616e6f6173837828616464723171393479356175616561733064657430" +
		"6d7572373075383777636536303766306e3534782864666678343833323268737532646c79796b66347436613270346770787732" +
		"6864373236306c3661777a37706b70336865716c767276783865736579776a38676266611a003d0900626f660062747881a36161" +
		"8378286164647231713934793561756165617330646574306d7572373075383777636536303766306e353478286466667834383332" +
		"3268737532646c79796b663474366132703467707877326864373236306c3661777a37706b70336865716c767276783865736579" +
		"776a3867616d1a12a79e60617401"

	// bridging request with the alonzo envelope cardano-cli emits:
	// tx 96cf88405128be043e8efa4223f2738aefe968e7dc2502cf6274f2eb7a76c329 in block
	// 2682996 (slot 73740732, babbage era)
	alonzoEnvelopeAuxDataHex = "d90103a100a101a66164656e657875736266611a000f424a626f66006173837828616464723171383967376e783975687578" +
		"3772746b733872763030347076617a78787837737238707828397568753375306d78736a377a63743072786c7676746676377137" +
		"3263347161336a366568796e35773265716d78373763646d787236727878716b7a7778363961746662726964676562747881a361" +
		"6181782836464241456336414337333132364438633163466166393464316145364361354337393435323143616d1b0000000b66" +
		"5535d8617400"
)

// txFromAuxData rebuilds the indexer.Tx the oracle would get for the given
// auxiliary_data, the way indexer/gouroboros/utils.go createTx does.
func txFromAuxData(t *testing.T, auxDataHex string) *indexer.Tx {
	t.Helper()

	metadata, err := hex.DecodeString(auxDataHex)
	require.NoError(t, err)

	return &indexer.Tx{
		Metadata: metadata,
		Valid:    true,
	}
}

// newCardanoTxProcessorsCollection wires the processors the same way
// oracle_cardano/oracle/oracle.go does in reactor mode with refunds disabled.
func newCardanoTxProcessorsCollection() *txProcessorsCollection {
	logger := hclog.NewNullLogger()
	refundRequestProcessor := successtxprocessors.NewRefundDisabledProcessor()

	return NewTxProcessorsCollection(
		[]core.CardanoTxSuccessProcessor{
			successtxprocessors.NewBatchExecutedProcessor(logger),
			successtxprocessors.NewHotWalletIncrementProcessor(logger),
			successtxprocessors.NewBridgingRequestedProcessor(refundRequestProcessor, logger),
		},
		[]core.CardanoTxFailedProcessor{
			failedtxprocessors.NewBatchExecutionFailedProcessor(logger),
		},
	)
}

// selectProcessor runs the oracle's standard processor selection over the tx.
func selectProcessor(tx *indexer.Tx) (core.CardanoTxSuccessProcessor, error) {
	cardanoTx := &core.CardanoTx{
		OriginChainID: common.ChainIDStrPrime,
		Tx:            *tx,
	}

	return newCardanoTxProcessorsCollection().getSuccess(
		cardanoTx, &cCore.AppConfig{})
}

// auxDataEnvelope names which of the auxiliary_data encodings babbage accepts a tx
// used. The era names mark when each alternative was introduced - babbage still
// accepts all of them, so the tx builder is the one that picks.
func auxDataEnvelope(t *testing.T, auxData []byte) string {
	t.Helper()

	var asTag cbor.RawTag
	if err := cbor.Unmarshal(auxData, &asTag); err == nil && asTag.Number == 259 {
		return "alonzo #6.259({0: {label: M}})"
	}

	var asArray []cbor.RawMessage
	if err := cbor.Unmarshal(auxData, &asArray); err == nil {
		return "shelley-ma [{label: M}, [scripts]]"
	}

	var asMap map[uint64]cbor.RawMessage

	require.NoError(t, cbor.Unmarshal(auxData, &asMap))

	if _, ok := asMap[0]; ok && len(asMap) == 1 {
		return "alonzo {0: {label: M}}"
	}

	return "shelley {label: M}"
}

func TestPrimeMainnetTxMetadataDecode(t *testing.T) {
	t.Run("shelley envelope", func(t *testing.T) {
		tx := txFromAuxData(t, shelleyEnvelopeAuxDataHex)

		require.Equal(t, "shelley {label: M}", auxDataEnvelope(t, tx.Metadata))

		txProcessor, err := selectProcessor(tx)
		require.NoError(t, err)
		require.Equal(t, common.BridgingTxTypeBridgingRequest, txProcessor.GetType())

		metadata, err := common.UnmarshalMetadata[common.BridgingRequestMetadata](
			common.MetadataEncodingTypeCbor, tx.Metadata)
		require.NoError(t, err)

		require.EqualValues(t, common.BridgingTxTypeBridgingRequest, metadata.BridgingTxType)
		require.Equal(t, "cardano", metadata.DestinationChainID)
		require.EqualValues(t, 4000000, metadata.BridgingFee)
		require.Len(t, metadata.Transactions, 1)
		require.EqualValues(t, 312974944, metadata.Transactions[0].Amount)
	})

	t.Run("alonzo envelope", func(t *testing.T) {
		tx := txFromAuxData(t, alonzoEnvelopeAuxDataHex)

		require.Equal(t, "alonzo #6.259({0: {label: M}})", auxDataEnvelope(t, tx.Metadata))

		txProcessor, err := selectProcessor(tx)
		require.NoError(t, err)
		require.Equal(t, common.BridgingTxTypeBridgingRequest, txProcessor.GetType())

		metadata, err := common.UnmarshalMetadata[common.BridgingRequestMetadata](
			common.MetadataEncodingTypeCbor, tx.Metadata)
		require.NoError(t, err)

		require.EqualValues(t, common.BridgingTxTypeBridgingRequest, metadata.BridgingTxType)
		require.Equal(t, "nexus", metadata.DestinationChainID)
		require.EqualValues(t, 1000010, metadata.BridgingFee)
		require.Len(t, metadata.Transactions, 1)
		require.EqualValues(t, 48961500632, metadata.Transactions[0].Amount)
	})

	t.Run("the envelope makes no difference to the decoded metadata", func(t *testing.T) {
		tx := txFromAuxData(t, shelleyEnvelopeAuxDataHex)

		fromShelley, err := common.UnmarshalMetadata[common.BridgingRequestMetadata](
			common.MetadataEncodingTypeCbor, tx.Metadata)
		require.NoError(t, err)

		// the shelley envelope is the bare metadata map, so re-wrapping it the way
		// cardano-cli would has to yield exactly the same metadata
		alonzo, err := cbor.Marshal(cbor.Tag{
			Number:  259,
			Content: map[uint64]cbor.RawMessage{0: cbor.RawMessage(tx.Metadata)},
		})
		require.NoError(t, err)

		fromAlonzo, err := common.UnmarshalMetadata[common.BridgingRequestMetadata](
			common.MetadataEncodingTypeCbor, alonzo)
		require.NoError(t, err)

		require.Equal(t, *fromShelley, *fromAlonzo)
	})
}
