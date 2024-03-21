# apex-bridge
Apex Bridge componentes written in Go

# How to go get private repo
`$ git config url."git@github.com:Ethernal-Tech/cardano-infrastructure.git".insteadOf "https://github.com/Ethernal-Tech/cardano-infrastructure"`
`$ GOPRIVATE=github.com/Ethernal-Tech/cardano-infrastructure go get github.com/Ethernal-Tech/cardano-infrastructure`

# Example how to generate go binding for smart contract
`$ solcjs -p --abi contractbinding/dummycontracts/TestContract.sol -o ./contractbinding/contractbuild && abigen --abi ./contractbinding/contractbuild/contractbinding_dummycontracts_TestContract_sol_TestContract.abi --pkg main --type TestContract --out ./contractbinding/TestContract.go --pkg contractbinding`

# Example how to generate keys for batcher
`$ go run cardano-wallet-cli/cmd/main.go init --addr 0x69b6eEAff0A5c5F80a242104B79F4aC5c40E5130 --directory ~/apex-bridge/batcher/keys --network prime --node-url https://polygon-mumbai-pokt.nodies.app --pk 93c91e490bfd3736d17d04f53a10093e9cf2435309f4be1f5751381c8e201d23 && go run cardano-wallet-cli/cmd/main.go init --addr 0x69b6eEAff0A5c5F80a242104B79F4aC5c40E5130 --directory ~/apex-bridge/batcher/keys --network vector --node-url https://polygon-mumbai-pokt.nodies.app --pk 93c91e490bfd3736d17d04f53a10093e9cf2435309f4be1f5751381c8e201d23`