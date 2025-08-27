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

func (c *BridgingAddressesCoordinatorImpl) GetAddressesAndAmounts(
	chainID uint8,
	cardanoCliBinary string,
	isRedistribution bool,
	protocolParams []byte,
	txOutputs *[]cardanowallet.TxOutput,
) ([]common.AddressAndAmount, error) {
	// Go through all addresses, sort them by the total amount of tokens (descending),
	// and choose the one with the biggest amount
	// Future improvement:
	// - add the stake pool saturation awareness
	var err error

	var amounts []common.AddressAndAmount

	if txOutputs == nil {
		return nil, fmt.Errorf("txOutputs cannot be nil")
	}

	if len(*txOutputs) == 0 && !isRedistribution {
		return nil, nil
	}

	remainingTokenAmounts := cardanowallet.GetOutputsSum(*txOutputs)
	requiredAdaAmount := remainingTokenAmounts[cardanowallet.AdaTokenName]

	c.logger.Debug("remainingTokenAmounts", remainingTokenAmounts)

	addrAmounts, potentialInputs, err := c.getTokensAmountByAddr(chainID, isRedistribution)
	if err != nil {
		return nil, err
	}

	if isRedistribution {
		amounts, err = c.redistributeTokens(remainingTokenAmounts, addrAmounts)
	} else {
		amounts, err = c.getAddressesAndAmountsToPayFrom(cardanoCliBinary, protocolParams, addrAmounts,
			remainingTokenAmounts, potentialInputs, txOutputs)
	}

	if err != nil {
		return nil, err
	}

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

func (c *BridgingAddressesCoordinatorImpl) getAddressesAndAmountsToPayFrom(
	cardanoCliBinary string,
	protocolParams []byte,
	addrAmounts []addrAmount,
	remainingTokenAmounts map[string]uint64,
	potentialInputs []*indexer.TxInputOutput,
	txOutputs *[]cardanowallet.TxOutput,
) ([]common.AddressAndAmount, error) {
	// If we have native tokens we need to add up min utxo lovelace amount
	// needed to be sent together with native token
	minUtxo := uint64(0)

	var err error

	if len(remainingTokenAmounts) > 1 {
		minUtxo, err = cardanotx.CalculateMinUtxoLovelaceAmount(
			cardanoCliBinary, protocolParams, addrAmounts[0].address, potentialInputs, *txOutputs)
		if err != nil {
			return nil, err
		}
	}

	// Sort by totalAmount descending
	sort.Slice(addrAmounts, func(i, j int) bool {
		return addrAmounts[i].
			totalTokenAmounts[cardanowallet.AdaTokenName] > addrAmounts[j].totalTokenAmounts[cardanowallet.AdaTokenName]
	})

	amounts := make([]common.AddressAndAmount, 0)

	c.logger.Debug("Available addresses to pay from", addrAmounts)

	for i := range addrAmounts {
		// Early exit if we've satisfied all required amounts
		if len(remainingTokenAmounts) == 1 && remainingTokenAmounts[cardanowallet.AdaTokenName] == 0 {
			break
		}

		c.logger.Debug("Processing address", i, "remainingTokenAmounts", remainingTokenAmounts)

		addrAmount := &addrAmounts[i]
		includeChange := minUtxo

		// Process native tokens
		nativeTokenChangeInUtxos := c.processNativeTokens(addrAmount, remainingTokenAmounts, minUtxo)

		// Process remaining lovelace amount
		requiredAdaAmount := remainingTokenAmounts[cardanowallet.AdaTokenName]
		if requiredAdaAmount > 0 {
			includeChange = c.processAdaAmount(
				addrAmount, remainingTokenAmounts, minUtxo, nativeTokenChangeInUtxos, i, addrAmounts, txOutputs)
		} else {
			addrAmount.includeInTx[cardanowallet.AdaTokenName] = 0
		}

		amounts = append(amounts, common.AddressAndAmount{
			Address:       addrAmount.address,
			AddressIndex:  addrAmount.addressIndex,
			TokensAmounts: addrAmount.includeInTx,
			IncludeChange: includeChange,
			UtxoCount:     addrAmount.utxoCount,
		})
	}

	return amounts, nil
}

func (c *BridgingAddressesCoordinatorImpl) redistributeAdaTokens(
	addrAmounts []addrAmount, totalAdaAmount uint64, requiredAdaAmount uint64,
) ([]common.AddressAndAmount, error) {
	if len(addrAmounts) == 0 {
		return nil, fmt.Errorf("cannot redistribute ADA tokens: no addresses provided")
	}

	remainigAdaAmount, ok := safeSubstract(totalAdaAmount, requiredAdaAmount)
	if !ok {
		return nil, fmt.Errorf("not enough ADA token funds to redistribute: available = %d, required = %d",
			totalAdaAmount,
			requiredAdaAmount)
	}

	adaTokensPerAddress := remainigAdaAmount / uint64(len(addrAmounts))
	extra := remainigAdaAmount % uint64(len(addrAmounts))

	addressAndAmounts := make([]common.AddressAndAmount, 0, len(addrAmounts))

	for i, addrAndAmount := range addrAmounts {
		addressAndAmount := common.AddressAndAmount{
			Address:       addrAndAmount.address,
			AddressIndex:  addrAndAmount.addressIndex,
			TokensAmounts: addrAndAmount.totalTokenAmounts,
			UtxoCount:     addrAndAmount.utxoCount,
		}

		addressAndAmount.TokensAmounts[cardanowallet.AdaTokenName] = adaTokensPerAddress
		if uint64(i) < extra { //nolint:gosec
			addressAndAmount.TokensAmounts[cardanowallet.AdaTokenName] += 1 // Distribute the remainder fairly
		}

		addressAndAmounts = append(addressAndAmounts, addressAndAmount)
	}

	return addressAndAmounts, nil
}

func (c *BridgingAddressesCoordinatorImpl) redistributeTokens(
	remainingTokenAmounts map[string]uint64, addrAmounts []addrAmount,
) ([]common.AddressAndAmount, error) {
	totalAdaAmount := uint64(0)
	for _, addrAmount := range addrAmounts {
		totalAdaAmount += addrAmount.totalTokenAmounts[cardanowallet.AdaTokenName]
	}

	requiredAdaAmount := remainingTokenAmounts[cardanowallet.AdaTokenName]
	delete(remainingTokenAmounts, cardanowallet.AdaTokenName)

	addressAndAmounts, err := c.redistributeAdaTokens(addrAmounts, totalAdaAmount, requiredAdaAmount)
	if err != nil {
		return nil, err
	}

	// subtract tokens that should be transferred
	for _, addrAmount := range addressAndAmounts {
		if len(remainingTokenAmounts) == 0 {
			break
		}

		// Update remaining token amounts
		for tokenName, requiredAmount := range remainingTokenAmounts {
			if addrAmount.TokensAmounts[tokenName] >= requiredAmount {
				addressChange, _ := safeSubstract(addrAmount.TokensAmounts[tokenName], requiredAmount)
				delete(remainingTokenAmounts, tokenName)

				if addressChange > 0 {
					addrAmount.TokensAmounts[tokenName] = addressChange
				} else {
					delete(addrAmount.TokensAmounts, tokenName)
				}
			} else {
				remainingTokenAmounts[tokenName] -= addrAmount.TokensAmounts[tokenName]
				delete(addrAmount.TokensAmounts, tokenName)
			}
		}
	}

	return addressAndAmounts, nil
}

func (c *BridgingAddressesCoordinatorImpl) getTokensAmountByAddr(
	chainID uint8, isRedistribution bool,
) ([]addrAmount, []*indexer.TxInputOutput, error) {
	db := c.dbs[common.ToStrChainID(chainID)]
	addresses := c.bridgingAddressesManager.GetAllPaymentAddresses(chainID)

	addrAmounts := make([]addrAmount, 0, len(addresses))
	potentialInputs := make([]*indexer.TxInputOutput, 0)

	// Calculate amount held by each address
	for i, address := range addresses {
		utxos, err := db.GetAllTxOutputs(address, true)
		if err != nil {
			return nil, nil, err
		}

		potentialInputs = append(potentialInputs, utxos...)

		totalTokenAmounts := cardanotx.GetSumMapFromTxInputOutput(utxos)

		addrAmounts = append(addrAmounts, addrAmount{
			address:           address,
			addressIndex:      uint8(i), //nolint:gosec
			totalTokenAmounts: totalTokenAmounts,
			includeInTx:       make(map[string]uint64),
			holdNativeTokens:  len(totalTokenAmounts) > 1 || isRedistribution,
			utxoCount:         len(utxos),
		})
	}

	return addrAmounts, potentialInputs, nil
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

			remainingTokenAmounts[tokenName] -= requiredAmount
			addrAmount.includeInTx[tokenName] = requiredAmount
		} else if addrAmount.totalTokenAmounts[tokenName] > 0 {
			// Take full amount, no change so we don't pay attention to lovelace
			remainingTokenAmounts[tokenName] -= addrAmount.totalTokenAmounts[tokenName]
			addrAmount.includeInTx[tokenName] = addrAmount.totalTokenAmounts[tokenName]
		}

		if remainingTokenAmounts[tokenName] == 0 {
			delete(remainingTokenAmounts, tokenName)
		}
	}

	return nativeTokenChangeInUtxos
}

