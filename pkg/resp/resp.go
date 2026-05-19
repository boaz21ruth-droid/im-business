package resp

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type envelope struct {
	ErrCode int    `json:"errCode"`
	ErrMsg  string `json:"errMsg"`
	Data    any    `json:"data"`
}

func OK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, envelope{ErrCode: 0, ErrMsg: "", Data: data})
}

func Fail(c *gin.Context, httpStatus, errCode int, errMsg string) {
	c.AbortWithStatusJSON(httpStatus, envelope{ErrCode: errCode, ErrMsg: errMsg})
}

func ErrBadRequest(c *gin.Context, msg string) {
	Fail(c, http.StatusBadRequest, 1001, msg)
}

// ErrUnauthorized uses code 1501 — the Flutter app's kickoff handler watches for this.
func ErrUnauthorized(c *gin.Context) {
	Fail(c, http.StatusUnauthorized, 1501, "token expired or invalid")
}

func ErrInternal(c *gin.Context, msg string) {
	Fail(c, http.StatusInternalServerError, 1002, msg)
}
