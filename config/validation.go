package config

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// ValidationError represents a configuration validation error
type ValidationError struct {
	Field   string
	Value   interface{}
	Message string
}

// Error implements the error interface
func (e ValidationError) Error() string {
	return fmt.Sprintf("validation error in field '%s': %s (value: %v)", e.Field, e.Message, e.Value)
}

// ValidateGameConfig validates a complete game configuration
func ValidateGameConfig(config *GameConfig) []ValidationError {
	var errors []ValidationError

	// Validate basic fields
	if err := validateTitle(config.Title); err != nil {
		errors = append(errors, ValidationError{
			Field:   "title",
			Value:   config.Title,
			Message: err.Error(),
		})
	}

	if err := validateDescription(config.Description); err != nil {
		errors = append(errors, ValidationError{
			Field:   "description",
			Value:   config.Description,
			Message: err.Error(),
		})
	}

	// Validate companies
	companyErrors := validateCompanies(config.Companies)
	errors = append(errors, companyErrors...)

	// Validate player configs
	playerConfigErrors := validatePlayerConfigs(config.PlayerConfigs)
	errors = append(errors, playerConfigErrors...)

	return errors
}

// validateTitle validates the game title
func validateTitle(title string) error {
	if title == "" {
		return fmt.Errorf("title cannot be empty")
	}
	if len(title) > 100 {
		return fmt.Errorf("title too long (max 100 characters)")
	}
	// Title should contain printable characters only
	if !isPrintableString(title) {
		return fmt.Errorf("title contains invalid characters")
	}
	return nil
}

// validateDescription validates the game description
func validateDescription(description string) error {
	if len(description) > 500 {
		return fmt.Errorf("description too long (max 500 characters)")
	}
	// Description can be empty, but if present should be printable
	if description != "" && !isPrintableString(description) {
		return fmt.Errorf("description contains invalid characters")
	}
	return nil
}

// validateCompanies validates the companies array
func validateCompanies(companies []Company) []ValidationError {
	var errors []ValidationError

	if len(companies) == 0 {
		errors = append(errors, ValidationError{
			Field:   "companies",
			Value:   len(companies),
			Message: "at least one company must be defined",
		})
		return errors
	}

	if len(companies) > 50 {
		errors = append(errors, ValidationError{
			Field:   "companies",
			Value:   len(companies),
			Message: "too many companies (max 50)",
		})
	}

	// Track company IDs to ensure uniqueness
	companyIDs := make(map[string]bool)

	for i, company := range companies {
		fieldPrefix := fmt.Sprintf("companies[%d]", i)

		// Validate company ID
		if err := validateCompanyID(company.ID); err != nil {
			errors = append(errors, ValidationError{
				Field:   fieldPrefix + ".id",
				Value:   company.ID,
				Message: err.Error(),
			})
		} else {
			// Check for duplicate IDs
			if companyIDs[company.ID] {
				errors = append(errors, ValidationError{
					Field:   fieldPrefix + ".id",
					Value:   company.ID,
					Message: "duplicate company ID",
				})
			}
			companyIDs[company.ID] = true
		}

		// Validate company name
		if err := validateCompanyName(company.Name); err != nil {
			errors = append(errors, ValidationError{
				Field:   fieldPrefix + ".name",
				Value:   company.Name,
				Message: err.Error(),
			})
		}

		// Validate shares
		if err := validateShares(company.Shares); err != nil {
			errors = append(errors, ValidationError{
				Field:   fieldPrefix + ".shares",
				Value:   company.Shares,
				Message: err.Error(),
			})
		}
	}

	return errors
}

