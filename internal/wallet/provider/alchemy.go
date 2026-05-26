package provider

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"time"
)

var alchemyNetworks = map[string]string{
	"eth":         "eth-mainnet",
	"polygon":     "polygon-mainnet",
	"arbitrum":    "arb-mainnet",
	"optimism":    "opt-mainnet",
	"eth_sepolia": "eth-sepolia",
}

type AlchemyProvider struct {
	mainPool  *keyPool
	testPool  *keyPool
	endpoints map[string]string // per-chain endpoint URLs (takes precedence over api_key)
	client    *http.Client
}

func NewAlchemyProvider(mainKeys, testKeys []string, endpoints map[string]string) *AlchemyProvider {
	m := newKeyPool(mainKeys)
	t := m
	if len(testKeys) > 0 {
		t = newKeyPool(testKeys)
	}
	return &AlchemyProvider{mainPool: m, testPool: t, endpoints: endpoints, client: &http.Client{Timeout: 15 * time.Second}}
}

func (p *AlchemyProvider) Name() string { return "alchemy" }

func (p *AlchemyProvider) GetTransfers(ctx context.Context, req TransferRequest) ([]TxRecord, error) {
	// Per-chain endpoint URL takes precedence over the shared api_key approach.
	var url string
	if ep := p.endpoints[req.Chain]; ep != "" {
		url = ep
	} else {
		network, ok := alchemyNetworks[req.Chain]
		if !ok {
			return nil, fmt.Errorf("alchemy %w: %s", ErrUnsupportedChain, req.Chain)
		}
		pool := p.mainPool
		if isTestnet[req.Chain] {
			pool = p.testPool
		}
		key := pool.pick()
		if key == "" {
			return nil, fmt.Errorf("alchemy %w: not configured for chain %s", ErrUnsupportedChain, req.Chain)
		}
		url = fmt.Sprintf("https://%s.g.alchemy.com/v2/%s", network, key)
	}

	base := map[string]any{
		"fromBlock":    "0x0",
		"toBlock":      "latest",
		"withMetadata": true,
		"order":        "desc",
		"maxCount":     fmt.Sprintf("0x%x", req.Limit),
	}
	if req.ContractAddress != "" {
		base["category"] = []string{"erc20"}
		base["contractAddresses"] = []string{req.ContractAddress}
	} else {
		base["category"] = []string{"external"}
	}

	seen := make(map[string]bool)
	var all []TxRecord

	for _, addrKey := range []string{"toAddress", "fromAddress"} {
		params := make(map[string]any, len(base)+1)
		for k, v := range base {
			params[k] = v
		}
		params[addrKey] = req.Address

		body, _ := json.Marshal(map[string]any{
			"id": 1, "jsonrpc": "2.0",
			"method": "alchemy_getAssetTransfers",
			"params": []any{params},
		})
		httpReq, _ := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
		httpReq.Header.Set("Content-Type", "application/json")

		resp, err := p.client.Do(httpReq)
		if err != nil {
			return nil, err
		}
		var out struct {
			Result *struct {
				Transfers []alchemyTransfer `json:"transfers"`
			} `json:"result"`
			Error *struct{ Message string `json:"message"` } `json:"error"`
		}
		json.NewDecoder(resp.Body).Decode(&out)
		resp.Body.Close()

		if out.Error != nil {
			return nil, fmt.Errorf("alchemy: %s", out.Error.Message)
		}
		if out.Result == nil {
			continue
		}
		for _, t := range out.Result.Transfers {
			if seen[t.Hash] {
				continue
			}
			seen[t.Hash] = true
			all = append(all, t.toRecord(req.Chain))
		}
	}
	return all, nil
}

type alchemyTransfer struct {
	Hash     string  `json:"hash"`
	From     string  `json:"from"`
	To       string  `json:"to"`
	Value    float64 `json:"value"`
	Asset    string  `json:"asset"`
	BlockNum string  `json:"blockNum"`
	RawContract struct {
		Value   string `json:"value"`
		Address string `json:"address"`
		Decimal string `json:"decimal"`
	} `json:"rawContract"`
	Metadata struct {
		BlockTimestamp string `json:"blockTimestamp"`
	} `json:"metadata"`
}

// --- Stream provider (Alchemy Notify / ADDRESS_ACTIVITY webhooks) ---

// alchemyWebhookNetworks maps our chain keys to Alchemy Notify network names.
var alchemyWebhookNetworks = map[string]string{
	"eth":         "ETH_MAINNET",
	"polygon":     "MATIC_MAINNET",
	"arbitrum":    "ARB_MAINNET",
	"optimism":    "OPT_MAINNET",
	"eth_sepolia": "ETH_SEPOLIA",
}

var alchemyWebhookNetworkToKey = map[string]string{
	"ETH_MAINNET":   "eth",
	"MATIC_MAINNET": "polygon",
	"ARB_MAINNET":   "arbitrum",
	"OPT_MAINNET":   "optimism",
	"ETH_SEPOLIA":   "eth_sepolia",
}

// AlchemyStreamProvider implements StreamProvider using Alchemy Notify ADDRESS_ACTIVITY webhooks.
type AlchemyStreamProvider struct {
	authToken  string // X-Alchemy-Token for Notify management API
	signingKey string // HMAC-SHA256 signing key provided by Alchemy
	webhookURL string
	client     *http.Client
}

func NewAlchemyStreamProvider(authToken, signingKey, webhookURL string) *AlchemyStreamProvider {
	return &AlchemyStreamProvider{
		authToken:  authToken,
		signingKey: signingKey,
		webhookURL: webhookURL,
		client:     &http.Client{Timeout: 15 * time.Second},
	}
}

