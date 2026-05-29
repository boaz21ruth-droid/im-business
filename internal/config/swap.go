package config

// SwapConfig describes runtime config served to the wallet's Swap page.
// It is returned verbatim to the Flutter client via GET /wallet/swap_config so
// API keys, RPC URLs, fee recipients, and router whitelists can be rotated
// without shipping a new app build.
//
// Keys here mirror the JSON the Flutter app's RemoteSwapConfig DTO expects.
// See lib/services/wallet/swap/remote_swap_config.dart in im-wallet-app.
type SwapConfig struct {
	Providers SwapProviders         `mapstructure:"providers" json:"providers"`
	Chains    map[string]SwapChain  `mapstructure:"chains"    json:"chains"`
	Limits    SwapLimits            `mapstructure:"limits"    json:"limits"`
}

type SwapProviders struct {
	ZeroEx SwapZeroExConfig `mapstructure:"zerox" json:"zerox"`
}

type SwapZeroExConfig struct {
	APIKey  string `mapstructure:"api_key"  json:"apiKey"`
	APIBase string `mapstructure:"api_base" json:"apiBase"`
	Version string `mapstructure:"version"  json:"version"`
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
