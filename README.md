# Apex Bridge componentes written in Go

# How to go get private repo
```shell
$ git config url."git@github.com:Ethernal-Tech/cardano-infrastructure.git".insteadOf "https://github.com/Ethernal-Tech/cardano-infrastructure"
$ GOPRIVATE=github.com/Ethernal-Tech/cardano-infrastructure go get github.com/Ethernal-Tech/cardano-infrastructure
```

# How to generate go binding for smart contract
```shell
$ solcjs -p --abi contractbinding/dummycontracts/TestContract.sol -o ./contractbinding/contractbuild && abigen --abi ./contractbinding/contractbuild/contractbinding_dummycontracts_TestContract_sol_TestContract.abi --pkg main --type TestContract --out ./contractbinding/TestContract.go --pkg contractbinding
```

# How to generate cardano keys for batcher
```shell
$ go run ./cli/cmd/main.go wallet-create --chain prime --dir /home/bbs/cardano --show-pk
```

# How to register chain for validator
```shell
$ go run ./cli/cmd/main.go register-chain \
        --chain prime \
        --dir /home/bbs/cardano \
        --validator-dir /home/bbs/blade \
        --addr addr_test1wrs0nrc0rvrfl7pxjl8vgqp5xuvt8j4n8a2lu8gef80wxhq4lmleh \
        --addr-fee addr_test1vqjysa7p4mhu0l25qknwznvj0kghtr29ud7zp732ezwtzec0w8g3u \
        --token-supply 20000000 \
        --block-frost https://cardano-preview.blockfrost.io/api/v0 \
        --block-frost-id preview7mGSjpyEKb24OxQ4cCxomxZ5axMs5PvE \
        --bridge-url https://polygon-mumbai.blockpi.network/v1/rpc/public \
        --bridge-addr 0x8F371EeFe210ad18a2Ce45d51B48E56aBa1a58A9
```