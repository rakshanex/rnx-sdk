# @rnx/sdk — RNX Testnet JS/TS SDK

> ⚠️ **TESTNET ONLY — NOT REAL MONEY — NOT MAINNET.** Read methods query a real
> EVM JSON-RPC endpoint directly. **Write access is limited to broadcasting a
> PRE-SIGNED raw transaction** — this SDK never handles, stores, or requests
> private keys. All returned values come straight from the RPC; nothing is
> fabricated. Point `rpcUrl` at an RNX Testnet node to use it.

## Status / honesty

- RNX mainnet is **NOT live**. Chain ID `194151` (`0x2F667`) is the reserved
  **future** mainnet id and the testnet rehearsal id — engineering foundation
  only.
- Decentralization = **BOOTSTRAP** (0 independent operators; operator-run). Not
  decentralized/trustless/permissionless/production-mainnet.
- **qUSD is a TEST quote asset — NOT USDT, NOT USDC.**
- Legacy chain `12345` unchanged. `rakshanex.bihar → 194.164.151.23` preserved.
- The SDK never fabricates validators/users/TVL/liquidity/volume/followers.

## Install (when published)

```bash
npm install @rnx/sdk ethers
```

Minimal dependency surface: only [`ethers`](https://docs.ethers.org/) (v6).

## Usage

```ts
import { RnxClient, RNX_CHAIN_ID } from "@rnx/sdk";

const client = new RnxClient({
  rpcUrl: "<TESTNET_RPC_URL>",
  chainId: RNX_CHAIN_ID, // 194151
  registryAddress: "0x...", // VERIDEX anchor registry (for veridexVerify)
});

// Reads (query the RPC directly):
const provider = client.connect();
const bal = await client.getBalance("0x...");
const blk = await client.getBlock("latest");
const t   = await client.getTransaction("0x...");
const out = await client.call({ to: "0x...", data: "0x..." });
const c   = client.contract(addr, abi); // read-bound by default
const v   = await client.veridexVerify("rakshanex.bihar"); // currentAnchor read

// Write (PRE-SIGNED only — sign in YOUR wallet, SDK just broadcasts):
// const resp = await client.sendRawTransaction(signedRawTx);
```

## API surface

| Method | Purpose |
|--------|---------|
| `connect()` | Return a cached ethers `JsonRpcProvider` (static network). |
| `getBalance(address, blockTag?)` | Native RNX (test) balance in wei. |
| `getBlock(blockHashOrTag, prefetchTxs?)` | Fetch a block. |
| `getTransaction(txHash)` | Fetch a transaction. |
| `call(tx, blockTag?)` | Read-only `eth_call`. |
| `contract(address, abi, runner?)` | Bind an ethers `Contract` (read by default). |
| `veridexVerify(subject, registryAddress?)` | Read `currentAnchor` on the registry (fail-closed). |
| `sendRawTransaction(signedRawTx)` | Broadcast a PRE-SIGNED raw tx. No key handling. |
| `nameHash(subject)` | keccak256 nameHash helper (offline). |

## Source

- `src/index.ts` — client + real implementations.
- `example.js` — read-only usage example (placeholder RPC).
- `package.json` / `tsconfig.json` — build config (dep: `ethers` v6).

## Verification

- `npx tsc --noEmit` — type-checks clean against ethers v6.
- `node --check example.js` — parses clean.
- `node example.js` — runs read-only; fails closed on an unreachable
  placeholder RPC (never fabricates data).

## Safety

- Never commit private keys. Testnet keys are developer-held only.
- Writes are PRE-SIGNED raw txs only; the SDK never signs or holds keys.
- No real transactions, deploys, or production changes result from this SDK.
