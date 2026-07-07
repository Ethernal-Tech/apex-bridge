package clisendtx

import (
	"context"
	"fmt"
	"math/big"
	"os"
	"strings"
	"time"

	cardanotx "github.com/Ethernal-Tech/apex-bridge/cardano"
	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/contractbinding"
	ethtxhelper "github.com/Ethernal-Tech/apex-bridge/eth/txhelper"
	infracommon "github.com/Ethernal-Tech/cardano-infrastructure/common"
	"github.com/Ethernal-Tech/cardano-infrastructure/sendtx"
	cardanowallet "github.com/Ethernal-Tech/cardano-infrastructure/wallet"
	solsendtx "github.com/Ethernal-Tech/solana-infrastructure/sendtx"
	solanawallet "github.com/Ethernal-Tech/solana-infrastructure/wallet"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/spf13/cobra"
)

const (
	operationFeeFlag                  = "operation-fee"
	fullSrcTokenNameFlag              = "src-token-name"          //nolint:gosec
	fullDstTokenNameFlag              = "dst-token-name"          //nolint:gosec
	tokenIDSrcFlag                    = "src-token-id"            //nolint:gosec
	tokenContractAddrSrcFlag          = "src-token-contract-addr" //nolint:gosec
	tokenContractAddrDstFlag          = "dst-token-contract-addr" //nolint:gosec
	nativeTokenWalletContractAddrFlag = "native-token-wallet-contract-addr"
	solanaURLFlag                     = "solana-url"
	solanaProgramIDFlag               = "program-id"
	rpcURLDstFlag                     = "rpc-url-dst"

	operationFeeFlagDesc                  = "operation fee"
	fullSrcTokenNameFlagDesc              = "denom of the token to transfer from source chain"    //nolint:gosec
	fullDstTokenNameFlagDesc              = "denom of the token to transfer to destination chain" //nolint:gosec
	tokenIDSrcFlagDesc                    = "token id from source chain"
	tokenContractAddrSrcFlagDesc          = "contract address of the source token (EVM source chains only)"
	tokenContractAddrDstFlagDesc          = "contract address of the destination token (EVM destination chains only)"
	nativeTokenWalletContractAddrFlagDesc = "address of native token wallet contract"
	solanaURLFlagDesc                     = "solana rpc url"
	solanaProgramIDFlagDesc               = "source skyline solana program id"
	rpcURLDstFlagDesc                     = "destination evm chain rpc url (EVM source chains only)"

	apexTokenID = uint16(1)
	adaTokenID  = uint16(2)

	solanaWaitForSignatureTimeout = 3 * time.Minute
)

type sendSkylineTxParams struct {
	txType string // cardano, evm or solana

	privateKeyRaw      string
	stakePrivateKeyRaw string
	receivers          []string
	chainIDSrc         string
	chainIDDst         string
	feeString          string
	operationFeeString string
	tokenIDSrc         uint16
	tokenFullNameSrc   string
	tokenFullNameDst   string
	chainIDsConfig     string

	ogmiosURLSrc    string
	networkIDSrc    uint
	testnetMagicSrc uint
	multisigAddrSrc string
	treasuryAddrSrc string
	ogmiosURLDst    string

	// evm
	gatewayAddress                   string
	nativeTokenWalletContractAddress string
	rpcURL                           string
	rpcURLDst                        string
	tokenContractAddrSrc             string
	tokenContractAddrDst             string

	solanaWallet           *solanawallet.Wallet
	solanaProgramPublicKey solana.PublicKey
	solanaTreasuryAddress  solana.PublicKey
	solanaProgramID        string
	solanaURL              string

	feeAmount          *big.Int
	operationFeeAmount *big.Int
	receiversParsed    []*receiverAmount
	wallet             *cardanowallet.Wallet
	chainIDConverter   *common.ChainIDConverter

	cardanoCliBinaryName string
}

type skylineChainCommand struct {
	use       string
	short     string
	chainType string
	addFlags  func(cmd *cobra.Command, params *sendSkylineTxParams)
}

func getSendSkylineTxCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   sendSkylineTxCommandUse,
		Short: "sends the transaction in skyline mode",
	}

	chainCommands := []skylineChainCommand{
		{
			use:       common.ChainTypeCardanoStr,
			short:     "send skyline transaction from cardano source chain",
			chainType: common.ChainTypeCardanoStr,
			addFlags:  addCardanoSkylineFlags,
		},
		{
			use:       common.ChainTypeEVMStr,
			short:     "send skyline transaction from evm source chain",
			chainType: common.ChainTypeEVMStr,
			addFlags:  addEvmSkylineFlags,
		},
		{
			use:       common.ChainTypeSolanaStr,
			short:     "send skyline transaction from solana source chain",
			chainType: common.ChainTypeSolanaStr,
			addFlags:  addSolanaSkylineFlags,
		},
	}

	for _, chainCommand := range chainCommands {
		params := &sendSkylineTxParams{txType: chainCommand.chainType}
		subCmd := &cobra.Command{
			Use:   chainCommand.use,
			Short: chainCommand.short,
			PreRunE: func(_ *cobra.Command, _ []string) error {
				return params.validateFlags()
			},
			Run: common.GetCliRunCommand(params),
		}

		addCommonSkylineFlags(subCmd, params)
		chainCommand.addFlags(subCmd, params)

		cmd.AddCommand(subCmd)
	}

	return cmd
}

