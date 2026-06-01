package intent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"time"
)

// CoW Protocol constants — same on every supported chain.
const (
	gpv2Settlement   = "0x9008D19f58AAbD9eD0D60971565AA8510560ab41" // EIP-712 verifyingContract
	gpv2VaultRelayer = "0xC92E8bdf79f0507f65a392b0ab4667716BFE0110" // ERC20 approval spender
	// zeroAppData is the "no app metadata" sentinel (a 32-byte zero hash).
	zeroAppData = "0x0000000000000000000000000000000000000000000000000000000000000000"
)

// cowChainSlugs maps our chain keys to CoW orderbook URL slugs. CoW supports a
// limited set; only these of ours.
var cowChainSlugs = map[string]string{
	"eth":      "mainnet",
	"arbitrum": "arbitrum_one",
}

var cowChainIDs = map[string]int{
	"eth":      1,
	"arbitrum": 42161,
}

// CowProvider integrates the CoW Protocol orderbook API.
type CowProvider struct {
	apiBase string
	client  *http.Client
}

func NewCowProvider(apiBase string) *CowProvider {
	if apiBase == "" {
		apiBase = "https://api.cow.fi"
	}
	return &CowProvider{apiBase: apiBase, client: &http.Client{Timeout: 12 * time.Second}}
}

func (p *CowProvider) Name() string { return "cow" }

func (p *CowProvider) SupportsChain(chainKey string) bool {
	_, ok := cowChainSlugs[chainKey]
	return ok
}

func (p *CowProvider) base(chainKey string) string {
	return fmt.Sprintf("%s/%s/api/v1", p.apiBase, cowChainSlugs[chainKey])
}

func (p *CowProvider) Quote(ctx context.Context, req Request) (*QuoteResult, error) {
	if !p.SupportsChain(req.ChainKey) {
		return nil, fmt.Errorf("cow %w: %s", ErrUnsupportedChain, req.ChainKey)
	}
	body, _ := json.Marshal(map[string]any{
		"sellToken":           req.SellToken,
		"buyToken":            req.BuyToken,
		"from":                req.From,
		"receiver":            req.From,
		"sellAmountBeforeFee": req.SellAmount.String(),
		"kind":                "sell",
		"partiallyFillable":   false,
		"sellTokenBalance":    "erc20",
		"buyTokenBalance":     "erc20",
		"signingScheme":       "eip712",
		"onchainOrder":        false,
		"appData":             zeroAppData,
	})
	var out struct {
		Quote struct {
			SellToken         string `json:"sellToken"`
			BuyToken          string `json:"buyToken"`
			Receiver          string `json:"receiver"`
			SellAmount        string `json:"sellAmount"`
			BuyAmount         string `json:"buyAmount"`
			ValidTo           int64  `json:"validTo"`
			AppData           string `json:"appData"`
			FeeAmount         string `json:"feeAmount"`
			Kind              string `json:"kind"`
			PartiallyFillable bool   `json:"partiallyFillable"`
			SellTokenBalance  string `json:"sellTokenBalance"`
			BuyTokenBalance   string `json:"buyTokenBalance"`
		} `json:"quote"`
		ID int64 `json:"id"`
	}
	if err := p.do(ctx, http.MethodPost, p.base(req.ChainKey)+"/quote", body, &out); err != nil {
		return nil, err
	}
	if out.Quote.SellAmount == "" || out.Quote.BuyAmount == "" {
		return nil, ErrNoQuote
	}

	// Modern CoW: the SIGNED order carries feeAmount="0" and the FULL sell amount
	// (quote.sellAmount + quote.feeAmount); the fee is folded into the limit price.
	sellAmt, _ := new(big.Int).SetString(out.Quote.SellAmount, 10)
	fee, _ := new(big.Int).SetString(out.Quote.FeeAmount, 10)
	if sellAmt == nil || fee == nil {
		return nil, ErrNoQuote
	}
	fullSell := new(big.Int).Add(sellAmt, fee)

	// buyAmount is the slippage floor on the quoted estimate.
	buyEst, _ := new(big.Int).SetString(out.Quote.BuyAmount, 10)
	if buyEst == nil {
		return nil, ErrNoQuote
	}
	minBuy := new(big.Int).Mul(buyEst, big.NewInt(int64(10000-req.SlippageBps)))
	minBuy.Div(minBuy, big.NewInt(10000))

	appData := out.Quote.AppData
	if appData == "" {
		appData = zeroAppData
	}

	return &QuoteResult{
		Order: Order{
			SellToken:         out.Quote.SellToken,
			BuyToken:          out.Quote.BuyToken,
			Receiver:          out.Quote.Receiver,
			SellAmount:        fullSell.String(),
			BuyAmount:         minBuy.String(),
			ValidTo:           out.Quote.ValidTo,
			AppData:           appData,
			FeeAmount:         "0",
			Kind:              "sell",
			PartiallyFillable: false,
			SellTokenBalance:  "erc20",
			BuyTokenBalance:   "erc20",
		},
		ChainID:           cowChainIDs[req.ChainKey],
		VerifyingContract: gpv2Settlement,
		ApprovalSpender:   gpv2VaultRelayer,
		QuoteID:           out.ID,
		ExpectedBuyAmount: out.Quote.BuyAmount,
	}, nil
}

func (p *CowProvider) Submit(ctx context.Context, chainKey string, sub SubmitRequest) (string, error) {
	if !p.SupportsChain(chainKey) {
		return "", fmt.Errorf("cow %w: %s", ErrUnsupportedChain, chainKey)
	}
	o := sub.Order
	body, _ := json.Marshal(map[string]any{
		"sellToken":         o.SellToken,
		"buyToken":          o.BuyToken,
		"receiver":          o.Receiver,
		"sellAmount":        o.SellAmount,
		"buyAmount":         o.BuyAmount,
		"validTo":           o.ValidTo,
		"appData":           o.AppData,
		"feeAmount":         o.FeeAmount,
		"kind":              o.Kind,
		"partiallyFillable": o.PartiallyFillable,
		"sellTokenBalance":  o.SellTokenBalance,
		"buyTokenBalance":   o.BuyTokenBalance,
		"signingScheme":     "eip712",
		"signature":         sub.Signature,
		"from":              sub.From,
		"quoteId":           sub.QuoteID,
	})
	// CoW returns the order UID as a bare JSON string.
	var uid string
	if err := p.do(ctx, http.MethodPost, p.base(chainKey)+"/orders", body, &uid); err != nil {
		return "", err
	}
	return uid, nil
}

func (p *CowProvider) Status(ctx context.Context, chainKey, orderUID string) (*OrderStatus, error) {
	if !p.SupportsChain(chainKey) {
		return nil, fmt.Errorf("cow %w: %s", ErrUnsupportedChain, chainKey)
	}
	var out struct {
		Status            string `json:"status"`
		ExecutedBuyAmount string `json:"executedBuyAmount"`
	}
	url := p.base(chainKey) + "/orders/" + orderUID
	if err := p.do(ctx, http.MethodGet, url, nil, &out); err != nil {
		return nil, err
	}
	return &OrderStatus{Status: out.Status, ExecutedBuyAmount: out.ExecutedBuyAmount}, nil
}



func (p *CowProvider) do(ctx context.Context, method, url string, body []byte, out any) error {
	var rdr io.Reader
	if body != nil {
		rdr = bytes.NewReader(body)
	}
	req, _ := http.NewRequestWithContext(ctx, method, url, rdr)
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := p.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("cow %s: status %d: %s", method, resp.StatusCode, string(respBody))
	}
	if out == nil {
		return nil
	}
	return json.Unmarshal(respBody, out)
}
