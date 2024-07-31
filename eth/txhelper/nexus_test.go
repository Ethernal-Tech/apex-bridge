package ethtxhelper

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/Ethernal-Tech/apex-bridge/contractbinding"
	eventTrackerStore "github.com/Ethernal-Tech/blockchain-event-tracker/store"
	eventTracker "github.com/Ethernal-Tech/blockchain-event-tracker/tracker"
	"github.com/Ethernal-Tech/ethgo"

	"github.com/hashicorp/go-hclog"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/require"
)

const (
	testNodeURL          = "http://127.0.0.1:12013"
	dummyMumbaiTestAccPk = "61deed8dda92a396e8e9dbcbb5a058bee274de1adc57b2067975691dacdd55c7"
)

var (
	scBytecodeTest, _ = hex.DecodeString("608060405234801561001057600080fd5b5061068d806100206000396000f3fe608060405234801561001057600080fd5b50600436106100365760003560e01c806398b1e06a1461003b578063fa398db814610057575b600080fd5b61005560048036038101906100509190610165565b610073565b005b610071600480360381019061006c91906102d5565b6100b0565b005b7f7adcde22575d10ee3d4e78ee24cc9f854ecc4ce2bc5fda5eadeb754384227db082826040516100a49291906103bb565b60405180910390a15050565b7f2b846d03da343b397a350d2e88aa5091d29b87dd95204dc125870a82860416c885858585856040516100e7959493929190610609565b60405180910390a15050505050565b600080fd5b600080fd5b600080fd5b600080fd5b600080fd5b60008083601f84011261012557610124610100565b5b8235905067ffffffffffffffff81111561014257610141610105565b5b60208301915083600182028301111561015e5761015d61010a565b5b9250929050565b6000806020838503121561017c5761017b6100f6565b5b600083013567ffffffffffffffff81111561019a576101996100fb565b5b6101a68582860161010f565b92509250509250929050565b600060ff82169050919050565b6101c8816101b2565b81146101d357600080fd5b50565b6000813590506101e5816101bf565b92915050565b600073ffffffffffffffffffffffffffffffffffffffff82169050919050565b6000610216826101eb565b9050919050565b6102268161020b565b811461023157600080fd5b50565b6000813590506102438161021d565b92915050565b60008083601f84011261025f5761025e610100565b5b8235905067ffffffffffffffff81111561027c5761027b610105565b5b6020830191508360208202830111156102985761029761010a565b5b9250929050565b6000819050919050565b6102b28161029f565b81146102bd57600080fd5b50565b6000813590506102cf816102a9565b92915050565b6000806000806000608086880312156102f1576102f06100f6565b5b60006102ff888289016101d6565b955050602061031088828901610234565b945050604086013567ffffffffffffffff811115610331576103306100fb565b5b61033d88828901610249565b93509350506060610350888289016102c0565b9150509295509295909350565b600082825260208201905092915050565b82818337600083830152505050565b6000601f19601f8301169050919050565b600061039a838561035d565b93506103a783858461036e565b6103b08361037d565b840190509392505050565b600060208201905081810360008301526103d681848661038e565b90509392505050565b6103e8816101b2565b82525050565b6103f78161020b565b82525050565b600082825260208201905092915050565b6000819050919050565b600080fd5b600080fd5b600080fd5b6000808335600160200384360303811261044457610443610422565b5b83810192508235915060208301925067ffffffffffffffff82111561046c5761046b610418565b5b6001820236038313156104825761048161041d565b5b509250929050565b600082825260208201905092915050565b60006104a7838561048a565b93506104b483858461036e565b6104bd8361037d565b840190509392505050565b60006104d760208401846102c0565b905092915050565b6104e88161029f565b82525050565b6000604083016105016000840184610427565b858303600087015261051483828461049b565b9250505061052560208401846104c8565b61053260208601826104df565b508091505092915050565b600061054983836104ee565b905092915050565b60008235600160400383360303811261056d5761056c610422565b5b82810191505092915050565b6000602082019050919050565b600061059283856103fd565b9350836020840285016105a48461040e565b8060005b878110156105e85784840389526105bf8284610551565b6105c9858261053d565b94506105d483610579565b925060208a019950506001810190506105a8565b50829750879450505050509392505050565b6106038161029f565b82525050565b600060808201905061061e60008301886103df565b61062b60208301876103ee565b818103604083015261063e818587610586565b905061064d60608301846105fa565b969550505050505056fea264697066735822122035cdd7353467dc05bfecadfb19774faf60c75d4612e6af0cb81255f4043fdf4b64736f6c63430008180033")
)

type es struct {
	test string
}

func (sub es) AddLog(log *ethgo.Log) error {
	fmt.Println("AddLog new event")
	fmt.Printf("%+v\n", log)

	events, _ := getEventSignatures([]string{"Deposit", "Withdraw"})

	switch log.Topics[0] {
	case events[0]:
		fmt.Println("Deposit")
	case events[1]:
		fmt.Println("Withdraw")
	default:
		fmt.Println("undefined event")
	}

	return nil
}

