// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

interface IPrecompileStructs {
    struct ValidatorChainData {
        // verifying key, verifying Fee key for Cardano
        // BLS for EVM
        uint256[4] key;
    }
}
