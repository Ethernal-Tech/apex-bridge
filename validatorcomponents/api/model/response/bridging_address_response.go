package response

import "github.com/Ethernal-Tech/apex-bridge/common"

type BridgingAddressResponse struct {
	// Chain ID
	ChainID string `json:"chainID"`
	// Bridging address index
	AddressIndex uint8 `json:"addressIndex"`
	// Bridging address
	Address string `json:"address"`
} // @name BridgingAddressResponse

func NewBridgingAddressResponse(
	chainID string, bridgingAddress common.AddressAndAmount,
) *BridgingAddressResponse {
	return &BridgingAddressResponse{
		ChainID:      chainID,
		AddressIndex: bridgingAddress.AddressIndex,
		Address:      bridgingAddress.Address,
	}
}
