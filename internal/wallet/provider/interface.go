package provider

import (
	"context"
	"errors"
	"net/http"
	"time"
)

// ErrUnsupportedChain is returned by providers when they don't support the requested chain.
// MultiProvider treats this as an expected skip (logged at Debug, not Warn).
var ErrUnsupportedChain = errors.New("unsupported chain")

type TransferRequest struct {
	Chain           string
	Address         string
	ContractAddress string // empty = native transfers
	Page            int
	Limit           int
	MinTimestamp    int64 // Unix ms; used by TronGrid poller for cursor-based fetching
}

type TxRecord struct {
	Hash           string
	From           string
	To             string
	Value          string // raw BigInt as decimal string
	Decimals       int
	TokenSymbol    *string
	TokenContract  *string
	BlockNumber    int64
	BlockTimestamp time.Time
	ChainKey       string
	Source         string
}

type TxProvider interface {
	GetTransfers(ctx context.Context, req TransferRequest) ([]TxRecord, error)
	Name() string
}

// WebhookTransfer is the unified transfer record from any stream provider webhook.
type WebhookTransfer struct {
	TxHash       string
	From         string
	To           string
	Value        string // raw BigInt as decimal string
	Decimals     int
	TokenSymbol  string
	TokenAddress string // empty = native transfer
}

// StreamProvider handles real-time webhook-based transaction notifications.
type StreamProvider interface {
	Name() string
	EnsureStream(ctx context.Context, chainKey string) (string, error)
	AddAddressToStream(ctx context.Context, streamID, address string) error
	VerifyWebhook(r *http.Request, body []byte) bool
	ParseWebhookPayload(body []byte) (chainKey string, confirmed bool, transfers []WebhookTransfer, err error)
}

// BulkStreamProvider extends StreamProvider for providers that require the entire address list
// to be replaced rather than added incrementally (e.g. QuickNode Streams filter functions).
type BulkStreamProvider interface {
	StreamProvider
	SetStreamAddresses(ctx context.Context, streamID string, addresses []string) error
}

// WSProvider subscribes to on-chain events via a persistent outbound WebSocket connection.
// Unlike StreamProvider (inbound webhooks), the provider dials the RPC node directly —
// no public URL or webhook registration required.
type WSProvider interface {
	Name() string
	// AddAddress registers an address to watch; safe to call concurrently.
	AddAddress(chainKey, address string)
	// Start blocks and reconnects on disconnect; run as a goroutine.
	Start(ctx context.Context, onTransfer func(chainKey string, t WebhookTransfer))
}
