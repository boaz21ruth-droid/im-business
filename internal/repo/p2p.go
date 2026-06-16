package repo

import (
	"context"

	"github.com/web1/im-business/internal/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type P2PRepo struct{ db *gorm.DB }

func NewP2PRepo(db *gorm.DB) *P2PRepo { return &P2PRepo{db: db} }

// Upsert inserts or updates an order (idempotent on replay).
func (r *P2PRepo) Upsert(ctx context.Context, o *model.P2POrder) error {
	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "order_id"}, {Name: "chain_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"buyer", "status", "updated_at"}),
		}).
		Create(o).Error
}

// UpdateStatus updates only the status field.
func (r *P2PRepo) UpdateStatus(ctx context.Context, chainID, orderID uint64, status uint8) error {
	return r.db.WithContext(ctx).
		Model(&model.P2POrder{}).
		Where("chain_id = ? AND order_id = ?", chainID, orderID).
		Updates(map[string]any{"status": status}).Error
}

// UpdateBuyerAndStatus sets buyer + status in one query (for OrderTaken event).
func (r *P2PRepo) UpdateBuyerAndStatus(ctx context.Context, chainID, orderID uint64, buyer string, status uint8) error {
	return r.db.WithContext(ctx).
		Model(&model.P2POrder{}).
		Where("chain_id = ? AND order_id = ?", chainID, orderID).
		Updates(map[string]any{"buyer": buyer, "status": status}).Error
}

// ListOpen returns all open orders, newest first.
func (r *P2PRepo) ListOpen(ctx context.Context, limit, offset int) ([]model.P2POrder, error) {
	var orders []model.P2POrder
	err := r.db.WithContext(ctx).
		Where("status = ?", model.P2PStatusOpen).
		Order("id DESC").
		Limit(limit).Offset(offset).
		Find(&orders).Error
	return orders, err
}

// GetByOnchainID fetches a single order by chain + contract order ID.
func (r *P2PRepo) GetByOnchainID(ctx context.Context, chainID, orderID uint64) (*model.P2POrder, error) {
	var o model.P2POrder
	err := r.db.WithContext(ctx).
		Where("chain_id = ? AND order_id = ?", chainID, orderID).
		First(&o).Error
	if err != nil {
		return nil, err
	}
	return &o, nil
}

// ListBySeller returns orders created by a given address.
func (r *P2PRepo) ListBySeller(ctx context.Context, seller string, limit, offset int) ([]model.P2POrder, error) {
	var orders []model.P2POrder
	err := r.db.WithContext(ctx).
		Where("seller = ?", seller).
		Order("id DESC").
		Limit(limit).Offset(offset).
		Find(&orders).Error
	return orders, err
}

// ListByBuyer returns orders taken by a given address.
func (r *P2PRepo) ListByBuyer(ctx context.Context, buyer string, limit, offset int) ([]model.P2POrder, error) {
	var orders []model.P2POrder
	err := r.db.WithContext(ctx).
		Where("buyer = ?", buyer).
		Order("id DESC").
		Limit(limit).Offset(offset).
		Find(&orders).Error
	return orders, err
}

// ListDisputed returns all disputed orders for admin review.
func (r *P2PRepo) ListDisputed(ctx context.Context) ([]model.P2POrder, error) {
	var orders []model.P2POrder
	err := r.db.WithContext(ctx).
		Where("status = ?", model.P2PStatusDisputed).
		Order("id DESC").
		Find(&orders).Error
	return orders, err
}
