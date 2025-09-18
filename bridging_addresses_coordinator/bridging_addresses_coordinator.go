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
	address            string
	totalTokenAmounts  map[string]uint64
	includeInTx        map[string]uint64
	addressIndex       uint8
	utxoCount          int
	insufficientChange bool
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

func (c *BridgingAddressesCoordinatorImpl) GetAddressToBridgeTo(
	chainID uint8,
	containsNativeTokens bool,
	isReward bool,
) (common.AddressAndAmount, error) {
	// Go through all addresses and find the one with the least amount of tokens
	// chose that one and send whole amount to it

	addressType := common.AddressTypeNormal
	firstAddressIndex := uint8(0)

	if isReward {
		addressType = common.AddressTypeReward
		firstAddressIndex = common.FirstRewardBridgingAddressIndex
	}

	addresses := c.bridgingAddressesManager.GetAllPaymentAddresses(chainID, addressType)

	if containsNativeTokens {
		return common.AddressAndAmount{Address: addresses[0], AddressIndex: firstAddressIndex}, nil
	}

	minAmount := uint64(0)
	index := 0
	db := c.dbs[common.ToStrChainID(chainID)]

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
			return common.AddressAndAmount{Address: address, AddressIndex: uint8(i) + firstAddressIndex}, nil //nolint:gosec
		}

		if i == 0 {
			minAmount = amount
		} else if amount < minAmount {
			minAmount = amount
			index = i
		}
	}

	return common.AddressAndAmount{Address: addresses[index], AddressIndex: uint8(index) + firstAddressIndex}, nil //nolint:gosec
}

func (c *BridgingAddressesCoordinatorImpl) GetAddressesAndAmountsForBatch(
	chainID uint8,
	cardanoCliBinary string,
	isRedistribution bool,
	protocolParams []byte,
	txOutputs cardanotx.TxOutputs,
) ([]common.AddressAndAmount, bool, error) {
	// Go through all addresses, sort them by the total amount of tokens (descending),
	// and choose the one with the biggest amount
	var (
		err     error
		amounts []common.AddressAndAmount
	)

	if len(txOutputs.Outputs) == 0 && !isRedistribution {
		return nil, isRedistribution, nil
	}

	requiredTokenAmounts := cardanowallet.GetOutputsSum(txOutputs.Outputs)
	requiredCurrencyAmount := requiredTokenAmounts[cardanowallet.AdaTokenName]

	c.logger.Debug("GetAddressesAndAmountsForBatch", "chain", common.ToStrChainID(chainID),
		"requiredTokenAmounts", requiredTokenAmounts)

	totalTokenAmounts, err := c.getTokensAmountByAddr(chainID)
	if err != nil {
		return nil, isRedistribution, err
	}

	changeMinUtxo, err := cardanotx.CalculateMinUtxoCurrencyAmount(
		cardanoCliBinary, protocolParams, totalTokenAmounts.addrAmounts[0].address,
		totalTokenAmounts.potentialInputs, txOutputs.Outputs)
	if err != nil {
		return nil, isRedistribution, err
	}

	changeMinUtxo = max(changeMinUtxo, common.MinUtxoAmountDefault)

	// Validate whether enough token funds exist
	if err := validateTokenFunds(requiredTokenAmounts, totalTokenAmounts.sum, changeMinUtxo); err != nil {
		return nil, isRedistribution, err
	}

	if isRedistribution {
		amounts, err = c.redistributeTokens(requiredTokenAmounts, totalTokenAmounts.addrAmounts, changeMinUtxo)
		if err != nil {
			c.logger.Warn("skipping redistribution", "error", err)
		}
	}

	if err != nil || !isRedistribution {
		amounts, err = c.getAddressesAndAmountsToPayFrom(
			totalTokenAmounts.addrAmounts, requiredTokenAmounts, changeMinUtxo)
		isRedistribution = false
	}

	if err != nil {
		return amounts, isRedistribution, err
	}

	result, err := handleSpecialConsolidation(
		requiredTokenAmounts[cardanowallet.AdaTokenName],
		requiredCurrencyAmount,
		amounts,
	)
	if err != nil {
		return result, isRedistribution, err
	}

	for tokenName, remainingAmount := range requiredTokenAmounts {
		if remainingAmount > 0 {
			return nil, isRedistribution,
				fmt.Errorf("not enough %s native token funds, missing: %v", tokenName, requiredTokenAmounts)
		}
	}

	return amounts, isRedistribution, nil
}

