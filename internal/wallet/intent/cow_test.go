package intent

import (
	"context"
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCow_QuoteFoldsfeeAndAppliesSlippage(t *testing.T) {
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/mainnet/api/v1/quote" {
			http.Error(w, "bad path", 404)
			return
		}
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"quote": map[string]any{
				"sellToken": "0xusdc", "buyToken": "0xusdt",
				"receiver":   "0xme",
				"sellAmount": "99832150", // after fee
				"buyAmount":  "100000000",
				"validTo":    1780308179,
				"appData":    zeroAppData,
				"feeAmount":  "167850",
				"kind":       "sell",
			},
			"id": 1194376838,
		})
	}))
	defer srv.Close()

	p := NewCowProvider(srv.URL)
	q, err := p.Quote(context.Background(), Request{
		ChainKey: "eth", SellToken: "0xusdc", BuyToken: "0xusdt",
		SellAmount: big.NewInt(100000000), From: "0xme", SlippageBps: 100,
	})
	if err != nil {
		t.Fatalf("Quote error: %v", err)
	}
	// signed order: full sell = 99832150 + 167850 = 100000000; fee zeroed.
	if q.Order.SellAmount != "100000000" {
		t.Fatalf("sellAmount = %s, want 100000000 (fee folded)", q.Order.SellAmount)
	}
	if q.Order.FeeAmount != "0" {
		t.Fatalf("feeAmount = %s, want 0", q.Order.FeeAmount)
	}
	// buyAmount = 100000000 * 9900/10000 = 99000000 (100 bps slippage floor).
	if q.Order.BuyAmount != "99000000" {
		t.Fatalf("buyAmount = %s, want 99000000", q.Order.BuyAmount)
	}
	if q.VerifyingContract != gpv2Settlement || q.ApprovalSpender != gpv2VaultRelayer {
		t.Fatalf("domain/approval consts wrong: %+v", q)
	}
	if q.ChainID != 1 || q.QuoteID != 1194376838 {
		t.Fatalf("chainId/quoteId wrong: %+v", q)
	}
	if gotBody["signingScheme"] != "eip712" || gotBody["kind"] != "sell" {
		t.Fatalf("request body wrong: %+v", gotBody)
	}
}

func TestCow_SubmitReturnsUID(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode("0xorderuid123")
	}))
	defer srv.Close()
	p := NewCowProvider(srv.URL)
	uid, err := p.Submit(context.Background(), "eth", SubmitRequest{
		Order: Order{FeeAmount: "0", Kind: "sell"}, Signature: "0xsig", From: "0xme", QuoteID: 1,
	})
	if err != nil || uid != "0xorderuid123" {
		t.Fatalf("submit: uid=%q err=%v", uid, err)
	}
}

func TestCow_StatusAndSupport(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status": "fulfilled", "executedBuyAmount": "99500000",
		})
	}))
	defer srv.Close()
	p := NewCowProvider(srv.URL)
	st, err := p.Status(context.Background(), "eth", "0xuid")
	if err != nil || st.Status != "fulfilled" || st.ExecutedBuyAmount != "99500000" {
		t.Fatalf("status: %+v err=%v", st, err)
	}
	if !p.SupportsChain("eth") || !p.SupportsChain("arbitrum") || p.SupportsChain("bsc") {
		t.Fatal("chain support wrong (eth/arbitrum yes, bsc no)")
	}
}
