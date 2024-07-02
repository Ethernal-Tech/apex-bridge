package request

type SignBridgingTxRequest struct {
	SigningKeyHex string `json:"signingKey"`
	NetworkID     uint   `json:"networkID"`
	TxRaw         string `json:"txRaw"`
	TxHash        string `json:"txHash"`
}
