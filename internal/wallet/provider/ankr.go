package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"time"
)

var ankrBlockchains = map[string]string{
	"eth":      "eth",
	"bsc":      "bsc",
	"polygon":  "polygon",
	"arbitrum": "arbitrum",
	"optimism": "optimism",
	// bsc_testnet and eth_sepolia omitted — not supported by Ankr Advanced API
}

type AnkrProvider struct {
	mainPool *keyPool
	client   *http.Client
}

// NewAnkrProvider accepts mainKeys and testKeys for signature consistency with other providers.
// Ankr does not support testnet chains, so testKeys is unused.
func NewAnkrProvider(mainKeys, _ []string) *AnkrProvider {
	return &AnkrProvider{mainPool: newKeyPool(mainKeys), client: &http.Client{Timeout: 15 * time.Second}}
}

func (p *AnkrProvider) Name() string { return "ankr" }

func (p *AnkrProvider) url() string {
	key := p.mainPool.pick()
	if key != "" {
		return "https://rpc.ankr.com/multichain/" + key
	}
	return "https://rpc.ankr.com/multichain"
}

func (p *AnkrProvider) GetTransfers(ctx context.Context, req TransferRequest) ([]TxRecord, error) {
	blockchain, ok := ankrBlockchains[req.Chain]
	if !ok {
		return nil, fmt.Errorf("ankr %w: %s", ErrUnsupportedChain, req.Chain)
	}
	if req.ContractAddress != "" {
		return p.getTokenTransfers(ctx, blockchain, req)
	}
	return p.getNativeTransfers(ctx, blockchain, req)
}

func (p *AnkrProvider) getTokenTransfers(ctx context.Context, blockchain string, req TransferRequest) ([]TxRecord, error) {
	params := map[string]any{
		"blockchain": blockchain, "address": req.Address,
		"contractAddress": req.ContractAddress,
		"pageSize":        req.Limit, "descOrder": true,
	}
	body, _ := json.Marshal(map[string]any{
		"jsonrpc": "2.0", "method": "ankr_getTokenTransfers", "params": params, "id": 1,
	})
	var out struct {
		Result *struct {
			Transfers []ankrTokenTransfer `json:"transfers"`
		} `json:"result"`
		Error *struct{ Message string `json:"message"` } `json:"error"`
	}
	if err := p.call(ctx, body, &out); err != nil {
		return nil, err
	}
	if out.Error != nil {
		return nil, fmt.Errorf("ankr: %s", out.Error.Message)
	}
	if out.Result == nil {
		return nil, nil
	}
	records := make([]TxRecord, 0, len(out.Result.Transfers))
	for _, t := range out.Result.Transfers {
		records = append(records, t.toRecord(req.Chain))
	}
	return records, nil
}

func (p *AnkrProvider) getNativeTransfers(ctx context.Context, blockchain string, req TransferRequest) ([]TxRecord, error) {
	params := map[string]any{
		"blockchain": blockchain, "address": req.Address,
		"pageSize": req.Limit, "descOrder": true,
	}
	body, _ := json.Marshal(map[string]any{
		"jsonrpc": "2.0", "method": "ankr_getTransactionsByAddress", "params": params, "id": 1,
	})
	var out struct {
		Result *struct {
			Transactions []ankrNativeTx `json:"transactions"`
		} `json:"result"`
		Error *struct{ Message string `json:"message"` } `json:"error"`
	}
	if err := p.call(ctx, body, &out); err != nil {
		return nil, err
	}
	if out.Error != nil {
		return nil, fmt.Errorf("ankr: %s", out.Error.Message)
	}
	if out.Result == nil {
		return nil, nil
	}
	records := make([]TxRecord, 0, len(out.Result.Transactions))
	for _, t := range out.Result.Transactions {
		records = append(records, t.toRecord(req.Chain))
	}
	return records, nil
}

func (p *AnkrProvider) call(ctx context.Context, body []byte, out any) error {
	req, _ := http.NewRequestWithContext(ctx, "POST", p.url(), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := p.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return json.NewDecoder(resp.Body).Decode(out)
}

type ankrTokenTransfer struct {
	TransactionHash string `json:"transactionHash"`
	FromAddress     string `json:"fromAddress"`
	ToAddress       string `json:"toAddress"`
	Value           string `json:"value"`
	ContractAddress string `json:"contractAddress"`
	TokenSymbol     string `json:"tokenSymbol"`
	TokenDecimals   int    `json:"tokenDecimals"`
	BlockHeight     int64  `json:"blockHeight"`
	Timestamp       int64  `json:"timestamp"`
}

func (t ankrTokenTransfer) toRecord(chain string) TxRecord {
	sym := t.TokenSymbol
	addr := t.ContractAddress
	return TxRecord{
		Hash: t.TransactionHash, From: t.FromAddress, To: t.ToAddress,
		Value: t.Value, Decimals: t.TokenDecimals,
		TokenSymbol: &sym, TokenContract: &addr,
		BlockNumber: t.BlockHeight, BlockTimestamp: time.Unix(t.Timestamp, 0).UTC(),
		ChainKey: chain, Source: "ankr",
	}
}

type ankrNativeTx struct {
	Hash      string `json:"hash"`
	From      string `json:"from"`
	To        string `json:"to"`
	Value     string `json:"value"`
	BlockNum  string `json:"blockNumber"`
	Timestamp int64  `json:"timestamp"`
}

func (t ankrNativeTx) toRecord(chain string) TxRecord {
	value := "0"
	if t.Value != "" {
		n := new(big.Int)
		n.SetString(strings.TrimPrefix(t.Value, "0x"), 16)
		value = n.String()
	}
	blockNum := int64(0)
	if t.BlockNum != "" {
		n := new(big.Int)
		n.SetString(strings.TrimPrefix(t.BlockNum, "0x"), 16)
		blockNum = n.Int64()
	}
	return TxRecord{
		Hash: t.Hash, From: t.From, To: t.To, Value: value, Decimals: 18,
		BlockNumber: blockNum, BlockTimestamp: time.Unix(t.Timestamp, 0).UTC(),
		ChainKey: chain, Source: "ankr",
	}
}
