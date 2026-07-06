package cliwalletcreate

import (
	"fmt"
	"strings"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/eth"
	ethtxhelper "github.com/Ethernal-Tech/apex-bridge/eth/txhelper"
	solanatx "github.com/Ethernal-Tech/apex-bridge/solana"
	"github.com/Ethernal-Tech/cardano-infrastructure/secrets"
	"github.com/spf13/cobra"
)

const (
	keyConfigFlag = "config"
	keyFlag       = "key"
	adminTypeFlag = "type"

	keyConfigFlagDesc = "path to secrets manager config file"
	keyFlagDesc       = "private key (hex ECDSA for admin/proxy; base58 or json keypair for solana types)"
	adminTypeFlagDesc = "type of wallet (admin, proxy, solana_admin, solana_program)"
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

	switch strings.ToLower(ip.adminType) {
	case "", "admin", "proxy", solanaAdminKeyType, solanaProgramKeyType:
	default:
		return fmt.Errorf("unsupported --%s: %s", adminTypeFlag, ip.adminType)
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

	switch strings.ToLower(ip.adminType) {
	case solanaAdminKeyType:
		return ip.storeSolanaKey(secretsManager, solanatx.GetKeyNameForSolanaAdmin(), solanaAdminKeyType)
	case solanaProgramKeyType:
		return ip.storeSolanaKey(secretsManager, solanatx.GetKeyNameForSolanaProgram(), solanaProgramKeyType)
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

func (ip *walletCreateBladeParams) storeSolanaKey(
	secretsManager secrets.SecretsManager,
	keyName string,
	walletType string,
) (common.ICommandResult, error) {
	privateKey, err := solanatx.PrivateKeyFromWalletString(ip.key)
	if err != nil {
		return nil, fmt.Errorf("invalid solana private key: %w", err)
	}

	if err := solanatx.StoreSolanaPrivateKey(secretsManager, keyName, privateKey); err != nil {
		return nil, err
	}

	return &evmCmdResult{
		ChainID:    walletType,
		PrivateKey: privateKey.String(),
		PublicKey:  privateKey.PublicKey().String(),
		Address:    privateKey.PublicKey().String(),
	}, nil
}
