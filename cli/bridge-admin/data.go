package clibridgeadmin

import "github.com/Ethernal-Tech/apex-bridge/common"

const (
	bridgeNodeURLFlag    = "bridge-url"
	chainIDFlag          = "chain"
	networkIDFlag        = "network-id"
	testnetMagicFlag     = "testnet-magic"
	ogmiosURLFlag        = "ogmios"
	amountFlag           = "amount"
	stakePrivateKeyFlag  = "stake-key"
	privateKeyFlag       = "key" // once these two key flags should be joined into one...
	bridgePrivateKeyFlag = "bridge-key"
	privateKeyConfigFlag = "key-config"
	addressFlag          = "addr"

	bridgeNodeURLFlagDesc    = "node URL of bridge chain"
	chainIDFlagDesc          = "chain ID (prime, vector, nexus, etc)"
	networkIDFlagDesc        = "network id"
	testnetMagicFlagDesc     = "testnet magic number. leave 0 for mainnet"
	ogmiosURLFlagDesc        = "ogmios url"
	stakePrivateKeyFlagDesc  = "wallet stake signing key"
	bridgePrivateKeyFlagDesc = "bridge admin private key"
	privateKeyConfigFlagDesc = "path to secrets manager config file"

	gasLimitMultiplier = 2.0
)

var (
	apexBridgeAdminScAddress = common.HexToAddress("0xABEF000000000000000000000000000000000006")
	apexBridgeScAddress      = common.HexToAddress("0xABEF000000000000000000000000000000000000")
)
