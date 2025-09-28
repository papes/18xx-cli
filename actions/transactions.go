package actions

import (
	"fmt"
	"time"

	"github.com/papes/18xxCli/state"
)

// BuyShareFromIPO handles purchasing shares from a company's IPO
func BuyShareFromIPO(gameState state.GameState, playerID, companyID string, shares state.ShareCount) (state.GameState, error) {
	// Validate inputs
	if err := state.ValidatePlayerID(playerID); err != nil {
		return gameState, fmt.Errorf("invalid player ID: %w", err)
	}
	if err := state.ValidateCompanyID(companyID); err != nil {
		return gameState, fmt.Errorf("invalid company ID: %w", err)
	}
	if err := state.ValidateShareCount(shares); err != nil {
		return gameState, fmt.Errorf("invalid share count: %w", err)
	}

	newState := gameState.Clone()

	player, exists := newState.Players[playerID]
	if !exists {
		return gameState, fmt.Errorf("player not found: %s", playerID)
	}

	company, exists := newState.Companies[companyID]
	if !exists {
		return gameState, fmt.Errorf("company not found: %s", companyID)
	}

	// Check if par value is set
	if company.ParValue == 0 {
		return gameState, fmt.Errorf("company par value not set")
	}

	// Calculate cost
	cost := state.Money(company.ParValue) * state.Money(shares)

	// Validate the share transaction
	if err := state.ValidateShareTransaction(shares, company.SharesInIPO, cost, player.Cash); err != nil {
		return gameState, fmt.Errorf("invalid share transaction: %w", err)
	}

	// Execute transaction
	player.Cash -= cost
	player.Shares[companyID] += shares
	company.SharesInIPO -= shares
	company.Treasury += cost

	// Add log entry
	logEntry := state.LogEntry{
		Timestamp: time.Now(),
		Action:    "buy_share_ipo",
		Details: map[string]interface{}{
			"player":  playerID,
			"company": companyID,
			"shares":  shares,
			"price":   company.ParValue,
			"total":   cost,
		},
	}
	newState.Log = append(newState.Log, logEntry)

	return newState, nil
}

// SellShareToBankPool handles selling shares to the bank pool
func SellShareToBankPool(gameState state.GameState, playerID, companyID string, shares state.ShareCount) (state.GameState, error) {
	// Validate inputs
	if err := state.ValidatePlayerID(playerID); err != nil {
		return gameState, fmt.Errorf("invalid player ID: %w", err)
	}
	if err := state.ValidateCompanyID(companyID); err != nil {
		return gameState, fmt.Errorf("invalid company ID: %w", err)
	}
	if err := state.ValidateShareCount(shares); err != nil {
		return gameState, fmt.Errorf("invalid share count: %w", err)
	}

	newState := gameState.Clone()

	player, exists := newState.Players[playerID]
	if !exists {
		return gameState, fmt.Errorf("player not found: %s", playerID)
	}

	company, exists := newState.Companies[companyID]
	if !exists {
		return gameState, fmt.Errorf("company not found: %s", companyID)
	}

	// Validate player owns enough shares
	if err := state.ValidateShareCount(player.Shares[companyID]); err != nil {
		return gameState, fmt.Errorf("invalid player shares: %w", err)
	}
	if player.Shares[companyID] < shares {
		return gameState, fmt.Errorf("player does not own enough shares: has %d, trying to sell %d", player.Shares[companyID], shares)
	}

	// Use current stock price (not par value) for bank sales
	price := company.StockPrice
	if price == 0 {
		// If no stock price set, use par value
		price = company.ParValue
	}

	// Calculate payment
	payment := state.Money(price) * state.Money(shares)

	// Validate payment amount
	if err := state.ValidateMoney(payment); err != nil {
		return gameState, fmt.Errorf("invalid payment amount: %w", err)
	}

	// Check if bank has enough money
	if newState.BankMoney < payment {
		return gameState, fmt.Errorf("bank has insufficient funds: has %d, needs %d", newState.BankMoney, payment)
	}

	// Execute transaction
	player.Cash += payment
	player.Shares[companyID] -= shares
	company.SharesInBank += shares
	newState.BankMoney -= payment

	// Add log entry
	logEntry := state.LogEntry{
		Timestamp: time.Now(),
		Action:    "sell_share_bank",
		Details: map[string]interface{}{
			"player":  playerID,
			"company": companyID,
			"shares":  shares,
			"price":   price,
			"total":   payment,
		},
	}
	newState.Log = append(newState.Log, logEntry)

	return newState, nil
}

