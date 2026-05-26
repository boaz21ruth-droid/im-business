package provider

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// Infura JSON-RPC: https://docs.infura.io/api
// Supports ETH, Polygon, Arbitrum, Optimism (ERC20 via eth_getLogs only; native not supported).
var infuraNetworks = map[string]string{
	"eth":         "mainnet",
	"polygon":     "polygon-mainnet",
	"arbitrum":    "arbitrum-mainnet",
	"optimism":    "optimism-mainnet",
	"eth_sepolia": "sepolia",
}

type InfuraProvider struct {
	mainPool *keyPool
	testPool *keyPool
	client   *http.Client
}

func NewInfuraProvider(mainKeys, testKeys []string) *InfuraProvider {
	m := newKeyPool(mainKeys)
	t := m
	if len(testKeys) > 0 {
		t = newKeyPool(testKeys)
	}
	return &InfuraProvider{mainPool: m, testPool: t, client: &http.Client{Timeout: 15 * time.Second}}
}

func (p *InfuraProvider) Name() string { return "infura" }

func (p *InfuraProvider) GetTransfers(ctx context.Context, req TransferRequest) ([]TxRecord, error) {
	network, ok := infuraNetworks[req.Chain]
	if !ok {
		return nil, fmt.Errorf("infura %w: %s", ErrUnsupportedChain, req.Chain)
	}
	if req.ContractAddress == "" {
		return nil, fmt.Errorf("infura %w: native transfers not supported", ErrUnsupportedChain)
	}
	pool := p.mainPool
	if isTestnet[req.Chain] {
		pool = p.testPool
	}
	key := pool.pick()
	if key == "" {
		return nil, fmt.Errorf("infura %w: api_key not configured", ErrUnsupportedChain)
	}
	rpcURL := fmt.Sprintf("https://%s.infura.io/v3/%s", network, key)
	records, err := fetchERC20ByLogs(ctx, p.client, rpcURL, req.ContractAddress, req.Address, req.Limit, "infura")
	if err != nil {
		return nil, err
	}
	for i := range records {
		records[i].ChainKey = req.Chain
	}
	return records, nil
}
