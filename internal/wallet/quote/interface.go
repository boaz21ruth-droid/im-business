// Package quote implements server-side swap quote aggregation. Multiple DEX
// aggregator providers (0x, 1inch, OKX, Paraswap, …) are queried concurrently
// and the best quote is returned to the client, which signs and broadcasts
// locally. Aggregator API keys live here and never reach the client.
//
// This mirrors the tx-history provider design in internal/wallet/provider,
// except the aggregator fans out concurrently and picks the best output rather
// than taking the first successful provider.
package quote

import (
	"context"
	"errors"
)

// ErrUnsupportedChain is returned by a provider when it cannot serve the
// requested chain. The aggregator treats this as an expected skip (Debug log),
// not a failure (Warn log).
var ErrUnsupportedChain = errors.New("unsupported chain")

// Provider is a single DEX aggregator integration.
type Provider interface {
	Name() string
	// SupportsChain reports whether this provider can quote on chainKey.
	SupportsChain(chainKey string) bool
	// Price is a soft quote for the live "you'll get X" preview.
	Price(ctx context.Context, req Request) (*PriceResult, error)
	// Quote is a firm quote: returns executable calldata ready to sign.
	Quote(ctx context.Context, req Request) (*Quote, error)
}
