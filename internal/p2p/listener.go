package p2p

import (
	"context"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/web1/im-business/internal/model"
	"github.com/web1/im-business/internal/p2p/contract"
	"github.com/web1/im-business/internal/repo"
	"go.uber.org/zap"
)

const (
	bscTestnetRPC    = "https://data-seed-prebsc-1-s1.binance.org:8545"
	bscTestnetChainID = 97
	pollInterval     = 12 * time.Second // ~1 BSC block
	confirmBlocks    = 3                // wait 3 blocks before indexing
)

type Listener struct {
	rpc          string
	chainID      uint64
	contractAddr common.Address
	repo         *repo.P2PRepo
	log          *zap.Logger
	startBlock   uint64
}

func NewListener(contractAddr string, repo *repo.P2PRepo, log *zap.Logger) *Listener {
	return &Listener{
		rpc:          bscTestnetRPC,
		chainID:      bscTestnetChainID,
		contractAddr: common.HexToAddress(contractAddr),
		repo:         repo,
		log:          log,
	}
}

// Run polls for new contract events until ctx is cancelled.
func (l *Listener) Run(ctx context.Context) {
	client, err := ethclient.DialContext(ctx, l.rpc)
	if err != nil {
		l.log.Error("p2p listener: dial failed", zap.Error(err))
		return
	}
	defer client.Close()

	// Parse ABI for log decoding
	parsed, err := abi.JSON(strings.NewReader(contract.P2PEscrowMetaData.ABI))
	if err != nil {
		l.log.Error("p2p listener: parse ABI", zap.Error(err))
		return
	}

	// Start from current head minus a small buffer on first run
	if l.startBlock == 0 {
		head, err := client.BlockNumber(ctx)
		if err != nil {
			l.log.Error("p2p listener: get block number", zap.Error(err))
			return
		}
		if head > 50 {
			l.startBlock = head - 50
		}
	}

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			head, err := client.BlockNumber(ctx)
			if err != nil {
				l.log.Warn("p2p listener: block number", zap.Error(err))
				continue
			}
			toBlock := head - confirmBlocks
			if toBlock <= l.startBlock {
				continue
			}
			l.fetchLogs(ctx, client, &parsed, l.startBlock+1, toBlock)
			l.startBlock = toBlock
		}
	}
}

func (l *Listener) fetchLogs(ctx context.Context, client *ethclient.Client, parsed *abi.ABI, from, to uint64) {
	query := ethereum.FilterQuery{
		FromBlock: new(big.Int).SetUint64(from),
		ToBlock:   new(big.Int).SetUint64(to),
		Addresses: []common.Address{l.contractAddr},
	}
	logs, err := client.FilterLogs(ctx, query)
	if err != nil {
		l.log.Warn("p2p listener: filter logs", zap.Error(err))
		return
	}
	for _, vLog := range logs {
		if err := l.handleLog(ctx, client, parsed, vLog); err != nil {
			l.log.Warn("p2p listener: handle log", zap.Error(err), zap.String("tx", vLog.TxHash.Hex()))
		}
	}
}

func (l *Listener) handleLog(ctx context.Context, client *ethclient.Client, parsed *abi.ABI, vLog types.Log) error {
	if len(vLog.Topics) == 0 {
		return nil
	}
	event, err := parsed.EventByID(vLog.Topics[0])
	if err != nil {
		return nil // unknown event, skip
	}

	switch event.Name {
	case "OrderCreated":
		return l.onOrderCreated(ctx, client, parsed, vLog)
	case "OrderTaken":
		return l.onOrderTaken(ctx, parsed, vLog)
	case "PaymentMarked":
		return l.onStatusChange(ctx, parsed, vLog, model.P2PStatusPaid)
	case "OrderReleased":
		return l.onStatusChange(ctx, parsed, vLog, model.P2PStatusReleased)
	case "OrderCancelled":
		return l.onStatusChange(ctx, parsed, vLog, model.P2PStatusCancelled)
	case "DisputeRaised":
		return l.onStatusChange(ctx, parsed, vLog, model.P2PStatusDisputed)
	case "DisputeResolved":
		return l.onDisputeResolved(ctx, parsed, vLog)
	}
	return nil
}

// onOrderCreated fetches full order data from the chain and inserts it.
func (l *Listener) onOrderCreated(ctx context.Context, client *ethclient.Client, parsed *abi.ABI, vLog types.Log) error {
	data := map[string]any{}
	if err := parsed.UnpackIntoMap(data, "OrderCreated", vLog.Data); err != nil {
		return err
	}
	orderID := new(big.Int).SetBytes(vLog.Topics[1].Bytes()).Uint64()
	seller := common.HexToAddress(vLog.Topics[2].Hex()).Hex()
	token := common.BytesToAddress(vLog.Data[12:32]).Hex()
	amount := new(big.Int).SetBytes(vLog.Data[32:64])

	// Fetch full order struct from contract for fiat details
	escrow, err := contract.NewP2PEscrow(l.contractAddr, client)
	if err != nil {
		return err
	}
	o, err := escrow.Orders(nil, new(big.Int).SetUint64(orderID))
	if err != nil {
		return err
	}

	order := &model.P2POrder{
		ChainID:      l.chainID,
		ContractAddr: l.contractAddr.Hex(),
		OrderID:      orderID,
		Seller:       seller,
		Token:        token,
		Amount:       amount.String(),
		FiatAmount:   o.FiatAmount.Uint64(),
		FiatCurrency: o.FiatCurrency,
		PayMethod:    o.PayMethod,
		Status:       model.P2PStatusOpen,
		TxHash:       vLog.TxHash.Hex(),
	}
	return l.repo.Upsert(ctx, order)
}

// onOrderTaken updates buyer address + status to Locked.
func (l *Listener) onOrderTaken(ctx context.Context, parsed *abi.ABI, vLog types.Log) error {
	orderID := new(big.Int).SetBytes(vLog.Topics[1].Bytes()).Uint64()
	buyer := common.HexToAddress(vLog.Topics[2].Hex()).Hex()
	return l.repo.UpdateBuyerAndStatus(ctx, l.chainID, orderID, buyer, model.P2PStatusLocked)
}

// onStatusChange updates only the status field.
func (l *Listener) onStatusChange(ctx context.Context, parsed *abi.ABI, vLog types.Log, status uint8) error {
	orderID := new(big.Int).SetBytes(vLog.Topics[1].Bytes()).Uint64()
	return l.repo.UpdateStatus(ctx, l.chainID, orderID, status)
}

// onDisputeResolved maps buyerWins → Released or Cancelled.
func (l *Listener) onDisputeResolved(ctx context.Context, parsed *abi.ABI, vLog types.Log) error {
	orderID := new(big.Int).SetBytes(vLog.Topics[1].Bytes()).Uint64()
	// Topics[2] is buyerWins (bool, padded to 32 bytes)
	buyerWins := vLog.Topics[2] != (common.Hash{}) && vLog.Topics[2][31] == 1
	status := model.P2PStatusCancelled
	if buyerWins {
		status = model.P2PStatusReleased
	}
	return l.repo.UpdateStatus(ctx, l.chainID, orderID, status)
}
