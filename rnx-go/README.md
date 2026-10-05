# rnx-sdk (Go)

A minimal, **dependency-free** Go SDK for the RNX EVM network. Read access is
over plain EVM JSON-RPC using only the standard library (`net/http` +
`encoding/json`). **No `go-ethereum`, no third-party modules.**

- **Read-only by design.** All read methods query the configured JSON-RPC
  endpoint directly.
- **Pre-signed writes only.** The single write method, `SendRawTransaction`,
  broadcasts a caller-supplied, already-signed transaction. The SDK **never
  constructs, holds, derives, or touches private keys.**
- **Fail-closed chain safety.** `CheckChain` errors on any chain-id mismatch and
  distinguishes the legacy chain (`12345`) from the reserved RNX chain
  (`194151`).

## Install

```bash
go get rnx-sdk
```

Module path: `rnx-sdk` · Go `1.21+`.

## Chain constants

| Constant         | Value              | Meaning                              |
| ---------------- | ------------------ | ------------------------------------ |
| `RNX_CHAIN_ID`   | `194151`           | Reserved future RNX chain (decimal)  |
| `RNXChainIDHex`  | `0x2f667`          | Same, hex quantity (eth_chainId form)|
| `LegacyChainID`  | `12345`            | Pre-migration legacy chain           |

## Quick start (read-only)

```go
package main

import (
    "context"
    "fmt"
    "log"

    rnx "rnx-sdk"
)

func main() {
    ctx := context.Background()

    c, err := rnx.New("https://rpc.rnx.example/placeholder") // replace with your RPC
    if err != nil {
        log.Fatal(err)
    }

    // Fail-closed: refuse to proceed unless we're really on RNX (194151).
    if err := c.CheckChain(ctx); err != nil {
        log.Fatal(err) // precise error if pointed at legacy 12345
    }

    id, _ := c.ChainID(ctx)
    bn, _ := c.BlockNumber(ctx)
    fmt.Printf("chain=%s height=%s\n", id, bn)

    bal, _ := c.GetBalance(ctx, "0x0000000000000000000000000000000000000001")
    fmt.Printf("balance=%s wei\n", bal)
}
```

## API

| Method | RPC | Notes |
| ------ | --- | ----- |
| `New(rpcURL, ...Option)` | — | No network call at construction |
| `ChainID(ctx)` | `eth_chainId` | returns `*big.Int` |
| `BlockNumber(ctx)` | `eth_blockNumber` | returns `*big.Int` |
| `GetBalance(ctx, addr)` | `eth_getBalance` | latest block |
| `GetBlock(ctx, n)` | `eth_getBlockByNumber` | nil `n` = latest |
| `GetTransaction(ctx, hash)` | `eth_getTransactionByHash` | raw JSON |
| `Call(ctx, to, data)` | `eth_call` | read-only, never signs |
| `VeridexVerify(ctx, registryAddr, nameHash)` | `eth_call` → `currentAnchor(bytes32)` | fail-closed status from on-chain state |
| `CheckChain(ctx)` | `eth_chainId` | fail-closed; distinguishes legacy 12345 |
| `SendRawTransaction(ctx, rawHex)` | `eth_sendRawTransaction` | **pre-signed only; no keys** |

### VERIDEX verification

`VeridexVerify` `eth_call`s the registry's `currentAnchor(bytes32)` view
(selector `0x6495ef8a`) and decodes
`(bytes32 commitmentHash, address owner, uint256 stateVersion, uint256 blockNumber, bool exists)`.
`VerificationStatus` is `"OK"` only when the on-chain `exists` flag is set,
otherwise `"NO_ANCHOR"` — the status is **never synthesized** and short/empty
returns are treated as `NO_ANCHOR` (fail-closed). The `nameHash` must be the
0x-prefixed 32-byte keccak256 nameHash of the subject, computed by the caller;
this SDK performs no hashing and needs no key.

### Pre-signed writes

```go
// You sign offline (hardware wallet / offline signer) and pass the raw blob.
rawSigned := "0x02f8..." // produced elsewhere; the SDK never sees your key
txHash, err := c.SendRawTransaction(ctx, rawSigned)
```

## Testing

```bash
go vet ./...
go test ./...
```

Tests are hermetic: a table-driven unit test covers chain-id hex↔dec parsing,
and an `httptest.Server` mock-RPC exercises `ChainID`, `BlockNumber`,
`GetBlock`, `CheckChain`, `VeridexVerify`, and `SendRawTransaction`. **No real
network endpoints are contacted and no keys are used.**

## Security posture

- stdlib-only read client (no heavy deps / supply-chain surface)
- no deploy, no private keys, no real transactions originated by the SDK
- chain id `194151` (future) vs `12345` (legacy) enforced fail-closed
- `qUSD` is a test asset; no hidden mint / no admin drain in the broader RNX
  contract set this SDK talks to
