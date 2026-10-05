/**
 * @rnx/sdk — RNX Testnet JavaScript/TypeScript SDK
 *
 * ⚠️ TESTNET ONLY — NOT REAL MONEY — NOT MAINNET.
 *
 * Real, working implementations against an EVM JSON-RPC endpoint using
 * `ethers` (v6). Read methods query the chain directly. WRITE methods accept a
 * PRE-SIGNED raw transaction ONLY — this SDK never handles, stores, derives, or
 * requests private keys. Signing happens entirely in the caller's wallet/signer
 * out of band; the SDK only broadcasts the already-signed bytes.
 *
 * Honesty invariants (do not violate):
 *  - RNX mainnet is NOT live. Chain ID 194151 (0x2F667) is the reserved FUTURE
 *    mainnet id and the testnet rehearsal id — engineering foundation only.
 *  - Decentralization = BOOTSTRAP (0 independent operators; operator-run).
 *    Never claim decentralized / trustless / permissionless.
 *  - qUSD is a TEST quote asset, NOT USDT, NOT USDC.
 *  - Never fabricate validators/users/TVL/liquidity/volume/followers. Every
 *    value returned here comes directly from the configured RPC — nothing is
 *    synthesized.
 *  - Legacy chain 12345 is unchanged.
 *
 * Dependency: `ethers` (v6). Keep dependencies minimal.
 */

import { ethers } from "ethers";

/** Reserved future mainnet / testnet rehearsal chain id (decimal). */
export const RNX_CHAIN_ID = 194151;
/** Same chain id in hex (0x2F667). */
export const RNX_CHAIN_ID_HEX = "0x2F667";

/**
 * ABI fragment for the registry's `currentAnchor` view used by
 * {@link RnxClient.veridexVerify}. Matches
 * `contracts/artifacts/BiharDomainRegistry.abi.json`
 * (selector 0x6495ef8a).
 */
export const CURRENT_ANCHOR_ABI = [
  {
    inputs: [{ internalType: "bytes32", name: "nameHash", type: "bytes32" }],
    name: "currentAnchor",
    outputs: [
      { internalType: "bytes32", name: "commitmentHash", type: "bytes32" },
      { internalType: "address", name: "owner", type: "address" },
      { internalType: "uint256", name: "stateVersion", type: "uint256" },
      { internalType: "uint256", name: "blockNumber", type: "uint256" },
      { internalType: "bool", name: "exists", type: "bool" },
    ],
    stateMutability: "view",
    type: "function",
  },
] as const;

/** Configuration for an {@link RnxClient}. */
export interface RnxClientConfig {
  /** JSON-RPC endpoint for RNX Testnet. */
  rpcUrl: string;
  /** Expected chain id. Defaults to {@link RNX_CHAIN_ID} (194151). */
  chainId?: number;
  /**
   * Address of the deployed VERIDEX anchor registry contract (the contract
   * exposing `currentAnchor`). Required for {@link RnxClient.veridexVerify}.
   */
  registryAddress?: string;
}

/**
 * Result of a VERIDEX verification. Values are read directly from the registry
 * contract via `currentAnchor` — never synthesized.
 */
export interface VeridexVerifyResult {
  /**
   * Fail-closed status derived ONLY from on-chain state:
   *  - "OK"        → an anchor exists for the subject's nameHash.
   *  - "NO_ANCHOR" → no anchor is recorded on-chain (fail-closed).
   */
  verificationStatus: "OK" | "NO_ANCHOR";
  /** The name/subject that was verified. */
  subject: string;
  /** keccak256 nameHash queried on-chain (0x-prefixed). */
  nameHash: string;
  /** On-chain commitment hash, or null when no anchor exists. */
  commitmentHash: string | null;
  /** Recorded owner address, or null when no anchor exists. */
  owner: string | null;
  /** Monotonic state version, or null when no anchor exists. */
  stateVersion: bigint | null;
  /** Block number recorded with the anchor, or null when none exists. */
  blockNumber: bigint | null;
}

/**
 * Compute the keccak256 nameHash for a subject string.
 *
 * This mirrors the registry's v2 nameHash input convention (keccak256 of the
 * UTF-8 subject). Exposed so callers can precompute/verify the same value.
 *
 * @param subject the name or proof subject (e.g. "rakshanex.bihar").
 * @returns the 0x-prefixed 32-byte keccak256 hash.
 */
export function nameHash(subject: string): string {
  return ethers.keccak256(ethers.toUtf8Bytes(subject));
}

/**
 * RNX Testnet client.
 *
 * Read methods query the configured JSON-RPC provider directly. Write access is
 * limited to {@link RnxClient.sendRawTransaction}, which broadcasts a
 * caller-provided PRE-SIGNED transaction. The SDK never touches private keys.
 */
export class RnxClient {
  readonly rpcUrl: string;
  readonly chainId: number;
  readonly registryAddress?: string;

  private _provider: ethers.JsonRpcProvider | null = null;

  constructor(config: RnxClientConfig) {
    if (!config || typeof config.rpcUrl !== "string" || !config.rpcUrl) {
      throw new TypeError("RnxClientConfig.rpcUrl is required.");
    }
    this.rpcUrl = config.rpcUrl;
    this.chainId = config.chainId ?? RNX_CHAIN_ID;
    this.registryAddress = config.registryAddress;
  }

  /**
   * Connect to the RNX Testnet RPC and return an ethers provider.
   *
   * The provider is created once and cached. A static network is supplied so
   * ethers does not auto-detect a different chain id.
   *
   * @returns an `ethers.JsonRpcProvider` bound to the configured RPC + chain id.
   */
  connect(): ethers.JsonRpcProvider {
    if (this._provider === null) {
      const network = ethers.Network.from(this.chainId);
      this._provider = new ethers.JsonRpcProvider(this.rpcUrl, network, {
        staticNetwork: network,
      });
    }
    return this._provider;
  }

