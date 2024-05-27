package response

type BridgingTxResponse struct {
	TxRaw         string   `json:"txRaw"`
	TxHash        string   `json:"txHash"`
	ReceiverAddrs []string `json:"receiverAddrs"`
	Amount        uint64   `json:"amount"`
}

func NewFullBridgingTxResponse(
	txRaw string, txHash string, receiverAddrs []string, amount uint64,
) *BridgingTxResponse {
	return &BridgingTxResponse{
		TxRaw:         txRaw,
		TxHash:        txHash,
		ReceiverAddrs: receiverAddrs,
		Amount:        amount,
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
