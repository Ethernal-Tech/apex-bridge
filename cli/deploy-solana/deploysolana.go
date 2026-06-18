package clideploysolana

import (
	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/spf13/cobra"
)

const deployProgramCommandUse = "deploy-program"
const upgradeProgramCommandUse = "upgrade-program"
const createALTCommandUse = "create-alt"
const extendALTCommandUse = "extend-alt"
const registerLockUnlockTokenCommandUse = "register-lock-unlock-token"
const registerMintBurnTokenCommandUse = "register-mint-burn-token"
const hotWalletIncrementCommandUse = "hot-wallet-increment"
const updateFeeConfigCommandUse = "update-fee-config"

var deployProgramParamsData = &deployProgramParams{}
var upgradeProgramParamsData = &upgradeProgramParams{}
var createALTParamsData = &createALTParams{}
var extendALTParamsData = &extendALTParams{}
var registerLockUnlockTokenParamsData = &registerLockUnlockTokenParams{}
var registerMintBurnTokenParamsData = &registerMintBurnTokenParams{}
var hotWalletIncrementParamsData = &hotWalletIncrementParams{}
var updateFeeConfigParamsData = &updateFeeConfigParams{}

func GetDeploySolanaCommand() *cobra.Command {
	cmdDeploySolana := &cobra.Command{
		Use:   "deploy-solana",
		Short: "deploy Solana program to a cluster",
	}

	cmdDeployProgram := &cobra.Command{
		Use:     deployProgramCommandUse,
		Short:   "deploy and initialize a Solana skyline program",
		PreRunE: runPreRun,
		Run:     common.GetCliRunCommand(deployProgramParamsData),
	}
	cmdUpgradeProgram := &cobra.Command{
		Use:     upgradeProgramCommandUse,
		Short:   "upgrade an existing Solana program using the solana CLI",
		PreRunE: runPreRun,
		Run:     common.GetCliRunCommand(upgradeProgramParamsData),
	}
	cmdCreateALT := &cobra.Command{
		Use:     createALTCommandUse,
		Short:   "create a new Solana address lookup table (ALT)",
		PreRunE: runPreRun,
		Run:     common.GetCliRunCommand(createALTParamsData),
	}
	cmdExtendALT := &cobra.Command{
		Use:     extendALTCommandUse,
		Short:   "extend Solana address lookup table (ALT) with bridge addresses",
		PreRunE: runPreRun,
		Run:     common.GetCliRunCommand(extendALTParamsData),
	}
	cmdRegisterLockUnlockToken := &cobra.Command{
		Use:     registerLockUnlockTokenCommandUse,
		Short:   "register lock/unlock token for skyline Solana program",
		PreRunE: runPreRun,
		Run:     common.GetCliRunCommand(registerLockUnlockTokenParamsData),
	}
	cmdRegisterMintBurnToken := &cobra.Command{
		Use:     registerMintBurnTokenCommandUse,
		Short:   "register mint/burn token for skyline Solana program",
		PreRunE: runPreRun,
		Run:     common.GetCliRunCommand(registerMintBurnTokenParamsData),
	}
	cmdHotWalletIncrement := &cobra.Command{
		Use:     hotWalletIncrementCommandUse,
		Short:   "submit hot wallet increment tx for skyline Solana program",
		PreRunE: runPreRun,
		Run:     common.GetCliRunCommand(hotWalletIncrementParamsData),
	}
	cmdUpdateFeeConfig := &cobra.Command{
		Use:     updateFeeConfigCommandUse,
		Short:   "update fee configuration for skyline Solana program",
		PreRunE: runPreRun,
		Run:     common.GetCliRunCommand(updateFeeConfigParamsData),
	}

	deployProgramParamsData.setFlags(cmdDeployProgram)
	upgradeProgramParamsData.setFlags(cmdUpgradeProgram)
	createALTParamsData.setFlags(cmdCreateALT)
	extendALTParamsData.setFlags(cmdExtendALT)
	registerLockUnlockTokenParamsData.setFlags(cmdRegisterLockUnlockToken)
	registerMintBurnTokenParamsData.setFlags(cmdRegisterMintBurnToken)
	hotWalletIncrementParamsData.setFlags(cmdHotWalletIncrement)
	updateFeeConfigParamsData.setFlags(cmdUpdateFeeConfig)
	cmdDeploySolana.AddCommand(cmdDeployProgram)
	cmdDeploySolana.AddCommand(cmdUpgradeProgram)
	cmdDeploySolana.AddCommand(cmdCreateALT)
	cmdDeploySolana.AddCommand(cmdExtendALT)
	cmdDeploySolana.AddCommand(cmdRegisterLockUnlockToken)
	cmdDeploySolana.AddCommand(cmdRegisterMintBurnToken)
	cmdDeploySolana.AddCommand(cmdHotWalletIncrement)
	cmdDeploySolana.AddCommand(cmdUpdateFeeConfig)

	return cmdDeploySolana
}

func runPreRun(cb *cobra.Command, _ []string) error {
	if cb.Use == deployProgramCommandUse {
		return deployProgramParamsData.validateFlags()
	}

	if cb.Use == upgradeProgramCommandUse {
		return upgradeProgramParamsData.validateFlags()
	}

	if cb.Use == createALTCommandUse {
		return createALTParamsData.validateFlags()
	}

	if cb.Use == extendALTCommandUse {
		return extendALTParamsData.validateFlags()
	}

	if cb.Use == registerLockUnlockTokenCommandUse {
		return registerLockUnlockTokenParamsData.validateFlags()
	}

	if cb.Use == registerMintBurnTokenCommandUse {
		return registerMintBurnTokenParamsData.validateFlags()
	}

	if cb.Use == hotWalletIncrementCommandUse {
		return hotWalletIncrementParamsData.validateFlags()
	}

	if cb.Use == updateFeeConfigCommandUse {
		return updateFeeConfigParamsData.validateFlags()
	}

	return nil
}