// BuyShareFromBankPool handles purchasing shares from the bank pool
func BuyShareFromBankPool(gameState state.GameState, playerID, companyID string, shares state.ShareCount) (state.GameState, error) {
	// Validate inputs
	if err := state.ValidatePlayerID(playerID); err != nil {
		return gameState, fmt.Errorf("invalid player ID: %w", err)
	}
	if err := state.ValidateCompanyID(companyID); err != nil {
		return gameState, fmt.Errorf("invalid company ID: %w", err)
	}
	if err := state.ValidateShareCount(shares); err != nil {
		return gameState, fmt.Errorf("invalid share count: %w", err)
	}

	newState := gameState.Clone()

	player, exists := newState.Players[playerID]
	if !exists {
		return gameState, fmt.Errorf("player not found: %s", playerID)
	}

	company, exists := newState.Companies[companyID]
	if !exists {
		return gameState, fmt.Errorf("company not found: %s", companyID)
	}

	// Use current stock price for bank purchases
	price := company.StockPrice
	if price == 0 {
		return gameState, fmt.Errorf("stock price not set")
	}

	// Calculate cost
	cost := state.Money(price) * state.Money(shares)

	// Validate the share transaction
	if err := state.ValidateShareTransaction(shares, company.SharesInBank, cost, player.Cash); err != nil {
		return gameState, fmt.Errorf("invalid share transaction: %w", err)
	}

	// Execute transaction
	player.Cash -= cost
	player.Shares[companyID] += shares
	company.SharesInBank -= shares
	newState.BankMoney += cost

	// Add log entry
	logEntry := state.LogEntry{
		Timestamp: time.Now(),
		Action:    "buy_share_bank",
		Details: map[string]interface{}{
			"player":  playerID,
			"company": companyID,
			"shares":  shares,
			"price":   price,
			"total":   cost,
		},
	}
	newState.Log = append(newState.Log, logEntry)

	return newState, nil
}

// SetParValue sets the par value for a company (usually done at IPO)
func SetParValue(gameState state.GameState, companyID string, parValue state.StockPrice) (state.GameState, error) {
	// Validate inputs
	if err := state.ValidateCompanyID(companyID); err != nil {
		return gameState, fmt.Errorf("invalid company ID: %w", err)
	}
	if err := state.ValidateStockPrice(parValue); err != nil {
		return gameState, fmt.Errorf("invalid par value: %w", err)
	}
	if parValue == 0 {
		return gameState, fmt.Errorf("par value must be greater than zero")
	}

	newState := gameState.Clone()

	company, exists := newState.Companies[companyID]
	if !exists {
		return gameState, fmt.Errorf("company not found: %s", companyID)
	}

	// Check if par value already set
	if company.ParValue != 0 {
		return gameState, fmt.Errorf("par value already set")
	}

	// Set par value and initial stock price
	company.ParValue = parValue
	company.StockPrice = parValue // Stock price starts at par

	// Add log entry
	logEntry := state.LogEntry{
		Timestamp: time.Now(),
		Action:    "set_par_value",
		Details: map[string]interface{}{
			"company":   companyID,
			"par_value": parValue,
		},
	}
	newState.Log = append(newState.Log, logEntry)

	return newState, nil
}

// SetStockPrice manually adjusts a company's stock price
func SetStockPrice(gameState state.GameState, companyID string, price state.StockPrice) (state.GameState, error) {
	// Validate inputs
	if err := state.ValidateCompanyID(companyID); err != nil {
		return gameState, fmt.Errorf("invalid company ID: %w", err)
	}
	if err := state.ValidateStockPrice(price); err != nil {
		return gameState, fmt.Errorf("invalid stock price: %w", err)
	}

	newState := gameState.Clone()

	company, exists := newState.Companies[companyID]
	if !exists {
		return gameState, fmt.Errorf("company not found: %s", companyID)
	}

	oldPrice := company.StockPrice
	company.StockPrice = price

	// Add log entry
	logEntry := state.LogEntry{
		Timestamp: time.Now(),
		Action:    "set_stock_price",
		Details: map[string]interface{}{
			"company":   companyID,
			"old_price": oldPrice,
			"new_price": price,
		},
	}
	newState.Log = append(newState.Log, logEntry)

	return newState, nil
}

