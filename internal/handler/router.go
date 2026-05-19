package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/web1/im-business/internal/middleware"
	jwtpkg "github.com/web1/im-business/pkg/jwt"
	"go.uber.org/zap"
)

func NewRouter(log *zap.Logger, jwt *jwtpkg.Service, acc *AccountHandler, user *UserHandler) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.Logger(log))

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
	}

	r.POST("/friend/search", auth, user.SearchFriendInfo)

	return r
}
