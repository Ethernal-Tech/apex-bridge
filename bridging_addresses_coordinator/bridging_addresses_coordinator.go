package bridgingaddressscoordinator

import (
	"fmt"
	"sort"

	cardanotx "github.com/Ethernal-Tech/apex-bridge/cardano"
	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
	cardanowallet "github.com/Ethernal-Tech/cardano-infrastructure/wallet"
	"github.com/hashicorp/go-hclog"
)

type addrAmount struct {
	address           string
	totalTokenAmounts map[string]uint64
	includeInTx       map[string]uint64
	addressIndex      uint8
	holdNativeTokens  bool
	utxoCount         int
}

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
	txOutputs *[]cardanowallet.TxOutput,
) ([]common.AddressAndAmount, error) {
	// Go through all addresses, sort them by the total amount of tokens (descending),
	// and choose the one with the biggest amount
	// Future improvement:
	// - add the stake pool saturation awareness
	var err error

	if len(*txOutputs) == 0 {
		return nil, nil
	}

	db := c.dbs[common.ToStrChainID(chainID)]
	addresses := c.bridgingAddressesManager.GetAllPaymentAddresses(chainID)

	remainingTokenAmounts := cardanowallet.GetOutputsSum(*txOutputs)
	containsTokens := len(remainingTokenAmounts) > 1
	c.logger.Debug("remainingTokenAmounts", remainingTokenAmounts)

	requiredAdaAmount := remainingTokenAmounts[cardanowallet.AdaTokenName]

	addrAmounts := make([]addrAmount, 0, len(addresses))
	potentialInputs := make([]*indexer.TxInputOutput, 0)

	// Calculate amount hold by each address
	for i, address := range addresses {
		utxos, err := db.GetAllTxOutputs(address, true)
		if err != nil {
			return nil, err
		}
		potentialInputs = append(potentialInputs, utxos...)

		totalTokenAmounts := make(map[string]uint64)
		holdNativeTokens := false

		for _, utxo := range utxos {
			totalTokenAmounts[cardanowallet.AdaTokenName] += utxo.Output.Amount

			for _, token := range utxo.Output.Tokens {
				if remainingTokenAmounts[token.TokenName()] > 0 {
					holdNativeTokens = true
					totalTokenAmounts[token.TokenName()] += token.Amount
				}
			}
		}

		addrAmounts = append(addrAmounts, addrAmount{
			address:           address,
			addressIndex:      uint8(i), //nolint:gosec
			totalTokenAmounts: totalTokenAmounts,
			includeInTx:       make(map[string]uint64),
			holdNativeTokens:  holdNativeTokens,
			utxoCount:         len(utxos),
		})
	}

	// If we have native tokens we need to add up min utxo lovelace amount
	// needed to be sent together with native token
	minUtxo := uint64(0)

	if containsTokens {
		minUtxo, err = cardanotx.CalculateMinUtxoLovelaceAmount(cardanoCliBinary, protocolParams, addresses[0], potentialInputs, *txOutputs)
		if err != nil {
			return nil, err
		}

		// minUtxo += common.MinUtxoAmountDefault
	}

	// Sort by totalAmount descending
	sort.Slice(addrAmounts, func(i, j int) bool {
		return addrAmounts[i].
			totalTokenAmounts[cardanowallet.AdaTokenName] > addrAmounts[j].totalTokenAmounts[cardanowallet.AdaTokenName]
	})

	amounts := make([]common.AddressAndAmount, 0)

	c.logger.Debug("addrAmounts", addrAmounts)

	for i := range addrAmounts {
		c.logger.Debug("remainingTokenAmounts loop", remainingTokenAmounts)
		// Early exit if we've satisfied all required amounts
		if len(remainingTokenAmounts) == 1 && remainingTokenAmounts[cardanowallet.AdaTokenName] == 0 {
			break
		}

		addrAmount := &addrAmounts[i]
		includeChange := uint64(0)

		c.logger.Debug("addrAmount Address", addrAmount.address)

		nativeTokenChangeInUtxos := c.processNativeTokens(addrAmount, remainingTokenAmounts, minUtxo)

		// Update remaining lovelace amount
		requiredAdaAmount := remainingTokenAmounts[cardanowallet.AdaTokenName]
		if requiredAdaAmount > 0 {
			includeChange = c.processAdaAmount(addrAmount, remainingTokenAmounts, minUtxo, nativeTokenChangeInUtxos, i, addrAmounts, txOutputs)
		} else {
			if nativeTokenChangeInUtxos {
				includeChange = minUtxo
			}
			addrAmount.includeInTx[cardanowallet.AdaTokenName] = 0
		}

		c.logger.Debug("------------------------------------------")

		amounts = append(amounts, common.AddressAndAmount{
			Address:       addrAmount.address,
			AddressIndex:  addrAmount.addressIndex,
			TokensAmounts: addrAmount.includeInTx,
			IncludeChnage: includeChange,
			UtxoCount:     addrAmount.utxoCount,
		})
	}

	c.logDebugAmounts(amounts)

	if remainingTokenAmounts[cardanowallet.AdaTokenName] > 0 {
		return nil, fmt.Errorf("%w: required %d, but missing %d",
			cardanowallet.ErrUTXOsCouldNotSelect, requiredAdaAmount, remainingTokenAmounts[cardanowallet.AdaTokenName])
	}

	for tokenName, remainingAmount := range remainingTokenAmounts {
		if remainingAmount > 0 {
			return nil, fmt.Errorf("not enough %s native token funds, missing: %v", tokenName, remainingTokenAmounts)
		}
	}

	return amounts, nil
}

