package repo

import (
	"context"
	"hash/fnv"
	"sync"
	"wallet_service/internal/domain"
)

type MemoryRepo struct {
	segments []segment
}

type segment struct {
	mu      sync.RWMutex
	wallets map[string]domain.Wallet
}

func NewMemoryRepo(segmentCnt int) Repository {
	if segmentCnt <= 0 {
		segmentCnt = 64
	}

	segments := make([]segment, segmentCnt)
	for i := 0; i < segmentCnt; i++ {
		segments[i] = segment{
			wallets: make(map[string]domain.Wallet),
		}
	}
	return &MemoryRepo{segments: segments}
}

func (m *MemoryRepo) CreateWallet(_ context.Context, wallet domain.Wallet) (domain.Wallet, error) {
	id := wallet.ID

	seg := m.segmentForID(id)
	seg.mu.Lock()
	seg.wallets[id] = wallet
	seg.mu.Unlock()

	return wallet, nil

}

func (m *MemoryRepo) GetWallet(_ context.Context, walletID string) (domain.Wallet, error) {
	seg := m.segmentForID(walletID)
	seg.mu.RLock()
	record, ok := seg.wallets[walletID]
	seg.mu.RUnlock()
	if !ok {
		return domain.Wallet{}, domain.ErrWalletNotFound
	}
	return record, nil
}

func (m *MemoryRepo) Transfer(_ context.Context, sourceID, destinationID string, amount int64) error {
	if amount <= 0 {
		return domain.ErrInvalidAmount
	}
	if sourceID == destinationID {
		return domain.ErrSameWallet
	}

	sourceIndex := m.segmentIndex(sourceID)
	destinationIndex := m.segmentIndex(destinationID)

	if sourceIndex == destinationIndex {
		seg := &m.segments[sourceIndex]
		seg.mu.Lock()
		defer seg.mu.Unlock()
		return applyTransfer(seg.wallets, sourceID, destinationID, amount)
	}

	firstIndex, secondIndex := sourceIndex, destinationIndex
	if secondIndex < firstIndex {
		firstIndex, secondIndex = secondIndex, firstIndex
	}

	first := &m.segments[firstIndex]
	second := &m.segments[secondIndex]

	first.mu.Lock()
	second.mu.Lock()
	defer second.mu.Unlock()
	defer first.mu.Unlock()

	sourceSeg := &m.segments[sourceIndex]
	destinationSeg := &m.segments[destinationIndex]

	sourceRecord, ok := sourceSeg.wallets[sourceID]
	if !ok {
		return domain.ErrWalletNotFound
	}

	destinationRecord, ok := destinationSeg.wallets[destinationID]
	if !ok {
		return domain.ErrWalletNotFound
	}

	if sourceRecord.Balance < amount {
		return domain.ErrInsufficientFunds
	}

	sourceRecord.Balance -= amount
	destinationRecord.Balance += amount

	sourceSeg.wallets[sourceID] = sourceRecord
	destinationSeg.wallets[destinationID] = destinationRecord

	return nil
}

func applyTransfer(wallets map[string]domain.Wallet, sourceID, destinationID string, amount int64) error {
	sourceRecord, ok := wallets[sourceID]
	if !ok {
		return domain.ErrWalletNotFound
	}

	destinationRecord, ok := wallets[destinationID]
	if !ok {
		return domain.ErrWalletNotFound
	}

	if sourceRecord.Balance < amount {
		return domain.ErrInsufficientFunds
	}

	sourceRecord.Balance -= amount
	destinationRecord.Balance += amount
	wallets[sourceID] = sourceRecord
	wallets[destinationID] = destinationRecord
	return nil
}

func (m *MemoryRepo) segmentForID(walletID string) *segment {
	return &m.segments[m.segmentIndex(walletID)]
}

func (m *MemoryRepo) segmentIndex(walletID string) int {
	hasher := fnv.New32a()
	_, _ = hasher.Write([]byte(walletID))
	return int(hasher.Sum32() % uint32(len(m.segments)))
}

func (m *MemoryRepo) setBalance(walletID string, balance int64) {
	seg := m.segmentForID(walletID)
	seg.mu.Lock()
	record := seg.wallets[walletID]
	record.Balance = balance
	seg.wallets[walletID] = record
	seg.mu.Unlock()
}
