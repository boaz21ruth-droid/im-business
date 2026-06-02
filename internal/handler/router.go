package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/web1/im-business/internal/middleware"
	jwtpkg "github.com/web1/im-business/pkg/jwt"
	"go.uber.org/zap"
)

func NewRouter(log *zap.Logger, jwt *jwtpkg.Service, acc *AccountHandler, user *UserHandler, wal *WalletHandler, wh *WebhookHandler, totp *TotpHandler) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.Logger(log))
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, token, operationID")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	auth := middleware.Auth(jwt)

	r.GET("/healthz", func(c *gin.Context) { c.String(200, "ok") })

	// Stub routes that Flutter demo calls but don't need real logic yet
	r.GET("/app/check", func(c *gin.Context) {
		c.JSON(200, gin.H{"errCode": 0, "errMsg": "", "data": nil})
	})
	r.POST("/client_config/get", func(c *gin.Context) {
		c.JSON(200, gin.H{"errCode": 0, "errMsg": "", "data": gin.H{
			"discoverPageURL":       "discover",
			"allowSendMsgNotFriend": "1",
		}})
	})

	a := r.Group("/account")
	{
		a.POST("/code/send", acc.SendCode)
		a.POST("/code/verify", acc.VerifyCode)
		a.POST("/register", acc.Register)
		a.POST("/login", acc.Login)
		a.POST("/password/reset", acc.ResetPassword)
		a.POST("/password/change", auth, acc.ChangePassword)
	}

	u := r.Group("/user", auth)
	{
		u.POST("/update", user.UpdateUserInfo)
		u.POST("/find/full", user.FindUsersFullInfo)
		u.POST("/search/full", user.SearchUsersFullInfo)
		// RTC token stub — voice/video calls not yet supported
		u.POST("/rtc/get_token", func(c *gin.Context) {
			c.JSON(200, gin.H{"errCode": 0, "errMsg": "", "data": gin.H{
				"token":   "",
				"liveURL": "",
			}})
		})
		u.POST("/totp/setup", totp.Setup)
		u.POST("/totp/enable", totp.Enable)
		u.POST("/totp/disable", totp.Disable)
		u.GET("/totp/status", totp.Status)
	}

	r.POST("/friend/search", auth, user.SearchFriendInfo)

	w := r.Group("/wallet", auth)
	{
		w.POST("/addresses", wal.RegisterAddresses)
		w.GET("/addresses", wal.GetAddresses)
		w.GET("/friend/address", wal.GetFriendAddress)
		w.GET("/tx-history", wal.GetTxHistory)
		w.GET("/swap_config", wal.GetSwapConfig)
		w.GET("/price", wal.GetSwapPrice)
		w.GET("/quote", wal.GetSwapQuote)
		w.GET("/bridge/quote", wal.GetBridgeQuote)
		w.GET("/bridge/status", wal.GetBridgeStatus)
		w.GET("/intent/quote", wal.GetIntentQuote)
		w.POST("/intent/order", wal.PostIntentOrder)
		w.GET("/intent/status", wal.GetIntentStatus)
		w.POST("/totp/verify", totp.WalletVerify)
	}

	r.POST("/webhook/:provider", wh.HandleWebhook)

	return r
}
