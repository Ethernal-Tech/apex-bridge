package response

type BridgingTxResponse struct {
	TxRaw             string `json:"txRaw"`
	TxHash            string `json:"txHash"`
	Amount            uint64 `json:"amount"`
	NativeTokenAmount uint64 `json:"nativeTokenAmount"`
}

func NewBridgingTxResponse(
	txRaw string, txHash string,
) *BridgingTxResponse {
	return &BridgingTxResponse{
		TxRaw:  txRaw,
		TxHash: txHash,
	}
}

func NewSkylineBridgingTxResponse(
	txRaw string, txHash string, amount uint64, nativeTokenAmount uint64,
) *BridgingTxResponse {
	return &BridgingTxResponse{
		TxRaw:             txRaw,
		TxHash:            txHash,
		Amount:            amount,
		NativeTokenAmount: nativeTokenAmount,
	}
}
