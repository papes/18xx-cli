package actions

import (
	"testing"

	"github.com/papes/18xxCli/state"
)

func setupCompanyTestGame() state.GameState {
	gameState := state.NewGameState()
	gameState = gameState.AddPlayer("alice", "Alice", 1000)
	gameState = gameState.AddPlayer("bob", "Bob", 1000)
	gameState = gameState.AddCompany("BO", "Baltimore & Ohio", 10)

	// Set up a basic company with par value and some shares sold
	gameState, _ = SetParValue(gameState, "BO", 80)
	gameState, _ = BuyShareFromIPO(gameState, "alice", "BO", 3) // Alice has 3 shares (30%)
	gameState, _ = BuyShareFromIPO(gameState, "bob", "BO", 2)   // Bob has 2 shares (20%)
	gameState, _ = SetPresident(gameState, "BO", "alice")       // Alice is president

	return gameState
}

func TestWithholdRevenue(t *testing.T) {
	gameState := setupCompanyTestGame()

	initialTreasury := gameState.Companies["BO"].Treasury

	// Company withholds $200
	finalState, err := WithholdRevenue(gameState, "BO", 200)
	if err != nil {
		t.Fatalf("WithholdRevenue failed: %v", err)
	}

	// Check treasury increased
	company := finalState.Companies["BO"]
	expectedTreasury := initialTreasury + 200
	if company.Treasury != expectedTreasury {
		t.Errorf("Expected treasury %d, got %d", expectedTreasury, company.Treasury)
	}

	// Check log entry
	if len(finalState.Log) != len(gameState.Log)+1 {
		t.Errorf("Expected 1 new log entry")
	}

	lastEntry := finalState.Log[len(finalState.Log)-1]
	if lastEntry.Action != "withhold_revenue" {
		t.Errorf("Expected 'withhold_revenue' action, got %s", lastEntry.Action)
	}
}

func TestPayDetailedDividend(t *testing.T) {
	gameState := setupCompanyTestGame()

	// Add money to company treasury
	gameState, _ = WithholdRevenue(gameState, "BO", 500)

	initialAliceCash := gameState.Players["alice"].Cash
	initialBobCash := gameState.Players["bob"].Cash

	// Pay $100 total dividend (Alice: 3 shares, Bob: 2 shares = 5 total shares)
	// So $20 per share: Alice gets $60, Bob gets $40
	finalState, err := PayDetailedDividend(gameState, "BO", 100)
	if err != nil {
		t.Fatalf("PayDetailedDividend failed: %v", err)
	}

	// Check player cash increased correctly
	alice := finalState.Players["alice"]
	bob := finalState.Players["bob"]

	expectedAliceDividend := state.Money(60) // 3 shares * $20
	expectedBobDividend := state.Money(40)   // 2 shares * $20

	if alice.Cash != initialAliceCash+expectedAliceDividend {
		t.Errorf("Expected Alice cash %d, got %d", initialAliceCash+expectedAliceDividend, alice.Cash)
	}

	if bob.Cash != initialBobCash+expectedBobDividend {
		t.Errorf("Expected Bob cash %d, got %d", initialBobCash+expectedBobDividend, bob.Cash)
	}

	// Check company treasury reduced
	company := finalState.Companies["BO"]
	// Initial treasury: $400 (from IPO sales), +$500 (withheld), -$100 (dividend) = $800
	expectedTreasury := state.Money(800)
	if company.Treasury != expectedTreasury {
		t.Errorf("Expected company treasury %d, got %d", expectedTreasury, company.Treasury)
	}

	// Check detailed log entry
	lastEntry := finalState.Log[len(finalState.Log)-1]
	if lastEntry.Action != "pay_detailed_dividend" {
		t.Errorf("Expected 'pay_detailed_dividend' action, got %s", lastEntry.Action)
	}

	// Check log details include player payouts
	details := lastEntry.Details
	if details["total_amount"] != state.Money(100) {
		t.Errorf("Expected total_amount 100 in log, got %v", details["total_amount"])
	}

	playerPayouts, ok := details["player_payouts"].(map[string]state.Money)
	if !ok {
		t.Fatal("Expected player_payouts in log details")
	}

	if playerPayouts["alice"] != expectedAliceDividend {
		t.Errorf("Expected Alice payout %d in log, got %d", expectedAliceDividend, playerPayouts["alice"])
	}

	if playerPayouts["bob"] != expectedBobDividend {
		t.Errorf("Expected Bob payout %d in log, got %d", expectedBobDividend, playerPayouts["bob"])
	}
}

func TestPayDetailedDividendInsufficientFunds(t *testing.T) {
	gameState := setupCompanyTestGame()

	// Try to pay dividend with insufficient company funds
	_, err := PayDetailedDividend(gameState, "BO", 1000)
	if err == nil {
		t.Error("Expected error for insufficient company funds")
	}
}