func (p *sendSkylineTxParams) validateFlags() error {
	if p.txType != "" &&
		p.txType != common.ChainTypeEVMStr &&
		p.txType != common.ChainTypeCardanoStr &&
		p.txType != common.ChainTypeSolanaStr {
		return fmt.Errorf("invalid --%s type not supported", txTypeFlag)
	}

	if err := p.parseAmountFlags(); err != nil {
		return err
	}

	if err := p.validateCommonFlags(); err != nil {
		return err
	}

	if err := p.validateTokenNames(); err != nil {
		return err
	}

	switch p.txType {
	case common.ChainTypeEVMStr:
		if err := p.validateEvmFlags(); err != nil {
			return err
		}
	case common.ChainTypeSolanaStr:
		if err := p.validateSolanaFlags(); err != nil {
			return err
		}
	case common.ChainTypeCardanoStr, "":
		if err := p.validateCardanoFlags(); err != nil {
			return err
		}
	}

	return p.parseReceivers()
}

func (p *sendSkylineTxParams) validateCommonFlags() error {
	if p.privateKeyRaw == "" {
		return fmt.Errorf("flag --%s not specified", privateKeyFlag)
	}

	if len(p.receivers) == 0 {
		return fmt.Errorf("--%s not specified", receiverFlag)
	}

	if p.chainIDsConfig == "" {
		return fmt.Errorf("--%s flag not specified", chainIDsConfigFlag)
	}

	if _, err := os.Stat(p.chainIDsConfig); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("config file does not exist: %s", p.chainIDsConfig)
		}

		return fmt.Errorf("failed to check config file: %s. err: %w", p.chainIDsConfig, err)
	}

	chainIDsConfig, err := common.LoadConfig[common.ChainIDsConfigFile](p.chainIDsConfig, "")
	if err != nil {
		return fmt.Errorf("failed to load chain IDs config: %w", err)
	}

	p.chainIDConverter = chainIDsConfig.ToChainIDConverter()

	if !p.chainIDConverter.IsExistingChainID(p.chainIDSrc) {
		return fmt.Errorf("--%s unsupported chain ID %s", srcChainIDFlag, p.chainIDSrc)
	}

	if !p.chainIDConverter.IsExistingChainID(p.chainIDDst) {
		return fmt.Errorf("--%s unsupported chain ID %s", dstChainIDFlag, p.chainIDDst)
	}

	srcChainConfig := common.GetChainConfig(p.chainIDSrc)
	minFeeForBridging, minOperationFee :=
		srcChainConfig.MinFeeForBridging, srcChainConfig.MinOperationFee

	specifiedBridgingFee, specifiedOperationFee := big.NewInt(0), big.NewInt(0)

	switch p.txType {
	case common.ChainTypeEVMStr:
		specifiedBridgingFee = p.feeAmount
		specifiedOperationFee = p.operationFeeAmount
	case common.ChainTypeCardanoStr:
		specifiedBridgingFee = common.DfmToWei(p.feeAmount)
		specifiedOperationFee = common.DfmToWei(p.operationFeeAmount)
	case common.ChainTypeSolanaStr:
		specifiedBridgingFee = common.LamportsToWei(p.feeAmount)
		specifiedOperationFee = common.LamportsToWei(p.operationFeeAmount)
	}

	if specifiedBridgingFee.Cmp(minFeeForBridging) == -1 {
		return fmt.Errorf("--%s invalid amount: %s", feeAmountFlag, specifiedBridgingFee.String())
	}

	if specifiedOperationFee.Cmp(minOperationFee) == -1 {
		return fmt.Errorf("--%s invalid amount: %s", operationFeeFlag, specifiedOperationFee.String())
	}

	return p.validateDestinationURLFlags()
}

func (p *sendSkylineTxParams) validateDestinationURLFlags() error {
	solanaURLIsDestFlag := p.txType != common.ChainTypeSolanaStr
	rpcURLIsDestFlag := p.txType != common.ChainTypeEVMStr

	switch {
	case p.chainIDConverter.IsEVMChainID(p.chainIDDst):
		if p.ogmiosURLDst != "" {
			return fmt.Errorf("--%s should not be used for an EVM destination; use --%s", ogmiosURLDstFlag, rpcURLFlag)
		}

		if solanaURLIsDestFlag && p.solanaURL != "" {
			return fmt.Errorf("--%s should not be used for an EVM destination; use --%s", solanaURLFlag, rpcURLFlag)
		}

		if p.rpcURL != "" && !common.IsValidHTTPURL(p.rpcURL) {
			return fmt.Errorf("invalid --%s: %s", rpcURLFlag, p.rpcURL)
		}

		if p.rpcURLDst != "" && !common.IsValidHTTPURL(p.rpcURLDst) {
			return fmt.Errorf("invalid --%s: %s", rpcURLDstFlag, p.rpcURLDst)
		}

	case p.chainIDConverter.IsCardanoChainID(p.chainIDDst):
		if rpcURLIsDestFlag && p.rpcURL != "" {
			return fmt.Errorf("--%s should not be used for a Cardano destination; use --%s", rpcURLFlag, ogmiosURLDstFlag)
		}

		if p.rpcURLDst != "" {
			return fmt.Errorf("--%s should not be used for a Cardano destination; use --%s", rpcURLDstFlag, ogmiosURLDstFlag)
		}

		if solanaURLIsDestFlag && p.solanaURL != "" {
			return fmt.Errorf("--%s should not be used for a Cardano destination; use --%s", solanaURLFlag, ogmiosURLDstFlag)
		}

		if p.ogmiosURLDst != "" && !common.IsValidHTTPURL(p.ogmiosURLDst) {
			return fmt.Errorf("invalid --%s: %s", ogmiosURLDstFlag, p.ogmiosURLDst)
		}

	case p.chainIDConverter.IsSolanaChainID(p.chainIDDst):
		if rpcURLIsDestFlag && p.rpcURL != "" {
			return fmt.Errorf("--%s should not be used for a Solana destination; use --%s", rpcURLFlag, solanaURLFlag)
		}

		if p.rpcURLDst != "" {
			return fmt.Errorf("--%s should not be used for a Solana destination; use --%s", rpcURLDstFlag, solanaURLFlag)
		}

		if p.ogmiosURLDst != "" {
			return fmt.Errorf("--%s should not be used for a Solana destination; use --%s", ogmiosURLDstFlag, solanaURLFlag)
		}

		if solanaURLIsDestFlag && p.solanaURL != "" && !common.IsValidHTTPURL(p.solanaURL) {
			return fmt.Errorf("invalid --%s: %s", solanaURLFlag, p.solanaURL)
		}
	}

	return nil
}

