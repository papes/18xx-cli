package logging

import (
	"strings"
	"testing"
	"time"

	"github.com/papes/18xxCli/state"
)

func TestNewLogger(t *testing.T) {
	logger := NewLogger()
	if logger.timezone != time.UTC {
		t.Error("Expected default timezone to be UTC")
	}
}

func TestSetTimezone(t *testing.T) {
	logger := NewLogger()
	est, _ := time.LoadLocation("America/New_York")
	logger.SetTimezone(est)

	if logger.timezone != est {
		t.Error("Expected timezone to be set to EST")
	}
}

func TestLogPlayerAction(t *testing.T) {
	logger := NewLogger()
	entry := logger.LogPlayerAction("add_player", "alice", "Alice", map[string]interface{}{
		"starting_cash": 600,
	})

	if entry.Action != "add_player" {
		t.Errorf("Expected action 'add_player', got %s", entry.Action)
	}

	if entry.Details["player_id"] != "alice" {
		t.Error("Expected player_id to be set in details")
	}

	if entry.Details["player_name"] != "Alice" {
		t.Error("Expected player_name to be set in details")
	}

	if entry.Details["starting_cash"] != 600 {
		t.Error("Expected starting_cash to be preserved in details")
	}
}

func TestLogPlayerActionNilDetails(t *testing.T) {
	logger := NewLogger()
	entry := logger.LogPlayerAction("add_player", "bob", "Bob", nil)

	if entry.Details == nil {
		t.Error("Expected details to be initialized")
	}

	if entry.Details["player_id"] != "bob" {
		t.Error("Expected player_id to be set even with nil input details")
	}
}

func TestLogCompanyAction(t *testing.T) {
	logger := NewLogger()
	entry := logger.LogCompanyAction("add_company", "BO", map[string]interface{}{
		"shares": 10,
	})

	if entry.Action != "add_company" {
		t.Errorf("Expected action 'add_company', got %s", entry.Action)
	}

	if entry.Details["company"] != "BO" {
		t.Error("Expected company to be set in details")
	}

	if entry.Details["shares"] != 10 {
		t.Error("Expected shares to be preserved in details")
	}
}

func TestLogShareTransaction(t *testing.T) {
	logger := NewLogger()
	entry := logger.LogShareTransaction("buy_share_ipo", "alice", "BO", 2, 67, 134)

	if entry.Action != "buy_share_ipo" {
		t.Errorf("Expected action 'buy_share_ipo', got %s", entry.Action)
	}

	expectedFields := map[string]interface{}{
		"player":  "alice",
		"company": "BO",
		"shares":  state.ShareCount(2),
		"price":   state.StockPrice(67),
		"total":   state.Money(134),
	}

	for key, expectedValue := range expectedFields {
		if entry.Details[key] != expectedValue {
			t.Errorf("Expected %s to be %v, got %v", key, expectedValue, entry.Details[key])
		}
	}
}

func TestLogMoneyTransfer(t *testing.T) {
	logger := NewLogger()
	entry := logger.LogMoneyTransfer("alice", "bob", 100, "player_to_player")

	if entry.Action != "transfer_money" {
		t.Errorf("Expected action 'transfer_money', got %s", entry.Action)
	}

	expectedFields := map[string]interface{}{
		"from":   "alice",
		"to":     "bob",
		"amount": state.Money(100),
		"type":   "player_to_player",
	}

	for key, expectedValue := range expectedFields {
		if entry.Details[key] != expectedValue {
			t.Errorf("Expected %s to be %v, got %v", key, expectedValue, entry.Details[key])
		}
	}
}

func TestLogDividendPayment(t *testing.T) {
	logger := NewLogger()
	entry := logger.LogDividendPayment("BO", "alice", "player", 2, 25, 50, 100)

	if entry.Action != "dividend_payment" {
		t.Errorf("Expected action 'dividend_payment', got %s", entry.Action)
	}

	expectedFields := map[string]interface{}{
		"company":            "BO",
		"recipient":          "alice",
		"recipient_type":     "player",
		"shares_owned":       state.ShareCount(2),
		"dividend_per_share": state.Money(25),
		"amount_received":    state.Money(50),
		"total_dividend":     state.Money(100),
	}

	for key, expectedValue := range expectedFields {
		if entry.Details[key] != expectedValue {
			t.Errorf("Expected %s to be %v, got %v", key, expectedValue, entry.Details[key])
		}
	}
}

