package ethtxhelper

import (
	"crypto/ecdsa"
	"math/big"
	"strings"

	apexcommon "github.com/Ethernal-Tech/apex-bridge/common"
	secretsInfra "github.com/Ethernal-Tech/cardano-infrastructure/secrets"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
)

type IEthTxWallet interface {
	GetTransactOpts(chainID *big.Int) (*bind.TransactOpts, error)
	GetAddress() common.Address
}

type EthTxWallet struct {
	addr       common.Address
	privateKey *ecdsa.PrivateKey
}

var _ IEthTxWallet = (*EthTxWallet)(nil)

func NewEthTxWalletFromSecretManager(sm secretsInfra.SecretsManager) (*EthTxWallet, error) {
	privateKey, err := sm.GetSecret(secretsInfra.ValidatorKey)
	if err != nil {
		return nil, err
	}

	return NewEthTxWallet(string(privateKey))
}

func NewEthTxWalletFromSecretManagerConfig(config *secretsInfra.SecretsManagerConfig) (*EthTxWallet, error) {
	sm, err := apexcommon.GetSecretsManagerFromConfig(config)
	if err != nil {
		return nil, err
	}

	return NewEthTxWalletFromSecretManager(sm)
}

func NewEthTxWallet(pk string) (*EthTxWallet, error) {
	bytes, err := apexcommon.DecodeHex(strings.Trim(strings.Trim(pk, "\n"), " "))
	if err != nil {
		return nil, err
	}

	privateKey, err := crypto.ToECDSA(bytes)
	if err != nil {
		return nil, err
	}

	// Get the Ethereum address from the public key
	ethereumAddress := crypto.PubkeyToAddress(privateKey.PublicKey)

	return &EthTxWallet{
		privateKey: privateKey,
		addr:       ethereumAddress,
	}, nil
}

func (w EthTxWallet) GetTransactOpts(chainID *big.Int) (*bind.TransactOpts, error) {
	return bind.NewKeyedTransactorWithChainID(w.privateKey, chainID)
}

func (w EthTxWallet) GetAddress() common.Address {
	return w.addr
}

func (w EthTxWallet) SignTx(chainID *big.Int, tx *types.Transaction) (*types.Transaction, error) {
	return types.SignTx(tx, types.NewLondonSigner(chainID), w.privateKey)
}

func TxOpts2LegacyTx(to string, data []byte, txOpts *bind.TransactOpts) *types.Transaction {
	return types.NewTransaction(txOpts.Nonce.Uint64(), common.HexToAddress(to), txOpts.Value, txOpts.GasLimit, txOpts.GasPrice, data)
}

func TxOpts2DynamicFeeTx(to string, chainID *big.Int, data []byte, txOpts *bind.TransactOpts) *types.Transaction {
	toAddr := common.HexToAddress(to)

	return types.NewTx(&types.DynamicFeeTx{
		ChainID:   chainID,
		Nonce:     txOpts.Nonce.Uint64(),
		To:        &toAddr,
		Gas:       txOpts.GasLimit,
		Value:     txOpts.Value,
		Data:      data,
		GasFeeCap: txOpts.GasFeeCap,
		GasTipCap: txOpts.GasTipCap,
	})
}
