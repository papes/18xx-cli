package state

import "time"

// Money represents currency amounts in the game
type Money int

// ShareCount represents number of shares
type ShareCount int

// StockPrice represents a stock price (can be 0 if not set)
type StockPrice int

// Player represents a player in the game
type Player struct {
	ID   string
	Name string
	Cash Money
	// Shares maps company ID to number of shares owned
	Shares map[string]ShareCount
}

// Company represents a railroad company
type Company struct {
	ID       string
	Name     string
	Treasury Money
	// StockPrice is current market price (0 if not set/IPO)
	StockPrice StockPrice
	// ParValue is the IPO price (0 if not set)
	ParValue StockPrice
	// SharesInIPO tracks shares available for purchase from IPO
	SharesInIPO ShareCount
	// SharesInBank tracks shares sold back to bank pool
	SharesInBank ShareCount
	// TotalShares is the total number of shares for this company (usually 10)
	TotalShares ShareCount
	// PresidentID is the current president (player with most shares)
	PresidentID string
}

// LogEntry represents a logged action in the game
type LogEntry struct {
	Timestamp time.Time
	Action    string
	Details   map[string]interface{}
}

// GameState represents the complete state of the game
type GameState struct {
	Players   map[string]*Player
	Companies map[string]*Company
	// Bank represents the bank's money (usually starts very high)
	BankMoney Money
	// Log contains all actions taken in the game
	Log []LogEntry
}

// Clone creates a deep copy of the GameState for immutable operations
func (gs GameState) Clone() GameState {
	newState := GameState{
		Players:   make(map[string]*Player),
		Companies: make(map[string]*Company),
		BankMoney: gs.BankMoney,
		Log:       make([]LogEntry, len(gs.Log)),
	}

	// Deep copy players
	for id, player := range gs.Players {
		newPlayer := &Player{
			ID:     player.ID,
			Name:   player.Name,
			Cash:   player.Cash,
			Shares: make(map[string]ShareCount),
		}
		for companyID, count := range player.Shares {
			newPlayer.Shares[companyID] = count
		}
		newState.Players[id] = newPlayer
	}

	// Deep copy companies
	for id, company := range gs.Companies {
		newCompany := &Company{
			ID:           company.ID,
			Name:         company.Name,
			Treasury:     company.Treasury,
			StockPrice:   company.StockPrice,
			ParValue:     company.ParValue,
			SharesInIPO:  company.SharesInIPO,
			SharesInBank: company.SharesInBank,
			TotalShares:  company.TotalShares,
			PresidentID:  company.PresidentID,
		}
		newState.Companies[id] = newCompany
	}

	// Copy log entries
	copy(newState.Log, gs.Log)

	return newState
}

// NewGameState creates a new empty game state with default bank money of 12000.
// This represents the starting state for a new 1830 game.
func NewGameState() GameState {
	return GameState{
		Players:   make(map[string]*Player),
		Companies: make(map[string]*Company),
		BankMoney: 12000, // Standard 1830 bank
		Log:       make([]LogEntry, 0),
	}
}

// AddPlayer adds a new player to the game with the specified starting cash.
// Note: Validation is handled by calling functions to maintain immutability.
// Returns a new GameState with the player added.
func (gs GameState) AddPlayer(id, name string, startingCash Money) GameState {
	// Note: Validation is handled by calling functions to maintain immutability
	newState := gs.Clone()
	newState.Players[id] = &Player{
		ID:     id,
		Name:   name,
		Cash:   startingCash,
		Shares: make(map[string]ShareCount),
	}
	return newState
}

// AddCompany adds a new company to the game with the specified total shares.
// The company starts with no treasury, no set par value or stock price,
// and all shares in the IPO. Note: Validation is handled by calling functions
// to maintain immutability. Returns a new GameState with the company added.
func (gs GameState) AddCompany(id, name string, totalShares ShareCount) GameState {
	// Note: Validation is handled by calling functions to maintain immutability
	newState := gs.Clone()
	newState.Companies[id] = &Company{
		ID:           id,
		Name:         name,
		Treasury:     0,
		StockPrice:   0, // Not set until floated
		ParValue:     0, // Not set until IPO
		SharesInIPO:  totalShares,
		SharesInBank: 0,
		TotalShares:  totalShares,
		PresidentID:  "", // No president initially
	}
	return newState
}

// CalculatePlayerValue calculates the total value of a player including
// cash on hand plus the market value of all shares owned.
// Returns 0 if the player doesn't exist. Share values are calculated
// using current stock prices; shares with no stock price set are valued at 0.
func (gs GameState) CalculatePlayerValue(playerID string) Money {
	player, exists := gs.Players[playerID]
	if !exists {
		return 0
	}

	totalValue := player.Cash

	// Add market value of all shares
	for companyID, shareCount := range player.Shares {
		if shareCount > 0 {
			company, companyExists := gs.Companies[companyID]
			if companyExists && company.StockPrice > 0 {
				shareValue := Money(company.StockPrice) * Money(shareCount)
				totalValue += shareValue
			}
		}
	}

	return totalValue
}
