package request

import (
	cardanowallet "github.com/Ethernal-Tech/cardano-infrastructure/wallet"
)

type CreateBridgingTxTransactionRequest struct {
	Addr   string `json:"addr"`
	Amount uint64 `json:"amount"`
}

type CreateBridgingTxRequest struct {
	SenderAddr             string                               `json:"senderAddr"`
	SenderAddrPolicyScript *cardanowallet.PolicyScript          `json:"senderAddrPolicyScript"`
	SourceChainID          string                               `json:"sourceChainId"`
	BridgingAddress        string                               `json:"bridgingAddress"`
	DestinationChainID     string                               `json:"destinationChainId"`
	Transactions           []CreateBridgingTxTransactionRequest `json:"transactions"`
	BridgingFee            uint64                               `json:"bridgingFee"`
}
