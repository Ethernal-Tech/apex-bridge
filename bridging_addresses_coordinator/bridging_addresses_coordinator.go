package bridgingaddressscoordinator

import (
	"fmt"
	"sort"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
	cardanowallet "github.com/Ethernal-Tech/cardano-infrastructure/wallet"
)

type BridgingAddressesCoordinatorImpl struct {
	bridgingAddressesManager common.BridgingAddressesManager
	dbs                      map[string]indexer.Database
}

var _ common.BridgingAddressesCoordinator = (*BridgingAddressesCoordinatorImpl)(nil)

func NewBridgingAddressesCoordinator(
	bridgingAddressesManager common.BridgingAddressesManager,
	dbs map[string]indexer.Database,
) common.BridgingAddressesCoordinator {
	return &BridgingAddressesCoordinatorImpl{
		bridgingAddressesManager: bridgingAddressesManager,
		dbs:                      dbs,
	}
}

func (c *BridgingAddressesCoordinatorImpl) GetAddressesAndAmountsToPayFrom(chainID uint8, cardanoCliBinary string, protocolParams []byte, txOutputs []cardanowallet.TxOutput) ([]common.AddressAndAmount, error) {
	// Go through all addresses, sort them by the total amount of tokens (descending), and choose the one with the biggest amount
	// Future improvement:
	// - add the stake pool saturation awareness
	db := c.dbs[common.ToStrChainID(chainID)]
	addresses := c.bridgingAddressesManager.GetAllPaymentAddresses(chainID)

	// TODO: extract methods that do this from cco_utils
	containsTokens := false
	remainingTokenAmounts := make(map[string]uint64)
	for _, txOutput := range txOutputs {
		//remainingAdaAmount += txOutput.Amount
		remainingTokenAmounts[cardanowallet.AdaTokenName] += txOutput.Amount
		for _, token := range txOutput.Tokens {
			containsTokens = true
			remainingTokenAmounts[token.TokenName()] += token.Amount
		}
	}

	if containsTokens {
		tokens, err := cardanowallet.GetTokensFromSumMap(remainingTokenAmounts)
		if err != nil {
			return nil, err
		}

		txBuilder, err := cardanowallet.NewTxBuilder(cardanoCliBinary)
		if err != nil {
			return nil, err
		}

		defer txBuilder.Dispose()

		minUtxo, err := txBuilder.SetProtocolParameters(protocolParams).CalculateMinUtxo(cardanowallet.TxOutput{
			Addr:   txOutputs[0].Addr,
			Amount: remainingTokenAmounts[cardanowallet.AdaTokenName],
			Tokens: tokens,
		})
		if err != nil {
			return nil, err
		}

		remainingTokenAmounts[cardanowallet.AdaTokenName] += minUtxo
	}

	requiredAdaAmount := remainingTokenAmounts[cardanowallet.AdaTokenName]

	type addrAmount struct {
		address           string
		totalTokenAmounts map[string]uint64
		addressIndex      uint8
	}

	addrAmounts := make([]addrAmount, 0, len(addresses))

	for i, address := range addresses {
		utxos, err := db.GetAllTxOutputs(address, true)
		if err != nil {
			return nil, err
		}

		totalTokenAmounts := make(map[string]uint64)
		for _, utxo := range utxos {
			totalTokenAmounts[cardanowallet.AdaTokenName] += utxo.Output.Amount
			for _, token := range utxo.Output.Tokens {
				if remainingTokenAmounts[token.TokenName()] > 0 {
					totalTokenAmounts[token.TokenName()] += token.Amount
				}
			}
		}

		addrAmounts = append(addrAmounts, addrAmount{
			address:           address,
			addressIndex:      uint8(i),
			totalTokenAmounts: totalTokenAmounts,
		})
	}

	// Sort by totalAmount descending
	sort.Slice(addrAmounts, func(i, j int) bool {
		return addrAmounts[i].totalTokenAmounts[cardanowallet.AdaTokenName] > addrAmounts[j].totalTokenAmounts[cardanowallet.AdaTokenName]
	})

	amounts := make([]common.AddressAndAmount, 0)

	for _, addrAmount := range addrAmounts {
		if remainingTokenAmounts[cardanowallet.AdaTokenName] == 0 && (len(remainingTokenAmounts) == 0 && containsTokens) {
			break
		}

		// Check if this address has enough tokens to cover remaining token amounts
		// canHandleTokens := true
		// for tokenPolicyID, requiredAmount := range remainingTokenAmounts {
		// 	if addrAmount.totalTokenAmounts[tokenPolicyID] < requiredAmount {
		// 		canHandleTokens = false
		// 		break
		// 	}
		// }
		//
		// if !canHandleTokens {
		// 	continue // Skip this address if it can't handle the required tokens
		// }

		// Update remaining token amounts
		for tokenName, requiredAmount := range remainingTokenAmounts {
			if tokenName == cardanowallet.AdaTokenName {
				// TODO: check if we need GetMinUtxoAmount(chainID)
				if addrAmount.totalTokenAmounts[tokenName] >= requiredAmount+common.MinUtxoAmountDefault {
					delete(remainingTokenAmounts, tokenName)
					addrAmount.totalTokenAmounts[tokenName] = requiredAmount
				} else {
					remainingTokenAmounts[tokenName] -= addrAmount.totalTokenAmounts[tokenName]
				}
			}

			if addrAmount.totalTokenAmounts[tokenName] >= requiredAmount {
				delete(remainingTokenAmounts, tokenName)
				addrAmount.totalTokenAmounts[tokenName] = requiredAmount
			} else {
				remainingTokenAmounts[tokenName] -= addrAmount.totalTokenAmounts[tokenName]
			}
		}

		amounts = append(amounts, common.AddressAndAmount{
			Address:       addrAmount.address,
			AddressIndex:  addrAmount.addressIndex,
			TokensAmounts: addrAmount.totalTokenAmounts,
		})
	}

	if remainingTokenAmounts[cardanowallet.AdaTokenName] > 0 {
		return nil, fmt.Errorf("not enough lovelace funds, required: %d, remaining: %d", requiredAdaAmount, remainingTokenAmounts[cardanowallet.AdaTokenName])
	}

	if len(remainingTokenAmounts) > 0 {
		return nil, fmt.Errorf("not enough native token funds, missing: %v", remainingTokenAmounts)
	}

	return amounts, nil
}

func (c *BridgingAddressesCoordinatorImpl) GetAddressesAndAmountsToStakeTo(chainID uint8, amount uint64) ([]common.AddressAndAmount, error) {
	// Go through all addresses and find the one with the least amount of tokens
	// chose that one and send whole amount to it
	// Future improvement:
	// - add the stake pool saturation awareness

	db := c.dbs[common.ToStrChainID(chainID)]
	addresses := c.bridgingAddressesManager.GetAllPaymentAddresses(chainID)
	amounts := make([]common.AddressAndAmount, 0)

	minAmount := uint64(0)
	index := 0
	for i, address := range addresses {
		utxos, err := db.GetAllTxOutputs(address, true)
		if err != nil {
			return nil, err
		}

		amount := uint64(0)
		for _, utxo := range utxos {
			amount += utxo.Output.Amount
		}

		if amount == 0 {
			amounts = append(amounts, common.AddressAndAmount{Address: address, AddressIndex: uint8(i)})
			return amounts, nil
		}

		if i == 0 {
			minAmount = amount
		} else if amount < minAmount {
			minAmount = amount
			index = i
		}
	}

	amounts = append(amounts, common.AddressAndAmount{Address: addresses[index], AddressIndex: uint8(index)})

	return amounts, nil
}
