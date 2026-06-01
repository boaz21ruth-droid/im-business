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

// paraswapChainIDs maps our chain keys to Paraswap `network` IDs.
var paraswapChainIDs = map[string]int{
	"eth":      1,
	"bsc":      56,
	"polygon":  137,
	"arbitrum": 42161,
	"optimism": 10,
}

// ParaswapProvider integrates the Paraswap/Velora API v6.2. Keyless (no KYC):
// an optional partner name attributes volume. Two calls: /prices (price) then
// /transactions (calldata). Requires token decimals.
// paraswapAMMVenues is an allow-list of on-chain AMM DEXs used in AMM-only
// (fork-test) mode. An allow-list is robust: any PMM/RFQ/limit-order venue
// (which can't settle on a frozen fork) is excluded by default.
const paraswapAMMVenues = "UniswapV2,UniswapV3,UniswapV4,SushiSwap,Curve,CurveV1,CurveV2,BalancerV1,BalancerV2,MakerPSM,PancakeSwapV3"

type ParaswapProvider struct {
	apiBase       string
	partner       string
	feeRecipients map[string]string
	feeBps        int
	ammOnly       bool
	client        *http.Client
}

func NewParaswapProvider(apiBase, partner string, feeRecipients map[string]string, feeBps int, ammOnly bool) *ParaswapProvider {
	if apiBase == "" {
		apiBase = "https://api.paraswap.io"
	}
	if partner == "" {
		partner = "im-wallet"
	}
	return &ParaswapProvider{
		apiBase:       apiBase,
		partner:       partner,
		feeRecipients: feeRecipients,
		feeBps:        feeBps,
		ammOnly:       ammOnly,
		client:        &http.Client{Timeout: 10 * time.Second},
	}
}

func (p *ParaswapProvider) Name() string { return "paraswap" }

func (p *ParaswapProvider) SupportsChain(chainKey string) bool {
	_, ok := paraswapChainIDs[chainKey]
	return ok // keyless — no API key needed
}

func (p *ParaswapProvider) getPriceRoute(ctx context.Context, req Request) (json.RawMessage, *paraswapRoute, error) {
	chainID := paraswapChainIDs[req.ChainKey]
	if chainID == 0 {
		return nil, nil, fmt.Errorf("paraswap %w: %s", ErrUnsupportedChain, req.ChainKey)
	}
	v := url.Values{}
	v.Set("srcToken", req.SellToken.APIAddress())
	v.Set("destToken", req.BuyToken.APIAddress())
	v.Set("amount", req.SellAmount.String())
	v.Set("srcDecimals", strconv.Itoa(req.SellToken.Decimals))
	v.Set("destDecimals", strconv.Itoa(req.BuyToken.Decimals))
	v.Set("side", "SELL")
	v.Set("network", strconv.Itoa(chainID))
	v.Set("version", "6.2") // pin Augustus v6.2; without this the default could drift
	v.Set("partner", p.partner)
	if p.ammOnly {
		v.Set("includeDEXS", paraswapAMMVenues)
	}
	u := fmt.Sprintf("%s/prices/?%s", p.apiBase, v.Encode())
	httpReq, _ := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, nil, fmt.Errorf("paraswap prices: status %d: %s", resp.StatusCode, string(body))
	}
	var out struct {
		PriceRoute json.RawMessage `json:"priceRoute"`
	}
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, nil, err
	}
	if len(out.PriceRoute) == 0 {
		return nil, nil, fmt.Errorf("paraswap prices: no route: %s", string(body))
	}
	var fields paraswapRoute
	if err := json.Unmarshal(out.PriceRoute, &fields); err != nil {
		return nil, nil, err
	}
	if fields.DestAmount == "" {
		return nil, nil, fmt.Errorf("paraswap prices: empty destAmount")
	}
	return out.PriceRoute, &fields, nil
}

type paraswapRoute struct {
	DestAmount         string `json:"destAmount"`
	GasCost            string `json:"gasCost"`
	TokenTransferProxy string `json:"tokenTransferProxy"`
}

