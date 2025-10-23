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
        --wrapped-token-supply 0 \
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
        --prime-utxo-min-amount <minimal UTXO value for prime> \
        --prime-min-fee-for-bridging <minimal bridging fee for prime> \
        --prime-block-confirmation-count <block confirmation count for prime> \
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
        --vector-utxo-min-amount <minimal UTXO value for vector> \
        --vector-min-fee-for-bridging<minimal bridging fee for vector> \
        --vector-block-confirmation-count <block confirmation count for vector> \
        --nexus-node-url <nexus node URL> \
        --nexus-ttl-block-inc <nexus ttl block increment> \
        --nexus-block-rounding-threshold <nexus block rounding threshold> \
        --nexus-starting-block <block number> \
        --nexus-min-fee-for-bridging <minimal bridging fee for nexus> \
        --evm-relayer-gas-fee-multiplier <gas fee multiplier for evm relayer> \
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
        --prime-network-address localhost:13001 \
        --prime-network-magic 142 \
        --prime-ogmios-url https://prime.ogmios.com \
        --vector-network-address localhost:23001 \
        --vector-network-magic 242 \
        --vector-ogmios-url https://vector.ogmios.com \
        --nexus-node-url localhost:5500 \
        --bridge-node-url https://bridge.com \
        --bridge-sc-address 0x816402271eE6D9078Fc8Cb537aDBDD58219485BB \
        --api-keys test_api_key_1
```

Skyline minimal example
``` shell
$ apex-bridge generate-configs skyline \
        --validator-data-dir ./blade-dir \
        --relayer-data-dir ./blade-dir \
        --prime-network-address localhost:13001 \
        --prime-network-magic 142 \
        --prime-ogmios-url https://prime.ogmios.com \
        --vector-network-address localhost:13001 \
        --vector-network-magic 242 \
        --vector-ogmios-url https://vector.ogmios.com \
        --cardano-network-address localhost:23001 \
        --cardano-ogmios-url https://vector.ogmios.com \
        --bridge-node-url https://bridge.com \
        --bridge-sc-address 0x816402271eE6D9078Fc8Cb537aDBDD58219485BB \
        --vector-cardano-token-name 29f8873beb52e126f207a2dfd50f7cff556806b5b4cba9834a7b26a8.4b6173685f546f6b656e \
        --cardano-prime-token-name 29f8873beb52e126f207a2dfd50f7cff556806b5b4cba9834a7b26a8.Route3 \
        --api-keys test_api_key_1
```

# Example of sending a transaction from the prime to the vector
```shell
$ apex-bridge sendtx \
        --key PRIME_WALLET_PRIVATE_KEY \
        --testnet-src 3311 \
        --addr-multisig-src addr_test1wrz24vv4tvfqsywkxn36rv5zagys2d7euafcgt50gmpgqpq4ju9uv \
        --ogmios-src http://ogmios.prime.testnet.apexfusion.org:1337 \
        --ogmios-dst http://ogmios.vector.testnet.apexfusion.org:1337 \
        --chain-src prime \
        --chain-dst vector \
        --receiver addr_test1v25acu09yv4z2jc026ss5hhgfu5nunfp9z7gkamae43t6fc8gx3pf:1_000_000 \
        --fee 1_100_000
```
- there is an optional `--stake-key` flag

# Example of sending a transaction from the vector to the prime
```shell
$ apex-bridge sendtx \
        --key VECTOR_WALLET_PRIVATE_KEY \
        --testnet-src 1127 \
        --addr-multisig-src addr_test1w2h482rf4gf44ek0rekamxksulazkr64yf2fhmm7f5gxjpsdm4zsg \
        --ogmios-src http://ogmios.vector.testnet.apexfusion.org:1337 \
        --ogmios-dst http://ogmios.prime.testnet.apexfusion.org:1337 \
        --chain-src vector \
        --chain-dst prime \
        --receiver addr_test1vrlt3wnp3hxermfyhfp2x9lu5u32275lf0yh3nvxkpjv7qgxl9f8y:1_234_567 \
        --fee 1_100_000 \
        --network-id-src 2
```
- there is an optional `--stake-key` flag

# Example of sending a transaction from the prime to the nexus
```shell
$ apex-bridge sendtx \
        --key PRIME_WALLET_PRIVATE_KEY \
        --ogmios-src http://ogmios.prime.testnet.apexfusion.org:1337 \
        --addr-multisig-src addr_test1wrz24vv4tvfqsywkxn36rv5zagys2d7euafcgt50gmpgqpq4ju9uv \
        --testnet-src 3311 \
        --chain-src prime \
        --chain-dst nexus \
        --receiver 0x4BC4892F8B01B9aFc99BCB827c39646EE78bCF06:1_000_000 \
        --fee 1_100_000 \
        --nexus-url https://testnet.af.route3.dev/json-rpc/p2-c
