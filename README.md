# Apex Bridge componentes written in Go

# How to go get private repo
```shell
$ git config url."git@github.com:Ethernal-Tech/cardano-infrastructure.git".insteadOf "https://github.com/Ethernal-Tech/cardano-infrastructure"
$ GOPRIVATE=github.com/Ethernal-Tech/cardano-infrastructure go get github.com/Ethernal-Tech/cardano-infrastructure
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

# How to generate key for blade admin
```shell
$ go run ./main.go wallet-create blade --type admin --key KEY --config CONFIG_PATTH
```

# How to generate key for blade proxy admin
```shell
$ go run ./main.go wallet-create blade --type proxy --key KEY --config CONFIG_PATTH
```

# How to register chain for validator
```shell
$ go run ./main.go register-chain \
        --chain prime \
        --type 0 \
        --validator-data-dir /home/bbs/blade \
        --token-supply 20000000 \
        --bridge-url https://polygon-mumbai.blockpi.network/v1/rpc/public \
        --bridge-addr 0x8F371EeFe210ad18a2Ce45d51B48E56aBa1a58A9        
```
- instead of `--validator-data-dir` it is possible to set blade configuration file with `--validator-config /path/config.json`.

# How to create multisig address
```shell
$ go run ./main.go create-address \
        --network-id network_ID \
        --testnet-magic 3311 \
        --bridge-url http://127.0.0.1:12013 \
        --bridge-addr 0xABEF000000000000000000000000000000000000 \
        --bridge-key BRIDGE_ADMIN_PRIVATE_KEY \
        --chain prime
```
- optional `--show-policy-script` flag
- instead of `--bridge-key` it is possible to set key secret manager configuration file with `--key-config /path/config.json`.

# How to generate config files
All options
``` shell
$ go run ./main.go generate-configs \
        --validator-data-dir <path to bridge chain data directory when using local secrets manager> \
        --validator-config <path to bridge chain secrets manager config file> \
        --output-dir <path to config jsons output directory> \
        --output-validator-components-file-name <validator components config json output file name>.json \
        --output-relayer-file-name <relayer config json output file name>.json \
        --bridge-node-url <node URL of bridge chain> \
        --bridge-sc-address <bridging smart contract address on bridge chain> \
        --relayer-data-dir <relayer data dir for secrets> \
        --relayer-config <relayer secrets config file path> \
        --dbs-path <path to where databases will be stored> \
        --logs-path <path to where logs will be stored> \
        --api-port <port at which API should run> \
        --api-keys <api key 1> \
        --api-keys <api key 2> \
        --empty-blocks-threshold <maximum number of empty blocks for blocks submitter to skip>
```
optionally, the --telemetry <prometheusip:port,datadogip:port> flag can be used if telemetry is desired

Minimal example
``` shell
$ go run ./main.go generate-configs \
        --validator-data-dir ./blade-dir \
        --relayer-data-dir ./blade-dir \
        --bridge-node-url https://polygon-mumbai-pokt.nodies.app \
        --bridge-sc-address 0x816402271eE6D9078Fc8Cb537aDBDD58219485BB \
        --api-keys test_api_key_1
```

Cardano chain all options
``` shell
$ apex-bridge generate-configs cardano-chain \
        --network-address <address of the chain network> \
        --network-magic <network magic of the chain network> \
        --network-id <network id of the chain network> \
        --ogmios-url <ogmios URL for the chain network> \
        --blockfrost-url <blockfrost URL for the chain network> \
        --blockfrost-api-key <blockfrost API key> \
        --socket-path <socket path for the chain network> \
        --ttl-slot-inc <ttl slot increment> \
        --slot-rounding-threshold <slot rounding threshold> \
        --starting-block <slot:hash> \
        --utxo-min-amount <minimal UTXO value for the chain> \
        --min-fee-for-bridging <minimal bridging fee> \
        --block-confirmation-count <block confirmation count> \
        --chain-id <chain id> \
        --allowed-directions <allowed bridging directions> \
        --output-validator-components-file-name <validator components config json output file name> \
        --output-relayer-file-name <relayer config json output file name>