// processAdaAmount handles the processing of lovelace amounts for a given address
// Updates the addrAmount.includeInTx map with the amount of lovelace to be included in the transaction
// Deducts the amount of lovelace from the remainingTokenAmounts map
// Returns the amount of lovelace to be included as change for this address during utxo selection
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
	c.logger.Debug("Address change", addrAmount.address, addressChange)

	if !ok && !nativeTokenChangeInUtxos {
		// Not enough lovelace on this address
		// Take all lovelace from this address
		addrAmount.includeInTx[cardanowallet.AdaTokenName] = availableAdaOnAddress
		remainingTokenAmounts[cardanowallet.AdaTokenName] -= availableAdaOnAddress

		return 0
	}

	if addressChange == 0 && !nativeTokenChangeInUtxos {
		// Exact amount that is needed
		addrAmount.includeInTx[cardanowallet.AdaTokenName] = requiredAdaAmount
		remainingTokenAmounts[cardanowallet.AdaTokenName] = 0

		return 0
	}

	if nativeTokenChangeInUtxos && (addressChange == 0 || !ok) {
		addrAmount.includeInTx[cardanowallet.AdaTokenName] = availableAdaOnAddress - requiredForChange
		remainingTokenAmounts[cardanowallet.AdaTokenName] -= availableAdaOnAddress - requiredForChange

		return requiredForChange
	}

	if addressChange >= requiredForChange {
		// Sufficient change amount, no need to carry over
		addrAmount.includeInTx[cardanowallet.AdaTokenName] = requiredAdaAmount
		remainingTokenAmounts[cardanowallet.AdaTokenName] = 0

		return requiredForChange
	}

	// Handle insufficient change
	return c.handleInsufficientChange(addrAmount, remainingTokenAmounts, availableAdaOnAddress,
		addressChange, requiredForChange, currentIndex, addrAmounts, txOutputs)
}

// handleInsufficientChange handles the case where there's insufficient change
// This can be handled in two ways:
// 1. If we can split the amount between the current and the next address
// 2. If we can't split the amount, we need to carry over the change to the next address
// Updates the addrAmount.includeInTx map with the amount of lovelace to be included in the transaction
// Deducts the amount of lovelace from the remainingTokenAmounts map
// Returns the amount of lovelace to be included as change for this address during utxo selection
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

	// Check if we have enough remaining lovelace after change
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

	// Handle carry over, send the change to the next address and include it in the transaction
	c.logger.Debug("Handle carry over")

	addrAmount.includeInTx[cardanowallet.AdaTokenName] = availableAdaOnAddress

	*txOutputs = append(*txOutputs, cardanowallet.NewTxOutput(
		addrAmounts[currentIndex+1].address,
		addressChange+addrAmounts[currentIndex+1].totalTokenAmounts[cardanowallet.AdaTokenName]))

	remainingTokenAmounts[cardanowallet.AdaTokenName] =
		addrAmounts[currentIndex+1].totalTokenAmounts[cardanowallet.AdaTokenName]

	return 0
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
