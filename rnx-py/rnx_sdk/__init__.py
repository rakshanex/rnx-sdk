"""rnx-sdk — RNX Testnet Python SDK (SKELETON).

⚠️ TESTNET ONLY — NOT REAL MONEY — NOT MAINNET.

This is a stub SDK. Every method below has a real, stable signature and a
docstring, but the implementations are intentionally not wired up: they raise
``NotImplementedError`` so that no fake data or fabricated on-chain state can be
returned. Endpoints are placeholders until the RNX Testnet infrastructure is
provisioned.

Honesty invariants (do not violate):
  - RNX mainnet is NOT live. Chain ID 194151 (0x2F667) is the reserved FUTURE
    mainnet id and the testnet rehearsal id — engineering foundation only.
  - Decentralization = BOOTSTRAP (0 independent operators; operator-run).
  - qUSD is a TEST quote asset, NOT USDT, NOT USDC.
  - Never fabricate validators/users/TVL/liquidity/volume/followers.
  - Legacy chain 12345 is unchanged.

Dependency: ``web3`` (web3.py). Keep dependencies minimal.
"""

from __future__ import annotations

from dataclasses import dataclass
from typing import Any, Optional

__all__ = [
    "RNX_CHAIN_ID",
    "RNX_CHAIN_ID_HEX",
    "RnxClient",
    "VeridexVerifyResult",
]

#: Reserved future mainnet / testnet rehearsal chain id (decimal).
RNX_CHAIN_ID: int = 194151
#: Same chain id in hex (0x2F667).
RNX_CHAIN_ID_HEX: str = "0x2F667"


@dataclass
class VeridexVerifyResult:
    """Result of a VERIDEX verification.

    The shape is stable; values come from the real VERIDEX layer once wired —
    they are never synthesized by this skeleton.

    Attributes:
        verification_status: e.g. "OK" | "STALE_PROOF" | "MISMATCH" |
            "UNTRUSTED" (fail-closed).
        subject: the name/subject that was verified.
    """

    verification_status: str
    subject: str


class RnxClient:
    """RNX Testnet client (skeleton).

    Construct with placeholder endpoints; call sites are stable even though the
    methods raise :class:`NotImplementedError` until implemented.
    """

    def __init__(
        self,
        rpc_url: str,
        chain_id: int = RNX_CHAIN_ID,
        veridex_endpoint: Optional[str] = None,
    ) -> None:
        """Initialize the client.

        Args:
            rpc_url: JSON-RPC endpoint for RNX Testnet (placeholder until
                provisioned).
            chain_id: expected chain id; defaults to :data:`RNX_CHAIN_ID`
                (194151).
            veridex_endpoint: VERIDEX verification endpoint (placeholder until
                provisioned).
        """
        self.rpc_url = rpc_url
        self.chain_id = chain_id
        self.veridex_endpoint = veridex_endpoint

    def connect(self) -> "Any":
        """Connect to the RNX Testnet RPC and return a ``web3.Web3`` instance.

        Returns:
            A ``web3.Web3`` bound to the configured RPC + chain id.

        Raises:
            NotImplementedError: until wired to a provisioned endpoint.
        """
        raise NotImplementedError(
            "RnxClient.connect is a TESTNET SKELETON stub and is not "
            "implemented yet. No data is returned to avoid fabricating "
            "on-chain state."
        )

    def get_balance(self, address: str) -> int:
        """Get the native RNX (test) balance of an address, in wei.

        Args:
            address: 0x-prefixed account address.

        Returns:
            The balance in wei. Test RNX has no monetary value.

        Raises:
            NotImplementedError: until implemented.
        """
        raise NotImplementedError(
            "RnxClient.get_balance is a TESTNET SKELETON stub and is not "
            "implemented yet."
        )

    def send_transaction(self, account: "Any", tx: dict) -> str:
        """Sign and broadcast a transaction on RNX Testnet.

        Args:
            account: a web3/eth-account signer holding a TESTNET key (never
                commit keys).
            tx: a transaction dict (to/value/data/...). ``chainId`` defaults to
                194151 for EIP-155 replay protection.

        Returns:
            The 0x-prefixed transaction hash once broadcast.

        Raises:
            NotImplementedError: until implemented.
        """
        raise NotImplementedError(
            "RnxClient.send_transaction is a TESTNET SKELETON stub and is not "
            "implemented yet."
        )

    def get_block(self, block_identifier: "Any") -> "Any":
        """Fetch a block by number, hash, or tag.

        Args:
            block_identifier: a block number, hash, or tag such as "latest".

        Returns:
            The block mapping, or ``None`` if not found.

        Raises:
            NotImplementedError: until implemented.
        """
        raise NotImplementedError(
            "RnxClient.get_block is a TESTNET SKELETON stub and is not "
            "implemented yet."
        )

    def get_transaction(self, tx_hash: str) -> "Any":
        """Fetch a transaction by hash.

        Args:
            tx_hash: 0x-prefixed transaction hash.

        Returns:
            The transaction mapping, or ``None`` if not found.

        Raises:
            NotImplementedError: until implemented.
        """
        raise NotImplementedError(
            "RnxClient.get_transaction is a TESTNET SKELETON stub and is not "
            "implemented yet."
        )

    def contract(self, address: str, abi: list) -> "Any":
        """Bind a web3 contract to this client's provider.

        Args:
            address: deployed contract address (0x-prefixed).
            abi: contract ABI (list of fragments).

        Returns:
            A ``web3`` contract instance.

        Raises:
            NotImplementedError: until implemented.
        """
        raise NotImplementedError(
            "RnxClient.contract is a TESTNET SKELETON stub and is not "
            "implemented yet."
        )

    def veridex_verify(self, subject: str) -> VeridexVerifyResult:
        """Verify a name/proof through the VERIDEX verification layer.

        Args:
            subject: the name or proof subject (e.g. "rakshanex.bihar").

        Returns:
            A :class:`VeridexVerifyResult` from VERIDEX (fail-closed).

        Raises:
            NotImplementedError: until wired to a VERIDEX endpoint.
        """
        raise NotImplementedError(
            "RnxClient.veridex_verify is a TESTNET SKELETON stub and is not "
            "implemented yet."
        )
