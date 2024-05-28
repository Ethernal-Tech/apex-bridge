package request

type SignBridgingTxRequest struct {
	SigningKeyHex string `json:"signingKey"`
	TxRaw         string `json:"txRaw"`
	TxHash        string `json:"txHash"`
}