func TestLogRevenue(t *testing.T) {
	logger := NewLogger()
	entry := logger.LogRevenue("BO", 200, "company")

	if entry.Action != "company_revenue" {
		t.Errorf("Expected action 'company_revenue', got %s", entry.Action)
	}

	if entry.Details["company"] != "BO" {
		t.Error("Expected company to be BO")
	}

	if entry.Details["amount"] != state.Money(200) {
		t.Error("Expected amount to be 200")
	}
}

func TestLogPriceChange(t *testing.T) {
	logger := NewLogger()
	entry := logger.LogPriceChange("set_stock_price", "BO", 67, 75)

	if entry.Action != "set_stock_price" {
		t.Errorf("Expected action 'set_stock_price', got %s", entry.Action)
	}

	expectedFields := map[string]interface{}{
		"company":   "BO",
		"old_price": state.StockPrice(67),
		"new_price": state.StockPrice(75),
	}

	for key, expectedValue := range expectedFields {
		if entry.Details[key] != expectedValue {
			t.Errorf("Expected %s to be %v, got %v", key, expectedValue, entry.Details[key])
		}
	}
}

func TestLogTrainPurchase(t *testing.T) {
	logger := NewLogger()
	entry := logger.LogTrainPurchase("BO", "4-train", 300)

	if entry.Action != "buy_train" {
		t.Errorf("Expected action 'buy_train', got %s", entry.Action)
	}

	expectedFields := map[string]interface{}{
		"company":    "BO",
		"train_type": "4-train",
		"cost":       state.Money(300),
	}

	for key, expectedValue := range expectedFields {
		if entry.Details[key] != expectedValue {
			t.Errorf("Expected %s to be %v, got %v", key, expectedValue, entry.Details[key])
		}
	}
}

func TestLogPresidentChange(t *testing.T) {
	logger := NewLogger()
	entry := logger.LogPresidentChange("BO", "alice", "bob")

	if entry.Action != "set_president" {
		t.Errorf("Expected action 'set_president', got %s", entry.Action)
	}

	expectedFields := map[string]interface{}{
		"company":       "BO",
		"old_president": "alice",
		"new_president": "bob",
	}

	for key, expectedValue := range expectedFields {
		if entry.Details[key] != expectedValue {
			t.Errorf("Expected %s to be %v, got %v", key, expectedValue, entry.Details[key])
		}
	}
}

func TestLogGameSetup(t *testing.T) {
	logger := NewLogger()
	details := map[string]interface{}{
		"game":         "1830",
		"player_count": 4,
	}
	entry := logger.LogGameSetup("setup_game", details)

	if entry.Action != "setup_game" {
		t.Errorf("Expected action 'setup_game', got %s", entry.Action)
	}

	if entry.Details["game"] != "1830" {
		t.Error("Expected game to be 1830")
	}

	if entry.Details["player_count"] != 4 {
		t.Error("Expected player_count to be 4")
	}
}

func TestFormatLogEntry(t *testing.T) {
	logger := NewLogger()

	tests := []struct {
		name     string
		entry    state.LogEntry
		contains []string
	}{
		{
			name: "buy share ipo",
			entry: state.LogEntry{
				Timestamp: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
				Action:    "buy_share_ipo",
				Details: map[string]interface{}{
					"player":  "alice",
					"company": "BO",
					"shares":  2,
					"total":   134,
				},
			},
			contains: []string{"12:00:00", "buy_share_ipo", "alice buys 2 shares of BO for $134"},
		},
		{
			name: "dividend payment",
			entry: state.LogEntry{
				Timestamp: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
				Action:    "dividend_payment",
				Details: map[string]interface{}{
					"recipient":          "alice",
					"company":            "BO",
					"amount_received":    50,
					"shares_owned":       2,
					"recipient_type":     "player",
				},
			},
			contains: []string{"12:00:00", "dividend_payment", "alice receives $50 (2 shares of BO)"},
		},
		{
			name: "transfer money",
			entry: state.LogEntry{
				Timestamp: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
				Action:    "transfer_money",
				Details: map[string]interface{}{
					"from":   "alice",
					"to":     "bob",
					"amount": 100,
				},
			},
			contains: []string{"12:00:00", "transfer_money", "alice → bob: $100"},
		},
		{
			name: "company revenue",
			entry: state.LogEntry{
				Timestamp: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
				Action:    "company_revenue",
				Details: map[string]interface{}{
					"company": "BO",
					"amount":  200,
				},
			},
			contains: []string{"12:00:00", "company_revenue", "BO: $200"},
		},
		{
			name: "basic action without special formatting",
			entry: state.LogEntry{
				Timestamp: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
				Action:    "unknown_action",
				Details: map[string]interface{}{
					"company": "BO",
				},
			},
			contains: []string{"12:00:00", "unknown_action", "BO"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := logger.FormatLogEntry(tt.entry)
			for _, expectedSubstring := range tt.contains {
				if !strings.Contains(result, expectedSubstring) {
					t.Errorf("Expected formatted entry to contain %q, got: %s", expectedSubstring, result)
				}
			}
		})
	}
}

