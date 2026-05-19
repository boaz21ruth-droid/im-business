package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	jwtpkg "github.com/web1/im-business/pkg/jwt"
)

func TestAuthMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	jwtSvc := jwtpkg.NewService("secret", 168*time.Hour)
	token, _ := jwtSvc.Sign("user42")

	r := gin.New()
	r.Use(Auth(jwtSvc))
	r.GET("/t", func(c *gin.Context) { c.String(200, c.GetString(UserIDKey)) })

	// valid token via "token" header
	req := httptest.NewRequest(http.MethodGet, "/t", nil)
	req.Header.Set("token", token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != 200 || w.Body.String() != "user42" {
		t.Errorf("valid token: %d %s", w.Code, w.Body.String())
	}

	// no token → 401
	req2 := httptest.NewRequest(http.MethodGet, "/t", nil)
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)
	if w2.Code != 401 {
		t.Errorf("no token: want 401, got %d", w2.Code)
	}
}
