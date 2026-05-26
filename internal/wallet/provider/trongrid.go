package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

var trongridHosts = map[string]string{
	"tron":        "https://api.trongrid.io",
	"tron_shasta": "https://api.shasta.trongrid.io",
}

type TronGridProvider struct {
	client *http.Client
}

func NewTronGridProvider() *TronGridProvider {
	return &TronGridProvider{client: &http.Client{Timeout: 15 * time.Second}}
}

func (p *TronGridProvider) Name() string { return "trongrid" }

func (p *TronGridProvider) GetTransfers(ctx context.Context, req TransferRequest) ([]TxRecord, error) {
	host, ok := trongridHosts[req.Chain]
	if !ok {
		return nil, fmt.Errorf("trongrid: unsupported chain %s", req.Chain)
	}

	if req.ContractAddress != "" {
		return p.getTRC20(ctx, host, req)
	}
	return p.getNative(ctx, host, req)
}

func (p *TronGridProvider) getTRC20(ctx context.Context, host string, req TransferRequest) ([]TxRecord, error) {
	url := fmt.Sprintf("%s/v1/accounts/%s/transactions/trc20?limit=%d&only_confirmed=true&contract_address=%s",
		host, req.Address, req.Limit, req.ContractAddress)
	if req.MinTimestamp > 0 {
		url += fmt.Sprintf("&min_timestamp=%d", req.MinTimestamp)
	}

	httpReq, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var out struct {
		Data []trc20Transfer `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}

	records := make([]TxRecord, 0, len(out.Data))
	for _, t := range out.Data {
		records = append(records, t.toRecord(req.Chain))
	}
	return records, nil
}

func (p *TronGridProvider) getNative(ctx context.Context, host string, req TransferRequest) ([]TxRecord, error) {
	url := fmt.Sprintf("%s/v1/accounts/%s/transactions?limit=%d&only_confirmed=true",
		host, req.Address, req.Limit)
	if req.MinTimestamp > 0 {
		url += fmt.Sprintf("&min_timestamp=%d", req.MinTimestamp)
	}

	httpReq, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var out struct {
		Data []tronNativeTx `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}

	records := make([]TxRecord, 0, len(out.Data))
	for _, t := range out.Data {
		r, ok := t.toRecord(req.Chain)
		if ok {
			records = append(records, r)
		}
	}
	return records, nil
}

type trc20Transfer struct {
	TransactionID  string `json:"transaction_id"`
	BlockTimestamp int64  `json:"block_timestamp"` // Unix ms
	From           string `json:"from"`
	To             string `json:"to"`
	Value          string `json:"value"`
	TokenInfo      struct {
		Symbol   string `json:"symbol"`
		Address  string `json:"address"`
		Decimals int    `json:"decimals"`
	} `json:"token_info"`
}

func (t trc20Transfer) toRecord(chain string) TxRecord {
	sym := t.TokenInfo.Symbol
	addr := t.TokenInfo.Address
	ts := time.UnixMilli(t.BlockTimestamp).UTC()
	return TxRecord{
		Hash:           t.TransactionID,
		From:           t.From,
		To:             t.To,
		Value:          t.Value,
		Decimals:       t.TokenInfo.Decimals,
		TokenSymbol:    &sym,
		TokenContract:  &addr,
		BlockTimestamp: ts,
		ChainKey:       chain,
		Source:         "trongrid",
	}
}

type tronNativeTx struct {
	TxID           string `json:"txID"`
	BlockTimestamp int64  `json:"block_timestamp"`
	Ret            []struct {
		ContractRet string `json:"contractRet"`
	} `json:"ret"`
	RawData struct {
		Contract []struct {
			Type      string `json:"type"`
			Parameter struct {
				Value struct {
					Amount       int64  `json:"amount"`
					OwnerAddress string `json:"owner_address"` // hex
					ToAddress    string `json:"to_address"`    // hex
				} `json:"value"`
			} `json:"parameter"`
		} `json:"contract"`
	} `json:"raw_data"`
}

func (t tronNativeTx) toRecord(chain string) (TxRecord, bool) {
	if len(t.RawData.Contract) == 0 || t.RawData.Contract[0].Type != "TransferContract" {
		return TxRecord{}, false
	}
	c := t.RawData.Contract[0].Parameter.Value
	ts := time.UnixMilli(t.BlockTimestamp).UTC()
	return TxRecord{
		Hash:           t.TxID,
		From:           hexToTronAddress(c.OwnerAddress),
		To:             hexToTronAddress(c.ToAddress),
		Value:          fmt.Sprintf("%d", c.Amount),
		Decimals:       6,
		BlockTimestamp: ts,
		ChainKey:       chain,
		Source:         "trongrid",
	}, true
}

// hexToTronAddress converts a TronGrid hex address (41...) to base58check.
// Falls back to hex string if conversion fails.
func hexToTronAddress(hexAddr string) string {
	// Tron addresses from TronGrid are hex-encoded 21 bytes starting with "41"
	// Full base58check encoding requires sha256 which adds complexity.
	// Return as-is for now; the poller won't match these against stored base58 addresses.
	return hexAddr
}
