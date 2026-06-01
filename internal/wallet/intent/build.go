package intent

import "github.com/web1/im-business/internal/config"

// BuildProvider constructs the intent provider from config. Only CoW today.
func BuildProvider(cfg config.IntentConfig) Provider {
	c := cfg.Providers["cow"]
	return NewCowProvider(c.APIBase)
}