func (p *AlchemyStreamProvider) Name() string { return "alchemy" }

func (p *AlchemyStreamProvider) EnsureStream(ctx context.Context, chainKey string) (string, error) {
	network, ok := alchemyWebhookNetworks[chainKey]
	if !ok {
		return "", fmt.Errorf("alchemy stream: unsupported chain %s", chainKey)
	}
	body, _ := json.Marshal(map[string]any{
		"network":      network,
		"webhook_type": "ADDRESS_ACTIVITY",
		"webhook_url":  p.webhookURL,
		"addresses":    []string{},
	})
	req, _ := http.NewRequestWithContext(ctx, "POST", "https://dashboard.alchemy.com/api/create-webhook", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Alchemy-Token", p.authToken)

	resp, err := p.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var out struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
		Error *struct{ Message string `json:"message"` } `json:"error"`
	}
	json.NewDecoder(resp.Body).Decode(&out)
	if out.Data.ID == "" {
		msg := ""
		if out.Error != nil {
			msg = out.Error.Message
		}
		return "", fmt.Errorf("alchemy create webhook failed: %s", msg)
	}
	return out.Data.ID, nil
}

func (p *AlchemyStreamProvider) AddAddressToStream(ctx context.Context, streamID, address string) error {
	body, _ := json.Marshal(map[string]any{
		"webhook_id":          streamID,
		"addresses_to_add":    []string{address},
		"addresses_to_remove": []string{},
	})
	req, _ := http.NewRequestWithContext(ctx, "POST", "https://dashboard.alchemy.com/api/update-webhook-addresses", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Alchemy-Token", p.authToken)

	resp, err := p.client.Do(req)
	if err != nil {
		return err
	}
	resp.Body.Close()
	if resp.StatusCode >= 400 {
		return fmt.Errorf("alchemy update addresses: HTTP %d", resp.StatusCode)
	}
	return nil
}

func (p *AlchemyStreamProvider) VerifyWebhook(r *http.Request, body []byte) bool {
	if p.signingKey == "" {
		return false
	}
	sig := r.Header.Get("x-alchemy-signature")
	mac := hmac.New(sha256.New, []byte(p.signingKey))
	mac.Write(body)
	computed := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(computed), []byte(sig))
}

type alchemyWebhookPayload struct {
	Type  string `json:"type"`
	Event struct {
		Network  string `json:"network"`
		Activity []struct {
			FromAddress string  `json:"fromAddress"`
			ToAddress   string  `json:"toAddress"`
			BlockNum    string  `json:"blockNum"`
			Hash        string  `json:"hash"`
			Value       float64 `json:"value"`
			Asset       string  `json:"asset"`
			RawContract struct {
				Value   string `json:"value"`
				Address string `json:"address"`
				Decimal string `json:"decimal"`
			} `json:"rawContract"`
		} `json:"activity"`
	} `json:"event"`
}

func (p *AlchemyStreamProvider) ParseWebhookPayload(body []byte) (chainKey string, confirmed bool, transfers []WebhookTransfer, err error) {
	var payload alchemyWebhookPayload
	if err = json.Unmarshal(body, &payload); err != nil {
		return
	}
	chainKey = alchemyWebhookNetworkToKey[payload.Event.Network]
	confirmed = true // Alchemy only sends confirmed transactions

	for _, a := range payload.Event.Activity {
		decimals := 18
		if a.RawContract.Decimal != "" {
			d := new(big.Int)
			d.SetString(strings.TrimPrefix(a.RawContract.Decimal, "0x"), 16)
			decimals = int(d.Int64())
		}
		value := ""
		if a.RawContract.Value != "" {
			n := new(big.Int)
			n.SetString(strings.TrimPrefix(a.RawContract.Value, "0x"), 16)
			value = n.String()
		} else {
			n := new(big.Float).SetFloat64(a.Value)
			n.Mul(n, new(big.Float).SetFloat64(1e18))
			i, _ := n.Int(nil)
			value = i.String()
		}
		transfers = append(transfers, WebhookTransfer{
			TxHash:       a.Hash,
			From:         a.FromAddress,
			To:           a.ToAddress,
			Value:        value,
			Decimals:     decimals,
			TokenSymbol:  a.Asset,
			TokenAddress: a.RawContract.Address,
		})
	}
	return
}

func (t alchemyTransfer) toRecord(chain string) TxRecord {
	r := TxRecord{Hash: t.Hash, From: t.From, To: t.To, ChainKey: chain, Source: "alchemy"}

	if t.BlockNum != "" {
		n := new(big.Int)
		n.SetString(strings.TrimPrefix(t.BlockNum, "0x"), 16)
		r.BlockNumber = n.Int64()
	}
	if t.Metadata.BlockTimestamp != "" {
		ts, _ := time.Parse(time.RFC3339, t.Metadata.BlockTimestamp)
		r.BlockTimestamp = ts
	}
	if t.RawContract.Value != "" {
		n := new(big.Int)
		n.SetString(strings.TrimPrefix(t.RawContract.Value, "0x"), 16)
		r.Value = n.String()
	} else {
		n := new(big.Float).SetFloat64(t.Value)
		n.Mul(n, new(big.Float).SetFloat64(1e18))
		i, _ := n.Int(nil)
		r.Value = i.String()
	}
	if t.RawContract.Decimal != "" {
		d := new(big.Int)
		d.SetString(strings.TrimPrefix(t.RawContract.Decimal, "0x"), 16)
		r.Decimals = int(d.Int64())
	} else {
		r.Decimals = 18
	}
	if t.RawContract.Address != "" {
		addr := t.RawContract.Address
		sym := t.Asset
		r.TokenContract = &addr
		r.TokenSymbol = &sym
	}
	return r
}
