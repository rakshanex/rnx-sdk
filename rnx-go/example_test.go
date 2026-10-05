package rnx

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// Example demonstrates read-only usage against a placeholder RPC endpoint.
// It makes NO real network calls at test time (it is an example of API shape,
// not an executed Example with Output).
func Example() {
	// Placeholder endpoint — replace with your RNX RPC. No key is ever used.
	c, err := New("https://rpc.rnx.example/placeholder")
	if err != nil {
		return
	}
	_ = c // read-only methods: c.ChainID, c.BlockNumber, c.GetBalance, ...
	fmt.Println("rnx client constructed")
	// Output: rnx client constructed
}

// --- Table-driven: chainId hex <-> dec parsing ---

func TestParseHexBigChainID(t *testing.T) {
	tests := []struct {
		name    string
		hex     string
		wantDec int64
		wantErr bool
	}{
		{"rnx future chain", RNXChainIDHex, RNX_CHAIN_ID, false},
		{"rnx future chain upper", "0x2F667", 194151, false},
		{"legacy chain", "0x3039", LegacyChainID, false},
		{"one", "0x1", 1, false},
		{"zero padded", "0x0000000000000001", 1, false},
		{"empty", "0x", 0, true},
		{"garbage", "0xzzzz", 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseHexBig(tt.hex)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error for %q, got %v", tt.hex, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error for %q: %v", tt.hex, err)
			}
			if got.Int64() != tt.wantDec {
				t.Fatalf("hex %q => %d, want %d", tt.hex, got.Int64(), tt.wantDec)
			}
		})
	}
}

// TestRNXChainIDConstantsConsistent ensures the hex constant matches the
// decimal constant (prevents drift between the two source-of-truth values).
func TestRNXChainIDConstantsConsistent(t *testing.T) {
	n, err := parseHexBig(RNXChainIDHex)
	if err != nil {
		t.Fatalf("RNXChainIDHex not parseable: %v", err)
	}
	if n.Int64() != RNX_CHAIN_ID {
		t.Fatalf("RNXChainIDHex %s => %d, but RNX_CHAIN_ID = %d", RNXChainIDHex, n.Int64(), RNX_CHAIN_ID)
	}
	if RNX_CHAIN_ID == LegacyChainID {
		t.Fatal("RNX_CHAIN_ID must differ from LegacyChainID")
	}
}

// --- Mock JSON-RPC server ---

// newMockRPC returns a test server that dispatches on JSON-RPC method name
// using the provided results map (method -> result value, JSON-marshaled).
func newMockRPC(t *testing.T, results map[string]interface{}) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var req rpcRequest
		if err := json.Unmarshal(body, &req); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		res, ok := results[req.Method]
		if !ok {
			_ = json.NewEncoder(w).Encode(rpcResponse{
				JSONRPC: "2.0", ID: req.ID,
				Error: &rpcError{Code: -32601, Message: "method not found: " + req.Method},
			})
			return
		}
		raw, _ := json.Marshal(res)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(rpcResponse{JSONRPC: "2.0", ID: req.ID, Result: raw})
	}))
}

func TestChainIDAgainstMock(t *testing.T) {
	srv := newMockRPC(t, map[string]interface{}{
		"eth_chainId": RNXChainIDHex,
	})
	defer srv.Close()

	c, err := New(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	id, err := c.ChainID(context.Background())
	if err != nil {
		t.Fatalf("ChainID: %v", err)
	}
	if id.Int64() != RNX_CHAIN_ID {
		t.Fatalf("ChainID = %d, want %d", id.Int64(), RNX_CHAIN_ID)
	}
}

func TestBlockNumberAgainstMock(t *testing.T) {
	srv := newMockRPC(t, map[string]interface{}{
		"eth_blockNumber": "0x10d4f", // 68943
	})
	defer srv.Close()

	c, err := New(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	bn, err := c.BlockNumber(context.Background())
	if err != nil {
		t.Fatalf("BlockNumber: %v", err)
	}
	if bn.Int64() != 68943 {
		t.Fatalf("BlockNumber = %d, want 68943", bn.Int64())
	}
}

// TestCheckChainFailClosed verifies CheckChain passes on RNX, and returns a
// precise, distinguishing error for the legacy chain and other mismatches.
func TestCheckChainFailClosed(t *testing.T) {
	cases := []struct {
		name      string
		chainHex  string
		wantOK    bool
		wantLegacy bool
	}{
		{"rnx ok", RNXChainIDHex, true, false},
		{"legacy", "0x3039", false, true},
		{"random mainnet", "0x1", false, false},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			srv := newMockRPC(t, map[string]interface{}{"eth_chainId": tt.chainHex})
			defer srv.Close()
			c, _ := New(srv.URL)
			err := c.CheckChain(context.Background())
			if tt.wantOK {
				if err != nil {
					t.Fatalf("expected OK, got %v", err)
				}
				return
			}
			if err == nil {
				t.Fatal("expected mismatch error, got nil")
			}
			if tt.wantLegacy && !strings.Contains(err.Error(), "LEGACY") {
				t.Fatalf("expected legacy-specific error, got %v", err)
			}
		})
	}
}