func (p *sendSkylineTxParams) validateTokenNames() error {
	if p.txType == common.ChainTypeCardanoStr || p.txType == "" {
		if p.tokenFullNameSrc == "" {
			p.tokenFullNameSrc = cardanowallet.AdaTokenName
		}

		if p.tokenFullNameSrc != cardanowallet.AdaTokenName {
			token, err := cardanowallet.NewTokenWithFullNameTry(p.tokenFullNameSrc)
			if err != nil {
				return fmt.Errorf("--%s invalid token name: %s", fullSrcTokenNameFlag, p.tokenFullNameSrc)
			}

			p.tokenFullNameSrc = token.String()
		}
	}

	if p.tokenFullNameDst == "" {
		return fmt.Errorf("--%s flag not specified", fullDstTokenNameFlag)
	}

	if p.chainIDConverter.IsCardanoChainID(p.chainIDDst) && p.tokenFullNameDst != cardanowallet.AdaTokenName {
		token, err := cardanowallet.NewTokenWithFullNameTry(p.tokenFullNameDst)
		if err != nil {
			return fmt.Errorf("--%s invalid token name: %s", fullDstTokenNameFlag, p.tokenFullNameDst)
		}

		p.tokenFullNameDst = token.String()
	}

	return nil
}

func (p *sendSkylineTxParams) parseAmountFlags() error {
	feeAmount, ok := new(big.Int).SetString(p.feeString, 0)
	if !ok {
		return fmt.Errorf("--%s invalid amount: %s", feeAmountFlag, p.feeString)
	}

	p.feeAmount = feeAmount

	operationFeeAmount, ok := new(big.Int).SetString(p.operationFeeString, 0)
	if !ok {
		return fmt.Errorf("--%s invalid amount: %s", operationFeeFlag, p.operationFeeString)
	}

	p.operationFeeAmount = operationFeeAmount

	return nil
}

func (p *sendSkylineTxParams) validateEvmFlags() error {
	if !p.chainIDConverter.IsEVMChainID(p.chainIDSrc) {
		return fmt.Errorf("--%s must be an EVM chain when using chain type %s", srcChainIDFlag, common.ChainTypeEVMStr)
	}

	if !common.IsValidHTTPURL(p.rpcURL) {
		return fmt.Errorf("--%s not specified or invalid", rpcURLFlag)
	}

	if p.gatewayAddress == "" {
		return fmt.Errorf("--%s not specified", gatewayAddressFlag)
	}

	if !common.IsValidAddress(common.ChainIDStrNexus, p.gatewayAddress, p.chainIDConverter) {
		return fmt.Errorf("invalid address for flag --%s", gatewayAddressFlag)
	}

	if p.tokenContractAddrSrc != "" &&
		!common.IsValidAddress(p.chainIDSrc, p.tokenContractAddrSrc, p.chainIDConverter) {
		return fmt.Errorf("invalid address for flag --%s", tokenContractAddrSrcFlag)
	}

	if p.nativeTokenWalletContractAddress != "" &&
		!common.IsValidAddress(common.ChainIDStrNexus, p.nativeTokenWalletContractAddress, p.chainIDConverter) {
		return fmt.Errorf("invalid address for flag --%s", nativeTokenWalletContractAddrFlag)
	}

	return nil
}

func (p *sendSkylineTxParams) validateCardanoFlags() error {
	if !p.chainIDConverter.IsCardanoChainID(p.chainIDSrc) {
		return fmt.Errorf("--%s must be a Cardano chain when using chain type %s", srcChainIDFlag, common.ChainTypeCardanoStr)
	}

	bytes, err := cardanotx.GetCardanoPrivateKeyBytes(p.privateKeyRaw)
	if err != nil {
		return fmt.Errorf("invalid --%s value %s", privateKeyFlag, p.privateKeyRaw)
	}

	var stakeBytes []byte
	if len(p.stakePrivateKeyRaw) > 0 {
		stakeBytes, err = cardanotx.GetCardanoPrivateKeyBytes(p.stakePrivateKeyRaw)
		if err != nil {
			return fmt.Errorf("invalid --%s value %s", stakePrivateKeyFlag, p.stakePrivateKeyRaw)
		}
	}

	p.wallet = cardanowallet.NewWallet(bytes, stakeBytes)

	if !common.IsValidHTTPURL(p.ogmiosURLSrc) {
		return fmt.Errorf("invalid --%s: %s", ogmiosURLSrcFlag, p.ogmiosURLSrc)
	}

	if p.multisigAddrSrc == "" {
		return fmt.Errorf("--%s not specified", multisigAddrSrcFlag)
	}

	if p.treasuryAddrSrc == "" {
		return fmt.Errorf("--%s not specified", treasuryAddrSrcFlag)
	}

	if p.tokenContractAddrDst != "" &&
		!common.IsValidAddress(p.chainIDDst, p.tokenContractAddrDst, p.chainIDConverter) {
		return fmt.Errorf("invalid address for flag --%s", tokenContractAddrDstFlag)
	}

	return nil
}

