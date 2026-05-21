package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/web1/im-business/internal/middleware"
	"github.com/web1/im-business/internal/service"
	"github.com/web1/im-business/pkg/resp"
)

type AccountHandler struct{ svc *service.AccountService }

func NewAccountHandler(svc *service.AccountService) *AccountHandler {
	return &AccountHandler{svc: svc}
}

func (h *AccountHandler) SendCode(c *gin.Context) {
	var req struct {
		AreaCode    string `json:"areaCode"`
		PhoneNumber string `json:"phoneNumber"`
		Email       string `json:"email"`
		UsedFor     int    `json:"usedFor"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.ErrBadRequest(c, err.Error())
		return
	}
	target := firstNE(req.Email, req.PhoneNumber)
	if target == "" {
		resp.ErrBadRequest(c, "email or phoneNumber required")
		return
	}
	if err := h.svc.SendCode(c.Request.Context(), target); err != nil {
		resp.ErrInternal(c, err.Error())
		return
	}
	resp.OK(c, nil)
}

// VerifyCode checks the submitted code without consuming it (consumed in Register/ResetPassword).
func (h *AccountHandler) VerifyCode(c *gin.Context) {
	var req struct {
		AreaCode    string `json:"areaCode"`
		PhoneNumber string `json:"phoneNumber"`
		Email       string `json:"email"`
		VerifyCode  string `json:"verifyCode"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.ErrBadRequest(c, err.Error())
		return
	}
	target := firstNE(req.Email, req.PhoneNumber)
	if target == "" || req.VerifyCode == "" {
		resp.ErrBadRequest(c, "email/phone and verifyCode required")
		return
	}
	if err := h.svc.VerifyCode(c.Request.Context(), target, req.VerifyCode); err != nil {
		resp.Fail(c, http.StatusBadRequest, 1007, err.Error())
		return
	}
	resp.OK(c, nil)
}

func (h *AccountHandler) Register(c *gin.Context) {
	var req struct {
		VerifyCode string `json:"verifyCode"`
		Platform   int    `json:"platform"`
		User       struct {
			Nickname    string `json:"nickname"`
			FaceURL     string `json:"faceURL"`
			Birth       int64  `json:"birth"`
			Gender      int8   `json:"gender"`
			Email       string `json:"email"`
			AreaCode    string `json:"areaCode"`
			PhoneNumber string `json:"phoneNumber"`
			Account     string `json:"account"`
			Password    string `json:"password"`
		} `json:"user"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.ErrBadRequest(c, err.Error())
		return
	}
	if req.VerifyCode == "" {
		resp.ErrBadRequest(c, "verifyCode required")
		return
	}

	result, err := h.svc.Register(c.Request.Context(), service.RegisterInput{
		Account:     req.User.Account,
		PhoneNumber: req.User.PhoneNumber,
		AreaCode:    req.User.AreaCode,
		Email:       req.User.Email,
		Password:    req.User.Password,
		Nickname:    req.User.Nickname,
		FaceURL:     req.User.FaceURL,
		Gender:      req.User.Gender,
		Birth:       req.User.Birth,
		VerifyCode:  req.VerifyCode,
		Platform:    req.Platform,
	})
	if err != nil {
		resp.Fail(c, http.StatusBadRequest, 1003, err.Error())
		return
	}
	resp.OK(c, gin.H{"userID": result.UserID, "imToken": result.IMToken, "chatToken": result.ChatToken})
}

func (h *AccountHandler) Login(c *gin.Context) {
	var req struct {
		AreaCode    string `json:"areaCode"`
		PhoneNumber string `json:"phoneNumber"`
		Account     string `json:"account"`
		Email       string `json:"email"`
		Password    string `json:"password"`
		Platform    int    `json:"platform"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.ErrBadRequest(c, err.Error())
		return
	}
	account := firstNE(req.Account, req.Email, req.PhoneNumber)
	if account == "" || req.Password == "" {
		resp.ErrBadRequest(c, "account and password required")
		return
	}
	result, err := h.svc.Login(c.Request.Context(), service.LoginInput{
		Account: account, Password: req.Password, Platform: req.Platform,
	})
	if err != nil {
		resp.Fail(c, http.StatusUnauthorized, 1004, err.Error())
		return
	}
	resp.OK(c, gin.H{"userID": result.UserID, "imToken": result.IMToken, "chatToken": result.ChatToken})
}

func (h *AccountHandler) ResetPassword(c *gin.Context) {
	var req struct {
		PhoneNumber string `json:"phoneNumber"`
		Email       string `json:"email"`
		Password    string `json:"password"`
		VerifyCode  string `json:"verifyCode"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.ErrBadRequest(c, err.Error())
		return
	}
	target := firstNE(req.Email, req.PhoneNumber)
	if err := h.svc.ResetPassword(c.Request.Context(), target, req.VerifyCode, req.Password); err != nil {
		resp.Fail(c, http.StatusBadRequest, 1005, err.Error())
		return
	}
	resp.OK(c, nil)
}

func (h *AccountHandler) ChangePassword(c *gin.Context) {
	userID := c.GetString(middleware.UserIDKey)
	var req struct {
		CurrentPassword string `json:"currentPassword"`
		NewPassword     string `json:"newPassword"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.ErrBadRequest(c, err.Error())
		return
	}
	if err := h.svc.ChangePassword(c.Request.Context(), userID, req.CurrentPassword, req.NewPassword); err != nil {
		resp.Fail(c, http.StatusBadRequest, 1006, err.Error())
		return
	}
	resp.OK(c, nil)
}

func firstNE(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}
