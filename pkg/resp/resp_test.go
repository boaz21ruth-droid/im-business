package resp

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestOK(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	OK(c, map[string]string{"key": "val"})

	var out struct {
		ErrCode int            `json:"errCode"`
		Data    map[string]any `json:"data"`
	}
	json.NewDecoder(w.Body).Decode(&out)
	if out.ErrCode != 0 {
		t.Errorf("errCode: want 0, got %d", out.ErrCode)
	}
	if out.Data["key"] != "val" {
		t.Errorf("data.key: want val, got %v", out.Data["key"])
	}
}

func TestFail(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	Fail(c, 400, 1001, "bad input")

	var out struct {
		ErrCode int    `json:"errCode"`
		ErrMsg  string `json:"errMsg"`
	}
	json.NewDecoder(w.Body).Decode(&out)
	if out.ErrCode != 1001 {
		t.Errorf("errCode: want 1001, got %d", out.ErrCode)
	}
	if out.ErrMsg != "bad input" {
		t.Errorf("errMsg: want 'bad input', got %s", out.ErrMsg)
	}
}