```
- there is an optional `--stake-key` flag

# Example of sending a transaction from the nexus to the prime
```shell
$ apex-bridge sendtx \
        --tx-type evm \
        --key NEXUS_WALLET_PRIVATE_KEY \
        --nexus-url https://testnet.af.route3.dev/json-rpc/p2-c \
        --gateway-addr GATEWAY_PROXY_ADDRESS \
        --chain-src nexus \
        --chain-dst prime \
        --receiver addr_test1vrlt3wnp3hxermfyhfp2x9lu5u32275lf0yh3nvxkpjv7qgxl9f8y:1000000000000000000 \
        --fee 1000010000000000000 \
        --ogmios-dst http://ogmios.prime.testnet.apexfusion.org:1337
```
- there is an optional `--stake-key` flag

# Example of sending a skyline transaction from the cardano to the prime
```shell
$ apex-bridge sendtx skyline \
        --key CARDANO_WALLET_PRIVATE_KEY \
        --ogmios-src http://ogmios.cardano.testnet.apexfusion.org:1337 \
        --ogmios-dst http://ogmios.prime.testnet.apexfusion.org:1337 \
        --addr-multisig-src addr_test1wrz24vv4tvfqsywkxn36rv5zagys2d7euafcgt50gmpgqpq4ju9uv \
        --testnet-src 3311 \
        --network-id-src 1 \
        --chain-src cardano \
        --chain-dst prime \
        --receiver addr_test1vrlt3wnp3hxermfyhfp2x9lu5u32275lf0yh3nvxkpjv7qgxl9f8y:1_234_567 \
        --fee 1_100_000 \
        --dst-token-name 72f3d1e6c885e4d0bdcf5250513778dbaa851c0b4bfe3ed4e1bcceb0.4b6173685f546f6b656e

```
- optional `--src-token-name` which can be used instead of `--dst-token-name`
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
- `BRIDGE_ADMIN_WALLET_PRIVATE_KEY` is the wallet used with `--blade-admin` when starting blade
- optional `gas-limit` flag if 5_242_880 of gas is not enough for transaction

Example with explicit bls keys:
```shell
$ apex-bridge deploy-evm \
        --url http://127.0.0.1:12001 \
        --key 1841ffaeb5015fa5547e42a2524214e9b55deda3cc26676ff9823bca98b25c94 \
        --dir /tmp \
        --bls-key 0x.... \
        --bls-key 0x.... \
        --bls-key 0x.... \
        --bls-key 0x.... \        
```
- optional `--min-fee`, min-fee value can be specified for the Gateway contract
- optional `--clone` which should be used with `--repo REPOSITORY_URL` and `--branch BRANCH_NAME` flags
- optional `--min-bridging-amount` - for the Gateway contract, new min-bridging-amount can be defined
- optional `gas-limit` flag if 5_242_880 of gas is not enough for transaction
- instead of `--key` it is possible to set key secret manager configuration file with `--key-config /path/config.json`.

# How to upgrade bridge/gateway contracts
```shell
$ apex-bridge deploy-evm upgrade \
        --url http://127.0.0.1:12001 \
        --key NEXUS_OR_EVM_PRIVATE_KEY \
        --dir /tmp \
        --branch main \
        --repo https://github.com/Ethernal-Tech/apex-bridge-smartcontracts \
        --contract Admin:0xABEF000000000000000000000000000000000006
```
- optional `--dynamic-tx`
- optional `--gas-limit` flag if 5_242_880 of gas is not enough for transaction
- optional `--clone` which should be used with `--repo REPOSITORY_URL` and `--branch BRANCH_NAME` flags
- optional `--contract` can contain additional function call `--contract Admin:0xABEF000000000000000000000000000000000006:functionName:fnArg1;fnArg2`
- `--key` for bridge SC is the key of `ProxyContractsAdmin`, and for nexus is the key of owner/initial deployer
- instead of `--key` it is possible to set key secret manager configuration file with `--key-config /path/config.json`.

# How to deploy bridge/gateway contracts
```shell
$ apex-bridge deploy-evm deploy-contract \
        --contract-dir /home/apex-bridge-smartcontracts \
        --contract-name BridgeAddresses \
        --dependencies 0xaBef000000000000000000000000000000000000 \
        --key NEXUS_OR_EVM_PRIVATE_KEY \
        --owner 0xb1995c4276F222cF11Ce78338A0F588579367286 \
        --upgrade-admin 0x7Af59817518702AB59936916c55bB552555BDEa5 \
        --url http://127.0.0.1:12001
```

- optional `--dynamic-tx`
- optional `--clone` which should be used with `--repo REPOSITORY_URL` and `--branch BRANCH_NAME` flags
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
- optional `--is-wrapped-token` bool flag
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
        --native-token-amount 200 \
        --addr 0xeefcd00000000000000000000000000000000000
```
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

