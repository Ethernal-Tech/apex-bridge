# Apex Bridge componentes written in Go

# How to go get private repo
```shell
$ git config url."git@github.com:Ethernal-Tech/cardano-infrastructure.git".insteadOf "https://github.com/Ethernal-Tech/cardano-infrastructure"
$ GOPRIVATE=github.com/Ethernal-Tech/cardano-infrastructure go get -u github.com/Ethernal-Tech/cardano-infrastructure
```

# How to generate go binding for smart contract(s)
Let say we will have all smart contract repositories in directory `/home/igor/development/ethernal/apex-bridge/`
- Clone them:
```shell
   git clone https://github.com/Ethernal-Tech/apex-bridge-smartcontracts/   
```
```shell
   git clone https://github.com/Ethernal-Tech/apex-evm-gateway
```
- Build them:
```shell
cd apex-bridge-smartcontracts && npm i && npx hardhat compile && cd ..
```
```shell
cd apex-evm-gateway && npm i && npx hardhat compile && cd ..
```
- To generate bridge bindings execute:
```shell
BASEPATH=/home/igor/development/ethernal/apex-bridge/apex-bridge-smartcontracts/
solcjs --base-path "${BASEPATH}" --include-path "${BASEPATH}node_modules" -p \
       --abi ${BASEPATH}contracts/Bridge.sol -o ./contractbinding/contractbuild --optimize
abigen --abi ./contractbinding/contractbuild/contracts_Bridge_sol_Bridge.abi --pkg main \
       --type BridgeContract --out ./contractbinding/BridgeContract.go --pkg contractbinding
```
- To generate nexus bindings execute:
```shell
BASEPATH=/home/igor/development/ethernal/apex-bridge/apex-evm-gateway/
solcjs --base-path "${BASEPATH}" --include-path "${BASEPATH}node_modules" -p \
       --abi ${BASEPATH}contracts/Gateway.sol -o ./contractbinding/contractbuild --optimize
abigen --abi ./contractbinding/contractbuild/contracts_Gateway_sol_Gateway.abi --pkg main \
       --type Gateway --out ./contractbinding/GatewayContract.go --pkg contractbinding
```

# How to generate blade secrets
```shell
$ blade secrets init --insecure --data-dir ./blade-secrets
```

# How to generate cardano keys for cardano batcher(s)
```shell
$ go run ./main.go wallet-create --chain prime --validator-data-dir /home/bbs/cardano --show-pk
```
- instead of `--validator-data-dir` it is possible to set blade configuration file with `--validator-config /path/config.json`.
- its possible to use `--type stake` flag if we want wallet which includes stake signing key also

# How to generate bls keys for evm batcher(s)
```shell
$ go run ./main.go wallet-create --chain nexus --validator-data-dir /home/bbs/cardano --show-pk --type evm
```
- instead of `--validator-data-dir` it is possible to set blade configuration file with `--validator-config /path/config.json`.

# How to generate ecdsa keys for evm relayer(s)
```shell
$ go run ./main.go wallet-create --chain nexus --validator-data-dir /home/bbs/cardano --show-pk --type relayer-evm
```
- instead of `--validator-data-dir` it is possible to set blade configuration file with `--validator-config /path/config.json`.

# How to register chain for validator
```shell
$ go run ./main.go register-chain \
        --chain prime \
        --type 0 \
        --validator-data-dir /home/bbs/blade \
        --addr addr_test1wrs0nrc0rvrfl7pxjl8vgqp5xuvt8j4n8a2lu8gef80wxhq4lmleh \
        --addr-fee addr_test1vqjysa7p4mhu0l25qknwznvj0kghtr29ud7zp732ezwtzec0w8g3u \
        --token-supply 20000000 \
        --bridge-url https://polygon-mumbai.blockpi.network/v1/rpc/public \
        --bridge-addr 0x8F371EeFe210ad18a2Ce45d51B48E56aBa1a58A9        
```
- instead of `--validator-data-dir` it is possible to set blade configuration file with `--validator-config /path/config.json`.