func TestExportLog(t *testing.T) {
	logger := NewLogger()
	gameState := state.NewGameState()

	// Add some log entries
	gameState.Log = []state.LogEntry{
		{
			Timestamp: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
			Action:    "add_player",
			Details: map[string]interface{}{
				"player_id": "alice",
			},
		},
		{
			Timestamp: time.Date(2023, 1, 1, 12, 1, 0, 0, time.UTC),
			Action:    "buy_share_ipo",
			Details: map[string]interface{}{
				"player":  "alice",
				"company": "BO",
			},
		},
	}

	// Test text export
	textResult, err := logger.ExportLog(gameState, "text")
	if err != nil {
		t.Errorf("Expected no error for text export, got: %v", err)
	}
	if !strings.Contains(textResult, "=== Game Log ===") {
		t.Error("Expected text export to contain header")
	}
	if !strings.Contains(textResult, "add_player") {
		t.Error("Expected text export to contain actions")
	}

	// Test JSON export
	jsonResult, err := logger.ExportLog(gameState, "json")
	if err != nil {
		t.Errorf("Expected no error for JSON export, got: %v", err)
	}
	if !strings.Contains(jsonResult, "timestamp") {
		t.Error("Expected JSON export to contain timestamp field")
	}
	if !strings.Contains(jsonResult, "add_player") {
		t.Error("Expected JSON export to contain actions")
	}

	// Test CSV export
	csvResult, err := logger.ExportLog(gameState, "csv")
	if err != nil {
		t.Errorf("Expected no error for CSV export, got: %v", err)
	}
	if !strings.Contains(csvResult, "Timestamp,Action,Details") {
		t.Error("Expected CSV export to contain headers")
	}
	if !strings.Contains(csvResult, "add_player") {
		t.Error("Expected CSV export to contain actions")
	}

	// Test unsupported format
	_, err = logger.ExportLog(gameState, "xml")
	if err == nil {
		t.Error("Expected error for unsupported export format")
	}
}

func TestTimestampInTimezone(t *testing.T) {
	logger := NewLogger()

	// Test with EST timezone
	est, _ := time.LoadLocation("America/New_York")
	logger.SetTimezone(est)

	entry := logger.LogPlayerAction("test", "alice", "Alice", nil)

	// The timestamp should be in EST
	if entry.Timestamp.Location() != est {
		t.Error("Expected timestamp to be in EST timezone")
	}
}

func TestLogEntryFields(t *testing.T) {
	logger := NewLogger()

	entry := logger.LogPlayerAction("test", "alice", "Alice", nil)

	// Check that all required fields are set
	if entry.Timestamp.IsZero() {
		t.Error("Expected timestamp to be set")
	}

	if entry.Action == "" {
		t.Error("Expected action to be set")
	}

	if entry.Details == nil {
		t.Error("Expected details to be initialized")
	}
}

func TestFormatLogEntryEdgeCases(t *testing.T) {
	logger := NewLogger()

	// Test with nil details
	entry := state.LogEntry{
		Timestamp: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
		Action:    "test_action",
		Details:   nil,
	}

	result := logger.FormatLogEntry(entry)
	if !strings.Contains(result, "test_action") {
		t.Error("Expected formatted entry to contain action")
	}

	// Test with empty details
	entry.Details = make(map[string]interface{})
	result = logger.FormatLogEntry(entry)
	if !strings.Contains(result, "test_action") {
		t.Error("Expected formatted entry to contain action with empty details")
	}
}