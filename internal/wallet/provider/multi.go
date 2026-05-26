package provider

import (
	"context"
	"errors"

	"go.uber.org/zap"
)

type MultiProvider struct {
	providers []TxProvider
	log       *zap.Logger
}

func NewMultiProvider(providers []TxProvider, log *zap.Logger) *MultiProvider {
	return &MultiProvider{providers: providers, log: log}
}

func (m *MultiProvider) GetTransfers(ctx context.Context, req TransferRequest) ([]TxRecord, error) {
	for _, p := range m.providers {
		result, err := p.GetTransfers(ctx, req)
		if err == nil {
			return result, nil
		}
		if errors.Is(err, ErrUnsupportedChain) {
			m.log.Debug("provider skipped (unsupported chain)", zap.String("name", p.Name()), zap.String("chain", req.Chain))
		} else {
			m.log.Warn("provider failed", zap.String("name", p.Name()), zap.Error(err))
		}
	}
	return nil, errors.New("all providers exhausted")
}

func (m *MultiProvider) Name() string { return "multi" }
