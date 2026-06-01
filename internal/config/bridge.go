package config

// BridgeConfig describes cross-chain (bridge) aggregator config. Like SwapConfig,
// credentials are server-side only (never serialized to the client).
type BridgeConfig struct {
	// FeeBps is the platform fee in basis points for cross-chain swaps. Kept at
	// 0 until the integrator is registered with the aggregator (e.g. LI.FI).
	FeeBps int `mapstructure:"fee_bps" json:"-"`
	// Providers holds per-aggregator credentials/identifiers, keyed by name
	// (e.g. "lifi").
	Providers map[string]BridgeProviderCreds `mapstructure:"providers" json:"-"`
}

// BridgeProviderCreds holds credentials for one bridge aggregator.
type BridgeProviderCreds struct {
	APIBase    string `mapstructure:"api_base"`
	APIKey     string `mapstructure:"api_key"`
	Integrator string `mapstructure:"integrator"`
}
