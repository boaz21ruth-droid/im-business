package config

// IntentConfig holds intent-swap (order-based) aggregator config. Server-side only.
type IntentConfig struct {
	Providers map[string]IntentProviderCreds `mapstructure:"providers" json:"-"`
}

type IntentProviderCreds struct {
	APIBase string `mapstructure:"api_base"`
	AppCode string `mapstructure:"app_code"`
}
