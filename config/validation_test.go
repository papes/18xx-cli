package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateTitle(t *testing.T) {
	tests := []struct {
		title     string
		expectErr bool
	}{
		{"1830", false},
		{"18xx Rail Game", false},
		{"", true},                                                                              // empty
		{strings.Repeat("a", 101), true},                                                       // too long
		{"Game\x00Name", true},                                                                 // invalid character
		{"Valid Title 123", false},                                                             // valid
		{"Title with (parentheses) and [brackets] and symbols: !@#$%^&*", false},             // valid symbols
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			err := validateTitle(tt.title)
			if (err != nil) != tt.expectErr {
				t.Errorf("validateTitle(%q) error = %v, expectErr %v", tt.title, err, tt.expectErr)
			}
		})
	}
}

func TestValidateDescription(t *testing.T) {
	tests := []struct {
		description string
		expectErr   bool
	}{
		{"A classic 18xx game", false},
		{"", false},                               // empty is allowed
		{strings.Repeat("a", 500), false},        // max length
		{strings.Repeat("a", 501), true},         // too long
		{"Description\x00with null", true},       // invalid character
		{"Valid description with symbols!", false}, // valid
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			err := validateDescription(tt.description)
			if (err != nil) != tt.expectErr {
				t.Errorf("validateDescription(%q) error = %v, expectErr %v", tt.description, err, tt.expectErr)
			}
		})
	}
}

func TestValidateCompanyID(t *testing.T) {
	tests := []struct {
		id        string
		expectErr bool
	}{
		{"BO", false},
		{"B_O", false},
		{"B-O", false},
		{"PRR", false},
		{"NYC123", false},
		{"", true},                      // empty
		{strings.Repeat("a", 11), true}, // too long
		{"B O", true},                   // space not allowed
		{"B@O", true},                   // special character not allowed
		{"123", false},                  // numbers only is fine
	}

	for _, tt := range tests {
		t.Run(tt.id, func(t *testing.T) {
			err := validateCompanyID(tt.id)
			if (err != nil) != tt.expectErr {
				t.Errorf("validateCompanyID(%q) error = %v, expectErr %v", tt.id, err, tt.expectErr)
			}
		})
	}
}

func TestValidateCompanyName(t *testing.T) {
	tests := []struct {
		name      string
		expectErr bool
	}{
		{"Baltimore & Ohio", false},
		{"Pennsylvania Railroad", false},
		{"", true},                               // empty
		{strings.Repeat("a", 100), false},       // max length
		{strings.Repeat("a", 101), true},        // too long
		{"Name\x00with null", true},             // invalid character
		{"Valid Company Name 123!", false},      // valid with symbols
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateCompanyName(tt.name)
			if (err != nil) != tt.expectErr {
				t.Errorf("validateCompanyName(%q) error = %v, expectErr %v", tt.name, err, tt.expectErr)
			}
		})
	}
}

func TestValidateShares(t *testing.T) {
	tests := []struct {
		shares    int
		expectErr bool
	}{
		{10, false},
		{1, false},
		{100, false},
		{0, true},   // zero not allowed
		{-1, true},  // negative not allowed
		{101, true}, // too high
	}

	for _, tt := range tests {
		t.Run(string(rune(tt.shares)), func(t *testing.T) {
			err := validateShares(tt.shares)
			if (err != nil) != tt.expectErr {
				t.Errorf("validateShares(%d) error = %v, expectErr %v", tt.shares, err, tt.expectErr)
			}
		})
	}
}

func TestValidatePlayerCount(t *testing.T) {
	tests := []struct {
		count     int
		expectErr bool
	}{
		{2, false},
		{3, false},
		{4, false},
		{8, false},
		{1, true}, // too low
		{9, true}, // too high
	}

	for _, tt := range tests {
		t.Run(string(rune(tt.count)), func(t *testing.T) {
			err := validatePlayerCount(tt.count)
			if (err != nil) != tt.expectErr {
				t.Errorf("validatePlayerCount(%d) error = %v, expectErr %v", tt.count, err, tt.expectErr)
			}
		})
	}
}

func TestValidateStartingMoney(t *testing.T) {
	tests := []struct {
		amount    int
		expectErr bool
	}{
		{600, false},
		{1, false},
		{10000, false},
		{0, true},     // zero not allowed
		{-1, true},    // negative not allowed
		{10001, true}, // too high
	}

	for _, tt := range tests {
		t.Run(string(rune(tt.amount)), func(t *testing.T) {
			err := validateStartingMoney(tt.amount)
			if (err != nil) != tt.expectErr {
				t.Errorf("validateStartingMoney(%d) error = %v, expectErr %v", tt.amount, err, tt.expectErr)
			}
		})
	}
}

