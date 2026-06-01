package handler

import (
	"errors"
	"math/big"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/web1/im-business/internal/config"
	"github.com/web1/im-business/internal/middleware"
	"github.com/web1/im-business/internal/wallet"
	"github.com/web1/im-business/internal/wallet/bridge"
	"github.com/web1/im-business/internal/wallet/intent"
	"github.com/web1/im-business/internal/wallet/quote"
	"github.com/web1/im-business/pkg/resp"
)

type WalletHandler struct {
	svc        *wallet.WalletService
	swapConfig config.SwapConfig
	agg        *quote.Aggregator
	bridge     bridge.Provider
	intent     intent.Provider
}

func NewWalletHandler(svc *wallet.WalletService, swapConfig config.SwapConfig, agg *quote.Aggregator, br bridge.Provider, in intent.Provider) *WalletHandler {
	return &WalletHandler{svc: svc, swapConfig: swapConfig, agg: agg, bridge: br, intent: in}
}

// GetSwapConfig handles GET /wallet/swap_config.
// Returns the runtime Swap configuration (RPCs, fee recipients, router
// whitelists, limits) so the client can rotate them without an app release.
// Aggregator credentials are NOT included (json:"-" on Providers) — the client
// never holds an aggregator API key.
func (h *WalletHandler) GetSwapConfig(c *gin.Context) {
	resp.OK(c, h.swapConfig)
}

// parseQuoteRequest builds a quote.Request from GET query params. A token param
// that is empty or "native" denotes the chain's native asset.
func parseQuoteRequest(c *gin.Context) (quote.Request, error) {
	chainKey := c.Query("chainKey")
	if chainKey == "" {
		return quote.Request{}, errors.New("chainKey is required")
	}
	amountStr := c.Query("sellAmount")
	amount, ok := new(big.Int).SetString(amountStr, 10)
	if !ok || amount.Sign() <= 0 {
		return quote.Request{}, errors.New("sellAmount must be a positive integer")
	}
	slippage := 50
	if s := c.Query("slippageBps"); s != "" {
		if v, err := strconv.Atoi(s); err == nil && v >= 0 && v <= 5000 {
			slippage = v
		}
	}
	return quote.Request{
		ChainKey:    chainKey,
		SellToken:   parseTokenParam(c.Query("sellToken"), c.Query("sellDecimals")),
		BuyToken:    parseTokenParam(c.Query("buyToken"), c.Query("buyDecimals")),
		SellAmount:  amount,
		Taker:       c.Query("taker"),
		SlippageBps: slippage,
	}, nil
}

func parseTokenParam(addr, decimals string) quote.Token {
	d, _ := strconv.Atoi(decimals)
	if addr == "" || addr == "native" {
		return quote.Token{Decimals: d}
	}
	return quote.Token{ContractAddress: addr, Decimals: d}
}

func writeQuoteErr(c *gin.Context, err error) {
	switch {
	case errors.Is(err, quote.ErrNoQuote):
		resp.Biz(c, resp.CodeSwapNoQuote, "no quote available")
	case errors.Is(err, quote.ErrUnsupportedChain):
		resp.Biz(c, resp.CodeSwapChainUnsupported, "chain not supported for swap")
	default:
		resp.Biz(c, resp.CodeSwapNoQuote, err.Error())
	}
}

// GetSwapPrice handles GET /wallet/price — soft quote for the live preview.
func (h *WalletHandler) GetSwapPrice(c *gin.Context) {
	req, err := parseQuoteRequest(c)
	if err != nil {
		resp.Biz(c, resp.CodeSwapBadParams, err.Error())
		return
	}
	if !h.agg.SupportsChain(req.ChainKey) {
		resp.Biz(c, resp.CodeSwapChainUnsupported, "chain not supported for swap")
		return
	}
	out, err := h.agg.BestPrice(c.Request.Context(), req)
	if err != nil {
		writeQuoteErr(c, err)
		return
	}
	resp.OK(c, out)
}

