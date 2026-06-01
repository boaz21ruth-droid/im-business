package quote

import (
	"bytes"
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

// kyberChainSlugs maps our chain keys to KyberSwap's URL chain slugs.
var kyberChainSlugs = map[string]string{
	"eth":      "ethereum",
	"bsc":      "bsc",
	"polygon":  "polygon",
	"arbitrum": "arbitrum",
	"optimism": "optimism",
}

// KyberSwapProvider integrates the KyberSwap Aggregator API v1. Keyless: an
// optional x-client-id improves rate limits. Two calls: /routes (price) then
// /route/build (calldata).
// kyberAMMSources is an allow-list of on-chain AMM source IDs used in AMM-only
// (fork-test) mode, so off-chain order/PMM/RFQ sources are excluded by default.
const kyberAMMSources = "uniswap,uniswapv3,sushiswap,curve,balancer-v2,maker-psm"

type KyberSwapProvider struct {
	apiBase       string
	clientID      string
	feeRecipients map[string]string
	feeBps        int
	ammOnly       bool
	client        *http.Client
}

func NewKyberSwapProvider(apiBase, clientID string, feeRecipients map[string]string, feeBps int, ammOnly bool) *KyberSwapProvider {
	if apiBase == "" {
		apiBase = "https://aggregator-api.kyberswap.com"
	}
	if clientID == "" {
		clientID = "im-wallet"
	}
	return &KyberSwapProvider{
		apiBase:       apiBase,
		clientID:      clientID,
		feeRecipients: feeRecipients,
		feeBps:        feeBps,
		ammOnly:       ammOnly,
		client:        &http.Client{Timeout: 10 * time.Second},
	}
}

func (p *KyberSwapProvider) Name() string { return "kyberswap" }

func (p *KyberSwapProvider) SupportsChain(chainKey string) bool {
	_, ok := kyberChainSlugs[chainKey]
	return ok // keyless — no API key needed
}

// routeSummary is opaque to us; we pass it back verbatim to /route/build.
type kyberRoutes struct {
	Code int    `json:"code"`
	Msg  string `json:"message"`
	Data struct {
		RouteSummary  json.RawMessage `json:"routeSummary"`
		RouterAddress string          `json:"routerAddress"`
	} `json:"data"`
}

// routeSummaryOut pulls just amountOut/gas out of the opaque summary.
type kyberRouteSummaryFields struct {
	AmountOut string `json:"amountOut"`
	Gas       string `json:"gas"`
}

func (p *KyberSwapProvider) getRoutes(ctx context.Context, req Request) (*kyberRoutes, error) {
	slug := kyberChainSlugs[req.ChainKey]
	if slug == "" {
		return nil, fmt.Errorf("kyberswap %w: %s", ErrUnsupportedChain, req.ChainKey)
	}
	v := url.Values{}
	v.Set("tokenIn", req.SellToken.APIAddress())
	v.Set("tokenOut", req.BuyToken.APIAddress())
	v.Set("amountIn", req.SellAmount.String())
	v.Set("gasInclude", "true")
	if recipient := p.feeRecipients[req.ChainKey]; recipient != "" && p.feeBps > 0 {
		v.Set("feeAmount", strconv.Itoa(p.feeBps))
		v.Set("chargeFeeBy", "currency_out")
		v.Set("isInBps", "true")
		v.Set("feeReceiver", recipient)
	}
	if p.ammOnly {
		v.Set("includedSources", kyberAMMSources)
	}
	u := fmt.Sprintf("%s/%s/api/v1/routes?%s", p.apiBase, slug, v.Encode())
	httpReq, _ := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	httpReq.Header.Set("x-client-id", p.clientID)

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("kyberswap routes: status %d: %s", resp.StatusCode, string(body))
	}
	var out kyberRoutes
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, err
	}
	if out.Code != 0 || len(out.Data.RouteSummary) == 0 {
		return nil, fmt.Errorf("kyberswap routes: code %d: %s", out.Code, out.Msg)
	}
	return &out, nil
}