func (p *sendSkylineTxParams) validateSolanaFlags() error {
	if !p.chainIDConverter.IsSolanaChainID(p.chainIDSrc) {
		return fmt.Errorf("--%s must be a Solana chain when using chain type %s", srcChainIDFlag, common.ChainTypeSolanaStr)
	}

	if !common.IsValidHTTPURL(p.solanaURL) {
		return fmt.Errorf("invalid --%s: %s", solanaURLFlag, p.solanaURL)
	}

	if p.solanaProgramID == "" {
		return fmt.Errorf("--%s not specified", solanaProgramIDFlag)
	}

	programID, err := solanawallet.PublicKeyFromAddress(p.solanaProgramID)
	if err != nil {
		return fmt.Errorf("invalid --%s: %s", solanaProgramIDFlag, p.solanaProgramID)
	}

	p.solanaProgramPublicKey = programID
	if p.treasuryAddrSrc == "" {
		return fmt.Errorf("--%s not specified", treasuryAddrSrcFlag)
	}

	treasury, err := solanawallet.PublicKeyFromAddress(p.treasuryAddrSrc)
	if err != nil {
		return fmt.Errorf("invalid --%s: %s", treasuryAddrSrcFlag, p.treasuryAddrSrc)
	}

	p.solanaTreasuryAddress = treasury

	wallet, err := solanawallet.NewWalletFromPrivateKey(p.privateKeyRaw)
	if err != nil {
		return fmt.Errorf("invalid --%s value", privateKeyFlag)
	}

	p.solanaWallet = wallet

	return nil
}

func (p *sendSkylineTxParams) parseReceivers() error {
	receivers := make([]*receiverAmount, 0, len(p.receivers))

	for i, x := range p.receivers {
		vals := strings.Split(x, ":")
		if len(vals) != 2 {
			return fmt.Errorf("--%s number %d is invalid: %s", receiverFlag, i, x)
		}

		amount, ok := new(big.Int).SetString(vals[1], 0)
		if !ok {
			return fmt.Errorf("--%s number %d has invalid amount: %s", receiverFlag, i, x)
		}

		if !common.IsValidAddress(p.chainIDDst, vals[0], p.chainIDConverter) {
			return fmt.Errorf("--%s number %d has invalid address: %s", receiverFlag, i, x)
		}

		receivers = append(receivers, &receiverAmount{
			ReceiverAddr: vals[0],
			Amount:       amount,
		})
	}

	p.receiversParsed = receivers

	return nil
}

func addCommonSkylineFlags(cmd *cobra.Command, p *sendSkylineTxParams) {
	cmd.Flags().StringVar(
		&p.privateKeyRaw,
		privateKeyFlag,
		"",
		privateKeyFlagDesc,
	)

	cmd.Flags().StringArrayVar(
		&p.receivers,
		receiverFlag,
		nil,
		receiverFlagDesc,
	)

	cmd.Flags().StringVar(
		&p.chainIDSrc,
		srcChainIDFlag,
		"",
		srcChainIDFlagDesc,
	)

	cmd.Flags().StringVar(
		&p.chainIDDst,
		dstChainIDFlag,
		"",
		dstChainIDFlagDesc,
	)

	cmd.Flags().StringVar(
		&p.feeString,
		feeAmountFlag,
		"0",
		feeAmountFlagDesc,
	)

	cmd.Flags().StringVar(
		&p.operationFeeString,
		operationFeeFlag,
		"0",
		operationFeeFlagDesc,
	)

	cmd.Flags().Uint16Var(
		&p.tokenIDSrc,
		tokenIDSrcFlag,
		0,
		tokenIDSrcFlagDesc,
	)

	cmd.Flags().StringVar(
		&p.tokenFullNameSrc,
		fullSrcTokenNameFlag,
		"",
		fullSrcTokenNameFlagDesc,
	)

	cmd.Flags().StringVar(
		&p.chainIDsConfig,
		chainIDsConfigFlag,
		"",
		chainIDsConfigFlagDesc,
	)

	cmd.Flags().StringVar(
		&p.tokenFullNameDst,
		fullDstTokenNameFlag,
		cardanowallet.AdaTokenName,
		fullDstTokenNameFlagDesc,
	)

	cmd.Flags().StringVar(
		&p.tokenContractAddrDst,
		tokenContractAddrDstFlag,
		"",
		tokenContractAddrDstFlagDesc,
	)

	cmd.MarkFlagsMutuallyExclusive(fullDstTokenNameFlag, tokenContractAddrDstFlag)
}

func addCardanoSkylineFlags(cmd *cobra.Command, p *sendSkylineTxParams) {
	cmd.Flags().StringVar(
		&p.stakePrivateKeyRaw,
		stakePrivateKeyFlag,
		"",
		stakePrivateKeyFlagDesc,
	)
	cmd.Flags().StringVar(
		&p.ogmiosURLSrc,
		ogmiosURLSrcFlag,
		"",
		ogmiosURLSrcFlagDesc,
	)
	cmd.Flags().UintVar(
		&p.networkIDSrc,
		networkIDSrcFlag,
		0,
		networkIDSrcFlagDesc,
	)
	cmd.Flags().UintVar(
		&p.testnetMagicSrc,
		testnetMagicFlag,
		0,
		testnetMagicFlagDesc,
	)
	cmd.Flags().StringVar(
		&p.multisigAddrSrc,
		multisigAddrSrcFlag,
		"",
		multisigAddrSrcFlagDesc,
	)
	cmd.Flags().StringVar(
		&p.treasuryAddrSrc,
		treasuryAddrSrcFlag,
		"",
		treasuryAddrSrcFlagDesc,
	)
	cmd.Flags().StringVar(
		&p.ogmiosURLDst,
		ogmiosURLDstFlag,
		"",
		ogmiosURLDstFlagDesc,
	)
	cmd.Flags().StringVar(
		&p.rpcURL,
		rpcURLFlag,
		"",
		rpcURLFlagDesc,
	)
	cmd.Flags().StringVar(
		&p.solanaURL,
		solanaURLFlag,
		"",
		solanaURLFlagDesc,
	)
	cmd.Flags().StringVar(
		&p.cardanoCliBinaryName,
		cardanoCliBinaryNameFlag,
		"",
		cardanoCliBinaryNameFlagDesc,
	)

	cmd.MarkFlagsMutuallyExclusive(ogmiosURLDstFlag, rpcURLFlag, solanaURLFlag)
}

