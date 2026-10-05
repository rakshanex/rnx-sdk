// Package rnx is a minimal, dependency-free Go SDK for the RNX EVM network.
//
// It speaks EVM JSON-RPC over plain net/http + encoding/json. There are NO
// third-party dependencies: go-ethereum is intentionally not required for the
// read path. Write access is PRE-SIGNED ONLY via SendRawTransaction; the SDK
// never constructs, holds, or touches private keys.
//
// Safety posture:
//   - CheckChain() is fail-closed: it errors on any chain id mismatch and
//     explicitly distinguishes the legacy chain (12345) from the reserved
//     future chain (194151).
//   - VeridexVerify() derives its status ONLY from on-chain state returned by
//     the registry's currentAnchor view; it never synthesizes a result.
package rnx

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"time"
)

const (
	// RNX_CHAIN_ID is the reserved future mainnet / testnet-rehearsal chain id
	// (decimal). All read/write operations are expected to target this chain.
	RNX_CHAIN_ID int64 = 194151

	// RNXChainIDHex is RNX_CHAIN_ID encoded as a 0x-prefixed hex quantity,
	// matching eth_chainId's return format.
	RNXChainIDHex = "0x2f667"

	// LegacyChainID is the pre-migration ("legacy") chain id. It is called out
	// explicitly so CheckChain can give a precise, actionable error rather than
	// a generic mismatch when a caller is still pointed at the old network.
	LegacyChainID int64 = 12345

	// currentAnchorSelector is the 4-byte function selector for
	// currentAnchor(bytes32) == keccak256("currentAnchor(bytes32)")[:4].
	// Matches contracts/artifacts/BiharDomainRegistry.abi.json.
	currentAnchorSelector = "0x6495ef8a"
)

// ErrChainMismatch is returned (wrapped) by CheckChain when the connected RPC
// reports a chain id other than RNX_CHAIN_ID.
var ErrChainMismatch = errors.New("rnx: chain id mismatch (fail-closed)")

// Client is a read-mostly RNX JSON-RPC client. It is safe for concurrent use
// by multiple goroutines once constructed.
type Client struct {
	rpcURL string
	http   *http.Client
	nextID func() int
}

// Option configures a Client.
type Option func(*Client)

// WithHTTPClient overrides the default *http.Client (e.g. to set custom
// timeouts, transports, or proxies in tests).
func WithHTTPClient(h *http.Client) Option {
	return func(c *Client) { c.http = h }
}

// New creates a Client bound to the given JSON-RPC endpoint URL. The endpoint
// is used verbatim; no network call is made at construction time.
func New(rpcURL string, opts ...Option) (*Client, error) {
	if strings.TrimSpace(rpcURL) == "" {
		return nil, errors.New("rnx: empty rpcURL")
	}
	id := 0
	c := &Client{
		rpcURL: rpcURL,
		http:   &http.Client{Timeout: 30 * time.Second},
		nextID: func() int { id++; return id },
	}
	for _, o := range opts {
		o(c)
	}
	return c, nil
}

// --- JSON-RPC envelope types ---