func (c *BridgingAddressesCoordinatorImpl) getAddressesAndAmountsToPayFrom(
	addrAmounts []addrAmount,
	requiredTokenAmounts map[string]uint64,
	changeMinUtxo uint64,
) ([]common.AddressAndAmount, error) {
	amounts := make([]common.AddressAndAmount, 0)

	// Sort by total currency amount descending
	sort.SliceStable(addrAmounts, func(i, j int) bool {
		return addrAmounts[i].
			totalTokenAmounts[cardanowallet.AdaTokenName] > addrAmounts[j].totalTokenAmounts[cardanowallet.AdaTokenName]
	})

	c.logger.Debug("Available addresses to pay from", addrAmounts)

	for i := range addrAmounts {
		// Early exit if we've satisfied all required amounts
		if len(requiredTokenAmounts) == 0 {
			break
		}

		c.logger.Debug("Processing address", i, "requiredTokenAmounts", requiredTokenAmounts)

		addrAmount := &addrAmounts[i]

		if addrAmount.totalTokenAmounts[cardanowallet.AdaTokenName] == 0 {
			continue
		}

		includeChange := common.MinUtxoAmountDefault

		// Process native tokens only from frist address if there are any
		if addrAmount.addressIndex == 0 && len(addrAmount.totalTokenAmounts) > 1 {
			includeChange = changeMinUtxo

			c.processNativeTokens(addrAmount, requiredTokenAmounts)
		}

		// Process remaining currency amount
		if requiredTokenAmounts[cardanowallet.AdaTokenName] > 0 {
			includeChange = c.processCurrencyAmount(
				addrAmount, changeMinUtxo, requiredTokenAmounts)
		} else {
			addrAmount.includeInTx[cardanowallet.AdaTokenName] = 0
		}

		amounts = append(amounts, common.AddressAndAmount{
			Address:           addrAmount.address,
			AddressIndex:      addrAmount.addressIndex,
			TokensAmounts:     addrAmount.includeInTx,
			IncludeChange:     includeChange,
			UtxoCount:         addrAmount.utxoCount,
			ShouldConsolidate: addrAmount.insufficientChange,
		})
	}

	return amounts, nil
}

func (c *BridgingAddressesCoordinatorImpl) redistributeCurrencyTokens(
	addrAmounts []addrAmount, totalCurrencyAmount uint64, requiredCurrencyAmount uint64,
	minUtxo uint64,
) ([]common.AddressAndAmount, error) {
	if len(addrAmounts) == 0 {
		return nil, fmt.Errorf("cannot redistribute currency tokens: no addresses provided")
	}

	remainingCurrencyAmount := totalCurrencyAmount - requiredCurrencyAmount
	currencyTokensPerAddress := remainingCurrencyAmount / uint64(len(addrAmounts))

	if currencyTokensPerAddress < minUtxo {
		return nil, fmt.Errorf("redistribution amount per address smaller than min utxo %d vs %d",
			currencyTokensPerAddress, minUtxo)
	}

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
	requiredTokenAmounts map[string]uint64, addrAmounts []addrAmount, minUtxo uint64,
) ([]common.AddressAndAmount, error) {
	totalCurrencyAmount := uint64(0)
	for _, addrAmount := range addrAmounts {
		totalCurrencyAmount += addrAmount.totalTokenAmounts[cardanowallet.AdaTokenName]
	}

	requiredCurrencyAmount := requiredTokenAmounts[cardanowallet.AdaTokenName]
	delete(requiredTokenAmounts, cardanowallet.AdaTokenName)

	addressAndAmounts, err := c.redistributeCurrencyTokens(
		addrAmounts, totalCurrencyAmount, requiredCurrencyAmount, minUtxo)
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

	addresses := c.bridgingAddressesManager.GetAllPaymentAddresses(chainID, common.AddressTypeNormal)

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

		availableTokens := addrAmount.totalTokenAmounts[tokenName]
		if availableTokens == 0 {
			continue
		}

		tokensTakenFromAddress := min(availableTokens, requiredAmount)
		c.logger.Debug("Taking native tokens from address", "address", addrAmount.address,
			"token", tokenName, "amount", tokensTakenFromAddress)

		requiredTokenAmounts[tokenName] -= tokensTakenFromAddress
		addrAmount.includeInTx[tokenName] = tokensTakenFromAddress
		addrAmount.totalTokenAmounts[tokenName] -= tokensTakenFromAddress

		if requiredTokenAmounts[tokenName] == 0 {
			delete(requiredTokenAmounts, tokenName)
		}

		if addrAmount.totalTokenAmounts[tokenName] == 0 {
			delete(addrAmount.totalTokenAmounts, tokenName)
		}
	}
}