func (p *KyberSwapProvider) Price(ctx context.Context, req Request) (*PriceResult, error) {
	routes, err := p.getRoutes(ctx, req)
	if err != nil {
		return nil, err
	}
	var rs kyberRouteSummaryFields
	if err := json.Unmarshal(routes.Data.RouteSummary, &rs); err != nil {
		return nil, err
	}
	if rs.AmountOut == "" {
		return nil, fmt.Errorf("kyberswap: empty amountOut")
	}
	dst, _ := new(big.Int).SetString(rs.AmountOut, 10)
	return &PriceResult{
		Provider:    p.Name(),
		BuyAmount:   rs.AmountOut,
		GasEstimate: rs.Gas,
		Fees:        estimatedOutputFee(req, dst, p.feeRecipients[req.ChainKey], p.feeBps),
	}, nil
}

func (p *KyberSwapProvider) Quote(ctx context.Context, req Request) (*Quote, error) {
	if req.Taker == "" {
		return nil, fmt.Errorf("kyberswap build: taker required")
	}
	routes, err := p.getRoutes(ctx, req)
	if err != nil {
		return nil, err
	}

	slug := kyberChainSlugs[req.ChainKey]
	buildBody, _ := json.Marshal(map[string]any{
		"routeSummary":      json.RawMessage(routes.Data.RouteSummary),
		"sender":            req.Taker,
		"recipient":         req.Taker,
		"slippageTolerance": req.SlippageBps,
		"source":            p.clientID,
	})
	u := fmt.Sprintf("%s/%s/api/v1/route/build", p.apiBase, slug)
	httpReq, _ := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewReader(buildBody))
	httpReq.Header.Set("x-client-id", p.clientID)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("kyberswap build: status %d: %s", resp.StatusCode, string(body))
	}
	var out struct {
		Code int `json:"code"`
		Msg  string `json:"message"`
		Data struct {
			AmountOut        string `json:"amountOut"`
			Data             string `json:"data"`
			RouterAddress    string `json:"routerAddress"`
			Gas              string `json:"gas"`
			TransactionValue string `json:"transactionValue"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, err
	}
	if out.Code != 0 || out.Data.Data == "" || out.Data.RouterAddress == "" {
		return nil, fmt.Errorf("kyberswap build: code %d: %s", out.Code, out.Msg)
	}

	dst, ok := new(big.Int).SetString(out.Data.AmountOut, 10)
	if !ok {
		return nil, fmt.Errorf("kyberswap build: bad amountOut %q", out.Data.AmountOut)
	}
	minBuy := new(big.Int).Mul(dst, big.NewInt(int64(10000-req.SlippageBps)))
	minBuy.Div(minBuy, big.NewInt(10000))

	value := out.Data.TransactionValue
	if value == "" {
		value = "0"
	}

	var approval *Approval
	if !req.SellToken.IsNative() {
		approval = &Approval{
			TokenAddress:   req.SellToken.ContractAddress,
			Spender:        out.Data.RouterAddress,
			RequiredAmount: req.SellAmount.String(),
		}
	}

	return &Quote{
		Provider:     p.Name(),
		BuyAmount:    out.Data.AmountOut,
		MinBuyAmount: minBuy.String(),
		To:           out.Data.RouterAddress,
		Data:         out.Data.Data,
		Value:        value,
		Gas:          out.Data.Gas,
		Approval:     approval,
		Fees:         estimatedOutputFee(req, dst, p.feeRecipients[req.ChainKey], p.feeBps),
	}, nil
}

// estimatedOutputFee approximates the platform fee taken from the buy token:
// the returned amount is post-fee, so fee ≈ out * feeBps / (10000 - feeBps).
// Shared by aggregators that don't return an explicit fee breakdown.
func estimatedOutputFee(req Request, out *big.Int, recipient string, feeBps int) Fees {
	if out == nil || recipient == "" || feeBps <= 0 {
		return Fees{}
	}
	fee := new(big.Int).Mul(out, big.NewInt(int64(feeBps)))
	fee.Div(fee, big.NewInt(int64(10000-feeBps)))
	return Fees{
		IntegratorFeeAmount: fee.String(),
		IntegratorFeeToken:  req.BuyToken.APIAddress(),
	}
}
