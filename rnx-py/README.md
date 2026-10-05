# rnx-sdk — RNX Testnet Python SDK (skeleton)

> ⚠️ **TESTNET ONLY — NOT REAL MONEY — NOT MAINNET.** This is an SDK *skeleton*.
> Every method has a real, stable signature with docstrings, but implementations
> raise `NotImplementedError` so no fake on-chain data is ever returned.
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
pip install rnx-sdk web3
```

Minimal dependency surface: only [`web3`](https://web3py.readthedocs.io/).

## Usage

```python
from rnx_sdk import RnxClient, RNX_CHAIN_ID

client = RnxClient(
    rpc_url="<PLACEHOLDER_TESTNET_RPC_URL>",
    chain_id=RNX_CHAIN_ID,  # 194151
    veridex_endpoint="<PLACEHOLDER_VERIDEX_ENDPOINT>",
)

# Each call raises NotImplementedError until wired to real infrastructure:
# w3  = client.connect()
# bal = client.get_balance("0x...")
# tx  = client.send_transaction(account, {"to": "0x...", "value": 0})
# blk = client.get_block("latest")
# t   = client.get_transaction("0x...")
# c   = client.contract("0x...", abi)
# v   = client.veridex_verify("rakshanex.bihar")
```

## API surface (stubs)

| Method | Purpose |
|--------|---------|
| `connect()` | Return a configured `web3.Web3` for RNX Testnet. |
| `get_balance(address)` | Native RNX (test) balance in wei. |
| `send_transaction(account, tx)` | Sign + broadcast a testnet transaction. |
| `get_block(block_identifier)` | Fetch a block. |
| `get_transaction(tx_hash)` | Fetch a transaction. |
| `contract(address, abi)` | Bind a web3 contract. |
| `veridex_verify(subject)` | Verify a name/proof via VERIDEX (fail-closed). |

## Source

- `rnx_sdk/__init__.py` — client + typed stubs.

## Safety

- Never commit private keys. Testnet keys are developer-held only.
- No real transactions, deploys, or production changes result from this skeleton.
