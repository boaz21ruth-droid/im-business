package handler

import (
	"math/big"
	"net/http"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gin-gonic/gin"
	"github.com/web1/im-business/internal/model"
	"github.com/web1/im-business/internal/p2p/contract"
	"github.com/web1/im-business/internal/repo"
	"github.com/web1/im-business/pkg/resp"
	"go.uber.org/zap"
)

const (
	bscTestnetRPC     = "https://data-seed-prebsc-1-s1.binance.org:8545"
	bscTestnetChainID = 97
)

type P2PHandler struct {
	repo         *repo.P2PRepo
	contractAddr string
	arbitratorKey string // hex private key for resolveDispute
	log          *zap.Logger
}

func NewP2PHandler(repo *repo.P2PRepo, contractAddr, arbitratorKey string, log *zap.Logger) *P2PHandler {
	return &P2PHandler{repo: repo, contractAddr: contractAddr, arbitratorKey: arbitratorKey, log: log}
}

// GET /p2p/orders?limit=20&offset=0
func (h *P2PHandler) ListOrders(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	if limit > 100 {
		limit = 100
	}
	orders, err := h.repo.ListOpen(c.Request.Context(), limit, offset)
	if err != nil {
		resp.ErrInternal(c, "db error")
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": toOrderDTOs(orders)})
}

// GET /p2p/orders/:id
func (h *P2PHandler) GetOrder(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "invalid id"})
		return
	}
	o, err := h.repo.GetByOnchainID(c.Request.Context(), bscTestnetChainID, id)
	if err != nil {
		resp.ErrInternal(c, "db error")
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": toOrderDTO(*o)})
}

// GET /p2p/my/selling?address=0x...
func (h *P2PHandler) MySelling(c *gin.Context) {
	addr := strings.ToLower(c.Query("address"))
	if addr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "address required"})
		return
	}
	orders, err := h.repo.ListBySeller(c.Request.Context(), addr, 50, 0)
	if err != nil {
		resp.ErrInternal(c, "db error")
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": toOrderDTOs(orders)})
}

// GET /p2p/my/buying?address=0x...
func (h *P2PHandler) MyBuying(c *gin.Context) {
	addr := strings.ToLower(c.Query("address"))
	if addr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "address required"})
		return
	}
	orders, err := h.repo.ListByBuyer(c.Request.Context(), addr, 50, 0)
	if err != nil {
		resp.ErrInternal(c, "db error")
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": toOrderDTOs(orders)})
}

// GET /p2p/admin/disputes  (admin only)
func (h *P2PHandler) ListDisputed(c *gin.Context) {
	orders, err := h.repo.ListDisputed(c.Request.Context())
	if err != nil {
		resp.ErrInternal(c, "db error")
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": toOrderDTOs(orders)})
}

// POST /p2p/admin/resolve  (admin only)
// Body: {"order_id": 1, "buyer_wins": true}
func (h *P2PHandler) ResolveDispute(c *gin.Context) {
	var req struct {
		OrderID   uint64 `json:"order_id" binding:"required"`
		BuyerWins bool   `json:"buyer_wins"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": err.Error()})
		return
	}

	client, err := ethclient.DialContext(c.Request.Context(), bscTestnetRPC)
	if err != nil {
		h.log.Error("p2p resolve: dial", zap.Error(err))
		resp.ErrInternal(c, "db error")
		return
	}
	defer client.Close()

	privKey, err := crypto.HexToECDSA(h.arbitratorKey)
	if err != nil {
		h.log.Error("p2p resolve: parse key", zap.Error(err))
		resp.ErrInternal(c, "db error")
		return
	}

	auth, err := bind.NewKeyedTransactorWithChainID(privKey, big.NewInt(bscTestnetChainID))
	if err != nil {
		resp.ErrInternal(c, "db error")
		return
	}

	escrow, err := contract.NewP2PEscrow(common.HexToAddress(h.contractAddr), client)
	if err != nil {
		resp.ErrInternal(c, "db error")
		return
	}

	tx, err := escrow.ResolveDispute(auth, new(big.Int).SetUint64(req.OrderID), req.BuyerWins)
	if err != nil {
		h.log.Error("p2p resolve: tx", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "data": gin.H{"tx_hash": tx.Hash().Hex()}})
}

// ── DTO ───────────────────────────────────────────────────────────────────────

type orderDTO struct {
	OrderID      uint64 `json:"order_id"`
	Seller       string `json:"seller"`
	Buyer        string `json:"buyer,omitempty"`
	Token        string `json:"token"`
	Amount       string `json:"amount"`
	FiatAmount   uint64 `json:"fiat_amount"`
	FiatCurrency string `json:"fiat_currency"`
	PayMethod    string `json:"pay_method"`
	Status       uint8  `json:"status"`
	StatusLabel  string `json:"status_label"`
	TxHash       string `json:"tx_hash"`
}

func toOrderDTO(o model.P2POrder) orderDTO {
	return orderDTO{
		OrderID:      o.OrderID,
		Seller:       o.Seller,
		Buyer:        o.Buyer,
		Token:        o.Token,
		Amount:       o.Amount,
		FiatAmount:   o.FiatAmount,
		FiatCurrency: o.FiatCurrency,
		PayMethod:    o.PayMethod,
		Status:       o.Status,
		StatusLabel:  model.P2PStatusLabel(o.Status),
		TxHash:       o.TxHash,
	}
}

func toOrderDTOs(orders []model.P2POrder) []orderDTO {
	dtos := make([]orderDTO, len(orders))
	for i, o := range orders {
		dtos[i] = toOrderDTO(o)
	}
	return dtos
}
