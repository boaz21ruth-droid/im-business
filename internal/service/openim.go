package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/web1/im-business/internal/config"
)

type OpenIMClient struct {
	cfg        config.OpenIMConfig
	http       *http.Client
	mu         sync.Mutex
	adminToken string
	tokenExp   time.Time
}

func NewOpenIMClient(cfg config.OpenIMConfig) *OpenIMClient {
	return &OpenIMClient{cfg: cfg, http: &http.Client{Timeout: 10 * time.Second}}
}

type openimEnvelope struct {
	ErrCode int             `json:"errCode"`
	ErrMsg  string          `json:"errMsg"`
	Data    json.RawMessage `json:"data"`
}

func (c *OpenIMClient) post(path string, body any, token string) (json.RawMessage, error) {
	b, _ := json.Marshal(body)
	req, _ := http.NewRequest(http.MethodPost, c.cfg.APIURL+path, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("operationID", strconv.FormatInt(time.Now().UnixNano(), 10))
	if token != "" {
		req.Header.Set("token", token)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("openim %s: %w", path, err)
	}
	defer resp.Body.Close()

	var out openimEnvelope
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("openim decode %s: %w", path, err)
	}
	if out.ErrCode != 0 {
		return nil, fmt.Errorf("openim %s errCode=%d: %s", path, out.ErrCode, out.ErrMsg)
	}
	return out.Data, nil
}

func (c *OpenIMClient) adminTok() (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.adminToken != "" && time.Now().Before(c.tokenExp) {
		return c.adminToken, nil
	}
	data, err := c.post("/auth/get_admin_token", map[string]any{
		"secret": c.cfg.Secret, "userID": c.cfg.AdminUserID, "platformID": 1,
	}, "")
	if err != nil {
		return "", err
	}
	var out struct {
		Token string `json:"token"`
	}
	json.Unmarshal(data, &out)
	c.adminToken = out.Token
	c.tokenExp = time.Now().Add(23 * time.Hour)
	return c.adminToken, nil
}

// RegisterUser creates a user in OpenIM. Safe to call multiple times (idempotent).
func (c *OpenIMClient) RegisterUser(userID, nickname, faceURL string) error {
	tok, err := c.adminTok()
	if err != nil {
		return err
	}
	_, err = c.post("/user/user_register", map[string]any{
		"users": []map[string]any{
			{"userID": userID, "nickname": nickname, "faceURL": faceURL},
		},
	}, tok)
	return err
}

// WalletTxNotification is the payload for contentType 1400 custom messages.
type WalletTxNotification struct {
	Type      string `json:"type"`
	Chain     string `json:"chain"`
	Symbol    string `json:"symbol"`
	Amount    string `json:"amount"`
	Direction string `json:"direction"` // "sent" | "received"
	Hash      string `json:"hash"`
}

// WalletTransferPayload is the data field of a contentType=101 custom message
// sent in the sender-receiver single chat when a wallet transfer is confirmed.
type WalletTransferPayload struct {
	Amount string `json:"amount"`
	Symbol string `json:"symbol"`
	Chain  string `json:"chain"`
	Hash   string `json:"hash"`
}

// SendWalletTransferMsg posts a custom message (contentType=101, sessionType=1)
// in the single chat between fromUserID and toUserID.
// viewType=916 maps to CustomMessageType.walletTransfer in Flutter.
// OpenIM single-chat messages are visible to both participants.
func (c *OpenIMClient) SendWalletTransferMsg(fromUserID, toUserID string, p WalletTransferPayload) error {
	data := map[string]any{
		"viewType": 916,
		"data":     p,
	}
	dataBytes, _ := json.Marshal(data)
	tok, err := c.adminTok()
	if err != nil {
		return err
	}
	_, err = c.post("/msg/send_msg", map[string]any{
		"sendID":      fromUserID,
		"recvID":      toUserID,
		"contentType": 101,
		"sessionType": 1,
		"content": map[string]any{
			"data":        string(dataBytes),
			"extension":   "",
			"description": "转账通知",
		},
	}, tok)
	return err
}

// SendWalletTxNotify delivers a wallet transaction notification to a user via OpenIM.
func (c *OpenIMClient) SendWalletTxNotify(userID string, n WalletTxNotification) error {
	content, _ := json.Marshal(n)
	tok, err := c.adminTok()
	if err != nil {
		return err
	}
	_, err = c.post("/msg/send_msg", map[string]any{
		"sendID":         c.cfg.AdminUserID,
		"recvID":         userID,
		"senderNickname": "Wallet",
		"content":        map[string]any{"content": string(content)},
		"contentType":    1400,
		"sessionType":    1,
	}, tok)
	return err
}

// GetUserToken returns an IM token for an existing OpenIM user.
func (c *OpenIMClient) GetUserToken(userID string, platformID int) (string, error) {
	tok, err := c.adminTok()
	if err != nil {
		return "", err
	}
	data, err := c.post("/auth/get_user_token", map[string]any{
		"userID": userID, "platformID": platformID,
	}, tok)
	if err != nil {
		return "", err
	}
	var out struct {
		Token string `json:"token"`
	}
	json.Unmarshal(data, &out)
	return out.Token, nil
}
