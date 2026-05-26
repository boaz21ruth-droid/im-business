package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// NodeReal Enhanced API: https://docs.nodereal.io/reference/enhanced-api
// Supports ETH, BSC (including testnet), Polygon, Arbitrum, Optimism.
var noderealNetworks = map[string]string{
	"eth":         "eth-mainnet",
	"bsc":         "bsc-mainnet",
	"bsc_testnet": "bsc-testnet",
	"polygon":     "polygon-mainnet",
	"arbitrum":    "arbitrum-one",
	"optimism":    "optimism-mainnet",
	"eth_sepolia": "eth-sepolia",
}

type NodeRealProvider struct {
	mainPool *keyPool
	testPool *keyPool
	client   *http.Client
}

func NewNodeRealProvider(mainKeys, testKeys []string) *NodeRealProvider {
	m := newKeyPool(mainKeys)
	t := m
	if len(testKeys) > 0 {
		t = newKeyPool(testKeys)
	}
	return &NodeRealProvider{mainPool: m, testPool: t, client: &http.Client{Timeout: 15 * time.Second}}
}

func (p *NodeRealProvider) Name() string { return "nodereal" }

func (p *NodeRealProvider) GetTransfers(ctx context.Context, req TransferRequest) ([]TxRecord, error) {
	pool := p.mainPool
	if isTestnet[req.Chain] {
		pool = p.testPool
	}
	key := pool.pick()
	if key == "" {
		return nil, fmt.Errorf("nodereal %w: api_key not configured", ErrUnsupportedChain)
	}
	network, ok := noderealNetworks[req.Chain]
	if !ok {
		return nil, fmt.Errorf("nodereal %w: %s", ErrUnsupportedChain, req.Chain)
	}
	if req.ContractAddress == "" {
		return p.getNative(ctx, network, key, req)
	}
	return p.getERC20(ctx, network, key, req)
}

func (p *NodeRealProvider) call(ctx context.Context, network, key, method string, params any, result any) error {
	body, _ := json.Marshal(map[string]any{
		"jsonrpc": "2.0", "method": method,
		"params": []any{params}, "id": 1,
	})
	endpoint := fmt.Sprintf("https://%s.nodereal.io/v1/%s", network, key)
	req, _ := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := p.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	var wrapper struct {
		Result json.RawMessage                                   `json:"result"`
		Error  *struct{ Message string `json:"message"` } `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&wrapper); err != nil {
		return err
	}
	if wrapper.Error != nil {
		return fmt.Errorf("nodereal: %s", wrapper.Error.Message)
	}
	return json.Unmarshal(wrapper.Result, result)
}

type noderealERC20Transfer struct {
	Hash            string `json:"hash"`
	From            string `json:"from"`
	To              string `json:"to"`
	ContractAddress string `json:"contractAddress"`
	Value           string `json:"value"`
	TokenSymbol     string `json:"tokenSymbol"`
	TokenDecimal    string `json:"tokenDecimal"`
	BlockNumber     string `json:"blockNumber"`
	TimeStamp       string `json:"timeStamp"`
}

func (p *NodeRealProvider) getERC20(ctx context.Context, network, key string, req TransferRequest) ([]TxRecord, error) {
	params := map[string]any{
		"address":         req.Address,
		"contractAddress": req.ContractAddress,
		"pageSize":        fmt.Sprintf("0x%x", req.Limit),
		"pageIndex":       "0x0",
	}
	var result struct {
		Details []noderealERC20Transfer `json:"details"`
	}
	if err := p.call(ctx, network, key, "nr_getTokenTransfers", params, &result); err != nil {
		return nil, err
	}
	records := make([]TxRecord, 0, len(result.Details))
	for _, t := range result.Details {
		sym := t.TokenSymbol
		addr := t.ContractAddress
		decimals := 18
		fmt.Sscanf(t.TokenDecimal, "%d", &decimals)
		blockNum := int64(0)
		fmt.Sscanf(t.BlockNumber, "%d", &blockNum)
		ts := int64(0)
		fmt.Sscanf(t.TimeStamp, "%d", &ts)
		records = append(records, TxRecord{
			Hash: t.Hash, From: t.From, To: t.To, Value: t.Value,
			Decimals: decimals, TokenSymbol: &sym, TokenContract: &addr,
			BlockNumber: blockNum, BlockTimestamp: unixSecToTime(ts),
			ChainKey: req.Chain, Source: "nodereal",
		})
	}
	return records, nil
}

type noderealNativeTx struct {
	Hash        string `json:"hash"`
	From        string `json:"from"`
	To          string `json:"to"`
	Value       string `json:"value"`
	BlockNumber string `json:"blockNumber"`
	TimeStamp   string `json:"timeStamp"`
}

func (p *NodeRealProvider) getNative(ctx context.Context, network, key string, req TransferRequest) ([]TxRecord, error) {
	params := map[string]any{
		"address":   req.Address,
		"pageSize":  fmt.Sprintf("0x%x", req.Limit),
		"pageIndex": "0x0",
	}
	var result struct {
		TxList []noderealNativeTx `json:"txList"`
	}
	if err := p.call(ctx, network, key, "nr_getTxByAccount", params, &result); err != nil {
		return nil, err
	}
	records := make([]TxRecord, 0, len(result.TxList))
	for _, t := range result.TxList {
		records = append(records, TxRecord{
			Hash: t.Hash, From: t.From, To: t.To,
			Value:          hexToBigDecStr(t.Value),
			Decimals:       18,
			BlockNumber:    hexToInt64(t.BlockNumber),
			BlockTimestamp: unixSecToTime(hexToInt64(t.TimeStamp)),
			ChainKey:       req.Chain,
			Source:         "nodereal",
		})
	}
	return records, nil
}
