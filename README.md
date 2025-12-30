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
solc --base-path "${BASEPATH}" --include-path "${BASEPATH}node_modules" \
       --abi ${BASEPATH}contracts/Bridge.sol -o ./contractbinding/contractbuild \
       --optimize --via-ir --overwrite
abigen --abi ./contractbinding/contractbuild/Bridge.abi --pkg main \
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

# How to generate bls keys for cardano relayer(s)
```shell
$ go run ./main.go wallet-create --chain prime --validator-data-dir /home/bbs/cardano --type relayer-cardano --show-pk
```
- instead of using `--validator-data-dir`, it is possible to set the blade configuration file with 
`--validator-config path_to_config/config.json`

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

# How to generate key cardano relayer
```shell
$ go run ./main.go wallet-create --chain prime --validator-data-dir /home/bbs/cardano --type relayer-cardano --network-id 1
```

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
        --config ./config.json \
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
$ go run ./main.go create-addresses \
        --config ./config.json \
        --network-id network_ID \
        --testnet-magic 3311 \
        --bridge-url http://127.0.0.1:12013 \
        --bridge-addr 0xABEF000000000000000000000000000000000000 \
        --bridge-key BRIDGE_ADMIN_PRIVATE_KEY \
        --generate-custodial-address \
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
        --bridge-node-url https://bridge.com \
        --bridge-sc-address 0x816402271eE6D9078Fc8Cb537aDBDD58219485BB \
        --api-keys test_api_key_1
```

Skyline minimal example
``` shell
$ apex-bridge generate-configs skyline \
        --validator-data-dir ./blade-dir \
        --bridge-node-url https://bridge.com \
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
        --min-operation-fee <minimal operation fee> \
        --chain-id <chain id> \
        --native-token-destination-chain-id <wrapped token destination chain id> \
        --native-token-name <wrapped token full name> \
        --allowed-directions <allowed bridging directions> \
        --output-validator-components-file-name <validator components config json output file name> \
        --output-relayer-file-name <relayer config json output file name> \
        --empty-blocks-threshold <maximum number of empty blocks for blocks submitter to skip> \
        --dbs-path <path to where databases will be stored> \
        --min-fee-for-bridging-tokens <minimal bridging fee for bridging tokens for the chain> \
        --minting-script-tx-input-hash <tx input hash used for referencing minting script> \
	--minting-script-tx-input-index <tx input index used for referencing minting script> \
	--nft-policy-id <the policy ID of the NFT used in the minting script> \
	--nft-name <the name of the NFT used in the minting script> \
	--relayer-address <relayer address for paying collaterals on the chain> \
        --output-dir <path to config jsons output directory> 
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
        --min-operation-fee 1999999 \
        --block-confirmation-count 10 \
        --chain-id "prime" \
        --native-token-destination-chain-id "cardano" \
        --native-token-name "29f8873beb52e126f207a2dfd50f7cff556806b5b4cba9834a7b26a8.4b6173685f546f6b656e" \
        --allowed-directions "cardano" \
        --output-validator-components-file-name "vc_config.json"
```

Evm chain config all options
``` shell
$ apex-bridge generate-configs evm-chain \
        --evm-node-url <node URL> \
        --evm-ttl-block-inc <ttl block increment> \
        --evm-block-rounding-threshold <block rounding threshold> \
        --evm-starting-block <block number> \
        --evm-min-fee-for-bridging <minimal bridging fee> \
        --min-operation-fee <minimal operation fee> \
        --evm-relayer-gas-fee-multiplier <gas fee multiplier for evm relayer> \
        --chain-id <evm chain id> \
        --allowed-directions <allowed bridging directions> \
        --allowed-directions <allowed bridging directions> \
        --output-validator-components-file-name <validator components config json output file name> \
        --output-relayer-file-name <relayer config json output file name> \
        --empty-blocks-threshold <maximum number of empty blocks for blocks submitter to skip> \
        --dbs-path <path to where databases will be stored> \
        --relayer-config <path to relayer secrets manager config file> \
        --relayer-data-dir <path to relayer secret directory when using local secrets manager> \
        --output-dir <path to config jsons output directory> 
