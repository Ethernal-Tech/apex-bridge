package ethcontracts

import (
	"context"
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	ethtxhelper "github.com/Ethernal-Tech/apex-bridge/eth/txhelper"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
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
	nodeUrl := "https://testnet.af.route3.dev/json-rpc/nexus-p2-c"
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
