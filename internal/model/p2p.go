package model

import "time"

// P2POrder mirrors on-chain order state, synced by the event listener.
type P2POrder struct {
	ID          uint64    `gorm:"primaryKey"`
	ChainID     uint64    `gorm:"not null;index"`
	ContractAddr string   `gorm:"not null"`
	OrderID     uint64    `gorm:"not null;uniqueIndex:idx_chain_order"`
	Seller      string    `gorm:"not null;index"`
	Buyer       string    `gorm:"index"`
	Token       string    `gorm:"not null"`
	Amount      string    `gorm:"not null"`       // big.Int as decimal string
	FiatAmount  uint64    `gorm:"not null"`       // ×100 (cents)
	FiatCurrency string   `gorm:"not null"`
	PayMethod   string    `gorm:"not null"`
	Status      uint8     `gorm:"not null;index"` // mirrors OrderStatus enum
	TxHash      string    `gorm:"not null"`       // creation tx
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// P2POrderStatus mirrors the Solidity enum.
const (
	P2PStatusOpen      uint8 = 0
	P2PStatusLocked    uint8 = 1
	P2PStatusPaid      uint8 = 2
	P2PStatusReleased  uint8 = 3
	P2PStatusCancelled uint8 = 4
	P2PStatusDisputed  uint8 = 5
)

func P2PStatusLabel(s uint8) string {
	switch s {
	case P2PStatusOpen:
		return "open"
	case P2PStatusLocked:
		return "locked"
	case P2PStatusPaid:
		return "paid"
	case P2PStatusReleased:
		return "released"
	case P2PStatusCancelled:
		return "cancelled"
	case P2PStatusDisputed:
		return "disputed"
	default:
		return "unknown"
	}
}
