package bridge

import (
	"context"
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestLiFi_QuoteParsesHexAndApproval(t *testing.T) {
	var gotQuery url.Values
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.Query()
		_ = json.NewEncoder(w).Encode(map[string]any{
			"tool": "across",
			"estimate": map[string]any{
				"approvalAddress":   "0x1231DEB6f5749EF6cE6943a275A1D3E7486F4EaE",
				"toAmount":          "4981314",
				"toAmountMin":       "4900000",
				"executionDuration": 2,
			},
			"transactionRequest": map[string]any{
				"to":       "0x1231DEB6f5749EF6cE6943a275A1D3E7486F4EaE",
				"data":     "0xabcd",
				"value":    "0x0",
				"gasLimit": "0x30d40", // 200000
				"gasPrice": "0x3b9aca00",
			},
		})
	}))
	defer srv.Close()

	p := NewLiFiProvider(srv.URL, "", "im-wallet", 0)
	q, err := p.Quote(context.Background(), Request{
		FromChain: "eth", ToChain: "arbitrum",
		FromToken: "0xUSDC", ToToken: "0xUSDCarb",
		FromAmount: big.NewInt(5000000), FromAddress: "0xme", SlippageBps: 100,
	})
	if err != nil {
		t.Fatalf("Quote error: %v", err)
	}
	// 100 bps -> slippage 0.01; chain ids mapped
	if gotQuery.Get("slippage") != "0.01" {
		t.Fatalf("slippage = %q", gotQuery.Get("slippage"))
	}
	if gotQuery.Get("fromChain") != "1" || gotQuery.Get("toChain") != "42161" {
		t.Fatalf("chain ids: %s -> %s", gotQuery.Get("fromChain"), gotQuery.Get("toChain"))
	}
	if q.Tool != "across" || q.ToAmount != "4981314" {
		t.Fatalf("bad quote: %+v", q)
	}
	if q.Value != "0" || q.Gas != "200000" { // hex normalized to decimal
		t.Fatalf("hex normalize: value=%q gas=%q", q.Value, q.Gas)
	}
	if q.Approval == nil || q.Approval.Spender != "0x1231DEB6f5749EF6cE6943a275A1D3E7486F4EaE" {
		t.Fatalf("approval: %+v", q.Approval)
	}
}

func TestLiFi_NativeSourceNoApproval(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("fromToken") != NativeToken {
			t.Errorf("native not mapped to zero address: %q", r.URL.Query().Get("fromToken"))
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"tool": "across",
			"estimate": map[string]any{"approvalAddress": "0xdiamond", "toAmount": "99", "toAmountMin": "98"},
			"transactionRequest": map[string]any{"to": "0xdiamond", "data": "0x", "value": "0x470de4df820000"},
		})
	}))
	defer srv.Close()

	p := NewLiFiProvider(srv.URL, "", "im-wallet", 0)
	q, err := p.Quote(context.Background(), Request{
		FromChain: "eth", ToChain: "polygon", FromToken: "native", ToToken: "native",
		FromAmount: big.NewInt(20000000000000000), FromAddress: "0xme", SlippageBps: 50,
	})
	if err != nil {
		t.Fatalf("Quote error: %v", err)
	}
	if q.Approval != nil {
		t.Fatalf("native source should have no approval: %+v", q.Approval)
	}
	if q.Value != "20000000000000000" { // 0x470de4df820000
		t.Fatalf("native value = %q", q.Value)
	}
}

func TestLiFi_StatusMapping(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status":           "DONE",
			"substatus":        "COMPLETED",
			"lifiExplorerLink": "https://scan.li.fi/tx/0xabc",
			"receiving":        map[string]any{"txHash": "0xdesttx", "amount": "4981000"},
		})
	}))
	defer srv.Close()

	p := NewLiFiProvider(srv.URL, "", "im-wallet", 0)
	st, err := p.Status(context.Background(), "eth", "arbitrum", "0xsrctx", "across")
	if err != nil {
		t.Fatalf("Status error: %v", err)
	}
	if st.Status != "DONE" || st.DestTxHash != "0xdesttx" || st.DestAmount != "4981000" {
		t.Fatalf("bad status: %+v", st)
	}
}

func TestLiFi_SupportsChain(t *testing.T) {
	p := NewLiFiProvider("", "", "", 0)
	if !p.SupportsChain("eth") || !p.SupportsChain("arbitrum") {
		t.Fatal("eth/arbitrum should be supported")
	}
	if p.SupportsChain("tron") {
		t.Fatal("tron should not be supported")
	}
}
