package actions

import (
	"testing"

	"github.com/papes/18xxCli/state"
)

func TestPayDividendWithConfig_IndividualLogging(t *testing.T) {
	// Set up a game state with players and a company
	gameState := state.NewGameState()
	gameState = gameState.AddPlayer("Alice", "Alice", 1000)
	gameState = gameState.AddPlayer("Bob", "Bob", 1000)
	gameState = gameState.AddCompany("BO", "Baltimore & Ohio", 10)

	// Give players shares (Alice: 3 shares, Bob: 2 shares)
	alice := gameState.Players["Alice"]
	alice.Shares["BO"] = 3
	bob := gameState.Players["Bob"]
	bob.Shares["BO"] = 2

	// Set up company with treasury and some shares in bank/IPO
	company := gameState.Companies["BO"]
	company.Treasury = 500
	company.SharesInIPO = 3
	company.SharesInBank = 2

	// Clear existing log entries
	gameState.Log = []state.LogEntry{}

	// Pay dividend: $100 total
	// With 5 player shares + 3 IPO + 2 bank = 10 total shares
	// Dividend per share = $10
	// Alice should get $30 (3 shares), Bob should get $20 (2 shares)
	newState, err := PayDividendWithConfig(gameState, "BO", 100, true, true)
	if err != nil {
		t.Fatalf("PayDividendWithConfig failed: %v", err)
	}

	// Check that we have the expected number of log entries
	// Should have: 2 individual payments + 1 company self-payment + 1 summary
	expectedEntries := 4
	if len(newState.Log) != expectedEntries {
		t.Fatalf("Expected %d log entries, got %d", expectedEntries, len(newState.Log))
	}

	// Verify individual payment logs
	var alicePayment, bobPayment, companySelfPayment, summaryEntry *state.LogEntry

	for i := range newState.Log {
		entry := &newState.Log[i]
		switch entry.Action {
		case "dividend_payment":
			if recipient, ok := entry.Details["recipient"].(string); ok {
				switch recipient {
				case "Alice":
					alicePayment = entry
				case "Bob":
					bobPayment = entry
				case "BO":
					companySelfPayment = entry
				}
			}
		case "pay_dividend_with_config":
			summaryEntry = entry
		}
	}

	// Verify Alice's payment
	if alicePayment == nil {
		t.Error("Alice's dividend payment not logged")
	} else {
		if amount, ok := alicePayment.Details["amount_received"].(state.Money); !ok || amount != 30 {
			t.Errorf("Alice should receive $30, got %v", alicePayment.Details["amount_received"])
		}
		if shares, ok := alicePayment.Details["shares_owned"].(state.ShareCount); !ok || shares != 3 {
			t.Errorf("Alice should own 3 shares, got %v", alicePayment.Details["shares_owned"])
		}
		if recipientType, ok := alicePayment.Details["recipient_type"].(string); !ok || recipientType != "player" {
			t.Errorf("Alice should be recipient_type 'player', got %v", recipientType)
		}
	}

	// Verify Bob's payment
	if bobPayment == nil {
		t.Error("Bob's dividend payment not logged")
	} else {
		if amount, ok := bobPayment.Details["amount_received"].(state.Money); !ok || amount != 20 {
			t.Errorf("Bob should receive $20, got %v", bobPayment.Details["amount_received"])
		}
		if shares, ok := bobPayment.Details["shares_owned"].(state.ShareCount); !ok || shares != 2 {
			t.Errorf("Bob should own 2 shares, got %v", bobPayment.Details["shares_owned"])
		}
	}

	// Verify company self-payment
	if companySelfPayment == nil {
		t.Error("Company self-payment not logged")
	} else {
		if amount, ok := companySelfPayment.Details["amount_received"].(state.Money); !ok || amount != 50 {
			t.Errorf("Company should receive $50 self-payout, got %v", companySelfPayment.Details["amount_received"])
		}
		if recipientType, ok := companySelfPayment.Details["recipient_type"].(string); !ok || recipientType != "company" {
			t.Errorf("Company should be recipient_type 'company', got %v", recipientType)
		}
	}

	// Verify summary entry
	if summaryEntry == nil {
		t.Error("Summary entry not logged")
	} else {
		if totalAmount, ok := summaryEntry.Details["total_amount"].(state.Money); !ok || totalAmount != 100 {
			t.Errorf("Summary should show total amount $100, got %v", summaryEntry.Details["total_amount"])
		}
		if individualPayments, ok := summaryEntry.Details["individual_payments"].(int); !ok || individualPayments != 3 {
			t.Errorf("Summary should show 3 individual payments, got %v", summaryEntry.Details["individual_payments"])
		}
	}

	// Verify actual money was transferred correctly
	if newState.Players["Alice"].Cash != 1030 {
		t.Errorf("Alice should have $1030, got $%d", newState.Players["Alice"].Cash)
	}
	if newState.Players["Bob"].Cash != 1020 {
		t.Errorf("Bob should have $1020, got $%d", newState.Players["Bob"].Cash)
	}
	if newState.Companies["BO"].Treasury != 450 {
		t.Errorf("Company should have $450 treasury (500 - 50 paid to players), got $%d", newState.Companies["BO"].Treasury)
	}
}

