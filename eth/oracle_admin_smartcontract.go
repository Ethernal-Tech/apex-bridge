package eth

import (
	"context"
	"fmt"

	"github.com/Ethernal-Tech/apex-bridge/contractbinding"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethcommon "github.com/ethereum/go-ethereum/common"
)

type IOracleAdminSmartContract interface {
	GetValidatorChangeStatus(ctx context.Context) (bool, error)
}

type OracleAdminSmartContractImpl struct {
	smartContractAddress ethcommon.Address
	ethHelper            *EthHelperWrapper
}

var _ IOracleAdminSmartContract = (*OracleAdminSmartContractImpl)(nil)

func NewOracleAdminSmartContract(
	smartContractAddress string, ethHelper *EthHelperWrapper,
) *OracleAdminSmartContractImpl {
	return &OracleAdminSmartContractImpl{
		smartContractAddress: ethcommon.HexToAddress(smartContractAddress),
		ethHelper:            ethHelper,
	}
}

func (asc *OracleAdminSmartContractImpl) GetValidatorChangeStatus(
	ctx context.Context,
) (bool, error) {
	ethTxHelper, err := asc.ethHelper.GetEthHelper()
	if err != nil {
		return false, fmt.Errorf("error while GetEthHelper: %w", err)
	}

	contract, err := contractbinding.NewAdminContract(
		asc.smartContractAddress,
		ethTxHelper.GetClient())
	if err != nil {
		return false, fmt.Errorf("error while NewAdminContract: %w", asc.ethHelper.ProcessError(err))
	}

	result, err := contract.ValidatorChange(&bind.CallOpts{
		Context: ctx,
	})
	if err != nil {
		return false, fmt.Errorf("error while ValidatorChange: %w", asc.ethHelper.ProcessError(err))
	}

	return result, nil
}
