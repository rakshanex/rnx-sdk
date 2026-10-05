"""rnx-sdk — read-only usage example (TESTNET ONLY — NOT REAL MONEY).

This example performs ONLY read operations against a placeholder testnet RPC.
It never signs or broadcasts a transaction and never handles a private key.

Run (after `pip install -e .` or `pip install web3`):
    RNX_RPC_URL="<testnet-rpc>" RNX_REGISTRY="0x..." python example.py

With the default placeholder RPC the network calls will simply fail to connect
— which is expected; this file demonstrates the API shape and the fail-closed
behavior, not live data.
"""

from __future__ import annotations

import os

from rnx_sdk import RNX_CHAIN_ID, RNX_CHAIN_ID_HEX, RnxClient, name_hash

# Placeholder testnet RPC. Override with the RNX_RPC_URL env var.
RPC_URL = os.environ.get("RNX_RPC_URL", "https://rpc.rnx-testnet.invalid")
# Placeholder registry address. Override with RNX_REGISTRY (required for verify).
REGISTRY = os.environ.get("RNX_REGISTRY", "")


def main() -> None:
    print("RNX SDK example — TESTNET ONLY, read-only. No real money.")
    print(f"chainId={RNX_CHAIN_ID} ({RNX_CHAIN_ID_HEX})")

    client = RnxClient(
        rpc_url=RPC_URL,
        chain_id=RNX_CHAIN_ID,
        registry_address=REGISTRY or None,
    )

    # Deterministic, offline: name_hash does not touch the network.
    print(f'name_hash("rakshanex.bihar") = {name_hash("rakshanex.bihar")}')

    # --- Read-only chain queries (require a reachable RPC) ---
    try:
        block = client.get_block("latest")
        print(f"latest block number: {block.get('number', 'n/a')}")

        addr = "0x0000000000000000000000000000000000000000"
        bal = client.get_balance(addr)
        print(f"balance({addr}) = {bal} wei (test RNX, no value)")

        if REGISTRY:
            # Fail-closed VERIDEX verify via currentAnchor (read-only).
            result = client.veridex_verify("rakshanex.bihar")
            print(f"veridex_verify: {result}")
        else:
            print("Set RNX_REGISTRY to run veridex_verify (currentAnchor read).")
    except Exception as err:  # noqa: BLE001 — demo: expected with placeholder RPC
        print(f"read query failed (expected with placeholder RPC): {err}")

    # NOTE: writes use client.send_raw_transaction(signed_raw_tx) with a
    # PRE-SIGNED transaction produced by YOUR account. This SDK never handles
    # private keys, so no write is performed in this read-only example.


if __name__ == "__main__":
    main()
