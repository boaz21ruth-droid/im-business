package bridge

import (
	"github.com/web1/im-business/internal/config"
)

// BuildProvider constructs the cross-chain bridge provider from config. Only
// LI.FI is wired today; the interface allows adding others later.
func BuildProvider(cfg config.BridgeConfig) Provider {
	c := cfg.Providers["lifi"]
	return NewLiFiProvider(c.APIBase, c.APIKey, c.Integrator, cfg.FeeBps)
}