// validatePlayerConfigs validates the player configurations array
func validatePlayerConfigs(playerConfigs []PlayerConfig) []ValidationError {
	var errors []ValidationError

	if len(playerConfigs) == 0 {
		errors = append(errors, ValidationError{
			Field:   "player_configs",
			Value:   len(playerConfigs),
			Message: "at least one player configuration must be defined",
		})
		return errors
	}

	if len(playerConfigs) > 10 {
		errors = append(errors, ValidationError{
			Field:   "player_configs",
			Value:   len(playerConfigs),
			Message: "too many player configurations (max 10)",
		})
	}

	// Track player counts to ensure uniqueness
	playerCounts := make(map[int]bool)

	for i, config := range playerConfigs {
		fieldPrefix := fmt.Sprintf("player_configs[%d]", i)

		// Validate player count
		if err := validatePlayerCount(config.PlayerCount); err != nil {
			errors = append(errors, ValidationError{
				Field:   fieldPrefix + ".player_count",
				Value:   config.PlayerCount,
				Message: err.Error(),
			})
		} else {
			// Check for duplicate player counts
			if playerCounts[config.PlayerCount] {
				errors = append(errors, ValidationError{
					Field:   fieldPrefix + ".player_count",
					Value:   config.PlayerCount,
					Message: "duplicate player count configuration",
				})
			}
			playerCounts[config.PlayerCount] = true
		}

		// Validate starting money
		if err := validateStartingMoney(config.StartingMoney); err != nil {
			errors = append(errors, ValidationError{
				Field:   fieldPrefix + ".starting_money",
				Value:   config.StartingMoney,
				Message: err.Error(),
			})
		}

		// Validate bank money
		if err := validateBankMoney(config.BankMoney); err != nil {
			errors = append(errors, ValidationError{
				Field:   fieldPrefix + ".bank_money",
				Value:   config.BankMoney,
				Message: err.Error(),
			})
		}

		// Cross-validation: bank money should be much larger than starting money
		if config.BankMoney > 0 && config.StartingMoney > 0 {
			if config.BankMoney < config.StartingMoney*config.PlayerCount {
				errors = append(errors, ValidationError{
					Field:   fieldPrefix,
					Value:   fmt.Sprintf("bank_money:%d, starting_money:%d, player_count:%d", config.BankMoney, config.StartingMoney, config.PlayerCount),
					Message: "bank money should be at least starting_money * player_count",
				})
			}
		}
	}

	return errors
}

// validateCompanyID validates a company ID
func validateCompanyID(id string) error {
	if id == "" {
		return fmt.Errorf("company ID cannot be empty")
	}
	if len(id) > 10 {
		return fmt.Errorf("company ID too long (max 10 characters)")
	}
	// Company ID should be alphanumeric with optional hyphens/underscores
	matched, _ := regexp.MatchString("^[a-zA-Z0-9_-]+$", id)
	if !matched {
		return fmt.Errorf("company ID must contain only alphanumeric characters, hyphens, and underscores")
	}
	return nil
}

// validateCompanyName validates a company name
func validateCompanyName(name string) error {
	if name == "" {
		return fmt.Errorf("company name cannot be empty")
	}
	if len(name) > 100 {
		return fmt.Errorf("company name too long (max 100 characters)")
	}
	if !isPrintableString(name) {
		return fmt.Errorf("company name contains invalid characters")
	}
	return nil
}

// validateShares validates the number of shares
func validateShares(shares int) error {
	if shares <= 0 {
		return fmt.Errorf("shares must be positive")
	}
	if shares > 100 {
		return fmt.Errorf("shares too high (max 100)")
	}
	return nil
}

// validatePlayerCount validates the player count
func validatePlayerCount(count int) error {
	if count < 2 {
		return fmt.Errorf("player count must be at least 2")
	}
	if count > 8 {
		return fmt.Errorf("player count too high (max 8)")
	}
	return nil
}

// validateStartingMoney validates starting money amount
func validateStartingMoney(amount int) error {
	if amount <= 0 {
		return fmt.Errorf("starting money must be positive")
	}
	if amount > 10000 {
		return fmt.Errorf("starting money too high (max 10000)")
	}
	return nil
}

// validateBankMoney validates bank money amount
func validateBankMoney(amount int) error {
	if amount <= 0 {
		return fmt.Errorf("bank money must be positive")
	}
	if amount > 1000000 {
		return fmt.Errorf("bank money too high (max 1000000)")
	}
	return nil
}

// isPrintableString checks if a string contains only printable characters
func isPrintableString(s string) bool {
	for _, r := range s {
		if r < 32 || r > 126 {
			// Allow common whitespace characters
			if r != ' ' && r != '\t' && r != '\n' && r != '\r' {
				return false
			}
		}
	}
	return true
}

// ValidateConfigFile validates a YAML config file without loading it into memory
func ValidateConfigFile(filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	// Check for basic YAML syntax errors
	var temp interface{}
	if err := yaml.Unmarshal(data, &temp); err != nil {
		return fmt.Errorf("invalid YAML syntax: %w", err)
	}

	// Try to unmarshal into our GameConfig structure
	var config GameConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to parse config structure: %w", err)
	}

	// Validate the configuration
	errors := ValidateGameConfig(&config)
	if len(errors) > 0 {
		var errorMessages []string
		for _, err := range errors {
			errorMessages = append(errorMessages, err.Error())
		}
		return fmt.Errorf("validation errors:\n%s", strings.Join(errorMessages, "\n"))
	}

	return nil
}

// ValidateAndLoadGameConfig validates and loads a game configuration
func ValidateAndLoadGameConfig(gameName string) (*GameConfig, error) {
	configPath := filepath.Join("config", "games", gameName+".yaml")

	// First validate the file
	if err := ValidateConfigFile(configPath); err != nil {
		return nil, err
	}

	// Then load it (we know it's valid now)
	return LoadGameConfig(gameName)
}