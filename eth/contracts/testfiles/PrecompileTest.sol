// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import "./console.sol";
import "./IPrecompileValidators.sol";
import "./IPrecompileStructs.sol";
/**
 * @title Storage
 * @dev Store & retrieve value in a variable
 * @custom:dev-run-script ./scripts/deploy_with_ethers.ts
 */
contract PrecompileTest is IPrecompileStructs {
    address public constant VALIDATOR_BLS_PRECOMPILE = 0x0000000000000000000000000000000000002060;
    uint256 public constant VALIDATOR_BLS_PRECOMPILE_GAS = 150000;

    event CustomResult(
        bool callSuccess,
        bool result
    );

    function deposit(
        bytes calldata _signature,
        uint256 _bitmap,
        bytes calldata _data,
        ValidatorChainData[] memory validatorsChainData
    ) external {
        bytes32 _hash = keccak256(_data);

        // verify signatures` for provided sig data and sigs bytes
        // solhint-disable-next-line avoid-low-level-calls
        // slither-disable-next-line low-level-calls,calls-loop
        (bool callSuccess, bytes memory returnData) = VALIDATOR_BLS_PRECOMPILE
            .staticcall{gas: VALIDATOR_BLS_PRECOMPILE_GAS}(
            abi.encodePacked(
                uint8(1),
                abi.encode(_hash, _signature, validatorsChainData, _bitmap)
            )
        );

        bool result = false;
        if (callSuccess) {
            result = abi.decode(returnData, (bool));
        } 
        emit CustomResult(callSuccess, result);
    }

    function depositDirect(
        bytes calldata _signature,
        uint256 _bitmap,
        bytes calldata _data
    ) external {
        IPrecompileValidators validators = (IPrecompileValidators)(address(0xfefBD392E59DFac2C30cd9c1dB89f4798348EA69));
        ValidatorChainData[] memory validatorsChainData = validators.getValidatorsChainData();
        bytes32 _hash = keccak256(_data);

        // verify signatures` for provided sig data and sigs bytes
        // solhint-disable-next-line avoid-low-level-calls
        // slither-disable-next-line low-level-calls,calls-loop
        (bool callSuccess, bytes memory returnData) = VALIDATOR_BLS_PRECOMPILE
            .staticcall{gas: VALIDATOR_BLS_PRECOMPILE_GAS}(
            abi.encodePacked(
                uint8(1),
                abi.encode(_hash, _signature, validatorsChainData, _bitmap)
            )
        );

        bool result = false;
        if (callSuccess) {
            result = abi.decode(returnData, (bool));
        } 
        emit CustomResult(callSuccess, result);
    }

    function depositOrig(
        bytes calldata _signature,
        uint256 _bitmap,
        bytes calldata _data
    ) external {
        IPrecompileValidators validators = (IPrecompileValidators)(address(0xfefBD392E59DFac2C30cd9c1dB89f4798348EA69));
        bytes32 _hash = keccak256(_data);
        bool valid = validators.isBlsSignatureValid(_hash, _signature, _bitmap);

        emit CustomResult(true, valid);
    }
}