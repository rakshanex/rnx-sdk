"""rnx-sdk — RNX Testnet Python SDK.

⚠️ TESTNET ONLY — NOT REAL MONEY — NOT MAINNET.

Real, working implementations against an EVM JSON-RPC endpoint using
``web3`` (web3.py). Read methods query the chain directly. WRITE methods accept
a PRE-SIGNED raw transaction ONLY — this SDK never handles, stores, derives, or
requests private keys. Signing happens entirely in the caller's wallet/account
out of band; the SDK only broadcasts the already-signed bytes.

Honesty invariants (do not violate):
  - RNX mainnet is NOT live. Chain ID 194151 (0x2F667) is the reserved FUTURE
    mainnet id and the testnet rehearsal id — engineering foundation only.
  - Decentralization = BOOTSTRAP (0 independent operators; operator-run). Never
    claim decentralized / trustless / permissionless.
  - qUSD is a TEST quote asset, NOT USDT, NOT USDC.
  - Never fabricate validators/users/TVL/liquidity/volume/followers. Every
    value returned here comes directly from the configured RPC — nothing is
    synthesized.
  - Legacy chain 12345 is unchanged.

Dependency: ``web3`` (web3.py >= 6). Keep dependencies minimal.
"""

from __future__ import annotations

from dataclasses import dataclass
from typing import Any, Optional

from web3 import Web3

__all__ = [
    "RNX_CHAIN_ID",
    "RNX_CHAIN_ID_HEX",
    "CURRENT_ANCHOR_ABI",
    "RnxClient",
    "VeridexVerifyResult",
    "name_hash",
]

#: Reserved future mainnet / testnet rehearsal chain id (decimal).
RNX_CHAIN_ID: int = 194151
#: Same chain id in hex (0x2F667).
RNX_CHAIN_ID_HEX: str = "0x2F667"

#: ABI fragment for the registry's ``currentAnchor`` view used by
#: :meth:`RnxClient.veridex_verify`. Matches
#: ``contracts/artifacts/BiharDomainRegistry.abi.json`` (selector 0x6495ef8a).
CURRENT_ANCHOR_ABI: list = [
    {
        "inputs": [
            {"internalType": "bytes32", "name": "nameHash", "type": "bytes32"}
        ],
        "name": "currentAnchor",
        "outputs": [
            {"internalType": "bytes32", "name": "commitmentHash", "type": "bytes32"},
            {"internalType": "address", "name": "owner", "type": "address"},
            {"internalType": "uint256", "name": "stateVersion", "type": "uint256"},
            {"internalType": "uint256", "name": "blockNumber", "type": "uint256"},
            {"internalType": "bool", "name": "exists", "type": "bool"},
        ],
        "stateMutability": "view",
        "type": "function",
    }
]


def name_hash(subject: str) -> str:
    """Compute the keccak256 nameHash for a subject string.

    Mirrors the registry's v2 nameHash input convention (keccak256 of the
    UTF-8 subject). Exposed so callers can precompute/verify the same value.

    Args:
        subject: the name or proof subject (e.g. "rakshanex.bihar").

    Returns:
        The 0x-prefixed 32-byte keccak256 hash.
    """
    return Web3.keccak(text=subject).hex()


@dataclass
class VeridexVerifyResult:
    """Result of a VERIDEX verification.

    Values are read directly from the registry contract via ``currentAnchor`` —
    never synthesized.

    Attributes:
        verification_status: fail-closed status derived ONLY from on-chain
            state: "OK" when an anchor exists, "NO_ANCHOR" otherwise.
        subject: the name/subject that was verified.
        name_hash: keccak256 nameHash queried on-chain (0x-prefixed).
        commitment_hash: on-chain commitment hash, or None when no anchor exists.
        owner: recorded owner address, or None when no anchor exists.
        state_version: monotonic state version, or None when no anchor exists.
        block_number: block number recorded with the anchor, or None.
    """

    verification_status: str
    subject: str
    name_hash: str
    commitment_hash: Optional[str]
    owner: Optional[str]
    state_version: Optional[int]
    block_number: Optional[int]