func TestValidateBankMoney(t *testing.T) {
	tests := []struct {
		amount    int
		expectErr bool
	}{
		{12000, false},
		{1, false},
		{1000000, false},
		{0, true},       // zero not allowed
		{-1, true},      // negative not allowed
		{1000001, true}, // too high
	}

	for _, tt := range tests {
		t.Run(string(rune(tt.amount)), func(t *testing.T) {
			err := validateBankMoney(tt.amount)
			if (err != nil) != tt.expectErr {
				t.Errorf("validateBankMoney(%d) error = %v, expectErr %v", tt.amount, err, tt.expectErr)
			}
		})
	}
}

func TestValidateCompanies(t *testing.T) {
	tests := []struct {
		name       string
		companies  []Company
		expectErrs int
	}{
		{
			name: "valid companies",
			companies: []Company{
				{ID: "BO", Name: "Baltimore & Ohio", Shares: 10},
				{ID: "PRR", Name: "Pennsylvania Railroad", Shares: 10},
			},
			expectErrs: 0,
		},
		{
			name:       "empty companies",
			companies:  []Company{},
			expectErrs: 1,
		},
		{
			name: "duplicate company IDs",
			companies: []Company{
				{ID: "BO", Name: "Baltimore & Ohio", Shares: 10},
				{ID: "BO", Name: "Boston & Maine", Shares: 10},
			},
			expectErrs: 1,
		},
		{
			name: "invalid company data",
			companies: []Company{
				{ID: "", Name: "Baltimore & Ohio", Shares: 10},      // empty ID
				{ID: "PRR", Name: "", Shares: 10},                  // empty name
				{ID: "NYC", Name: "New York Central", Shares: 0},   // zero shares
			},
			expectErrs: 3,
		},
		{
			name: "too many companies",
			companies: func() []Company {
				companies := make([]Company, 51)
				for i := range companies {
					companies[i] = Company{ID: fmt.Sprintf("C%d", i), Name: "Company", Shares: 10}
				}
				return companies
			}(),
			expectErrs: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := validateCompanies(tt.companies)
			if len(errors) != tt.expectErrs {
				t.Errorf("validateCompanies() got %d errors, expected %d", len(errors), tt.expectErrs)
				for _, err := range errors {
					t.Logf("Error: %s", err.Error())
				}
			}
		})
	}
}

func TestValidatePlayerConfigs(t *testing.T) {
	tests := []struct {
		name          string
		playerConfigs []PlayerConfig
		expectErrs    int
	}{
		{
			name: "valid player configs",
			playerConfigs: []PlayerConfig{
				{PlayerCount: 3, StartingMoney: 600, BankMoney: 12000},
				{PlayerCount: 4, StartingMoney: 600, BankMoney: 12000},
			},
			expectErrs: 0,
		},
		{
			name:          "empty player configs",
			playerConfigs: []PlayerConfig{},
			expectErrs:    1,
		},
		{
			name: "duplicate player counts",
			playerConfigs: []PlayerConfig{
				{PlayerCount: 3, StartingMoney: 600, BankMoney: 12000},
				{PlayerCount: 3, StartingMoney: 700, BankMoney: 15000},
			},
			expectErrs: 1,
		},
		{
			name: "invalid player config data",
			playerConfigs: []PlayerConfig{
				{PlayerCount: 1, StartingMoney: 600, BankMoney: 12000},   // invalid player count
				{PlayerCount: 3, StartingMoney: 0, BankMoney: 12000},     // invalid starting money
				{PlayerCount: 4, StartingMoney: 600, BankMoney: 0},       // invalid bank money
			},
			expectErrs: 3,
		},
		{
			name: "bank money too low",
			playerConfigs: []PlayerConfig{
				{PlayerCount: 4, StartingMoney: 600, BankMoney: 1000}, // bank money < starting money * player count
			},
			expectErrs: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := validatePlayerConfigs(tt.playerConfigs)
			if len(errors) != tt.expectErrs {
				t.Errorf("validatePlayerConfigs() got %d errors, expected %d", len(errors), tt.expectErrs)
				for _, err := range errors {
					t.Logf("Error: %s", err.Error())
				}
			}
		})
	}
}

