# Apex Bridge componentes written in Go

# How to go get private repo
```shell
$ git config url."git@github.com:Ethernal-Tech/cardano-infrastructure.git".insteadOf "https://github.com/Ethernal-Tech/cardano-infrastructure"
$ GOPRIVATE=github.com/Ethernal-Tech/cardano-infrastructure go get -u github.com/Ethernal-Tech/cardano-infrastructure
```

# How to generate go binding for smart contract(s)
- Let's say we will place smart contract repositories in the directory `/home/igor/development/ethernal/apex-bridge/`
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
- Generate bridge bindings with the command:
```shell
BASEPATH=/home/igor/development/ethernal/apex-bridge/apex-bridge-smartcontracts/
solcjs --base-path "${BASEPATH}" --include-path "${BASEPATH}node_modules" -p \
       --abi ${BASEPATH}contracts/Bridge.sol -o ./contractbinding/contractbuild --optimize
abigen --abi ./contractbinding/contractbuild/contracts_Bridge_sol_Bridge.abi --pkg main \
       --type BridgeContract --out ./contractbinding/BridgeContract.go --pkg contractbinding
```
- Generate nexus bindings with the command:
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
- instead of using `--validator-data-dir`, it is possible to set the blade configuration file with `--validator-config /path_to_config/config.json`
- It's possible to use the `--type stake` flag if we want a wallet that includes the stake signing key as well

# How to generate bls keys for evm batcher(s)
```shell
$ go run ./main.go wallet-create --chain nexus --validator-data-dir /home/bbs/cardano --type evm --show-pk
```
- instead of using `--validator-data-dir`, it is possible to set the blade configuration file with 
`--validator-config path_to_config/config.json`

# How to generate ecdsa keys for evm relayer(s)
```shell
$ go run ./main.go wallet-create --chain nexus --validator-data-dir /home/bbs/cardano --type relayer-evm --show-pk
```
- instead of using `--validator-data-dir`, it is possible to set the blade configuration file with 
`--validator-config path_to_config/config.json`

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
        --validator-data-dir <path to bridge chain data directory when using local secrets manager> \
        --validator-config <path to bridge chain secrets manager config file> \        
        --output-dir <path to config jsons output directory> \
        --output-validator-components-file-name <validator components config json output file name>.json \
        --output-relayer-file-name <relayer config json output file name>.json \
        --prime-network-address <address of prime network> \
        --prime-network-id <network id of prime network> \
        --prime-network-magic <network magic of prime network> \
        --prime-ogmios-url <ogmios URL for prime network> \
        --prime-blockfrost-url <blockfrost URL for prime network> \
        --prime-blockfrost-api-key <blockfrost API key for prime network> \
        --prime-socket-path <socket path for prime network> \
        --prime-ttl-slot-inc <ttl slot increment for prime> \
        --prime-slot-rounding-threshold <prime slot rounding threshold> \
        --prime-starting-block <slot:hash> \
        --vector-network-address <address of vector network> \
        --vector-network-magic <network magic of vector network> \
        --vector-network-id <network id of vector network> \
        --vector-blockfrost-url <blockfrost URL for vector network> \
        --vector-ogmios-url <ogmios URL for vector network> \
        --vector-blockfrost-api-key <blockfrost API key for vector network> \
        --vector-socket-path <socket path for vector network> \
        --vector-ttl-slot-inc <ttl slot increment for vector> \
        --vector-slot-rounding-threshold <vector slot rounding threshold> \
        --vector-starting-block <slot:hash> \
        --nexus-node-url <nexus node URL> \
        --nexus-sc-address <nexus smart contract address> \
        --nexus-relayer-addr <nexus relayer address> \
        --nexus-ttl-block-inc <nexus ttl block increment> \
        --nexus-block-rounding-threshold <nexus block rounding threshold> \
        --nexus-starting-block <block number> \
        --bridge-node-url <node URL of bridge chain> \
        --bridge-sc-address <bridging smart contract address on bridge chain> \
        --relayer-data-dir <relayer data dir for secrets> \
        --relayer-config <relayer secrets config file path> \
        --dbs-path <path to where databases will be stored> \
        --logs-path <path to where logs will be stored> \
        --api-port <port at which API should run> \
        --api-keys <api key 1> \
        --api-keys <api key 2>
