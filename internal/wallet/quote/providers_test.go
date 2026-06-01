package quote

import (
	"context"
	"encoding/json"
	"io"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestKyberSwap_QuoteParsesAndChargesFee(t *testing.T) {
	var routesQuery, clientID string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clientID = r.Header.Get("x-client-id")
		switch r.URL.Path {
		case "/ethereum/api/v1/routes":
			routesQuery = r.URL.RawQuery
			_ = json.NewEncoder(w).Encode(map[string]any{
				"code": 0, "message": "ok",
				"data": map[string]any{
					"routeSummary":  map[string]any{"amountOut": "200000000", "gas": "250000"},
					"routerAddress": "0xRouter",
				},
			})
		case "/ethereum/api/v1/route/build":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"code": 0,
				"data": map[string]any{
					"amountOut": "200000000", "data": "0xdeadbeef",
					"routerAddress": "0xRouter", "gas": "260000", "transactionValue": "0",
				},
			})
		default:
			http.Error(w, "404", 404)
		}
	}))
	defer srv.Close()

	p := NewKyberSwapProvider(srv.URL, "im-wallet", map[string]string{"eth": "0xfee"}, 30, false)
	q, err := p.Quote(context.Background(), Request{
		ChainKey: "eth", SellToken: Token{ContractAddress: "0xUSDC"},
		BuyToken: Token{ContractAddress: "0xUSDT"}, SellAmount: big.NewInt(1e9),
		Taker: "0xt", SlippageBps: 50,
	})
	if err != nil {
		t.Fatalf("Quote error: %v", err)
	}
	if clientID != "im-wallet" {
		t.Fatalf("x-client-id = %q", clientID)
	}
	// fee params on /routes
	for _, want := range []string{"feeReceiver=0xfee", "chargeFeeBy=currency_out", "feeAmount=30", "isInBps=true"} {
		if !contains(routesQuery, want) {
			t.Fatalf("routes query %q missing %q", routesQuery, want)
		}
	}
	if q.Provider != "kyberswap" || q.To != "0xRouter" || q.Data != "0xdeadbeef" {
		t.Fatalf("bad quote: %+v", q)
	}
	if q.MinBuyAmount != "199000000" { // 200000000 * 9950/10000
		t.Fatalf("minBuy = %q", q.MinBuyAmount)
	}
	if q.Approval == nil || q.Approval.Spender != "0xRouter" {
		t.Fatalf("approval spender wrong: %+v", q.Approval)
	}
}

func TestParaswap_QuoteSendsDecimalsAndDirectFee(t *testing.T) {
	var pricesQuery string
	var buildBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/prices/":
			pricesQuery = r.URL.RawQuery
			_ = json.NewEncoder(w).Encode(map[string]any{
				"priceRoute": map[string]any{
					"destAmount": "201000000", "gasCost": "200000",
					"tokenTransferProxy": "0xProxy",
				},
			})
		case r.URL.Path == "/transactions/1":
			b, _ := io.ReadAll(r.Body)
			_ = json.Unmarshal(b, &buildBody)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"to": "0xAugustus", "data": "0xcafe", "value": "0", "gas": "210000",
			})
		default:
			http.Error(w, "404", 404)
		}
	}))
	defer srv.Close()

	p := NewParaswapProvider(srv.URL, "im-wallet", map[string]string{"eth": "0xfee"}, 30, false)
	q, err := p.Quote(context.Background(), Request{
		ChainKey: "eth", SellToken: Token{ContractAddress: "0xUSDC", Decimals: 6},
		BuyToken: Token{ContractAddress: "0xUSDT", Decimals: 6}, SellAmount: big.NewInt(1e9),
		Taker: "0xt", SlippageBps: 100,
	})
	if err != nil {
		t.Fatalf("Quote error: %v", err)
	}
	for _, want := range []string{"srcDecimals=6", "destDecimals=6", "side=SELL", "network=1"} {
		if !contains(pricesQuery, want) {
			t.Fatalf("prices query %q missing %q", pricesQuery, want)
		}
	}
	if buildBody["isDirectFeeTransfer"] != true {
		t.Fatalf("isDirectFeeTransfer not set: %+v", buildBody)
	}
	if buildBody["partnerAddress"] != "0xfee" {
		t.Fatalf("partnerAddress = %v", buildBody["partnerAddress"])
	}
	if q.To != "0xAugustus" || q.Approval == nil || q.Approval.Spender != "0xProxy" {
		t.Fatalf("bad quote/approval: to=%s approval=%+v", q.To, q.Approval)
	}
}

func contains(haystack, needle string) bool {
	return len(haystack) >= len(needle) && (indexOf(haystack, needle) >= 0)
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
