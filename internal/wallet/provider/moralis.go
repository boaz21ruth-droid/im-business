package provider

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"golang.org/x/crypto/sha3"
)


var moralisChainIDs = map[string]string{
	"eth":         "0x1",
	"bsc":         "0x38",
	"polygon":     "0x89",
	"arbitrum":    "0xa4b1",
	"optimism":    "0xa",
	"bsc_testnet": "0x61",
	"eth_sepolia": "0xaa36a7",
}

type MoralisProvider struct {
	apiKey        string
	webhookSecret string
	webhookURL    string
	client        *http.Client
}

func NewMoralisProvider(apiKey, webhookSecret, webhookURL string) *MoralisProvider {
	return &MoralisProvider{
		apiKey:        apiKey,
		webhookSecret: webhookSecret,
		webhookURL:    webhookURL,
		client:        &http.Client{Timeout: 15 * time.Second},
	}
}

func (p *MoralisProvider) Name() string { return "moralis" }

func (p *MoralisProvider) GetTransfers(ctx context.Context, req TransferRequest) ([]TxRecord, error) {
	chainID, ok := moralisChainIDs[req.Chain]
	if !ok {
		return nil, fmt.Errorf("moralis %w: %s", ErrUnsupportedChain, req.Chain)
	}
	if req.ContractAddress == "" {
		return p.getNativeTransfers(ctx, chainID, req)
	}
	return p.getERC20Transfers(ctx, chainID, req)
}

