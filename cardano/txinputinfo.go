package cardanotx

import (
	cardanowallet "github.com/Ethernal-Tech/cardano-infrastructure/wallet"
)

type TxInputInfos struct {
	MultiSig    *TxInputInfo
	MultiSigFee *TxInputInfo
}

type TxOutputs struct {
	Outputs []cardanowallet.TxOutput
	Sum     uint64
}

type TxInputInfo struct {
	cardanowallet.TxInputs
	PolicyScript *cardanowallet.PolicyScript
	Address      string
}