type rpcRequest struct {
	JSONRPC string        `json:"jsonrpc"`
	ID      int           `json:"id"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *rpcError) Error() string {
	return fmt.Sprintf("rnx: rpc error %d: %s", e.Code, e.Message)
}

type rpcResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      int             `json:"id"`
	Result  json.RawMessage `json:"result"`
	Error   *rpcError       `json:"error"`
}

// call performs a single JSON-RPC request and unmarshals Result into out.
func (c *Client) call(ctx context.Context, method string, params []interface{}, out interface{}) error {
	if params == nil {
		params = []interface{}{}
	}
	reqBody := rpcRequest{JSONRPC: "2.0", ID: c.nextID(), Method: method, Params: params}
	buf, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("rnx: marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.rpcURL, bytes.NewReader(buf))
	if err != nil {
		return fmt.Errorf("rnx: build http request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(httpReq)
	if err != nil {
		return fmt.Errorf("rnx: http do: %w", err)
	}
	defer resp.Body.Close()

	dec := json.NewDecoder(resp.Body)
	var rpcResp rpcResponse
	if err := dec.Decode(&rpcResp); err != nil {
		return fmt.Errorf("rnx: decode response (http %d): %w", resp.StatusCode, err)
	}
	if rpcResp.Error != nil {
		return rpcResp.Error
	}
	if out == nil {
		return nil
	}
	if err := json.Unmarshal(rpcResp.Result, out); err != nil {
		return fmt.Errorf("rnx: unmarshal result for %s: %w", method, err)
	}
	return nil
}

// --- Read methods ---

// ChainID returns the chain id reported by the RPC (eth_chainId), decoded from
// its hex-quantity form to a *big.Int.
func (c *Client) ChainID(ctx context.Context) (*big.Int, error) {
	var hex string
	if err := c.call(ctx, "eth_chainId", nil, &hex); err != nil {
		return nil, err
	}
	return parseHexBig(hex)
}

// BlockNumber returns the latest block height (eth_blockNumber).
func (c *Client) BlockNumber(ctx context.Context) (*big.Int, error) {
	var hex string
	if err := c.call(ctx, "eth_blockNumber", nil, &hex); err != nil {
		return nil, err
	}
	return parseHexBig(hex)
}

// GetBalance returns the wei balance of addr at the latest block
// (eth_getBalance).
func (c *Client) GetBalance(ctx context.Context, addr string) (*big.Int, error) {
	if !isHexAddress(addr) {
		return nil, fmt.Errorf("rnx: invalid address %q", addr)
	}
	var hex string
	if err := c.call(ctx, "eth_getBalance", []interface{}{addr, "latest"}, &hex); err != nil {
		return nil, err
	}
	return parseHexBig(hex)
}

// Block is a lightly-typed view over an eth_getBlockByNumber result. Unknown
// fields are preserved in Raw for callers that need more detail.
type Block struct {
	Number       *big.Int        `json:"-"`
	Hash         string          `json:"hash"`
	ParentHash   string          `json:"parentHash"`
	Timestamp    *big.Int        `json:"-"`
	Transactions []string        `json:"-"`
	Raw          json.RawMessage `json:"-"`
}

// GetBlock returns block number n (eth_getBlockByNumber, without full tx
// bodies — transaction hashes only). A nil n means the latest block.
func (c *Client) GetBlock(ctx context.Context, n *big.Int) (*Block, error) {
	tag := "latest"
	if n != nil {
		tag = "0x" + n.Text(16)
	}
	var raw json.RawMessage
	if err := c.call(ctx, "eth_getBlockByNumber", []interface{}{tag, false}, &raw); err != nil {
		return nil, err
	}
	if len(raw) == 0 || string(raw) == "null" {
		return nil, fmt.Errorf("rnx: block %s not found", tag)
	}
	var shape struct {
		Number       string   `json:"number"`
		Hash         string   `json:"hash"`
		ParentHash   string   `json:"parentHash"`
		Timestamp    string   `json:"timestamp"`
		Transactions []string `json:"transactions"`
	}
	if err := json.Unmarshal(raw, &shape); err != nil {
		return nil, fmt.Errorf("rnx: decode block: %w", err)
	}
	b := &Block{
		Hash:         shape.Hash,
		ParentHash:   shape.ParentHash,
		Transactions: shape.Transactions,
		Raw:          raw,
	}
	if shape.Number != "" {
		if num, err := parseHexBig(shape.Number); err == nil {
			b.Number = num
		}
	}
	if shape.Timestamp != "" {
		if ts, err := parseHexBig(shape.Timestamp); err == nil {
			b.Timestamp = ts
		}
	}
	return b, nil
}

// GetTransaction returns the transaction with the given hash
// (eth_getTransactionByHash). The raw JSON object is returned so callers can
// decode exactly the fields they need without pulling in a heavy dependency.
func (c *Client) GetTransaction(ctx context.Context, hash string) (json.RawMessage, error) {
	if !strings.HasPrefix(hash, "0x") || len(hash) != 66 {
		return nil, fmt.Errorf("rnx: invalid tx hash %q", hash)
	}
	var raw json.RawMessage
	if err := c.call(ctx, "eth_getTransactionByHash", []interface{}{hash}, &raw); err != nil {
		return nil, err
	}
	if len(raw) == 0 || string(raw) == "null" {
		return nil, fmt.Errorf("rnx: transaction %s not found", hash)
	}
	return raw, nil
}

// Call performs a read-only eth_call against `to` with the ABI-encoded `data`
// (0x-prefixed) at the latest block. It returns the raw 0x-prefixed return
// bytes. This method NEVER sends a transaction and NEVER signs anything.
func (c *Client) Call(ctx context.Context, to, data string) (string, error) {
	if !isHexAddress(to) {
		return "", fmt.Errorf("rnx: invalid 'to' address %q", to)
	}
	if !strings.HasPrefix(data, "0x") {
		return "", fmt.Errorf("rnx: calldata must be 0x-prefixed")
	}
	callObj := map[string]string{"to": to, "data": data}
	var out string
	if err := c.call(ctx, "eth_call", []interface{}{callObj, "latest"}, &out); err != nil {
		return "", err
	}
	return out, nil
}

// --- Chain safety ---

// CheckChain verifies the connected RPC is RNX (RNX_CHAIN_ID). It is
// fail-closed: ANY mismatch returns an error. It distinguishes the legacy
// chain (12345) so operators get a precise diagnostic.
func (c *Client) CheckChain(ctx context.Context) error {
	id, err := c.ChainID(ctx)
	if err != nil {
		return fmt.Errorf("rnx: CheckChain could not read chain id: %w", err)
	}
	if id.Int64() == RNX_CHAIN_ID {
		return nil
	}
	if id.Int64() == LegacyChainID {
		return fmt.Errorf("%w: connected to LEGACY chain %d, expected RNX %d — update your RPC endpoint",
			ErrChainMismatch, LegacyChainID, RNX_CHAIN_ID)
	}
	return fmt.Errorf("%w: got %s, expected RNX %d", ErrChainMismatch, id.String(), RNX_CHAIN_ID)
}

// --- VERIDEX ---

// VeridexVerifyResult holds the on-chain state returned by the registry's
// currentAnchor view. Status is derived ONLY from on-chain state (fail-closed):
// "OK" when an anchor exists, "NO_ANCHOR" otherwise. Nothing is synthesized.
type VeridexVerifyResult struct {
	VerificationStatus string
	NameHash           string
	CommitmentHash     string
	Owner              string
	StateVersion       *big.Int
	BlockNumber        *big.Int
	Exists             bool
}

// VeridexVerify reads currentAnchor(nameHash) on the VERIDEX registry via
// eth_call and reports verification status purely from the on-chain result.
//
// nameHash must be the 0x-prefixed 32-byte keccak256 nameHash of the subject
// (computed by the caller / another SDK). This function performs no hashing and
// needs no key.
func (c *Client) VeridexVerify(ctx context.Context, registryAddr, nameHash string) (*VeridexVerifyResult, error) {
	if !isHexAddress(registryAddr) {
		return nil, fmt.Errorf("rnx: invalid registry address %q", registryAddr)
	}
	word, err := hexWord32(nameHash)
	if err != nil {
		return nil, fmt.Errorf("rnx: invalid nameHash: %w", err)
	}
	data := currentAnchorSelector + word
	ret, err := c.Call(ctx, registryAddr, data)
	if err != nil {
		return nil, err
	}
	return decodeCurrentAnchor(nameHash, ret)
}

// decodeCurrentAnchor decodes the ABI-encoded return of
// currentAnchor(bytes32) -> (bytes32,address,uint256,uint256,bool).
// Five 32-byte words. Fail-closed: short/empty returns => NO_ANCHOR.
func decodeCurrentAnchor(nameHash, ret string) (*VeridexVerifyResult, error) {
	res := &VeridexVerifyResult{
		VerificationStatus: "NO_ANCHOR",
		NameHash:           nameHash,
		Exists:             false,
	}
	h := strings.TrimPrefix(ret, "0x")
	if len(h) < 5*64 {
		// Empty / malformed return: fail-closed, treat as no anchor.
		return res, nil
	}
	words := make([]string, 5)
	for i := 0; i < 5; i++ {
		words[i] = h[i*64 : (i+1)*64]
	}
	res.CommitmentHash = "0x" + words[0]
	res.Owner = "0x" + words[1][24:] // last 20 bytes of the address word
	if sv, ok := new(big.Int).SetString(words[2], 16); ok {
		res.StateVersion = sv
	}
	if bn, ok := new(big.Int).SetString(words[3], 16); ok {
		res.BlockNumber = bn
	}
	exists, ok := new(big.Int).SetString(words[4], 16)
	if ok && exists.Sign() != 0 {
		res.Exists = true
		res.VerificationStatus = "OK"
	}
	return res, nil
}

// --- Write (pre-signed only) ---

// SendRawTransaction broadcasts a caller-provided PRE-SIGNED raw transaction
// (eth_sendRawTransaction) and returns the resulting transaction hash.
//
// This is the ONLY write method in the SDK. It never signs, never derives, and
// never touches private keys: the caller is solely responsible for producing a
// correctly signed rawHex (e.g. via an offline signer / hardware wallet).
func (c *Client) SendRawTransaction(ctx context.Context, rawHex string) (string, error) {
	if !strings.HasPrefix(rawHex, "0x") || len(rawHex) <= 2 {
		return "", fmt.Errorf("rnx: raw transaction must be non-empty 0x-prefixed hex")
	}
	var txHash string
	if err := c.call(ctx, "eth_sendRawTransaction", []interface{}{rawHex}, &txHash); err != nil {
		return "", err
	}
	return txHash, nil
}

// --- helpers ---

// parseHexBig parses a 0x-prefixed hex quantity into a *big.Int.
func parseHexBig(h string) (*big.Int, error) {
	s := strings.TrimPrefix(h, "0x")
	if s == "" {
		return nil, fmt.Errorf("rnx: empty hex quantity %q", h)
	}
	n, ok := new(big.Int).SetString(s, 16)
	if !ok {
		return nil, fmt.Errorf("rnx: invalid hex quantity %q", h)
	}
	return n, nil
}

// isHexAddress reports whether s looks like a 20-byte 0x-prefixed address.
func isHexAddress(s string) bool {
	if !strings.HasPrefix(s, "0x") || len(s) != 42 {
		return false
	}
	for _, r := range s[2:] {
		if !isHexDigit(r) {
			return false
		}
	}
	return true
}

// hexWord32 validates a 0x-prefixed 32-byte value and returns the bare
// (unprefixed, lower-case, 64-hex-char) word.
func hexWord32(s string) (string, error) {
	h := strings.TrimPrefix(s, "0x")
	if len(h) != 64 {
		return "", fmt.Errorf("expected 32-byte (64 hex char) value, got %d chars", len(h))
	}
	for _, r := range h {
		if !isHexDigit(r) {
			return "", fmt.Errorf("non-hex character in %q", s)
		}
	}
	return strings.ToLower(h), nil
}

func isHexDigit(r rune) bool {
	return (r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F')
}
