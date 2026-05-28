package wallet

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/web1/im-business/internal/config"
	"github.com/web1/im-business/internal/model"
	"github.com/web1/im-business/internal/repo"
	"github.com/web1/im-business/internal/service"
	"github.com/web1/im-business/internal/wallet/provider"
)

// testnetFallback maps testnet chain keys to their mainnet equivalent for address lookup.
// Used when a user registered before testnet chains were added to their address map.
var testnetFallback = map[string]string{
	"bsc_testnet": "bsc",
	"eth_sepolia": "eth",
	"tron_shasta": "tron",
}

// evmChains lists EVM chain keys that support stream providers.
var evmChains = map[string]bool{
	"eth": true, "bsc": true, "polygon": true,
	"arbitrum": true, "optimism": true,
	"bsc_testnet": true, "eth_sepolia": true,
}

var tronChains = map[string]bool{
	"tron": true, "tron_shasta": true,
}

// TxResponse is the API response record returned to Flutter.
type TxResponse struct {
	Hash           string `json:"hash"`
	From           string `json:"from"`
	To             string `json:"to"`
	Value          string `json:"value"`
	Decimals       int    `json:"decimals"`
	TokenSymbol    string `json:"tokenSymbol,omitempty"`
	TokenContract  string `json:"tokenContract,omitempty"`
	BlockTimestamp string `json:"blockTimestamp"`
	ChainKey       string `json:"chainKey"`
	Direction      string `json:"direction"` // "sent" | "received"
}

type WalletService struct {
	repo            *repo.WalletRepo
	cache           *WalletCache
	openim          *service.OpenIMClient
	streamProviders []provider.StreamProvider
	wsProviders     []provider.WSProvider
	providers       map[string]*provider.MultiProvider // chain → provider chain
	tronPoller      *TronPoller
	log             *zap.Logger
}

func NewWalletService(
	repo *repo.WalletRepo,
	cache *WalletCache,
	openim *service.OpenIMClient,
	streamProviders []provider.StreamProvider,
	wsProviders []provider.WSProvider,
	providers map[string]*provider.MultiProvider,
	tronPoller *TronPoller,
	log *zap.Logger,
) *WalletService {
	return &WalletService{
		repo: repo, cache: cache, openim: openim,
		streamProviders: streamProviders, wsProviders: wsProviders,
		providers: providers, tronPoller: tronPoller, log: log,
	}
}

// RegisterAddresses upserts wallet addresses, registers stream providers, and registers Tron addresses.
// Stream registration runs in the background with retries to tolerate transient DNS failures.
func (s *WalletService) RegisterAddresses(ctx context.Context, userID string, addresses map[string]string) error {
	// Synthesize testnet addresses from mainnet equivalents for wallets created before testnet support.
	for testnet, mainnet := range testnetFallback {
		if _, ok := addresses[testnet]; !ok {
			if addr, ok := addresses[mainnet]; ok {
				addresses[testnet] = addr
			}
		}
	}

	for chainKey, address := range addresses {
		if err := s.repo.UpsertAddress(userID, chainKey, address); err != nil {
			s.log.Error("upsert address", zap.String("chain", chainKey), zap.Error(err))
			continue
		}

		if evmChains[chainKey] {
			for _, sp := range s.streamProviders {
				go s.registerStreamAddressWithRetry(sp, chainKey, address)
			}
			for _, wp := range s.wsProviders {
				wp.AddAddress(chainKey, address)
			}
		}

		if tronChains[chainKey] {
			s.tronPoller.AddAddress(userID, chainKey, address)
		}
	}
	return nil
}

func (s *WalletService) GetRegisteredAddresses(ctx context.Context, userID string) (map[string]string, error) {
	rows, err := s.repo.GetAddressesByUserID(userID)
	if err != nil {
		return nil, err
	}
	addresses := make(map[string]string, len(rows))
	for _, row := range rows {
		addresses[row.ChainKey] = row.Address
	}
	return addresses, nil
}

