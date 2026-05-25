# Wallet Transaction History — Backend Service Design

## Context

The Flutter wallet app (im-wallet-app) needs near real-time transaction history per asset.
Client-side approaches (direct RPC `eth_getLogs`, Explorer APIs without key) are unreliable:
public nodes cap block ranges, testnet Explorer APIs reject keyless requests.

This spec defines a backend service in **im-business** that:
- Receives new transactions in near real-time via provider webhooks (~5–15s after on-chain confirmation)
- Serves history queries from DB/cache (millisecond response)
- Notifies the Flutter app via existing OpenIM channel
- Supports EVM chains (ETH, BSC, Polygon, Arbitrum, Optimism + testnets) and Tron

---

## Architecture Overview

```
Flutter App
  │
  ├─ POST /wallet/addresses      Register addresses on first wallet unlock
  └─ GET  /wallet/tx-history     Load history page (chain + contract + page)

im-business
  ├─ WalletAddressService        Saves addresses, registers Moralis Streams
  ├─ WalletTxHandler             GET /tx-history: Redis → DB → Provider fallback
  ├─ MoralisWebhookHandler       POST /webhook/moralis: verify → store → notify
  ├─ TronPoller (goroutine)      Polls TronGrid every 30s (Moralis has no Tron support)
  ├─ MultiTxProvider             Priority-ordered provider chain per chain key
  ├─ WalletTxRepository          PostgreSQL reads/writes
  ├─ WalletTxCache               Redis cache (TTL 2min, invalidated on webhook)
  └─ WalletNotifyService         Sends OpenIM custom message on new tx

External
  ├─ Moralis Streams             EVM webhook delivery + history API
  ├─ Alchemy                     EVM history API (ETH/Polygon/Arbitrum/Optimism)
  ├─ Ankr Advanced API           EVM history API (all chains incl. BSC)
  ├─ Covalent                    EVM history API fallback
  └─ TronGrid                    Tron history polling
```

**New transaction flow:**
```
User sends USDT on BSC
  → Moralis detects on-chain confirmation (~5–15s)
  → POST /webhook/moralis to im-business
  → Verify HMAC signature
  → Write to wallet_transactions
  → Invalidate Redis cache for that address
  → Send OpenIM custom message to user
  → App receives message → refreshes history page (reads DB, <5ms)
```

**History page load flow:**
```
App opens history page
  → GET /wallet/tx-history?chain=bsc&contract=0x337...&page=1
  → Redis hit → return immediately
  → Redis miss → query DB
  → DB result < limit → call Provider (Ankr → Moralis → Covalent)
  → Write new records to DB + Redis
  → Return to app
```

---

## Provider Abstraction Layer

### Interfaces

```go
// internal/wallet/provider/interface.go

type TransferRequest struct {
    Chain           string
    Address         string
    ContractAddress string  // empty = native transfers
    Page            int
    Limit           int
}

type TxProvider interface {
    GetTransfers(ctx context.Context, req TransferRequest) ([]TxRecord, error)
    Name() string
}

type StreamProvider interface {
    RegisterAddress(ctx context.Context, chain, address string) error
    UnregisterAddress(ctx context.Context, chain, address string) error
    VerifyWebhook(r *http.Request, body []byte) bool
}
```

### Provider Priority Per Chain (config-driven)

```yaml
# config/config.example.yaml
wallet:
  tx_providers:
    eth:          [alchemy, ankr, covalent, moralis]
    bsc:          [ankr, covalent, moralis]
    polygon:      [alchemy, ankr, covalent, moralis]
    arbitrum:     [alchemy, ankr, covalent, moralis]
    optimism:     [alchemy, ankr, covalent, moralis]
    bsc_testnet:  [ankr, moralis]
    eth_sepolia:  [alchemy, moralis]
    tron:         [trongrid]
    tron_shasta:  [trongrid]
  stream_provider: moralis
  moralis:
    api_key: "YOUR_MORALIS_API_KEY"
    webhook_secret: "YOUR_MORALIS_WEBHOOK_SECRET"
  alchemy:
    api_key: "YOUR_ALCHEMY_API_KEY"
  ankr:
    api_key: ""         # optional; works without key, higher quota with key
  covalent:
    api_key: "YOUR_COVALENT_API_KEY"
```

`MultiTxProvider` tries providers in order; on error logs and advances to next.
If all fail, returns whatever is currently in DB (stale data is better than an error).

### Free Tier Summary

