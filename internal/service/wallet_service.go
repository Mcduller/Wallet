package service

import (
	"context"
	"wallet_service/internal/domain"
	"wallet_service/internal/id"
	"wallet_service/internal/repo"
)

type WalletServiceImpl struct {
	repository  repo.Repository
	idGenerator id.IDGenerator
}

func NewWalletService(repository repo.Repository, generator id.IDGenerator) WalletService {
	return &WalletServiceImpl{
		repository:  repository,
		idGenerator: generator,
	}
}

func (w *WalletServiceImpl) CreateWallet(ctx context.Context) (*domain.Wallet, error) {
	wallet := domain.Wallet{
		ID:      w.idGenerator.NextWalletID(),
		Balance: 0,
	}

	wallet, err := w.repository.CreateWallet(ctx, wallet)
	if err != nil {
		return nil, err
	}

	return (&wallet).Clone(), nil
}

func (w *WalletServiceImpl) GetWallet(ctx context.Context, id string) (*domain.Wallet, error) {
	wallet, err := w.repository.GetWallet(ctx, id)
	if err != nil {
		return nil, domain.ErrWalletNotFound
	}

	return (&wallet).Clone(), nil
}

func (w *WalletServiceImpl) Transfer(ctx context.Context, sourceID, destinationID string, amount int) error {
	if amount < 0 {
		return domain.ErrInvalidAmount
	}

	if sourceID == destinationID {
		return domain.ErrSameWallet
	}

	return w.repository.Transfer(ctx, sourceID, destinationID, int64(amount))
}
