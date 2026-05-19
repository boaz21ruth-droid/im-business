package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	jwtpkg "github.com/web1/im-business/pkg/jwt"
	"github.com/web1/im-business/pkg/resp"
)

const UserIDKey = "userID"

func Auth(jwt *jwtpkg.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("token")
		if token == "" {
			if auth := c.GetHeader("Authorization"); strings.HasPrefix(auth, "Bearer ") {
				token = strings.TrimPrefix(auth, "Bearer ")
			}
		}
		if token == "" {
			resp.ErrUnauthorized(c)
			return
		}
		userID, err := jwt.Verify(token)
		if err != nil {
			resp.ErrUnauthorized(c)
			return
		}
		c.Set(UserIDKey, userID)
		c.Next()
	}
}
