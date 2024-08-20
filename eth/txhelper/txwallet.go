package ethtxhelper

import (
	"crypto/ecdsa"
	"encoding/hex"
	"math/big"
	"strings"

	apexcommon "github.com/Ethernal-Tech/apex-bridge/common"
	secretsInfra "github.com/Ethernal-Tech/cardano-infrastructure/secrets"
	"github.com/Ethernal-Tech/cardano-infrastructure/wallet"
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

func NewEthTxWallet(pk string) (*EthTxWallet, error) {
	var (
		bytes []byte
		err   error
	)

	if len(pk) == wallet.KeySize {
		bytes = ([]byte)(pk)
	} else {
		bytes, err = apexcommon.DecodeHex(strings.TrimSpace(pk))
		if err != nil {
			return nil, err
		}
	}

	privateKey, err := crypto.ToECDSA(bytes)
	if err != nil {
		return nil, err
	}

	return &EthTxWallet{
		privateKey: privateKey,
		addr:       crypto.PubkeyToAddress(privateKey.PublicKey), // Get the Ethereum address from the public key
	}, nil
}

func GenerateNewEthTxWallet() (*EthTxWallet, error) {
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		return nil, err
	}

	return &EthTxWallet{
		privateKey: privateKey,
		addr:       crypto.PubkeyToAddress(privateKey.PublicKey), // Get the Ethereum address from the public key
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

func (w EthTxWallet) Save(secretsManager secretsInfra.SecretsManager, key string) error {
	privateKeyBytes := crypto.FromECDSA(w.privateKey)

	return secretsManager.SetSecret(key, []byte(hex.EncodeToString(privateKeyBytes)))
}

func (w EthTxWallet) GetHexData() (string, string, string) {
	privateKeyBytes := crypto.FromECDSA(w.privateKey)
	publicKeyBytes := crypto.FromECDSAPub(&w.privateKey.PublicKey)

	return hex.EncodeToString(privateKeyBytes), hex.EncodeToString(publicKeyBytes), w.addr.Hex()
}

func TxOpts2LegacyTx(to string, data []byte, txOpts *bind.TransactOpts) *types.Transaction {
	return types.NewTransaction(
		txOpts.Nonce.Uint64(), common.HexToAddress(to), txOpts.Value,
		txOpts.GasLimit, txOpts.GasPrice, data)
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
