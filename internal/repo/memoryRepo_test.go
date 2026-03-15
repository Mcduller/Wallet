package repo

import (
	"context"
	"sync"
	"testing"
	"wallet_service/internal/domain"
)

func TestMemoryRepo_TransferIsSafeUnderConcurrentAccess(t *testing.T) {
	repo := NewMemoryRepo(32)
	ctx := context.Background()

	source, err := repo.CreateWallet(ctx, domain.Wallet{
		ID:      "source",
		Balance: 10_000,
	})
	if err != nil {
		t.Fatalf("create source wallet: %v", err)
	}

	destination, err := repo.CreateWallet(ctx, domain.Wallet{
		ID:      "destination",
		Balance: 0,
	})
	if err != nil {
		t.Fatalf("create destination wallet: %v", err)
	}

	const workers = 100
	const amount = int64(10)

	var wg sync.WaitGroup
	errs := make(chan error, workers)

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			errs <- repo.Transfer(ctx, source.ID, destination.ID, amount)
		}()
	}

	wg.Wait()
	close(errs)

	for err := range errs {
		if err != nil {
			t.Fatalf("expected no transfer errors, got %v", err)
		}
	}

	gotSource, err := repo.GetWallet(ctx, source.ID)
	if err != nil {
		t.Fatalf("get source wallet: %v", err)
	}

	gotDestination, err := repo.GetWallet(ctx, destination.ID)
	if err != nil {
		t.Fatalf("get destination wallet: %v", err)
	}

	if gotSource.Balance != 9_000 {
		t.Fatalf("expected source balance 9000, got %d", gotSource.Balance)
	}

	if gotDestination.Balance != 1_000 {
		t.Fatalf("expected destination balance 1000, got %d", gotDestination.Balance)
	}
}
