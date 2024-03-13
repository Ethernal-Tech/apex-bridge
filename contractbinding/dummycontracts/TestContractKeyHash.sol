// SPDX-License-Identifier: GPL-3.0
pragma solidity >0.7.0 < 0.9.0;
/**
* @title TestContract
* @dev store or retrieve variable value
*/

contract TestContractKeyHash {
    struct ValidatorCardanoData {
        string KeyHash;
        string KeyHashFee;
        bytes VerifyingKey;
        bytes VerifyingKeyFee;
    }

    mapping(address => mapping(string => ValidatorCardanoData)) public validatorCardanoDataMap;

    function setValidatorCardanoData(string calldata chainID, ValidatorCardanoData calldata vd) public{
		validatorCardanoDataMap[msg.sender][chainID] = vd;
	}
}