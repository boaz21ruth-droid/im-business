package config

// SwapConfig describes runtime config served to the wallet's Swap page.
// It is returned verbatim to the Flutter client via GET /wallet/swap_config so
// API keys, RPC URLs, fee recipients, and router whitelists can be rotated
// without shipping a new app build.
//
// Keys here mirror the JSON the Flutter app's RemoteSwapConfig DTO expects.
// See lib/services/wallet/swap/remote_swap_config.dart in im-wallet-app.
type SwapConfig struct {
	// FeeBps is the platform fee in basis points, injected into each provider's
	// own fee mechanism server-side (e.g. 0x swapFeeBps).
	FeeBps int `mapstructure:"fee_bps" json:"feeBps"`
	// Providers holds per-aggregator credentials. json:"-" — NEVER sent to the
	// client; aggregation and API keys are server-side only.
	Providers map[string]SwapProviderCreds `mapstructure:"providers" json:"-"`
	// QuoteProviders is the ordered set of aggregators to query per chain key,
	// mirroring wallet.tx_providers. json:"-" — server-side only.
	QuoteProviders map[string][]string `mapstructure:"quote_providers" json:"-"`
	Chains         map[string]SwapChain `mapstructure:"chains" json:"chains"`
	Limits         SwapLimits           `mapstructure:"limits" json:"limits"`
}

// SwapProviderCreds holds the API credentials/identifiers for one aggregator.
// Keyless aggregators (KyberSwap, Paraswap) use ClientID/Partner instead of a key.
type SwapProviderCreds struct {
	APIKey   string `mapstructure:"api_key"`
	APIBase  string `mapstructure:"api_base"`
	Version  string `mapstructure:"version"`
	ClientID string `mapstructure:"client_id"` // KyberSwap x-client-id (optional)
	Partner  string `mapstructure:"partner"`   // Paraswap partner name (optional)
}

type SwapChain struct {
	RPCs           []string `mapstructure:"rpcs"            json:"rpcs"`
	FeeRecipient   string   `mapstructure:"fee_recipient"   json:"feeRecipient"`
	AllowedRouters []string `mapstructure:"allowed_routers" json:"allowedRouters"`
}

type SwapLimits struct {
	LargeAmountUsdThreshold      float64 `mapstructure:"large_amount_usd_threshold"      json:"largeAmountUsdThreshold"`
	PriceDriftBps                int     `mapstructure:"price_drift_bps"                 json:"priceDriftBps"`
	ApproveReceiptTimeoutSeconds int     `mapstructure:"approve_receipt_timeout_seconds" json:"approveReceiptTimeoutSeconds"`
}