class RnxClient:
    """RNX Testnet client.

    Read methods query the configured JSON-RPC provider directly. Write access
    is limited to :meth:`send_raw_transaction`, which broadcasts a
    caller-provided PRE-SIGNED transaction. The SDK never touches private keys.
    """

    def __init__(
        self,
        rpc_url: str,
        chain_id: int = RNX_CHAIN_ID,
        registry_address: Optional[str] = None,
    ) -> None:
        """Initialize the client.

        Args:
            rpc_url: JSON-RPC endpoint for RNX Testnet.
            chain_id: expected chain id; defaults to :data:`RNX_CHAIN_ID`
                (194151).
            registry_address: address of the deployed VERIDEX anchor registry
                contract (exposing ``currentAnchor``). Required for
                :meth:`veridex_verify`.
        """
        if not rpc_url or not isinstance(rpc_url, str):
            raise ValueError("rpc_url is required.")
        self.rpc_url = rpc_url
        self.chain_id = chain_id
        self.registry_address = registry_address
        self._w3: Optional[Web3] = None

    def connect(self) -> Web3:
        """Connect to the RNX Testnet RPC and return a ``web3.Web3`` instance.

        The instance is created once and cached.

        Returns:
            A ``web3.Web3`` bound to the configured RPC endpoint.
        """
        if self._w3 is None:
            self._w3 = Web3(Web3.HTTPProvider(self.rpc_url))
        return self._w3

    def get_balance(self, address: str, block_identifier: Any = "latest") -> int:
        """Get the native RNX (test) balance of an address, in wei.

        Args:
            address: 0x-prefixed account address.
            block_identifier: block tag/number (defaults to "latest").

        Returns:
            The balance in wei. Test RNX has no monetary value.
        """
        w3 = self.connect()
        checksummed = Web3.to_checksum_address(address)
        return int(w3.eth.get_balance(checksummed, block_identifier))

    def send_raw_transaction(self, signed_raw_tx: str) -> str:
        """Broadcast a PRE-SIGNED raw transaction.

        SECURITY: this SDK never handles private keys. The caller must sign the
        transaction out of band (hardware wallet, external account, or an
        eth-account ``Account`` the caller controls) and pass the serialized,
        already-signed bytes here. This method only submits those bytes via
        ``eth_sendRawTransaction``.

        Args:
            signed_raw_tx: 0x-prefixed, fully-signed, serialized transaction.

        Returns:
            The 0x-prefixed transaction hash accepted by the node.
        """
        if not isinstance(signed_raw_tx, str) or not signed_raw_tx.startswith("0x"):
            raise ValueError(
                "send_raw_transaction expects a 0x-prefixed PRE-SIGNED raw "
                "transaction. This SDK never signs or handles private keys."
            )
        w3 = self.connect()
        tx_hash = w3.eth.send_raw_transaction(signed_raw_tx)
        return tx_hash.hex()

    def get_block(self, block_identifier: Any = "latest", full_transactions: bool = False) -> Any:
        """Fetch a block by number, hash, or tag.

        Args:
            block_identifier: a block number, hash, or tag such as "latest".
            full_transactions: include full transaction objects when True.

        Returns:
            The block mapping (``web3`` ``AttributeDict``).
        """
        w3 = self.connect()
        return w3.eth.get_block(block_identifier, full_transactions=full_transactions)

    def get_transaction(self, tx_hash: str) -> Any:
        """Fetch a transaction by hash.

        Args:
            tx_hash: 0x-prefixed transaction hash.

        Returns:
            The transaction mapping (``web3`` ``AttributeDict``).
        """
        w3 = self.connect()
        return w3.eth.get_transaction(tx_hash)

    def call(self, tx: dict, block_identifier: Any = "latest") -> bytes:
        """Perform a read-only contract call (``eth_call``).

        No state is mutated and no signature is required.

        Args:
            tx: a call request dict (to/data/...).
            block_identifier: block tag/number (defaults to "latest").

        Returns:
            The raw ABI-encoded return bytes.
        """
        w3 = self.connect()
        return w3.eth.call(tx, block_identifier)

    def contract(self, address: str, abi: list) -> Any:
        """Bind a web3 contract to this client's provider for reads.

        Args:
            address: deployed contract address (0x-prefixed).
            abi: contract ABI (list of fragments).

        Returns:
            A ``web3`` contract instance.
        """
        w3 = self.connect()
        return w3.eth.contract(address=Web3.to_checksum_address(address), abi=abi)

    def veridex_verify(
        self, subject: str, registry_address: Optional[str] = None
    ) -> VeridexVerifyResult:
        """Verify a name/proof by reading ``currentAnchor`` on the registry.

        READ-ONLY on-chain query — fail-closed: if no anchor exists on-chain,
        the status is "NO_ANCHOR". No value is ever fabricated.

        Args:
            subject: the name or proof subject (e.g. "rakshanex.bihar").
            registry_address: optional override of the configured registry
                address.

        Returns:
            A :class:`VeridexVerifyResult` sourced entirely from on-chain state.
        """
        addr = registry_address or self.registry_address
        if not addr:
            raise ValueError(
                "veridex_verify requires a registry contract address "
                "(registry_address constructor arg or method argument)."
            )
        nh = name_hash(subject)
        registry = self.contract(addr, CURRENT_ANCHOR_ABI)
        commitment_hash, owner, state_version, block_number, exists = (
            registry.functions.currentAnchor(Web3.to_bytes(hexstr=nh)).call()
        )

        if not exists:
            return VeridexVerifyResult(
                verification_status="NO_ANCHOR",
                subject=subject,
                name_hash=nh,
                commitment_hash=None,
                owner=None,
                state_version=None,
                block_number=None,
            )
        commitment_hex = (
            commitment_hash.hex()
            if isinstance(commitment_hash, (bytes, bytearray))
            else str(commitment_hash)
        )
        if not commitment_hex.startswith("0x"):
            commitment_hex = "0x" + commitment_hex
        return VeridexVerifyResult(
            verification_status="OK",
            subject=subject,
            name_hash=nh,
            commitment_hash=commitment_hex,
            owner=owner,
            state_version=int(state_version),
            block_number=int(block_number),
        )
