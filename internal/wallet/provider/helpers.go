package provider

import (
	"math/big"
	"strings"
	"sync/atomic"
	"time"
)

func hexToBigDecStr(h string) string {
	n := new(big.Int)
	n.SetString(strings.TrimPrefix(strings.ToLower(h), "0x"), 16)
	return n.String()
}

func hexToInt64(h string) int64 {
	n := new(big.Int)
	n.SetString(strings.TrimPrefix(strings.ToLower(h), "0x"), 16)
	return n.Int64()
}

func unixSecToTime(sec int64) time.Time {
	return time.Unix(sec, 0).UTC()
}

// keyPool holds a set of API keys and returns them in round-robin order.
// Safe for concurrent use.
type keyPool struct {
	keys    []string
	counter atomic.Uint32
}

func newKeyPool(keys []string) *keyPool {
	return &keyPool{keys: keys}
}

// pick returns the next key in round-robin order, or "" if empty.
func (p *keyPool) pick() string {
	if len(p.keys) == 0 {
		return ""
	}
	return p.keys[p.counter.Add(1)%uint32(len(p.keys))]
}

func (p *keyPool) empty() bool { return len(p.keys) == 0 }

// isTestnet marks chain keys that correspond to test networks.
// Providers that support separate testnet keys use this to pick the right pool.
var isTestnet = map[string]bool{
	"bsc_testnet": true,
	"eth_sepolia": true,
	"tron_shasta": true,
}