  /**
   * Get the native RNX (test) balance of an address, in wei.
   *
   * @param address 0x-prefixed account address.
   * @param blockTag optional block tag (defaults to "latest").
   * @returns the balance as a bigint (wei). Test RNX has no monetary value.
   */
  async getBalance(
    address: string,
    blockTag: ethers.BlockTag = "latest",
  ): Promise<bigint> {
    if (!ethers.isAddress(address)) {
      throw new TypeError(`Invalid address: ${address}`);
    }
    return this.connect().getBalance(address, blockTag);
  }

  /**
   * Broadcast a PRE-SIGNED raw transaction.
   *
   * SECURITY: this SDK never handles private keys. The caller must sign the
   * transaction out of band (hardware wallet, browser wallet, or an
   * eth-account/ethers Wallet the caller controls) and pass the serialized,
   * already-signed bytes here. This method only submits those bytes via
   * `eth_sendRawTransaction`.
   *
   * @param signedRawTx 0x-prefixed, fully-signed, serialized transaction.
   * @returns the transaction response once accepted by the node.
   */
  async sendRawTransaction(
    signedRawTx: string,
  ): Promise<ethers.TransactionResponse> {
    if (typeof signedRawTx !== "string" || !signedRawTx.startsWith("0x")) {
      throw new TypeError(
        "sendRawTransaction expects a 0x-prefixed PRE-SIGNED raw transaction. " +
          "This SDK never signs or handles private keys.",
      );
    }
    return this.connect().broadcastTransaction(signedRawTx);
  }

  /**
   * Fetch a block by number or tag.
   *
   * @param blockHashOrTag a block number, hash, or tag such as "latest".
   * @param prefetchTxs whether to include full transaction objects.
   * @returns the block, or null if not found.
   */
  async getBlock(
    blockHashOrTag: ethers.BlockTag | string,
    prefetchTxs = false,
  ): Promise<ethers.Block | null> {
    return this.connect().getBlock(blockHashOrTag, prefetchTxs);
  }

  /**
   * Fetch a transaction by hash.
   *
   * @param txHash 0x-prefixed transaction hash.
   * @returns the transaction, or null if not found.
   */
  async getTransaction(
    txHash: string,
  ): Promise<ethers.TransactionResponse | null> {
    return this.connect().getTransaction(txHash);
  }

  /**
   * Perform a read-only contract call (`eth_call`). No state is mutated and no
   * signature is required.
   *
   * @param tx a call request (to/data/...).
   * @param blockTag optional block tag (defaults to "latest").
   * @returns the 0x-prefixed ABI-encoded return data.
   */
  async call(
    tx: ethers.TransactionRequest,
    blockTag: ethers.BlockTag = "latest",
  ): Promise<string> {
    return this.connect().call({ ...tx, blockTag });
  }

  /**
   * Build an ethers Contract bound to this client's provider for READS, or to a
   * caller-supplied signer/runner for writes.
   *
   * NOTE: if you pass a signer, signing is the caller's responsibility and
   * happens in the caller's wallet — this SDK still never sees private keys. For
   * read-only usage, omit `runner` and the cached provider is used.
   *
   * @param address deployed contract address (0x-prefixed).
   * @param abi contract ABI fragments.
   * @param runner optional signer/provider; defaults to the read provider.
   * @returns an `ethers.Contract` instance.
   */
  contract(
    address: string,
    abi: ethers.InterfaceAbi,
    runner?: ethers.ContractRunner,
  ): ethers.Contract {
    if (!ethers.isAddress(address)) {
      throw new TypeError(`Invalid contract address: ${address}`);
    }
    return new ethers.Contract(address, abi, runner ?? this.connect());
  }

  /**
   * Verify a name/proof by reading `currentAnchor` on the VERIDEX registry
   * contract. This is a READ-ONLY on-chain query — fail-closed: if no anchor
   * exists on-chain, status is "NO_ANCHOR". No value is ever fabricated.
   *
   * @param subject the name or proof subject (e.g. "rakshanex.bihar").
   * @param registryAddress optional override of the configured registry address.
   * @returns a {@link VeridexVerifyResult} sourced entirely from on-chain state.
   */
  async veridexVerify(
    subject: string,
    registryAddress?: string,
  ): Promise<VeridexVerifyResult> {
    const addr = registryAddress ?? this.registryAddress;
    if (!addr || !ethers.isAddress(addr)) {
      throw new TypeError(
        "veridexVerify requires a valid registry contract address " +
          "(config.registryAddress or the registryAddress argument).",
      );
    }
    const nh = nameHash(subject);
    const registry = this.contract(addr, CURRENT_ANCHOR_ABI as unknown as ethers.InterfaceAbi);
    const result = await registry.currentAnchor(nh);

    // ethers returns a Result tuple matching the ABI outputs.
    const commitmentHash: string = result[0];
    const owner: string = result[1];
    const stateVersion: bigint = result[2];
    const blockNumber: bigint = result[3];
    const exists: boolean = result[4];

    if (!exists) {
      return {
        verificationStatus: "NO_ANCHOR",
        subject,
        nameHash: nh,
        commitmentHash: null,
        owner: null,
        stateVersion: null,
        blockNumber: null,
      };
    }
    return {
      verificationStatus: "OK",
      subject,
      nameHash: nh,
      commitmentHash,
      owner,
      stateVersion,
      blockNumber,
    };
  }
}
