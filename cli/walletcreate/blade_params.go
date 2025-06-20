package cliwalletcreate

import (
	"fmt"
	"strings"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/eth"
	ethtxhelper "github.com/Ethernal-Tech/apex-bridge/eth/txhelper"
	"github.com/spf13/cobra"
)

const (
	keyConfigFlag = "config"
	keyFlag       = "key"
	adminTypeFlag = "type"

	keyConfigFlagDesc = "path to secrets manager config file"
	keyFlagDesc       = "hexadecimal representation of ECDSA key"
	adminTypeFlagDesc = "type of wallet (admin or proxy)"
)

type walletCreateBladeParams struct {
	keyConfig string
	key       string
	adminType string
}

func (ip *walletCreateBladeParams) validateFlags() error {
	if ip.keyConfig == "" {
		return fmt.Errorf("--%s not specified", keyConfigFlag)
	}

	if ip.key == "" {
		return fmt.Errorf("--%s not specified", keyFlag)
	}

	return nil
}

func (ip *walletCreateBladeParams) setFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(
		&ip.keyConfig,
		keyConfigFlag,
		"",
		keyConfigFlagDesc,
	)

	cmd.Flags().StringVar(
		&ip.key,
		keyFlag,
		"",
		keyFlagDesc,
	)

	cmd.Flags().StringVar(
		&ip.adminType,
		adminTypeFlag,
		"",
		adminTypeFlagDesc,
	)
}

func (ip *walletCreateBladeParams) Execute(outputter common.OutputFormatter) (common.ICommandResult, error) {
	secretsManager, err := common.GetSecretsManager("", ip.keyConfig, false)
	if err != nil {
		return nil, err
	}

	evmWallet, err := ethtxhelper.NewEthTxWallet(ip.key)
	if err != nil {
		return nil, err
	}

	keyName := eth.GetKeyNameForBladeAdmin(strings.ToLower(ip.adminType) == "proxy")

	if secretsManager.HasSecret(keyName) {
		if err := secretsManager.RemoveSecret(keyName); err != nil {
			return nil, err
		}
	}

	if err := evmWallet.Save(secretsManager, keyName); err != nil {
		return nil, err
	}

	pk, pub, addr := evmWallet.GetHexData()

	return &evmCmdResult{
		PrivateKey: pk,
		PublicKey:  pub,
		Address:    addr,
	}, nil
}