# How to create multisig address
```shell
$ go run ./main.go create-address \
        --network-id network_ID \
        --key 582068fc463c29900b00122423c7e6a39469987786314e07a5e7f5eae76a5fe671bf \
        --key 58209a9cefaa636d75dffa3a3a5ab446a191beac92b09ac82da513640e8e35935202
        ...
```

# How to generate config files
All options
``` shell
$ go run ./main.go generate-configs \
        --validator-data-dir "<path to bridge chain data directory when using local secrets manager>" \
        --validator-config-path "<path to to bridge chain secrets manager config file>" \        
        --output-dir "<path to config jsons output directory>" \
        --output-validator-components-file-name "<validator components config json output file name>.json" \
        --output-relayer-file-name "<relayer config json output file name>.json" \
        --prime-network-address "<address of prime network>" \
        --prime-network-id <network id of prime network> \
        --prime-network-magic <network magic of prime network> \
        --prime-ogmios-url "<ogmios URL for prime network>" \
        --prime-blockfrost-url "<blockfrost URL for prime network>" \
        --prime-blockfrost-api-key "<blockfrost API key for prime network>" \
        --prime-socket-path "<socket path for prime network>" \
        --prime-ttl-slot-inc <ttl slot increment for prime> \
        --prime-slot-rounding-threshold <take slot from sc if zero otherwise calculate slot from tip with rounding> \
        --vector-network-address "<address of vector network>" \
        --vector-network-magic <network magic of vector network> \
        --vector-network-id <network id of vector network> \
        --vector-blockfrost-url "<blockfrost URL for vector network>" \
        --vector-ogmios-url "<ogmios URL for vector network>" \
        --vector-blockfrost-api-key "<blockfrost API key for vector network>" \
        --vector-socket-path "<socket path for vector network>" \
        --vector-ttl-slot-inc <ttl slot increment for vector> \
        --vector-slot-rounding-threshold <take slot from sc if zero otherwise calculate slot from tip with rounding> \
        --bridge-node-url "<node URL of bridge chain>" \
        --bridge-sc-address "<bridging smart contract address on bridge chain>" \
        --dbs-path "<path to where databases will be stored>" \
        --logs-path "<path to where logs will be stored>" \
        --api-port <port at which API should run> \
        --api-keys "<api key 1>" \
        --api-keys "<api key 2>"
```
optionally, the --telemetry <prometheusip:port,datadogip:port> flag can be used if telemetry is desired

Minimal example
``` shell
$ go run ./main.go generate-configs \
        --validator-data-dir "./blade-dir" \
        --prime-network-address "localhost:13001" \
        --prime-network-magic 142 \
        --prime-blockfrost-url "https://cardano-preview.blockfrost.io/api/v0" \
        --vector-network-address "localhost:23001" \
        --vector-network-magic 242 \
        --vector-blockfrost-url "https://cardano-preview.blockfrost.io/api/v0" \
        --bridge-node-url "https://polygon-mumbai-pokt.nodies.app" \
        --bridge-sc-address "0x816402271eE6D9078Fc8Cb537aDBDD58219485BB" \
        --api-keys "test_api_key_1"
```

# How to Send a Bridging Transaction from Prime to Vector (and Vice Versa)
```shell
$ go run ./main.go sendtx \
        --key 58201825bce09711e1563fc1702587da6892d1d869894386323bd4378ea5e3d6cba0 \
        --ogmios-src http://localhost:1337 \
        --receiver addr_test1vzkcuz4e9c07hl90gjyf66xr4eutt8wfchafupdzwgs5cyc7996zx:1_000_010 \
        --receiver addr_test1wrvy8aw0trr4a93ujufac0l9jeh43p7a7z74dz8kljx2yxguldndk:2_000_010 \
        --testnet-src 42 \
        --chain-dst vector \
        --addr-multisig-src addr_test1wzk57y7l9q6qxdyrm4a3nlp535w5l8xglg0kvtl8hp9l8rgpj7q2x \
        --fee 1_100_000 \
        --ogmios-dst http://localhost:1338 
```