func (p *MoralisProvider) getERC20Transfers(ctx context.Context, chainID string, req TransferRequest) ([]TxRecord, error) {
	url := fmt.Sprintf("https://deep-index.moralis.io/api/v2.2/%s/erc20/transfers?chain=%s&limit=%d&order=DESC",
		req.Address, chainID, req.Limit)
	url += "&contract_addresses[0]=" + req.ContractAddress

	httpReq, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
	httpReq.Header.Set("X-API-Key", p.apiKey)

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var out struct {
		Result []moralisTransfer `json:"result"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	records := make([]TxRecord, 0, len(out.Result))
	for _, t := range out.Result {
		records = append(records, t.toRecord(req.Chain))
	}
	return records, nil
}

// getNativeTransfers fetches native coin transfers (ETH/BNB/MATIC etc.) via /transactions.
func (p *MoralisProvider) getNativeTransfers(ctx context.Context, chainID string, req TransferRequest) ([]TxRecord, error) {
	url := fmt.Sprintf("https://deep-index.moralis.io/api/v2.2/%s/transactions?chain=%s&limit=%d&order=DESC",
		req.Address, chainID, req.Limit)

	httpReq, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
	httpReq.Header.Set("X-API-Key", p.apiKey)

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var out struct {
		Result []moralisNativeTx `json:"result"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	records := make([]TxRecord, 0, len(out.Result))
	for _, t := range out.Result {
		records = append(records, t.toRecord(req.Chain))
	}
	return records, nil
}

// erc20TransferABI is the minimal ABI for the ERC20 Transfer event.
// Moralis requires explicit event definitions instead of "all contract events".
var erc20TransferABI = []map[string]any{
	{
		"anonymous": false,
		"inputs": []map[string]any{
			{"indexed": true, "name": "from", "type": "address"},
			{"indexed": true, "name": "to", "type": "address"},
			{"indexed": false, "name": "value", "type": "uint256"},
		},
		"name": "Transfer",
		"type": "event",
	},
}

// EnsureStream creates (or looks up existing) Moralis stream for a chain.
func (p *MoralisProvider) EnsureStream(ctx context.Context, chainKey string) (streamID string, err error) {
	chainID, ok := moralisChainIDs[chainKey]
	if !ok {
		return "", fmt.Errorf("moralis: unsupported chain %s", chainKey)
	}

	body, _ := json.Marshal(map[string]any{
		"webhookUrl":          p.webhookURL,
		"description":         "IM Wallet",
		"tag":                 "im-wallet-" + chainKey,
		"chains":              []string{chainID},
		"includeContractLogs": true,
		"includeNativeTxs":    false,
		"includeInternalTxs":  false,
		// Specify the Transfer event so Moralis knows it's not "all events"
		"abi":    erc20TransferABI,
		"topic0": []string{"Transfer(address,address,uint256)"},
	})

	req, _ := http.NewRequestWithContext(ctx, "POST", "https://api.moralis.io/streams/evm", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", p.apiKey)

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
		return "", fmt.Errorf("moralis stream create failed: %s", out.Error)
	}
	return out.ID, nil
}

// AddAddressToStream adds a wallet address to an existing Moralis stream.
func (p *MoralisProvider) AddAddressToStream(ctx context.Context, streamID, address string) error {
	body, _ := json.Marshal(map[string]any{"address": address})
	url := fmt.Sprintf("https://api.moralis.io/streams/evm/%s/address", streamID)

	req, _ := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", p.apiKey)

	resp, err := p.client.Do(req)
	if err != nil {
		return err
	}
	resp.Body.Close()
	if resp.StatusCode >= 400 {
		return fmt.Errorf("moralis add address: HTTP %d", resp.StatusCode)
	}
	return nil
}

// VerifyWebhook checks the x-signature header against keccak256(body + secret).
func (p *MoralisProvider) VerifyWebhook(r *http.Request, body []byte) bool {
	signature := r.Header.Get("x-signature")
	h := sha3.NewLegacyKeccak256()
	h.Write(body)
	h.Write([]byte(p.webhookSecret))
	computed := hex.EncodeToString(h.Sum(nil))
	sig := strings.TrimPrefix(strings.ToLower(signature), "0x")
	return computed == sig
}

// moralisChainIDToKey maps Moralis hex chain IDs back to our chain keys.
func moralisChainIDToKey(chainID string) string {
	for key, id := range moralisChainIDs {
		if id == strings.ToLower(chainID) {
			return key
		}
	}
	return ""
}

// ParseWebhookPayload parses a Moralis webhook body into a normalized transfer list.
func (p *MoralisProvider) ParseWebhookPayload(body []byte) (chainKey string, confirmed bool, transfers []WebhookTransfer, err error) {
	var payload MoralisWebhookPayload
	if err = json.Unmarshal(body, &payload); err != nil {
		return
	}
	chainKey = moralisChainIDToKey(payload.ChainID)
	confirmed = payload.Confirmed
	for _, t := range payload.ERC20Transfers {
		decimals := 18
		fmt.Sscanf(t.TokenDecimals, "%d", &decimals)
		transfers = append(transfers, WebhookTransfer{
			TxHash:       t.TransactionHash,
			From:         t.From,
			To:           t.To,
			Value:        t.Value,
			Decimals:     decimals,
			TokenSymbol:  t.TokenSymbol,
			TokenAddress: t.TokenAddress,
		})
	}
	return
}

type moralisTransfer struct {
	TransactionHash string `json:"transaction_hash"`
	FromAddress     string `json:"from_address"`
	ToAddress       string `json:"to_address"`
	Value           string `json:"value"`
	TokenSymbol     string `json:"token_symbol"`
	TokenDecimals   string `json:"token_decimals"` // string in API response
	ContractAddress string `json:"address"`
	BlockTimestamp  string `json:"block_timestamp"` // ISO 8601
	BlockNumber     string `json:"block_number"`
}

func (t moralisTransfer) toRecord(chain string) TxRecord {
	sym := t.TokenSymbol
	addr := t.ContractAddress
	ts, _ := time.Parse(time.RFC3339, t.BlockTimestamp)
	decimals := 18
	fmt.Sscanf(t.TokenDecimals, "%d", &decimals)
	blockNum := int64(0)
	fmt.Sscanf(t.BlockNumber, "%d", &blockNum)
	return TxRecord{
		Hash:           t.TransactionHash,
		From:           t.FromAddress,
		To:             t.ToAddress,
		Value:          t.Value,
		Decimals:       decimals,
		TokenSymbol:    &sym,
		TokenContract:  &addr,
		BlockNumber:    blockNum,
		BlockTimestamp: ts,
		ChainKey:       chain,
		Source:         "moralis",
	}
}

type moralisNativeTx struct {
	Hash           string `json:"hash"`
	FromAddress    string `json:"from_address"`
	ToAddress      string `json:"to_address"`
	Value          string `json:"value"` // wei, decimal string
	BlockTimestamp string `json:"block_timestamp"`
	BlockNumber    string `json:"block_number"`
}

func (t moralisNativeTx) toRecord(chain string) TxRecord {
	ts, _ := time.Parse(time.RFC3339, t.BlockTimestamp)
	blockNum := int64(0)
	fmt.Sscanf(t.BlockNumber, "%d", &blockNum)
	return TxRecord{
		Hash:           t.Hash,
		From:           t.FromAddress,
		To:             t.ToAddress,
		Value:          t.Value,
		Decimals:       18,
		BlockNumber:    blockNum,
		BlockTimestamp: ts,
		ChainKey:       chain,
		Source:         "moralis",
	}
}

// MoralisWebhookPayload is the structure of Moralis webhook POST body.
type MoralisWebhookPayload struct {
	ChainID        string                    `json:"chainId"`
	Confirmed      bool                      `json:"confirmed"`
	StreamID       string                    `json:"streamId"`
	Tag            string                    `json:"tag"`
	ERC20Transfers []MoralisWebhookTransfer  `json:"erc20Transfers"`
}

type MoralisWebhookTransfer struct {
	TransactionHash string `json:"transactionHash"`
	From            string `json:"from"`
	To              string `json:"to"`
	Value           string `json:"value"`
	TokenSymbol     string `json:"tokenSymbol"`
	TokenDecimals   string `json:"tokenDecimals"`
	TokenAddress    string `json:"tokenAddress"`
}
