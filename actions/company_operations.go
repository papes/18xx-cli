package actions

import (
	"fmt"
	"time"

	"github.com/papes/18xxCli/state"
)

// WithholdRevenue adds revenue to company treasury without paying dividends
func WithholdRevenue(gameState state.GameState, companyID string, amount state.Money) (state.GameState, error) {
	// Validate inputs
	if err := state.ValidateCompanyID(companyID); err != nil {
		return gameState, fmt.Errorf("invalid company ID: %w", err)
	}
	if err := state.ValidateMoney(amount); err != nil {
		return gameState, fmt.Errorf("invalid amount: %w", err)
	}
	if amount == 0 {
		return gameState, fmt.Errorf("revenue amount must be greater than zero")
	}

	newState := gameState.Clone()

	company, exists := newState.Companies[companyID]
	if !exists {
		return gameState, fmt.Errorf("company not found: %s", companyID)
	}

	// Add revenue to company treasury
	company.Treasury += amount

	// Add log entry
	logEntry := state.LogEntry{
		Timestamp: time.Now(),
		Action:    "withhold_revenue",
		Details: map[string]interface{}{
			"company": companyID,
			"amount":  amount,
		},
	}
	newState.Log = append(newState.Log, logEntry)

	return newState, nil
}

// PayDetailedDividend pays dividends with detailed per-player breakdown
func PayDetailedDividend(gameState state.GameState, companyID string, totalAmount state.Money) (state.GameState, error) {
	// Validate inputs
	if err := state.ValidateCompanyID(companyID); err != nil {
		return gameState, fmt.Errorf("invalid company ID: %w", err)
	}
	if err := state.ValidateMoney(totalAmount); err != nil {
		return gameState, fmt.Errorf("invalid dividend amount: %w", err)
	}
	if totalAmount == 0 {
		return gameState, fmt.Errorf("dividend amount must be greater than zero")
	}

	newState := gameState.Clone()

	company, exists := newState.Companies[companyID]
	if !exists {
		return gameState, fmt.Errorf("company not found: %s", companyID)
	}

	// Calculate total shares outstanding (owned by players)
	totalShares := state.ShareCount(0)
	playerPayouts := make(map[string]state.Money)

	for _, player := range newState.Players {
		shares := player.Shares[companyID]
		if shares > 0 {
			totalShares += shares
		}
	}

	if totalShares == 0 {
		return gameState, fmt.Errorf("no shares outstanding for company %s", companyID)
	}

	// Check if company has enough money
	if company.Treasury < totalAmount {
		return gameState, fmt.Errorf("company has insufficient funds for dividend: has %d, needs %d", company.Treasury, totalAmount)
	}

	// Calculate dividend per share
	dividendPerShare := totalAmount / state.Money(totalShares)

	// Pay dividends to each player and track amounts
	for id, player := range newState.Players {
		shares := player.Shares[companyID]
		if shares > 0 {
			dividend := state.Money(shares) * dividendPerShare
			player.Cash += dividend
			playerPayouts[id] = dividend
		}
	}

	// Reduce company treasury
	company.Treasury -= totalAmount

	// Log individual payments for each entity that receives money
	timestamp := time.Now()

	// Log payment to each player that owns shares
	for playerID, dividend := range playerPayouts {
		if dividend > 0 {
			shares := newState.Players[playerID].Shares[companyID]
			logEntry := state.LogEntry{
				Timestamp: timestamp,
				Action:    "dividend_payment",
				Details: map[string]interface{}{
					"company":            companyID,
					"recipient":          playerID,
					"recipient_type":     "player",
					"shares_owned":       shares,
					"dividend_per_share": dividendPerShare,
					"amount_received":    dividend,
					"total_dividend":     totalAmount,
				},
			}
			newState.Log = append(newState.Log, logEntry)
		}
	}

	// Add summary log entry for the overall dividend action
	summaryLogEntry := state.LogEntry{
		Timestamp: timestamp,
		Action:    "pay_detailed_dividend",
		Details: map[string]interface{}{
			"company":             companyID,
			"total_amount":        totalAmount,
			"dividend_per_share":  dividendPerShare,
			"total_shares":        totalShares,
			"player_payouts":      playerPayouts,
			"individual_payments": len(playerPayouts),
		},
	}
	newState.Log = append(newState.Log, summaryLogEntry)

	return newState, nil
}