func (s *WalletService) registerStreamAddressWithRetry(sp provider.StreamProvider, chainKey, address string) {
	delays := []time.Duration{0, 2 * time.Second, 5 * time.Second, 15 * time.Second, 30 * time.Second}
	for i, delay := range delays {
		if delay > 0 {
			time.Sleep(delay)
		}
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		err := s.doRegisterStreamAddress(ctx, sp, chainKey, address)
		cancel()
		if err == nil {
			if i > 0 {
				s.log.Info("stream register succeeded after retry",
					zap.String("provider", sp.Name()), zap.String("chain", chainKey), zap.Int("attempt", i+1))
			}
			return
		}
		s.log.Warn("stream register failed",
			zap.String("provider", sp.Name()), zap.String("chain", chainKey), zap.Int("attempt", i+1), zap.Error(err))
	}
	s.log.Error("stream register failed after all retries",
		zap.String("provider", sp.Name()), zap.String("chain", chainKey))
}

func (s *WalletService) doRegisterStreamAddress(ctx context.Context, sp provider.StreamProvider, chainKey, address string) error {
	streamID, found, err := s.repo.GetStreamIDByProvider(sp.Name(), chainKey)
	if err != nil {
		return err
	}
	if !found {
		streamID, err = sp.EnsureStream(ctx, chainKey)
		if err != nil {
			return err
		}
		if err := s.repo.UpsertStreamByProvider(sp.Name(), chainKey, streamID); err != nil {
			return err
		}
	}

	// BulkStreamProviders require the full address list to regenerate their filter.
	if bulk, ok := sp.(provider.BulkStreamProvider); ok {
		allAddrs, err := s.repo.GetAddressesByChain(chainKey)
		if err != nil {
			return err
		}
		return bulk.SetStreamAddresses(ctx, streamID, allAddrs)
	}

	return sp.AddAddressToStream(ctx, streamID, address)
}

// GetHistory returns paginated tx history for a user/chain/contract.
// fromCache=true means all providers failed and data may be stale.
func (s *WalletService) GetHistory(ctx context.Context, userID, chainKey, contract string, page, limit int) ([]TxResponse, bool, error) {
	address, err := s.repo.FindAddressByUserChain(userID, chainKey)
	if err != nil || address == "" {
		// Fall back to mainnet address for testnet chains (handles users registered before
		// testnet chains were added to the address map).
		if fallback, ok := testnetFallback[chainKey]; ok {
			address, err = s.repo.FindAddressByUserChain(userID, fallback)
		}
		if err != nil || address == "" {
			return nil, false, fmt.Errorf("no address for chain %s", chainKey)
		}
	}

	// Cache hit
	if cached, ok := s.cache.GetHistory(ctx, chainKey, address, contract, page); ok {
		return toResponses(cached, address), false, nil
	}

	// DB query
	dbRows, err := s.repo.QueryTxs(userID, chainKey, address, contract, page, limit)
	if err != nil {
		s.log.Error("query txs", zap.Error(err))
	}

	// If DB has enough, cache and return
	if len(dbRows) >= limit {
		records := dbRowsToRecords(dbRows)
		s.cache.SetHistory(ctx, chainKey, address, contract, page, records)
		return toResponses(records, address), false, nil
	}

	// Try providers
	p, hasProvider := s.providers[chainKey]
	if hasProvider {
		req := provider.TransferRequest{
			Chain:           chainKey,
			Address:         address,
			ContractAddress: contract,
			Page:            page,
			Limit:           limit,
		}
		provRecords, provErr := p.GetTransfers(ctx, req)
		if provErr == nil && len(provRecords) > 0 {
			for _, r := range provRecords {
				tx := recordToModel(r, userID, address)
				s.repo.InsertTx(tx) //nolint: all errors logged elsewhere
			}
			// Re-query merged data
			dbRows, _ = s.repo.QueryTxs(userID, chainKey, address, contract, page, limit)
			records := dbRowsToRecords(dbRows)
			s.cache.SetHistory(ctx, chainKey, address, contract, page, records)
			return toResponses(records, address), false, nil
		}
		if provErr != nil {
			s.log.Warn("all providers failed", zap.String("chain", chainKey), zap.Error(provErr))
		}
	}

	// Return whatever DB has (may be stale or empty)
	records := dbRowsToRecords(dbRows)
	return toResponses(records, address), true, nil
}