// GetSwapQuote handles GET /wallet/quote — firm quote with executable calldata.
// taker is required so the aggregator can produce signable calldata.
func (h *WalletHandler) GetSwapQuote(c *gin.Context) {
	req, err := parseQuoteRequest(c)
	if err != nil {
		resp.Biz(c, resp.CodeSwapBadParams, err.Error())
		return
	}
	if req.Taker == "" {
		resp.Biz(c, resp.CodeSwapBadParams, "taker is required")
		return
	}
	if !h.agg.SupportsChain(req.ChainKey) {
		resp.Biz(c, resp.CodeSwapChainUnsupported, "chain not supported for swap")
		return
	}
	out, err := h.agg.BestQuote(c.Request.Context(), req)
	if err != nil {
		writeQuoteErr(c, err)
		return
	}
	resp.OK(c, out)
}

// GetBridgeQuote handles GET /wallet/bridge/quote — a cross-chain route with a
// source-chain signable tx. The client signs+broadcasts it and then polls
// /wallet/bridge/status for delivery.
func (h *WalletHandler) GetBridgeQuote(c *gin.Context) {
	fromChain := c.Query("fromChain")
	toChain := c.Query("toChain")
	if fromChain == "" || toChain == "" {
		resp.Biz(c, resp.CodeBridgeBadParams, "fromChain and toChain are required")
		return
	}
	amount, ok := new(big.Int).SetString(c.Query("fromAmount"), 10)
	if !ok || amount.Sign() <= 0 {
		resp.Biz(c, resp.CodeBridgeBadParams, "fromAmount must be a positive integer")
		return
	}
	from := c.Query("fromAddress")
	if from == "" {
		resp.Biz(c, resp.CodeBridgeBadParams, "fromAddress is required")
		return
	}
	if !h.bridge.SupportsChain(fromChain) || !h.bridge.SupportsChain(toChain) {
		resp.Biz(c, resp.CodeBridgeChainUnsupported, "chain not supported for bridging")
		return
	}
	slippage := 100
	if s := c.Query("slippageBps"); s != "" {
		if v, err := strconv.Atoi(s); err == nil && v >= 0 && v <= 5000 {
			slippage = v
		}
	}
	q, err := h.bridge.Quote(c.Request.Context(), bridge.Request{
		FromChain:   fromChain,
		ToChain:     toChain,
		FromToken:   c.Query("fromToken"),
		ToToken:     c.Query("toToken"),
		FromAmount:  amount,
		FromAddress: from,
		ToAddress:   c.Query("toAddress"),
		SlippageBps: slippage,
	})
	if err != nil {
		switch {
		case errors.Is(err, bridge.ErrNoRoute):
			resp.Biz(c, resp.CodeBridgeNoRoute, "no bridge route available")
		case errors.Is(err, bridge.ErrUnsupportedChain):
			resp.Biz(c, resp.CodeBridgeChainUnsupported, "chain not supported for bridging")
		default:
			resp.Biz(c, resp.CodeBridgeNoRoute, err.Error())
		}
		return
	}
	resp.OK(c, q)
}

// GetBridgeStatus handles GET /wallet/bridge/status — cross-chain delivery state
// for a broadcast source tx.
func (h *WalletHandler) GetBridgeStatus(c *gin.Context) {
	txHash := c.Query("txHash")
	if txHash == "" {
		resp.Biz(c, resp.CodeBridgeBadParams, "txHash is required")
		return
	}
	st, err := h.bridge.Status(c.Request.Context(), c.Query("fromChain"), c.Query("toChain"), txHash, c.Query("tool"))
	if err != nil {
		resp.Biz(c, resp.CodeBridgeNoRoute, err.Error())
		return
	}
	resp.OK(c, st)
}

// GetIntentQuote handles GET /wallet/intent/quote — a ready-to-sign EIP-712
// order (CoW) + domain/approval metadata. The client signs it and POSTs it back.
func (h *WalletHandler) GetIntentQuote(c *gin.Context) {
	chainKey := c.Query("chainKey")
	from := c.Query("from")
	amount, ok := new(big.Int).SetString(c.Query("sellAmount"), 10)
	if chainKey == "" || from == "" || !ok || amount.Sign() <= 0 {
		resp.Biz(c, resp.CodeIntentBadParams, "chainKey, from, positive sellAmount required")
		return
	}
	if !h.intent.SupportsChain(chainKey) {
		resp.Biz(c, resp.CodeIntentChainUnsupported, "chain not supported for intent swap")
		return
	}
	slippage := 50
	if s := c.Query("slippageBps"); s != "" {
		if v, err := strconv.Atoi(s); err == nil && v >= 0 && v <= 5000 {
			slippage = v
		}
	}
	out, err := h.intent.Quote(c.Request.Context(), intent.Request{
		ChainKey:    chainKey,
		SellToken:   c.Query("sellToken"),
		BuyToken:    c.Query("buyToken"),
		SellAmount:  amount,
		From:        from,
		SlippageBps: slippage,
	})
	if err != nil {
		writeIntentErr(c, err)
		return
	}
	resp.OK(c, out)
}