func addEvmSkylineFlags(cmd *cobra.Command, p *sendSkylineTxParams) {
	cmd.Flags().StringVar(
		&p.gatewayAddress,
		gatewayAddressFlag,
		"",
		gatewayAddressFlagDesc,
	)
	cmd.Flags().StringVar(
		&p.nativeTokenWalletContractAddress,
		nativeTokenWalletContractAddrFlag,
		"",
		nativeTokenWalletContractAddrFlagDesc,
	)
	cmd.Flags().StringVar(
		&p.rpcURL,
		rpcURLFlag,
		"",
		rpcURLFlagDesc,
	)
	cmd.Flags().StringVar(
		&p.ogmiosURLDst,
		ogmiosURLDstFlag,
		"",
		ogmiosURLDstFlagDesc,
	)
	cmd.Flags().StringVar(
		&p.solanaURL,
		solanaURLFlag,
		"",
		solanaURLFlagDesc,
	)
	cmd.Flags().StringVar(
		&p.tokenContractAddrSrc,
		tokenContractAddrSrcFlag,
		"",
		tokenContractAddrSrcFlagDesc,
	)
	cmd.Flags().StringVar(
		&p.rpcURLDst,
		rpcURLDstFlag,
		"",
		rpcURLDstFlagDesc,
	)

	cmd.MarkFlagsMutuallyExclusive(ogmiosURLDstFlag, solanaURLFlag, rpcURLDstFlag)
}

func addSolanaSkylineFlags(cmd *cobra.Command, p *sendSkylineTxParams) {
	cmd.Flags().StringVar(
		&p.solanaURL,
		solanaURLFlag,
		"",
		solanaURLFlagDesc,
	)
	cmd.Flags().StringVar(
		&p.solanaProgramID,
		solanaProgramIDFlag,
		"",
		solanaProgramIDFlagDesc,
	)
	cmd.Flags().StringVar(
		&p.treasuryAddrSrc,
		treasuryAddrSrcFlag,
		"",
		treasuryAddrSrcFlagDesc,
	)
	cmd.Flags().StringVar(
		&p.rpcURL,
		rpcURLFlag,
		"",
		rpcURLFlagDesc,
	)
	cmd.Flags().StringVar(
		&p.ogmiosURLDst,
		ogmiosURLDstFlag,
		"",
		ogmiosURLDstFlagDesc,
	)

	cmd.MarkFlagsMutuallyExclusive(ogmiosURLDstFlag, rpcURLFlag)
}

func (p *sendSkylineTxParams) Execute(
	outputter common.OutputFormatter,
) (common.ICommandResult, error) {
	ctx := context.Background()

	switch p.txType {
	case common.ChainTypeEVMStr:
		return p.executeEvm(ctx, outputter)
	case common.ChainTypeSolanaStr:
		return p.executeSolana(ctx, outputter)
	case common.ChainTypeCardanoStr, "":
		return p.executeCardano(ctx, outputter)
	default:
		return nil, fmt.Errorf("txType not supported")
	}
}

