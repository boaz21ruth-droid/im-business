package config

import (
	"os"
	"testing"
)

func TestEnvOverride(t *testing.T) {
	yaml := `
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
wallet:
  moralis:
    webhook_url: "http://localhost:10008/webhook/moralis"
`
	f, _ := os.CreateTemp("", "cfg*.yaml")
	f.WriteString(yaml)
	f.Close()
	defer os.Remove(f.Name())

	t.Setenv("SERVER_MODE", "release")
	t.Setenv("POSTGRES_DSN", "host=postgres-business port=5432")
	t.Setenv("REDIS_ADDR", "redis:6379")
	t.Setenv("OPENIM_API_URL", "http://openim-server:10002")
	t.Setenv("WALLET_MORALIS_WEBHOOK_URL", "")

	cfg, err := Load(f.Name())
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if cfg.Server.Mode != "release" {
		t.Errorf("server.mode = %q, want release", cfg.Server.Mode)
	}
	if cfg.Postgres.DSN != "host=postgres-business port=5432" {
		t.Errorf("postgres.dsn = %q, want overridden", cfg.Postgres.DSN)
	}
	if cfg.Redis.Addr != "redis:6379" {
		t.Errorf("redis.addr = %q, want redis:6379", cfg.Redis.Addr)
	}
	if cfg.OpenIM.APIURL != "http://openim-server:10002" {
		t.Errorf("openim.api_url = %q, want overridden", cfg.OpenIM.APIURL)
	}
}
