// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import "./IPrecompileStructs.sol";

interface IPrecompileValidators is IPrecompileStructs {
    function isBlsSignatureValid(
        bytes32 _hash,
        bytes calldata _signature,
        uint256 _bitmap
    ) external view returns (bool);

    function getValidatorsChainData()
        external
        view
        returns (ValidatorChainData[] memory);
}
