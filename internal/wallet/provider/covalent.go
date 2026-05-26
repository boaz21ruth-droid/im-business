package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

var covalentChains = map[string]string{
	"eth":         "eth-mainnet",
	"bsc":         "bsc-mainnet",
	"polygon":     "matic-mainnet",
	"arbitrum":    "arbitrum-mainnet",
	"optimism":    "optimism-mainnet",
	"bsc_testnet": "bsc-testnet",
	"eth_sepolia": "eth-sepolia-testnet",
}

type CovalentProvider struct {
	mainPool *keyPool
	testPool *keyPool
	client   *http.Client
}

func NewCovalentProvider(mainKeys, testKeys []string) *CovalentProvider {
	m := newKeyPool(mainKeys)
	t := m
	if len(testKeys) > 0 {
		t = newKeyPool(testKeys)
	}
	return &CovalentProvider{mainPool: m, testPool: t, client: &http.Client{Timeout: 15 * time.Second}}
}

func (p *CovalentProvider) Name() string { return "covalent" }

func (p *CovalentProvider) GetTransfers(ctx context.Context, req TransferRequest) ([]TxRecord, error) {
	chainName, ok := covalentChains[req.Chain]
	if !ok {
		return nil, fmt.Errorf("covalent %w: %s", ErrUnsupportedChain, req.Chain)
	}
	pool := p.mainPool
	if isTestnet[req.Chain] {
		pool = p.testPool
	}
	key := pool.pick()
	if key == "" {
		return nil, fmt.Errorf("covalent %w: api_key not configured", ErrUnsupportedChain)
	}

	url := fmt.Sprintf(
		"https://api.covalenthq.com/v1/%s/address/%s/transfers_v2/?page-size=%d&page-number=%d",
		chainName, req.Address, req.Limit, req.Page,
	)
	if req.ContractAddress != "" {
		url += "&contract-address=" + req.ContractAddress
	}

	httpReq, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
	httpReq.SetBasicAuth(key, "")

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var out struct {
		Data *struct {
			Items []covalentItem `json:"items"`
		} `json:"data"`
		Error        bool   `json:"error"`
		ErrorMessage string `json:"error_message"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	if out.Error {
		return nil, fmt.Errorf("covalent: %s", out.ErrorMessage)
	}
	if out.Data == nil {
		return nil, nil
	}

	var records []TxRecord
	for _, item := range out.Data.Items {
		for _, t := range item.Transfers {
			records = append(records, t.toRecord(item, req.Chain))
		}
	}
	return records, nil
}

type covalentItem struct {
	BlockSignedAt string             `json:"block_signed_at"`
	BlockHeight   int64              `json:"block_height"`
	TxHash        string             `json:"tx_hash"`
	FromAddress   string             `json:"from_address"`
	ToAddress     string             `json:"to_address"`
	Transfers     []covalentTransfer `json:"transfers"`
}

type covalentTransfer struct {
	ContractDecimals int    `json:"contract_decimals"`
	ContractTicker   string `json:"contract_ticker_symbol"`
	ContractAddress  string `json:"contract_address"`
	Delta            string `json:"delta"`
	FromAddress      string `json:"from_address"`
	ToAddress        string `json:"to_address"`
}

func (t covalentTransfer) toRecord(item covalentItem, chain string) TxRecord {
	from := t.FromAddress
	if from == "" {
		from = item.FromAddress
	}
	to := t.ToAddress
	if to == "" {
		to = item.ToAddress
	}
	ts, _ := time.Parse(time.RFC3339, item.BlockSignedAt)
	decimals := t.ContractDecimals
	if decimals == 0 {
		decimals = 18
	}
	var sym, addr *string
	if t.ContractTicker != "" {
		s := t.ContractTicker
		sym = &s
	}
	if t.ContractAddress != "" {
		a := t.ContractAddress
		addr = &a
	}
	return TxRecord{
		Hash: item.TxHash, From: from, To: to,
		Value: t.Delta, Decimals: decimals,
		TokenSymbol: sym, TokenContract: addr,
		BlockNumber: item.BlockHeight, BlockTimestamp: ts,
		ChainKey: chain, Source: "covalent",
	}
}
