/**
 * @rnx/sdk — read-only usage example (TESTNET ONLY — NOT REAL MONEY).
 *
 * This example performs ONLY read operations against a placeholder testnet RPC.
 * It never signs or broadcasts a transaction and never handles a private key.
 *
 * Run (after `npm install` and `npm run build`):
 *   RNX_RPC_URL="<testnet-rpc>" RNX_REGISTRY="0x..." node example.js
 *
 * With the default placeholder RPC the network calls will simply fail to
 * connect — which is expected; this file demonstrates the API shape and the
 * fail-closed behavior, not live data.
 */

import { RnxClient, RNX_CHAIN_ID, RNX_CHAIN_ID_HEX, nameHash } from "./dist/index.js";

// Placeholder testnet RPC. Override with the RNX_RPC_URL env var.
const RPC_URL = process.env.RNX_RPC_URL || "https://rpc.rnx-testnet.invalid";
// Placeholder registry address. Override with RNX_REGISTRY (required for verify).
const REGISTRY = process.env.RNX_REGISTRY || "";

async function main() {
  console.log("RNX SDK example — TESTNET ONLY, read-only. No real money.");
  console.log(`chainId=${RNX_CHAIN_ID} (${RNX_CHAIN_ID_HEX})`);

  const client = new RnxClient({
    rpcUrl: RPC_URL,
    chainId: RNX_CHAIN_ID,
    registryAddress: REGISTRY || undefined,
  });

  // Deterministic, offline: nameHash does not touch the network.
  console.log(`nameHash("rakshanex.bihar") = ${nameHash("rakshanex.bihar")}`);

  // --- Read-only chain queries (require a reachable RPC) ---
  try {
    const block = await client.getBlock("latest");
    console.log(`latest block number: ${block?.number ?? "n/a"}`);

    const addr = "0x0000000000000000000000000000000000000000";
    const bal = await client.getBalance(addr);
    console.log(`balance(${addr}) = ${bal} wei (test RNX, no value)`);

    if (REGISTRY) {
      // Fail-closed VERIDEX verify via currentAnchor (read-only).
      const result = await client.veridexVerify("rakshanex.bihar");
      console.log("veridexVerify:", result);
    } else {
      console.log("Set RNX_REGISTRY to run veridexVerify (currentAnchor read).");
    }
  } catch (err) {
    // Expected with a placeholder/unreachable RPC. We never fabricate data.
    console.log(`read query failed (expected with placeholder RPC): ${err.message}`);
  }

  // NOTE: writes use client.sendRawTransaction(signedRawTx) with a PRE-SIGNED
  // transaction produced by YOUR wallet. This SDK never handles private keys,
  // so no write is performed in this read-only example.
}

main();
