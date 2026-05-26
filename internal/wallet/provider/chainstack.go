package provider

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// Chainstack JSON-RPC: https://docs.chainstack.com
// Endpoint URLs are per-chain and configured by the user in the dashboard.
// ERC20 transfers via eth_getLogs only; native not supported.

type ChainstackProvider struct {
	endpoints map[string]string // chain → RPC URL
	client    *http.Client
}

func NewChainstackProvider(endpoints map[string]string) *ChainstackProvider {
	return &ChainstackProvider{endpoints: endpoints, client: &http.Client{Timeout: 15 * time.Second}}
}

func (p *ChainstackProvider) Name() string { return "chainstack" }

func (p *ChainstackProvider) GetTransfers(ctx context.Context, req TransferRequest) ([]TxRecord, error) {
	rpcURL, ok := p.endpoints[req.Chain]
	if !ok || rpcURL == "" {
		return nil, fmt.Errorf("chainstack %w: %s not configured", ErrUnsupportedChain, req.Chain)
	}
	if req.ContractAddress == "" {
		return nil, fmt.Errorf("chainstack %w: native transfers not supported", ErrUnsupportedChain)
	}
	records, err := fetchERC20ByLogs(ctx, p.client, rpcURL, req.ContractAddress, req.Address, req.Limit, "chainstack")
	if err != nil {
		return nil, err
	}
	for i := range records {
		records[i].ChainKey = req.Chain
	}
	return records, nil
}
