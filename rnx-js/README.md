# @rnx/sdk — RNX Testnet JS/TS SDK (skeleton)

> ⚠️ **TESTNET ONLY — NOT REAL MONEY — NOT MAINNET.** This is an SDK *skeleton*.
> Every method has a real, stable signature with docstrings, but implementations
> throw `NotImplementedError` so no fake on-chain data is ever returned.
> Endpoints are placeholders until RNX Testnet infrastructure is provisioned.

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
  rpcUrl: "<PLACEHOLDER_TESTNET_RPC_URL>",
  chainId: RNX_CHAIN_ID, // 194151
  veridexEndpoint: "<PLACEHOLDER_VERIDEX_ENDPOINT>",
});

// Each call throws NotImplementedError until wired to real infrastructure:
// const provider = client.connect();
// const bal = await client.getBalance("0x...");
// const tx  = await client.sendTransaction(signer, { to, value });
// const blk = await client.getBlock("latest");
// const t   = await client.getTransaction("0x...");
// const c   = client.contract(addr, abi, signer);
// const v   = await client.veridexVerify("rakshanex.bihar");
```

## API surface (stubs)

| Method | Purpose |
|--------|---------|
| `connect()` | Return an ethers `JsonRpcProvider` for RNX Testnet. |
| `getBalance(address)` | Native RNX (test) balance in wei. |
| `sendTransaction(signer, tx)` | Sign + broadcast a testnet transaction. |
| `getBlock(blockHashOrTag)` | Fetch a block. |
| `getTransaction(txHash)` | Fetch a transaction. |
| `contract(address, abi, signerOrProvider?)` | Bind an ethers `Contract`. |
| `veridexVerify(subject)` | Verify a name/proof via VERIDEX (fail-closed). |

## Source

- `src/index.ts` — client + typed stubs.

## Safety

- Never commit private keys. Testnet keys are developer-held only.
- No real transactions, deploys, or production changes result from this
  skeleton.