// This test requires running blade setup so that we can deploy sc and execute txs
func TestNexus(t *testing.T) {
	t.Skip()

	wallet, err := NewEthTxWallet(dummyMumbaiTestAccPk)
	require.NoError(t, err)

	txHelper, err := NewEThTxHelper(
		WithNodeURL(testNodeURL), WithGasFeeMultiplier(150),
		WithZeroGasPrice(false), WithDefaultGasLimit(0))
	require.NoError(t, err)

	ctx, cancelCtx := context.WithTimeout(context.Background(), time.Second*60)
	defer cancelCtx()

	abiData, err := contractbinding.TestGatewayMetaData.GetAbi()
	require.NoError(t, err)

	addr, hash, err := txHelper.Deploy(ctx, wallet, bind.TransactOpts{}, *abiData, scBytecodeTest)
	require.NoError(t, err)
	require.NotEqual(t, common.Address{}, addr)

	receipt, err := txHelper.WaitForReceipt(ctx, hash, true)
	require.NoError(t, err)
	require.Equal(t, types.ReceiptStatusSuccessful, receipt.Status)
	require.Equal(t, common.HexToAddress(addr), receipt.ContractAddress)

	scAddress := addr

	contract, err := contractbinding.NewTestGateway(common.HexToAddress(scAddress), txHelper.GetClient())
	require.NoError(t, err)

	tx, err := txHelper.SendTx(ctx, wallet, bind.TransactOpts{},
		func(txOpts *bind.TransactOpts) (*types.Transaction, error) {
			return contract.Deposit(txOpts, []byte{1, 2, 3, 4, 5})
		})
	require.NoError(t, err)

	receipt, err = txHelper.WaitForReceipt(ctx, tx.Hash().String(), true)

	require.NoError(t, err)
	require.Equal(t, types.ReceiptStatusSuccessful, receipt.Status)

	tx2, err := txHelper.SendTx(ctx, wallet, bind.TransactOpts{},
		func(txOpts *bind.TransactOpts) (*types.Transaction, error) {
			return contract.Withdraw(txOpts, 1, wallet.addr, []contractbinding.TestGatewayReceiverWithdraw{
				{
					Receiver: "",
					Amount:   big.NewInt(121),
				},
			}, big.NewInt(150))
		})
	require.NoError(t, err)

	receipt, err = txHelper.WaitForReceipt(ctx, tx2.Hash().String(), true)

	fmt.Println("Block Hash: ", receipt.BlockHash)
	fmt.Printf("Receipt %+v: \n", receipt)
	fmt.Println("Block Number: ", receipt.BlockNumber)

	require.NoError(t, err)
	require.Equal(t, types.ReceiptStatusSuccessful, receipt.Status)

	trackerConfig := &eventTracker.EventTrackerConfig{
		SyncBatchSize: 10,
		RPCEndpoint:   "http://127.0.0.1:12013",
		PollInterval:  2 * time.Second,
	}

	trackerConfig.Logger = hclog.Default().Named("test logger")
	trackerConfig.NumBlockConfirmations = 10
	// bridgingAddress := "0xABEF000000000000000000000000000000000000"
	scAddress2 := ethgo.HexToAddress(scAddress)

	events := []string{"Deposit", "Withdraw"}

	eventSigs, err := getEventSignatures(events)
	if err != nil {
		fmt.Println("failed to get event signatures", err)
	}

	logFilter := make(map[ethgo.Address][]ethgo.Hash)
	logFilter[scAddress2] = append(logFilter[scAddress2], eventSigs...)

	trackerConfig.LogFilter = logFilter

	sub := es{
		test: "Test subscriber",
	}

	trackerConfig.EventSubscriber = sub

	trackerStore, err := eventTrackerStore.NewBoltDBEventTrackerStore("my.db")
	if err != nil {
		fmt.Println("failed to init event store!!!")

		return
	}

	lastTracked, _ := trackerStore.GetLastProcessedBlock()
	start := uint64(0)

	if lastTracked > start {
		start = lastTracked
	}

	err = trackerStore.InsertLastProcessedBlock(start)
	if err != nil {
		fmt.Println("failed to insert last processed block in tracker store")

		return
	}

	ethTracker, err := eventTracker.NewEventTracker(trackerConfig, trackerStore, start)
	if err != nil {
		fmt.Println("failed to init event tracker")

		return
	}

	if err := ethTracker.Start(); err != nil {
		fmt.Println("failed to start the tracker")

		return
	}
	defer ethTracker.Close()

	<-context.Background().Done()
	fmt.Println("Closing...")
}

func getEventSignatures(events []string) ([]ethgo.Hash, error) {
	abi, err := contractbinding.GatewayMetaData.GetAbi()
	if err != nil {
		return nil, err
	}

	hashes := make([]ethgo.Hash, len(events))
	for i, ev := range events {
		hashes[i] = ethgo.Hash(abi.Events[ev].ID)
	}

	return hashes, nil
}
