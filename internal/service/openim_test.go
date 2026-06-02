package service

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/web1/im-business/internal/config"
)

func TestGetUserToken(t *testing.T) {
	var adminCalled, userTokenCalled bool

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/auth/get_admin_token":
			adminCalled = true
			json.NewEncoder(w).Encode(map[string]any{
				"errCode": 0,
				"data":    map[string]any{"token": "admin-tok", "expireTimeSeconds": 86400},
			})
		case "/auth/get_user_token":
			userTokenCalled = true
			json.NewEncoder(w).Encode(map[string]any{
				"errCode": 0,
				"data":    map[string]any{"token": "im-tok", "expireTimeSeconds": 86400},
			})
		default:
			w.WriteHeader(404)
		}
	}))
	defer srv.Close()

	c := NewOpenIMClient(config.OpenIMConfig{
		APIURL: srv.URL, AdminUserID: "imAdmin", Secret: "openIM123",
	})
	tok, err := c.GetUserToken("user1", 2)
	if err != nil {
		t.Fatalf("GetUserToken: %v", err)
	}
	if tok != "im-tok" {
		t.Errorf("want im-tok, got %s", tok)
	}
	if !adminCalled || !userTokenCalled {
		t.Error("expected both admin and user token endpoints to be called")
	}
}

func TestAdminTokenCached(t *testing.T) {
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/auth/get_admin_token" {
			calls++
			json.NewEncoder(w).Encode(map[string]any{
				"errCode": 0,
				"data":    map[string]any{"token": "admin", "expireTimeSeconds": 86400},
			})
		} else {
			json.NewEncoder(w).Encode(map[string]any{
				"errCode": 0,
				"data":    map[string]any{"token": "im", "expireTimeSeconds": 86400},
			})
		}
	}))
	defer srv.Close()

	c := NewOpenIMClient(config.OpenIMConfig{APIURL: srv.URL, AdminUserID: "imAdmin", Secret: "s"})
	c.GetUserToken("u1", 2)
	c.GetUserToken("u2", 2)
	if calls != 1 {
		t.Errorf("admin token should be fetched once, got %d calls", calls)
	}
}

func TestSendWalletTransferMsg(t *testing.T) {
	var capturedBody map[string]any

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/auth/get_admin_token":
			json.NewEncoder(w).Encode(openimEnvelope{
				Data: json.RawMessage(`{"token":"test-tok"}`),
			})
		case "/msg/send_msg":
			json.NewDecoder(r.Body).Decode(&capturedBody)
			json.NewEncoder(w).Encode(openimEnvelope{})
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := NewOpenIMClient(config.OpenIMConfig{
		APIURL:      srv.URL,
		AdminUserID: "admin",
		Secret:      "s3cr3t",
	})

	err := c.SendWalletTransferMsg("alice", "bob", WalletTransferPayload{
		Amount: "0.1", Symbol: "ETH", Chain: "eth", Hash: "0xabc",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedBody["sendID"] != "alice" {
		t.Errorf("sendID: got %v, want alice", capturedBody["sendID"])
	}
	if capturedBody["recvID"] != "bob" {
		t.Errorf("recvID: got %v, want bob", capturedBody["recvID"])
	}
	if int(capturedBody["contentType"].(float64)) != 101 {
		t.Errorf("contentType: got %v, want 101", capturedBody["contentType"])
	}
	if int(capturedBody["sessionType"].(float64)) != 1 {
		t.Errorf("sessionType: got %v, want 1", capturedBody["sessionType"])
	}
}
