// @ts-nocheck
// test/helpers.js - Test helper functions for building ScriptContext
import { readFileSync } from "fs";
import { Program } from "@helios-lang/compiler";
import {
  makeIntData,
  makeByteArrayData,
  makeListData,
  makeMapData,
  makeConstrData
} from "@helios-lang/uplc";
import { makeMintingPolicyHash } from "@helios-lang/ledger"

/**
 * Compile the Helios validator and return the compiled UPLC program
 */
export function compileValidator() {
  const src = readFileSync(new URL("../mint_validator.hl", import.meta.url)).toString();
  const program = new Program(src);

  const mintinPolicyHash = makeMintingPolicyHash("14b249936a64cbc96bde5a46e04174e7fb58b565103d0c3a32f8d61f")
  const nftPolicySet = program.changeParam("minting_policy::NFT_POLICY", mintinPolicyHash)
  
  if (!nftPolicySet) {
    throw new Error("Failed to set parameter NFT_POLICY");
  }

  const nftName = makeByteArrayData(hexToBytes("54657374546F6B656E"))
  const nftNameSet = program.changeParam("minting_policy::NFT_NAME", nftName)

  if (!nftNameSet) {
    throw new Error("Failed to set parameter NFT_NAME");
  }

  return program.compile();
}

/**
 * Convert hex string to Uint8Array
 */
export function hexToBytes(hex) {
  const clean = hex.startsWith("0x") ? hex.slice(2) : hex;
  const bytes = new Uint8Array(clean.length / 2);
  for (let i = 0; i < bytes.length; i++) {
    bytes[i] = parseInt(clean.substr(i * 2, 2), 16);
  }
  return bytes;
}

/**
 * Build a Value map containing exactly 1 unit of an NFT
 * Value encoding: Map<PolicyId(ByteArray), Map<TokenName(ByteArray), Int>>
 */
export function makeNftValue(nftPolicy, nftName) {
  const policy = makeByteArrayData(hexToBytes(nftPolicy));
  const tokenName = makeByteArrayData(hexToBytes(nftName));
  const inner = makeMapData([[tokenName, makeIntData(1)]]);
  return makeMapData([[policy, inner]]);
}

/**
 * Create a dummy TxInput with the given value
 */
export function makeDummyTxIn(valueData, placeNFTInInput = true) {
  // A TxInInfo is usually a pair (outRef, txOut)
  // We'll encode outRef as a ConstrData with (txId :: ByteArray, index :: Int)
  const outRef = makeConstrData(0, [
    makeByteArrayData(hexToBytes("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")), // txId
    makeIntData(0) // index
  ]);

  // txOut: encode minimal TxOut structure: (address, value, datumHash/inline datum)
  const txOut = makeConstrData(0, [
    makeByteArrayData(new Uint8Array([1])), // address as bytearray placeholder
    placeNFTInInput ? valueData : makeMapData([]), // value containing the NFT
    makeIntData(0) // datum placeholder
  ]);

  // TxInInfo as map or pair depending on the expected Plutus encoding.
  return makeConstrData(0, [outRef, txOut]);
}

/**
 * Create a dummy TxOutput with the given value
 */
export function makeDummyTxOut(valueData, placeNFTInOutput = true) {
  return makeConstrData(0, [
    makeByteArrayData(new Uint8Array([1])),
    placeNFTInOutput ? valueData : makeMapData([]), // value containing the NFT
    makeIntData(0) // datum placeholder
  ]);
}

/**
 * Create a dummy TxInfo with the given value in inputs/outputs/mint
 */
export function makeDummyTxInfo(valueData, placeNFTInOutput = true, placeNFTInInput = true) {
  const inputs = makeListData([makeDummyTxIn(valueData, placeNFTInInput)]); // list of TxInInfo
  const referenceInputs = makeListData([]);       // none
  const outputs = makeListData([makeDummyTxOut(valueData, placeNFTInOutput)]); // created outputs by tx
  const fee = makeMapData([]); // empty = zero fee (or fill a map with currency -> amount)
  const mint = valueData; // include token data
  const dcert = makeListData([]); // none
  const wdrl = makeMapData([]); // none
  const valid_range = makeConstrData(0, [makeIntData(0), makeIntData(1000)]); // POSIX range start..end

  // TxInfo is commonly a constructor with all those fields in order
  return makeConstrData(0, [
    inputs,
    referenceInputs,
    outputs,
    fee,
    mint,
    dcert,
    wdrl,
    valid_range
  ]);
}

/**
 * Create a Minting purpose for the NFT policy
 */
export function makeMintingPurpose(policyId) {
  // ScriptPurpose::Minting(policyId)
  const policy = makeByteArrayData(hexToBytes(policyId));
  // encode purpose with tag 1 for Minting
  return makeConstrData(1, [policy]);
}

/**
 * Decode and log error details from evaluation result
 */
export function decodeError(result) {
  if ("left" in result) {
    console.log("Error message:", result.left.error || "(empty)");
    console.log("\nCall sites:");
    result.left.callSites.forEach((site, i) => {
      console.log(`  [${i}]:`, site);
    });
  }
}
