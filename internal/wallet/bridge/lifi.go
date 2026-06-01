package bridge

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// lifiChainIDs maps our chain keys to numeric chain IDs LI.FI uses.
var lifiChainIDs = map[string]int{
	"eth":      1,
	"bsc":      56,
	"polygon":  137,
	"arbitrum": 42161,
	"optimism": 10,
}

// LiFiProvider integrates the LI.FI cross-chain API (li.quest/v1).
type LiFiProvider struct {
	apiBase    string
	apiKey     string
	integrator string
	feeBps     int
	client     *http.Client
}

func NewLiFiProvider(apiBase, apiKey, integrator string, feeBps int) *LiFiProvider {
	if apiBase == "" {
		apiBase = "https://li.quest/v1"
	}
	if integrator == "" {
		integrator = "im-wallet"
	}
	return &LiFiProvider{
		apiBase:    apiBase,
		apiKey:     apiKey,
		integrator: integrator,
		feeBps:     feeBps,
		client:     &http.Client{Timeout: 15 * time.Second},
	}
}

func (p *LiFiProvider) Name() string { return "lifi" }

func (p *LiFiProvider) SupportsChain(chainKey string) bool {
	_, ok := lifiChainIDs[chainKey]
	return ok
}

// tokenAddr maps "" / "native" to LI.FI's native (zero) address.
func tokenAddr(v string) string {
	if v == "" || v == "native" {
		return NativeToken
	}
	return v
}

func (p *LiFiProvider) headers(r *http.Request) {
	r.Header.Set("Accept", "application/json")
	if p.apiKey != "" {
		r.Header.Set("x-lifi-api-key", p.apiKey)
	}
}

func (p *LiFiProvider) get(ctx context.Context, path string, q url.Values, out any) (int, error) {
	u := fmt.Sprintf("%s/%s?%s", p.apiBase, path, q.Encode())
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	p.headers(req)
	resp, err := p.client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return resp.StatusCode, fmt.Errorf("lifi %s: status %d: %s", path, resp.StatusCode, string(body))
	}
	return resp.StatusCode, json.Unmarshal(body, out)
}

func (p *LiFiProvider) Quote(ctx context.Context, req Request) (*Quote, error) {
	if !p.SupportsChain(req.FromChain) || !p.SupportsChain(req.ToChain) {
		return nil, fmt.Errorf("lifi %w: %s->%s", ErrUnsupportedChain, req.FromChain, req.ToChain)
	}
	to := req.ToAddress
	if to == "" {
		to = req.FromAddress
	}
	q := url.Values{}
	q.Set("fromChain", strconv.Itoa(lifiChainIDs[req.FromChain]))
	q.Set("toChain", strconv.Itoa(lifiChainIDs[req.ToChain]))
	q.Set("fromToken", tokenAddr(req.FromToken))
	q.Set("toToken", tokenAddr(req.ToToken))
	q.Set("fromAmount", req.FromAmount.String())
	q.Set("fromAddress", req.FromAddress)
	q.Set("toAddress", to)
	q.Set("slippage", strconv.FormatFloat(float64(req.SlippageBps)/10000.0, 'f', -1, 64))
	q.Set("integrator", p.integrator)
	if p.feeBps > 0 {
		q.Set("fee", strconv.FormatFloat(float64(p.feeBps)/10000.0, 'f', -1, 64))
	}

	var out struct {
		Tool     string `json:"tool"`
		Estimate struct {
			ApprovalAddress   string `json:"approvalAddress"`
			ToAmount          string `json:"toAmount"`
			ToAmountMin       string `json:"toAmountMin"`
			ExecutionDuration int    `json:"executionDuration"`
		} `json:"estimate"`
		TransactionRequest struct {
			To       string `json:"to"`
			Data     string `json:"data"`
			Value    string `json:"value"`    // hex
			GasPrice string `json:"gasPrice"` // hex
			GasLimit string `json:"gasLimit"` // hex
		} `json:"transactionRequest"`
	}
	status, err := p.get(ctx, "quote", q, &out)
	if err != nil {
		if status == http.StatusNotFound {
			return nil, ErrNoRoute
		}
		return nil, err
	}
	if out.TransactionRequest.To == "" || out.Estimate.ToAmount == "" {
		return nil, ErrNoRoute
	}

	var approval *Approval
	if tokenAddr(req.FromToken) != NativeToken && out.Estimate.ApprovalAddress != "" {
		approval = &Approval{
			TokenAddress:   req.FromToken,
			Spender:        out.Estimate.ApprovalAddress,
			RequiredAmount: req.FromAmount.String(),
		}
	}

	return &Quote{
		Tool:                 out.Tool,
		FromChain:            req.FromChain,
		ToChain:              req.ToChain,
		ToAmount:             out.Estimate.ToAmount,
		ToAmountMin:          out.Estimate.ToAmountMin,
		To:                   out.TransactionRequest.To,
		Data:                 out.TransactionRequest.Data,
		Value:                hexToDec(out.TransactionRequest.Value),
		Gas:                  hexToDec(out.TransactionRequest.GasLimit),
		GasPrice:             hexToDec(out.TransactionRequest.GasPrice),
		Approval:             approval,
		ExecutionDurationSec: out.Estimate.ExecutionDuration,
	}, nil
}

func (p *LiFiProvider) Status(ctx context.Context, fromChain, toChain, txHash, tool string) (*Status, error) {
	q := url.Values{}
	q.Set("txHash", txHash)
	if id, ok := lifiChainIDs[fromChain]; ok {
		q.Set("fromChain", strconv.Itoa(id))
	}
	if id, ok := lifiChainIDs[toChain]; ok {
		q.Set("toChain", strconv.Itoa(id))
	}
	if tool != "" {
		q.Set("bridge", tool)
	}

	var out struct {
		Status            string `json:"status"`
		Substatus         string `json:"substatus"`
		SubstatusMessage  string `json:"substatusMessage"`
		LifiExplorerLink  string `json:"lifiExplorerLink"`
		Receiving         struct {
			TxHash string `json:"txHash"`
			Amount string `json:"amount"`
		} `json:"receiving"`
	}
	status, err := p.get(ctx, "status", q, &out)
	if err != nil {
		// Right after broadcast (or for a not-yet-indexed tx) LI.FI 404s. Treat
		// that as "not found yet" so the client keeps polling rather than erroring.
		if status == http.StatusNotFound {
			return &Status{Status: "NOT_FOUND"}, nil
		}
		return nil, err
	}
	return &Status{
		Status:     out.Status,
		Substatus:  out.Substatus,
		Message:    out.SubstatusMessage,
		DestTxHash: out.Receiving.TxHash,
		DestAmount: out.Receiving.Amount,
		Explorer:   out.LifiExplorerLink,
	}, nil
}

// hexToDec converts a 0x-prefixed hex string to a decimal string. Returns "" for
// empty/unparsable input. LI.FI returns value/gas as hex; the client expects decimal.
func hexToDec(h string) string {
	h = strings.TrimSpace(h)
	if h == "" || h == "0x" {
		return ""
	}
	n, ok := new(big.Int).SetString(strings.TrimPrefix(h, "0x"), 16)
	if !ok {
		return ""
	}
	return n.String()
}
