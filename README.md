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

# How to generate blade secrets
```shell
$ blade secrets init --insecure --data-dir ./blade-secrets
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
- instead of `--validator-dir` it is possible to set blade configuration file with `--validator-config /path/config.json`.
- instead of `--block-frost` cardano cli can be used with these two flags: `--socket-path socket_path` and `--testnet some_number`

Block frost/socket path to local Cardano node is used only for retrieving utxos for `addr` and `addr-fee`

# How to create multisig address
```shell
$ go run ./cli/cmd/main.go create-address \
        --testnet some_number \
        --prefix  some_prefix \
        --key 582068fc463c29900b00122423c7e6a39469987786314e07a5e7f5eae76a5fe671bf \
        --key 58209a9cefaa636d75dffa3a3a5ab446a191beac92b09ac82da513640e8e35935202
        ...
```

# How to generate config files
All options
``` shell
$ go run ./cli/cmd/main.go generate-configs \
        --output-dir "<Path to config jsons output directory>" \
        --output-validator-components-file-name "<Validator components config json output file name>.json" \
        --output-relayer-file-name "<Relayer config json output file name>.json" \
        --prime-network-address "<Address of prime network>" \
        --prime-network-magic <Network magic of prime network> \
        --prime-keys-dir "<Path to cardano keys directory for prime network>" \
        --prime-blockfrost-url "<Blockfrost URL for prime network>" \
        --prime-blockfrost-api-key "<Blockfrost API key for prime network>" \
        --prime-socket-path "<Socket path for prime network>" \
        --vector-network-address "<Address of vector network>" \
        --vector-network-magic <Network magic of vector network> \
        --vector-keys-dir "<Path to cardano keys directory for vector network>" \
        --vector-blockfrost-url "<Blockfrost URL for vector network>" \
        --vector-blockfrost-api-key "<Blockfrost API key for vector network>" \
        --vector-socket-path "<Socket path for vector network>" \
        --bridge-node-url "<Node URL of bridge chain>" \
        --bridge-sc-address "<Bridging smart contract address on bridge chain>" \
        --bridge-validator-data-dir "<Path to bridge chain data directory when using local secrets manager>" \
        --bridge-validator-config-path "<Path to to bridge chain secrets manager config file>" \
        --dbs-path "<Path to where databases will be stored>" \
        --logs-path "<Path to where logs will be stored>" \
        --api-port <Port at which API should run> \
        --api-keys "<api key 1>" \
        --api-keys "<api key 2>"
```

Minimal example
``` shell
$ go run ./cli/cmd/main.go generate-configs \
        --prime-network-address "localhost:13001" \
        --prime-network-magic 142 \
        --prime-blockfrost-url "https://cardano-preview.blockfrost.io/api/v0" \
        --vector-network-address "localhost:23001" \
        --vector-network-magic 242 \
        --vector-blockfrost-url "https://cardano-preview.blockfrost.io/api/v0" \
        --bridge-node-url "https://polygon-mumbai-pokt.nodies.app" \
        --bridge-sc-address "0x816402271eE6D9078Fc8Cb537aDBDD58219485BB" \
        --bridge-validator-data-dir "./blade-dir" \
        --api-keys "test_api_key_1"
```