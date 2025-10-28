// @ts-nocheck
// test/validator.spec.ts
import {
  makeIntData,
  makeMapData,
  makeConstrData,
  makeUplcDataValue,
  makeBasicUplcLogger
} from "@helios-lang/uplc";
import { expect, test } from "vitest";
import {
  compileValidator,
  makeNftValue,
  makeDummyTxInfo,
  makeMintingPurpose,
  decodeError
} from "./helpers.js";

const NFT_POLICY = "14b249936a64cbc96bde5a46e04174e7fb58b565103d0c3a32f8d61f";
const NFT_NAME = "54657374546F6B656E";
const MINTING_POLICY = "14b249936a64cbc96bde5a46e04174e7fb58b565103d0c3a32f8d61e";

test("validator should pass", () => {
  const uplc = compileValidator();

  // build redeemer matching the validator signature (redeemer must equal 4)
  const redeemer = makeUplcDataValue(makeIntData(4));

  // build txInfo & purpose then ScriptContext as a constructor (txInfo, purpose)
  const nftValue = makeNftValue(NFT_POLICY, NFT_NAME);
  const txInfo = makeDummyTxInfo(nftValue);
  const purpose = makeMintingPurpose(MINTING_POLICY);
  const scriptContextData = makeConstrData(0, [txInfo, purpose]);
  const scriptContext = makeUplcDataValue(scriptContextData);

  // evaluate
  const { result } = uplc.eval([redeemer, scriptContext]);

  expect("left" in result).toBe(false);
  // if right, check for expected return type (unit)
  const r = result.right;
  if (typeof r === "string") {
    throw new Error(`unexpected string result: ${r}`);
  }

  expect(r.kind).toBe("unit");
});

// test("validator should fail with wrong redeemer", () => {
//     const uplc = compileValidator();
//   
//     // build redeemer with wrong value (5 instead of 4)
//     const redeemer = makeUplcDataValue(makeIntData(5));
//   
//     // build txInfo & purpose then ScriptContext as a constructor (txInfo, purpose)
//     const nftValue = makeNftValue(NFT_POLICY, NFT_NAME);
//     const txInfo = makeDummyTxInfo(nftValue);
//     const purpose = makeMintingPurpose(MINTING_POLICY);
//     const scriptContextData = makeConstrData(0, [txInfo, purpose]);
//     const scriptContext = makeUplcDataValue(scriptContextData);
//   
//     // evaluate
//     const { result } = uplc.eval([redeemer, scriptContext]);
//   
//     // This test expects the validator to fail
//     expect("left" in result).toBe(true);
// });

test("validator should fail for missing NFT", () => {
  const uplc = compileValidator();
  // build redeemer matching the validator signature (redeemer must equal 4)
  const redeemer = makeUplcDataValue(makeIntData(4));

  // build txInfo & purpose then ScriptContext as a constructor (txInfo, purpose)
  const nftValue = makeMapData([]);
  const txInfo = makeDummyTxInfo(nftValue);
  const purpose = makeMintingPurpose(MINTING_POLICY);
  const scriptContextData = makeConstrData(0, [txInfo, purpose]);
  const scriptContext = makeUplcDataValue(scriptContextData);
  // evaluate

  const logger = makeBasicUplcLogger();

  const { result } = uplc.eval([redeemer, scriptContext], logger);

  // This test expects the validator to fail
  expect("left" in result).toBe(true);
}); 

test("validator should fail for wrong NFT", () => {
    const uplc = compileValidator();
    // build redeemer matching the validator signature (redeemer must equal 4)
    const redeemer = makeUplcDataValue(makeIntData(4));
  
    // build txInfo & purpose then ScriptContext as a constructor (txInfo, purpose)
    const nftValue = makeNftValue(NFT_POLICY, "WrongNFT");    
    const txInfo = makeDummyTxInfo(nftValue);
    const purpose = makeMintingPurpose(MINTING_POLICY);
    const scriptContextData = makeConstrData(0, [txInfo, purpose]);
    const scriptContext = makeUplcDataValue(scriptContextData);
    // evaluate
  
    const logger = makeBasicUplcLogger();
  
    const { result } = uplc.eval([redeemer, scriptContext], logger);
  
    // This test expects the validator to fail
    expect("left" in result).toBe(true);
});

test("validator should fail for missing NFT in input", () => {
    const uplc = compileValidator();
    // build redeemer matching the validator signature (redeemer must equal 4)
    const redeemer = makeUplcDataValue(makeIntData(4));
    
    // build txInfo & purpose then ScriptContext as a constructor (txInfo, purpose)
    const nftValue = makeNftValue(NFT_POLICY, NFT_NAME);
    const txInfo = makeDummyTxInfo(nftValue, true, false);
    const purpose = makeMintingPurpose(MINTING_POLICY);
    const scriptContextData = makeConstrData(0, [txInfo, purpose]);
    const scriptContext = makeUplcDataValue(scriptContextData);
  
    // evaluate
    const logger = makeBasicUplcLogger();
    const { result } = uplc.eval([redeemer, scriptContext], logger);
    // This test expects the validator to fail
    expect("left" in result).toBe(true);
});

test("validator should fail for missing NFT in output", () => {
  const uplc = compileValidator();
  // build redeemer matching the validator signature (redeemer must equal 4)
  const redeemer = makeUplcDataValue(makeIntData(4));
  
  // build txInfo & purpose then ScriptContext as a constructor (txInfo, purpose)
  const nftValue = makeNftValue(NFT_POLICY, NFT_NAME);
  const txInfo = makeDummyTxInfo(nftValue, false);
  const purpose = makeMintingPurpose(MINTING_POLICY);
  const scriptContextData = makeConstrData(0, [txInfo, purpose]);
  const scriptContext = makeUplcDataValue(scriptContextData);

  // evaluate
  const logger = makeBasicUplcLogger();
  const { result } = uplc.eval([redeemer, scriptContext], logger);
  // This test expects the validator to fail
  expect("left" in result).toBe(true);
});
