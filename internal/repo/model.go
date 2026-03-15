package repo

import "wallet_service/internal/domain"

// WalletModel is the GORM model for wallet
type WalletModel struct {
	ID      string `gorm:"primaryKey;size:64"`
	Balance int64  `gorm:"not null;default:0"`
}

func (WalletModel) TableName() string {
	return "wallets"
}

func toModel(w domain.Wallet) WalletModel {
	return WalletModel{
		ID:      w.ID,
		Balance: w.Balance,
	}
}

func toDomain(m WalletModel) domain.Wallet {
	return domain.Wallet{
		ID:      m.ID,
		Balance: m.Balance,
	}
}
