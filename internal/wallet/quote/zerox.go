package quote

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// zeroExChainIDs maps our chain keys to 0x `chainId` values. Mirrors the map
// that used to live in the Flutter ZeroExProvider (now deleted).
var zeroExChainIDs = map[string]int{
	"eth":      1,
	"bsc":      56,
	"polygon":  137,
	"arbitrum": 42161,
	"optimism": 10,
}

// ZeroExProvider integrates 0x Swap API v2 (allowance-holder endpoints).
type ZeroExProvider struct {
	apiKey        string
	apiBase       string
	version       string
	feeRecipients map[string]string // chainKey -> platform fee recipient ("" = no fee)
	feeBps        int
	client        *http.Client
}

// NewZeroExProvider constructs a 0x provider. apiBase/version fall back to the
// 0x v2 defaults when empty.
func NewZeroExProvider(apiKey, apiBase, version string, feeRecipients map[string]string, feeBps int) *ZeroExProvider {
	if apiBase == "" {
		apiBase = "https://api.0x.org"
	}
	if version == "" {
		version = "v2"
	}
	return &ZeroExProvider{
		apiKey:        apiKey,
		apiBase:       apiBase,
		version:       version,
		feeRecipients: feeRecipients,
		feeBps:        feeBps,
		client:        &http.Client{Timeout: 10 * time.Second},
	}
}

func (p *ZeroExProvider) Name() string { return "zerox" }

func (p *ZeroExProvider) SupportsChain(chainKey string) bool {
	_, ok := zeroExChainIDs[chainKey]
	return ok && p.apiKey != ""
}

func (p *ZeroExProvider) params(req Request) url.Values {
	v := url.Values{}
	v.Set("chainId", strconv.Itoa(zeroExChainIDs[req.ChainKey]))
	v.Set("sellToken", req.SellToken.APIAddress())
	v.Set("buyToken", req.BuyToken.APIAddress())
	v.Set("sellAmount", req.SellAmount.String())
	if req.Taker != "" {
		v.Set("taker", req.Taker)
	}
	v.Set("slippageBps", strconv.Itoa(req.SlippageBps))
	if recipient := p.feeRecipients[req.ChainKey]; recipient != "" && p.feeBps > 0 {
		v.Set("swapFeeBps", strconv.Itoa(p.feeBps))
		v.Set("swapFeeRecipient", recipient)
		v.Set("swapFeeToken", req.BuyToken.APIAddress())
	}
	return v
}

func (p *ZeroExProvider) headers(r *http.Request) {
	r.Header.Set("0x-api-key", p.apiKey)
	r.Header.Set("0x-version", p.version)
}

func (p *ZeroExProvider) do(ctx context.Context, path string, req Request, out any) error {
	if !p.SupportsChain(req.ChainKey) {
		return fmt.Errorf("zerox %w: %s", ErrUnsupportedChain, req.ChainKey)
	}
	u := fmt.Sprintf("%s/swap/allowance-holder/%s?%s", p.apiBase, path, p.params(req).Encode())
	httpReq, _ := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	p.headers(httpReq)

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("zerox %s: status %d: %s", path, resp.StatusCode, string(body))
	}
	return json.Unmarshal(body, out)
}

type zeroExFeeItem struct {
	Amount string `json:"amount"`
	Token  string `json:"token"`
}

type zeroExFees struct {
	IntegratorFee *zeroExFeeItem `json:"integratorFee"`
	ZeroExFee     *zeroExFeeItem `json:"zeroExFee"`
	GasFee        *zeroExFeeItem `json:"gasFee"`
}

func (f zeroExFees) toFees() Fees {
	out := Fees{}
	if f.IntegratorFee != nil {
		out.IntegratorFeeAmount = f.IntegratorFee.Amount
		out.IntegratorFeeToken = f.IntegratorFee.Token
	}
	if f.ZeroExFee != nil {
		out.ZeroExFeeAmount = f.ZeroExFee.Amount
	}
	if f.GasFee != nil {
		out.GasFeeAmount = f.GasFee.Amount
	}
	return out
}

func (p *ZeroExProvider) Price(ctx context.Context, req Request) (*PriceResult, error) {
	var out struct {
		BuyAmount string     `json:"buyAmount"`
		Gas       string     `json:"gas"`
		Fees      zeroExFees `json:"fees"`
	}
	if err := p.do(ctx, "price", req, &out); err != nil {
		return nil, err
	}
	if out.BuyAmount == "" {
		return nil, fmt.Errorf("zerox price: empty buyAmount")
	}
	return &PriceResult{
		Provider:    p.Name(),
		BuyAmount:   out.BuyAmount,
		GasEstimate: out.Gas,
		Fees:        out.Fees.toFees(),
	}, nil
}

func (p *ZeroExProvider) Quote(ctx context.Context, req Request) (*Quote, error) {
	var out struct {
		BuyAmount   string `json:"buyAmount"`
		Transaction struct {
			To       string `json:"to"`
			Data     string `json:"data"`
			Value    string `json:"value"`
			Gas      string `json:"gas"`
			GasPrice string `json:"gasPrice"`
		} `json:"transaction"`
		Fees   zeroExFees `json:"fees"`
		Issues *struct {
			Allowance *struct {
				Spender string `json:"spender"`
			} `json:"allowance"`
		} `json:"issues"`
	}
	if err := p.do(ctx, "quote", req, &out); err != nil {
		return nil, err
	}
	if out.BuyAmount == "" || out.Transaction.To == "" {
		return nil, fmt.Errorf("zerox quote: incomplete response")
	}

	buyAmount, ok := new(big.Int).SetString(out.BuyAmount, 10)
	if !ok {
		return nil, fmt.Errorf("zerox quote: bad buyAmount %q", out.BuyAmount)
	}
	// minBuyAmount = buyAmount * (10000 - slippageBps) / 10000
	minBuy := new(big.Int).Mul(buyAmount, big.NewInt(int64(10000-req.SlippageBps)))
	minBuy.Div(minBuy, big.NewInt(10000))

	value := out.Transaction.Value
	if value == "" {
		value = "0"
	}

	var approval *Approval
	if out.Issues != nil && out.Issues.Allowance != nil && !req.SellToken.IsNative() {
		approval = &Approval{
			TokenAddress:   req.SellToken.ContractAddress,
			Spender:        out.Issues.Allowance.Spender,
			RequiredAmount: req.SellAmount.String(),
		}
	}

	return &Quote{
		Provider:     p.Name(),
		BuyAmount:    out.BuyAmount,
		MinBuyAmount: minBuy.String(),
		To:           out.Transaction.To,
		Data:         out.Transaction.Data,
		Value:        value,
		Gas:          out.Transaction.Gas,
		GasPrice:     out.Transaction.GasPrice,
		Approval:     approval,
		Fees:         out.Fees.toFees(),
	}, nil
}