func (p *sendSkylineTxParams) executeCardano(ctx context.Context, outputter common.OutputFormatter) (
	common.ICommandResult, error,
) {
	receivers := toSkylineCardanoMetadata(p.receiversParsed, p.tokenIDSrc)
	networkID := cardanowallet.CardanoNetworkType(p.networkIDSrc)

	srcConfig := common.GetChainConfig(p.chainIDSrc)
	dstConfig := common.GetChainConfig(p.chainIDDst)

	srcTokens := map[uint16]sendtx.ApexToken{
		p.tokenIDSrc: {
			FullName:          p.tokenFullNameSrc,
			IsWrappedCurrency: p.tokenFullNameDst == cardanowallet.AdaTokenName,
		},
	}

	currencyTokenID := apexTokenID
	if p.chainIDSrc == common.ChainIDStrCardano {
		currencyTokenID = adaTokenID
	}

	if p.tokenIDSrc != currencyTokenID {
		srcTokens[currencyTokenID] = sendtx.ApexToken{
			FullName:          cardanowallet.AdaTokenName,
			IsWrappedCurrency: false,
		}
	}

	txSender := sendtx.NewTxSender(
		map[string]sendtx.ChainConfig{
			p.chainIDSrc: {
				CardanoCliBinary:           cardanowallet.ResolveCardanoCliBinary(p.cardanoCliBinaryName),
				TxProvider:                 cardanowallet.NewTxProviderOgmios(p.ogmiosURLSrc),
				TestNetMagic:               p.testnetMagicSrc,
				TTLSlotNumberInc:           ttlSlotNumberInc,
				DefaultMinFeeForBridging:   common.WeiToDfm(srcConfig.MinFeeForBridging).Uint64(),
				MinFeeForBridgingTokens:    common.WeiToDfm(srcConfig.MinFeeForBridging).Uint64(),
				MinOperationFeeAmount:      common.WeiToDfm(srcConfig.MinOperationFee).Uint64(),
				MinUtxoValue:               srcConfig.MinUtxoAmount,
				MinColCoinsAllowedToBridge: common.WeiToDfm(srcConfig.MinColCoinsAllowedToBridge).Uint64(),
				Tokens:                     srcTokens,
				TreasuryAddress:            p.treasuryAddrSrc,
			},
			p.chainIDDst: {
				MinUtxoValue:             dstConfig.MinUtxoAmount,
				DefaultMinFeeForBridging: common.WeiToDfm(dstConfig.MinFeeForBridging).Uint64(),
				MinFeeForBridgingTokens:  common.WeiToDfm(dstConfig.MinFeeForBridging).Uint64(),
				MinOperationFeeAmount:    common.WeiToDfm(dstConfig.MinOperationFee).Uint64(),
			},
		},
		sendtx.WithMinAmountToBridge(srcConfig.MinUtxoAmount),
	)

	senderAddr, err := cardanotx.GetAddress(networkID, p.wallet)
	if err != nil {
		return nil, err
	}

	txInfo, _, err := txSender.CreateBridgingTx(
		ctx,
		sendtx.BridgingTxDto{
			SrcChainID:      p.chainIDSrc,
			DstChainID:      p.chainIDDst,
			SenderAddr:      senderAddr.String(),
			Receivers:       receivers,
			BridgingAddress: p.multisigAddrSrc,
			BridgingFee:     p.feeAmount.Uint64(),
			OperationFee:    p.operationFeeAmount.Uint64(),
		},
	)
	if err != nil {
		return nil, err
	}

	_, _ = outputter.Write([]byte("Submiting bridging transaction..."))
	outputter.WriteOutput()

	err = txSender.SubmitTx(ctx, p.chainIDSrc, txInfo.TxRaw, p.wallet)
	if err != nil {
		return nil, err
	}

	_, _ = outputter.Write([]byte(fmt.Sprintf("transaction has been submitted: %s", txInfo.TxHash)))
	outputter.WriteOutput()

	_, err = p.waitForSkylineDestinationTx(ctx)
	if err != nil {
		return nil, err
	}

	_, _ = outputter.Write([]byte("Transaction has been bridged"))
	outputter.WriteOutput()

	return CmdResult{
		SenderAddr: senderAddr.String(),
		ChainID:    p.chainIDDst,
		Receipts:   p.receiversParsed,
		TxHash:     txInfo.TxHash,
	}, nil
}

func (p *sendSkylineTxParams) executeSolana(ctx context.Context, outputter common.OutputFormatter) (
	common.ICommandResult, error,
) {
	txProvider, err := solanawallet.NewProvider(p.solanaURL, nil)
	if err != nil {
		return nil, err
	}

	txSender := solsendtx.NewTxSender(txProvider, &solsendtx.ChainConfig{
		TreasuryAddress: p.solanaTreasuryAddress,
	})

	txReceivers := make([]solsendtx.BridgingTxReceiver, 0, len(p.receiversParsed))

	for idx, rec := range p.receiversParsed {
		if !rec.Amount.IsUint64() {
			return nil, fmt.Errorf("--%s number %d has amount too large for Solana: %s", receiverFlag, idx, rec.Amount.String())
		}

		txReceivers = append(txReceivers, solsendtx.BridgingTxReceiver{
			Address: rec.ReceiverAddr,
			TokenAmount: solanawallet.TokenAmount{
				TokenID: p.tokenIDSrc,
				Amount:  rec.Amount.Uint64(),
			},
		})
	}

	txDto := solsendtx.BridgeRequestDto{
		Ctx:          ctx,
		ProgramID:    p.solanaProgramPublicKey,
		DstChainID:   p.chainIDDst,
		SenderAddr:   p.solanaWallet.PublicKey.String(),
		Receivers:    txReceivers,
		BridgingFee:  p.feeAmount.Uint64(),
		OperationFee: p.operationFeeAmount.Uint64(),
	}

	recentBlockhash, err := txProvider.GetLatestBlockhash(ctx)
	if err != nil {
		return nil, err
	}

	tx, err := txSender.CreateTx(
		ctx,
		p.solanaWallet.PublicKey,
		solsendtx.InstructionTypeBridgingRequest,
		recentBlockhash,
		txDto,
	)
	if err != nil {
		return nil, err
	}

	_, err = tx.Sign(func(key solana.PublicKey) *solana.PrivateKey {
		return &p.solanaWallet.PrivateKey
	})
	if err != nil {
		return nil, fmt.Errorf("error while signing tx: %w", err)
	}

	_, _ = outputter.Write([]byte("Submiting bridging transaction..."))
	outputter.WriteOutput()

	sig, err := txSender.SendTx(ctx, tx)
	if err != nil {
		return nil, err
	}

	if err := txProvider.WaitForSignature(ctx, *sig, rpc.CommitmentFinalized, solanaWaitForSignatureTimeout); err != nil {
		return nil, fmt.Errorf("wait for bridging request confirmation: %w", err)
	}

	_, _ = outputter.Write([]byte(fmt.Sprintf("transaction has been submitted: %s", sig.String())))
	outputter.WriteOutput()

	_, err = p.waitForSkylineDestinationTx(ctx)
	if err != nil {
		return nil, err
	}

	_, _ = outputter.Write([]byte("Transaction has been bridged"))
	outputter.WriteOutput()

	return CmdResult{
		SenderAddr: p.solanaWallet.PublicKey.String(),
		ChainID:    p.chainIDDst,
		Receipts:   p.receiversParsed,
		TxHash:     sig.String(),
	}, nil
}

