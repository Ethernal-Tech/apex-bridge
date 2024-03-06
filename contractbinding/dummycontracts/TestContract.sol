// SPDX-License-Identifier: GPL-3.0
pragma solidity >0.7.0 < 0.9.0;
/**
* @title TestContract
* @dev store or retrieve variable value
*/

contract TestContract {
    struct ConfirmedBatch {
        string id;
        string rawTransaction;
        string[] multisigSignatures;
        string[] feePayerMultisigSignatures;
    }
    
    uint256 value;
    ConfirmedBatch public confirmedBatch;

    function setValue(uint256 number) public {
        value = number;
    }

    function getValue() public view returns (uint256) {
		return value;
	}


    function setConfirmedBatch(ConfirmedBatch memory _newValue) public {
        confirmedBatch = _newValue;
    }

    function getConfirmedBatch() public view returns (ConfirmedBatch memory) {
        return confirmedBatch;
    }
}