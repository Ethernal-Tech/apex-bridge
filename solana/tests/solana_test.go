package client

import (
	"context"
	"testing"

	"github.com/Ethernal-Tech/apex-bridge/solana/client"
	testvalidator "github.com/Ethernal-Tech/apex-bridge/solana/tests/test_validator"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/gagliardetto/solana-go/rpc/ws"
	"github.com/stretchr/testify/require"
)

func TestSolana(t *testing.T) {
	cli := rpc.New(rpc.LocalNet_RPC)
	testValidator := testvalidator.NewTestValidator(cli)
	require.NoError(t, testValidator.StartTestNode())
	require.NoError(t, testValidator.WaitForNode())

	defer testValidator.Close()

	wsCli, err := ws.Connect(context.Background(), rpc.LocalNet_WS)
	require.NoError(t, err)

	solanaClient := client.NewSolanaClient(cli, wsCli)

	feePayer, err := solana.PrivateKeyFromSolanaKeygenFile("./client/test.json")
	require.NoError(t, err)

	require.NoError(t, solanaClient.Airdrop(feePayer.PublicKey(), 10_000_000_000))

}