// ProcessWebhookTransfer stores an incoming webhook transfer and notifies the affected user.
func (s *WalletService) ProcessWebhookTransfer(ctx context.Context, chainKey string, t provider.WebhookTransfer) {
	userIDs, err := s.repo.FindUsersByAddress(t.From)
	if err == nil {
		toIDs, _ := s.repo.FindUsersByAddress(t.To)
		userIDs = append(userIDs, toIDs...)
	}

	seen := make(map[string]bool)
	for _, userID := range userIDs {
		if seen[userID] {
			continue
		}
		seen[userID] = true

		address, _ := s.repo.FindAddressByUserChain(userID, chainKey)

		sym := t.TokenSymbol
		addr := t.TokenAddress
		tx := &model.WalletTx{
			UserID:        userID,
			ChainKey:      chainKey,
			Address:       address,
			TxHash:        t.TxHash,
			FromAddress:   t.From,
			ToAddress:     t.To,
			Value:         t.Value,
			Decimals:      t.Decimals,
			TokenSymbol:   &sym,
			TokenContract: &addr,
			Source:        "webhook",
		}
		if err := s.repo.InsertTx(tx); err != nil {
			s.log.Error("webhook insert tx", zap.Error(err))
		}

		s.cache.InvalidateAddress(ctx, chainKey, address)

		direction := "received"
		if t.From == address {
			direction = "sent"
		}
		notif := service.WalletTxNotification{
			Type:      "wallet_tx",
			Chain:     chainKey,
			Symbol:    sym,
			Amount:    formatAmount(t.Value, t.Decimals),
			Direction: direction,
			Hash:      t.TxHash,
		}
		if err := s.openim.SendWalletTxNotify(userID, notif); err != nil {
			s.log.Warn("openim notify", zap.Error(err))
		}
	}
}

func dbRowsToRecords(rows []model.WalletTx) []provider.TxRecord {
	records := make([]provider.TxRecord, 0, len(rows))
	for _, row := range rows {
		r := provider.TxRecord{
			Hash:        row.TxHash,
			From:        row.FromAddress,
			To:          row.ToAddress,
			Value:       row.Value,
			Decimals:    row.Decimals,
			BlockNumber: row.BlockNumber,
			ChainKey:    row.ChainKey,
			Source:      row.Source,
		}
		if row.BlockTimestamp != nil {
			r.BlockTimestamp = *row.BlockTimestamp
		}
		if row.TokenSymbol != nil {
			r.TokenSymbol = row.TokenSymbol
		}
		if row.TokenContract != nil {
			r.TokenContract = row.TokenContract
		}
		records = append(records, r)
	}
	return records
}

func toResponses(records []provider.TxRecord, userAddress string) []TxResponse {
	resp := make([]TxResponse, 0, len(records))
	for _, r := range records {
		direction := "received"
		if eqAddr(r.From, userAddress) {
			direction = "sent"
		}
		tr := TxResponse{
			Hash:           r.Hash,
			From:           r.From,
			To:             r.To,
			Value:          r.Value,
			Decimals:       r.Decimals,
			BlockTimestamp: r.BlockTimestamp.UTC().Format("2006-01-02T15:04:05Z"),
			ChainKey:       r.ChainKey,
			Direction:      direction,
		}
		if r.TokenSymbol != nil {
			tr.TokenSymbol = *r.TokenSymbol
		}
		if r.TokenContract != nil {
			tr.TokenContract = *r.TokenContract
		}
		resp = append(resp, tr)
	}
	return resp
}

func eqAddr(a, b string) bool {
	if len(a) == 0 || len(b) == 0 {
		return false
	}
	al := len(a)
	bl := len(b)
	// Case-insensitive for EVM, exact for Tron
	if al == bl {
		for i := 0; i < al; i++ {
			ca, cb := a[i], b[i]
			if ca >= 'A' && ca <= 'Z' {
				ca += 32
			}
			if cb >= 'A' && cb <= 'Z' {
				cb += 32
			}
			if ca != cb {
				return false
			}
		}
		return true
	}
	return false
}

