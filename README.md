# apex-bridge
Apex Bridge componentes written in Go

# How to go get private repo
`$ git config url."git@github.com:Ethernal-Tech/cardano-infrastructure.git".insteadOf "https://github.com/Ethernal-Tech/cardano-infrastructure"`
`$ GOPRIVATE=github.com/Ethernal-Tech/cardano-infrastructure go get github.com/Ethernal-Tech/cardano-infrastructure`

# Example how to generate go binding for smart contract
`$ solcjs -p --abi contractbinding/dummycontracts/TestContract.sol -o ./contractbinding/contractbuild && abigen --abi ./contractbinding/contractbuild/contractbinding_dummycontracts_TestContract_sol_TestContract.abi --pkg main --type TestContract --out ./contractbinding/TestContract.go --pkg contractbinding`