// PayDividendWithConfig pays dividends according to game configuration settings
func PayDividendWithConfig(gameState state.GameState, companyID string, totalAmount state.Money, bankSharesPayCompany, ipoSharesPayCompany bool) (state.GameState, error) {
	newState := gameState.Clone()

	company, exists := newState.Companies[companyID]
	if !exists {
		return gameState, fmt.Errorf("company not found: %s", companyID)
	}

	// Calculate total shares outstanding (owned by players)
	totalPlayerShares := state.ShareCount(0)
	playerPayouts := make(map[string]state.Money)

	for _, player := range newState.Players {
		shares := player.Shares[companyID]
		if shares > 0 {
			totalPlayerShares += shares
		}
	}

	// Calculate total shares for dividend calculation (including bank/IPO if configured)
	totalSharesForDividend := totalPlayerShares
	if bankSharesPayCompany {
		totalSharesForDividend += company.SharesInBank
	}
	if ipoSharesPayCompany {
		totalSharesForDividend += company.SharesInIPO
	}

	if totalSharesForDividend == 0 {
		return gameState, fmt.Errorf("no shares for dividend calculation for company %s", companyID)
	}

	// Calculate dividend per share
	dividendPerShare := totalAmount / state.Money(totalSharesForDividend)

	// Pay dividends to players (always happens regardless of settings)
	totalPlayerPayout := state.Money(0)
	for id, player := range newState.Players {
		shares := player.Shares[companyID]
		if shares > 0 {
			dividend := state.Money(shares) * dividendPerShare
			player.Cash += dividend
			playerPayouts[id] = dividend
			totalPlayerPayout += dividend
		}
	}

	// Calculate company self-payout from bank/IPO shares
	companySelfPayout := state.Money(0)
	if bankSharesPayCompany && company.SharesInBank > 0 {
		companySelfPayout += state.Money(company.SharesInBank) * dividendPerShare
	}
	if ipoSharesPayCompany && company.SharesInIPO > 0 {
		companySelfPayout += state.Money(company.SharesInIPO) * dividendPerShare
	}

	// Check if company has enough money for total payout
	totalPayout := totalPlayerPayout + companySelfPayout
	if company.Treasury < totalPayout {
		return gameState, fmt.Errorf("company has insufficient funds for dividend: has %d, needs %d", company.Treasury, totalPayout)
	}

	// Execute payouts
	company.Treasury -= totalPlayerPayout // Money paid to players goes out of company
	// Company self-payout (from bank/IPO shares) stays in company treasury (company keeps its own dividend)

	// Log individual payments for each entity that receives money
	timestamp := time.Now()

	// Log payment to each player that owns shares
	for playerID, dividend := range playerPayouts {
		if dividend > 0 {
			shares := newState.Players[playerID].Shares[companyID]
			logEntry := state.LogEntry{
				Timestamp: timestamp,
				Action:    "dividend_payment",
				Details: map[string]interface{}{
					"company":            companyID,
					"recipient":          playerID,
					"recipient_type":     "player",
					"shares_owned":       shares,
					"dividend_per_share": dividendPerShare,
					"amount_received":    dividend,
					"total_dividend":     totalAmount,
				},
			}
			newState.Log = append(newState.Log, logEntry)
		}
	}

	// Log company self-payout if applicable
	if companySelfPayout > 0 {
		selfPayoutDetails := make(map[string]interface{})
		selfPayoutDetails["company"] = companyID
		selfPayoutDetails["recipient"] = companyID
		selfPayoutDetails["recipient_type"] = "company"
		selfPayoutDetails["dividend_per_share"] = dividendPerShare
		selfPayoutDetails["amount_received"] = companySelfPayout
		selfPayoutDetails["total_dividend"] = totalAmount

		if bankSharesPayCompany && company.SharesInBank > 0 {
			selfPayoutDetails["bank_shares"] = company.SharesInBank
			selfPayoutDetails["bank_share_payout"] = state.Money(company.SharesInBank) * dividendPerShare
		}
		if ipoSharesPayCompany && company.SharesInIPO > 0 {
			selfPayoutDetails["ipo_shares"] = company.SharesInIPO
			selfPayoutDetails["ipo_share_payout"] = state.Money(company.SharesInIPO) * dividendPerShare
		}

		logEntry := state.LogEntry{
			Timestamp: timestamp,
			Action:    "dividend_payment",
			Details:   selfPayoutDetails,
		}
		newState.Log = append(newState.Log, logEntry)
	}

	// Add summary log entry for the overall dividend action
	summaryLogEntry := state.LogEntry{
		Timestamp: timestamp,
		Action:    "pay_dividend_with_config",
		Details: map[string]interface{}{
			"company":             companyID,
			"total_amount":        totalAmount,
			"dividend_per_share":  dividendPerShare,
			"total_shares":        totalSharesForDividend,
			"player_shares":       totalPlayerShares,
			"bank_shares":         company.SharesInBank,
			"ipo_shares":          company.SharesInIPO,
			"bank_shares_pay":     bankSharesPayCompany,
			"ipo_shares_pay":      ipoSharesPayCompany,
			"player_payouts":      playerPayouts,
			"company_self_payout": companySelfPayout,
			"total_player_payout": totalPlayerPayout,
			"individual_payments": len(playerPayouts) + (func() int {
				if companySelfPayout > 0 {
					return 1
				} else {
					return 0
				}
			})(),
		},
	}
	newState.Log = append(newState.Log, summaryLogEntry)

	return newState, nil
}

