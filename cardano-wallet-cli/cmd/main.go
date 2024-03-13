package main

import cardanowalletcli "github.com/Ethernal-Tech/apex-bridge/cardano-wallet-cli"

func main() {
	cardanowalletcli.NewRootCommand().Execute()
}
