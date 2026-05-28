package config

import (
	"strings"

	"github.com/spf13/viper"
)

type SmtpConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	From     string `mapstructure:"from"`
}

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Postgres PostgresConfig `mapstructure:"postgres"`
	Redis    RedisConfig    `mapstructure:"redis"`
	JWT      JWTConfig      `mapstructure:"jwt"`
	OpenIM   OpenIMConfig   `mapstructure:"openim"`
	Wallet   WalletConfig   `mapstructure:"wallet"`
	TOTP     TOTPConfig     `mapstructure:"totp"`
	SMTP     *SmtpConfig    `mapstructure:"smtp"` // optional; nil = dev mode (mock code)
}

type TOTPConfig struct {
	// EncryptKey is a hex string for AES-256 (64 hex chars = 32 bytes).
	// Used to encrypt totp_secret at rest in the users table.
	EncryptKey string `mapstructure:"encrypt_key"`
}

// ProviderKeyConfig holds API keys for a provider.
// Supports a single key (api_key), multiple rotating keys (api_keys), and
// optional separate keys for testnet chains (testnet_api_keys).
// api_key is kept for backward compatibility; api_keys takes precedence when set.
type ProviderKeyConfig struct {
	APIKey         string   `mapstructure:"api_key"`
	APIKeys        []string `mapstructure:"api_keys"`
	TestnetAPIKeys []string `mapstructure:"testnet_api_keys"`
}

// MainKeys returns the effective mainnet key list.
func (c ProviderKeyConfig) MainKeys() []string {
	if len(c.APIKeys) > 0 {
		return c.APIKeys
	}
	if c.APIKey != "" {
		return []string{c.APIKey}
	}
	return nil
}

// TestKeys returns testnet keys, falling back to MainKeys if not configured.
func (c ProviderKeyConfig) TestKeys() []string {
	if len(c.TestnetAPIKeys) > 0 {
		return c.TestnetAPIKeys
	}
	return c.MainKeys()
}

type WalletConfig struct {
	TxProviders       map[string][]string   `mapstructure:"tx_providers"`
	StreamProvider    string                `mapstructure:"stream_provider"`
	DisabledProviders []string              `mapstructure:"disabled_providers"`
	Moralis           MoralisProviderConfig `mapstructure:"moralis"`
	Alchemy           AlchemyConfig         `mapstructure:"alchemy"`
	Ankr              ProviderKeyConfig     `mapstructure:"ankr"`
	Covalent          ProviderKeyConfig     `mapstructure:"covalent"`
	NodeReal          ProviderKeyConfig     `mapstructure:"nodereal"`
	Infura            ProviderKeyConfig     `mapstructure:"infura"`
	GetBlock          ProviderKeyConfig     `mapstructure:"getblock"`
	QuickNode         QuickNodeConfig       `mapstructure:"quicknode"`
	Chainstack        struct {
		Endpoints map[string]string `mapstructure:"endpoints"`
	} `mapstructure:"chainstack"`
}

// AlchemyConfig holds both tx-history keys and Notify (stream) credentials.
// Endpoints takes precedence over api_key for tx-history: set per-chain endpoint URLs
// when each Alchemy App is bound to a single network (the common case).
// api_key / api_keys are the fallback for a shared key that covers all networks.
type AlchemyConfig struct {
	Endpoints         map[string]string `mapstructure:"endpoints"`           // per-chain full endpoint URLs
	APIKey            string            `mapstructure:"api_key"`             // fallback single key
	APIKeys           []string          `mapstructure:"api_keys"`            // fallback rotating keys
	TestnetAPIKeys    []string          `mapstructure:"testnet_api_keys"`
	AuthToken         string            `mapstructure:"auth_token"`          // X-Alchemy-Token for Notify management API
	WebhookSigningKey string            `mapstructure:"webhook_signing_key"` // HMAC-SHA256 key for webhook verification
	WebhookURL        string            `mapstructure:"webhook_url"`
}

func (c AlchemyConfig) MainKeys() []string {
	if len(c.APIKeys) > 0 {
		return c.APIKeys
	}
	if c.APIKey != "" {
		return []string{c.APIKey}
	}
	return nil
}

func (c AlchemyConfig) TestKeys() []string {
	if len(c.TestnetAPIKeys) > 0 {
		return c.TestnetAPIKeys
	}
	return c.MainKeys()
}

// QuickNodeConfig holds both tx-history endpoint URLs and Streams credentials.
type QuickNodeConfig struct {
	Endpoints     map[string]string `mapstructure:"endpoints"`
	APIKey        string            `mapstructure:"api_key"`        // x-api-key for Streams management API
	WebhookSecret string            `mapstructure:"webhook_secret"` // x-qn-secret header value
	WebhookURL    string            `mapstructure:"webhook_url"`
}

type MoralisProviderConfig struct {
	APIKey        string `mapstructure:"api_key"`
	WebhookSecret string `mapstructure:"webhook_secret"`
	WebhookURL    string `mapstructure:"webhook_url"`
}

type ServerConfig struct {
	Port int    `mapstructure:"port"`
	Mode string `mapstructure:"mode"`
}

type PostgresConfig struct {
	DSN string `mapstructure:"dsn"`
}

type RedisConfig struct {
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

type JWTConfig struct {
	Secret      string `mapstructure:"secret"`
	ExpireHours int    `mapstructure:"expire_hours"`
}

type OpenIMConfig struct {
	APIURL      string `mapstructure:"api_url"`
	AdminUserID string `mapstructure:"admin_user_id"`
	Secret      string `mapstructure:"secret"`
}

func Load(cfgFile string) (*Config, error) {
	viper.SetConfigFile(cfgFile)
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}
	var cfg Config
	return &cfg, viper.Unmarshal(&cfg)
}