```

Add cardano chain config minimal example
``` shell
$ apex-bridge generate-configs cardano-chain \
        --network-address localhost:13001 \
        --network-magic 142 \
        --network-id 3 \
        --ogmios-url https://prime.ogmios.com \
        --ttl-slot-inc 6 \
        --min-fee-for-bridging 10000000 \
        --block-confirmation-count 10 \
        --chain-id "prime" \
        --allowed-directions "vector" \
        --output-validator-components-file-name "vc_config.json"
```

Add evm chain config minimal example
``` shell
$ apex-bridge generate-configs evm-chain \
        --evm-node-url <node URL> \
        --evm-ttl-block-inc <ttl block increment> \
        --evm-block-rounding-threshold <block rounding threshold> \
        --evm-starting-block <block number> \
        --evm-min-fee-for-bridging <minimal bridging fee> \
        --evm-relayer-gas-fee-multiplier <gas fee multiplier for evm relayer> \
        --chain-id <evm chain id> \
        --allowed-directions <allowed bridging directions> \
        --allowed-directions <allowed bridging directions> \
        --output-validator-components-file-name <validator components config json output file name> \
        --output-relayer-file-name <relayer config json output file name>
```

Add evm chain config minimal example
``` shell
$ apex-bridge generate-configs evm-chain \
        --evm-node-url localhost:5500 \
        --evm-ttl-block-inc 10 \
        --evm-block-rounding-threshold 5 \
        --evm-starting-block 3 \
        --evm-min-fee-for-bridging 10000000 \
        --evm-relayer-gas-fee-multiplier 4 \
        --chain-id "nexus" \
        --allowed-directions "prime" \
        --allowed-directions "vector" \
        --output-validator-components-file-name "vc_config.json"
```

# Example of sending a transaction from the prime to the vector
```shell
apex-bridge sendtx \
        --key PRIME_WALLET_PRIVATE_KEY \
        --testnet-src 3311 \
        --addr-multisig-src addr_test1wrz24vv4tvfqsywkxn36rv5zagys2d7euafcgt50gmpgqpq4ju9uv \
        --ogmios-src http://ogmios.prime.testnet.apexfusion.org:1337 \
        --ogmios-dst http://ogmios.vector.testnet.apexfusion.org:1337 \
        --chain-dst vector \
        --receiver addr_test1v25acu09yv4z2jc026ss5hhgfu5nunfp9z7gkamae43t6fc8gx3pf:1_000_000 \
        --fee 1_100_000
```
- there is an optional `--stake-key` flag

# Example of sending a transaction from the vector to the prime
```shell
apex-bridge sendtx \
        --key VECTOR_WALLET_PRIVATE_KEY \
        --testnet-src 1127 \
        --addr-multisig-src vector_test1w2h482rf4gf44ek0rekamxksulazkr64yf2fhmm7f5gxjpsdm4zsg \
        --ogmios-src http://ogmios.vector.testnet.apexfusion.org:1337 \
        --ogmios-dst http://ogmios.prime.testnet.apexfusion.org:1337 \
        --chain-dst prime \
        --receiver addr_test1vrlt3wnp3hxermfyhfp2x9lu5u32275lf0yh3nvxkpjv7qgxl9f8y:1_234_567 \
        --fee 1_100_000 \
        --network-id-src 2
```
- there is an optional `--stake-key` flag

# Example of sending a transaction from the prime to the nexus
```shell
apex-bridge sendtx \
        --key PRIME_WALLET_PRIVATE_KEY \
        --ogmios-src http://ogmios.prime.testnet.apexfusion.org:1337 \
        --addr-multisig-src addr_test1wrz24vv4tvfqsywkxn36rv5zagys2d7euafcgt50gmpgqpq4ju9uv \
        --testnet-src 3311 \
        --chain-dst nexus \
        --receiver 0x4BC4892F8B01B9aFc99BCB827c39646EE78bCF06:1_000_000 \
        --fee 1_100_000 \
        --nexus-url https://testnet.af.route3.dev/json-rpc/p2-c
