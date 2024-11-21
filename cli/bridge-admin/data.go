package clibridgeadmin

import "github.com/Ethernal-Tech/apex-bridge/common"

const (
	bridgeNodeURLFlag = "bridge-url"
	chainIDFlag       = "chain"
	amountFlag        = "amount"
	privateKeyFlag    = "key"
	addressFlag       = "addr"

	bridgeNodeURLFlagDesc = "node URL of bridge chain"
	chainIDFlagDesc       = "chain ID (prime, vector, etc)"
	privateKeyFlagDesc    = "wallet private signing key"

	gasLimitMultiplier = 2.0
)

var (
	apexBridgeAdminScAddress = common.HexToAddress("0xABEF000000000000000000000000000000000006")
)
