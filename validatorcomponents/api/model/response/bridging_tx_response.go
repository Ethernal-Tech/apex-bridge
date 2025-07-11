package response

import "encoding/hex"

type BridgingTxResponse struct {
	TxRaw             string `json:"txRaw"`
	TxHash            string `json:"txHash"`
	BridgingFee       uint64 `json:"bridgingFee"`
	Amount            uint64 `json:"amount"`
	NativeTokenAmount uint64 `json:"nativeTokenAmount"`
} // @name BridgingTxResponse

func NewBridgingTxResponse(
	txRaw []byte, txHash string, bridgingFee uint64, amount uint64, nativeTokenAmount uint64,
) *BridgingTxResponse {
	return &BridgingTxResponse{
		TxRaw:             hex.EncodeToString(txRaw),
		TxHash:            txHash,
		BridgingFee:       bridgingFee,
		Amount:            amount,
		NativeTokenAmount: nativeTokenAmount,
	}
}
