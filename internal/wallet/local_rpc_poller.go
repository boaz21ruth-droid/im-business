package wallet

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/web1/im-business/internal/repo"
	"github.com/web1/im-business/internal/wallet/provider"
)

const localRpcPollInterval = 2 * time.Second

// LocalRpcPoller polls a local EVM-compatible RPC (e.g. Anvil fork) for new blocks
// and feeds matching transactions into ProcessWebhookTransfer so they are stored
// in the DB and trigger IM notifications — identical to the Moralis/Alchemy path.
type LocalRpcPoller struct {
	chainKey   string
	rpcURL     string
	mu         sync.RWMutex
	addresses  map[string]struct{} // monitored wallet addresses, all lowercase
	repo       *repo.WalletRepo
	onTransfer func(ctx context.Context, chainKey string, t provider.WebhookTransfer)
	log        *zap.Logger
	client     *http.Client
	rpcSeq     int
}

func NewLocalRpcPoller(chainKey, rpcURL string, walletRepo *repo.WalletRepo, log *zap.Logger) *LocalRpcPoller {
	return &LocalRpcPoller{
		chainKey:   chainKey,
		rpcURL:     rpcURL,
		addresses:  make(map[string]struct{}),
		repo:       walletRepo,
		onTransfer: func(_ context.Context, _ string, _ provider.WebhookTransfer) {},
		log:        log,
		client:     &http.Client{Timeout: 10 * time.Second},
	}
}

// ChainKey returns the chain this poller monitors.
func (p *LocalRpcPoller) ChainKey() string { return p.chainKey }

// SetOnTransfer wires in the transfer handler (call after WalletService is created).
func (p *LocalRpcPoller) SetOnTransfer(fn func(ctx context.Context, chainKey string, t provider.WebhookTransfer)) {
	p.onTransfer = fn
}

// AddAddress registers an address for monitoring. Safe to call concurrently.
func (p *LocalRpcPoller) AddAddress(address string) {
	p.mu.Lock()
	p.addresses[strings.ToLower(address)] = struct{}{}
	p.mu.Unlock()
}

// loadAddresses pre-populates the address set from DB so existing users are
// covered without needing to re-register after a server restart.
func (p *LocalRpcPoller) loadAddresses() {
	addrs, err := p.repo.GetAddressesByChain(p.chainKey)
	if err != nil {
		p.log.Warn("local rpc poller: load addresses", zap.String("chain", p.chainKey), zap.Error(err))
		return
	}
	p.mu.Lock()
	for _, a := range addrs {
		p.addresses[strings.ToLower(a)] = struct{}{}
	}
	p.mu.Unlock()
	p.log.Info("local rpc poller: loaded addresses", zap.String("chain", p.chainKey), zap.Int("count", len(addrs)))
}

// Start begins polling. Blocks until ctx is done.
func (p *LocalRpcPoller) Start(ctx context.Context) {
	p.loadAddresses()

	startBlock, err := p.blockNumber()
	if err != nil {
		p.log.Error("local rpc poller: cannot get start block — aborting", zap.String("chain", p.chainKey), zap.Error(err))
		return
	}
	nextBlock := startBlock
	p.log.Info("local rpc poller: started", zap.String("chain", p.chainKey), zap.String("rpc", p.rpcURL), zap.Uint64("from_block", nextBlock))

	ticker := time.NewTicker(localRpcPollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			current, err := p.blockNumber()
			if err != nil {
				p.log.Warn("local rpc poller: block number", zap.String("chain", p.chainKey), zap.Error(err))
				continue
			}
			for b := nextBlock; b <= current; b++ {
				if err := p.processBlock(ctx, b); err != nil {
					p.log.Warn("local rpc poller: process block", zap.String("chain", p.chainKey), zap.Uint64("block", b), zap.Error(err))
				}
			}
			if current >= nextBlock {
				nextBlock = current + 1
			}
		}
	}
}

func (p *LocalRpcPoller) processBlock(ctx context.Context, blockNum uint64) error {
	block, err := p.getBlock(blockNum)
	if err != nil {
		return err
	}

	p.mu.RLock()
	addrSnap := make(map[string]struct{}, len(p.addresses))
	for k := range p.addresses {
		addrSnap[k] = struct{}{}
	}
	p.mu.RUnlock()

	for _, tx := range block.Transactions {
		if tx.To == nil {
			continue // contract creation — no recipient
		}
		from := strings.ToLower(tx.From)
		to := strings.ToLower(*tx.To)
		_, fromHit := addrSnap[from]
		_, toHit := addrSnap[to]
		if !fromHit && !toHit {
			continue
		}
		p.onTransfer(ctx, p.chainKey, provider.WebhookTransfer{
			TxHash:   tx.Hash,
			From:     tx.From,
			To:       *tx.To,
			Value:    hexToDecimal(tx.Value),
			Decimals: 18,
		})
	}
	return nil
}

// ── JSON-RPC plumbing ────────────────────────────────────────────────────────

type lrpcRequest struct {
	JSONRPC string `json:"jsonrpc"`
	Method  string `json:"method"`
	Params  []any  `json:"params"`
	ID      int    `json:"id"`
}

type lrpcResponse struct {
	Result json.RawMessage `json:"result"`
	Error  *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

type lrpcBlock struct {
	Transactions []lrpcTx `json:"transactions"`
}

type lrpcTx struct {
	Hash  string  `json:"hash"`
	From  string  `json:"from"`
	To    *string `json:"to"`
	Value string  `json:"value"`
}

func (p *LocalRpcPoller) call(method string, params []any) (json.RawMessage, error) {
	p.rpcSeq++
	body, _ := json.Marshal(lrpcRequest{JSONRPC: "2.0", Method: method, Params: params, ID: p.rpcSeq})
	resp, err := p.client.Post(p.rpcURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("local rpc %s: %w", method, err)
	}
	defer resp.Body.Close()
	var out lrpcResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("local rpc decode %s: %w", method, err)
	}
	if out.Error != nil {
		return nil, fmt.Errorf("local rpc %s: code=%d %s", method, out.Error.Code, out.Error.Message)
	}
	return out.Result, nil
}

func (p *LocalRpcPoller) blockNumber() (uint64, error) {
	raw, err := p.call("eth_blockNumber", nil)
	if err != nil {
		return 0, err
	}
	var hex string
	if err := json.Unmarshal(raw, &hex); err != nil {
		return 0, err
	}
	return lrpcParseHex(hex)
}

func (p *LocalRpcPoller) getBlock(num uint64) (*lrpcBlock, error) {
	raw, err := p.call("eth_getBlockByNumber", []any{fmt.Sprintf("0x%x", num), true})
	if err != nil {
		return nil, err
	}
	if string(raw) == "null" {
		return &lrpcBlock{}, nil
	}
	var block lrpcBlock
	if err := json.Unmarshal(raw, &block); err != nil {
		return nil, err
	}
	return &block, nil
}

func lrpcParseHex(hex string) (uint64, error) {
	hex = strings.TrimPrefix(hex, "0x")
	if hex == "" {
		return 0, nil
	}
	var n uint64
	_, err := fmt.Sscanf(hex, "%x", &n)
	return n, err
}

func hexToDecimal(hex string) string {
	hex = strings.TrimPrefix(hex, "0x")
	if hex == "" {
		return "0"
	}
	n, ok := new(big.Int).SetString(hex, 16)
	if !ok {
		return "0"
	}
	return n.String()
}
