package parser

import (
	"testing"

	"github.com/papes/18xxCli/config"
	"github.com/papes/18xxCli/state"
)

func TestAllCommands_ComprehensiveCoverage(t *testing.T) {
	registry := CreateDefaultRegistry()
	executor := NewCommandExecutor(registry)

	// Create a base game state
	gameState := state.NewGameState()
	gameState = gameState.AddPlayer("alice", "Alice", 600)
	gameState = gameState.AddPlayer("bob", "Bob", 700)
	gameState = gameState.AddCompany("BO", "Baltimore & Ohio", 10)
	gameState = gameState.AddCompany("PRR", "Pennsylvania Railroad", 10)

	tests := []struct {
		name        string
		command     string
		expectError bool
	}{
		// Basic commands
		{"help", "help", false},
		{"status", "status", false},
		{"list games", "list games", true}, // Will fail without config directory

		// Player management
		{"add player", "add player charlie 800", false},
		{"add player invalid", "add player", true},
		{"add player bad amount", "add player dave abc", true},

		// Company management
		{"add company", "add company NYC 10", false},
		{"add company invalid", "add company", true},
		{"add company bad shares", "add company TEST abc", true},

		// Set par value
		{"set par value", "set par BO 100", false},
		{"set par invalid company", "set par NONEXISTENT 100", true},
		{"set par invalid amount", "set par BO abc", true},
		{"set par missing args", "set par", true},

		// Set stock price
		{"set stock price", "set price BO 120", false},
		{"set price invalid company", "set price NONEXISTENT 120", true},
		{"set price invalid amount", "set price BO abc", true},
		{"set price missing args", "set price", true},

		// Buy shares from IPO (need to set par first)
		{"buy ipo shares", "buy ipo alice BO 2", true}, // Will fail without par
		{"buy ipo invalid player", "buy ipo NONEXISTENT BO 2", true},
		{"buy ipo invalid company", "buy ipo alice NONEXISTENT 2", true},
		{"buy ipo invalid shares", "buy ipo alice BO abc", true},
		{"buy ipo missing args", "buy ipo", true},

		// Buy shares from bank (need stock price)
		{"buy bank shares", "buy bank alice BO 1", true}, // Will fail without stock price
		{"buy bank invalid player", "buy bank NONEXISTENT BO 1", true},
		{"buy bank invalid company", "buy bank alice NONEXISTENT 1", true},
		{"buy bank invalid shares", "buy bank alice BO abc", true},
		{"buy bank missing args", "buy bank", true},

		// Sell shares to bank (need to own shares)
		{"sell bank shares", "sell bank alice BO 1", true}, // Will fail without owning shares
		{"sell bank invalid player", "sell bank NONEXISTENT BO 1", true},
		{"sell bank invalid company", "sell bank alice NONEXISTENT 1", true},
		{"sell bank invalid shares", "sell bank alice BO abc", true},
		{"sell bank missing args", "sell bank", true},

		// Pay dividends (need config and shares)
		{"pay dividend", "pay dividend BO 50", true}, // Will fail without config
		{"pay dividend invalid company", "pay dividend NONEXISTENT 50", true},
		{"pay dividend invalid amount", "pay dividend BO abc", true},
		{"pay dividend missing args", "pay dividend", true},

		// Transfer money
		{"transfer money", "transfer alice bob 100", false},
		{"transfer invalid from", "transfer NONEXISTENT bob 100", true},
		{"transfer invalid to", "transfer alice NONEXISTENT 100", true},
		{"transfer invalid amount", "transfer alice bob abc", true},
		{"transfer missing args", "transfer", true},

		// Company operations
		{"withhold revenue", "withhold BO 200", false},
		{"withhold invalid company", "withhold NONEXISTENT 200", true},
		{"withhold invalid amount", "withhold BO abc", true},
		{"withhold missing args", "withhold", true},

		{"buy train", "buy train BO '4-train' 300", true}, // Will fail without money
		{"buy train invalid company", "buy train NONEXISTENT '4-train' 300", true},
		{"buy train invalid cost", "buy train BO '4-train' abc", true},
		{"buy train missing args", "buy train", true},

		{"set president", "set president BO alice", false},
		{"set president invalid company", "set president NONEXISTENT alice", true},
		{"set president invalid player", "set president BO NONEXISTENT", true},
		{"set president missing args", "set president", true},

		{"company revenue", "revenue BO 150", false},
		{"revenue invalid company", "revenue NONEXISTENT 150", true},
		{"revenue invalid amount", "revenue BO abc", true},
		{"revenue missing args", "revenue", true},

		// Game setup (need config file)
		{"setup game", "setup game 1830 alice bob", true}, // Will fail without config file
		{"setup invalid game", "setup game NONEXISTENT alice bob", true},
		{"setup missing args", "setup game", true},

		// Load companies command
		{"load companies", "load companies 1830", true}, // Will fail without config file
		{"load invalid config", "load companies NONEXISTENT", true},
		{"load invalid config", "load game NONEXISTENT", true},
		{"load missing args", "load game", true},

		// Invalid commands
		{"unknown command", "unknown_command", true},
		{"empty command", "", true},

		// More command variations for coverage
		{"load config missing args", "load", true},
		{"setup missing args", "setup", true},
		{"set missing args", "set", true},
		{"buy missing args", "buy", true},
		{"sell missing args", "sell", true},
		{"pay missing args", "pay", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := executor.Execute(tt.command, gameState, nil)

			if tt.expectError && result.Error == nil {
				t.Errorf("Expected error for command %q but got none", tt.command)
			}

			if !tt.expectError && result.Error != nil {
				t.Errorf("Expected no error for command %q but got: %v", tt.command, result.Error)
			}
		})
	}
}

