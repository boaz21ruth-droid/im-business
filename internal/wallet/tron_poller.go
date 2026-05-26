package wallet

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/web1/im-business/internal/model"
	"github.com/web1/im-business/internal/repo"
	"github.com/web1/im-business/internal/service"
	"github.com/web1/im-business/internal/wallet/provider"
)

const (
	pollInterval    = 30 * time.Second
	maxFailures     = 5
	pauseDuration   = 10 * time.Minute
)

type addressEntry struct {
	userID   string
	chainKey string
	address  string
}

type TronPoller struct {
	mu          sync.RWMutex
	addresses   map[string]addressEntry // address → entry
	failCount   map[string]int
	pauseUntil  map[string]time.Time

	addCh    chan addressEntry
	repo     *repo.WalletRepo
	cache    *WalletCache
	openim   *service.OpenIMClient
	trongrid *provider.TronGridProvider
	log      *zap.Logger
}

func NewTronPoller(repo *repo.WalletRepo, cache *WalletCache, openim *service.OpenIMClient, trongrid *provider.TronGridProvider, log *zap.Logger) *TronPoller {
	return &TronPoller{
		addresses:  make(map[string]addressEntry),
		failCount:  make(map[string]int),
		pauseUntil: make(map[string]time.Time),
		addCh:      make(chan addressEntry, 64),
		repo:       repo,
		cache:      cache,
		openim:     openim,
		trongrid:   trongrid,
		log:        log,
	}
}

func (p *TronPoller) AddAddress(userID, chainKey, address string) {
	p.addCh <- addressEntry{userID: userID, chainKey: chainKey, address: address}
}

func (p *TronPoller) Start(ctx context.Context) {
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case entry := <-p.addCh:
			p.mu.Lock()
			p.addresses[entry.address] = entry
			p.mu.Unlock()
		case <-ticker.C:
			p.mu.RLock()
			entries := make([]addressEntry, 0, len(p.addresses))
			for _, v := range p.addresses {
				entries = append(entries, v)
			}
			p.mu.RUnlock()

			for _, entry := range entries {
				if t, ok := p.pauseUntil[entry.address]; ok && time.Now().Before(t) {
					continue
				}
				p.pollOne(ctx, entry)
			}
		}
	}
}

func (p *TronPoller) pollOne(ctx context.Context, entry addressEntry) {
	cursor := p.cache.GetTronCursor(ctx, entry.address)

	req := provider.TransferRequest{
		Chain:        entry.chainKey,
		Address:      entry.address,
		Limit:        50,
		MinTimestamp: cursor + 1,
	}

	records, err := p.trongrid.GetTransfers(ctx, req)
	if err != nil {
		p.failCount[entry.address]++
		p.log.Warn("tron poll failed", zap.String("address", entry.address), zap.Error(err))
		if p.failCount[entry.address] >= maxFailures {
			p.pauseUntil[entry.address] = time.Now().Add(pauseDuration)
			p.failCount[entry.address] = 0
			p.log.Warn("tron polling paused", zap.String("address", entry.address),
				zap.Duration("for", pauseDuration))
		}
		return
	}
	p.failCount[entry.address] = 0

	var maxTS int64
	for _, r := range records {
		ts := r.BlockTimestamp.UnixMilli()
		if ts > maxTS {
			maxTS = ts
		}

		direction := "received"
		if r.From == entry.address {
			direction = "sent"
		}

		tx := recordToModel(r, entry.userID, entry.address)
		if err := p.repo.InsertTx(tx); err != nil {
			p.log.Error("insert tron tx", zap.Error(err))
			continue
		}

		p.cache.InvalidateAddress(ctx, entry.chainKey, entry.address)

		amount := formatAmount(r.Value, r.Decimals)
		sym := ""
		if r.TokenSymbol != nil {
			sym = *r.TokenSymbol
		}
		notif := service.WalletTxNotification{
			Type:      "wallet_tx",
			Chain:     entry.chainKey,
			Symbol:    sym,
			Amount:    amount,
			Direction: direction,
			Hash:      r.Hash,
		}
		if err := p.openim.SendWalletTxNotify(entry.userID, notif); err != nil {
			p.log.Warn("openim notify failed", zap.Error(err))
		}
	}

	if maxTS > cursor {
		p.cache.SetTronCursor(ctx, entry.address, maxTS)
	}
}

// recordToModel converts a provider TxRecord to a model.WalletTx for DB storage.
func recordToModel(r provider.TxRecord, userID, address string) *model.WalletTx {
	ts := r.BlockTimestamp
	tx := &model.WalletTx{
		UserID:         userID,
		ChainKey:       r.ChainKey,
		Address:        address,
		TxHash:         r.Hash,
		FromAddress:    r.From,
		ToAddress:      r.To,
		Value:          r.Value,
		Decimals:       r.Decimals,
		TokenSymbol:    r.TokenSymbol,
		TokenContract:  r.TokenContract,
		BlockNumber:    r.BlockNumber,
		BlockTimestamp: &ts,
		Source:         r.Source,
	}
	return tx
}

// formatAmount converts a raw BigInt string to a human-readable decimal amount.
func formatAmount(rawValue string, decimals int) string {
	if rawValue == "" || rawValue == "0" {
		return "0"
	}
	if decimals <= 0 {
		return rawValue
	}
	if len(rawValue) <= decimals {
		return fmt.Sprintf("0.%0*s", decimals, rawValue)
	}
	intPart := rawValue[:len(rawValue)-decimals]
	fracPart := rawValue[len(rawValue)-decimals:]
	// Trim trailing zeros
	trimmed := fracPart
	for len(trimmed) > 2 && trimmed[len(trimmed)-1] == '0' {
		trimmed = trimmed[:len(trimmed)-1]
	}
	return intPart + "." + trimmed
}