func TestBuyTrain(t *testing.T) {
	gameState := setupCompanyTestGame()

	// Add money to company treasury
	gameState, _ = WithholdRevenue(gameState, "BO", 500)

	initialTreasury := gameState.Companies["BO"].Treasury

	// Buy a 2-train for $80
	finalState, err := BuyTrain(gameState, "BO", "2", 80)
	if err != nil {
		t.Fatalf("BuyTrain failed: %v", err)
	}

	// Check treasury reduced
	company := finalState.Companies["BO"]
	expectedTreasury := initialTreasury - 80
	if company.Treasury != expectedTreasury {
		t.Errorf("Expected treasury %d, got %d", expectedTreasury, company.Treasury)
	}

	// Check log entry
	lastEntry := finalState.Log[len(finalState.Log)-1]
	if lastEntry.Action != "buy_train" {
		t.Errorf("Expected 'buy_train' action, got %s", lastEntry.Action)
	}

	details := lastEntry.Details
	if details["train_type"] != "2" {
		t.Errorf("Expected train_type '2' in log, got %v", details["train_type"])
	}

	if details["cost"] != state.Money(80) {
		t.Errorf("Expected cost 80 in log, got %v", details["cost"])
	}
}

func TestBuyTrainInsufficientFunds(t *testing.T) {
	gameState := setupCompanyTestGame()

	// Try to buy expensive train with insufficient funds
	_, err := BuyTrain(gameState, "BO", "6", 1000)
	if err == nil {
		t.Error("Expected error for insufficient company funds")
	}
}

func TestFloatCompany(t *testing.T) {
	gameState := state.NewGameState()
	gameState = gameState.AddPlayer("alice", "Alice", 1000)
	gameState = gameState.AddCompany("PRR", "Pennsylvania Railroad", 10)
	gameState, _ = SetParValue(gameState, "PRR", 76)

	// Sell 6 shares (60%) to float the company
	gameState, _ = BuyShareFromIPO(gameState, "alice", "PRR", 6)

	initialTreasury := gameState.Companies["PRR"].Treasury

	// Float the company
	finalState, err := FloatCompany(gameState, "PRR")
	if err != nil {
		t.Fatalf("FloatCompany failed: %v", err)
	}

	// Check treasury increased by share proceeds
	company := finalState.Companies["PRR"]
	sharesSold := state.ShareCount(6)
	expectedProceeds := state.Money(sharesSold) * state.Money(76) // 6 * 76 = 456
	expectedTreasury := initialTreasury + expectedProceeds

	if company.Treasury != expectedTreasury {
		t.Errorf("Expected treasury %d, got %d", expectedTreasury, company.Treasury)
	}

	// Check log entry
	lastEntry := finalState.Log[len(finalState.Log)-1]
	if lastEntry.Action != "float_company" {
		t.Errorf("Expected 'float_company' action, got %s", lastEntry.Action)
	}

	details := lastEntry.Details
	if details["shares_sold"] != sharesSold {
		t.Errorf("Expected shares_sold %d in log, got %v", sharesSold, details["shares_sold"])
	}

	if details["proceeds"] != expectedProceeds {
		t.Errorf("Expected proceeds %d in log, got %v", expectedProceeds, details["proceeds"])
	}
}

func TestSetPresident(t *testing.T) {
	gameState := setupCompanyTestGame()

	// Change president from alice to bob
	finalState, err := SetPresident(gameState, "BO", "bob")
	if err != nil {
		t.Fatalf("SetPresident failed: %v", err)
	}

	// Check president changed
	company := finalState.Companies["BO"]
	if company.PresidentID != "bob" {
		t.Errorf("Expected president 'bob', got %s", company.PresidentID)
	}

	// Check log entry
	lastEntry := finalState.Log[len(finalState.Log)-1]
	if lastEntry.Action != "set_president" {
		t.Errorf("Expected 'set_president' action, got %s", lastEntry.Action)
	}

	details := lastEntry.Details
	if details["old_president"] != "alice" {
		t.Errorf("Expected old_president 'alice' in log, got %v", details["old_president"])
	}

	if details["new_president"] != "bob" {
		t.Errorf("Expected new_president 'bob' in log, got %v", details["new_president"])
	}
}

func TestSetPresidentInvalidPlayer(t *testing.T) {
	gameState := setupCompanyTestGame()

	// Try to set non-existent player as president
	_, err := SetPresident(gameState, "BO", "charlie")
	if err == nil {
		t.Error("Expected error for non-existent player")
	}
}

func TestCompanyRevenue(t *testing.T) {
	gameState := setupCompanyTestGame()

	initialTreasury := gameState.Companies["BO"].Treasury

	// Company receives $300 revenue
	finalState, err := CompanyRevenue(gameState, "BO", 300)
	if err != nil {
		t.Fatalf("CompanyRevenue failed: %v", err)
	}

	// Check treasury increased
	company := finalState.Companies["BO"]
	expectedTreasury := initialTreasury + 300
	if company.Treasury != expectedTreasury {
		t.Errorf("Expected treasury %d, got %d", expectedTreasury, company.Treasury)
	}

	// Check log entry
	lastEntry := finalState.Log[len(finalState.Log)-1]
	if lastEntry.Action != "company_revenue" {
		t.Errorf("Expected 'company_revenue' action, got %s", lastEntry.Action)
	}

	details := lastEntry.Details
	if details["amount"] != state.Money(300) {
		t.Errorf("Expected amount 300 in log, got %v", details["amount"])
	}
}

func TestCompanyOperationsImmutability(t *testing.T) {
	gameState := setupCompanyTestGame()
	originalTreasury := gameState.Companies["BO"].Treasury

	// Perform operation and check original state unchanged
	WithholdRevenue(gameState, "BO", 100)

	if gameState.Companies["BO"].Treasury != originalTreasury {
		t.Error("Original state was modified by WithholdRevenue")
	}
}
