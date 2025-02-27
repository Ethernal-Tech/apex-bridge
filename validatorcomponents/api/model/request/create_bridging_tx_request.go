package request

type CreateBridgingTxTransactionRequest struct {
	Addr          string `json:"addr"`
	Amount        uint64 `json:"amount"`
	IsNativeToken bool   `json:"isNativeToken"`
}

type CreateBridgingTxRequest struct {
	SenderAddr         string                               `json:"senderAddr"`
	SourceChainID      string                               `json:"sourceChainId"`
	DestinationChainID string                               `json:"destinationChainId"`
	Transactions       []CreateBridgingTxTransactionRequest `json:"transactions"`
	BridgingFee        uint64                               `json:"bridgingFee"`
	OperationFee       uint64                               `json:"operationFee"`
}
