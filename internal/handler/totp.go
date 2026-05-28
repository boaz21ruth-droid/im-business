package handler

import (
	"context"
	"errors"

	"github.com/gin-gonic/gin"

	"github.com/web1/im-business/internal/middleware"
	"github.com/web1/im-business/internal/service"
	"github.com/web1/im-business/pkg/resp"
)

type TotpHandler struct {
	svc     *service.TotpService
	userSvc *service.UserService
}

func NewTotpHandler(svc *service.TotpService, userSvc *service.UserService) *TotpHandler {
	return &TotpHandler{svc: svc, userSvc: userSvc}
}

// Setup → POST /user/totp/setup
//
// Generates a new secret, caches it (10 min), and returns the secret + otpauth URL.
// The client should render the otpauth URL as a QR code and prompt the user to scan.
func (h *TotpHandler) Setup(c *gin.Context) {
	userID := c.GetString(middleware.UserIDKey)
	accountName := h.accountName(c.Request.Context(), userID)
	secret, url, err := h.svc.Setup(c.Request.Context(), userID, accountName)
	if errors.Is(err, service.ErrTotpAlreadyEnabled) {
		resp.ErrTotpAlreadyEnabled(c)
		return
	}
	if err != nil {
		resp.ErrInternal(c, err.Error())
		return
	}
	resp.OK(c, gin.H{
		"secret":     secret,
		"otpauthUrl": url,
	})
}

// Enable → POST /user/totp/enable
//
// Body: {"code": "123456"}
func (h *TotpHandler) Enable(c *gin.Context) {
	var req struct {
		Code string `json:"code"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Code == "" {
		resp.ErrBadRequest(c, "code required")
		return
	}
	userID := c.GetString(middleware.UserIDKey)
	err := h.svc.Enable(c.Request.Context(), userID, req.Code)
	switch {
	case errors.Is(err, service.ErrTotpAlreadyEnabled):
		resp.ErrTotpAlreadyEnabled(c)
	case errors.Is(err, service.ErrTotpSetupExpired):
		resp.ErrTotpSetupExpired(c)
	case errors.Is(err, service.ErrTotpInvalidCode):
		resp.ErrTotpInvalid(c)
	case err != nil:
		resp.ErrInternal(c, err.Error())
	default:
		resp.OK(c, nil)
	}
}

// Disable → POST /user/totp/disable
//
// Body: {"code": "123456"}  (current code, proves possession of authenticator)
func (h *TotpHandler) Disable(c *gin.Context) {
	var req struct {
		Code string `json:"code"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Code == "" {
		resp.ErrBadRequest(c, "code required")
		return
	}
	userID := c.GetString(middleware.UserIDKey)
	err := h.svc.Disable(c.Request.Context(), userID, req.Code)
	switch {
	case errors.Is(err, service.ErrTotpNotEnabled):
		resp.ErrTotpNotEnabled(c)
	case errors.Is(err, service.ErrTotpInvalidCode):
		resp.ErrTotpInvalid(c)
	case err != nil:
		resp.ErrInternal(c, err.Error())
	default:
		resp.OK(c, nil)
	}
}

// Status → GET /user/totp/status
func (h *TotpHandler) Status(c *gin.Context) {
	userID := c.GetString(middleware.UserIDKey)
	enabled, err := h.svc.Status(c.Request.Context(), userID)
	if err != nil {
		resp.ErrInternal(c, err.Error())
		return
	}
	resp.OK(c, gin.H{"enabled": enabled})
}

// WalletVerify → POST /wallet/totp/verify
//
// Body: {"code": "123456"}
// For users who haven't enabled TOTP, returns {valid: true} unconditionally so
// the client can call this without conditional logic.
func (h *TotpHandler) WalletVerify(c *gin.Context) {
	var req struct {
		Code string `json:"code"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.ErrBadRequest(c, "code required")
		return
	}
	userID := c.GetString(middleware.UserIDKey)
	err := h.svc.Verify(c.Request.Context(), userID, req.Code)
	switch {
	case errors.Is(err, service.ErrTotpLocked):
		resp.ErrTotpLocked(c)
	case errors.Is(err, service.ErrTotpInvalidCode):
		resp.ErrTotpInvalid(c)
	case err != nil:
		resp.ErrInternal(c, err.Error())
	default:
		resp.OK(c, gin.H{"valid": true})
	}
}

// accountName is the display name shown next to the issuer in the user's
// Authenticator app. Prefer email → phone → userID.
func (h *TotpHandler) accountName(ctx context.Context, userID string) string {
	users, err := h.userSvc.FindByIDs(ctx, []string{userID})
	if err != nil || len(users) == 0 {
		return userID
	}
	u := users[0]
	if u.Email != "" {
		return u.Email
	}
	if u.PhoneNumber != "" {
		return u.PhoneNumber
	}
	return userID
}
