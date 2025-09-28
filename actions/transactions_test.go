package actions

import (
	"testing"

	"github.com/papes/18xxCli/state"
)

func setupTestGame() state.GameState {
	gameState := state.NewGameState()
	gameState = gameState.AddPlayer("alice", "Alice", 600)
	gameState = gameState.AddPlayer("bob", "Bob", 600)
	gameState = gameState.AddCompany("BO", "Baltimore & Ohio", 10)
	gameState = gameState.AddCompany("PRR", "Pennsylvania Railroad", 10)
	return gameState
}

func TestBuyShareFromIPO(t *testing.T) {
	gameState := setupTestGame()

	// Set par value first
	newState, err := SetParValue(gameState, "BO", 76)
	if err != nil {
		t.Fatalf("SetParValue failed: %v", err)
	}

	// Buy shares from IPO
	finalState, err := BuyShareFromIPO(newState, "alice", "BO", 2)
	if err != nil {
		t.Fatalf("BuyShareFromIPO failed: %v", err)
	}

	// Check player cash reduced
	alice := finalState.Players["alice"]
	expectedCash := state.Money(600 - 2*76) // 600 - 152 = 448
	if alice.Cash != expectedCash {
		t.Errorf("Expected Alice cash %d, got %d", expectedCash, alice.Cash)
	}

	// Check player owns shares
	if alice.Shares["BO"] != 2 {
		t.Errorf("Expected Alice to own 2 BO shares, got %d", alice.Shares["BO"])
	}

	// Check company IPO shares reduced
	company := finalState.Companies["BO"]
	if company.SharesInIPO != 8 {
		t.Errorf("Expected 8 shares in IPO, got %d", company.SharesInIPO)
	}

	// Check company treasury increased
	expectedTreasury := state.Money(2 * 76) // 152
	if company.Treasury != expectedTreasury {
		t.Errorf("Expected company treasury %d, got %d", expectedTreasury, company.Treasury)
	}

	// Check log entry was added
	if len(finalState.Log) != 2 { // SetParValue + BuyShare
		t.Errorf("Expected 2 log entries, got %d", len(finalState.Log))
	}

	// Check original state unchanged
	originalAlice := gameState.Players["alice"]
	if originalAlice.Cash != 600 {
		t.Error("Original state was modified")
	}
}

func TestBuyShareFromIPOInsufficientFunds(t *testing.T) {
	gameState := setupTestGame()

	// Set high par value
	newState, err := SetParValue(gameState, "BO", 100)
	if err != nil {
		t.Fatalf("SetParValue failed: %v", err)
	}

	// Try to buy more shares than player can afford
	_, err = BuyShareFromIPO(newState, "alice", "BO", 7) // 7 * 100 = 700, but alice has 600
	if err == nil {
		t.Error("Expected error for insufficient funds")
	}

	if err.Error() != "invalid share transaction: insufficient funds for share purchase: cost 700, available 600" {
		t.Errorf("Expected 'invalid share transaction: insufficient funds for share purchase: cost 700, available 600' error, got: %v", err)
	}
}

func TestBuyShareFromIPONoParValue(t *testing.T) {
	gameState := setupTestGame()

	// Try to buy without setting par value
	_, err := BuyShareFromIPO(gameState, "alice", "BO", 2)
	if err == nil {
		t.Error("Expected error when buying from IPO without par value")
	}

	if err.Error() != "company par value not set" {
		t.Errorf("Expected 'company par value not set' error, got: %v", err)
	}
}

func TestBuyShareFromIPOInsufficientShares(t *testing.T) {
	gameState := setupTestGame()

	// Set par value
	newState, err := SetParValue(gameState, "BO", 76)
	if err != nil {
		t.Fatalf("SetParValue failed: %v", err)
	}

	// Try to buy more shares than available
	_, err = BuyShareFromIPO(newState, "alice", "BO", 11) // Only 10 shares total
	if err == nil {
		t.Error("Expected error for insufficient shares in IPO")
	}

	if err.Error() != "invalid share transaction: insufficient shares available: requested 11, available 10" {
		t.Errorf("Expected 'invalid share transaction: insufficient shares available: requested 11, available 10' error, got: %v", err)
	}
}

