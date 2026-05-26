package wallet

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/web1/im-business/internal/wallet/provider"
)

const historyTTL = 2 * time.Minute

type WalletCache struct {
	rdb *redis.Client
}

func NewCache(rdb *redis.Client) *WalletCache {
	return &WalletCache{rdb: rdb}
}

func historyKey(chain, address, contract string, page int) string {
	return fmt.Sprintf("wallet:history:%s:%s:%s:%d", chain, address, contract, page)
}

func (c *WalletCache) GetHistory(ctx context.Context, chain, address, contract string, page int) ([]provider.TxRecord, bool) {
	data, err := c.rdb.Get(ctx, historyKey(chain, address, contract, page)).Bytes()
	if err != nil {
		return nil, false
	}
	var records []provider.TxRecord
	if err := json.Unmarshal(data, &records); err != nil {
		return nil, false
	}
	return records, true
}

func (c *WalletCache) SetHistory(ctx context.Context, chain, address, contract string, page int, records []provider.TxRecord) {
	data, _ := json.Marshal(records)
	c.rdb.Set(ctx, historyKey(chain, address, contract, page), data, historyTTL)
}

// InvalidateAddress deletes all history cache keys for a given chain+address.
func (c *WalletCache) InvalidateAddress(ctx context.Context, chain, address string) {
	pattern := fmt.Sprintf("wallet:history:%s:%s:*", chain, address)
	var cursor uint64
	for {
		keys, next, err := c.rdb.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			break
		}
		if len(keys) > 0 {
			c.rdb.Del(ctx, keys...)
		}
		cursor = next
		if cursor == 0 {
			break
		}
	}
}

func tronCursorKey(address string) string {
	return "wallet:tron:cursor:" + address
}

func (c *WalletCache) GetTronCursor(ctx context.Context, address string) int64 {
	v, err := c.rdb.Get(ctx, tronCursorKey(address)).Int64()
	if err != nil {
		return 0
	}
	return v
}

func (c *WalletCache) SetTronCursor(ctx context.Context, address string, ts int64) {
	c.rdb.Set(ctx, tronCursorKey(address), ts, 0)
}