// PayDividend pays dividends to all shareholders of a company
func PayDividend(gameState state.GameState, companyID string, dividendPerShare state.Money) (state.GameState, error) {
	// Validate inputs
	if err := state.ValidateCompanyID(companyID); err != nil {
		return gameState, fmt.Errorf("invalid company ID: %w", err)
	}
	if err := state.ValidateMoney(dividendPerShare); err != nil {
		return gameState, fmt.Errorf("invalid dividend per share: %w", err)
	}
	if dividendPerShare == 0 {
		return gameState, fmt.Errorf("dividend per share must be greater than zero")
	}

	newState := gameState.Clone()

	company, exists := newState.Companies[companyID]
	if !exists {
		return gameState, fmt.Errorf("company not found: %s", companyID)
	}

	// Calculate total dividend payout
	totalShares := state.ShareCount(0)
	for _, player := range newState.Players {
		totalShares += player.Shares[companyID]
	}

	totalPayout := state.Money(totalShares) * dividendPerShare

	// Validate dividend amount
	if err := state.ValidateDividendAmount(totalPayout, company.Treasury); err != nil {
		return gameState, fmt.Errorf("invalid dividend: %w", err)
	}

	// Pay dividends to each player
	for _, player := range newState.Players {
		playerShares := player.Shares[companyID]
		if playerShares > 0 {
			dividend := state.Money(playerShares) * dividendPerShare
			player.Cash += dividend
		}
	}

	// Reduce company treasury
	company.Treasury -= totalPayout

	// Add log entry
	logEntry := state.LogEntry{
		Timestamp: time.Now(),
		Action:    "pay_dividend",
		Details: map[string]interface{}{
			"company":            companyID,
			"dividend_per_share": dividendPerShare,
			"total_shares":       totalShares,
			"total_payout":       totalPayout,
		},
	}
	newState.Log = append(newState.Log, logEntry)

	return newState, nil
}

// TransferMoney transfers money between players or between player and company
func TransferMoney(gameState state.GameState, fromID, toID string, amount state.Money) (state.GameState, error) {
	// Validate inputs
	if err := state.ValidateMoney(amount); err != nil {
		return gameState, fmt.Errorf("invalid transfer amount: %w", err)
	}
	if amount == 0 {
		return gameState, fmt.Errorf("transfer amount must be greater than zero")
	}
	if fromID == "" || toID == "" {
		return gameState, fmt.Errorf("from and to IDs cannot be empty")
	}
	if fromID == toID {
		return gameState, fmt.Errorf("cannot transfer money to self")
	}

	newState := gameState.Clone()

	// Handle player-to-player transfers
	fromPlayer, fromIsPlayer := newState.Players[fromID]
	toPlayer, toIsPlayer := newState.Players[toID]

	if fromIsPlayer && toIsPlayer {
		// Player to player transfer
		if err := state.ValidateTransferAmount(amount, fromPlayer.Cash); err != nil {
			return gameState, fmt.Errorf("player transfer failed: %w", err)
		}

		fromPlayer.Cash -= amount
		toPlayer.Cash += amount

		logEntry := state.LogEntry{
			Timestamp: time.Now(),
			Action:    "transfer_money",
			Details: map[string]interface{}{
				"from":   fromID,
				"to":     toID,
				"amount": amount,
				"type":   "player_to_player",
			},
		}
		newState.Log = append(newState.Log, logEntry)

		return newState, nil
	}

	// Handle company-to-player or player-to-company transfers
	fromCompany, fromIsCompany := newState.Companies[fromID]
	toCompany, toIsCompany := newState.Companies[toID]

	if fromIsPlayer && toIsCompany {
		// Player to company
		if err := state.ValidateTransferAmount(amount, fromPlayer.Cash); err != nil {
			return gameState, fmt.Errorf("player to company transfer failed: %w", err)
		}

		fromPlayer.Cash -= amount
		toCompany.Treasury += amount

	} else if fromIsCompany && toIsPlayer {
		// Company to player
		if err := state.ValidateTransferAmount(amount, fromCompany.Treasury); err != nil {
			return gameState, fmt.Errorf("company to player transfer failed: %w", err)
		}

		fromCompany.Treasury -= amount
		toPlayer.Cash += amount

	} else if fromIsCompany && toIsCompany {
		// Company to company
		if err := state.ValidateTransferAmount(amount, fromCompany.Treasury); err != nil {
			return gameState, fmt.Errorf("company to company transfer failed: %w", err)
		}

		fromCompany.Treasury -= amount
		toCompany.Treasury += amount

	} else {
		return gameState, fmt.Errorf("invalid transfer: %s to %s", fromID, toID)
	}

	logEntry := state.LogEntry{
		Timestamp: time.Now(),
		Action:    "transfer_money",
		Details: map[string]interface{}{
			"from":   fromID,
			"to":     toID,
			"amount": amount,
		},
	}
	newState.Log = append(newState.Log, logEntry)

	return newState, nil
}