func (p *ParaswapProvider) Price(ctx context.Context, req Request) (*PriceResult, error) {
	_, route, err := p.getPriceRoute(ctx, req)
	if err != nil {
		return nil, err
	}
	dst, _ := new(big.Int).SetString(route.DestAmount, 10)
	return &PriceResult{
		Provider:    p.Name(),
		BuyAmount:   route.DestAmount,
		GasEstimate: route.GasCost,
		Fees:        estimatedOutputFee(req, dst, p.feeRecipients[req.ChainKey], p.feeBps),
	}, nil
}

func (p *ParaswapProvider) Quote(ctx context.Context, req Request) (*Quote, error) {
	if req.Taker == "" {
		return nil, fmt.Errorf("paraswap build: taker required")
	}
	priceRoute, route, err := p.getPriceRoute(ctx, req)
	if err != nil {
		return nil, err
	}

	chainID := paraswapChainIDs[req.ChainKey]
	bodyMap := map[string]any{
		"srcToken":     req.SellToken.APIAddress(),
		"destToken":    req.BuyToken.APIAddress(),
		"srcAmount":    req.SellAmount.String(),
		"srcDecimals":  req.SellToken.Decimals,
		"destDecimals": req.BuyToken.Decimals,
		"priceRoute":   priceRoute,
		"userAddress":  req.Taker,
		"slippage":     req.SlippageBps,
		"partner":      p.partner,
	}
	if recipient := p.feeRecipients[req.ChainKey]; recipient != "" && p.feeBps > 0 {
		bodyMap["partnerAddress"] = recipient
		bodyMap["partnerFeeBps"] = p.feeBps
		bodyMap["takeSurplus"] = false
		// Transfer the fee to partnerAddress per-swap instead of accruing it in
		// Paraswap's FeeClaimer (the default) for later manual claiming.
		bodyMap["isDirectFeeTransfer"] = true
	}
	reqBody, _ := json.Marshal(bodyMap)

	// ignoreChecks=true skips Paraswap's allowance/balance pre-checks so we get
	// calldata even before the user has approved.
	u := fmt.Sprintf("%s/transactions/%d?ignoreChecks=true", p.apiBase, chainID)
	httpReq, _ := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewReader(reqBody))
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("paraswap transactions: status %d: %s", resp.StatusCode, string(body))
	}
	var tx struct {
		To       string `json:"to"`
		Data     string `json:"data"`
		Value    string `json:"value"`
		GasPrice string `json:"gasPrice"`
		Gas      string `json:"gas"`
	}
	if err := json.Unmarshal(body, &tx); err != nil {
		return nil, err
	}
	if tx.To == "" || tx.Data == "" {
		return nil, fmt.Errorf("paraswap transactions: incomplete: %s", string(body))
	}

	dst, ok := new(big.Int).SetString(route.DestAmount, 10)
	if !ok {
		return nil, fmt.Errorf("paraswap: bad destAmount %q", route.DestAmount)
	}
	minBuy := new(big.Int).Mul(dst, big.NewInt(int64(10000-req.SlippageBps)))
	minBuy.Div(minBuy, big.NewInt(10000))

	value := tx.Value
	if value == "" {
		value = "0"
	}

	// Paraswap pulls tokens via the TokenTransferProxy, which is the approval
	// spender (not the tx `to`).
	var approval *Approval
	if !req.SellToken.IsNative() && route.TokenTransferProxy != "" {
		approval = &Approval{
			TokenAddress:   req.SellToken.ContractAddress,
			Spender:        route.TokenTransferProxy,
			RequiredAmount: req.SellAmount.String(),
		}
	}

	return &Quote{
		Provider:     p.Name(),
		BuyAmount:    route.DestAmount,
		MinBuyAmount: minBuy.String(),
		To:           tx.To,
		Data:         tx.Data,
		Value:        value,
		Gas:          tx.Gas,
		GasPrice:     tx.GasPrice,
		Approval:     approval,
		Fees:         estimatedOutputFee(req, dst, p.feeRecipients[req.ChainKey], p.feeBps),
	}, nil
}
