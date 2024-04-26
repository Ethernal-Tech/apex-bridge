package clicreateaddress

import (
	"errors"
	"fmt"
	"strings"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/cardano-infrastructure/wallet"
	"github.com/spf13/cobra"
)

const (
	keyFlag               = "key"
	testnetMagicFlag      = "testnet"
	addrPrefixReplaceFlag = "prefix"

	keyFlagDesc               = "cardano verification key for validator"
	testnetMagicFlagDesc      = "testnet magic number. leave 0 for mainnet"
	addrPrefixReplaceFlagDesc = "prefix replacement for a address"
)

type createAddressParams struct {
	keys         []string
	testnetMagic uint
	addrPrefix   string
}

func (ip *createAddressParams) validateFlags() error {
	if len(ip.keys) == 0 {
		return errors.New("keys not specified")
	}

	existing := make(map[string]bool, len(ip.keys))

	for i, vk := range ip.keys {
		if vk == "" {
			return errors.New("empty key")
		}

		vkBytes, err := common.DecodeHex(vk)
		if err != nil {
			return fmt.Errorf("invalid key: %s", vk)
		}

		keyHash, err := wallet.GetKeyHash(vkBytes)
		if err != nil {
			return fmt.Errorf("invalid key: %s", vk)
		}

		if existing[keyHash] {
			return fmt.Errorf("duplicate key: %s", vk)
		}

		existing[keyHash] = true
		ip.keys[i] = keyHash // overwrite key with hash key
	}

	return nil
}

func (ip *createAddressParams) setFlags(cmd *cobra.Command) {
	cmd.Flags().StringArrayVar(
		&ip.keys,
		keyFlag,
		nil,
		keyFlagDesc,
	)

	cmd.Flags().UintVar(
		&ip.testnetMagic,
		testnetMagicFlag,
		0,
		testnetMagicFlagDesc,
	)

	cmd.Flags().StringVar(
		&ip.addrPrefix,
		addrPrefixReplaceFlag,
		"",
		addrPrefixReplaceFlagDesc,
	)
}

func (ip *createAddressParams) Execute() (common.ICommandResult, error) {
	atLeast := common.GetRequiredSignaturesForConsensus(uint64(len(ip.keys)))

	script, err := wallet.NewPolicyScript(ip.keys, int(atLeast))
	if err != nil {
		return nil, err
	}

	address, err := script.CreateMultiSigAddress(ip.testnetMagic)
	if err != nil {
		return nil, err
	}

	if ip.addrPrefix != "" {
		address = strings.Replace(
			strings.Replace(address, "addr_test", ip.addrPrefix, 1),
			"addr", ip.addrPrefix, 1,
		)
	}

	return &CmdResult{
		address: address,
	}, nil
}
