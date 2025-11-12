package cliwalletcreate

import (
	"bytes"
	"encoding/hex"
	"fmt"

	"github.com/Ethernal-Tech/apex-bridge/common"
	cardanowallet "github.com/Ethernal-Tech/cardano-infrastructure/wallet"
)

type cardanoCmdResult struct {
	Multisig        *cardanowallet.Wallet `json:"multisig"`
	Fee             *cardanowallet.Wallet `json:"fee"`
	KeyHash         string                `json:"keyHash"`
	StakeKeyHash    string                `json:"stakeKeyHash"`
	KeyHashFee      string                `json:"keyHashFee"`
	StakeKeyHashFee string                `json:"stakeKeyHashFee"`
	ChainID         string                `json:"chainID"`
	showPrivateKey  bool
}

func (res cardanoCmdResult) GetOutput() string {
	var buffer bytes.Buffer

	buffer.WriteString("\n[SECRETS ")
	buffer.WriteString(res.ChainID)
	buffer.WriteString("]\n")
	buffer.WriteString("[Multisig]\n")
	buffer.WriteString(common.FormatKV(res.getColumnData(res.Multisig, res.KeyHash, res.StakeKeyHash)))
	buffer.WriteString("\n")
	buffer.WriteString("[MultisigFee]\n")
	buffer.WriteString(common.FormatKV(res.getColumnData(res.Fee, res.KeyHashFee, res.StakeKeyHashFee)))
	buffer.WriteString("\n")

	return buffer.String()
}

func (res cardanoCmdResult) getColumnData(
	wallet *cardanowallet.Wallet, keyHash, stakeKeyHash string,
) (vals []string) {
	if res.showPrivateKey {
		vals = append(vals, fmt.Sprintf("Signing Key|%s", hex.EncodeToString(wallet.SigningKey)))
	}

	vals = append(vals,
		fmt.Sprintf("Verifying Key|%s", hex.EncodeToString(wallet.VerificationKey)),
		fmt.Sprintf("Key Hash|%s", keyHash))

	if len(wallet.StakeVerificationKey) > 0 {
		if res.showPrivateKey {
			vals = append(vals, fmt.Sprintf("Stake Signing Key|%s", hex.EncodeToString(wallet.StakeSigningKey)))
		}

		vals = append(vals,
			fmt.Sprintf("Stake Verifying Key|%s", hex.EncodeToString(wallet.StakeVerificationKey)),
			fmt.Sprintf("Stake Key Hash|%s", stakeKeyHash))
	}

	return vals
}

type evmCmdResult struct {
	ChainID        string `json:"chainID"`
	PrivateKey     string `json:"privateKey"`
	PublicKey      string `json:"publicKey"`
	Address        string `json:"address"`
	showPrivateKey bool
}

func (r evmCmdResult) GetOutput() string {
	var (
		buffer bytes.Buffer
		vals   []string
	)

	if r.showPrivateKey {
		vals = append(vals, fmt.Sprintf("Private Key|%s", r.PrivateKey))
	}

	vals = append(vals, fmt.Sprintf("Public Key|%s", r.PublicKey))

	if r.Address != "" {
		vals = append(vals, fmt.Sprintf("Address|%s", r.Address))
	}

	buffer.WriteString("\n[SECRETS ")
	buffer.WriteString(r.ChainID)
	buffer.WriteString("]\n")
	buffer.WriteString(common.FormatKV(vals))
	buffer.WriteString("\n")

	return buffer.String()
}

type cardanoRelayerCmdResult struct {
	ChainID        string                `json:"chainID"`
	Wallet         *cardanowallet.Wallet `json:"wallet"`
	Address        string                `json:"address"`
	showPrivateKey bool
}

func (r cardanoRelayerCmdResult) GetOutput() string {
	var (
		buffer bytes.Buffer
		vals   []string
	)

	if r.showPrivateKey {
		vals = append(vals, fmt.Sprintf("Signing Key|%s", hex.EncodeToString(r.Wallet.SigningKey)))
	}

	vals = append(vals,
		fmt.Sprintf("Verifying Key|%s", hex.EncodeToString(r.Wallet.VerificationKey)),
		fmt.Sprintf("Address|%s", r.Address),
	)

	buffer.WriteString("\n[RELAYER SECRETS ")
	buffer.WriteString(r.ChainID)
	buffer.WriteString("]\n")
	buffer.WriteString(common.FormatKV(vals))
	buffer.WriteString("\n")

	return buffer.String()
}
