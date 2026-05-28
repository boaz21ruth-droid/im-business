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

// TOTP-related error codes (1600–1604).
const (
	CodeTotpInvalid       = 1600
	CodeTotpLocked        = 1601
	CodeTotpAlreadyEnable = 1602
	CodeTotpNotEnabled    = 1603
	CodeTotpSetupExpired  = 1604
)

func ErrTotpInvalid(c *gin.Context) {
	Fail(c, http.StatusBadRequest, CodeTotpInvalid, "验证码无效")
}

func ErrTotpLocked(c *gin.Context) {
	Fail(c, http.StatusTooManyRequests, CodeTotpLocked, "验证失败次数过多，请稍后再试")
}

func ErrTotpAlreadyEnabled(c *gin.Context) {
	Fail(c, http.StatusBadRequest, CodeTotpAlreadyEnable, "已绑定，请先解绑")
}

func ErrTotpNotEnabled(c *gin.Context) {
	Fail(c, http.StatusBadRequest, CodeTotpNotEnabled, "未绑定")
}

func ErrTotpSetupExpired(c *gin.Context) {
	Fail(c, http.StatusBadRequest, CodeTotpSetupExpired, "二维码已过期，请刷新后重新扫描")
}
