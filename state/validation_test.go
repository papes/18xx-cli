package state

import (
	"testing"
)

func TestValidateMoney(t *testing.T) {
	tests := []struct {
		name    string
		amount  Money
		wantErr bool
	}{
		{"valid zero", 0, false},
		{"valid positive", 100, false},
		{"valid large", 999999, false},
		{"invalid negative", -1, true},
		{"invalid too large", 1000001, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateMoney(tt.amount)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateMoney() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateShareCount(t *testing.T) {
	tests := []struct {
		name    string
		shares  ShareCount
		wantErr bool
	}{
		{"valid zero", 0, false},
		{"valid positive", 10, false},
		{"valid max", 100, false},
		{"invalid negative", -1, true},
		{"invalid too large", 101, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateShareCount(tt.shares)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateShareCount() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateStockPrice(t *testing.T) {
	tests := []struct {
		name    string
		price   StockPrice
		wantErr bool
	}{
		{"valid zero", 0, false},
		{"valid positive", 100, false},
		{"valid max", 1000, false},
		{"invalid negative", -1, true},
		{"invalid too large", 1001, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateStockPrice(tt.price)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateStockPrice() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidatePlayerID(t *testing.T) {
	tests := []struct {
		name     string
		playerID string
		wantErr  bool
	}{
		{"valid short", "p1", false},
		{"valid medium", "player1", false},
		{"valid long", "this_is_a_very_long_player_name_but_still_valid", false},
		{"invalid empty", "", true},
		{"invalid too long", "this_player_name_is_way_too_long_and_exceeds_the_maximum_allowed_length_for_player_identifiers", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePlayerID(tt.playerID)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePlayerID() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateCompanyID(t *testing.T) {
	tests := []struct {
		name      string
		companyID string
		wantErr   bool
	}{
		{"valid short", "PRR", false},
		{"valid medium", "PENNSYLVANIA", false},
		{"valid max length", "12345678901234567890", false},
		{"invalid empty", "", true},
		{"invalid too long", "123456789012345678901", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCompanyID(tt.companyID)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateCompanyID() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateTransferAmount(t *testing.T) {
	tests := []struct {
		name      string
		amount    Money
		available Money
		wantErr   bool
	}{
		{"valid transfer", 100, 200, false},
		{"valid exact amount", 100, 100, false},
		{"invalid zero", 0, 100, true},
		{"invalid negative", -50, 100, true},
		{"invalid insufficient", 200, 100, true},
		{"invalid too large", 1000001, 2000000, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTransferAmount(tt.amount, tt.available)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateTransferAmount() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateDividendAmount(t *testing.T) {
	tests := []struct {
		name     string
		amount   Money
		treasury Money
		wantErr  bool
	}{
		{"valid dividend", 100, 200, false},
		{"valid exact treasury", 100, 100, false},
		{"invalid zero", 0, 100, true},
		{"invalid negative", -50, 100, true},
		{"invalid insufficient treasury", 200, 100, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDividendAmount(tt.amount, tt.treasury)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateDividendAmount() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateShareTransaction(t *testing.T) {
	tests := []struct {
		name        string
		shares      ShareCount
		available   ShareCount
		cost        Money
		playerCash  Money
		wantErr     bool
	}{
		{"valid transaction", 2, 5, 200, 300, false},
		{"valid exact shares", 5, 5, 200, 300, false},
		{"valid exact cash", 2, 5, 300, 300, false},
		{"invalid zero shares", 0, 5, 200, 300, true},
		{"invalid negative shares", -1, 5, 200, 300, true},
		{"invalid insufficient shares", 6, 5, 200, 300, true},
		{"invalid insufficient cash", 2, 5, 400, 300, true},
		{"invalid negative cost", 2, 5, -100, 300, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateShareTransaction(tt.shares, tt.available, tt.cost, tt.playerCash)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateShareTransaction() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}