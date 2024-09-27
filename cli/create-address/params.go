package clicreateaddress

import (
	"context"
	"errors"
	"fmt"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/eth"
	"github.com/Ethernal-Tech/cardano-infrastructure/wallet"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/hashicorp/go-hclog"
	"github.com/spf13/cobra"
)

const (
	keyFlag           = "key"
	networkIDFlag     = "network-id"
	chainIDFlag       = "chain"
	bridgeNodeURLFlag = "bridge-url"
	bridgeSCAddrFlag  = "bridge-addr"

	keyFlagDesc           = "cardano verification key for validator"
	networkIDFlagDesc     = "network ID"
	bridgeNodeURLFlagDesc = "bridge node url"
	bridgeSCAddrFlagDesc  = "bridge smart contract address"
	chainIDFlagDesc       = "cardano chain ID (prime, vector, etc)"
)

type createAddressParams struct {
	keys      []string
	networkID uint

	bridgeNodeURL string
	bridgeSCAddr  string
	chainID       string
}

func (ip *createAddressParams) validateFlags() error {
	if ip.bridgeNodeURL != "" {
		if !common.IsValidHTTPURL(ip.bridgeNodeURL) {
			return fmt.Errorf("invalid --%s flag", bridgeNodeURLFlag)
		}

		if !ethcommon.IsHexAddress(ip.bridgeSCAddr) {
			return fmt.Errorf("invalid --%s flag", bridgeSCAddrFlag)
		}

		if !common.IsExistingChainID(ip.chainID) {
			return fmt.Errorf("unexisting chain: %s", ip.chainID)
		}

		return nil
	}

	if len(ip.keys) == 0 {
		return errors.New("keys not specified")
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

	cmd.Flags().StringVar(
		&ip.bridgeNodeURL,
		bridgeNodeURLFlag,
		"",
		bridgeNodeURLFlagDesc,
	)

	cmd.Flags().StringVar(
		&ip.bridgeSCAddr,
		bridgeSCAddrFlag,
		"",
		bridgeSCAddrFlagDesc,
	)

	cmd.Flags().StringVar(
		&ip.chainID,
		chainIDFlag,
		common.ChainIDStrPrime,
		chainIDFlagDesc,
	)

	cmd.MarkFlagsMutuallyExclusive(bridgeNodeURLFlag, keyFlag)
	cmd.MarkFlagsMutuallyExclusive(bridgeSCAddrFlag, keyFlag)
	cmd.MarkFlagsMutuallyExclusive(chainIDFlag, keyFlag)
}

func (ip *createAddressParams) Execute() (common.ICommandResult, error) {
	if len(ip.keys) > 0 {
		keys, err := getKeyHashesFromInput(ip.keys)
		if err != nil {
			return nil, err
		}

		addr, err := getAddress(ip.networkID, keys)
		if err != nil {
			return nil, err
		}

		return &CmdResult{
			address: addr,
		}, nil
	}

	multisigHashes, feeHashes, err := getKeyHashesFromBridge(
		context.Background(), ip.bridgeNodeURL, ip.bridgeSCAddr, ip.chainID)
	if err != nil {
		return nil, err
	}

	multisigAddr, err := getAddress(ip.networkID, multisigHashes)
	if err != nil {
		return nil, err
	}

	feeAddr, err := getAddress(ip.networkID, feeHashes)
	if err != nil {
		return nil, err
	}

	return &CmdResult{
		multisigAddress: multisigAddr,
		address:         feeAddr,
	}, nil
}

func getAddress(networkIDInt uint, keys []string) (string, error) {
	networkID := wallet.CardanoNetworkType(networkIDInt)
	atLeast := common.GetRequiredSignaturesForConsensus(uint64(len(keys)))
	script := wallet.NewPolicyScript(keys, int(atLeast))
	cliUtils := wallet.NewCliUtils(wallet.ResolveCardanoCliBinary(networkID))

	policyID, err := cliUtils.GetPolicyID(script)
	if err != nil {
		return "", fmt.Errorf("failed to generate policy id: %w", err)
	}

	addr, err := wallet.NewPolicyScriptAddress(networkID, policyID)
	if err != nil {
		return "", fmt.Errorf("failed to create address: %w", err)
	}

	return addr.String(), nil
}

func getKeyHashesFromBridge(
	ctx context.Context, nodeURL, addr, chainID string,
) ([]string, []string, error) {
	bridgeSC := eth.NewBridgeSmartContract(nodeURL, addr, false, hclog.NewNullLogger())

	validatorsData, err := bridgeSC.GetValidatorsChainData(ctx, chainID)
	if err != nil {
		return nil, nil, err
	}

	multisigHashes := make([]string, len(validatorsData))
	feeHashes := make([]string, len(validatorsData))

	for i, x := range validatorsData {
		keyHashMultisig, err := wallet.GetKeyHash(x.Key[0].Bytes())
		if err != nil {
			return nil, nil, err
		}

		keyHashFee, err := wallet.GetKeyHash(x.Key[1].Bytes())
		if err != nil {
			return nil, nil, err
		}

		multisigHashes[i] = keyHashMultisig
		feeHashes[i] = keyHashFee
	}

	return multisigHashes, feeHashes, nil
}

func getKeyHashesFromInput(keys []string) ([]string, error) {
	existing := make(map[string]bool, len(keys))
	result := make([]string, len(keys))

	for i, vk := range keys {
		if vk == "" {
			return nil, errors.New("empty key")
		}

		vkBytes, err := common.DecodeHex(vk)
		if err != nil {
			return nil, fmt.Errorf("invalid key: %s", vk)
		}

		keyHash, err := wallet.GetKeyHash(vkBytes)
		if err != nil {
			return nil, fmt.Errorf("invalid key: %s", vk)
		}

		if existing[keyHash] {
			return nil, fmt.Errorf("duplicate key: %s", vk)
		}

		existing[keyHash] = true
		result[i] = keyHash // overwrite key with hash key
	}

	return result, nil
}
