package repo

import (
	"strings"

	"github.com/web1/im-business/internal/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type WalletRepo struct {
	db *gorm.DB
}

func NewWalletRepo(db *gorm.DB) *WalletRepo {
	return &WalletRepo{db: db}
}

func (r *WalletRepo) UpsertAddress(userID, chainKey, address string) error {
	row := model.WalletAddress{UserID: userID, ChainKey: chainKey, Address: address}
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}, {Name: "chain_key"}},
		DoUpdates: clause.AssignmentColumns([]string{"address"}),
	}).Create(&row).Error
}

func (r *WalletRepo) GetAddressesByUserID(userID string) ([]model.WalletAddress, error) {
	var rows []model.WalletAddress
	return rows, r.db.Where("user_id = ?", userID).Find(&rows).Error
}

// FindUsersByAddress returns all userIDs whose wallet address matches (case-insensitive for EVM).
func (r *WalletRepo) FindUsersByAddress(address string) ([]string, error) {
	var rows []model.WalletAddress
	err := r.db.Where("LOWER(address) = LOWER(?)", address).Find(&rows).Error
	if err != nil {
		return nil, err
	}
	seen := make(map[string]bool)
	var ids []string
	for _, row := range rows {
		if !seen[row.UserID] {
			seen[row.UserID] = true
			ids = append(ids, row.UserID)
		}
	}
	return ids, nil
}

// FindAddressByUserChain returns the stored address for a given user + chain.
func (r *WalletRepo) FindAddressByUserChain(userID, chainKey string) (string, error) {
	var row model.WalletAddress
	err := r.db.Where("user_id = ? AND chain_key = ?", userID, chainKey).First(&row).Error
	if err != nil {
		return "", err
	}
	return row.Address, nil
}

// InsertTx inserts a transaction record, ignoring duplicates (idempotent).
func (r *WalletRepo) InsertTx(tx *model.WalletTx) error {
	return r.db.Clauses(clause.OnConflict{DoNothing: true}).Create(tx).Error
}

// QueryTxs returns paginated transaction history for a user/chain/address/contract.
func (r *WalletRepo) QueryTxs(userID, chainKey, address, contract string, page, limit int) ([]model.WalletTx, error) {
	q := r.db.Where("user_id = ? AND chain_key = ? AND address = ?", userID, chainKey, address)
	if contract != "" {
		q = q.Where("LOWER(token_contract) = LOWER(?)", contract)
	} else {
		q = q.Where("token_contract IS NULL OR token_contract = ''")
	}
	offset := (page - 1) * limit
	var rows []model.WalletTx
	return rows, q.Order("block_timestamp DESC NULLS LAST, id DESC").
		Offset(offset).Limit(limit).Find(&rows).Error
}

func (r *WalletRepo) UpsertStream(chainKey, streamID string) error {
	row := model.WalletMoralisStream{ChainKey: chainKey, StreamID: streamID}
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "chain_key"}},
		DoUpdates: clause.AssignmentColumns([]string{"stream_id"}),
	}).Create(&row).Error
}

func (r *WalletRepo) GetStreamID(chainKey string) (string, bool, error) {
	var row model.WalletMoralisStream
	err := r.db.Where("chain_key = ?", chainKey).First(&row).Error
	if err != nil {
		if strings.Contains(err.Error(), "record not found") {
			return "", false, nil
		}
		return "", false, err
	}
	return row.StreamID, true, nil
}

// UpsertStreamByProvider stores (or updates) a stream ID for a given provider + chain.
func (r *WalletRepo) UpsertStreamByProvider(providerName, chainKey, streamID string) error {
	row := model.WalletStream{ProviderName: providerName, ChainKey: chainKey, StreamID: streamID}
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "provider_name"}, {Name: "chain_key"}},
		DoUpdates: clause.AssignmentColumns([]string{"stream_id"}),
	}).Create(&row).Error
}

// GetStreamIDByProvider returns the stream ID for a given provider + chain, if present.
func (r *WalletRepo) GetStreamIDByProvider(providerName, chainKey string) (string, bool, error) {
	var row model.WalletStream
	err := r.db.Where("provider_name = ? AND chain_key = ?", providerName, chainKey).First(&row).Error
	if err != nil {
		if strings.Contains(err.Error(), "record not found") {
			return "", false, nil
		}
		return "", false, err
	}
	return row.StreamID, true, nil
}

// GetAddressesByChain returns all distinct wallet addresses registered for a chain.
func (r *WalletRepo) GetAddressesByChain(chainKey string) ([]string, error) {
	var rows []model.WalletAddress
	if err := r.db.Where("chain_key = ?", chainKey).Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]string, 0, len(rows))
	for _, row := range rows {
		out = append(out, row.Address)
	}
	return out, nil
}