```
- there is an optional `--stake-key` flag

# Example of sending a transaction from the nexus to the prime
```shell
apex-bridge sendtx \
        --tx-type evm \
        --key NEXUS_WALLET_PRIVATE_KEY \
        --nexus-url https://testnet.af.route3.dev/json-rpc/p2-c \
        --gateway-addr GATEWAY_PROXY_ADDRESS \
        --chain-dst prime \
        --receiver addr_test1vrlt3wnp3hxermfyhfp2x9lu5u32275lf0yh3nvxkpjv7qgxl9f8y:1000000000000000000 \
        --fee 1000010000000000000 \
        --ogmios-dst http://ogmios.prime.testnet.apexfusion.org:1337
```
- there is an optional `--stake-key` flag

# How to Deploy Nexus Smart Contracts
Default example (bls keys are retrieved from bridge and gateway address is updated on the bridge):
```shell
$ apex-bridge deploy-evm \
        --url http://127.0.0.1:12001 \
        --key NEXUS_OR_EVM_PRIVATE_KEY \
        --dir /tmp \
        --clone \
        --bridge-url http://127.0.0.1:12013 \
        --bridge-addr 0xABEF000000000000000000000000000000000000 \
        --bridge-key BRIDGE_ADMIN_WALLET_PRIVATE_KEY \
```
- instead of `--key` and `--bridge-key` it is possible to set key secret manager configuration file with `--key-config /path/config.json`.
- `--key` for bridge SC is the key of `ProxyContractsAdmin`, and for nexus is the key of owner/initial deployer
-- `BRIDGE_ADMIN_WALLET_PRIVATE_KEY` is the wallet used with `--blade-admin` when starting blade

Example with explicit bls keys:
```shell
$ apex-bridge deploy-evm \
        --url http://127.0.0.1:12001 \
        --key 1841ffaeb5015fa5547e42a2524214e9b55deda3cc26676ff9823bca98b25c94 \
        --dir /tmp \
        --clone \
        --bls-key 0x.... \
        --bls-key 0x.... \
        --bls-key 0x.... \
        --bls-key 0x.... \        
```
- optional `--min-fee`, min-fee value can be specified for the Gateway contract
- optional `--min-bridging-amount` - for the Gateway contract, new min-bridging-amount can be defined
- instead of `--key` it is possible to set key secret manager configuration file with `--key-config /path/config.json`.

# How to upgrade bridge/gateway contracts
```shell
$ apex-bridge deploy-evm upgrade \
        --url http://127.0.0.1:12001 \
        --key NEXUS_OR_EVM_PRIVATE_KEY \
        --dir /tmp \
        --clone \
        --branch main \
        --repo https://github.com/Ethernal-Tech/apex-bridge-smartcontracts \
        --contract Admin:0xABEF000000000000000000000000000000000006
```
- optional `--dynamic-tx`
- `--key` for bridge SC is the key of `ProxyContractsAdmin`, and for nexus is the key of owner/initial deployer
- instead of `--key` it is possible to set key secret manager configuration file with `--key-config /path/config.json`.

# How to Set validators data on Nexus Smart Contracts
Default example (bls keys are retrieved from bridge):
```shell
$ apex-bridge deploy-evm set-validators-chain-data \
        --url http://127.0.0.1:12001 \
        --key NEXUS_OR_EVM_PRIVATE_KEY \
        --dir /tmp \
        --clone \
        --bridge-url http://127.0.0.1:12013 \
        --bridge-addr 0xABEF000000000000000000000000000000000000 \
        --validators-proxy-addr 0x157E8D7DA7A2282aDe8678390A4ad6ba83B0FD9E \
```
- `--key` for bridge SC is the key of `ProxyContractsAdmin`, and for nexus is the key of owner/initial deployer
- instead of `--key` it is possible to set key secret manager configuration file with `--key-config /path/config.json`.

Example with explicit bls keys:
```shell
$ apex-bridge deploy-evm set-validators-chain-data \
        --url http://127.0.0.1:12001 \
        --key 1841ffaeb5015fa5547e42a2524214e9b55deda3cc26676ff9823bca98b25c94 \
        --dir /tmp \
        --clone \
        --bls-key 0x.... \
        --bls-key 0x.... \
        --bls-key 0x.... \
        --bls-key 0x.... \
        --validators-proxy-addr 0x157E8D7DA7A2282aDe8678390A4ad6ba83B0FD9E \