// processNativeTokens handles the processing of native tokens for a given address
func (c *BridgingAddressesCoordinatorImpl) processNativeTokens(
	addrAmount *addrAmount,
	remainingTokenAmounts map[string]uint64,
	minUtxo uint64,
) bool {
	nativeTokenChangeInUtxos := false

	for tokenName, requiredAmount := range remainingTokenAmounts {
		if tokenName == cardanowallet.AdaTokenName {
			continue
		}

		if addrAmount.totalTokenAmounts[tokenName] > requiredAmount {
			nativeTokenChangeInUtxos = true
			// Take partial amount, be sure that we have enough lovelace for change
			if addrAmount.totalTokenAmounts[cardanowallet.AdaTokenName] < 2*minUtxo {
				// We don't have enough lovelace to cover the change
				continue
			}

			// Handle needed ada for change
			addrAmount.totalTokenAmounts[cardanowallet.AdaTokenName] -= minUtxo

			remainingTokenAmounts[tokenName] -= requiredAmount
			delete(remainingTokenAmounts, tokenName)

			addrAmount.includeInTx[tokenName] = requiredAmount
		} else if addrAmount.totalTokenAmounts[tokenName] > 0 {
			// Take full amount, no change so we don't pay attention to lovelace
			remainingTokenAmounts[tokenName] -= addrAmount.totalTokenAmounts[tokenName]
			addrAmount.includeInTx[tokenName] = addrAmount.totalTokenAmounts[tokenName]

			if remainingTokenAmounts[tokenName] == 0 {
				delete(remainingTokenAmounts, tokenName)
			}
		}
	}

	return nativeTokenChangeInUtxos
}

// processAdaAmount handles the processing of ADA amounts for a given address
func (c *BridgingAddressesCoordinatorImpl) processAdaAmount(
	addrAmount *addrAmount,
	remainingTokenAmounts map[string]uint64,
	minUtxo uint64,
	nativeTokenChangeInUtxos bool,
	currentIndex int,
	addrAmounts []addrAmount,
	txOutputs *[]cardanowallet.TxOutput,
) uint64 {
	requiredAdaAmount := remainingTokenAmounts[cardanowallet.AdaTokenName]
	availableAdaOnAddress := addrAmount.totalTokenAmounts[cardanowallet.AdaTokenName]

	requiredForChange := common.MinUtxoAmountDefault
	if nativeTokenChangeInUtxos {
		requiredForChange = minUtxo
	}

	addressChange, ok := safeSubstract(availableAdaOnAddress, requiredAdaAmount)
	c.logger.Debug("addressChange", addressChange)
	c.logger.Debug("nativeTokenChangeInUtxos", nativeTokenChangeInUtxos)

	if !ok {
		// Not enough ADA on this address
		addrAmount.includeInTx[cardanowallet.AdaTokenName] = availableAdaOnAddress
		remainingTokenAmounts[cardanowallet.AdaTokenName] -= availableAdaOnAddress
		return 0
	}

	if addressChange == 0 && !nativeTokenChangeInUtxos {
		// Exact amount
		addrAmount.includeInTx[cardanowallet.AdaTokenName] = requiredAdaAmount
		remainingTokenAmounts[cardanowallet.AdaTokenName] = 0
		return 0
	}

	if addressChange >= requiredForChange {
		// Sufficient change amount
		addrAmount.includeInTx[cardanowallet.AdaTokenName] = requiredAdaAmount
		remainingTokenAmounts[cardanowallet.AdaTokenName] = 0
		return requiredForChange
	}

	// Handle insufficient change - need carry over logic
	return c.handleInsufficientChange(addrAmount, remainingTokenAmounts, availableAdaOnAddress,
		addressChange, requiredForChange, currentIndex, addrAmounts, txOutputs)
}