```

Add evm chain config minimal example
``` shell
$ apex-bridge generate-configs evm-chain \
        --evm-node-url localhost:5500 \
        --evm-ttl-block-inc 10 \
        --evm-block-rounding-threshold 5 \
        --evm-starting-block 3 \
        --evm-min-fee-for-bridging 10000000 \
        --min-operation-fee 0 \
        --evm-relayer-gas-fee-multiplier 4 \
        --chain-id "nexus" \
        --allowed-directions "prime" \
        --allowed-directions "vector" \
        --output-validator-components-file-name "vc_config.json"
```

# Example of sending a transaction from the prime to the vector
```shell
$ apex-bridge sendtx \
        --config ./config.json \
        --tx-type cardano \
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
        --config ./config.json \
        --tx-type cardano \
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
        --config ./config.json \
        --tx-type cardano \
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
        --config ./config.json \
        --tx-type evm \
        --key NEXUS_WALLET_PRIVATE_KEY \
        --nexus-url https://testnet.af.route3.dev/json-rpc/p2-c \
        --gateway-addr GATEWAY_PROXY_ADDRESS \
        --chain-src nexus \
        --currency-token-id 1 \
        --fee 1000010000000000000 \
        --chain-dst prime \
        --receiver addr_test1vrlt3wnp3hxermfyhfp2x9lu5u32275lf0yh3nvxkpjv7qgxl9f8y:1000000000000000000 \
        --ogmios-dst http://ogmios.prime.testnet.apexfusion.org:1337
```
- there is an optional `--stake-key` flag

# Example of sending a skyline transaction from the cardano to the prime
```shell
$ apex-bridge sendtx skyline \
        --config ./config.json \
        --tx-type cardano \
        --key CARDANO_WALLET_PRIVATE_KEY \
        --ogmios-src http://ogmios.cardano.testnet.apexfusion.org:1337 \
        --addr-multisig-src addr_test1wrz24vv4tvfqsywkxn36rv5zagys2d7euafcgt50gmpgqpq4ju9uv \
        --testnet-src 3311 \
        --network-id-src 1 \
        --chain-src cardano \
        --src-token-id 1 \
        --fee 1_100_000 \
        --ogmios-dst http://ogmios.prime.testnet.apexfusion.org:1337 \
        --chain-dst prime \
        --receiver addr_test1vrlt3wnp3hxermfyhfp2x9lu5u32275lf0yh3nvxkpjv7qgxl9f8y:1_234_567 \
        --dst-token-name 72f3d1e6c885e4d0bdcf5250513778dbaa851c0b4bfe3ed4e1bcceb0.4b6173685f546f6b656e
```
- optional `--src-token-name` which can be used instead of `--dst-token-name`
- there is an optional `--stake-key` flag

# Example of sending a skyline transaction from the cardano to the nexus
```shell
$ apex-bridge sendtx skyline \
        --config ./config.json \
        --tx-type cardano \
        --key CARDANO_WALLET_PRIVATE_KEY \
        --ogmios-src http://ogmios.cardano.testnet.apexfusion.org:1337 \
        --addr-multisig-src addr_test1wrz24vv4tvfqsywkxn36rv5zagys2d7euafcgt50gmpgqpq4ju9uv \
        --testnet-src 3311 \
        --network-id-src 1 \
        --chain-src cardano \
        --src-token-id 1 \
        --src-token-name 72f3d1e6c885e4d0bdcf5250513778dbaa851c0b4bfe3ed4e1bcceb0.4b6173685f546f6b656e \
        --fee 1_100_000 \
        --chain-dst nexus \
        --nexus-url https://testnet.af.route3.dev/json-rpc/p2-c \
        --receiver 0x1111:1_234_567 \
        --dst-token-contract-addr 0x22222
```
- there is an optional `--stake-key` flag

# Example of sending a skyline transaction from the nexus to the cardano
```shell
$ apex-bridge sendtx skyline \
        --config ./config.json \
        --tx-type evm \
        --key EVM_WALLET_PRIVATE_KEY \
        --nexus-url https://testnet.af.route3.dev/json-rpc/p2-c \
        --gateway-addr 0x3333 \
        --native-token-wallet-contract-addr 0x4444 \
        --chain-src nexus \
        --src-token-id 1 \
        --src-token-contract-addr 0x22222 \
        --fee 1_100_000 \
        --chain-dst cardano \
        --ogmios-dst http://ogmios.cardano.testnet.apexfusion.org:1337 \
        --receiver addr_test1wrz24vv4tvfqsywkxn36rv5zagys2d7euafcgt50gmpgqpq4ju9uv:1_234_567 \
        --dst-token-name 72f3d1e6c885e4d0bdcf5250513778dbaa851c0b4bfe3ed4e1bcceb0.4b6173685f546f6b656e