// BuyTrain purchases a train for a company
func BuyTrain(gameState state.GameState, companyID string, trainType string, cost state.Money) (state.GameState, error) {
	newState := gameState.Clone()

	company, exists := newState.Companies[companyID]
	if !exists {
		return gameState, fmt.Errorf("company not found: %s", companyID)
	}

	// Check if company has enough money
	if company.Treasury < cost {
		return gameState, fmt.Errorf("company has insufficient funds for train: has %d, needs %d", company.Treasury, cost)
	}

	// Deduct cost from company treasury
	company.Treasury -= cost

	// Add log entry
	logEntry := state.LogEntry{
		Timestamp: time.Now(),
		Action:    "buy_train",
		Details: map[string]interface{}{
			"company":    companyID,
			"train_type": trainType,
			"cost":       cost,
		},
	}
	newState.Log = append(newState.Log, logEntry)

	return newState, nil
}

// FloatCompany handles company floating (when 60% shares are sold)
func FloatCompany(gameState state.GameState, companyID string) (state.GameState, error) {
	newState := gameState.Clone()

	company, exists := newState.Companies[companyID]
	if !exists {
		return gameState, fmt.Errorf("company not found: %s", companyID)
	}

	// Calculate IPO proceeds (shares sold * par value)
	sharesSold := company.TotalShares - company.SharesInIPO
	proceeds := state.Money(sharesSold) * state.Money(company.ParValue)

	// Add proceeds to company treasury
	company.Treasury += proceeds

	// Add log entry
	logEntry := state.LogEntry{
		Timestamp: time.Now(),
		Action:    "float_company",
		Details: map[string]interface{}{
			"company":     companyID,
			"shares_sold": sharesSold,
			"proceeds":    proceeds,
		},
	}
	newState.Log = append(newState.Log, logEntry)

	return newState, nil
}

// SetPresident changes the president of a company
func SetPresident(gameState state.GameState, companyID, newPresidentID string) (state.GameState, error) {
	newState := gameState.Clone()

	company, exists := newState.Companies[companyID]
	if !exists {
		return gameState, fmt.Errorf("company not found: %s", companyID)
	}

	_, exists = newState.Players[newPresidentID]
	if !exists {
		return gameState, fmt.Errorf("player not found: %s", newPresidentID)
	}

	oldPresidentID := company.PresidentID
	company.PresidentID = newPresidentID

	// Add log entry
	logEntry := state.LogEntry{
		Timestamp: time.Now(),
		Action:    "set_president",
		Details: map[string]interface{}{
			"company":       companyID,
			"old_president": oldPresidentID,
			"new_president": newPresidentID,
		},
	}
	newState.Log = append(newState.Log, logEntry)

	return newState, nil
}

// CompanyRevenue adds revenue to a company (for operating rounds)
func CompanyRevenue(gameState state.GameState, companyID string, amount state.Money) (state.GameState, error) {
	newState := gameState.Clone()

	company, exists := newState.Companies[companyID]
	if !exists {
		return gameState, fmt.Errorf("company not found: %s", companyID)
	}

	// Add revenue to company treasury
	company.Treasury += amount

	// Add log entry
	logEntry := state.LogEntry{
		Timestamp: time.Now(),
		Action:    "company_revenue",
		Details: map[string]interface{}{
			"company": companyID,
			"amount":  amount,
		},
	}
	newState.Log = append(newState.Log, logEntry)

	return newState, nil
}
