package clicreateaddress

import (
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/cardano-infrastructure/wallet"
	"github.com/spf13/cobra"
)

const (
	keyFlag       = "key"
	networkIDFlag = "network-id"

	keyFlagDesc       = "cardano verification key for validator"
	networkIDFlagDesc = "network ID"
)

type createAddressParams struct {
	keys      []string
	networkID uint
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
		&ip.networkID,
		networkIDFlag,
		0,
		networkIDFlagDesc,
	)
}

func (ip *createAddressParams) Execute() (common.ICommandResult, error) {
	atLeast := common.GetRequiredSignaturesForConsensus(uint64(len(ip.keys)))

	script, err := wallet.NewPolicyScript(ip.keys, int(atLeast))
	if err != nil {
		return nil, fmt.Errorf("failed to create policy script: %w", err)
	}

	policyId, err := script.GetPolicyID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate policy id: %w", err)
	}

	policyIdBytes, err := hex.DecodeString(policyId)
	if err != nil {
		return nil, fmt.Errorf("failed to decode policy id string: %w", err)
	}

	stakeCredential, err := wallet.NewStakeCredential(policyIdBytes, true)
	if err != nil {
		return nil, fmt.Errorf("failed to create stake credential: %w", err)
	}

	return &CmdResult{
		address: (wallet.EnterpriseAddress{
			Network: wallet.CardanoNetworkType(ip.networkID),
			Payment: stakeCredential,
		}).String(),
	}, nil
}
