// SPDX-License-Identifier: GPL-3.0
pragma solidity >0.7.0 <0.9.0;
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

    // Dummy function
    function setConfirmedBatch(ConfirmedBatch memory _newValue) public {
        confirmedBatch = _newValue;
    }

    function getConfirmedBatch(
        string calldata destinationChain
    ) external view returns (ConfirmedBatch memory batch) {
        if (
            keccak256(abi.encodePacked(destinationChain)) ==
            keccak256(abi.encodePacked("prime"))
        ) {
            batch = confirmedBatch;
        } else {
            batch = confirmedBatch;
        }
    }

    // Batcher

    bool batcher1submitted = true;
    bool batcher2submitted = true;
    bool batcher3submitted = true;
    bool shouldRetrieve = false;
    string[] multisigSigs;
    string[] multisigFeeSigs;

    struct SignedBatch {
        string id;
        string destinationChainId;
        string rawTransaction;
        string multisigSignature;
        string feePayerMultisigSignature;
        ConfirmedTransaction[] includedTransactions;
        UTXOs usedUTXOs;
    }

    struct ConfirmedTransaction {
        uint256 nonce;
        //mapping(string => uint256) receivers;
        Receiver[] receivers;
    }

    struct Receiver {
        string destinationAddress;
        uint256 amount;
    }

    struct UTXOs {
        UTXO[] multisigOwnedUTXOs;
        UTXO[] feePayerOwnedUTXOs;
    }

    struct UTXO {
        string txHash;
        uint256 txIndex;
        uint256 amount;
    }

    function shouldCreateBatch(
        string calldata destinationChain
    ) external view returns (bool) {
        if (
            keccak256(abi.encodePacked(destinationChain)) ==
            keccak256(abi.encodePacked("prime1"))
        ) return batcher1submitted;
        else if (
            keccak256(abi.encodePacked(destinationChain)) ==
            keccak256(abi.encodePacked("prime2"))
        ) return batcher2submitted;
        else if (
            keccak256(abi.encodePacked(destinationChain)) ==
            keccak256(abi.encodePacked("prime3"))
        ) return batcher3submitted;
        return false;
    }

    // Dummy for testing
    function shouldRelayerRetrieve() external view returns (bool) {
        return shouldRetrieve;
    }
    function resetShouldRetrieve() external {
        shouldRetrieve = !shouldRetrieve;
    }

    function submitSignedBatch(SignedBatch calldata signedBatch) external {
        if (
            keccak256(abi.encodePacked(signedBatch.destinationChainId)) ==
            keccak256(abi.encodePacked("prime1"))
        ) batcher1submitted = false;
        if (
            keccak256(abi.encodePacked(signedBatch.destinationChainId)) ==
            keccak256(abi.encodePacked("prime2"))
        ) batcher2submitted = false;
        if (
            keccak256(abi.encodePacked(signedBatch.destinationChainId)) ==
            keccak256(abi.encodePacked("prime3"))
        ) batcher3submitted = false;

        multisigSigs.push(signedBatch.multisigSignature);
        multisigFeeSigs.push(signedBatch.feePayerMultisigSignature);

        if (multisigSigs.length == 3) {
            ConfirmedBatch memory newBatch;
            newBatch.id = "1337";
            newBatch.rawTransaction = signedBatch.rawTransaction;
            newBatch.multisigSignatures = multisigSigs;
            newBatch.feePayerMultisigSignatures = multisigFeeSigs;

            if (!shouldRetrieve) {
                setConfirmedBatch(newBatch);
                shouldRetrieve = true;

                delete multisigSigs;
                delete multisigFeeSigs;
                batcher1submitted = true;
                batcher2submitted = true;
                batcher3submitted = true;
            }
        }
    }
}