```
- `--key` for bridge SC is the key of `ProxyContractsAdmin`, and for nexus is the key of owner/initial deployer
- instead of `--key` it is possible to set key secret manager configuration file with `--key-config /path/config.json`.

# Bridge admin commands
```shell
$ apex-bridge bridge-admin get-chain-token-quantity \
        --bridge-url http://localhost:12013 \
        --chain prime --chain nexus --chain vector
```

```shell
$ apex-bridge bridge-admin update-chain-token-quantity \
        --bridge-url http://localhost:12013 \
        --chain nexus --amount 300 \
        --key 922769e22b70614d4172fc899126785841f4de7d7c009fc338923ce50683023d
```
- instead of `--key` it is possible to set key secret manager configuration file with `--key-config /path/config.json`.

```shell
$ apex-bridge bridge-admin set-min-amounts \
        --url http://127.0.0.1:12001 \
        --key 922769e22b70614d4172fc899126785841f4de7d7c009fc338923ce50683023d \
        --contract-addr 0xeefcd00000000000000000000000000000000013 \
        --min-fee 200 \
        --min-bridging-amount 100 
```
- instead of `--key` it is possible to set key secret manager configuration file with `--key-config /path/config.json`.


```shell
$ apex-bridge bridge-admin defund \
        --bridge-url http://localhost:12013 \
        --chain nexus \
        --key 922769e22b70614d4172fc899126785841f4de7d7c009fc338923ce50683023d \
        --amount 100 \
        --addr 0xeefcd00000000000000000000000000000000000
```
- there is optional `--native-token-amount` flag
- instead of `--key` it is possible to set key secret manager configuration file with `--key-config /path/config.json`.

```shell
$ apex-bridge bridge-admin set-additional-data \
        --bridge-url http://localhost:12013 \
        --bridge-key 922769e22b70614d4172fc899126785841f4de7d7c009fc338923ce50683023d \
        --chain nexus \
        --bridging-addr 0xeefcd00000000000000000000000000000000022 \
        --fee-addr 0xeefcd00000000000000000000000000000000021
```
- instead of `--bridge-key` it is possible to set key secret manager configuration file with `--key-config /path/config.json`.

```shell
$ apex-bridge bridge-admin get-validators-data \
        --bridge-url http://localhost:12013 \
        --bridge-addr 0xeefcd00000000000000000000000000000000022 \
        --config ./config.json
```

```shell
$ apex-bridge bridge-admin get-bridging-addresses-balances \
        --config ./config.json \
        --indexer-dbs-path /e2e-bridge-data-tmp-Test_OnlyRunApexBridge_WithNexusAndVector/validator_1/bridging-dbs/validatorcomponents \
        --prime-wallet-addr addr_test1wrapsqy073nhdx7tz4j54q4aanhzqqgfpydftysvqyqw50cgz9hpl \
        --vector-wallet-addr addr_test1wffkxzsjpdnkn4vzk7v8wgygcqvztn8ndmte8294rp2l2uqgnp993 \
        --nexus-wallet-addr 0x2ac7dEB534901E63FBd5CEC49929B8830F3FaFF4 \
```

# How to get bridge and gateway smart contract version
```shell
apex-bridge sc-version \
        --node-url http://127.0.0.1:12013 \
        --addr 0xaBef000000000000000000000000000000000000:Bridge \
        --addr 0xaBef000000000000000000000000000000000001:ClaimsHelper \
        --addr 0xaBef000000000000000000000000000000000002:Claims \
        --addr 0xaBef000000000000000000000000000000000003:SignedBatches \
        --addr 0xaBef000000000000000000000000000000000004:Slots \
        --addr 0xaBef000000000000000000000000000000000005:Validators \
        --addr 0xaBef000000000000000000000000000000000006:Admin \
```