func (p *sendSkylineTxParams) executeEvm(ctx context.Context, outputter common.OutputFormatter) (
	common.ICommandResult, error,
) {
	contractAddress := common.HexToAddress(p.gatewayAddress)
	chainID := p.chainIDConverter.ToChainIDNum(p.chainIDDst)
	receivers, totalTokenAmount := toSkylineGatewayStruct(p.receiversParsed, p.tokenIDSrc)

	totalAmount := big.NewInt(0)
	totalAmount.Add(totalAmount, p.feeAmount)
	totalAmount.Add(totalAmount, p.operationFeeAmount)

	// If transferring native currency, add total token amount to total amount
	if p.tokenFullNameSrc == cardanowallet.AdaTokenName {
		totalAmount.Add(totalAmount, totalTokenAmount)
	}

	wallet, err := ethtxhelper.NewEthTxWallet(p.privateKeyRaw)
	if err != nil {
		return nil, err
	}

	txHelper, err := getTxHelper(p.rpcURL)
	if err != nil {
		return nil, err
	}

	if p.tokenContractAddrSrc != "" {
		_, _ = outputter.Write([]byte("submitting approve tx..."))
		outputter.WriteOutput()

		parsed, err := abi.JSON(strings.NewReader(approveERC20ABI))
		if err != nil {
			return nil, err
		}

		client := txHelper.GetClient()

		erc20Contract := bind.NewBoundContract(
			common.HexToAddress(p.tokenContractAddrSrc),
			parsed,
			client,
			client,
			client,
		)

		tx, err := infracommon.ExecuteWithRetry(ctx, func(ctx context.Context) (*types.Transaction, error) {
			return txHelper.SendTx(ctx, wallet, bind.TransactOpts{},
				func(opts *bind.TransactOpts) (*types.Transaction, error) {
					return erc20Contract.Transact(
						opts, "approve", ethcommon.HexToAddress(p.nativeTokenWalletContractAddress), totalTokenAmount)
				})
		})
		if err != nil {
			return nil, err
		}

		_, _ = outputter.Write([]byte(fmt.Sprintf("approve transaction has been submitted: %s", tx.Hash())))
		outputter.WriteOutput()

		receipt, err := txHelper.WaitForReceipt(ctx, tx.Hash().String())
		if err != nil {
			return nil, err
		} else if receipt.Status != types.ReceiptStatusSuccessful {
			return nil, fmt.Errorf("approve transaction receipt status is unsuccessful, receipt: %+v", receipt)
		}
	}

	contract, err := contractbinding.NewGateway(contractAddress, txHelper.GetClient())
	if err != nil {
		return nil, err
	}

	_, _ = outputter.Write([]byte("Estimating gas..."))
	outputter.WriteOutput()

	abi, err := contractbinding.GatewayMetaData.GetAbi()
	if err != nil {
		return nil, err
	}

	estimatedGas, _, err := txHelper.EstimateGas(
		ctx, wallet.GetAddress(), contractAddress, totalAmount, gasLimitMultiplier,
		abi, "withdraw", chainID, receivers, p.feeAmount, p.operationFeeAmount)
	if err != nil {
		return nil, err
	}

	_, _ = outputter.Write([]byte("Submiting bridging transaction..."))
	outputter.WriteOutput()

	tx, err := txHelper.SendTx(ctx, wallet,
		bind.TransactOpts{
			GasLimit: estimatedGas,
			Value:    totalAmount,
		},
		func(txOpts *bind.TransactOpts) (*types.Transaction, error) {
			return contract.Withdraw(
				txOpts, chainID, receivers, p.feeAmount, p.operationFeeAmount,
			)
		})
	if err != nil {
		return nil, err
	}

	_, _ = outputter.Write([]byte(fmt.Sprintf("transaction has been submitted: %s", tx.Hash())))
	outputter.WriteOutput()

	receipt, err := txHelper.WaitForReceipt(ctx, tx.Hash().String())
	if err != nil {
		return nil, err
	} else if receipt.Status != types.ReceiptStatusSuccessful {
		return nil, fmt.Errorf("transaction receipt status is unsuccessful, receipt: %+v", receipt)
	}

	waitedDestinationTx, err := p.waitForSkylineDestinationTx(ctx)
	if err != nil {
		return nil, err
	}

	if waitedDestinationTx {
		_, _ = outputter.Write([]byte("Transaction has been bridged"))
		outputter.WriteOutput()
	}

	return CmdResult{
		SenderAddr: wallet.GetAddress().String(),
		ChainID:    p.chainIDDst,
		Receipts:   p.receiversParsed,
		TxHash:     receipt.TxHash.String(),
	}, nil
}

func waitForEvmSkylineTx(
	ctx context.Context, client *ethclient.Client, tokenContractAddr string, receivers []*receiverAmount) error {
	if tokenContractAddr == "" {
		return waitForTx(ctx, receivers, func(ctx context.Context, addr string) (*big.Int, error) {
			return client.BalanceAt(ctx, common.HexToAddress(addr), nil)
		})
	}

	return waitForTx(ctx, receivers, func(ctx context.Context, addr string) (*big.Int, error) {
		return getERC20Balance(client, common.HexToAddress(tokenContractAddr), common.HexToAddress(addr))
	})
}

func waitForCardanoSkylineTx(
	ctx context.Context, txUtxoRetriever cardanowallet.IUTxORetriever,
	tokenName string, receivers []*receiverAmount) error {
	return waitForTx(ctx, receivers, func(ctx context.Context, addr string) (*big.Int, error) {
		utxos, err := txUtxoRetriever.GetUtxos(ctx, addr)
		if err != nil {
			return nil, err
		}

		return new(big.Int).SetUint64(cardanowallet.GetUtxosSum(utxos)[tokenName]), nil
	})
}

