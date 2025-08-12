package bridgingaddressscoordinator

import (
	"fmt"
	"sort"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
	cardanowallet "github.com/Ethernal-Tech/cardano-infrastructure/wallet"
	"github.com/hashicorp/go-hclog"
)

type BridgingAddressesCoordinatorImpl struct {
	bridgingAddressesManager common.BridgingAddressesManager
	dbs                      map[string]indexer.Database
	logger                   hclog.Logger
}

var _ common.BridgingAddressesCoordinator = (*BridgingAddressesCoordinatorImpl)(nil)

func NewBridgingAddressesCoordinator(
	bridgingAddressesManager common.BridgingAddressesManager,
	dbs map[string]indexer.Database,
	logger hclog.Logger,
) common.BridgingAddressesCoordinator {
	return &BridgingAddressesCoordinatorImpl{
		bridgingAddressesManager: bridgingAddressesManager,
		dbs:                      dbs,
		logger:                   logger,
	}
}

func (c *BridgingAddressesCoordinatorImpl) GetAddressesAndAmountsToPayFrom(
	chainID uint8,
	cardanoCliBinary string,
	protocolParams []byte,
	txOutputs []cardanowallet.TxOutput,
) ([]common.AddressAndAmount, error) {
	// Go through all addresses, sort them by the total amount of tokens (descending),
	// and choose the one with the biggest amount
	// Future improvement:
	// - add the stake pool saturation awareness
	if len(txOutputs) == 0 {
		return nil, nil
	}

	db := c.dbs[common.ToStrChainID(chainID)]
	addresses := c.bridgingAddressesManager.GetAllPaymentAddresses(chainID)

	// TODO: should we extract methods that do this from cco_utils
	//containsTokens := false
	remainingTokenAmounts := make(map[string]uint64)

	for _, txOutput := range txOutputs {
		remainingTokenAmounts[cardanowallet.AdaTokenName] += txOutput.Amount

		for _, token := range txOutput.Tokens {
			//containsTokens = true
			remainingTokenAmounts[token.TokenName()] += token.Amount
		}
	}

	c.logger.Debug("remainingTokenAmounts", remainingTokenAmounts)

	// If we have native tokens we need to add up min utxo lovelace amount
	// needed to be sent together with native token
	/* if containsTokens {
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
	} */

	requiredAdaAmount := remainingTokenAmounts[cardanowallet.AdaTokenName]

	type addrAmount struct {
		address           string
		totalTokenAmounts map[string]uint64
		addressIndex      uint8
	}

	addrAmounts := make([]addrAmount, 0, len(addresses))

	// Calculate amount hold by each address
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
			addressIndex:      uint8(i), //nolint:gosec
			totalTokenAmounts: totalTokenAmounts,
		})
	}

	// Sort by totalAmount descending
	sort.Slice(addrAmounts, func(i, j int) bool {
		return addrAmounts[i].
			totalTokenAmounts[cardanowallet.AdaTokenName] > addrAmounts[j].totalTokenAmounts[cardanowallet.AdaTokenName]
	})

	amounts := make([]common.AddressAndAmount, 0)

	/* if remainingTokenAmounts[cardanowallet.AdaTokenName] != 0 {
		remainingTokenAmounts[cardanowallet.AdaTokenName] += common.MinUtxoAmountDefault
	} */

	c.logger.Debug("remainingTokenAmounts1", remainingTokenAmounts)
	c.logger.Debug("addrAmounts", addrAmounts)

	// Pick addresses and amounts to be taken from
	for _, addrAmount := range addrAmounts {
		if len(remainingTokenAmounts) == 0 {
			break
		}

		fullAmount := false

		c.logger.Debug("addrAmount Address", addrAmount.address)

		// Update remaining token amounts
		for tokenName, requiredAmount := range remainingTokenAmounts {
			c.logger.Debug("remainingTokenAmounts requiredAmount", requiredAmount)

			if addrAmount.totalTokenAmounts[tokenName] >= requiredAmount {
				addressChange, ok := safeSubstract(addrAmount.totalTokenAmounts[tokenName], requiredAmount)
				c.logger.Debug("address change", addressChange)

				if ok && addressChange > common.MinUtxoAmountDefault {
					c.logger.Debug("delete from remainingTokenAmounts", tokenName)
					delete(remainingTokenAmounts, tokenName)
				} else if ok {
					/* requiredAmount += addressChange
					fullAmount = true */
					requiredAmount += addressChange - common.MinUtxoAmountDefault
					newAddrChange, _ := safeSubstract(addrAmount.totalTokenAmounts[tokenName], requiredAmount)
					c.logger.Debug("requiredAmount", requiredAmount, "newAddrChange", newAddrChange)

					remainingTokenAmounts[tokenName] -= requiredAmount
					c.logger.Debug("remainingTokenAmounts for token", tokenName, "amount", remainingTokenAmounts[tokenName])
				}

				/* if tokenName == cardanowallet.AdaTokenName && requiredAmount < common.MinUtxoAmountDefault {
					requiredAmount = common.MinUtxoAmountDefault
					c.logger.Debug("requiredAmount is MinUtxoAmountDefault", requiredAmount)
				} */

				addrAmount.totalTokenAmounts[tokenName] = requiredAmount
				c.logger.Debug("new addr amount for token", tokenName, "amount", addrAmount.totalTokenAmounts[tokenName])
			} else {
				remainingTokenAmounts[tokenName] -= addrAmount.totalTokenAmounts[tokenName]

				fullAmount = true
				c.logger.Debug("[FULL] addr amount for token", tokenName, "amount", addrAmount.totalTokenAmounts[tokenName])
				c.logger.Debug("[FULL] remainingTokenAmounts for token", tokenName, "amount", remainingTokenAmounts[tokenName])
			}
		}

		c.logger.Debug("------------------------------------------")

		amounts = append(amounts, common.AddressAndAmount{
			Address:       addrAmount.address,
			AddressIndex:  addrAmount.addressIndex,
			TokensAmounts: addrAmount.totalTokenAmounts,
			FullAmount:    fullAmount,
		})
	}

	c.logger.Debug("amounts:")

	for _, am := range amounts {
		c.logger.Debug("address", am.Address)
		c.logger.Debug("address index", am.AddressIndex)
		c.logger.Debug("tokens amount", am.TokensAmounts)
		c.logger.Debug("full amount", am.FullAmount)
		c.logger.Debug("-------------------------------------")
	}

	if remainingTokenAmounts[cardanowallet.AdaTokenName] > 0 {
		return nil, fmt.Errorf("not enough lovelace funds, required: %d, remaining: %d",
			requiredAdaAmount, remainingTokenAmounts[cardanowallet.AdaTokenName])
	}

	if len(remainingTokenAmounts) > 0 {
		return nil, fmt.Errorf("not enough native token funds, missing: %v", remainingTokenAmounts)
	}

	return amounts, nil
}

