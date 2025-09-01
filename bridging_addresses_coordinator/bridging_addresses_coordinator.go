package bridgingaddressscoordinator

import (
	"fmt"
	"sort"

	cardanotx "github.com/Ethernal-Tech/apex-bridge/cardano"
	"github.com/Ethernal-Tech/apex-bridge/common"
	oracleCore "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
	cardanowallet "github.com/Ethernal-Tech/cardano-infrastructure/wallet"
	"github.com/hashicorp/go-hclog"
)

type addrAmount struct {
	address           string
	totalTokenAmounts map[string]uint64
	includeInTx       map[string]uint64
	addressIndex      uint8
	utxoCount         int
}

type tokensAmountPerAddress struct {
	addrAmounts     []addrAmount
	potentialInputs []*indexer.TxInputOutput
	sum             map[string]uint64
}

type BridgingAddressesCoordinatorImpl struct {
	bridgingAddressesManager common.BridgingAddressesManager
	dbs                      map[string]indexer.Database
	cardanoChains            map[string]*oracleCore.CardanoChainConfig
	logger                   hclog.Logger
}

var _ common.BridgingAddressesCoordinator = (*BridgingAddressesCoordinatorImpl)(nil)

func NewBridgingAddressesCoordinator(
	bridgingAddressesManager common.BridgingAddressesManager,
	dbs map[string]indexer.Database,
	cardanoChains map[string]*oracleCore.CardanoChainConfig,
	logger hclog.Logger,
) common.BridgingAddressesCoordinator {
	return &BridgingAddressesCoordinatorImpl{
		bridgingAddressesManager: bridgingAddressesManager,
		dbs:                      dbs,
		cardanoChains:            cardanoChains,
		logger:                   logger,
	}
}

func (c *BridgingAddressesCoordinatorImpl) GetAddressesAndAmountsForBatch(
	chainID uint8,
	cardanoCliBinary string,
	isRedistribution bool,
	protocolParams []byte,
	txOutputs *cardanotx.TxOutputs,
) ([]common.AddressAndAmount, error) {
	// Go through all addresses, sort them by the total amount of tokens (descending),
	// and choose the one with the biggest amount
	var err error

	var amounts []common.AddressAndAmount

	if txOutputs == nil {
		return nil, fmt.Errorf("txOutputs cannot be nil")
	}

	if len(txOutputs.Outputs) == 0 && !isRedistribution {
		return nil, nil
	}

	requiredTokenAmounts := cardanowallet.GetOutputsSum(txOutputs.Outputs)
	requiredCurrencyAmount := requiredTokenAmounts[cardanowallet.AdaTokenName]

	c.logger.Debug("GetAddressesAndAmountsForBatch", "chain", common.ToStrChainID(chainID),
		"requiredTokenAmounts", requiredTokenAmounts)

	totalTokenAmounts, err := c.getTokensAmountByAddr(chainID)
	if err != nil {
		return nil, err
	}

	changeMinUtxo := common.MinUtxoAmountDefault
	if len(requiredTokenAmounts) > 1 {
		changeMinUtxo, err = cardanotx.CalculateMinUtxoLovelaceAmount(
			cardanoCliBinary, protocolParams, totalTokenAmounts.addrAmounts[0].address,
			totalTokenAmounts.potentialInputs, txOutputs.Outputs)
		if err != nil {
			return nil, err
		}
	}

	for token, amount := range requiredTokenAmounts {
		if token == cardanowallet.AdaTokenName {
			if amount != totalTokenAmounts.sum[token] &&
				amount > totalTokenAmounts.sum[token]-changeMinUtxo {
				return nil, fmt.Errorf("not enough %s token funds for batch: available = %d, required = %d",
					token, totalTokenAmounts.sum[token], amount)
			}
		}

		if amount > totalTokenAmounts.sum[token] {
			return nil, fmt.Errorf("not enough %s token funds for batch: available = %d, required = %d",
				token, totalTokenAmounts.sum[token], amount)
		}
	}

	if isRedistribution {
		amounts, err = c.redistributeTokens(requiredTokenAmounts, totalTokenAmounts.addrAmounts)
	} else {
		amounts, err = c.getAddressesAndAmountsToPayFrom(
			totalTokenAmounts.addrAmounts, totalTokenAmounts.sum[cardanowallet.AdaTokenName],
			requiredTokenAmounts, &txOutputs.Outputs)
	}

	if err != nil {
		return nil, err
	}

	if requiredTokenAmounts[cardanowallet.AdaTokenName] > 0 {
		return nil, fmt.Errorf("%w: required %d, but missing %d",
			cardanowallet.ErrUTXOsCouldNotSelect, requiredCurrencyAmount, requiredTokenAmounts[cardanowallet.AdaTokenName])
	}

	for tokenName, remainingAmount := range requiredTokenAmounts {
		if remainingAmount > 0 {
			return nil, fmt.Errorf("not enough %s native token funds, missing: %v", tokenName, requiredTokenAmounts)
		}
	}

	return amounts, nil
}

