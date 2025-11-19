package core

import (
	"fmt"
	"slices"

	cCore "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
)

func IsTxDirectionAllowed(appConfig *cCore.AppConfig, srcChainID string, metadata *BridgingRequestEthMetadata) error {
	destChainID := metadata.DestinationChainID

	allowedDirection, ok := appConfig.BridgingSettings.AllowedDirections[srcChainID][destChainID]
	if !ok {
		return fmt.Errorf("transaction direction not allowed: %s -> %s", srcChainID, destChainID)
	}

	for _, tx := range metadata.Transactions {
		if tx.ColoredCoinID == 0 && !allowedDirection.CurrencyBirdgingAllowed {
			return fmt.Errorf(
				"currency bridging not allowed for direction %s to %s",
				srcChainID, destChainID,
			)
		}

		if tx.ColoredCoinID > 0 && !slices.Contains(allowedDirection.ColoredCoins, tx.ColoredCoinID) {
			return fmt.Errorf(
				"colored coin (%d) not allowed for direction %s to %s",
				tx.ColoredCoinID, srcChainID, destChainID,
			)
		}
	}

	return nil
}