func safeSubstract(a, b uint64) (uint64, bool) {
	if a >= b {
		return a - b, true
	}

	return 0, false
}

func (c *BridgingAddressesCoordinatorImpl) GetAddressesAndAmountsToStakeTo(
	chainID uint8, amount uint64,
) (common.AddressAndAmount, error) {
	// Go through all addresses and find the one with the least amount of tokens
	// chose that one and send whole amount to it
	// Future improvement:
	// - add the stake pool saturation awareness
	db := c.dbs[common.ToStrChainID(chainID)]
	addresses := c.bridgingAddressesManager.GetAllPaymentAddresses(chainID)

	minAmount := uint64(0)
	index := 0

	for i, address := range addresses {
		utxos, err := db.GetAllTxOutputs(address, true)
		if err != nil {
			return common.AddressAndAmount{}, err
		}

		amount := uint64(0)
		for _, utxo := range utxos {
			amount += utxo.Output.Amount
		}

		if amount == 0 {
			return common.AddressAndAmount{Address: address, AddressIndex: uint8(i)}, nil //nolint:gosec
		}

		if i == 0 {
			minAmount = amount
		} else if amount < minAmount {
			minAmount = amount
			index = i
		}
	}

	return common.AddressAndAmount{Address: addresses[index], AddressIndex: uint8(index)}, nil
}

func (c *BridgingAddressesCoordinatorImpl) GetAllAddresses(chainID uint8) []string {
	// Return all addresses
	return c.bridgingAddressesManager.GetAllPaymentAddresses(chainID)
}
