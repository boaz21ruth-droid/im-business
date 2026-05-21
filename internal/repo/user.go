package repo

import (
	"errors"
	"strings"

	"github.com/web1/im-business/internal/model"
	"gorm.io/gorm"
)

type UserRepo struct {
	db *gorm.DB
}

func NewUserRepo(db *gorm.DB) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) Create(u *model.User) error {
	return r.db.Create(u).Error
}

// FindByAccount looks up by account handle, email, or phone (any of the three).
func (r *UserRepo) FindByAccount(account string) (*model.User, error) {
	var u model.User
	err := r.db.Where("account = ? OR email = ? OR phone_number = ?", account, account, account).
		First(&u).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &u, err
}

func (r *UserRepo) FindByUserID(userID string) (*model.User, error) {
	var u model.User
	err := r.db.Where("user_id = ?", userID).First(&u).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &u, err
}

func (r *UserRepo) FindByIDs(userIDs []string) ([]model.User, error) {
	var users []model.User
	return users, r.db.Where("user_id IN ?", userIDs).Find(&users).Error
}

func (r *UserRepo) Search(keyword string, offset, limit int) ([]model.User, error) {
	var users []model.User
	like := "%" + keyword + "%"
	err := r.db.Where(
		"nickname ILIKE ? OR account ILIKE ? OR phone_number LIKE ? OR email ILIKE ? OR user_id = ?",
		like, like, like, like, keyword,
	).Offset(offset).Limit(limit).Find(&users).Error
	return users, err
}

func (r *UserRepo) Update(userID string, fields map[string]any) error {
	return r.db.Model(&model.User{}).Where("user_id = ?", userID).Updates(fields).Error
}

func (r *UserRepo) ExistsByAccount(account, email, phone string) (bool, error) {
	var clauses []string
	var args []any
	if account != "" {
		clauses = append(clauses, "account = ?")
		args = append(args, account)
	}
	if email != "" {
		clauses = append(clauses, "email = ?")
		args = append(args, email)
	}
	if phone != "" {
		clauses = append(clauses, "phone_number = ?")
		args = append(args, phone)
	}
	if len(clauses) == 0 {
		return false, nil
	}
	var count int64
	err := r.db.Model(&model.User{}).
		Where(strings.Join(clauses, " OR "), args...).
		Count(&count).Error
	return count > 0, err
}