| Provider | ETH | BSC | Polygon | Arbitrum | Optimism | Webhook | Free Quota |
|---|---|---|---|---|---|---|---|
| Alchemy  | ✓ | ✗ | ✓ | ✓ | ✓ | ✓ | 300M CU/month |
| Ankr     | ✓ | ✓ | ✓ | ✓ | ✓ | ✗ | 30M req/day |
| Covalent | ✓ | ✓ | ✓ | ✓ | ✓ | ✗ | 4 req/s |
| Moralis  | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | 40k CU/day |

Strategy: preserve Moralis quota for webhooks; use Alchemy + Ankr for history queries.

---

## Data Model

### PostgreSQL

```sql
-- Watched addresses per user
CREATE TABLE wallet_addresses (
    id         BIGSERIAL    PRIMARY KEY,
    user_id    VARCHAR(64)  NOT NULL,
    chain_key  VARCHAR(32)  NOT NULL,
    address    VARCHAR(128) NOT NULL,
    created_at TIMESTAMPTZ  DEFAULT NOW(),
    UNIQUE(user_id, chain_key)
);

-- Transaction records (webhook-pushed + history-fetched)
CREATE TABLE wallet_transactions (
    id              BIGSERIAL    PRIMARY KEY,
    user_id         VARCHAR(64)  NOT NULL,
    chain_key       VARCHAR(32)  NOT NULL,
    address         VARCHAR(128) NOT NULL,  -- user's address (from or to)
    tx_hash         VARCHAR(128) NOT NULL,
    from_address    VARCHAR(128) NOT NULL,
    to_address      VARCHAR(128) NOT NULL,
    value           VARCHAR(64)  NOT NULL,  -- raw BigInt string; no float
    decimals        INT          NOT NULL,
    token_symbol    VARCHAR(32),            -- NULL = native transfer
    token_contract  VARCHAR(128),           -- NULL = native transfer
    block_number    BIGINT,
    block_timestamp TIMESTAMPTZ,
    source          VARCHAR(32),            -- 'webhook','alchemy','ankr',… (debug)
    created_at      TIMESTAMPTZ  DEFAULT NOW(),
    UNIQUE(chain_key, tx_hash, address)     -- same tx stored once per involved address
);
CREATE INDEX idx_wtx ON wallet_transactions(user_id, chain_key, address, block_timestamp DESC);

-- One Moralis Stream per EVM chain (manages webhook lifecycle)
CREATE TABLE wallet_moralis_streams (
    chain_key  VARCHAR(32)  PRIMARY KEY,
    stream_id  VARCHAR(128) NOT NULL,
    created_at TIMESTAMPTZ  DEFAULT NOW()
);
```

### Redis

```
wallet:history:{chain_key}:{address}:{page}   TTL 2min    History page cache
wallet:tron:cursor:{address}                  no TTL      Last processed block_timestamp (ms)
```

---

## API Endpoints

### `POST /wallet/addresses`
Authenticated (chatToken). Called by app on first wallet unlock.

**Request:**
```json
{
  "addresses": {
    "eth":      "0xABC...",
    "bsc":      "0xABC...",
    "polygon":  "0xABC...",
    "arbitrum": "0xABC...",
    "optimism": "0xABC...",
    "tron":     "TABC..."
  }
}
```

**Backend actions:**
1. Upsert rows in `wallet_addresses`
2. For each EVM chain: add address to Moralis Stream for that chain (create stream if first time)
3. Send Tron address to TronPoller via channel

**Response:** `{ "errCode": 0 }`

---

### `GET /wallet/tx-history`
Authenticated (chatToken).

**Query params:**

| Param | Required | Description |
|---|---|---|
| `chain` | yes | e.g. `bsc`, `eth` |
| `contract` | no | Token contract address. Omit for native transfers |
| `page` | no | Default 1 |
| `limit` | no | Default 20, max 50 |

**Response:**
```json
{
  "errCode": 0,
  "data": {
    "records": [
      {
        "hash":           "0x...",
        "from":           "0x...",
        "to":             "0x...",
        "value":          "10000000",
        "decimals":       6,
        "tokenSymbol":    "USDT",
        "tokenContract":  "0x337...",
        "blockTimestamp": "2026-05-25T10:00:00Z",
        "chainKey":       "bsc",
        "direction":      "received"
      }
    ],
    "hasMore": true,
    "fromCache": false
  }
}
```

