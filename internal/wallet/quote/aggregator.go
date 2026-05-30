package quote

import (
	"context"
	"errors"
	"math/big"
	"sort"
	"sync"
	"time"

	"go.uber.org/zap"
)

// ErrNoQuote means no provider returned a usable quote for the request.
var ErrNoQuote = errors.New("no quote available")

// defaultProviderTimeout caps how long any single provider may take before the
// aggregator gives up on it and uses whatever the others returned.
const defaultProviderTimeout = 4 * time.Second

// Aggregator fans out a request to all providers configured for the chain,
// concurrently, and returns the best result (highest buyAmount) plus the full
// ranked list. Unlike provider.MultiProvider (first-success fallback), it waits
// for all and compares — a slow/failed provider never blocks the winner.
type Aggregator struct {
	providersByChain map[string][]Provider
	timeout          time.Duration
	log              *zap.Logger
}

// NewAggregator builds an aggregator from a per-chain ordered provider map.
func NewAggregator(providersByChain map[string][]Provider, log *zap.Logger) *Aggregator {
	return &Aggregator{
		providersByChain: providersByChain,
		timeout:          defaultProviderTimeout,
		log:              log,
	}
}

// SupportsChain reports whether any provider is configured for chainKey.
func (a *Aggregator) SupportsChain(chainKey string) bool {
	return len(a.providersByChain[chainKey]) > 0
}

// BestPrice returns the best soft quote across providers for req.ChainKey.
func (a *Aggregator) BestPrice(ctx context.Context, req Request) (*PriceResponse, error) {
	results := fanOut(ctx, a, req, func(ctx context.Context, p Provider) (*PriceResult, string, error) {
		r, err := p.Price(ctx, req)
		if r != nil {
			return r, r.BuyAmount, err
		}
		return nil, "", err
	})
	if len(results) == 0 {
		return nil, ErrNoQuote
	}
	return &PriceResponse{Best: results[0], All: results}, nil
}

// BestQuote returns the best firm quote across providers for req.ChainKey.
func (a *Aggregator) BestQuote(ctx context.Context, req Request) (*QuoteResponse, error) {
	results := fanOut(ctx, a, req, func(ctx context.Context, p Provider) (*Quote, string, error) {
		q, err := p.Quote(ctx, req)
		if q != nil {
			return q, q.BuyAmount, err
		}
		return nil, "", err
	})
	if len(results) == 0 {
		return nil, ErrNoQuote
	}
	return &QuoteResponse{Best: results[0], All: results}, nil
}

// fanOut runs call against every provider that supports req.ChainKey,
// concurrently with a per-provider timeout, collects the successes, and returns
// them sorted by buyAmount descending. call returns (result, buyAmountStr, err).
func fanOut[T any](ctx context.Context, a *Aggregator, req Request, call func(context.Context, Provider) (T, string, error)) []T {
	provs := a.providersByChain[req.ChainKey]

	type scored struct {
		val       T
		buyAmount *big.Int
	}

	var (
		mu      sync.Mutex
		scoredR []scored
		wg      sync.WaitGroup
	)

	for _, p := range provs {
		if !p.SupportsChain(req.ChainKey) {
			continue
		}
		wg.Add(1)
		go func(p Provider) {
			defer wg.Done()
			cctx, cancel := context.WithTimeout(ctx, a.timeout)
			defer cancel()

			val, amtStr, err := call(cctx, p)
			if err != nil {
				if errors.Is(err, ErrUnsupportedChain) {
					a.log.Debug("quote provider skipped (unsupported chain)",
						zap.String("provider", p.Name()), zap.String("chain", req.ChainKey))
				} else {
					a.log.Warn("quote provider failed",
						zap.String("provider", p.Name()), zap.String("chain", req.ChainKey), zap.Error(err))
				}
				return
			}
			amt, ok := new(big.Int).SetString(amtStr, 10)
			if !ok || amt.Sign() <= 0 {
				a.log.Warn("quote provider returned unparsable buyAmount",
					zap.String("provider", p.Name()), zap.String("buyAmount", amtStr))
				return
			}
			mu.Lock()
			scoredR = append(scoredR, scored{val: val, buyAmount: amt})
			mu.Unlock()
		}(p)
	}
	wg.Wait()

	// Best net output is approximated by gross buyAmount. Gas-adjusted ranking
	// would need a gas→buyToken price oracle; deferred until that exists.
	sort.SliceStable(scoredR, func(i, j int) bool {
		return scoredR[i].buyAmount.Cmp(scoredR[j].buyAmount) > 0
	})

	out := make([]T, len(scoredR))
	for i, s := range scoredR {
		out[i] = s.val
	}
	return out
}