func TestSellShareToBankPool(t *testing.T) {
	gameState := setupTestGame()

	// Set par value and buy shares first
	gameState, _ = SetParValue(gameState, "BO", 76)
	gameState, _ = BuyShareFromIPO(gameState, "alice", "BO", 3)

	// Set stock price (different from par)
	gameState, _ = SetStockPrice(gameState, "BO", 90)

	// Sell 1 share to bank
	finalState, err := SellShareToBankPool(gameState, "alice", "BO", 1)
	if err != nil {
		t.Fatalf("SellShareToBankPool failed: %v", err)
	}

	// Check player cash increased by stock price (not par value)
	alice := finalState.Players["alice"]
	expectedCash := state.Money(600 - 3*76 + 90) // 600 - 228 + 90 = 462
	if alice.Cash != expectedCash {
		t.Errorf("Expected Alice cash %d, got %d", expectedCash, alice.Cash)
	}

	// Check player owns fewer shares
	if alice.Shares["BO"] != 2 {
		t.Errorf("Expected Alice to own 2 BO shares, got %d", alice.Shares["BO"])
	}

	// Check shares moved to bank pool
	company := finalState.Companies["BO"]
	if company.SharesInBank != 1 {
		t.Errorf("Expected 1 share in bank pool, got %d", company.SharesInBank)
	}

	// Check bank money reduced
	expectedBankMoney := state.Money(12000 - 90) // Bank paid for the share
	if finalState.BankMoney != expectedBankMoney {
		t.Errorf("Expected bank money %d, got %d", expectedBankMoney, finalState.BankMoney)
	}
}

func TestSellShareToBankPoolInsufficientShares(t *testing.T) {
	gameState := setupTestGame()

	// Try to sell shares player doesn't own
	_, err := SellShareToBankPool(gameState, "alice", "BO", 1)
	if err == nil {
		t.Error("Expected error when selling shares player doesn't own")
	}

	if err.Error() != "player does not own enough shares: has 0, trying to sell 1" {
		t.Errorf("Expected 'player does not own enough shares: has 0, trying to sell 1' error, got: %v", err)
	}
}

func TestBuyShareFromBankPool(t *testing.T) {
	gameState := setupTestGame()

	// Set up scenario: par value, buy shares, set stock price, sell to bank
	gameState, _ = SetParValue(gameState, "BO", 76)
	gameState, _ = BuyShareFromIPO(gameState, "alice", "BO", 3)
	gameState, _ = SetStockPrice(gameState, "BO", 85)
	gameState, _ = SellShareToBankPool(gameState, "alice", "BO", 1) // 1 share in bank pool

	// Bob buys from bank pool
	finalState, err := BuyShareFromBankPool(gameState, "bob", "BO", 1)
	if err != nil {
		t.Fatalf("BuyShareFromBankPool failed: %v", err)
	}

	// Check Bob's cash reduced by stock price
	bob := finalState.Players["bob"]
	expectedCash := state.Money(600 - 85)
	if bob.Cash != expectedCash {
		t.Errorf("Expected Bob cash %d, got %d", expectedCash, bob.Cash)
	}

	// Check Bob owns shares
	if bob.Shares["BO"] != 1 {
		t.Errorf("Expected Bob to own 1 BO share, got %d", bob.Shares["BO"])
	}

	// Check bank pool reduced
	company := finalState.Companies["BO"]
	if company.SharesInBank != 0 {
		t.Errorf("Expected 0 shares in bank pool, got %d", company.SharesInBank)
	}

	// Check bank money back to original (paid 85 for sale, received 85 for purchase)
	expectedBankMoney := state.Money(12000) // 12000 - 85 + 85 = 12000
	if finalState.BankMoney != expectedBankMoney {
		t.Errorf("Expected bank money %d, got %d", expectedBankMoney, finalState.BankMoney)
	}
}

