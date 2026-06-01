// Package intent implements intent-based ("sign an order, a solver fills it")
// swaps via CoW Protocol. The client signs an EIP-712 order (gasless at
// execution, MEV-protected); this package proxies CoW's orderbook for quote,
// submission, and status. Mirrors internal/wallet/bridge.
package intent

import (
	"context"
	"errors"
)

var ErrUnsupportedChain = errors.New("unsupported chain")
var ErrNoQuote = errors.New("no quote available")

// Provider is an intent (order-based) swap integration.
type Provider interface {
	Name() string
	SupportsChain(chainKey string) bool
	// Quote returns a ready-to-sign EIP-712 order + domain/approval metadata.
	Quote(ctx context.Context, req Request) (*QuoteResult, error)
	// Submit posts a signed order to the orderbook, returning its UID.
	Submit(ctx context.Context, chainKey string, sub SubmitRequest) (string, error)
	// Status reports an order's settlement state.
	Status(ctx context.Context, chainKey, orderUID string) (*OrderStatus, error)
}
