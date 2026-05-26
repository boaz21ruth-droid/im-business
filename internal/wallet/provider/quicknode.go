package provider

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// QuickNode add-on API: qn_getWalletTokenTransactions
// https://marketplace.quicknode.com/add-ons/token-and-nft-api
// Requires per-chain endpoint URLs configured in the QuickNode dashboard.
// Covers ERC20 token transfers. Native transfers not supported by this add-on.

type QuickNodeProvider struct {
	endpoints map[string]string // chain → QuickNode endpoint URL
	client    *http.Client
}

func NewQuickNodeProvider(endpoints map[string]string) *QuickNodeProvider {
	return &QuickNodeProvider{endpoints: endpoints, client: &http.Client{Timeout: 15 * time.Second}}
}

func (p *QuickNodeProvider) Name() string { return "quicknode" }

type quicknodeTxResult struct {
	Address      string         `json:"address"`
	Transactions []quicknodeTx  `json:"transactions"`
}

type quicknodeTx struct {
	BlockNumber     int64   `json:"blockNumber"`
	Date            string  `json:"date"`            // ISO8601
	Hash            string  `json:"hash"`
	From            string  `json:"from"`
	To              string  `json:"to"`
	ContractAddress *string `json:"contractAddress"` // nil = native
	Value           string  `json:"value"`           // decimal string
	Symbol          string  `json:"symbol"`
	Decimals        int     `json:"decimals"`
}

// --- Stream provider (QuickNode Streams) ---

// quicknodeStreamNetworks maps our chain keys to QuickNode stream network names.
var quicknodeStreamNetworks = map[string]string{
	"eth":         "ethereum-mainnet",
	"bsc":         "bsc-mainnet",
	"polygon":     "polygon-mainnet",
	"arbitrum":    "arbitrum-mainnet",
	"optimism":    "optimism-mainnet",
	"eth_sepolia": "ethereum-sepolia",
	"bsc_testnet": "bsc-testnet",
}

var quicknodeNetworkToChainKey = map[string]string{
	"ethereum-mainnet": "eth",
	"bsc-mainnet":      "bsc",
	"polygon-mainnet":  "polygon",
	"arbitrum-mainnet": "arbitrum",
	"optimism-mainnet": "optimism",
	"ethereum-sepolia": "eth_sepolia",
	"bsc-testnet":      "bsc_testnet",
}

// QuickNodeStreamProvider implements BulkStreamProvider using QuickNode Streams.
// QuickNode Streams require regenerating the entire JavaScript filter function when
// the monitored address set changes, hence BulkStreamProvider (not incremental).
type QuickNodeStreamProvider struct {
	apiKey        string            // x-api-key for Streams management API
	webhookSecret string            // x-qn-secret sent with each webhook delivery
	webhookURL    string
	supportedNets map[string]bool   // which chain keys this instance covers
	client        *http.Client
}

func NewQuickNodeStreamProvider(apiKey, webhookSecret, webhookURL string, supportedChains []string) *QuickNodeStreamProvider {
	nets := make(map[string]bool, len(supportedChains))
	for _, c := range supportedChains {
		nets[c] = true
	}
	return &QuickNodeStreamProvider{
		apiKey:        apiKey,
		webhookSecret: webhookSecret,
		webhookURL:    webhookURL,
		supportedNets: nets,
		client:        &http.Client{Timeout: 20 * time.Second},
	}
}

func (p *QuickNodeStreamProvider) Name() string { return "quicknode" }

