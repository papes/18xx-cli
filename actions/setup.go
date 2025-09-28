package actions

import (
	"fmt"
	"time"

	"github.com/papes/18xxCli/state"
)

// AddPlayer safely adds a new player to the game with validation
func AddPlayer(gameState state.GameState, id, name string, startingCash state.Money) (state.GameState, error) {
	// Validate inputs
	if err := state.ValidatePlayerID(id); err != nil {
		return gameState, fmt.Errorf("invalid player ID: %w", err)
	}
	if err := state.ValidatePlayerName(name); err != nil {
		return gameState, fmt.Errorf("invalid player name: %w", err)
	}
	if err := state.ValidateMoney(startingCash); err != nil {
		return gameState, fmt.Errorf("invalid starting cash: %w", err)
	}

	// Check if player already exists
	if _, exists := gameState.Players[id]; exists {
		return gameState, fmt.Errorf("player with ID %s already exists", id)
	}

	// Add player to state
	newState := gameState.AddPlayer(id, name, startingCash)

	// Add log entry
	logEntry := state.LogEntry{
		Timestamp: time.Now(),
		Action:    "add_player",
		Details: map[string]interface{}{
			"player_id":      id,
			"player_name":    name,
			"starting_cash":  startingCash,
		},
	}
	newState.Log = append(newState.Log, logEntry)

	return newState, nil
}

// AddCompany safely adds a new company to the game with validation
func AddCompany(gameState state.GameState, id, name string, totalShares state.ShareCount) (state.GameState, error) {
	// Validate inputs
	if err := state.ValidateCompanyID(id); err != nil {
		return gameState, fmt.Errorf("invalid company ID: %w", err)
	}
	if err := state.ValidateCompanyName(name); err != nil {
		return gameState, fmt.Errorf("invalid company name: %w", err)
	}
	if err := state.ValidateShareCount(totalShares); err != nil {
		return gameState, fmt.Errorf("invalid total shares: %w", err)
	}
	if totalShares == 0 {
		return gameState, fmt.Errorf("total shares must be greater than zero")
	}

	// Check if company already exists
	if _, exists := gameState.Companies[id]; exists {
		return gameState, fmt.Errorf("company with ID %s already exists", id)
	}

	// Add company to state
	newState := gameState.AddCompany(id, name, totalShares)

	// Add log entry
	logEntry := state.LogEntry{
		Timestamp: time.Now(),
		Action:    "add_company",
		Details: map[string]interface{}{
			"company_id":    id,
			"company_name":  name,
			"total_shares":  totalShares,
		},
	}
	newState.Log = append(newState.Log, logEntry)

	return newState, nil
}

// RemovePlayer safely removes a player from the game
func RemovePlayer(gameState state.GameState, playerID string) (state.GameState, error) {
	// Validate inputs
	if err := state.ValidatePlayerID(playerID); err != nil {
		return gameState, fmt.Errorf("invalid player ID: %w", err)
	}

	player, exists := gameState.Players[playerID]
	if !exists {
		return gameState, fmt.Errorf("player not found: %s", playerID)
	}

	// Check if player owns any shares
	hasShares := false
	for _, shares := range player.Shares {
		if shares > 0 {
			hasShares = true
			break
		}
	}
	if hasShares {
		return gameState, fmt.Errorf("cannot remove player %s: still owns shares", playerID)
	}

	// Remove player
	newState := gameState.Clone()
	delete(newState.Players, playerID)

	// Add log entry
	logEntry := state.LogEntry{
		Timestamp: time.Now(),
		Action:    "remove_player",
		Details: map[string]interface{}{
			"player_id":   playerID,
			"player_name": player.Name,
			"final_cash":  player.Cash,
		},
	}
	newState.Log = append(newState.Log, logEntry)

	return newState, nil
}