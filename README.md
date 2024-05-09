# Apex Bridge componentes written in Go

# How to go get private repo
```shell
$ git config url."git@github.com:Ethernal-Tech/cardano-infrastructure.git".insteadOf "https://github.com/Ethernal-Tech/cardano-infrastructure"
$ GOPRIVATE=github.com/Ethernal-Tech/cardano-infrastructure go get github.com/Ethernal-Tech/cardano-infrastructure
```

# How to generate go binding for smart contract
```shell
$ npm install
$ solcjs --base-path "." --include-path "node_modules" -p --abi contracts/Bridge.sol -o ./contractbuild --optimize
abigen --abi ./contractbuild/contracts_Bridge_sol_Bridge.abi --pkg main --type BridgeContract --out ./contractbuild/BridgeContract.go --pkg contractbinding
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
        --keys-dir /home/bbs/cardano \
        --bridge-validator-data-dir /home/bbs/blade \
        --addr addr_test1wrs0nrc0rvrfl7pxjl8vgqp5xuvt8j4n8a2lu8gef80wxhq4lmleh \
        --addr-fee addr_test1vqjysa7p4mhu0l25qknwznvj0kghtr29ud7zp732ezwtzec0w8g3u \
        --token-supply 20000000 \
        --blockfrost https://cardano-preview.blockfrost.io/api/v0 \
        --blockfrost-api-key preview7mGSjpyEKb24OxQ4cCxomxZ5axMs5PvE \
        --bridge-url https://polygon-mumbai.blockpi.network/v1/rpc/public \
        --bridge-addr 0x8F371EeFe210ad18a2Ce45d51B48E56aBa1a58A9        
```
- instead of `--bridge-validator-data-dir` it is possible to set blade configuration file with `--bridge-validator-config /path/config.json`.
- there is possibility to use one of these tx providers:
- blockfrost with `--blockfrost URL` and `--blockfrost-api-key API_KEY` flags
- ogmios with `--ogmios URL` flag
- cardano cli with  `--socket-path SOCKET_PATH` and `--network-magic NUMBER` flags

Block frost/ogmios/socket path to local Cardano node is used only for retrieving utxos for `addr` and `addr-fee`

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
        --output-dir "<path to config jsons output directory>" \
        --output-validator-components-file-name "<validator components config json output file name>.json" \
        --output-relayer-file-name "<relayer config json output file name>.json" \
        --prime-network-address "<address of prime network>" \
        --prime-network-magic <network magic of prime network> \
        --prime-keys-dir "<path to cardano keys directory for prime network>" \
        --prime-ogmios-url "<ogmios URL for prime network>" \
        --prime-blockfrost-url "<blockfrost URL for prime network>" \
        --prime-blockfrost-api-key "<blockfrost API key for prime network>" \
        --prime-socket-path "<socket path for prime network>" \
        --vector-network-address "<address of vector network>" \
        --vector-network-magic <network magic of vector network> \
        --vector-keys-dir "<path to cardano keys directory for vector network>" \
        --vector-blockfrost-url "<blockfrost URL for vector network>" \
        --vector-ogmios-url "<ogmios URL for vector network>" \
        --vector-blockfrost-api-key "<blockfrost API key for vector network>" \
        --vector-socket-path "<socket path for vector network>" \
        --bridge-node-url "<node URL of bridge chain>" \
        --bridge-sc-address "<bridging smart contract address on bridge chain>" \
        --bridge-validator-data-dir "<path to bridge chain data directory when using local secrets manager>" \
        --bridge-validator-config-path "<path to to bridge chain secrets manager config file>" \
        --dbs-path "<path to where databases will be stored>" \
        --logs-path "<path to where logs will be stored>" \
        --api-port <port at which API should run> \
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