// PostIntentOrder handles POST /wallet/intent/order — submit a signed order.
func (h *WalletHandler) PostIntentOrder(c *gin.Context) {
	var body struct {
		ChainKey string               `json:"chainKey"`
		Order    intent.Order         `json:"order"`
		Sig      string               `json:"signature"`
		From     string               `json:"from"`
		QuoteID  int64                `json:"quoteId"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || body.ChainKey == "" || body.Sig == "" {
		resp.Biz(c, resp.CodeIntentBadParams, "chainKey, order, signature required")
		return
	}
	uid, err := h.intent.Submit(c.Request.Context(), body.ChainKey, intent.SubmitRequest{
		Order: body.Order, Signature: body.Sig, From: body.From, QuoteID: body.QuoteID,
	})
	if err != nil {
		writeIntentErr(c, err)
		return
	}
	resp.OK(c, gin.H{"orderUid": uid})
}

// GetIntentStatus handles GET /wallet/intent/status — order settlement state.
func (h *WalletHandler) GetIntentStatus(c *gin.Context) {
	chainKey := c.Query("chainKey")
	orderUID := c.Query("orderUid")
	if chainKey == "" || orderUID == "" {
		resp.Biz(c, resp.CodeIntentBadParams, "chainKey and orderUid required")
		return
	}
	st, err := h.intent.Status(c.Request.Context(), chainKey, orderUID)
	if err != nil {
		writeIntentErr(c, err)
		return
	}
	resp.OK(c, st)
}

func writeIntentErr(c *gin.Context, err error) {
	switch {
	case errors.Is(err, intent.ErrUnsupportedChain):
		resp.Biz(c, resp.CodeIntentChainUnsupported, "chain not supported for intent swap")
	case errors.Is(err, intent.ErrNoQuote):
		resp.Biz(c, resp.CodeIntentNoQuote, "no intent quote available")
	default:
		resp.Biz(c, resp.CodeIntentNoQuote, err.Error())
	}
}

// RegisterAddresses handles POST /wallet/addresses.
func (h *WalletHandler) RegisterAddresses(c *gin.Context) {
	var req struct {
		Addresses map[string]string `json:"addresses"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || len(req.Addresses) == 0 {
		resp.ErrBadRequest(c, "invalid addresses")
		return
	}

	userID := c.GetString(middleware.UserIDKey)
	if err := h.svc.RegisterAddresses(c.Request.Context(), userID, req.Addresses); err != nil {
		resp.ErrInternal(c, err.Error())
		return
	}
	resp.OK(c, nil)
}

// GetAddresses handles GET /wallet/addresses.
func (h *WalletHandler) GetAddresses(c *gin.Context) {
	userID := c.GetString(middleware.UserIDKey)
	addresses, err := h.svc.GetRegisteredAddresses(c.Request.Context(), userID)
	if err != nil {
		resp.ErrInternal(c, err.Error())
		return
	}
	resp.OK(c, gin.H{
		"hasWallet": len(addresses) > 0,
		"addresses": addresses,
	})
}

// GetTxHistory handles GET /wallet/tx-history.
func (h *WalletHandler) GetTxHistory(c *gin.Context) {
	chainKey := c.Query("chain")
	if chainKey == "" {
		resp.ErrBadRequest(c, "chain is required")
		return
	}

	contract := c.Query("contract")
	page := 1
	limit := 20

	if p := c.Query("page"); p != "" {
		if v, err := strconv.Atoi(p); err == nil && v > 0 {
			page = v
		}
	}
	if l := c.Query("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 && v <= 50 {
			limit = v
		}
	}

	userID := c.GetString(middleware.UserIDKey)
	records, fromCache, err := h.svc.GetHistory(c.Request.Context(), userID, chainKey, contract, page, limit)
	if err != nil {
		resp.ErrBadRequest(c, err.Error())
		return
	}

	resp.OK(c, gin.H{
		"records":   records,
		"hasMore":   len(records) == limit,
		"fromCache": fromCache,
	})
}
