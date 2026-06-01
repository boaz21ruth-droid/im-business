// Package bridge implements cross-chain swap quoting/tracking via a bridge
// aggregator (LI.FI). Unlike same-chain DEX quotes (internal/wallet/quote), a
// cross-chain swap is asynchronous: the client signs+broadcasts one SOURCE-chain
// tx, then the destination delivery is tracked by polling Status. Aggregator API
// keys live here and never reach the client.
package bridge

import (
	"context"
	"errors"
)

// ErrUnsupportedChain — the from/to chain isn't supported by the provider.
var ErrUnsupportedChain = errors.New("unsupported chain")

// ErrNoRoute — no bridge route exists for the requested pair/amount.
var ErrNoRoute = errors.New("no bridge route")

// Provider is a cross-chain bridge aggregator integration.
type Provider interface {
	Name() string
	SupportsChain(chainKey string) bool
	// Quote returns a source-chain signable tx + route metadata.
	Quote(ctx context.Context, req Request) (*Quote, error)
	// Status reports the cross-chain progress for a broadcast source tx.
	Status(ctx context.Context, fromChain, toChain, txHash, tool string) (*Status, error)
}