// processCurrencyAmount handles the processing of currency amounts for a given address
// Updates the addrAmount.includeInTx map with the amount of currency to be included in the transaction
// Deducts the amount of currency from the requiredTokenAmounts map
// Returns the amount of currency to be included as change for this address during utxo selection
func (c *BridgingAddressesCoordinatorImpl) processCurrencyAmount(
	addrAmount *addrAmount,
	changeMinUtxo uint64,
	requiredTokenAmounts map[string]uint64,
) uint64 {
	minChange := c.spendCurrencyFromAddress(addrAmount, changeMinUtxo, requiredTokenAmounts)

	// Clean up if fully satisfied
	if requiredTokenAmounts[cardanowallet.AdaTokenName] == 0 {
		delete(requiredTokenAmounts, cardanowallet.AdaTokenName)
	}

	return minChange
}

func (c *BridgingAddressesCoordinatorImpl) spendCurrencyFromAddress(
	addrAmount *addrAmount,
	changeMinUtxo uint64,
	requiredTokenAmounts map[string]uint64,
) uint64 {
	requiredCurrencyAmount := requiredTokenAmounts[cardanowallet.AdaTokenName]
	availableCurrencyOnAddress := addrAmount.totalTokenAmounts[cardanowallet.AdaTokenName]

	// We handle this address in a different way compared to other addresses
	if addrAmount.addressIndex == 0 && len(addrAmount.totalTokenAmounts) > 1 {
		maxSpendable := availableCurrencyOnAddress - changeMinUtxo
		amountToSpend := min(maxSpendable, requiredCurrencyAmount)

		addrAmount.includeInTx[cardanowallet.AdaTokenName] = amountToSpend
		requiredTokenAmounts[cardanowallet.AdaTokenName] -= amountToSpend

		return changeMinUtxo
	}

	const defaultMinChange = common.MinUtxoAmountDefault

	if availableCurrencyOnAddress <= requiredCurrencyAmount {
		// Not enough currency on this address or exact amount
		// Take all currency from this address
		addrAmount.includeInTx[cardanowallet.AdaTokenName] = availableCurrencyOnAddress
		requiredTokenAmounts[cardanowallet.AdaTokenName] -= availableCurrencyOnAddress

		return 0
	}

	addressChange := availableCurrencyOnAddress - requiredCurrencyAmount
	c.logger.Debug("Address change", addrAmount.address, addressChange)

	if addressChange >= defaultMinChange {
		// Sufficient change amount
		addrAmount.includeInTx[cardanowallet.AdaTokenName] = requiredCurrencyAmount
		requiredTokenAmounts[cardanowallet.AdaTokenName] = 0

		return defaultMinChange
	}

	amountToSpend := availableCurrencyOnAddress - defaultMinChange
	addrAmount.includeInTx[cardanowallet.AdaTokenName] = amountToSpend
	requiredTokenAmounts[cardanowallet.AdaTokenName] -= amountToSpend
	addrAmount.insufficientChange = true

	return defaultMinChange
}

func validateTokenFunds(
	requiredTokenAmounts map[string]uint64,
	totalTokenAmounts map[string]uint64,
	changeMinUtxo uint64,
) error {
	for token, requiredAmount := range requiredTokenAmounts {
		available := totalTokenAmounts[token]

		// Adjust available ADA if required != total
		if token == cardanowallet.AdaTokenName && requiredAmount != available {
			available -= changeMinUtxo
		}

		if requiredAmount > available {
			return fmt.Errorf(
				"not enough %s token funds for batch: available = %d, required = %d",
				token, totalTokenAmounts[token], requiredAmount,
			)
		}
	}

	return nil
}

func handleSpecialConsolidation(
	requiredCurrencyAmounts uint64,
	initalRequiredCurrencyAmount uint64,
	amounts []common.AddressAndAmount,
) ([]common.AddressAndAmount, error) {
	if requiredCurrencyAmounts == 0 {
		return nil, nil
	}

	var (
		containsZeroAddr            bool
		insufficientChangeAddresses []common.AddressAndAmount
	)

	for _, addr := range amounts {
		if addr.AddressIndex == 0 {
			containsZeroAddr = true

			continue
		}

		if addr.ShouldConsolidate {
			insufficientChangeAddresses = append(insufficientChangeAddresses, addr)
		}
	}

	if len(insufficientChangeAddresses) == 0 {
		return nil, fmt.Errorf(
			"%w: required %d, but missing %d",
			cardanowallet.ErrUTXOsCouldNotSelect,
			initalRequiredCurrencyAmount,
			requiredCurrencyAmounts,
		)
	}

	if containsZeroAddr {
		return []common.AddressAndAmount{insufficientChangeAddresses[0]}, cardanotx.ErrInsufficientChange
	}

	return insufficientChangeAddresses, cardanotx.ErrInsufficientChange
}

func safeSubtract(a, b uint64) (uint64, bool) {
	if a >= b {
		return a - b, true
	}

	return 0, false
}
