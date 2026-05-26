package handler

import (
	"context"
	"io"

	"github.com/gin-gonic/gin"
	"github.com/web1/im-business/internal/wallet"
	"github.com/web1/im-business/internal/wallet/provider"
)

type WebhookHandler struct {
	walletSvc       *wallet.WalletService
	streamProviders map[string]provider.StreamProvider
}

func NewWebhookHandler(walletSvc *wallet.WalletService, streamProviders map[string]provider.StreamProvider) *WebhookHandler {
	return &WebhookHandler{walletSvc: walletSvc, streamProviders: streamProviders}
}

// HandleWebhook handles POST /webhook/:provider.
// No auth header — protected by per-provider signature verification.
func (h *WebhookHandler) HandleWebhook(c *gin.Context) {
	providerName := c.Param("provider")
	sp, ok := h.streamProviders[providerName]
	if !ok {
		c.JSON(404, gin.H{"error": "unknown provider"})
		return
	}

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(400, gin.H{"error": "bad request"})
		return
	}

	if !sp.VerifyWebhook(c.Request, body) {
		c.JSON(401, gin.H{"error": "invalid signature"})
		return
	}

	// Respond immediately — providers retry on non-2xx.
	c.JSON(200, gin.H{})

	chainKey, confirmed, transfers, err := sp.ParseWebhookPayload(body)
	if err != nil || !confirmed || chainKey == "" || len(transfers) == 0 {
		return
	}

	ctx := context.Background()
	for _, t := range transfers {
		h.walletSvc.ProcessWebhookTransfer(ctx, chainKey, t)
	}
}
