package cardanotx

import (
	"errors"

	"github.com/Ethernal-Tech/apex-bridge/common"
	cardanowallet "github.com/Ethernal-Tech/cardano-infrastructure/wallet"
)

var (
	ErrInsufficientChange = errors.New("insufficient change, special consolidation required")
)

type TxInputInfos struct {
	MultiSig    []*TxInputInfo
	MultiSigFee *TxInputInfo
}

type TxOutputs = common.TxOutputs

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
	Certificates      []*CertificatesWithScript
	RegistrationFee   uint64
	DeregistrationFee uint64
}
