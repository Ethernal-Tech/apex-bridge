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
	txOutputs *[]cardanowallet.TxOutput,
) ([]common.AddressAndAmount, error) {
	// Go through all addresses, sort them by the total amount of tokens (descending),
	// and choose the one with the biggest amount
	// Future improvement:
	// - add the stake pool saturation awareness
	if len(*txOutputs) == 0 {
		return nil, nil
	}

	db := c.dbs[common.ToStrChainID(chainID)]
	addresses := c.bridgingAddressesManager.GetAllPaymentAddresses(chainID)

	// TODO: should we extract methods that do this from cco_utils
	containsTokens := false
	remainingTokenAmounts := make(map[string]uint64)

	for _, txOutput := range *txOutputs {
		remainingTokenAmounts[cardanowallet.AdaTokenName] += txOutput.Amount

		for _, token := range txOutput.Tokens {
			containsTokens = true
			remainingTokenAmounts[token.TokenName()] += token.Amount
		}
	}

	c.logger.Debug("remainingTokenAmounts", remainingTokenAmounts)

	// If we have native tokens we need to add up min utxo lovelace amount
	// needed to be sent together with native token
	minUtxo := uint64(0)

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

		minUtxo, err = txBuilder.SetProtocolParameters(protocolParams).CalculateMinUtxo(cardanowallet.TxOutput{
			Addr:   addresses[0],
			Amount: remainingTokenAmounts[cardanowallet.AdaTokenName],
			Tokens: tokens,
		})
		if err != nil {
			return nil, err
		}

		minUtxo += common.MinUtxoAmountDefault
	}

	requiredAdaAmount := remainingTokenAmounts[cardanowallet.AdaTokenName]

	type addrAmount struct {
		address           string
		totalTokenAmounts map[string]uint64
		includeInTx       map[string]uint64
		addressIndex      uint8
		holdNativeTokens  bool
		utxoCount         int
	}

	addrAmounts := make([]addrAmount, 0, len(addresses))

	// Calculate amount hold by each address
	for i, address := range addresses {
		utxos, err := db.GetAllTxOutputs(address, true)
		if err != nil {
			return nil, err
		}

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

	// Sort by totalAmount descending
	sort.Slice(addrAmounts, func(i, j int) bool {
		return addrAmounts[i].
			totalTokenAmounts[cardanowallet.AdaTokenName] > addrAmounts[j].totalTokenAmounts[cardanowallet.AdaTokenName]
	})

	amounts := make([]common.AddressAndAmount, 0)

	c.logger.Debug("remainingTokenAmounts1", remainingTokenAmounts)
	c.logger.Debug("addrAmounts", addrAmounts)

	for i, addrAmount := range addrAmounts {
		if len(remainingTokenAmounts) == 0 {
			break
		}

		includeChange := uint64(0)

		c.logger.Debug("addrAmount Address", addrAmount.address)

		nativeTokenChangeInUtxos := false
		// Update remaining native token amounts
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
				includeChange = minUtxo
			} else if addrAmount.totalTokenAmounts[tokenName] > 0 {
				// Take full amount, no chnage so we don't pay attention to lovelace
				remainingTokenAmounts[tokenName] -= addrAmount.totalTokenAmounts[tokenName]
				addrAmount.includeInTx[tokenName] = addrAmount.totalTokenAmounts[tokenName]

				if remainingTokenAmounts[tokenName] == 0 {
					delete(remainingTokenAmounts, tokenName)
				}
			}
		}

		// Update remaining lovelace amount
		requiredAdaAmount := remainingTokenAmounts[cardanowallet.AdaTokenName]
		if requiredAdaAmount > 0 {
			availableAdaOnAddress := addrAmount.totalTokenAmounts[cardanowallet.AdaTokenName]

			requiredForChange := common.MinUtxoAmountDefault

			if nativeTokenChangeInUtxos {
				requiredForChange = minUtxo
			}

			addressChange, ok := safeSubstract(availableAdaOnAddress, requiredAdaAmount)
			c.logger.Debug("addressChange", addressChange)
			c.logger.Debug("nativeTokenChangeInUtxos", nativeTokenChangeInUtxos)

			if ok {
				if addressChange == 0 && !nativeTokenChangeInUtxos {
					// Exact amount
					addrAmount.includeInTx[cardanowallet.AdaTokenName] = requiredAdaAmount
					includeChange = 0
					remainingTokenAmounts[cardanowallet.AdaTokenName] = 0
				} else if addressChange >= requiredForChange {
					// Sufficient change amount
					addrAmount.includeInTx[cardanowallet.AdaTokenName] = requiredAdaAmount
					includeChange = requiredForChange
					remainingTokenAmounts[cardanowallet.AdaTokenName] = 0
				} else {
					c.logger.Debug("i", i)
					includeChange = requiredForChange + addressChange
					if availableAdaOnAddress-includeChange > common.MinUtxoAmountDefault && i <= len(addrAmounts)-2 {
						if addrAmounts[i+1].totalTokenAmounts[cardanowallet.AdaTokenName]-
							(remainingTokenAmounts[cardanowallet.AdaTokenName]-availableAdaOnAddress+includeChange) >= common.MinUtxoAmountDefault {
							addrAmount.includeInTx[cardanowallet.AdaTokenName] = availableAdaOnAddress - includeChange
							remainingTokenAmounts[cardanowallet.AdaTokenName] -= availableAdaOnAddress - includeChange
						} else {
							c.logger.Debug("Handle carry over")
							addrAmount.includeInTx[cardanowallet.AdaTokenName] = availableAdaOnAddress
							*txOutputs = append(*txOutputs, cardanowallet.NewTxOutput(
								addrAmounts[i+1].address,
								addressChange+addrAmounts[i+1].totalTokenAmounts[cardanowallet.AdaTokenName]))
							remainingTokenAmounts[cardanowallet.AdaTokenName] = addrAmounts[i+1].totalTokenAmounts[cardanowallet.AdaTokenName]
							includeChange = 0
						}
					} else if i <= len(addrAmounts)-2 {
						c.logger.Debug("Handle carry over")
						// We have to handle carry over, the case where
						// we need to send the insufficient change from one address to another
						// To do this we need to update txOutput and include other address amount
						// into the tx
						// We are able to do this only when the address that has carry over
						// isn't the last address that is being checked

						addrAmount.includeInTx[cardanowallet.AdaTokenName] = availableAdaOnAddress
						*txOutputs = append(*txOutputs, cardanowallet.NewTxOutput(
							addrAmounts[i+1].address,
							addressChange+addrAmounts[i+1].totalTokenAmounts[cardanowallet.AdaTokenName]))
						remainingTokenAmounts[cardanowallet.AdaTokenName] = addrAmounts[i+1].totalTokenAmounts[cardanowallet.AdaTokenName]
						includeChange = 0
					}

				}
			} else {
				addrAmount.includeInTx[cardanowallet.AdaTokenName] = availableAdaOnAddress
				includeChange = 0
				remainingTokenAmounts[cardanowallet.AdaTokenName] -= availableAdaOnAddress
			}

			if remainingTokenAmounts[cardanowallet.AdaTokenName] == 0 {
				delete(remainingTokenAmounts, cardanowallet.AdaTokenName)
			}
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

		if len(remainingTokenAmounts) == 0 {
			break
		}
	}

	c.logger.Debug("amounts:")

	for _, am := range amounts {
		c.logger.Debug("address", am.Address)
		c.logger.Debug("address index", am.AddressIndex)
		c.logger.Debug("tokens amount", am.TokensAmounts)
		c.logger.Debug("include change", am.IncludeChnage)
		c.logger.Debug("-------------------------------------")
	}

	if remainingTokenAmounts[cardanowallet.AdaTokenName] > 0 {
		return nil, fmt.Errorf("%w: %d vs %d",
			cardanowallet.ErrUTXOsCouldNotSelect, requiredAdaAmount, remainingTokenAmounts[cardanowallet.AdaTokenName])
	}

	for tokenName, remainingAmount := range remainingTokenAmounts {
		if remainingAmount > 0 {
			return nil, fmt.Errorf("not enough %s native token funds, missing: %v", tokenName, remainingTokenAmounts)
		}
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

	return common.AddressAndAmount{Address: addresses[index], AddressIndex: uint8(index)}, nil //nolint:gosec
}

func (c *BridgingAddressesCoordinatorImpl) GetAllAddresses(chainID uint8) []string {
	// Return all addresses
	return c.bridgingAddressesManager.GetAllPaymentAddresses(chainID)
}
