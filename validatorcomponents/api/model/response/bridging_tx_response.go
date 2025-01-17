package response

type BridgingTxResponse struct {
	TxRaw             string `json:"txRaw"`
	TxHash            string `json:"txHash"`
	BridgingFee       uint64 `json:"bridgingFee"`
	Amount            uint64 `json:"amount"`
	NativeTokenAmount uint64 `json:"nativeTokenAmount"`
}

func NewFullBridgingTxResponse(
	txRaw string, txHash string, bridgingFee uint64,
) *BridgingTxResponse {
	return &BridgingTxResponse{
		TxRaw:       txRaw,
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

func NewFullSkylineBridgingTxResponse(
	txRaw string, txHash string, bridgingFee uint64, amount uint64, nativeTokenAmount uint64,
) *BridgingTxResponse {
	return &BridgingTxResponse{
		TxRaw:             txRaw,
		TxHash:            txHash,
		BridgingFee:       bridgingFee,
		Amount:            amount,
		NativeTokenAmount: nativeTokenAmount,
	}
}
