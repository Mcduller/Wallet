package service

import (
	"context"
	"wallet_service/internal/domain"
)

type WalletService interface {
	CreateWallet(ctx context.Context) (*domain.Wallet, error)
	GetWallet(ctx context.Context, id string) (*domain.Wallet, error)
	Transfer(ctx context.Context, sourceID, destinationID string, amount int) error
}