func TestPayDetailedDividend_IndividualLogging(t *testing.T) {
	// Set up a game state with players and a company
	gameState := state.NewGameState()
	gameState = gameState.AddPlayer("Charlie", "Charlie", 500)
	gameState = gameState.AddPlayer("Dave", "Dave", 500)
	gameState = gameState.AddCompany("PRR", "Pennsylvania Railroad", 10)

	// Give players shares (Charlie: 4 shares, Dave: 1 share)
	charlie := gameState.Players["Charlie"]
	charlie.Shares["PRR"] = 4
	dave := gameState.Players["Dave"]
	dave.Shares["PRR"] = 1

	// Set up company with treasury
	company := gameState.Companies["PRR"]
	company.Treasury = 200

	// Clear existing log entries
	gameState.Log = []state.LogEntry{}

	// Pay dividend: $75 total
	// With 5 player shares total
	// Dividend per share = $15
	// Charlie should get $60 (4 shares), Dave should get $15 (1 share)
	newState, err := PayDetailedDividend(gameState, "PRR", 75)
	if err != nil {
		t.Fatalf("PayDetailedDividend failed: %v", err)
	}

	// Check that we have the expected number of log entries
	// Should have: 2 individual payments + 1 summary
	expectedEntries := 3
	if len(newState.Log) != expectedEntries {
		t.Fatalf("Expected %d log entries, got %d", expectedEntries, len(newState.Log))
	}

	// Verify individual payment logs
	var charliePayment, davePayment, summaryEntry *state.LogEntry

	for i := range newState.Log {
		entry := &newState.Log[i]
		switch entry.Action {
		case "dividend_payment":
			if recipient, ok := entry.Details["recipient"].(string); ok {
				switch recipient {
				case "Charlie":
					charliePayment = entry
				case "Dave":
					davePayment = entry
				}
			}
		case "pay_detailed_dividend":
			summaryEntry = entry
		}
	}

	// Verify Charlie's payment
	if charliePayment == nil {
		t.Error("Charlie's dividend payment not logged")
	} else {
		if amount, ok := charliePayment.Details["amount_received"].(state.Money); !ok || amount != 60 {
			t.Errorf("Charlie should receive $60, got %v", charliePayment.Details["amount_received"])
		}
		if shares, ok := charliePayment.Details["shares_owned"].(state.ShareCount); !ok || shares != 4 {
			t.Errorf("Charlie should own 4 shares, got %v", charliePayment.Details["shares_owned"])
		}
	}

	// Verify Dave's payment
	if davePayment == nil {
		t.Error("Dave's dividend payment not logged")
	} else {
		if amount, ok := davePayment.Details["amount_received"].(state.Money); !ok || amount != 15 {
			t.Errorf("Dave should receive $15, got %v", davePayment.Details["amount_received"])
		}
		if shares, ok := davePayment.Details["shares_owned"].(state.ShareCount); !ok || shares != 1 {
			t.Errorf("Dave should own 1 share, got %v", davePayment.Details["shares_owned"])
		}
	}

	// Verify summary entry
	if summaryEntry == nil {
		t.Error("Summary entry not logged")
	} else {
		if individualPayments, ok := summaryEntry.Details["individual_payments"].(int); !ok || individualPayments != 2 {
			t.Errorf("Summary should show 2 individual payments, got %v", summaryEntry.Details["individual_payments"])
		}
	}

	// Verify actual money was transferred correctly
	if newState.Players["Charlie"].Cash != 560 {
		t.Errorf("Charlie should have $560, got $%d", newState.Players["Charlie"].Cash)
	}
	if newState.Players["Dave"].Cash != 515 {
		t.Errorf("Dave should have $515, got $%d", newState.Players["Dave"].Cash)
	}
	if newState.Companies["PRR"].Treasury != 125 {
		t.Errorf("Company should have $125 treasury (200 - 75), got $%d", newState.Companies["PRR"].Treasury)
	}
}
