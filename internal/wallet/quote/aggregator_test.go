package quote

import (
	"context"
	"errors"
	"math/big"
	"testing"

	"go.uber.org/zap"
)

// mockProvider is a configurable Provider for aggregator tests.
type mockProvider struct {
	name      string
	chains    map[string]bool
	buyAmount string
	err       error
}

func (m *mockProvider) Name() string                  { return m.name }
func (m *mockProvider) SupportsChain(c string) bool    { return m.chains[c] }
func (m *mockProvider) Price(_ context.Context, _ Request) (*PriceResult, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &PriceResult{Provider: m.name, BuyAmount: m.buyAmount}, nil
}
func (m *mockProvider) Quote(_ context.Context, req Request) (*Quote, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &Quote{Provider: m.name, BuyAmount: m.buyAmount, To: "0xrouter", Data: "0x", Value: "0"}, nil
}

func eth(provs ...Provider) *Aggregator {
	return NewAggregator(map[string][]Provider{"eth": provs}, zap.NewNop())
}

func req() Request {
	return Request{ChainKey: "eth", SellAmount: big.NewInt(1), SlippageBps: 50}
}

func TestBestQuote_PicksHighestBuyAmount(t *testing.T) {
	a := eth(
		&mockProvider{name: "a", chains: map[string]bool{"eth": true}, buyAmount: "100"},
		&mockProvider{name: "b", chains: map[string]bool{"eth": true}, buyAmount: "300"},
		&mockProvider{name: "c", chains: map[string]bool{"eth": true}, buyAmount: "200"},
	)
	out, err := a.BestQuote(context.Background(), req())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Best.Provider != "b" {
		t.Fatalf("expected winner b, got %s", out.Best.Provider)
	}
	if len(out.All) != 3 || out.All[0].Provider != "b" || out.All[2].Provider != "a" {
		t.Fatalf("ranked list wrong: %+v", out.All)
	}
}

func TestBestQuote_SkipsErrorsAndUnsupported(t *testing.T) {
	a := eth(
		&mockProvider{name: "boom", chains: map[string]bool{"eth": true}, err: errors.New("api 500")},
		&mockProvider{name: "unsup", chains: map[string]bool{"polygon": true}, buyAmount: "999"}, // not eth
		&mockProvider{name: "ok", chains: map[string]bool{"eth": true}, buyAmount: "150"},
	)
	out, err := a.BestQuote(context.Background(), req())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Best.Provider != "ok" || len(out.All) != 1 {
		t.Fatalf("expected only 'ok' to survive, got %+v", out.All)
	}
}

func TestBestQuote_AllFail(t *testing.T) {
	a := eth(&mockProvider{name: "boom", chains: map[string]bool{"eth": true}, err: errors.New("x")})
	_, err := a.BestQuote(context.Background(), req())
	if !errors.Is(err, ErrNoQuote) {
		t.Fatalf("expected ErrNoQuote, got %v", err)
	}
}

func TestBestQuote_UnknownChain(t *testing.T) {
	a := eth(&mockProvider{name: "a", chains: map[string]bool{"eth": true}, buyAmount: "1"})
	_, err := a.BestQuote(context.Background(), Request{ChainKey: "solana", SellAmount: big.NewInt(1)})
	if !errors.Is(err, ErrNoQuote) {
		t.Fatalf("expected ErrNoQuote for unconfigured chain, got %v", err)
	}
}
