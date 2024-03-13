package cardanotx

import (
	"path"

	cardanowallet "github.com/Ethernal-Tech/cardano-infrastructure/wallet"
)

const (
	multisigPath    = "multisig"
	multisigFeePath = "multisigfee"
)

type CardanoWallet struct {
	MultiSig    cardanowallet.IWallet
	MultiSigFee cardanowallet.IWallet
}

func GenerateWallet(directory string, isStake bool, forceRegenerate bool) (*CardanoWallet, error) {
	var walletMngr cardanowallet.IWalletManager
	if isStake {
		walletMngr = cardanowallet.NewStakeWalletManager()
	} else {
		walletMngr = cardanowallet.NewWalletManager()
	}

	walletMultiSig, err := walletMngr.Create(path.Join(directory, multisigPath), forceRegenerate)
	if err != nil {
		return nil, err
	}

	walletMultiSigFee, err := walletMngr.Create(path.Join(directory, multisigFeePath), forceRegenerate)
	if err != nil {
		return nil, err
	}

	return &CardanoWallet{
		MultiSig:    walletMultiSig,
		MultiSigFee: walletMultiSigFee,
	}, nil
}

func LoadWallet(directory string, isStake bool) (*CardanoWallet, error) {
	var walletMngr cardanowallet.IWalletManager
	if isStake {
		walletMngr = cardanowallet.NewStakeWalletManager()
	} else {
		walletMngr = cardanowallet.NewWalletManager()
	}

	walletMultiSig, err := walletMngr.Load(path.Join(directory, multisigPath))
	if err != nil {
		return nil, err
	}

	walletMultiSigFee, err := walletMngr.Load(path.Join(directory, multisigFeePath))
	if err != nil {
		return nil, err
	}

	return &CardanoWallet{
		MultiSig:    walletMultiSig,
		MultiSigFee: walletMultiSigFee,
	}, nil
}
