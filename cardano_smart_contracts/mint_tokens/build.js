import { readFileSync } from "node:fs"
import { fileURLToPath } from "node:url"
import path from "node:path"
import { bytesToHex } from "@helios-lang/codec-utils"
import { Program } from "@helios-lang/compiler"
import { makeMintingPolicyHash } from "@helios-lang/ledger"
import { makeByteArrayData } from "@helios-lang/uplc";
import { hexToBytes } from "./test/helpers.js";

// get CLI arguments
// usage: node build.js <NFT_POLICY_ID> <NFT_NAME_HEX>
const [, , nftPolicyIdArg, nftNameHexArg] = process.argv;

if (!nftPolicyIdArg || !nftNameHexArg) {
  throw new Error("Usage: node build.js <NFT_POLICY_ID> <NFT_NAME_HEX>");
}

const dirname = path.dirname(fileURLToPath(import.meta.url))
const src = readFileSync(path.join(dirname, "mint_validator.hl")).toString()

// compile the validator
const program = new Program(src)

const mintinPolicyHash = makeMintingPolicyHash(nftPolicyIdArg)
const nftPolicySet = program.changeParam("minting_policy::NFT_POLICY", mintinPolicyHash)

if (!nftPolicySet) {
  throw new Error("Failed to set parameter NFT_POLICY");
}

const nftName = makeByteArrayData(hexToBytes(nftNameHexArg))
const nftNameSet = program.changeParam("minting_policy::NFT_NAME", nftName)

if (!nftNameSet) {
  throw new Error("Failed to set parameter NFT_NAME");
}

const uplc = program.compile(true)

// === Build the JSON structure
const plutusScript = {
  type: "PlutusScriptV2",
  description: "",
  cborHex: bytesToHex(uplc.toCbor())
}

console.log(JSON.stringify(plutusScript));
