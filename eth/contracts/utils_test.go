package ethcontracts

import (
	"context"
	"embed"
	"encoding/hex"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Ethernal-Tech/apex-bridge/contractbinding"
	ethtxhelper "github.com/Ethernal-Tech/apex-bridge/eth/txhelper"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//go:embed testfiles/*
var content embed.FS

func TestNexus(t *testing.T) {
	const (
		solFileName     = "TestCustom.sol"
		consoleFileName = "console.sol"
		testWait        = time.Second * 3
		gasLimit        = 2_300_000
		feeMultiplier   = 170
	)

	type Test struct {
		method      string
		callSuccess bool
		result      bool
	}

	tests := []Test{
		{method: "valid", callSuccess: true, result: true},
		{method: "valid", callSuccess: true, result: true},
		{method: "invalid", callSuccess: true, result: false},
		{method: "noQuorum", callSuccess: false, result: false},
		{method: "valid", callSuccess: true, result: true},
		{method: "invalid", callSuccess: true, result: false},
		{method: "valid", callSuccess: true, result: true},
		{method: "valid", callSuccess: true, result: true},
		{method: "noQuorum", callSuccess: false, result: false},
	}

	solBytes, err := content.ReadFile(filepath.Join("testfiles", solFileName))
	require.NoError(t, err)

	consoleBytes, err := content.ReadFile(filepath.Join("testfiles", consoleFileName))
	require.NoError(t, err)

	workingPath, err := os.MkdirTemp("", "TestNexusArtifacts")
	require.NoError(t, err)

	defer os.RemoveAll(workingPath)

	solFilePath := filepath.Join(workingPath, solFileName)
	consoleFilePath := filepath.Join(workingPath, consoleFileName)

	require.NoError(t, os.WriteFile(solFilePath, solBytes, 0770))
	require.NoError(t, os.WriteFile(consoleFilePath, consoleBytes, 0770))

	artifact, _, err := CompileAndLoadContract(solFilePath, solFilePath)
	require.NoError(t, err)

	ctx := context.Background()
	nodeUrl := "https://testnet.af.route3.dev/json-rpc/p2-c"
	// nodeUrl = "http://localhost:8545"
	pk := "4f1fe56b4f1d454bfea42fe20629dc61e99e783c4680773e6f3d663b8e984150"

	txHelper, err := ethtxhelper.NewEThTxHelper(
		ethtxhelper.WithNodeURL(nodeUrl),
		ethtxhelper.WithDynamicTx(true),
		ethtxhelper.WithZeroGasPrice(false),
		ethtxhelper.WithDefaultGasLimit(gasLimit),
		ethtxhelper.WithNonceRetrieveCounterFunc(),
		ethtxhelper.WithGasFeeMultiplier(feeMultiplier))
	require.NoError(t, err)

	wallet, err := ethtxhelper.NewEthTxWallet(pk)
	require.NoError(t, err)

	eth := NewEthContractUtils(txHelper, wallet, 1.1)

	addrEth := deploy(t, ctx, txHelper, wallet, artifact)

	for _, test := range tests {
		fmt.Println("executing", test.method)

		txHash, err := eth.ExecuteMethod(ctx, artifact, addrEth, test.method)
		require.NoError(t, err)

		receipt, err := txHelper.WaitForReceipt(ctx, txHash, true)
		require.NoError(t, err)
		require.Equal(t, types.ReceiptStatusSuccessful, receipt.Status)

		event := artifact.Abi.Events["CustomResult"]
		found := false
		mp := map[string]interface{}{}

		for _, logData := range receipt.Logs {
			if len(logData.Topics) < 1 || event.ID != logData.Topics[0] {
				continue
			}

			require.NoError(t, event.Inputs.UnpackIntoMap(mp, logData.Data))
			require.Equal(t, test.callSuccess, mp["callSuccess"].(bool))
			require.Equal(t, test.result, mp["result"].(bool))

			found = true
		}

		require.True(t, found)

		select {
		case <-ctx.Done():
			return
		case <-time.After(testWait):
		}
	}
}

func TestNexusPrecompile(t *testing.T) {
	const (
		solFileName          = "PrecompileTest.sol"
		consoleFileName      = "console.sol"
		interfaceSolFileName = "IPrecompileValidators.sol"
		structsSolFileName   = "IPrecompileStructs.sol"
		testWait             = time.Second * 3
		gasLimit             = 2_300_000
		feeMultiplier        = 170
	)

	solBytes, err := content.ReadFile(filepath.Join("testfiles", solFileName))
	require.NoError(t, err)

	consoleBytes, err := content.ReadFile(filepath.Join("testfiles", consoleFileName))
	require.NoError(t, err)

	interfaceBytes, err := content.ReadFile(filepath.Join("testfiles", interfaceSolFileName))
	require.NoError(t, err)

	structsBytes, err := content.ReadFile(filepath.Join("testfiles", structsSolFileName))
	require.NoError(t, err)

	workingPath, err := os.MkdirTemp("", "TestNexusArtifacts")
	require.NoError(t, err)

	defer os.RemoveAll(workingPath)

	solFilePath := filepath.Join(workingPath, solFileName)
	consoleFilePath := filepath.Join(workingPath, consoleFileName)
	interfaceFilePath := filepath.Join(workingPath, interfaceSolFileName)
	structsFilePath := filepath.Join(workingPath, structsSolFileName)

	require.NoError(t, os.WriteFile(solFilePath, solBytes, 0770))
	require.NoError(t, os.WriteFile(consoleFilePath, consoleBytes, 0770))
	require.NoError(t, os.WriteFile(interfaceFilePath, interfaceBytes, 0770))
	require.NoError(t, os.WriteFile(structsFilePath, structsBytes, 0770))

	artifact, _, err := CompileAndLoadContract(solFilePath, solFilePath)
	require.NoError(t, err)

	ctx := context.Background()
	nodeUrl := "https://testnet.af.route3.dev/json-rpc/p2-c"
	nodeUrl = "http://localhost:8545"
	pk := "4f1fe56b4f1d454bfea42fe20629dc61e99e783c4680773e6f3d663b8e984150"

	txHelper, err := ethtxhelper.NewEThTxHelper(
		ethtxhelper.WithNodeURL(nodeUrl),
		ethtxhelper.WithDynamicTx(true),
		ethtxhelper.WithZeroGasPrice(false),
		ethtxhelper.WithDefaultGasLimit(gasLimit),
		ethtxhelper.WithNonceRetrieveCounterFunc(),
		ethtxhelper.WithGasFeeMultiplier(feeMultiplier))
	require.NoError(t, err)

	wallet, err := ethtxhelper.NewEthTxWallet(pk)
	require.NoError(t, err)

	eth := NewEthContractUtils(txHelper, wallet, 1.5)

	// addrEth := common.HexToAddress("0xD00B005053ef3dF3FC8f5DD89e479410AcA2b595")
	addrEth := deploy(t, ctx, txHelper, wallet, artifact)

	signature, _ := hex.DecodeString("0e5c559fa8d70287f2d78008f9ac39fa514bea03fe5e6dd38540cba57e799aa32ca15dbb778ca5213e4b891d9a24e7ba77ecd6d3169b6df3368760b4254c4316")
	bitmap := big.NewInt(29)
	rawTx, _ := hex.DecodeString("0000000000000000000000000000000000000000000000000000000000000020000000000000000000000000000000000000000000000000000000000000008600000000000000000000000000000000000000000000000000000000000375820000000000000000000000000000000000000000000000000f43fc2c04ee0000000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000000010000000000000000000000004bc4892f8b01b9afc99bcb827c39646ee78bcf060000000000000000000000000000000000000000000000000de0b6b3a7640000")

	validatorsChainData, err := getValidatorsChainData("", "")
	require.NoError(t, err)

	txHash, err := eth.ExecuteMethod(ctx, artifact, addrEth, "deposit", signature, bitmap, rawTx, validatorsChainData)
	require.NoError(t, err)

	receipt, err := txHelper.WaitForReceipt(ctx, txHash, true)
	require.NoError(t, err)
	require.Equal(t, types.ReceiptStatusSuccessful, receipt.Status)
	for _, x := range receipt.Logs {
		if x.Topics[0] == artifact.Abi.Events["CustomResult"].ID {
			dt, err := artifact.Abi.Events["CustomResult"].Inputs.Unpack(x.Data)
			require.NoError(t, err)
			assert.True(t, dt[0].(bool))
			assert.True(t, dt[1].(bool))

			break
		}
	}
}

func deploy(
	t *testing.T, ctx context.Context,
	txHelper ethtxhelper.IEthTxHelper, wallet ethtxhelper.IEthTxWallet,
	artifact *Artifact,
) (addrResult common.Address) {
	addr, hash, err := txHelper.Deploy(ctx, wallet, bind.TransactOpts{}, *artifact.Abi, artifact.Bytecode)
	require.NoError(t, err)
	require.NotEqual(t, common.Address{}, addr)

	addrResult = common.HexToAddress(addr)

	receipt, err := txHelper.WaitForReceipt(ctx, hash, true)
	require.NoError(t, err)
	require.Equal(t, types.ReceiptStatusSuccessful, receipt.Status)
	require.Equal(t, addrResult, receipt.ContractAddress)

	return addrResult
}

func getValidatorsChainData(nodeURL, addr string) ([]contractbinding.IGatewayStructsValidatorChainData, error) {
	if nodeURL == "" {
		nodeURL = "https://testnet.af.route3.dev/json-rpc/p2-c"
	}

	if addr == "" {
		addr = "0xfefBD392E59DFac2C30cd9c1dB89f4798348EA69"
	}

	client, err := ethclient.Dial(nodeURL)
	if err != nil {
		return nil, err
	}

	c, err := contractbinding.NewValidators(
		common.HexToAddress(addr), client)
	if err != nil {
		return nil, err
	}

	return c.GetValidatorsChainData(&bind.CallOpts{})
}
