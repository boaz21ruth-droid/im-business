package provider

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// GetBlock JSON-RPC: https://getblock.io/docs/
// Supports ETH, BSC, Polygon (ERC20 via eth_getLogs only; native not supported).
var getblockNetworks = map[string]string{
	"eth":     "eth",
	"bsc":     "bsc",
	"polygon": "matic",
}

type GetBlockProvider struct {
	mainPool *keyPool
	client   *http.Client
}

// NewGetBlockProvider accepts mainKeys and testKeys for signature consistency.
// GetBlock only supports mainnet chains so testKeys is unused.
func NewGetBlockProvider(mainKeys, _ []string) *GetBlockProvider {
	return &GetBlockProvider{mainPool: newKeyPool(mainKeys), client: &http.Client{Timeout: 15 * time.Second}}
}

func (p *GetBlockProvider) Name() string { return "getblock" }

func (p *GetBlockProvider) GetTransfers(ctx context.Context, req TransferRequest) ([]TxRecord, error) {
	network, ok := getblockNetworks[req.Chain]
	if !ok {
		return nil, fmt.Errorf("getblock %w: %s", ErrUnsupportedChain, req.Chain)
	}
	if req.ContractAddress == "" {
		return nil, fmt.Errorf("getblock %w: native transfers not supported", ErrUnsupportedChain)
	}
	key := p.mainPool.pick()
	if key == "" {
		return nil, fmt.Errorf("getblock %w: api_key not configured", ErrUnsupportedChain)
	}
	rpcURL := fmt.Sprintf("https://%s.getblock.io/%s/mainnet/", network, key)
	records, err := fetchERC20ByLogs(ctx, p.client, rpcURL, req.ContractAddress, req.Address, req.Limit, "getblock")
	if err != nil {
		return nil, err
	}
	for i := range records {
		records[i].ChainKey = req.Chain
	}
	return records, nil
}