// BuildStreamProviders constructs the list of active stream providers from config.
// A provider is included only if its required credential is configured.
func BuildStreamProviders(cfg config.WalletConfig, log *zap.Logger) []provider.StreamProvider {
	var result []provider.StreamProvider

	if cfg.Moralis.APIKey != "" {
		result = append(result, provider.NewMoralisProvider(
			cfg.Moralis.APIKey, cfg.Moralis.WebhookSecret, cfg.Moralis.WebhookURL))
		log.Info("stream provider enabled", zap.String("name", "moralis"))
	}

	if cfg.Alchemy.AuthToken != "" {
		result = append(result, provider.NewAlchemyStreamProvider(
			cfg.Alchemy.AuthToken, cfg.Alchemy.WebhookSigningKey, cfg.Alchemy.WebhookURL))
		log.Info("stream provider enabled", zap.String("name", "alchemy"))
	}

	if cfg.QuickNode.APIKey != "" {
		// Derive supported chains from configured endpoint URLs.
		chains := make([]string, 0, len(cfg.QuickNode.Endpoints))
		for chain := range cfg.QuickNode.Endpoints {
			chains = append(chains, chain)
		}
		result = append(result, provider.NewQuickNodeStreamProvider(
			cfg.QuickNode.APIKey, cfg.QuickNode.WebhookSecret, cfg.QuickNode.WebhookURL, chains))
		log.Info("stream provider enabled", zap.String("name", "quicknode"))
	}

	return result
}

// BuildWSProviders constructs outbound WebSocket stream providers from config.
func BuildWSProviders(cfg config.WalletConfig, log *zap.Logger) []provider.WSProvider {
	var result []provider.WSProvider

	if len(cfg.Alchemy.MainKeys()) > 0 {
		result = append(result, provider.NewAlchemyWSProvider(cfg.Alchemy.MainKeys(), cfg.Alchemy.Endpoints, log))
		log.Info("ws provider enabled", zap.String("name", "alchemy_ws"))
	}

	return result
}

// StartWSProviders starts all WSProvider goroutines and returns when ctx is done.
func (s *WalletService) StartWSProviders(ctx context.Context) {
	for _, wp := range s.wsProviders {
		go wp.Start(ctx, func(chainKey string, t provider.WebhookTransfer) {
			s.ProcessWebhookTransfer(context.Background(), chainKey, t)
		})
	}
}

// BuildProviders constructs per-chain MultiProvider from config.
func BuildProviders(cfg config.WalletConfig, log *zap.Logger) map[string]*provider.MultiProvider {
	disabled := make(map[string]bool, len(cfg.DisabledProviders))
	for _, name := range cfg.DisabledProviders {
		disabled[name] = true
	}

	allProviders := map[string]provider.TxProvider{
		"alchemy":    provider.NewAlchemyProvider(cfg.Alchemy.MainKeys(), cfg.Alchemy.TestKeys(), cfg.Alchemy.Endpoints),
		"ankr":       provider.NewAnkrProvider(cfg.Ankr.MainKeys(), cfg.Ankr.TestKeys()),
		"moralis":    provider.NewMoralisProvider(cfg.Moralis.APIKey, cfg.Moralis.WebhookSecret, cfg.Moralis.WebhookURL),
		"covalent":   provider.NewCovalentProvider(cfg.Covalent.MainKeys(), cfg.Covalent.TestKeys()),
		"trongrid":   provider.NewTronGridProvider(),
		"nodereal":   provider.NewNodeRealProvider(cfg.NodeReal.MainKeys(), cfg.NodeReal.TestKeys()),
		"infura":     provider.NewInfuraProvider(cfg.Infura.MainKeys(), cfg.Infura.TestKeys()),
		"quicknode":  provider.NewQuickNodeProvider(cfg.QuickNode.Endpoints),
		"getblock":   provider.NewGetBlockProvider(cfg.GetBlock.MainKeys(), cfg.GetBlock.TestKeys()),
		"chainstack": provider.NewChainstackProvider(cfg.Chainstack.Endpoints),
	}

	result := make(map[string]*provider.MultiProvider)
	for chain, names := range cfg.TxProviders {
		var ordered []provider.TxProvider
		for _, name := range names {
			if disabled[name] {
				log.Info("provider disabled by config", zap.String("name", name))
				continue
			}
			if p, ok := allProviders[name]; ok {
				ordered = append(ordered, p)
			}
		}
		if len(ordered) > 0 {
			result[chain] = provider.NewMultiProvider(ordered, log)
		}
	}
	return result
}
