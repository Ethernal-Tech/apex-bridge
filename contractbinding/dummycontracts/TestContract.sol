// SPDX-License-Identifier: GPL-3.0
pragma solidity >0.7.0 <0.9.0;
/**
 * @title TestContract
 * @dev store or retrieve variable value
 */

contract TestContract {
    uint256 value;

    function setValue(uint256 number) public {
        value = number;
    }

    function getValue() public view returns (uint256) {
        return value;
    }

    // Batches
    function submitSignedBatch(
        SignedBatch calldata signedBatch
    ) external onlyValidator {}

    // Queries

    // Will determine if enough transactions are confirmed, or the timeout between two batches is exceeded.
    // It will also check if the given validator already submitted a signed batch and return the response accordingly.
    function shouldCreateBatch(
        string calldata destinationChain
    ) external view returns (bool batch) {}

    // Will return confirmed transactions until NEXT_BATCH_TIMEOUT_BLOCK or maximum number of transactions that
    // can be included in the batch, if the maximum number of transactions in a batch has been exceeded
    function getConfirmedTransactions(
        string calldata destinationChain
    )
        external
        view
        returns (ConfirmedTransaction[] memory confirmedTransactions)
    {
        ConfirmedTransaction[] memory dummyArray;
        return dummyArray;
    }

    // Will return available UTXOs that can cover the cost of bridging transactions included in some batch.
    // Each Batcher will first call the GetConfirmedTransactions() and then calculate (off-chain) how many tokens
    // should be transfered to users and send this info through the 'txCost' parameter. Based on this input and
    // number of UTXOs that need to be consolidated, the smart contract will return UTXOs belonging to the multisig address
    // that can cover the expenses. Additionaly, this method will return available UTXOs belonging to fee payer
    // multisig address that will cover the network fees (see chapter "2.2.2.3 Batcher" for more details)
    function getAvailableUTXOs(
        string calldata destinationChain,
        uint256 txCost
    ) external view returns (UTXOs memory availableUTXOs) {
        UTXOs memory dummyArray;
        return dummyArray;
    }

    function getConfirmedBatch(
        string calldata destinationChain
    ) external view returns (ConfirmedBatch memory batch) {
        ConfirmedBatch memory dummyArray;
        return dummyArray;
    }

    // only allowed for validators
    modifier onlyValidator() {
        _;
    }

    struct SignedBatch {
        string id;
        string destinationChainId;
        string rawTransaction;
        string multisigSignature;
        string feePayerMultisigSignature;
        ConfirmedTransaction[] includedTransactions;
        UTXOs usedUTXOs;
    }

    struct ConfirmedBatch {
        string id;
        string rawTransaction;
        string[] multisigSignatures;
        string[] feePayerMultisigSignatures;
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
}
