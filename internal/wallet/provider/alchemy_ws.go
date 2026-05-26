package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

// AlchemyWSProvider implements WSProvider by subscribing to eth_subscribe("logs") over
// a persistent WebSocket connection. Uses the same api_key / endpoints as AlchemyProvider;
// no auth_token or public webhook URL is required.
type AlchemyWSProvider struct {
	mainPool  *keyPool
	endpoints map[string]string // optional per-chain WSS URL override

	mu      sync.RWMutex
	watched map[string]map[string]bool // chain → lowercase address set

	log *zap.Logger
}

func NewAlchemyWSProvider(mainKeys []string, endpoints map[string]string, log *zap.Logger) *AlchemyWSProvider {
	return &AlchemyWSProvider{
		mainPool:  newKeyPool(mainKeys),
		endpoints: endpoints,
		watched:   make(map[string]map[string]bool),
		log:       log,
	}
}

func (p *AlchemyWSProvider) Name() string { return "alchemy_ws" }

func (p *AlchemyWSProvider) AddAddress(chainKey, address string) {
	lower := strings.ToLower(address)
	p.mu.Lock()
	if p.watched[chainKey] == nil {
		p.watched[chainKey] = make(map[string]bool)
	}
	p.watched[chainKey][lower] = true
	p.mu.Unlock()
}

// Start launches one goroutine per supported Alchemy chain.
func (p *AlchemyWSProvider) Start(ctx context.Context, onTransfer func(chainKey string, t WebhookTransfer)) {
	for chainKey := range alchemyNetworks {
		go p.listenChain(ctx, chainKey, onTransfer)
	}
	<-ctx.Done()
}

func (p *AlchemyWSProvider) wssURL(chainKey string) string {
	// Per-chain endpoint takes precedence; convert https→wss if needed.
	if ep := p.endpoints[chainKey]; ep != "" {
		u := strings.Replace(ep, "https://", "wss://", 1)
		return strings.Replace(u, "http://", "ws://", 1)
	}
	key := p.mainPool.pick()
	if key == "" {
		return ""
	}
	return fmt.Sprintf("wss://%s.g.alchemy.com/v2/%s", alchemyNetworks[chainKey], key)
}

func (p *AlchemyWSProvider) listenChain(ctx context.Context, chainKey string, onTransfer func(string, WebhookTransfer)) {
	backoff := 2 * time.Second
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		wssURL := p.wssURL(chainKey)
		if wssURL == "" {
			return // api_key not configured
		}

		if err := p.connectAndListen(ctx, chainKey, wssURL, onTransfer); err != nil && ctx.Err() == nil {
			p.log.Warn("alchemy ws disconnected",
				zap.String("chain", chainKey), zap.Duration("retry_in", backoff), zap.Error(err))
			select {
			case <-ctx.Done():
				return
			case <-time.After(backoff):
			}
			if backoff < 2*time.Minute {
				backoff *= 2
			}
		} else {
			backoff = 2 * time.Second
		}
	}
}

type alchemySubResult struct {
	Address         string   `json:"address"`
	Topics          []string `json:"topics"`
	Data            string   `json:"data"`
	TransactionHash string   `json:"transactionHash"`
	Removed         bool     `json:"removed"`
}

type alchemyWSMsg struct {
	Method string `json:"method"`
	Params *struct {
		Result alchemySubResult `json:"result"`
	} `json:"params"`
}

func (p *AlchemyWSProvider) connectAndListen(ctx context.Context, chainKey, wssURL string, onTransfer func(string, WebhookTransfer)) error {
	conn, _, err := websocket.DefaultDialer.DialContext(ctx, wssURL, nil)
	if err != nil {
		return err
	}
	defer conn.Close()

	// Close connection when context is cancelled.
	go func() {
		<-ctx.Done()
		conn.Close()
	}()

	// Subscribe to all ERC20 Transfer events on this chain.
	sub, _ := json.Marshal(map[string]any{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "eth_subscribe",
		"params":  []any{"logs", map[string]any{"topics": []string{erc20TransferTopic}}},
	})
	if err := conn.WriteMessage(websocket.TextMessage, sub); err != nil {
		return err
	}
	p.log.Info("alchemy ws subscribed", zap.String("chain", chainKey))

	for {
		_, raw, err := conn.ReadMessage()
		if err != nil {
			return err
		}

		var msg alchemyWSMsg
		if err := json.Unmarshal(raw, &msg); err != nil || msg.Method != "eth_subscription" || msg.Params == nil {
			continue
		}
		r := msg.Params.Result
		if r.Removed || len(r.Topics) < 3 {
			continue
		}

		from := "0x" + strings.ToLower(r.Topics[1][len(r.Topics[1])-40:])
		to := "0x" + strings.ToLower(r.Topics[2][len(r.Topics[2])-40:])

		p.mu.RLock()
		set := p.watched[chainKey]
		relevant := set != nil && (set[from] || set[to])
		p.mu.RUnlock()

		if !relevant {
			continue
		}

		onTransfer(chainKey, WebhookTransfer{
			TxHash:       r.TransactionHash,
			From:         from,
			To:           to,
			Value:        hexToBigDecStr(r.Data),
			Decimals:     18,
			TokenAddress: strings.ToLower(r.Address),
		})
	}
}