func (c *BridgingAddressesCoordinatorImpl) getAddressesAndAmountsToPayFrom(
	addrAmounts []addrAmount,
	totalCurrecnyAmountAvailableOnAddresses uint64,
	requiredTokenAmounts map[string]uint64,
	txOutputs *[]cardanowallet.TxOutput,
) ([]common.AddressAndAmount, error) {
	nativeTokensAvailable := len(addrAmounts[0].totalTokenAmounts) > 1

	// Sort by total lovelace amount descending
	sort.SliceStable(addrAmounts, func(i, j int) bool {
		return addrAmounts[i].
			totalTokenAmounts[cardanowallet.AdaTokenName] > addrAmounts[j].totalTokenAmounts[cardanowallet.AdaTokenName]
	})

	amounts := make([]common.AddressAndAmount, 0)

	c.logger.Debug("Available addresses to pay from", addrAmounts)

	for i := range addrAmounts {
		// Early exit if we've satisfied all required amounts
		if len(requiredTokenAmounts) == 0 {
			break
		}

		c.logger.Debug("Processing address", i, "requiredTokenAmounts", requiredTokenAmounts)

		addrAmount := &addrAmounts[i]
		includeChange := common.MinUtxoAmountDefault
		requiredCurrencyAmount := requiredTokenAmounts[cardanowallet.AdaTokenName]

		// Process native tokens only from frist address if there are any
		processCurrency := true
		predefinedCurrencyAmount := uint64(0)

		if addrAmount.addressIndex == 0 && nativeTokensAvailable {
			c.processNativeTokens(addrAmount, requiredTokenAmounts)

			// Check if we need to take some currency from this address
			availableOnRestOfTheAddresses :=
				totalCurrecnyAmountAvailableOnAddresses - addrAmount.totalTokenAmounts[cardanowallet.AdaTokenName]
			change, _ := safeSubtract(availableOnRestOfTheAddresses, common.MinUtxoAmountDefault)

			if change >= requiredCurrencyAmount {
				processCurrency = false
			} else {
				predefinedCurrencyAmount = requiredCurrencyAmount - availableOnRestOfTheAddresses
			}
		}

		// Process remaining lovelace amount
		if requiredCurrencyAmount > 0 && processCurrency {
			includeChange = c.processCurrencyAmount(
				addrAmount, predefinedCurrencyAmount, requiredTokenAmounts, i, addrAmounts, txOutputs)

			if requiredTokenAmounts[cardanowallet.AdaTokenName] == 0 {
				delete(requiredTokenAmounts, cardanowallet.AdaTokenName)
			}
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

func (c *BridgingAddressesCoordinatorImpl) redistributeCurrencyTokens(
	addrAmounts []addrAmount, totalCurrencyAmount uint64, requiredCurrencyAmount uint64,
) ([]common.AddressAndAmount, error) {
	if len(addrAmounts) == 0 {
		return nil, fmt.Errorf("cannot redistribute currency tokens: no addresses provided")
	}

	remainingCurrencyAmount := totalCurrencyAmount - requiredCurrencyAmount
	currencyTokensPerAddress := remainingCurrencyAmount / uint64(len(addrAmounts))
	extra := remainingCurrencyAmount % uint64(len(addrAmounts))

	addressAndAmounts := make([]common.AddressAndAmount, 0, len(addrAmounts))

	for i, addrAndAmount := range addrAmounts {
		addressAndAmount := common.AddressAndAmount{
			Address:       addrAndAmount.address,
			AddressIndex:  addrAndAmount.addressIndex,
			TokensAmounts: addrAndAmount.totalTokenAmounts,
			UtxoCount:     addrAndAmount.utxoCount,
		}

		addressAndAmount.TokensAmounts[cardanowallet.AdaTokenName] = currencyTokensPerAddress
		if uint64(i) < extra { //nolint:gosec
			addressAndAmount.TokensAmounts[cardanowallet.AdaTokenName] += 1 // Distribute the remainder fairly
		}

		addressAndAmounts = append(addressAndAmounts, addressAndAmount)
	}

	return addressAndAmounts, nil
}

func (c *BridgingAddressesCoordinatorImpl) redistributeTokens(
	requiredTokenAmounts map[string]uint64, addrAmounts []addrAmount,
) ([]common.AddressAndAmount, error) {
	totalCurrencyAmount := uint64(0)
	for _, addrAmount := range addrAmounts {
		totalCurrencyAmount += addrAmount.totalTokenAmounts[cardanowallet.AdaTokenName]
	}

	requiredCurrencyAmount := requiredTokenAmounts[cardanowallet.AdaTokenName]
	delete(requiredTokenAmounts, cardanowallet.AdaTokenName)

	addressAndAmounts, err := c.redistributeCurrencyTokens(addrAmounts, totalCurrencyAmount, requiredCurrencyAmount)
	if err != nil {
		return nil, err
	}

	// subtract tokens that should be transferred
	for _, addrAmount := range addressAndAmounts {
		// Update remaining token amounts
		for tokenName, requiredAmount := range requiredTokenAmounts {
			if addrAmount.TokensAmounts[tokenName] >= requiredAmount {
				addressChange, _ := safeSubtract(addrAmount.TokensAmounts[tokenName], requiredAmount)
				delete(requiredTokenAmounts, tokenName)

				if addressChange > 0 {
					addrAmount.TokensAmounts[tokenName] = addressChange
				} else {
					delete(addrAmount.TokensAmounts, tokenName)
				}
			} else {
				requiredTokenAmounts[tokenName] -= addrAmount.TokensAmounts[tokenName]
				delete(addrAmount.TokensAmounts, tokenName)
			}
		}
	}

	return addressAndAmounts, nil
}

func (c *BridgingAddressesCoordinatorImpl) getTokensAmountByAddr(
	chainID uint8,
) (*tokensAmountPerAddress, error) {
	chainIDStr := common.ToStrChainID(chainID)

	db, ok := c.dbs[chainIDStr]
	if !ok {
		return nil, fmt.Errorf("failed to get appropriate db for chain %s", chainIDStr)
	}

	addresses := c.bridgingAddressesManager.GetAllPaymentAddresses(chainID)

	addrAmounts := make([]addrAmount, 0, len(addresses))
	potentialInputs := make([]*indexer.TxInputOutput, 0)
	sum := make(map[string]uint64)

	config, ok := c.cardanoChains[chainIDStr]
	if !ok {
		return nil, fmt.Errorf("failed to get appropriate config for chain %s", chainIDStr)
	}

	knownTokens, err := cardanotx.GetKnownTokens(&config.CardanoChainConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to get known tokens: %w for chain %s", err, chainIDStr)
	}

	// Calculate amount held by each address
	for i, address := range addresses {
		utxos, err := db.GetAllTxOutputs(address, true)
		if err != nil {
			return nil, err
		}

		if i == 0 {
			utxos = cardanotx.FilterOutUtxosWithUnknownTokens(utxos, knownTokens...)
		} else {
			utxos = cardanotx.FilterOutUtxosWithUnknownTokens(utxos)
		}

		potentialInputs = append(potentialInputs, utxos...)

		totalTokenAmounts := cardanotx.GetSumMapFromTxInputOutput(utxos)

		for token, amount := range totalTokenAmounts {
			sum[token] += amount
		}

		addrAmounts = append(addrAmounts, addrAmount{
			address:           address,
			addressIndex:      uint8(i), //nolint:gosec
			totalTokenAmounts: totalTokenAmounts,
			includeInTx:       make(map[string]uint64),
			utxoCount:         len(utxos),
		})
	}

	return &tokensAmountPerAddress{
		addrAmounts:     addrAmounts,
		potentialInputs: potentialInputs,
		sum:             sum,
	}, nil
}

// processNativeTokens handles the processing of native tokens for a given address
func (c *BridgingAddressesCoordinatorImpl) processNativeTokens(
	addrAmount *addrAmount,
	requiredTokenAmounts map[string]uint64,
) {
	for tokenName, requiredAmount := range requiredTokenAmounts {
		if tokenName == cardanowallet.AdaTokenName {
			continue
		}

		if addrAmount.totalTokenAmounts[tokenName] >= requiredAmount {
			// Take partial amount and satisfy
			requiredTokenAmounts[tokenName] -= requiredAmount
			addrAmount.includeInTx[tokenName] = requiredAmount
		} else if addrAmount.totalTokenAmounts[tokenName] > 0 {
			// Take full amount, no change so we don't pay attention to lovelace
			requiredTokenAmounts[tokenName] -= addrAmount.totalTokenAmounts[tokenName]
			addrAmount.includeInTx[tokenName] = addrAmount.totalTokenAmounts[tokenName]
		}

		if requiredTokenAmounts[tokenName] == 0 {
			delete(requiredTokenAmounts, tokenName)
		}
	}
}

// processCurrencyAmount handles the processing of lovelace amounts for a given address
// Updates the addrAmount.includeInTx map with the amount of lovelace to be included in the transaction
// Deducts the amount of lovelace from the requiredTokenAmounts map
// Returns the amount of lovelace to be included as change for this address during utxo selection
func (c *BridgingAddressesCoordinatorImpl) processCurrencyAmount(
	addrAmount *addrAmount,
	predefinedCurrencyAmount uint64,
	requiredTokenAmounts map[string]uint64,
	currentIndex int,
	addrAmounts []addrAmount,
	txOutputs *[]cardanowallet.TxOutput,
) uint64 {
	requiredCurrencyAmount := predefinedCurrencyAmount
	if requiredCurrencyAmount == 0 {
		requiredCurrencyAmount = requiredTokenAmounts[cardanowallet.AdaTokenName]
	}

	availableCurrencyOnAddress := addrAmount.totalTokenAmounts[cardanowallet.AdaTokenName]

	requiredForChange := common.MinUtxoAmountDefault
	addressChange, ok := safeSubtract(availableCurrencyOnAddress, requiredCurrencyAmount)
	c.logger.Debug("Address change", addrAmount.address, addressChange)

	if !ok {
		// Not enough lovelace on this address
		// Take all lovelace from this address
		addrAmount.includeInTx[cardanowallet.AdaTokenName] = availableCurrencyOnAddress
		requiredTokenAmounts[cardanowallet.AdaTokenName] -= availableCurrencyOnAddress

		return 0
	}

	if addressChange == 0 {
		// Exact amount that is needed
		addrAmount.includeInTx[cardanowallet.AdaTokenName] = requiredCurrencyAmount
		requiredTokenAmounts[cardanowallet.AdaTokenName] = 0

		return 0
	}

	if addressChange >= requiredForChange {
		// Sufficient change amount, no need to carry over
		addrAmount.includeInTx[cardanowallet.AdaTokenName] = requiredCurrencyAmount
		requiredTokenAmounts[cardanowallet.AdaTokenName] -= requiredCurrencyAmount

		return requiredForChange
	}

	// Handle insufficient change
	return c.handleInsufficientChange(addrAmount, requiredTokenAmounts, availableCurrencyOnAddress,
		addressChange, requiredForChange, currentIndex, addrAmounts, txOutputs)
}

// handleInsufficientChange handles the case where there's insufficient change
// This can be handled in two ways:
// 1. If we can split the amount between the current and the next address
// 2. If we can't split the amount, we need to carry over the change to the next address
// Updates the addrAmount.includeInTx map with the amount of lovelace to be included in the transaction
// Deducts the amount of lovelace from the requiredTokenAmounts map
// Returns the amount of lovelace to be included as change for this address during utxo selection
func (c *BridgingAddressesCoordinatorImpl) handleInsufficientChange(
	addrAmount *addrAmount,
	requiredTokenAmounts map[string]uint64,
	availableCurrencyOnAddress uint64,
	addressChange uint64,
	requiredForChange uint64,
	currentIndex int,
	addrAmounts []addrAmount,
	txOutputs *[]cardanowallet.TxOutput,
) uint64 {
	includeChange := requiredForChange + addressChange

	// Check if we have enough remaining lovelace after change
	if availableCurrencyOnAddress-includeChange > common.MinUtxoAmountDefault {
		nextAddressCurrencyAmount := addrAmounts[currentIndex+1].totalTokenAmounts[cardanowallet.AdaTokenName]
		remainingRequired := requiredTokenAmounts[cardanowallet.AdaTokenName] - availableCurrencyOnAddress + includeChange

		if nextAddressCurrencyAmount-remainingRequired >= common.MinUtxoAmountDefault {
			// Can split the amount
			addrAmount.includeInTx[cardanowallet.AdaTokenName] = availableCurrencyOnAddress - includeChange
			requiredTokenAmounts[cardanowallet.AdaTokenName] -= availableCurrencyOnAddress - includeChange

			return includeChange
		}
	}

	// Handle carry over, send the change to the next address and include it in the transaction
	c.logger.Debug("Handle carry over")

	addrAmount.includeInTx[cardanowallet.AdaTokenName] = availableCurrencyOnAddress

	*txOutputs = append(*txOutputs, cardanowallet.NewTxOutput(
		addrAmounts[currentIndex+1].address,
		addressChange+addrAmounts[currentIndex+1].totalTokenAmounts[cardanowallet.AdaTokenName]))

	requiredTokenAmounts[cardanowallet.AdaTokenName] =
		addrAmounts[currentIndex+1].totalTokenAmounts[cardanowallet.AdaTokenName]

	return 0
}

func safeSubtract(a, b uint64) (uint64, bool) {
	if a >= b {
		return a - b, true
	}

	return 0, false
}

func (c *BridgingAddressesCoordinatorImpl) GetAddressToBridgeTo(
	chainID uint8,
	containsNativeTokens bool,
) (common.AddressAndAmount, error) {
	// Go through all addresses and find the one with the least amount of tokens
	// chose that one and send whole amount to it
	db := c.dbs[common.ToStrChainID(chainID)]
	addresses := c.bridgingAddressesManager.GetAllPaymentAddresses(chainID)

	if containsNativeTokens {
		return common.AddressAndAmount{Address: addresses[0], AddressIndex: 0}, nil
	}

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
