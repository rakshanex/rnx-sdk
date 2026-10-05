/**
 * @rnx/sdk — RNX Testnet JavaScript/TypeScript SDK (SKELETON)
 *
 * ⚠️ TESTNET ONLY — NOT REAL MONEY — NOT MAINNET.
 *
 * This is a stub SDK. Every method below has a real, stable signature and a
 * docstring, but the implementations are intentionally not wired up: they throw
 * `NotImplementedError` so that no fake data or fabricated on-chain state can be
 * returned. Endpoints are placeholders until the RNX Testnet infrastructure is
 * provisioned.
 *
 * Honesty invariants (do not violate):
 *  - RNX mainnet is NOT live. Chain ID 194151 (0x2F667) is the reserved FUTURE
 *    mainnet id and the testnet rehearsal id — engineering foundation only.
 *  - Decentralization = BOOTSTRAP (0 independent operators; operator-run).
 *  - qUSD is a TEST quote asset, NOT USDT, NOT USDC.
 *  - Never fabricate validators/users/TVL/liquidity/volume/followers.
 *  - Legacy chain 12345 is unchanged.
 *
 * Dependency: `ethers` (v6). Keep dependencies minimal.
 */

import { ethers } from "ethers";

/** Reserved future mainnet / testnet rehearsal chain id (decimal). */
export const RNX_CHAIN_ID = 194151;
/** Same chain id in hex (0x2F667). */
export const RNX_CHAIN_ID_HEX = "0x2F667";

/** Thrown by every stub until the corresponding feature is implemented. */
export class NotImplementedError extends Error {
  constructor(method: string) {
    super(
      `${method} is a TESTNET SKELETON stub and is not implemented yet. ` +
        `No data is returned to avoid fabricating on-chain state.`,
    );
    this.name = "NotImplementedError";
  }
}

/** Configuration for an {@link RnxClient}. All endpoints are placeholders. */
export interface RnxClientConfig {
  /** JSON-RPC endpoint for RNX Testnet. Placeholder until provisioned. */
  rpcUrl: string;
  /** Expected chain id. Defaults to {@link RNX_CHAIN_ID} (194151). */
  chainId?: number;
  /** VERIDEX verification endpoint. Placeholder until provisioned. */
  veridexEndpoint?: string;
}

/** Result of a VERIDEX verification. Shape is stable; values come from the
 * real VERIDEX layer once wired — never synthesized here. */
export interface VeridexVerifyResult {
  /** e.g. "OK" | "STALE_PROOF" | "MISMATCH" | "UNTRUSTED" (fail-closed). */
  verificationStatus: string;
  /** The name/subject that was verified. */
  subject: string;
}

/**
 * RNX Testnet client (skeleton).
 *
 * Construct with placeholder endpoints; call sites are type-stable even though
 * the methods throw {@link NotImplementedError} until implemented.
 */
export class RnxClient {
  readonly rpcUrl: string;
  readonly chainId: number;
  readonly veridexEndpoint?: string;

  constructor(config: RnxClientConfig) {
    this.rpcUrl = config.rpcUrl;
    this.chainId = config.chainId ?? RNX_CHAIN_ID;
    this.veridexEndpoint = config.veridexEndpoint;
  }

  /**
   * Connect to the RNX Testnet RPC and return an ethers provider.
   *
   * @returns an `ethers.JsonRpcProvider` bound to the configured RPC + chain id.
   * @throws {NotImplementedError} until wired to a provisioned endpoint.
   */
  connect(): ethers.JsonRpcProvider {
    throw new NotImplementedError("RnxClient.connect");
  }

  /**
   * Get the native RNX (test) balance of an address, in wei.
   *
   * @param address 0x-prefixed account address.
   * @returns the balance as a bigint (wei). Test RNX has no monetary value.
   * @throws {NotImplementedError} until implemented.
   */
  async getBalance(address: string): Promise<bigint> {
    void address;
    throw new NotImplementedError("RnxClient.getBalance");
  }

  /**
   * Sign and broadcast a transaction on RNX Testnet.
   *
   * @param signer an ethers signer holding a TESTNET key (never commit keys).
   * @param tx a transaction request (to/value/data/...). chainId defaults to 194151.
   * @returns the transaction response once broadcast.
   * @throws {NotImplementedError} until implemented.
   */
  async sendTransaction(
    signer: ethers.Signer,
    tx: ethers.TransactionRequest,
  ): Promise<ethers.TransactionResponse> {
    void signer;
    void tx;
    throw new NotImplementedError("RnxClient.sendTransaction");
  }

  /**
   * Fetch a block by number or tag.
   *
   * @param blockHashOrTag a block number, hash, or tag such as "latest".
   * @returns the block, or null if not found.
   * @throws {NotImplementedError} until implemented.
   */
  async getBlock(
    blockHashOrTag: ethers.BlockTag | string,
  ): Promise<ethers.Block | null> {
    void blockHashOrTag;
    throw new NotImplementedError("RnxClient.getBlock");
  }

  /**
   * Fetch a transaction by hash.
   *
   * @param txHash 0x-prefixed transaction hash.
   * @returns the transaction, or null if not found.
   * @throws {NotImplementedError} until implemented.
   */
  async getTransaction(
    txHash: string,
  ): Promise<ethers.TransactionResponse | null> {
    void txHash;
    throw new NotImplementedError("RnxClient.getTransaction");
  }

  /**
   * Build an ethers Contract bound to this client's provider/signer.
   *
   * @param address deployed contract address (0x-prefixed).
   * @param abi contract ABI fragments.
   * @param signerOrProvider optional signer (writes) or provider (reads).
   * @returns an `ethers.Contract` instance.
   * @throws {NotImplementedError} until implemented.
   */
  contract(
    address: string,
    abi: ethers.InterfaceAbi,
    signerOrProvider?: ethers.Signer | ethers.Provider,
  ): ethers.Contract {
    void address;
    void abi;
    void signerOrProvider;
    throw new NotImplementedError("RnxClient.contract");
  }

  /**
   * Verify a name/proof through the VERIDEX verification layer.
   *
   * @param subject the name or proof subject (e.g. "rakshanex.bihar").
   * @returns a {@link VeridexVerifyResult} from VERIDEX (fail-closed).
   * @throws {NotImplementedError} until wired to a VERIDEX endpoint.
   */
  async veridexVerify(subject: string): Promise<VeridexVerifyResult> {
    void subject;
    throw new NotImplementedError("RnxClient.veridexVerify");
  }
}
