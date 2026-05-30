package handler

import (
	"errors"
	"math/big"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/web1/im-business/internal/config"
	"github.com/web1/im-business/internal/middleware"
	"github.com/web1/im-business/internal/wallet"
	"github.com/web1/im-business/internal/wallet/quote"
	"github.com/web1/im-business/pkg/resp"
)

type WalletHandler struct {
	svc        *wallet.WalletService
	swapConfig config.SwapConfig
	agg        *quote.Aggregator
}

func NewWalletHandler(svc *wallet.WalletService, swapConfig config.SwapConfig, agg *quote.Aggregator) *WalletHandler {
	return &WalletHandler{svc: svc, swapConfig: swapConfig, agg: agg}
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
		SellToken:   parseTokenParam(c.Query("sellToken")),
		BuyToken:    parseTokenParam(c.Query("buyToken")),
		SellAmount:  amount,
		Taker:       c.Query("taker"),
		SlippageBps: slippage,
	}, nil
}

func parseTokenParam(v string) quote.Token {
	if v == "" || v == "native" {
		return quote.Token{}
	}
	return quote.Token{ContractAddress: v}
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
