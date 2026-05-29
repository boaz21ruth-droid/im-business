package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/web1/im-business/internal/config"
	"github.com/web1/im-business/internal/middleware"
	"github.com/web1/im-business/internal/wallet"
	"github.com/web1/im-business/pkg/resp"
)

type WalletHandler struct {
	svc        *wallet.WalletService
	swapConfig config.SwapConfig
}

func NewWalletHandler(svc *wallet.WalletService, swapConfig config.SwapConfig) *WalletHandler {
	return &WalletHandler{svc: svc, swapConfig: swapConfig}
}

// GetSwapConfig handles GET /wallet/swap_config.
// Returns the runtime Swap configuration so the Flutter client can rotate API
// keys, RPC URLs, fee recipients, and router whitelists without an app release.
func (h *WalletHandler) GetSwapConfig(c *gin.Context) {
	resp.OK(c, h.swapConfig)
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