func TestValidateGameConfig(t *testing.T) {
	validConfig := &GameConfig{
		Title:       "1830",
		Description: "Railroad game",
		Companies: []Company{
			{ID: "BO", Name: "Baltimore & Ohio", Shares: 10},
		},
		PlayerConfigs: []PlayerConfig{
			{PlayerCount: 3, StartingMoney: 600, BankMoney: 12000},
		},
		BankSharesPayCompany: true,
		IPOSharesPayCompany:  false,
	}

	errors := ValidateGameConfig(validConfig)
	if len(errors) != 0 {
		t.Errorf("Expected valid config to have no errors, got %d", len(errors))
		for _, err := range errors {
			t.Logf("Error: %s", err.Error())
		}
	}

	invalidConfig := &GameConfig{
		Title:         "", // invalid title
		Description:   strings.Repeat("a", 501), // invalid description
		Companies:     []Company{}, // invalid companies
		PlayerConfigs: []PlayerConfig{}, // invalid player configs
	}

	errors = ValidateGameConfig(invalidConfig)
	if len(errors) == 0 {
		t.Error("Expected invalid config to have errors")
	}
}

func TestValidateConfigFile(t *testing.T) {
	// Create a temporary directory for test files
	tempDir := t.TempDir()

	// Test valid config file
	validConfigContent := `
title: "Test Game"
description: "A test game"
companies:
  - id: "BO"
    name: "Baltimore & Ohio"
    shares: 10
player_configs:
  - player_count: 3
    starting_money: 600
    bank_money: 12000
bank_shares_pay_company: true
ipo_shares_pay_company: false
`

	validConfigPath := filepath.Join(tempDir, "valid.yaml")
	err := os.WriteFile(validConfigPath, []byte(validConfigContent), 0644)
	if err != nil {
		t.Fatal(err)
	}

	err = ValidateConfigFile(validConfigPath)
	if err != nil {
		t.Errorf("Expected valid config file to pass validation, got: %v", err)
	}

	// Test invalid YAML syntax
	invalidYAMLContent := `
title: "Test Game"
description: "A test game"
companies:
  - id: "BO"
    name: "Baltimore & Ohio"
    shares: 10
  invalid yaml here
`

	invalidYAMLPath := filepath.Join(tempDir, "invalid_yaml.yaml")
	err = os.WriteFile(invalidYAMLPath, []byte(invalidYAMLContent), 0644)
	if err != nil {
		t.Fatal(err)
	}

	err = ValidateConfigFile(invalidYAMLPath)
	if err == nil {
		t.Error("Expected invalid YAML to fail validation")
	}

	// Test invalid config content
	invalidConfigContent := `
title: ""
description: "A test game"
companies: []
player_configs: []
`

	invalidConfigPath := filepath.Join(tempDir, "invalid_config.yaml")
	err = os.WriteFile(invalidConfigPath, []byte(invalidConfigContent), 0644)
	if err != nil {
		t.Fatal(err)
	}

	err = ValidateConfigFile(invalidConfigPath)
	if err == nil {
		t.Error("Expected invalid config content to fail validation")
	}

	// Test non-existent file
	err = ValidateConfigFile("/nonexistent/file.yaml")
	if err == nil {
		t.Error("Expected non-existent file to fail validation")
	}
}

func TestValidationError(t *testing.T) {
	err := ValidationError{
		Field:   "test_field",
		Value:   "test_value",
		Message: "test message",
	}

	expected := "validation error in field 'test_field': test message (value: test_value)"
	if err.Error() != expected {
		t.Errorf("Expected error message %q, got %q", expected, err.Error())
	}
}

func TestIsPrintableString(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"Hello World", true},
		{"Hello\tWorld", true}, // tab allowed
		{"Hello\nWorld", true}, // newline allowed
		{"Hello\rWorld", true}, // carriage return allowed
		{"Hello\x00World", false}, // null character not allowed
		{"", true}, // empty string is printable
		{"!@#$%^&*()_+-={}[]|\\:;\"'<>?,./", true}, // all printable symbols
		{"Hello\x1FWorld", false}, // control character not allowed
		{"Hello\x7FWorld", false}, // DEL character not allowed
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := isPrintableString(tt.input)
			if result != tt.expected {
				t.Errorf("isPrintableString(%q) = %v, expected %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestValidateAndLoadGameConfig(t *testing.T) {
	// This test would require setting up actual config files
	// For now, we'll test that it fails gracefully with non-existent files
	_, err := ValidateAndLoadGameConfig("nonexistent_game")
	if err == nil {
		t.Error("Expected error when loading non-existent game config")
	}
}