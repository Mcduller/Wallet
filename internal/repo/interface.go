package repo

import (
	"context"
	"wallet_service/internal/domain"
)

type Repository interface {
	CreateWallet(ctx context.Context, wallet domain.Wallet) (domain.Wallet, error)
	GetWallet(ctx context.Context, walletID string) (domain.Wallet, error)
	Transfer(ctx context.Context, sourceID, destinationID string, amount int64) error
}

type RepositoryFactory interface {
	GetRepository(ctx context.Context, repoType RepoType) (Repository, error)
}
