package main

import (
	"encoding/hex"
	"fmt"
	"math"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
	db "github.com/Ethernal-Tech/cardano-infrastructure/indexer/db"
	"github.com/Ethernal-Tech/cardano-infrastructure/logger"
	"github.com/hashicorp/go-hclog"
)

func main() {
	networkMagic := uint32(42)
	address := "localhost:3000"   // "/tmp/cardano-133064331/node-spo1/node.sock"
	startBlockHash := []byte(nil) // from genesis
	startSlot := uint64(0)
	startBlockNum := uint64(math.MaxUint64)
	addressesOfInterest := []string{}

	// for test net
	address = "preprod-node.play.dev.cardano.org:3001"
	networkMagic = 1

	// for main net
	address = "backbone.cardano-mainnet.iohk.io:3001"
	networkMagic = uint32(764824073)

	startBlockHash, _ = hex.DecodeString("5d9435abf2a829142aaae08720afa05980efaa6ad58e47ebd4cffadc2f3c45d8")
	startSlot = uint64(76592549)
	startBlockNum = 7999980
	addressesOfInterest = []string{
		"addr1v9kganeshgdqyhwnyn9stxxgl7r4y2ejfyqjn88n7ncapvs4sugsd",
	}

	logger, err := logger.NewLogger(logger.LoggerConfig{
		LogLevel:      hclog.Debug,
		JSONLogFormat: false,
		AppendFile:    true,
		LogFilePath:   "logs/cardano_indexer",
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	dbs, err := db.NewDatabaseInit("", "burek.db")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		logger.Error("Open database failed", "err", err)
		os.Exit(1)
	}

	defer dbs.Close()

	confirmedHandler := func(txs []*indexer.Tx) error {
		logger.Info("Confirmed txs", "len", len(txs))

		confirmedTxs, err := dbs.GetUnprocessedConfirmedTxs(0)
		if err != nil {
			return err
		}

		for _, b := range confirmedTxs {
			logger.Info("Tx has been processed", "tx", b)
		}

		return dbs.MarkConfirmedTxsProcessed(confirmedTxs)
	}

	indexerConfig := &indexer.BlockIndexerConfig{
		StartingBlockPoint: &indexer.BlockPoint{
			BlockSlot:   startSlot,
			BlockHash:   startBlockHash,
			BlockNumber: startBlockNum,
		},
		AddressCheck:           indexer.AddressCheckAll,
		ConfirmationBlockCount: 10,
		AddressesOfInterest:    addressesOfInterest,
		SoftDeleteUtxo:         true,
	}
	syncerConfig := &indexer.BlockSyncerConfig{
		NetworkMagic:   networkMagic,
		NodeAddress:    address,
		RestartOnError: true,
		RestartDelay:   time.Second * 2,
		KeepAlive:      true,
	}

	blockIndexer := indexer.NewBlockIndexer(indexerConfig, confirmedHandler, dbs, logger.Named("block_indexer"))

	blockSyncer := indexer.NewBlockSyncer(syncerConfig, blockIndexer, logger.Named("block_syncer"))
	defer blockSyncer.Close()

	err = blockSyncer.Sync()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		logger.Error("Start syncing failed", "err", err)
		os.Exit(1)
	}

	signalChannel := make(chan os.Signal, 1)
	// Notify the signalChannel when the interrupt signal is received (Ctrl+C)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)

	select {
	case <-signalChannel:
	case <-blockSyncer.ErrorCh():
	}
}