// TestVeridexVerifyDecode exercises the fail-closed currentAnchor decoding via
// a mock eth_call. An "exists" word of 1 => OK; empty return => NO_ANCHOR.
func TestVeridexVerifyDecode(t *testing.T) {
	nameHash := "0x" + strings.Repeat("ab", 32)
	owner := "1111111111111111111111111111111111111111"
	// 5 words: commitment, owner(padded), stateVersion=7, blockNumber=42, exists=1
	ret := "0x" +
		strings.Repeat("cd", 32) + // commitment
		strings.Repeat("0", 24) + owner + // address right-aligned
		fmt.Sprintf("%064x", 7) +
		fmt.Sprintf("%064x", 42) +
		fmt.Sprintf("%064x", 1)

	srv := newMockRPC(t, map[string]interface{}{
		"eth_call": ret,
	})
	defer srv.Close()
	c, _ := New(srv.URL)

	reg := "0x" + strings.Repeat("22", 20)
	res, err := c.VeridexVerify(context.Background(), reg, nameHash)
	if err != nil {
		t.Fatalf("VeridexVerify: %v", err)
	}
	if res.VerificationStatus != "OK" {
		t.Fatalf("status = %q, want OK", res.VerificationStatus)
	}
	if !res.Exists {
		t.Fatal("expected Exists=true")
	}
	if res.StateVersion == nil || res.StateVersion.Int64() != 7 {
		t.Fatalf("stateVersion = %v, want 7", res.StateVersion)
	}
	if res.BlockNumber == nil || res.BlockNumber.Int64() != 42 {
		t.Fatalf("blockNumber = %v, want 42", res.BlockNumber)
	}
	if !strings.HasSuffix(res.Owner, owner) {
		t.Fatalf("owner = %q, want suffix %q", res.Owner, owner)
	}

	// Fail-closed: empty return => NO_ANCHOR.
	noAnchor, err := decodeCurrentAnchor(nameHash, "0x")
	if err != nil {
		t.Fatalf("decode empty: %v", err)
	}
	if noAnchor.VerificationStatus != "NO_ANCHOR" || noAnchor.Exists {
		t.Fatalf("empty return should be NO_ANCHOR, got %+v", noAnchor)
	}
}

// TestSendRawTransactionNoKeys confirms the only write method forwards a
// pre-signed blob verbatim and validates input; it never signs.
func TestSendRawTransactionNoKeys(t *testing.T) {
	srv := newMockRPC(t, map[string]interface{}{
		"eth_sendRawTransaction": "0x" + strings.Repeat("ff", 32),
	})
	defer srv.Close()
	c, _ := New(srv.URL)

	if _, err := c.SendRawTransaction(context.Background(), "nothex"); err == nil {
		t.Fatal("expected error for non-0x raw tx")
	}
	hash, err := c.SendRawTransaction(context.Background(), "0xdeadbeef")
	if err != nil {
		t.Fatalf("SendRawTransaction: %v", err)
	}
	if len(hash) != 66 {
		t.Fatalf("unexpected tx hash %q", hash)
	}
}

func TestAddressValidation(t *testing.T) {
	c, _ := New("https://placeholder.invalid")
	if _, err := c.GetBalance(context.Background(), "0x123"); err == nil {
		t.Fatal("expected invalid address error")
	}
}

func TestGetBlockAgainstMock(t *testing.T) {
	block := map[string]interface{}{
		"number":       "0x2a",
		"hash":         "0x" + strings.Repeat("11", 32),
		"parentHash":   "0x" + strings.Repeat("00", 32),
		"timestamp":    "0x5f5e100",
		"transactions": []string{"0x" + strings.Repeat("22", 32)},
	}
	srv := newMockRPC(t, map[string]interface{}{"eth_getBlockByNumber": block})
	defer srv.Close()
	c, _ := New(srv.URL)

	b, err := c.GetBlock(context.Background(), big.NewInt(42))
	if err != nil {
		t.Fatalf("GetBlock: %v", err)
	}
	if b.Number == nil || b.Number.Int64() != 42 {
		t.Fatalf("block number = %v, want 42", b.Number)
	}
	if len(b.Transactions) != 1 {
		t.Fatalf("expected 1 tx, got %d", len(b.Transactions))
	}
}