// handleInsufficientChange handles the case where there's insufficient change and needs carry over
func (c *BridgingAddressesCoordinatorImpl) handleInsufficientChange(
	addrAmount *addrAmount,
	remainingTokenAmounts map[string]uint64,
	availableAdaOnAddress uint64,
	addressChange uint64,
	requiredForChange uint64,
	currentIndex int,
	addrAmounts []addrAmount,
	txOutputs *[]cardanowallet.TxOutput,
) uint64 {
	includeChange := requiredForChange + addressChange

	// Check if we can handle carry over (not the last address)
	if currentIndex > len(addrAmounts)-2 {
		// Last address, can't carry over
		addrAmount.includeInTx[cardanowallet.AdaTokenName] = availableAdaOnAddress
		remainingTokenAmounts[cardanowallet.AdaTokenName] -= availableAdaOnAddress
		return 0
	}

	// Check if we have enough remaining ADA after change
	if availableAdaOnAddress-includeChange > common.MinUtxoAmountDefault {
		nextAddressAda := addrAmounts[currentIndex+1].totalTokenAmounts[cardanowallet.AdaTokenName]
		remainingRequired := remainingTokenAmounts[cardanowallet.AdaTokenName] - availableAdaOnAddress + includeChange

		if nextAddressAda-remainingRequired >= common.MinUtxoAmountDefault {
			// Can split the amount
			addrAmount.includeInTx[cardanowallet.AdaTokenName] = availableAdaOnAddress - includeChange
			remainingTokenAmounts[cardanowallet.AdaTokenName] -= availableAdaOnAddress - includeChange
			return includeChange
		}
	}

	// Handle carry over
	c.logger.Debug("Handle carry over")
	addrAmount.includeInTx[cardanowallet.AdaTokenName] = availableAdaOnAddress
	*txOutputs = append(*txOutputs, cardanowallet.NewTxOutput(
		addrAmounts[currentIndex+1].address,
		addressChange+addrAmounts[currentIndex+1].totalTokenAmounts[cardanowallet.AdaTokenName]))
	remainingTokenAmounts[cardanowallet.AdaTokenName] = addrAmounts[currentIndex+1].totalTokenAmounts[cardanowallet.AdaTokenName]
	return 0
}

// logDebugAmounts logs the amounts for debugging
func (c *BridgingAddressesCoordinatorImpl) logDebugAmounts(amounts []common.AddressAndAmount) {
	c.logger.Debug("amounts:")
	for _, am := range amounts {
		c.logger.Debug("address", am.Address)
		c.logger.Debug("address index", am.AddressIndex)
		c.logger.Debug("tokens amount", am.TokensAmounts)
		c.logger.Debug("include change", am.IncludeChnage)
		c.logger.Debug("-------------------------------------")
	}
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

	return common.AddressAndAmount{Address: addresses[index], AddressIndex: uint8(index)}, nil //nolint:gosec
}

func (c *BridgingAddressesCoordinatorImpl) GetAllAddresses(chainID uint8) []string {
	// Return all addresses
	return c.bridgingAddressesManager.GetAllPaymentAddresses(chainID)
}
