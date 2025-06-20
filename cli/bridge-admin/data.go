package clibridgeadmin

import "github.com/Ethernal-Tech/apex-bridge/common"

const (
	bridgeNodeURLFlag          = "bridge-url"
	chainIDFlag                = "chain"
	amountFlag                 = "amount"
	privateKeyFlag             = "key" // once these two key flags should be joined into one...
	bridgePrivateKeyFlag       = "bridge-key"
	bridgePrivateKeyConfigFlag = "bridge-key-config"
	addressFlag                = "addr"

	bridgeNodeURLFlagDesc          = "node URL of bridge chain"
	chainIDFlagDesc                = "chain ID (prime, vector, nexus, etc)"
	bridgePrivateKeyFlagDesc       = "bridge admin private key"
	bridgePrivateKeyConfigFlagDesc = "path to secrets manager config file"

	gasLimitMultiplier = 2.0
)

var (
	apexBridgeAdminScAddress = common.HexToAddress("0xABEF000000000000000000000000000000000006")
	apexBridgeScAddress      = common.HexToAddress("0xABEF000000000000000000000000000000000000")
)
