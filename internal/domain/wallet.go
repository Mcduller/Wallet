package domain

import "errors"

var (
	ErrWalletNotFound    = errors.New("wallet not found")
	ErrInsufficientFunds = errors.New("insufficient funds")
	ErrInvalidAmount     = errors.New("amount must be greater than zero")
	ErrSameWallet        = errors.New("source and destination wallets must differ")
	ErrInvalidWalletID   = errors.New("wallet id is required")
)

type Wallet struct {
	ID      string `json:"id"`
	Balance int64  `json:"balance"`
}

func NewWallet(id string, balance int64) *Wallet {
	return &Wallet{
		ID:      id,
		Balance: balance,
	}
}

func (w *Wallet) Clone() *Wallet {
	return &Wallet{
		ID:      w.ID,
		Balance: w.Balance,
	}
}
