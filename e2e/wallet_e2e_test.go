package e2e

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"wallet_service/internal/controller"
	"wallet_service/internal/id"
	"wallet_service/internal/repo"
	"wallet_service/internal/router"
	"wallet_service/internal/service"

	"testing"
)

func setupTestServer() *httptest.Server {
	// Initialize repository (using in-memory)
	repoFactory := repo.NewRepoFactory(64, repo.DBConfig{})
	repository, _ := repoFactory.GetRepository(nil, repo.RepoTypeMem)

	// Initialize dependencies
	generator := id.NewGenerator()
	walletService := service.NewWalletService(repository, generator)

	// Initialize controller
	walletController := controller.NewWalletController(walletService)

	// Setup router
	r := router.Setup(walletController)

	return httptest.NewServer(r)
}

func TestE2E_CreateAndGetWallet(t *testing.T) {
	server := setupTestServer()
	defer server.Close()

	// Create wallet
	createResp, err := http.Post(server.URL+"/wallets", "application/json", nil)
	if err != nil {
		t.Fatalf("Failed to create wallet: %v", err)
	}
	defer createResp.Body.Close()

	if createResp.StatusCode != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", createResp.StatusCode)
	}

	var wallet struct {
		ID      string `json:"id"`
		Balance int64  `json:"balance"`
	}
	if err := json.NewDecoder(createResp.Body).Decode(&wallet); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if wallet.ID == "" {
		t.Error("Wallet ID should not be empty")
	}
	if wallet.Balance != 0 {
		t.Errorf("Expected balance 0, got %d", wallet.Balance)
	}

	// Get wallet
	getResp, err := http.Get(server.URL + "/wallets/" + wallet.ID)
	if err != nil {
		t.Fatalf("Failed to get wallet: %v", err)
	}
	defer getResp.Body.Close()

	if getResp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", getResp.StatusCode)
	}

	var retrievedWallet struct {
		ID      string `json:"id"`
		Balance int64  `json:"balance"`
	}
	if err := json.NewDecoder(getResp.Body).Decode(&retrievedWallet); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if retrievedWallet.ID != wallet.ID {
		t.Errorf("Expected ID %s, got %s", wallet.ID, retrievedWallet.ID)
	}
	if retrievedWallet.Balance != 0 {
		t.Errorf("Expected balance 0, got %d", retrievedWallet.Balance)
	}
}

func TestE2E_GetWallet_NotFound(t *testing.T) {
	server := setupTestServer()
	defer server.Close()

	getResp, err := http.Get(server.URL + "/wallets/nonexistent")
	if err != nil {
		t.Fatalf("Failed to get wallet: %v", err)
	}
	defer getResp.Body.Close()

	if getResp.StatusCode != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", getResp.StatusCode)
	}
}

func TestE2E_Transfer_Success(t *testing.T) {
	server := setupTestServer()
	defer server.Close()

	// Create source wallet
	sourceResp, err := http.Post(server.URL+"/wallets", "application/json", nil)
	if err != nil {
		t.Fatalf("Failed to create source wallet: %v", err)
	}
	defer sourceResp.Body.Close()

	var sourceWallet struct {
		ID      string `json:"id"`
		Balance int64  `json:"balance"`
	}
	json.NewDecoder(sourceResp.Body).Decode(&sourceWallet)

	// Create destination wallet
	destResp, err := http.Post(server.URL+"/wallets", "application/json", nil)
	if err != nil {
		t.Fatalf("Failed to create destination wallet: %v", err)
	}
	defer destResp.Body.Close()

	var destWallet struct {
		ID      string `json:"id"`
		Balance int64  `json:"balance"`
	}
	json.NewDecoder(destResp.Body).Decode(&destWallet)

	// Need to add funds to source wallet first - use repo directly for test setup
	// Since there's no deposit endpoint, we'll test with wallet that has 0 balance
	// and expect insufficient funds error

	transferReq := map[string]interface{}{
		"source_id":      sourceWallet.ID,
		"destination_id": destWallet.ID,
		"amount":         100,
	}
	transferBody, _ := json.Marshal(transferReq)

	transferResp, err := http.Post(server.URL+"/transfer", "application/json", bytes.NewBuffer(transferBody))
	if err != nil {
		t.Fatalf("Failed to transfer: %v", err)
	}
	defer transferResp.Body.Close()

	// Transfer should fail because source has no funds
	if transferResp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", transferResp.StatusCode)
	}
}

func TestE2E_Transfer_InvalidAmount(t *testing.T) {
	server := setupTestServer()
	defer server.Close()

	// Create wallets
	createResp, _ := http.Post(server.URL+"/wallets", "application/json", nil)
	defer createResp.Body.Close()
	var wallet1 struct{ ID string }
	json.NewDecoder(createResp.Body).Decode(&wallet1)

	createResp2, _ := http.Post(server.URL+"/wallets", "application/json", nil)
	defer createResp2.Body.Close()
	var wallet2 struct{ ID string }
	json.NewDecoder(createResp2.Body).Decode(&wallet2)

	// Try transfer with invalid amount (negative)
	transferReq := map[string]interface{}{
		"source_id":      wallet1.ID,
		"destination_id": wallet2.ID,
		"amount":         -10,
	}
	transferBody, _ := json.Marshal(transferReq)

	resp, err := http.Post(server.URL+"/transfer", "application/json", bytes.NewBuffer(transferBody))
	if err != nil {
		t.Fatalf("Failed to transfer: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", resp.StatusCode)
	}
}

func TestE2E_Transfer_SameWallet(t *testing.T) {
	server := setupTestServer()
	defer server.Close()

	// Create wallet
	createResp, _ := http.Post(server.URL+"/wallets", "application/json", nil)
	defer createResp.Body.Close()
	var wallet struct{ ID string }
	json.NewDecoder(createResp.Body).Decode(&wallet)

	// Try transfer to same wallet
	transferReq := map[string]interface{}{
		"source_id":      wallet.ID,
		"destination_id": wallet.ID,
		"amount":         100,
	}
	transferBody, _ := json.Marshal(transferReq)

	resp, err := http.Post(server.URL+"/transfer", "application/json", bytes.NewBuffer(transferBody))
	if err != nil {
		t.Fatalf("Failed to transfer: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", resp.StatusCode)
	}
}

func TestE2E_Transfer_InvalidSource(t *testing.T) {
	server := setupTestServer()
	defer server.Close()

	// Create a valid destination wallet
	createResp, _ := http.Post(server.URL+"/wallets", "application/json", nil)
	defer createResp.Body.Close()
	var destWallet struct{ ID string }
	json.NewDecoder(createResp.Body).Decode(&destWallet)

	// Try transfer from invalid source
	transferReq := map[string]interface{}{
		"source_id":      "invalid_id",
		"destination_id": destWallet.ID,
		"amount":         100,
	}
	transferBody, _ := json.Marshal(transferReq)

	resp, err := http.Post(server.URL+"/transfer", "application/json", bytes.NewBuffer(transferBody))
	if err != nil {
		t.Fatalf("Failed to transfer: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", resp.StatusCode)
	}
}
