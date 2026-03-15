package service

import (
	"context"
	"sync"
	"testing"
	"wallet_service/internal/domain"
	"wallet_service/internal/id"
	"wallet_service/internal/repo"
)

func newTestService() WalletService {
	generator := id.NewGenerator()
	repository := repo.NewMemoryRepo(32)
	return NewWalletService(repository, generator)
}

func TestWalletService_CreateWallet(t *testing.T) {
	svc := newTestService()
	ctx := context.Background()

	wallet, err := svc.CreateWallet(ctx)
	if err != nil {
		t.Fatalf("CreateWallet failed: %v", err)
	}

	if wallet.ID == "" {
		t.Error("expected wallet ID to be generated")
	}

	if wallet.Balance != 0 {
		t.Errorf("expected initial balance 0, got %d", wallet.Balance)
	}
}

func TestWalletService_CreateWallet_Multiple(t *testing.T) {
	svc := newTestService()
	ctx := context.Background()

	wallet1, err := svc.CreateWallet(ctx)
	if err != nil {
		t.Fatalf("CreateWallet failed: %v", err)
	}

	wallet2, err := svc.CreateWallet(ctx)
	if err != nil {
		t.Fatalf("CreateWallet failed: %v", err)
	}

	if wallet1.ID == wallet2.ID {
		t.Error("expected different wallet IDs")
	}
}

func TestWalletService_GetWallet_Success(t *testing.T) {
	svc := newTestService()
	ctx := context.Background()

	created, err := svc.CreateWallet(ctx)
	if err != nil {
		t.Fatalf("CreateWallet failed: %v", err)
	}

	got, err := svc.GetWallet(ctx, created.ID)
	if err != nil {
		t.Fatalf("GetWallet failed: %v", err)
	}

	if got.ID != created.ID {
		t.Errorf("expected ID %s, got %s", created.ID, got.ID)
	}

	if got.Balance != created.Balance {
		t.Errorf("expected balance %d, got %d", created.Balance, got.Balance)
	}
}

func TestWalletService_GetWallet_NotFound(t *testing.T) {
	svc := newTestService()
	ctx := context.Background()

	_, err := svc.GetWallet(ctx, "nonexistent")
	if err != domain.ErrWalletNotFound {
		t.Errorf("expected ErrWalletNotFound, got %v", err)
	}
}

func TestWalletService_Transfer_Success(t *testing.T) {
	svc := newTestService()
	ctx := context.Background()

	source, err := svc.CreateWallet(ctx)
	if err != nil {
		t.Fatalf("CreateWallet failed: %v", err)
	}

	dest, err := svc.CreateWallet(ctx)
	if err != nil {
		t.Fatalf("CreateWallet failed: %v", err)
	}

	// Add balance to source via repo
	memRepo := svc.(*WalletServiceImpl).repository.(*repo.MemoryRepo)
	_, _ = memRepo.CreateWallet(ctx, domain.Wallet{ID: source.ID, Balance: 1000})

	err = svc.Transfer(ctx, source.ID, dest.ID, 500)
	if err != nil {
		t.Fatalf("Transfer failed: %v", err)
	}

	gotSource, _ := svc.GetWallet(ctx, source.ID)
	gotDest, _ := svc.GetWallet(ctx, dest.ID)

	if gotSource.Balance != 500 {
		t.Errorf("expected source balance 500, got %d", gotSource.Balance)
	}

	if gotDest.Balance != 500 {
		t.Errorf("expected dest balance 500, got %d", gotDest.Balance)
	}
}

func TestWalletService_Transfer_NegativeAmount(t *testing.T) {
	svc := newTestService()
	ctx := context.Background()

	err := svc.Transfer(ctx, "a", "b", -100)
	if err != domain.ErrInvalidAmount {
		t.Errorf("expected ErrInvalidAmount, got %v", err)
	}
}

func TestWalletService_Transfer_SameWallet(t *testing.T) {
	svc := newTestService()
	ctx := context.Background()

	err := svc.Transfer(ctx, "wallet1", "wallet1", 100)
	if err != domain.ErrSameWallet {
		t.Errorf("expected ErrSameWallet, got %v", err)
	}
}

func TestWalletService_Transfer_InsufficientFunds(t *testing.T) {
	svc := newTestService()
	ctx := context.Background()

	source, err := svc.CreateWallet(ctx)
	if err != nil {
		t.Fatalf("CreateWallet failed: %v", err)
	}

	dest, err := svc.CreateWallet(ctx)
	if err != nil {
		t.Fatalf("CreateWallet failed: %v", err)
	}

	// Source has 0 balance
	err = svc.Transfer(ctx, source.ID, dest.ID, 100)
	if err != domain.ErrInsufficientFunds {
		t.Errorf("expected ErrInsufficientFunds, got %v", err)
	}
}

func TestWalletService_Transfer_Concurrent(t *testing.T) {
	svc := newTestService()
	ctx := context.Background()

	source, _ := svc.CreateWallet(ctx)
	dest, _ := svc.CreateWallet(ctx)

	memRepo := svc.(*WalletServiceImpl).repository.(*repo.MemoryRepo)
	_, _ = memRepo.CreateWallet(ctx, domain.Wallet{ID: source.ID, Balance: 10_000})

	const workers = 50
	const amount = 100

	var wg sync.WaitGroup
	errChan := make(chan error, workers)

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			errChan <- svc.Transfer(ctx, source.ID, dest.ID, amount)
		}()
	}

	wg.Wait()
	close(errChan)

	for err := range errChan {
		if err != nil {
			t.Fatalf("transfer error: %v", err)
		}
	}

	gotSource, _ := svc.GetWallet(ctx, source.ID)
	gotDest, _ := svc.GetWallet(ctx, dest.ID)

	if gotSource.Balance != 5_000 {
		t.Errorf("expected source balance 5000, got %d", gotSource.Balance)
	}

	if gotDest.Balance != 5_000 {
		t.Errorf("expected dest balance 5000, got %d", gotDest.Balance)
	}
}
