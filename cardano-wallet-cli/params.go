package cardanowalletcli

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"path"
	"path/filepath"

	cardanotx "github.com/Ethernal-Tech/apex-bridge/cardano"
	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/contractbinding"
	ethtxhelper "github.com/Ethernal-Tech/apex-bridge/eth/txhelper"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/spf13/cobra"
)

const (
	directoryFlag           = "directory"
	nodeUrlFlag             = "node-url" // https://polygon-mumbai-pokt.nodies.app
	smartContractAddrFlag   = "addr"     // 0x69b6eEAff0A5c5F80a242104B79F4aC5c40E5130
	validatorPrivateKeyFlag = "pk"
	networkIDFlag           = "network"
	generateStakeKeyFlag    = "stake"
	forceRegenerateFlag     = "force"
	showPrivateKeyFlag      = "show-pk"

	directoryFlagDesc           = "the directory where the wallet keys will be stored"
	nodeUrlFlagDesc             = "cardano node url where tx will be submitted"
	smartContractAddrFlagDesc   = "smart contract address"
	validatorPrivateKeyFlagDesc = "validator's private key for the bridge"
	networkIDFlagDesc           = "network ID for which key is generated (prime, vector, etc)"
	generateStakeKeyFlagDesc    = "generate stake keys"
	forceRegenerateFlagDesc     = "force regenerating keys even if they exist in specified directory"
	showPrivateKeyFlagDesc      = "show private key in output" //nolint:gosec
)

type initParams struct {
	directory           string
	nodeUrl             string
	smartContractAddr   string
	validatorPrivateKey string
	networkID           string
	generateStakeKey    bool
	forceRegenerate     bool
	showPrivateKey      bool

	wallet      *ethtxhelper.EthTxWallet
	ethTxHelper ethtxhelper.IEthTxHelper
}

func (ip *initParams) validateFlags() error {
	if !common.IsValidURL(ip.nodeUrl) {
		return fmt.Errorf("invalid node url: %s", ip.nodeUrl)
	}

	if filepath.Clean(ip.directory) == "" {
		return fmt.Errorf("invalid directory: %s", ip.directory)
	}

	addrDecoded, err := common.DecodeHex(ip.smartContractAddr)
	if err != nil || len(addrDecoded) == 0 || len(addrDecoded) > 20 {
		return fmt.Errorf("invalid smart contract address: %s", ip.smartContractAddr)
	}

	if ip.networkID == "" {
		return errors.New("networkID is empty, --network flag not specified")
	}

	ethTxHelper, err := ethtxhelper.NewEThTxHelper(ethtxhelper.WithNodeUrl(ip.nodeUrl))
	if err != nil {
		return fmt.Errorf("can not connect to node: %v", err)
	}

	wallet, err := ethtxhelper.NewEthTxWallet(ip.validatorPrivateKey)
	if err != nil {
		return fmt.Errorf("invalid validator private key: %v", err)
	}

	ip.wallet = wallet
	ip.ethTxHelper = ethTxHelper

	return nil
}

func (ip *initParams) setFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(
		&ip.directory,
		directoryFlag,
		"",
		directoryFlagDesc,
	)

	cmd.Flags().StringVar(
		&ip.nodeUrl,
		nodeUrlFlag,
		"",
		nodeUrlFlagDesc,
	)

	cmd.Flags().StringVar(
		&ip.smartContractAddr,
		smartContractAddrFlag,
		"",
		smartContractAddrFlagDesc,
	)

	cmd.Flags().StringVar(
		&ip.validatorPrivateKey,
		validatorPrivateKeyFlag,
		"",
		validatorPrivateKeyFlagDesc,
	)

	cmd.Flags().StringVar(
		&ip.networkID,
		networkIDFlag,
		"",
		networkIDFlagDesc,
	)

	cmd.Flags().BoolVar(
		&ip.generateStakeKey,
		generateStakeKeyFlag,
		false,
		generateStakeKeyFlagDesc,
	)

	cmd.Flags().BoolVar(
		&ip.forceRegenerate,
		forceRegenerateFlag,
		false,
		forceRegenerateFlagDesc,
	)

	cmd.Flags().BoolVar(
		&ip.showPrivateKey,
		showPrivateKeyFlag,
		false,
		showPrivateKeyFlagDesc,
	)
}

func (ip *initParams) Execute() (common.ICommandResult, error) {
	wallet, err := cardanotx.GenerateWallet(path.Join(ip.directory, ip.networkID), ip.generateStakeKey, ip.forceRegenerate)
	if err != nil {
		return nil, err
	}

	contract, err := contractbinding.NewTestContractKeyHash(common.HexToAddress(ip.smartContractAddr), ip.ethTxHelper.GetClient())
	if err != nil {
		return nil, err
	}

	// first call is just for creating tx
	tx, err := ip.ethTxHelper.SendTx(context.Background(),
		ip.wallet, bind.TransactOpts{}, true, func(txOpts *bind.TransactOpts) (*types.Transaction, error) {
			return contract.SetValidatorCardanoData(txOpts, ip.networkID, contractbinding.TestContractKeyHashValidatorCardanoData{
				KeyHash:         wallet.MultiSig.GetKeyHash(),
				VerifyingKey:    wallet.MultiSig.GetVerificationKey(),
				KeyHashFee:      wallet.MultiSigFee.GetKeyHash(),
				VerifyingKeyFee: wallet.MultiSigFee.GetVerificationKey(),
			})
		})
	if err != nil {
		return nil, err
	}

	receipt, err := ip.ethTxHelper.WaitForReceipt(context.Background(), tx.Hash().String(), true)
	if err != nil {
		return nil, err
	}

	return &CmdResult{
		SigningKey:      hex.EncodeToString(wallet.MultiSig.GetSigningKey()),
		VerifyingKey:    hex.EncodeToString(wallet.MultiSig.GetVerificationKey()),
		KeyHash:         wallet.MultiSig.GetKeyHash(),
		SigningKeyFee:   hex.EncodeToString(wallet.MultiSigFee.GetSigningKey()),
		VerifyingKeyFee: hex.EncodeToString(wallet.MultiSigFee.GetVerificationKey()),
		KeyHashFee:      wallet.MultiSigFee.GetKeyHash(),
		showPrivateKey:  ip.showPrivateKey,
		networkID:       ip.networkID,
		blockHash:       receipt.BlockHash.String(),
	}, nil
}
