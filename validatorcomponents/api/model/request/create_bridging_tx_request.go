package request

type CreateBridgingTxTransactionRequest struct {
	Addr   string `json:"addr"`
	Amount uint64 `json:"amount"`
}

type CreateBridgingTxRequest struct {
	SenderAddr         string                               `json:"senderAddr"`
	SourceChainID      string                               `json:"sourceChainId"`
	DestinationChainID string                               `json:"destinationChainId"`
	Transactions       []CreateBridgingTxTransactionRequest `json:"transactions"`
}
