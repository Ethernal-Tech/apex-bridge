// SPDX-License-Identifier: GPL-3.0
pragma solidity >0.7.0 < 0.9.0;
/**
* @title TestContract
* @dev store or retrieve variable value
*/

contract TestContract {

	uint256 value;

	function setValue(uint256 number) public{
		value = number;
	}

	function getValue() public view returns (uint256){
		return value;
	}
}