```
optionally, the --telemetry <prometheusip:port,datadogip:port> flag can be used if telemetry is desired

Minimal example
``` shell
$ go run ./main.go generate-configs \
        --validator-data-dir ./blade-dir \
        --relayer-data-dir ./blade-dir \
        --prime-network-address localhost:13001 \
        --prime-network-magic 142 \
        --prime-blockfrost-url https://cardano-preview.blockfrost.io/api/v0 \
        --vector-network-address localhost:23001 \
        --vector-network-magic 242 \
        --vector-blockfrost-url https://cardano-preview.blockfrost.io/api/v0 \
        --nexus-node-url localhost:5500 \
        --nexus-sc-address 0x816402271eE6DA078Fc8Cb537aDBDD58219485BB \
        --nexus-relayer-addr 0x816402271FE6D9078Fc8Cb537aDBDD58219485BB \
        --bridge-node-url https://polygon-mumbai-pokt.nodies.app \
        --bridge-sc-address 0x816402271eE6D9078Fc8Cb537aDBDD58219485BB \
        --api-keys test_api_key_1
```

# How to Send a Bridging Transaction from Prime to Vector (and Vice Versa)
```shell
$ go run ./main.go sendtx \
        --key 58201825bce09711e1563fc1702587da6892d1d869894386323bd4378ea5e3d6cba0 \
        --ogmios-src http://localhost:1337 \
        --addr-multisig-src addr_test1vzkcuz4e9c07hl90gjyf66xr4eutt8wfchafupdzwgs5cyc7996zx \
        --testnet-src 3311 \
        --chain-dst vector \
        --receiver vector_test1vgxk3ha6hmftgjzrjlrxrndmqrg43y862pu909r87q8kpas0c0mzc:1_000_010 \
        --fee 1_100_000 \
        --ogmios-dst http://localhost:1338 
```

# How to Send a Bridging Transaction from Nexus to Prime
```shell
$ go run ./main.go sendtx \
        --tx-type evm \
        --key a7dc97a721e4ef7503bfc120fdf343030c2070d0cc0db8d1d9384ecf0bcd0aaf \
        --nexus-url http://127.0.0.1:12001 \
        --gateway-addr 0x1BBe094400C854934e46ce53E1f4Dd94f04F9376 \
        --chain-dst prime \
        --receiver addr_test1wq8vxyeq8waqx8cw8p400psmy0x2jd6jt78kffkrs670ctgfnasnk:1000000000000000000 \
        --fee 1000010000000000000 \
        --ogmios-dst http://localhost:1337
```

# How to Send a Bridging Transaction from Prime to Nexus
```shell
$ go run ./main.go sendtx \
        --key 58201825bce09711e1563fc1702587da6892d1d869894386323bd4378ea5e3d6cba0 \
        --ogmios-src http://localhost:1337 \
        --addr-multisig-src addr_test1vzkcuz4e9c07hl90gjyf66xr4eutt8wfchafupdzwgs5cyc7996zx \
        --testnet-src 3311 \
        --chain-dst nexus \
        --receiver 0x71C7656EC7ab88b098defB751B7401B5f6d8976F:1_000_000 \
        --fee 1_100_000 \
        --nexus-url http://127.0.0.1:12001
```

# How to Deploy Nexus Smart Contracts
```shell
$ go run ./main.go deploy-evm \
        --bridge-url http://localhost:12013 \
        -url http://127.0.0.1:12001 \
        --key 1841ffaeb5015fa5547e42a2524214e9b55deda3cc26676ff9823bca98b25c94 \
        --dir /home/igor/development \
        --clone
```