func waitForSolanaSkylineTx(
	ctx context.Context, solanaURL, tokenName string, receivers []*receiverAmount) error {
	return waitForTx(ctx, receivers, func(ctx context.Context, addr string) (*big.Int, error) {
		balance, err := solanawallet.GetAddressBalanceWithTokenNameLamports(ctx, addr, solanaURL, tokenName)
		if err != nil {
			return nil, err
		}

		return balance[tokenName], nil
	})
}

// receiversInDfm converts receiver amounts to DFM.
// Cardano inputs are already in DFM; Solana inputs are in lamports; EVM inputs are in Wei.
func (p *sendSkylineTxParams) receiversInDfm() []*receiverAmount {
	switch p.txType {
	case common.ChainTypeSolanaStr:
		return convertReceiversAmounts(p.receiversParsed, common.LamportsToDfm)
	case common.ChainTypeEVMStr:
		return convertReceiversAmounts(p.receiversParsed, common.WeiToDfm)
	default: // Cardano — already in DFM
		return p.receiversParsed
	}
}

func (p *sendSkylineTxParams) waitForSkylineDestinationTx(ctx context.Context) (bool, error) {
	receiversDfm := p.receiversInDfm()

	if p.chainIDConverter.IsCardanoChainID(p.chainIDDst) && p.ogmiosURLDst != "" {
		return true, waitForCardanoSkylineTx(
			ctx,
			cardanowallet.NewTxProviderOgmios(p.ogmiosURLDst),
			p.tokenFullNameDst,
			receiversDfm,
		)
	}

	if p.chainIDConverter.IsEVMChainID(p.chainIDDst) {
		rpcURLDst := p.rpcURL
		if p.txType == common.ChainTypeEVMStr {
			rpcURLDst = p.rpcURLDst
		}

		if rpcURLDst != "" {
			txHelper, err := getTxHelper(rpcURLDst)
			if err != nil {
				return false, err
			}

			return true, waitForEvmSkylineTx(ctx, txHelper.GetClient(), p.tokenContractAddrDst,
				convertReceiversAmounts(receiversDfm, common.DfmToWei))
		}

		return false, nil
	}

	if p.chainIDConverter.IsSolanaChainID(p.chainIDDst) && p.solanaURL != "" {
		return true, waitForSolanaSkylineTx(ctx, p.solanaURL, p.tokenFullNameDst,
			convertReceiversAmounts(receiversDfm, common.DfmToLamports))
	}

	return false, nil
}

func convertReceiversAmounts(
	receivers []*receiverAmount,
	convertAmount func(*big.Int) *big.Int,
) []*receiverAmount {
	convertedReceivers := make([]*receiverAmount, len(receivers))
	for i, rec := range receivers {
		convertedReceivers[i] = &receiverAmount{
			ReceiverAddr: rec.ReceiverAddr,
			Amount:       convertAmount(rec.Amount),
		}
	}

	return convertedReceivers
}

const approveERC20ABI = `
[
    {
      "inputs": [
        {
          "internalType": "address",
          "name": "spender",
          "type": "address"
        },
        {
          "internalType": "uint256",
          "name": "value",
          "type": "uint256"
        }
      ],
      "name": "approve",
      "outputs": [
        {
          "internalType": "bool",
          "name": "",
          "type": "bool"
        }
      ],
      "stateMutability": "nonpayable",
      "type": "function"
    }
]`

const balanceOfERC20ABI = `
[
    {
      "inputs": [
        {
          "internalType": "address",
          "name": "account",
          "type": "address"
        }
      ],
      "name": "balanceOf",
      "outputs": [
        {
          "internalType": "uint256",
          "name": "",
          "type": "uint256"
        }
      ],
      "stateMutability": "view",
      "type": "function"
    }
]`

func getERC20Balance(
	client *ethclient.Client, tokenContractAddr ethcommon.Address, addr ethcommon.Address,
) (*big.Int, error) {
	parsedABI, err := abi.JSON(strings.NewReader(balanceOfERC20ABI))
	if err != nil {
		return nil, fmt.Errorf("failed to parse abi. err: %w", err)
	}

	contract := bind.NewBoundContract(tokenContractAddr, parsedABI, client, client, client)

	var out []interface{}

	err = contract.Call(&bind.CallOpts{}, &out, "balanceOf", addr)
	if err != nil {
		return nil, fmt.Errorf("failed to call contract. err: %w", err)
	}

	balance, ok := out[0].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("failed to convert erc20 balanceOf result to big.Int")
	}

	return balance, nil
}

func toSkylineCardanoMetadata(receivers []*receiverAmount, tokenID uint16) []sendtx.BridgingTxReceiver {
	metadataReceivers := make([]sendtx.BridgingTxReceiver, len(receivers))
	for idx, rec := range receivers {
		metadataReceivers[idx] = sendtx.BridgingTxReceiver{
			Addr:    rec.ReceiverAddr,
			Amount:  rec.Amount.Uint64(),
			TokenID: tokenID,
		}
	}

	return metadataReceivers
}

func toSkylineGatewayStruct(receivers []*receiverAmount, tokenID uint16) (
	[]contractbinding.IGatewayStructsReceiverWithdraw, *big.Int,
) {
	total := big.NewInt(0)

	gatewayOutputs := make([]contractbinding.IGatewayStructsReceiverWithdraw, len(receivers))
	for idx, rec := range receivers {
		gatewayOutputs[idx] = contractbinding.IGatewayStructsReceiverWithdraw{
			Receiver: rec.ReceiverAddr,
			Amount:   rec.Amount,
			TokenId:  tokenID,
		}

		total.Add(total, rec.Amount)
	}

	return gatewayOutputs, total
}
