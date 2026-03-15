package repo

import (
	"context"
	"errors"
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"

	"wallet_service/internal/domain"
)

type MySQLRepo struct {
	db *gorm.DB
}

func NewMySQLRepo(cfg DBConfig) (*MySQLRepo, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.DBName,
	)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Auto migrate
	if err := db.AutoMigrate(&WalletModel{}); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return &MySQLRepo{db: db}, nil
}

func (r *MySQLRepo) CreateWallet(ctx context.Context, wallet domain.Wallet) (domain.Wallet, error) {
	model := toModel(wallet)
	if err := r.db.WithContext(ctx).Create(&model).Error; err != nil {
		return domain.Wallet{}, err
	}
	return toDomain(model), nil
}

func (r *MySQLRepo) GetWallet(ctx context.Context, walletID string) (domain.Wallet, error) {
	var model WalletModel
	if err := r.db.WithContext(ctx).Where("id = ?", walletID).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.Wallet{}, domain.ErrWalletNotFound
		}
		return domain.Wallet{}, err
	}
	return toDomain(model), nil
}

func (r *MySQLRepo) Transfer(ctx context.Context, sourceID, destinationID string, amount int64) error {
	firstID, secondID := sourceID, destinationID
	if secondID < firstID {
		firstID, secondID = secondID, firstID
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var first WalletModel
		if err := tx.
			Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ?", firstID).
			First(&first).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return domain.ErrWalletNotFound
			}
			return err
		}

		var second WalletModel
		if err := tx.
			Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ?", secondID).
			First(&second).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return domain.ErrWalletNotFound
			}
			return err
		}

		var source, dest *WalletModel
		if first.ID == sourceID {
			source = &first
			dest = &second
		} else {
			source = &second
			dest = &first
		}

		result := tx.Model(&WalletModel{}).
			Where("id = ? AND balance >= ?", source.ID, amount).
			Update("balance", gorm.Expr("balance - ?", amount))
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return domain.ErrInsufficientFunds
		}

		if err := tx.Model(&WalletModel{}).
			Where("id = ?", dest.ID).
			Update("balance", gorm.Expr("balance + ?", amount)).Error; err != nil {
			return err
		}

		return nil
	})
}

// Ensure MySQLRepo implements Repository
var _ Repository = (*MySQLRepo)(nil)
