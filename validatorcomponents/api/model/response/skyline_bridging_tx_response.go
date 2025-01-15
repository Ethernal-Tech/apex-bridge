package response

type SkylineBridgingTxResponse struct {
	TxRaw             string `json:"txRaw"`
	TxHash            string `json:"txHash"`
	BridgingFee       uint64 `json:"bridgingFee"`
	Amount            uint64 `json:"amount"`
	NativeTokenAmount uint64 `json:"nativeTokenAmount"`
}

func NewFullSkylineBridgingTxResponse(
	txRaw string, txHash string, bridgingFee uint64, amount uint64, nativeTokenAmount uint64,
) *SkylineBridgingTxResponse {
	return &SkylineBridgingTxResponse{
		TxRaw:             txRaw,
		TxHash:            txHash,
		BridgingFee:       bridgingFee,
		Amount:            amount,
		NativeTokenAmount: nativeTokenAmount,
	}
}

func NewSkylineBridgingTxResponse(
	txRaw string, txHash string,
) *SkylineBridgingTxResponse {
	return &SkylineBridgingTxResponse{
		TxRaw:  txRaw,
		TxHash: txHash,
	}
}
