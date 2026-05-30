package quote

import (
	"go.uber.org/zap"

	"github.com/web1/im-business/internal/config"
)

// BuildAggregator constructs the quote aggregator from swap config. It mirrors
// wallet.BuildProviders: instantiate every known provider from its credentials,
// then assemble the per-chain ordered list from `quote_providers`.
//
// New providers (1inch, OKX, Paraswap, Uniswap) are added to the `all` map here
// as they are implemented.
func BuildAggregator(cfg config.SwapConfig, log *zap.Logger) *Aggregator {
	feeRecipients := make(map[string]string, len(cfg.Chains))
	for chain, ch := range cfg.Chains {
		feeRecipients[chain] = ch.FeeRecipient
	}

	all := map[string]Provider{}
	if c, ok := cfg.Providers["zerox"]; ok {
		all["zerox"] = NewZeroExProvider(c.APIKey, c.APIBase, c.Version, feeRecipients, cfg.FeeBps)
	}
	// TODO(M2+): all["oneinch"], all["okx"], all["paraswap"], all["uniswap"].

	byChain := map[string][]Provider{}
	for chain, names := range cfg.QuoteProviders {
		var ordered []Provider
		for _, name := range names {
			if p, ok := all[name]; ok {
				ordered = append(ordered, p)
			} else {
				log.Warn("quote provider configured but not implemented", zap.String("name", name), zap.String("chain", chain))
			}
		}
		if len(ordered) > 0 {
			byChain[chain] = ordered
		}
	}

	return NewAggregator(byChain, log)
}