```
- `--src-token-name` flag should be set to `lovelace` only when sending the native currency and omitted in all other cases.
- there is an optional `--stake-key` flag

# How to Deploy Nexus Smart Contracts
Default example (bls keys are retrieved from bridge and gateway address is updated on the bridge):
```shell
$ apex-bridge deploy-evm \
        --config ./config.json \
        --url http://127.0.0.1:12001 \
        --key NEXUS_OR_EVM_PRIVATE_KEY \
        --dir /tmp \
        --clone \
        --min-fee 100\
        --min-bridging-amount 200\
        --min-token-bridging-amount 300\
        --min-operation-fee 400\
        --currency-token-id 1\
        --bridge-url http://127.0.0.1:12013 \
        --bridge-addr 0xABEF000000000000000000000000000000000000 \
        --bridge-key BRIDGE_ADMIN_WALLET_PRIVATE_KEY \
```
- all amounts should be entered in wei decimals
- `--currency-token-id` flag is the ecosystem token id of the currency of the chain the SCs are being deployed to
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
        --config ./config.json \
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
        --config ./config.json \
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
        --config ./config.json \
        --bridge-url http://localhost:12013 \
        --chain prime --chain nexus --chain vector
```

```shell
$ apex-bridge bridge-admin update-chain-token-quantity \
        --config ./config.json \
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
        --min-bridging-amount 100 \
        --min-token-bridging-amount 200 \
        --min-operation-fee 100 
```
- all amounts should be entered in wei decimals
- instead of `--key` it is possible to set key secret manager configuration file with `--key-config /path/config.json`.


```shell
$ apex-bridge bridge-admin defund \
        --config ./config.json \
        --bridge-url http://localhost:12013 \
        --chain nexus \
        --key 922769e22b70614d4172fc899126785841f4de7d7c009fc338923ce50683023d \
        --amount 100 \
        --native-token-amount 200 \
        --addr 0xeefcd00000000000000000000000000000000000
```
- instead of `--key` it is possible to set key secret manager configuration file with `--key-config /path/config.json`.
- there is an optional `--native-token-amount` flag, which refers to wrapped currency when `--token-id` flag is not passed
- there is an optional `--token-id` flag used for colored coins in combination with `--native-token-amount` flag

```shell
$ apex-bridge bridge-admin set-additional-data \
        --config ./config.json \
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
        --config ./config.json \
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
        --config ./config.json \
        --bridge-address-index 0 \
        --bridge-url http://localhost:12001 \
        --chain prime \
        --key 922769e22b70614d4172fc899126785841f4de7d7c009fc338923ce50683023d 
```
- instead of `--key` it is possible to set key secret manager configuration file with `--key-config /path/config.json`.

```shell
$ apex-bridge bridge-admin update-bridging-addrs-count \
        --config ./config.json \
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
$ apex-bridge bridge-admin deploy-cardano-script \
        --key PRIME_WALLET_PRIVATE_KEY \
        --ogmios http://ogmios.prime.testnet.apexfusion.org:1337 \
        --network-id 1 \
        --testnet-magic 3311 \
        --nft-policy-id 14b249936a64cbc96bde5a46e04174e7fb58b565103d0c3a32f8d61f \
        --nft-name-hex 54657374546F6B656E
```
- optional `--plutus-script-dir`

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
$ apex-bridge bridge-admin register-gateway-token \
        --config ./config.json \
        --node-url http://localhost:12001 \
        --gateway-address 0x020202 \
        --gas-limit 10_000_000 \
        --token-sc-address 0x03030300 \
        --token-id 2 \
        --token-name USDT \
        --token-symbol USDT \
        --key 922769e22b70614d4172fc899126785841f4de7d7c009fc338923ce50683023d
```
- instead of `--key` it is possible to set key secret manager configuration file with `--key-config /path/config.json`.
- `--token-sc-address` is only passed if the token contract is not owned by the bridge

```shell
$ apex-bridge bridge-admin redistribute-bridging-addresses-tokens \
        --config ./config.json \
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
