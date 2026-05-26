package model

import "time"

type WalletAddress struct {
	ID        int64     `gorm:"primaryKey;autoIncrement"`
	UserID    string    `gorm:"size:64;not null;uniqueIndex:uidx_wallet_addr,priority:1"`
	ChainKey  string    `gorm:"size:32;not null;uniqueIndex:uidx_wallet_addr,priority:2"`
	Address   string    `gorm:"size:128;not null"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
}

func (WalletAddress) TableName() string { return "wallet_addresses" }

type WalletTx struct {
	ID             int64      `gorm:"primaryKey;autoIncrement"`
	UserID         string     `gorm:"size:64;not null;index:idx_wtx,priority:1"`
	ChainKey       string     `gorm:"size:32;not null;uniqueIndex:uidx_wtx,priority:1;index:idx_wtx,priority:2"`
	Address        string     `gorm:"size:128;not null;uniqueIndex:uidx_wtx,priority:2;index:idx_wtx,priority:3"`
	TxHash         string     `gorm:"size:128;not null;uniqueIndex:uidx_wtx,priority:3"`
	FromAddress    string     `gorm:"size:128;not null"`
	ToAddress      string     `gorm:"size:128;not null"`
	Value          string     `gorm:"size:64;not null"`
	Decimals       int        `gorm:"not null"`
	TokenSymbol    *string    `gorm:"size:32"`
	TokenContract  *string    `gorm:"size:128"`
	BlockNumber    int64
	BlockTimestamp *time.Time
	Source         string    `gorm:"size:32"`
	CreatedAt      time.Time `gorm:"autoCreateTime"`
}

func (WalletTx) TableName() string { return "wallet_txes" }

type WalletMoralisStream struct {
	ChainKey  string    `gorm:"primaryKey;size:32"`
	StreamID  string    `gorm:"size:128;not null"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
}

func (WalletMoralisStream) TableName() string { return "wallet_moralis_streams" }

// WalletStream stores stream IDs keyed by (provider_name, chain_key).
// Supersedes WalletMoralisStream; old table kept for historical data.
type WalletStream struct {
	ProviderName string    `gorm:"primaryKey;size:32"`
	ChainKey     string    `gorm:"primaryKey;size:32"`
	StreamID     string    `gorm:"size:128;not null"`
	CreatedAt    time.Time `gorm:"autoCreateTime"`
}

func (WalletStream) TableName() string { return "wallet_streams" }