func TestBuyShareFromBankPoolNoShares(t *testing.T) {
	gameState := setupTestGame()

	// Try to buy from empty bank pool
	_, err := BuyShareFromBankPool(gameState, "alice", "BO", 1)
	if err == nil {
		t.Error("Expected error when buying from empty bank pool")
	}

	// The error will be "stock price not set" since we never set one
	if err.Error() != "stock price not set" {
		t.Errorf("Expected 'stock price not set' error, got: %v", err)
	}
}

func TestSetParValue(t *testing.T) {
	gameState := setupTestGame()

	newState, err := SetParValue(gameState, "BO", 82)
	if err != nil {
		t.Fatalf("SetParValue failed: %v", err)
	}

	company := newState.Companies["BO"]
	if company.ParValue != 82 {
		t.Errorf("Expected par value 82, got %d", company.ParValue)
	}

	// Check log entry
	if len(newState.Log) != 1 {
		t.Errorf("Expected 1 log entry, got %d", len(newState.Log))
	}

	// Try to set par value again (should fail)
	_, err = SetParValue(newState, "BO", 90)
	if err == nil {
		t.Error("Expected error when setting par value twice")
	}
}

func TestSetStockPrice(t *testing.T) {
	gameState := setupTestGame()

	newState, err := SetStockPrice(gameState, "BO", 95)
	if err != nil {
		t.Fatalf("SetStockPrice failed: %v", err)
	}

	company := newState.Companies["BO"]
	if company.StockPrice != 95 {
		t.Errorf("Expected stock price 95, got %d", company.StockPrice)
	}

	// Check log entry
	if len(newState.Log) != 1 {
		t.Errorf("Expected 1 log entry, got %d", len(newState.Log))
	}
}

func TestPayDividend(t *testing.T) {
	gameState := setupTestGame()

	// Set up scenario: both players own shares
	gameState, _ = SetParValue(gameState, "BO", 76)
	gameState, _ = BuyShareFromIPO(gameState, "alice", "BO", 3)
	gameState, _ = BuyShareFromIPO(gameState, "bob", "BO", 2)

	// Company has some money in treasury
	company := gameState.Companies["BO"]
	company.Treasury = 500 // From share sales: 5 * 76 = 380, add some operating revenue

	// Pay $10 per share dividend
	finalState, err := PayDividend(gameState, "BO", 10)
	if err != nil {
		t.Fatalf("PayDividend failed: %v", err)
	}

	// Check players received dividends
	alice := finalState.Players["alice"]
	expectedAliceCash := state.Money(600 - 3*76 + 3*10) // Original - share cost + dividends
	if alice.Cash != expectedAliceCash {
		t.Errorf("Expected Alice cash %d, got %d", expectedAliceCash, alice.Cash)
	}

	bob := finalState.Players["bob"]
	expectedBobCash := state.Money(600 - 2*76 + 2*10) // Original - share cost + dividends
	if bob.Cash != expectedBobCash {
		t.Errorf("Expected Bob cash %d, got %d", expectedBobCash, bob.Cash)
	}

	// Check company treasury reduced
	finalCompany := finalState.Companies["BO"]
	expectedTreasury := state.Money(500 - 5*10) // Total dividend payout: 5 shares * $10
	if finalCompany.Treasury != expectedTreasury {
		t.Errorf("Expected company treasury %d, got %d", expectedTreasury, finalCompany.Treasury)
	}
}

func TestTransferMoney(t *testing.T) {
	gameState := setupTestGame()

	// Transfer money from alice to bob
	finalState, err := TransferMoney(gameState, "alice", "bob", 100)
	if err != nil {
		t.Fatalf("TransferMoney failed: %v", err)
	}

	alice := finalState.Players["alice"]
	bob := finalState.Players["bob"]

	if alice.Cash != 500 { // 600 - 100
		t.Errorf("Expected Alice cash 500, got %d", alice.Cash)
	}

	if bob.Cash != 700 { // 600 + 100
		t.Errorf("Expected Bob cash 700, got %d", bob.Cash)
	}

	// Check log entry
	if len(finalState.Log) != 1 {
		t.Errorf("Expected 1 log entry, got %d", len(finalState.Log))
	}
}