`direction` is computed server-side from `address` vs `from`/`to` — Flutter doesn't need to know the user's address to render sent/received.

`fromCache: true` signals the app that all providers failed and data may be stale.

---

### `POST /webhook/moralis`
No auth header. Protected by HMAC-SHA3 signature verification.

**Processing:**
1. Read raw body; verify `x-signature` header against `HMAC-SHA3(body, moralis_webhook_secret)`
2. Return 401 immediately if invalid
3. Parse `erc20Transfers` array from Moralis payload
4. For each transfer: look up `wallet_addresses` by `from` and `to` address
5. For each matched user: `INSERT … ON CONFLICT DO NOTHING` into `wallet_transactions`
6. `DEL wallet:history:{chain_key}:{address}:*` from Redis
7. Call `WalletNotifyService.Notify(userID, tx)`
8. Return 200 (Moralis retries on non-2xx)

---

## Tron Poller

Goroutine started at server boot. Polls TronGrid every 30 seconds for all registered Tron addresses.

**Cursor strategy:** Redis key `wallet:tron:cursor:{address}` stores the `block_timestamp` (ms) of the last processed transaction. Each poll fetches only records newer than the cursor, preventing duplicate writes and notifications.

**Endpoints polled per address:**
```
TRX:   GET /v1/accounts/{addr}/transactions?limit=20&only_confirmed=true&min_timestamp={cursor}
TRC20: GET /v1/accounts/{addr}/transactions/trc20?limit=20&only_confirmed=true&min_timestamp={cursor}
```

**Dynamic address registration:** New Tron addresses are sent to the poller via a buffered channel, no restart required.

**Failure handling:** Single-address failure is logged and skipped. After 5 consecutive failures for one address, it is paused for 10 minutes before resuming.

---

## OpenIM Notification

On new transaction (webhook or Tron poll):

```json
{
  "contentType": 1400,
  "content": {
    "type":      "wallet_tx",
    "chain":     "bsc",
    "symbol":    "USDT",
    "amount":    "10.00",
    "direction": "received",
    "hash":      "0x..."
  }
}
```

Sent via existing OpenIM admin API (`/msg/send_msg`). Flutter's OpenIM SDK receives this as a custom message, triggers a history page refresh if the user is currently viewing it, or shows a banner notification otherwise.

---

## Error Handling

| Scenario | Handling |
|---|---|
| Provider quota exceeded / timeout | Try next provider in chain; if all fail, return DB data with `fromCache: true` |
| Webhook HMAC invalid | Return 401, log — never process |
| Duplicate webhook delivery | `ON CONFLICT DO NOTHING` — fully idempotent |
| DB write failure | Log error; do not fail the HTTP response; still attempt OpenIM notify |
| Redis unavailable | Bypass cache, query DB/provider directly — transparent to user |
| Tron Poller single failure | Log, skip, retry next cycle |
| Tron Poller 5 consecutive failures | Pause address 10 minutes, log alert |
| OpenIM notify failure | Log only — notification is best-effort, not critical |

**Principle:** History queries degrade gracefully (Provider → DB → empty list with `fromCache: true`). No external service failure causes a 5xx to the client.

---

## File Structure (new files in im-business)

```
internal/
└── wallet/
    ├── handler.go              POST /wallet/addresses, GET /wallet/tx-history
    ├── webhook_handler.go      POST /webhook/moralis
    ├── service.go              Address registration + history orchestration
    ├── notify_service.go       OpenIM notification
    ├── tron_poller.go          30s Tron polling goroutine
    ├── repository.go           PostgreSQL read/write
    ├── cache.go                Redis get/set/invalidate
    ├── model.go                TxRecord, TransferRequest structs
    └── provider/
        ├── interface.go
        ├── multi.go            Priority fallback chain
        ├── alchemy.go
        ├── ankr.go
        ├── moralis.go
        ├── covalent.go
        └── trongrid.go
```

Routes added in `internal/handler/router.go`:
```
POST /wallet/addresses       → wallet.Handler (auth middleware)
GET  /wallet/tx-history      → wallet.Handler (auth middleware)
POST /webhook/moralis        → wallet.WebhookHandler (no auth, HMAC verified internally)
```

---

## Out of Scope

- Push notifications via APNs/FCM (OpenIM channel is sufficient for now)
- Pagination beyond page 10 (history older than ~200 records fetched on demand)
- NFT transfers
- Cross-chain aggregated view (per-chain queries only)
