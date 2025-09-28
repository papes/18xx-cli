package state

import (
	"fmt"
)

// ValidateMoney ensures money values are non-negative and within reasonable bounds
func ValidateMoney(amount Money) error {
	if amount < 0 {
		return fmt.Errorf("money amount cannot be negative: %d", amount)
	}
	// Maximum reasonable amount to prevent overflow issues
	const maxMoney = Money(1000000) // 1 million
	if amount > maxMoney {
		return fmt.Errorf("money amount too large: %d (max: %d)", amount, maxMoney)
	}
	return nil
}

// ValidateShareCount ensures share counts are non-negative and within reasonable bounds
func ValidateShareCount(shares ShareCount) error {
	if shares < 0 {
		return fmt.Errorf("share count cannot be negative: %d", shares)
	}
	// Maximum shares for any company (typically 10-20)
	const maxShares = ShareCount(100)
	if shares > maxShares {
		return fmt.Errorf("share count too large: %d (max: %d)", shares, maxShares)
	}
	return nil
}

// ValidateStockPrice ensures stock prices are non-negative and within reasonable bounds
func ValidateStockPrice(price StockPrice) error {
	if price < 0 {
		return fmt.Errorf("stock price cannot be negative: %d", price)
	}
	// Maximum reasonable stock price
	const maxPrice = StockPrice(1000)
	if price > maxPrice {
		return fmt.Errorf("stock price too large: %d (max: %d)", price, maxPrice)
	}
	return nil
}

// ValidatePlayerID ensures player IDs are valid
func ValidatePlayerID(playerID string) error {
	if playerID == "" {
		return fmt.Errorf("player ID cannot be empty")
	}
	if len(playerID) > 50 {
		return fmt.Errorf("player ID too long: %d characters (max: 50)", len(playerID))
	}
	return nil
}

// ValidateCompanyID ensures company IDs are valid
func ValidateCompanyID(companyID string) error {
	if companyID == "" {
		return fmt.Errorf("company ID cannot be empty")
	}
	if len(companyID) > 20 {
		return fmt.Errorf("company ID too long: %d characters (max: 20)", len(companyID))
	}
	return nil
}

// ValidatePlayerName ensures player names are valid
func ValidatePlayerName(name string) error {
	if name == "" {
		return fmt.Errorf("player name cannot be empty")
	}
	if len(name) > 100 {
		return fmt.Errorf("player name too long: %d characters (max: 100)", len(name))
	}
	return nil
}

// ValidateCompanyName ensures company names are valid
func ValidateCompanyName(name string) error {
	if name == "" {
		return fmt.Errorf("company name cannot be empty")
	}
	if len(name) > 100 {
		return fmt.Errorf("company name too long: %d characters (max: 100)", len(name))
	}
	return nil
}

// ValidateTransferAmount validates money transfer amounts
func ValidateTransferAmount(amount Money, available Money) error {
	if err := ValidateMoney(amount); err != nil {
		return err
	}
	if amount == 0 {
		return fmt.Errorf("transfer amount must be greater than zero")
	}
	if amount > available {
		return fmt.Errorf("insufficient funds: requested %d, available %d", amount, available)
	}
	return nil
}

// ValidateDividendAmount validates dividend payment amounts
func ValidateDividendAmount(amount Money, companyTreasury Money) error {
	if err := ValidateMoney(amount); err != nil {
		return err
	}
	if amount == 0 {
		return fmt.Errorf("dividend amount must be greater than zero")
	}
	if amount > companyTreasury {
		return fmt.Errorf("company has insufficient funds for dividend: requested %d, available %d", amount, companyTreasury)
	}
	return nil
}

// ValidateShareTransaction validates share buy/sell transactions
func ValidateShareTransaction(shares ShareCount, available ShareCount, cost Money, playerCash Money) error {
	if err := ValidateShareCount(shares); err != nil {
		return err
	}
	if shares == 0 {
		return fmt.Errorf("share count must be greater than zero")
	}
	if shares > available {
		return fmt.Errorf("insufficient shares available: requested %d, available %d", shares, available)
	}
	if err := ValidateMoney(cost); err != nil {
		return err
	}
	if cost > playerCash {
		return fmt.Errorf("insufficient funds for share purchase: cost %d, available %d", cost, playerCash)
	}
	return nil
}