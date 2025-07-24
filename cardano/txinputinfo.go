package cardanotx

import (
	cardanowallet "github.com/Ethernal-Tech/cardano-infrastructure/wallet"
)

type TxInputInfos struct {
	MultiSig    []*TxInputInfo
	MultiSigFee *TxInputInfo
}

type TxOutputs struct {
	Outputs []cardanowallet.TxOutput
	Sum     map[string]uint64
}

type TxInputInfo struct {
	cardanowallet.TxInputs
	PolicyScript *cardanowallet.PolicyScript
	Address      string
}

type CertificatesWithScript struct {
	PolicyScript *cardanowallet.PolicyScript
	Certificates []cardanowallet.ICertificate
}

type CertificatesData struct {
	Certificates    []*CertificatesWithScript
	RegistrationFee uint64
}
