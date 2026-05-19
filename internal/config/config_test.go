package config

import (
	"os"
	"testing"
)

func TestLoad(t *testing.T) {
	content := `
server:
  port: 10008
  mode: debug
postgres:
  dsn: "host=localhost user=test password=test dbname=test port=5432 sslmode=disable"
redis:
  addr: "localhost:6379"
  password: ""
  db: 5
jwt:
  secret: "test-secret"
  expire_hours: 168
openim:
  api_url: "http://localhost:10002"
  admin_user_id: "imAdmin"
  secret: "openIM123"
`
	f, _ := os.CreateTemp("", "cfg*.yaml")
	f.WriteString(content)
	f.Close()
	defer os.Remove(f.Name())

	cfg, err := Load(f.Name())
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Server.Port != 10008 {
		t.Errorf("port: want 10008, got %d", cfg.Server.Port)
	}
	if cfg.OpenIM.Secret != "openIM123" {
		t.Errorf("openim secret: want openIM123, got %s", cfg.OpenIM.Secret)
	}
}