func (p *QuickNodeStreamProvider) EnsureStream(ctx context.Context, chainKey string) (string, error) {
	network, ok := quicknodeStreamNetworks[chainKey]
	if !ok || !p.supportedNets[chainKey] {
		return "", fmt.Errorf("quicknode stream: unsupported chain %s", chainKey)
	}

	filterFn := generateQNFilterFunction(nil)
	body, _ := json.Marshal(map[string]any{
		"name":     "im-wallet-" + chainKey,
		"network":  network,
		"dataset":  "logs",
		"isActive": true,
		// Infrastructure-level filter: only ERC20 Transfer events reach our function.
		"filters": []map[string]any{
			{"type": "log_sig", "value": "0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"},
		},
		"destinations": []map[string]any{
			{
				"type": "webhook",
				"name": "im-wallet-wh-" + chainKey,
				"payload": map[string]any{
					"type": "filter_function",
					"fields": map[string]any{
						"filter_function": base64.StdEncoding.EncodeToString([]byte(filterFn)),
					},
				},
				"destinationAttributes": map[string]any{
					"url":          p.webhookURL,
					"compression":  "none",
					"headers":      map[string]string{},
					"max_retry":    3,
					"post_timeout": 10,
					"secret":       p.webhookSecret,
				},
			},
		},
	})

	req, _ := http.NewRequestWithContext(ctx, "POST", "https://api.quicknode.com/streams/rest/v1/streams", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", p.apiKey)

	resp, err := p.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var out struct {
		ID    string `json:"id"`
		Error string `json:"message"`
	}
	json.NewDecoder(resp.Body).Decode(&out)
	if out.ID == "" {
		return "", fmt.Errorf("quicknode create stream failed: %s", out.Error)
	}
	return out.ID, nil
}

// AddAddressToStream delegates to SetStreamAddresses but cannot know the full list here.
// Callers that need incremental updates should use SetStreamAddresses directly via the
// BulkStreamProvider interface; the wallet service does this automatically.
func (p *QuickNodeStreamProvider) AddAddressToStream(_ context.Context, _ string, _ string) error {
	return nil // handled by SetStreamAddresses in wallet service
}

// SetStreamAddresses updates the QuickNode stream's filter function with the complete
// address list. This replaces the previous filter entirely.
func (p *QuickNodeStreamProvider) SetStreamAddresses(ctx context.Context, streamID string, addresses []string) error {
	filterFn := generateQNFilterFunction(addresses)
	body, _ := json.Marshal(map[string]any{
		"destinations": []map[string]any{
			{
				"payload": map[string]any{
					"type": "filter_function",
					"fields": map[string]any{
						"filter_function": base64.StdEncoding.EncodeToString([]byte(filterFn)),
					},
				},
			},
		},
	})

	url := "https://api.quicknode.com/streams/rest/v1/streams/" + streamID
	req, _ := http.NewRequestWithContext(ctx, "PATCH", url, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", p.apiKey)

	resp, err := p.client.Do(req)
	if err != nil {
		return err
	}
	resp.Body.Close()
	if resp.StatusCode >= 400 {
		return fmt.Errorf("quicknode update stream: HTTP %d", resp.StatusCode)
	}
	return nil
}

func (p *QuickNodeStreamProvider) VerifyWebhook(r *http.Request, body []byte) bool {
	if p.webhookSecret == "" {
		return true // no secret configured, pass through
	}
	return r.Header.Get("x-qn-secret") == p.webhookSecret
}

// quicknodeStreamPayload is the outer envelope QuickNode Streams sends to the webhook.
// The Data array contains objects returned by our filter function (one entry per
// matched ERC20 Transfer log).
type quicknodeStreamPayload struct {
	StreamID string `json:"streamId"`
	Network  string `json:"network"`
	Data     []struct {
		TxHash       string `json:"txHash"`
		From         string `json:"from"`
		To           string `json:"to"`
		Value        string `json:"value"`
		TokenAddress string `json:"tokenAddress"`
		BlockNumber  int64  `json:"blockNumber"`
	} `json:"data"`
}

func (p *QuickNodeStreamProvider) ParseWebhookPayload(body []byte) (chainKey string, confirmed bool, transfers []WebhookTransfer, err error) {
	var payload quicknodeStreamPayload
	if err = json.Unmarshal(body, &payload); err != nil {
		return
	}
	chainKey = quicknodeNetworkToChainKey[payload.Network]
	confirmed = true // QuickNode Streams only delivers confirmed blocks
	for _, d := range payload.Data {
		transfers = append(transfers, WebhookTransfer{
			TxHash:       d.TxHash,
			From:         d.From,
			To:           d.To,
			Value:        d.Value,
			Decimals:     18,
			TokenAddress: d.TokenAddress,
		})
	}
	return
}

// generateQNFilterFunction returns a JavaScript filter function for QuickNode Streams
// using the Logs dataset. Each item in data.streamData is already a single log entry;
// the infrastructure-level topic0 filter ensures only ERC20 Transfer events arrive.
func generateQNFilterFunction(addresses []string) string {
	addrsJSON, _ := json.Marshal(addresses)
	return fmt.Sprintf(`function main(data) {
    const ADDRS = new Set(%s);
    const results = [];
    for (const log of (data.streamData || [])) {
        if (!log.topics || log.topics.length < 3) continue;
        const from = '0x' + log.topics[1].slice(26).toLowerCase();
        const to   = '0x' + log.topics[2].slice(26).toLowerCase();
        if (!ADDRS.has(from) && !ADDRS.has(to)) continue;
        results.push({
            txHash: log.transactionHash,
            from: from, to: to,
            value: BigInt('0x' + log.data.slice(2)).toString(10),
            tokenAddress: log.address.toLowerCase(),
            blockNumber: parseInt(log.blockNumber, 16)
        });
    }
    return results;
}`, string(addrsJSON))
}

func (p *QuickNodeProvider) GetTransfers(ctx context.Context, req TransferRequest) ([]TxRecord, error) {
	rpcURL, ok := p.endpoints[req.Chain]
	if !ok || rpcURL == "" {
		return nil, fmt.Errorf("quicknode %w: %s not configured", ErrUnsupportedChain, req.Chain)
	}
	if req.ContractAddress == "" {
		return nil, fmt.Errorf("quicknode %w: native transfers not supported", ErrUnsupportedChain)
	}

	body, _ := json.Marshal(map[string]any{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "qn_getWalletTokenTransactions",
		"params":  []any{map[string]any{"wallet": req.Address, "page": req.Page, "perPage": req.Limit}},
	})
	httpReq, _ := http.NewRequestWithContext(ctx, "POST", rpcURL, bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")
	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var wrapper struct {
		Result quicknodeTxResult `json:"result"`
		Error  *struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&wrapper); err != nil {
		return nil, err
	}
	if wrapper.Error != nil {
		return nil, fmt.Errorf("quicknode rpc %d: %s", wrapper.Error.Code, wrapper.Error.Message)
	}

	records := make([]TxRecord, 0, len(wrapper.Result.Transactions))
	for _, t := range wrapper.Result.Transactions {
		if t.ContractAddress == nil {
			continue // skip native
		}
		// filter by contract address if specified
		if req.ContractAddress != "" && *t.ContractAddress != req.ContractAddress {
			continue
		}
		ts, _ := time.Parse(time.RFC3339, t.Date)
		sym := t.Symbol
		contract := *t.ContractAddress
		records = append(records, TxRecord{
			Hash:           t.Hash,
			From:           t.From,
			To:             t.To,
			Value:          t.Value,
			Decimals:       t.Decimals,
			TokenSymbol:    &sym,
			TokenContract:  &contract,
			BlockNumber:    t.BlockNumber,
			BlockTimestamp: ts,
			ChainKey:       req.Chain,
			Source:         "quicknode",
		})
	}
	return records, nil
}
