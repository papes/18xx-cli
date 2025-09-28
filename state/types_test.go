package state

import (
	"testing"
)

func TestNewGameState(t *testing.T) {
	state := NewGameState()

	if state.BankMoney != 12000 {
		t.Errorf("Expected bank money 12000, got %d", state.BankMoney)
	}

	if len(state.Players) != 0 {
		t.Errorf("Expected empty players map, got %d players", len(state.Players))
	}

	if len(state.Companies) != 0 {
		t.Errorf("Expected empty companies map, got %d companies", len(state.Companies))
	}

	if len(state.Log) != 0 {
		t.Errorf("Expected empty log, got %d entries", len(state.Log))
	}
}

func TestAddPlayer(t *testing.T) {
	state := NewGameState()

	// Add a player
	newState := state.AddPlayer("alice", "Alice", 600)

	// Original state should be unchanged
	if len(state.Players) != 0 {
		t.Errorf("Original state was modified - expected 0 players, got %d", len(state.Players))
	}

	// New state should have the player
	if len(newState.Players) != 1 {
		t.Errorf("Expected 1 player in new state, got %d", len(newState.Players))
	}

	player, exists := newState.Players["alice"]
	if !exists {
		t.Fatal("Player 'alice' not found in new state")
	}

	if player.ID != "alice" {
		t.Errorf("Expected player ID 'alice', got '%s'", player.ID)
	}

	if player.Name != "Alice" {
		t.Errorf("Expected player name 'Alice', got '%s'", player.Name)
	}

	if player.Cash != 600 {
		t.Errorf("Expected player cash 600, got %d", player.Cash)
	}

	if player.Shares == nil {
		t.Error("Player shares map should be initialized")
	}

	if len(player.Shares) != 0 {
		t.Errorf("Expected empty shares map, got %d entries", len(player.Shares))
	}
}

func TestAddCompany(t *testing.T) {
	state := NewGameState()

	// Add a company
	newState := state.AddCompany("BO", "Baltimore & Ohio", 10)

	// Original state should be unchanged
	if len(state.Companies) != 0 {
		t.Errorf("Original state was modified - expected 0 companies, got %d", len(state.Companies))
	}

	// New state should have the company
	if len(newState.Companies) != 1 {
		t.Errorf("Expected 1 company in new state, got %d", len(newState.Companies))
	}

	company, exists := newState.Companies["BO"]
	if !exists {
		t.Fatal("Company 'BO' not found in new state")
	}

	if company.ID != "BO" {
		t.Errorf("Expected company ID 'BO', got '%s'", company.ID)
	}

	if company.Name != "Baltimore & Ohio" {
		t.Errorf("Expected company name 'Baltimore & Ohio', got '%s'", company.Name)
	}

	if company.Treasury != 0 {
		t.Errorf("Expected company treasury 0, got %d", company.Treasury)
	}

	if company.StockPrice != 0 {
		t.Errorf("Expected stock price 0 (not set), got %d", company.StockPrice)
	}

	if company.ParValue != 0 {
		t.Errorf("Expected par value 0 (not set), got %d", company.ParValue)
	}

	if company.SharesInIPO != 10 {
		t.Errorf("Expected 10 shares in IPO, got %d", company.SharesInIPO)
	}

	if company.SharesInBank != 0 {
		t.Errorf("Expected 0 shares in bank, got %d", company.SharesInBank)
	}

	if company.TotalShares != 10 {
		t.Errorf("Expected 10 total shares, got %d", company.TotalShares)
	}
}

func TestGameStateClone(t *testing.T) {
	// Create a state with some data
	state := NewGameState()
	state = state.AddPlayer("alice", "Alice", 600)
	state = state.AddCompany("BO", "Baltimore & Ohio", 10)

	// Clone it
	cloned := state.Clone()

	// Verify they're equal but separate
	if len(cloned.Players) != len(state.Players) {
		t.Errorf("Cloned players count mismatch: expected %d, got %d", len(state.Players), len(cloned.Players))
	}

	if len(cloned.Companies) != len(state.Companies) {
		t.Errorf("Cloned companies count mismatch: expected %d, got %d", len(state.Companies), len(cloned.Companies))
	}

	if cloned.BankMoney != state.BankMoney {
		t.Errorf("Cloned bank money mismatch: expected %d, got %d", state.BankMoney, cloned.BankMoney)
	}

	// Verify they're separate objects (modify clone, check original unchanged)
	clonedPlayer := cloned.Players["alice"]
	clonedPlayer.Cash = 999

	originalPlayer := state.Players["alice"]
	if originalPlayer.Cash != 600 {
		t.Errorf("Original player cash was modified! Expected 600, got %d", originalPlayer.Cash)
	}

	// Test deep copy of shares map
	clonedPlayer.Shares["BO"] = 2

	if len(originalPlayer.Shares) != 0 {
		t.Errorf("Original player shares was modified! Expected empty map, got %d entries", len(originalPlayer.Shares))
	}
}

func TestGameStateImmutability(t *testing.T) {
	// Test that operations return new states and don't modify originals
	state1 := NewGameState()
	state2 := state1.AddPlayer("alice", "Alice", 600)
	state3 := state2.AddCompany("BO", "Baltimore & Ohio", 10)

	// Each state should be different
	if len(state1.Players) != 0 {
		t.Error("state1 should have no players")
	}

	if len(state2.Players) != 1 {
		t.Error("state2 should have 1 player")
	}

	if len(state2.Companies) != 0 {
		t.Error("state2 should have no companies")
	}

	if len(state3.Players) != 1 {
		t.Error("state3 should have 1 player")
	}

	if len(state3.Companies) != 1 {
		t.Error("state3 should have 1 company")
	}
}