func TestRegistryMethods(t *testing.T) {
	registry := CreateDefaultRegistry()

	// Test List method
	commands := registry.List()
	if len(commands) == 0 {
		t.Error("Registry should contain commands")
	}

	// Test GetSuggestions method
	suggestions := registry.GetSuggestions("he")
	found := false
	for _, suggestion := range suggestions {
		if suggestion == "help" {
			found = true
			break
		}
	}
	if !found {
		t.Error("GetSuggestions should return 'help' for prefix 'he'")
	}

	// Test GetSuggestions with no matches
	noSuggestions := registry.GetSuggestions("xyz")
	if len(noSuggestions) != 0 {
		t.Error("GetSuggestions should return empty slice for 'xyz'")
	}

	// Test Description method on a command
	helpCmd, exists := registry.Get("help")
	if !exists {
		t.Error("Help command should exist")
	}

	description := helpCmd.Description()
	if description == "" {
		t.Error("Help command should have a description")
	}
}

func TestUtilityFunctions(t *testing.T) {
	// Test formatMoney (just returns number as string)
	tests := []struct {
		input    int
		expected string
	}{
		{0, "0"},
		{100, "100"},
		{1000, "1000"},
		{1234567, "1234567"},
		{-100, "-100"},
	}

	for _, tt := range tests {
		result := formatMoney(tt.input)
		if result != tt.expected {
			t.Errorf("formatMoney(%d) = %q, expected %q", tt.input, result, tt.expected)
		}
	}

	// Test boolToString
	if boolToString(true) != "true" {
		t.Error("boolToString(true) should return 'true'")
	}
	if boolToString(false) != "false" {
		t.Error("boolToString(false) should return 'false'")
	}

	// Test parseStringAsInt
	val, err := parseStringAsInt("123")
	if err != nil || val != 123 {
		t.Errorf("parseStringAsInt('123') should return 123, nil, got %d, %v", val, err)
	}

	_, err = parseStringAsInt("abc")
	if err == nil {
		t.Error("parseStringAsInt('abc') should return an error")
	}

	// Test joinStrings
	result := joinStrings([]string{"a", "b", "c"}, ", ")
	if result != "a, b, c" {
		t.Errorf("joinStrings should return 'a, b, c', got %q", result)
	}

	// Test empty slice
	result = joinStrings([]string{}, ", ")
	if result != "" {
		t.Errorf("joinStrings with empty slice should return empty string, got %q", result)
	}

	// Test formatGameStatus
	gameState := state.NewGameState()
	gameState = gameState.AddPlayer("alice", "Alice", 600)
	gameState = gameState.AddCompany("BO", "Baltimore & Ohio", 10)
	statusText := formatGameStatus(gameState)
	if statusText == "" {
		t.Error("formatGameStatus should return non-empty string")
	}
}

func TestCommandValidationErrors(t *testing.T) {
	registry := CreateDefaultRegistry()
	executor := NewCommandExecutor(registry)
	gameState := state.NewGameState()

	// Test various validation error cases
	tests := []struct {
		name    string
		command string
	}{
		{"negative money", "add player alice -100"},
		{"negative shares", "add company TEST 'Test Corp' -5"},
		{"zero shares", "add company TEST 'Test Corp' 0"},
		{"zero money transfer", "transfer alice bob 0"},
		{"self transfer", "transfer alice alice 100"},
		{"invalid par value", "set par BO 0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := executor.Execute(tt.command, gameState, nil)
			if result.Error == nil {
				t.Errorf("Expected validation error for command %q", tt.command)
			}
		})
	}
}

func TestConfigIntegration(t *testing.T) {
	registry := CreateDefaultRegistry()
	executor := NewCommandExecutor(registry)
	gameState := state.NewGameState()

	// Create a test game config
	gameConfig := &config.GameConfig{
		Title:       "Test Game",
		Description: "Test Description",
		Companies: []config.Company{
			{ID: "TEST", Name: "Test Company", Shares: 10},
		},
		PlayerConfigs: []config.PlayerConfig{
			{PlayerCount: 2, StartingMoney: 500, BankMoney: 10000},
		},
		BankSharesPayCompany: true,
		IPOSharesPayCompany:  false,
	}

	// Test commands with game config
	result := executor.Execute("status", gameState, gameConfig)
	if result.Error != nil {
		t.Errorf("Status command with config should work, got error: %v", result.Error)
	}

	// Test dividend with config - need actual shares outstanding
	gameState = gameState.AddCompany("TEST", "Test Company", 10)
	gameState = gameState.AddPlayer("testplayer", "Test Player", 1000)
	// Add some shares to a player first
	gameState.Players["testplayer"].Shares["TEST"] = 2
	gameState.Companies["TEST"].Treasury = 500
	result = executor.Execute("pay dividend TEST 50", gameState, gameConfig)
	if result.Error != nil {
		t.Errorf("Pay dividend with config should work, got error: %v", result.Error)
	}
}