To register stake address and delegate it to the stake pool use: 
```shell
$ apex-bridge bridge-admin delegate-address-to-stake-pool \
        --bridge-address-index 0 \
        --bridge-url http://localhost:12001 \
        --chain prime \
        --key 922769e22b70614d4172fc899126785841f4de7d7c009fc338923ce50683023d \
        --stake-pool pool1hvsmu7l9c23ltrncj6lkgmr6ncth7s8tx67zyj2fxl8054xyjz6 \
        --do-registration
```
- instead of `--key` it is possible to set key secret manager configuration file with `--key-config /path/config.json`.

For redelegation of stake address to another stake pool use: 
```shell
$ apex-bridge bridge-admin delegate-address-to-stake-pool \
        --bridge-address-index 0 \
        --bridge-url http://localhost:12001 \
        --chain prime \
        --key 922769e22b70614d4172fc899126785841f4de7d7c009fc338923ce50683023d \
        --stake-pool pool1hvsmu7l9c23ltrncj6lkgmr6ncth7s8tx67zyj2fxl8054xyjz6
```
- instead of `--key` it is possible to set key secret manager configuration file with `--key-config /path/config.json`.

For deregistration of stake address use:
```shell
$ apex-bridge bridge-admin deregister-stake-address \
        --bridge-address-index 0 \
        --bridge-url http://localhost:12001 \
        --chain prime \
        --key 922769e22b70614d4172fc899126785841f4de7d7c009fc338923ce50683023d 
```
- instead of `--key` it is possible to set key secret manager configuration file with `--key-config /path/config.json`.

```shell
$ apex-bridge bridge-admin update-bridging-addrs-count \
        --bridge-url http://localhost:12001 \
        --bridging-addresses-count 5 \
        --chain prime \
        --key 922769e22b70614d4172fc899126785841f4de7d7c009fc338923ce50683023d
```
- instead of `--key` it is possible to set key secret manager configuration file with `--key-config /path/config.json`.

```shell
$ apex-bridge bridge-admin get-validators-data \
        --bridge-url http://localhost:12013 \
        --bridge-addr 0xeefcd00000000000000000000000000000000022 \
        --config ./config.json
```

```shell
$ apex-bridge bridge-admin mint-native-token \
        --key PRIME_WALLET_PRIVATE_KEY \
        --ogmios http://ogmios.prime.testnet.apexfusion.org:1337 \
        --network-id 1 \
        --testnet-magic 3311 \
        --token-name testt \
        --amount 10
```
- optional `--stake-key` and `--show-policy-script` flags
- optional `--validity-slot NUMBER` or `--validity-slot-inc NUMBER` flag. Second one uses ogmios `getTipData`.slot + inc

```shell
$ apex-bridge bridge-admin get-bridging-addresses-balances \
        --config ./config.json \
        --indexer-dbs-path /e2e-bridge-data-tmp-Test_OnlyRunApexBridge_WithNexusAndVector/validator_1/bridging-dbs/validatorcomponents \
        --prime-wallet-addr addr_test1wrapsqy073nhdx7tz4j54q4aanhzqqgfpydftysvqyqw50cgz9hpl \
        --vector-wallet-addr addr_test1wffkxzsjpdnkn4vzk7v8wgygcqvztn8ndmte8294rp2l2uqgnp993 \
        --nexus-wallet-addr 0x2ac7dEB534901E63FBd5CEC49929B8830F3FaFF4
```

```shell
$ apex-bridge bridge-admin get-bridging-addresses-balances skyline \
        --config ./config.json \
        --indexer-dbs-path /e2e-bridge-data-tmp-Test_OnlyRunSkylineBridge/validator_1/bridging-dbs/validatorcomponents \
        --prime-wallet-addr addr_test1wpg8ayttfkr2gvj47p2qkekhrx7w0ecjfdedh6ewrzjhnyg0t7rzg \
        --cardano-wallet-addr addr_test1wrzslpc4stfp78r774k96gxgv4nl2nluc84nv8xkdm0pv7cp4j05f
```

```shell
$ apex-bridge bridge-admin delegate-address-to-stake-pool \
        --bridge-address-index 0 \
        --bridge-url http://localhost:12001 \
        --chain prime \
        --key 922769e22b70614d4172fc899126785841f4de7d7c009fc338923ce50683023d \
        --stake-pool pool1hvsmu7l9c23ltrncj6lkgmr6ncth7s8tx67zyj2fxl8054xyjz6
```
- instead of `--key` it is possible to set key secret manager configuration file with `--key-config /path/config.json`.

```shell
$ apex-bridge bridge-admin redistribute-bridging-addresses-tokens \
        --bridge-url http://localhost:12001 \
        --chain prime \
        --key 922769e22b70614d4172fc899126785841f4de7d7c009fc338923ce50683023d
```
- instead of `--key` it is possible to set key secret manager configuration file with `--key-config /path/config.json`.

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
