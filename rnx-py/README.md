# rnx-sdk — RNX Testnet Python SDK

> ⚠️ **TESTNET ONLY — NOT REAL MONEY — NOT MAINNET.** Read methods query a real
> EVM JSON-RPC endpoint directly. **Write access is limited to broadcasting a
> PRE-SIGNED raw transaction** — this SDK never handles, stores, or requests
> private keys. All returned values come straight from the RPC; nothing is
> fabricated. Point `rpc_url` at an RNX Testnet node to use it.

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
pip install rnx-sdk web3
```

Minimal dependency surface: only [`web3`](https://web3py.readthedocs.io/).

## Usage

```python
from rnx_sdk import RnxClient, RNX_CHAIN_ID

client = RnxClient(
    rpc_url="<TESTNET_RPC_URL>",
    chain_id=RNX_CHAIN_ID,  # 194151
    registry_address="0x...",  # VERIDEX anchor registry (for veridex_verify)
)

# Reads (query the RPC directly):
w3  = client.connect()
bal = client.get_balance("0x...")
blk = client.get_block("latest")
t   = client.get_transaction("0x...")
out = client.call({"to": "0x...", "data": "0x..."})
c   = client.contract("0x...", abi)
v   = client.veridex_verify("rakshanex.bihar")  # currentAnchor read

# Write (PRE-SIGNED only — sign in YOUR account, SDK just broadcasts):
# tx_hash = client.send_raw_transaction(signed_raw_tx)
```

## API surface

| Method | Purpose |
|--------|---------|
| `connect()` | Return a cached `web3.Web3` (HTTP provider). |
| `get_balance(address, block_identifier?)` | Native RNX (test) balance in wei. |
| `get_block(block_identifier, full_transactions?)` | Fetch a block. |
| `get_transaction(tx_hash)` | Fetch a transaction. |
| `call(tx, block_identifier?)` | Read-only `eth_call`. |
| `contract(address, abi)` | Bind a web3 contract (read). |
| `veridex_verify(subject, registry_address?)` | Read `currentAnchor` (fail-closed). |
| `send_raw_transaction(signed_raw_tx)` | Broadcast a PRE-SIGNED raw tx. No key handling. |
| `name_hash(subject)` | keccak256 nameHash helper (offline). |

## Source

- `rnx_sdk/__init__.py` — client + real implementations.
- `example.py` — read-only usage example (placeholder RPC).
- `pyproject.toml` — build config (dep: `web3`).

## Verification

- `python -m py_compile rnx_sdk/__init__.py example.py` — compiles clean.
- `python example.py` — runs read-only; fails closed on an unreachable
  placeholder RPC (never fabricates data).

## Safety

- Never commit private keys. Testnet keys are developer-held only.
- Writes are PRE-SIGNED raw txs only; the SDK never signs or holds keys.
- No real transactions, deploys, or production changes result from this SDK.
