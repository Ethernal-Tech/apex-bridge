package response

import "encoding/hex"

type BridgingTxResponse struct {
	TxRaw       string `json:"txRaw"`
	TxHash      string `json:"txHash"`
	BridgingFee uint64 `json:"bridgingFee"`
}

func NewFullBridgingTxResponse(
	txRaw []byte, txHash string, bridgingFee uint64,
) *BridgingTxResponse {
	return &BridgingTxResponse{
		TxRaw:       hex.EncodeToString(txRaw),
		TxHash:      txHash,
		BridgingFee: bridgingFee,
	}
}

func NewBridgingTxResponse(
	txRaw string, txHash string,
) *BridgingTxResponse {
	return &BridgingTxResponse{
		TxRaw:  txRaw,
		TxHash: txHash,
	}
}
