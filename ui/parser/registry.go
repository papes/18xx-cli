package parser

import (
	"fmt"

	"github.com/papes/18xxCli/actions"
	"github.com/papes/18xxCli/config"
	"github.com/papes/18xxCli/state"
)

// CreateDefaultRegistry creates and initializes a command registry with all available commands
func CreateDefaultRegistry() *CommandRegistry {
	registry := NewCommandRegistry()

	// Register commands directly here to avoid import cycles
	registry.Register(&HelpCommand{
		BaseCommand: NewBaseCommand("help", "Show help information", "help [command]", "h", "?"),
		registry:    registry,
	})

	registry.Register(&StatusCommand{
		BaseCommand: NewBaseCommand("status", "Show current game state", "status", "st"),
	})

	registry.Register(&AddPlayerCommand{
		BaseCommand: NewBaseCommand("add player", "Add a new player", "add player <name> <cash>", "ap"),
	})

	registry.Register(&AddPlayersCommand{
		BaseCommand: NewBaseCommand("add players", "Add multiple players", "add players <name1> <name2> ... <cash>", "aps"),
	})

	registry.Register(&BuyShareCommand{
		BaseCommand: NewBaseCommand("buy", "Buy shares", "buy <ipo|bank> <player> <company> <shares>", "b"),
	})

	registry.Register(&SellShareCommand{
		BaseCommand: NewBaseCommand("sell", "Sell shares", "sell bank <player> <company> <shares>", "s"),
	})

	registry.Register(&SetCommand{
		BaseCommand: NewBaseCommand("set", "Set values", "set <par|price> <company> <value>", "sp"),
	})

	// Game setup commands
	registry.Register(&SetupGameCommand{
		BaseCommand: NewBaseCommand("setup game", "Setup complete game", "setup game <game> <player1> <player2> ..."),
	})

	registry.Register(&ListGamesCommand{
		BaseCommand: NewBaseCommand("list games", "List available games", "list games"),
	})

	registry.Register(&LoadCompaniesCommand{
		BaseCommand: NewBaseCommand("load companies", "Load companies from game", "load companies <game>"),
	})

	// Company commands
	registry.Register(&AddCompanyCommand{
		BaseCommand: NewBaseCommand("add company", "Add a new company", "add company <name> <shares>"),
	})

	registry.Register(&AddCompaniesCommand{
		BaseCommand: NewBaseCommand("add companies", "Add multiple companies", "add companies <name1> <name2> ... <shares>"),
	})

	// Financial commands
	registry.Register(&RevenueCommand{
		BaseCommand: NewBaseCommand("revenue", "Add revenue to company", "revenue <company> <amount>"),
	})

	registry.Register(&WithholdCommand{
		BaseCommand: NewBaseCommand("withhold", "Withhold revenue", "withhold <company> <amount>"),
	})

	registry.Register(&PayDividendCommand{
		BaseCommand: NewBaseCommand("pay dividend", "Pay dividend", "pay dividend <company> <total>"),
	})

	registry.Register(&TransferCommand{
		BaseCommand: NewBaseCommand("transfer", "Transfer money", "transfer <from> <to> <amount>"),
	})

	// Company operations
	registry.Register(&BuyTrainCommand{
		BaseCommand: NewBaseCommand("buy train", "Buy train for company", "buy train <company> <type> <cost>"),
	})

	registry.Register(&FloatCommand{
		BaseCommand: NewBaseCommand("float", "Float company", "float <company>"),
	})

	registry.Register(&SetPresidentCommand{
		BaseCommand: NewBaseCommand("set president", "Set company president", "set president <company> <player>"),
	})

	registry.Register(&LogsCommand{
		BaseCommand: NewBaseCommand("logs", "View all game logs", "logs", "log"),
	})

	return registry
}

// Command implementations to avoid import cycle

type HelpCommand struct {
	BaseCommand
	registry *CommandRegistry
}

func (c *HelpCommand) Validate(args []interface{}) error {
	if len(args) > 1 {
		return NewValidationError("too many arguments")
	}
	return nil
}

func (c *HelpCommand) Execute(args []interface{}, gameState state.GameState, gameConfig *config.GameConfig) CommandResult {
	message := `Available commands:
  SETUP:
    help                              - Show this help
    status                            - Show current game state
    add player <name> <cash>          - Add a new player
    add players <name1> <name2> ... <cash> - Add multiple players with same cash
    add company <name> <shares>       - Add a new company
    add companies <name1> <name2> ... <shares> - Add multiple companies with same shares
    setup game <game> <player1> <player2> ... - Complete game setup with players & companies
    load companies <game>             - Load companies only from game config
    list games                        - Show available game configurations

  STOCK OPERATIONS:
    set par <company> <price>         - Set company par value
    buy <ipo|bank> <player> <company> <shares> - Buy shares from IPO or bank pool
    sell bank <player> <company> <shares> - Sell shares to bank pool
    set price <company> <price>       - Set stock market price
    set president <company> <player>  - Set company president

  COMPANY OPERATIONS:
    revenue <company> <amount>        - Add revenue to company
    withhold <company> <amount>       - Withhold revenue (no dividends)
    pay dividend <company> <total>    - Pay dividend using current game rules
    buy train <company> <type> <cost> - Buy train for company
    float <company>                   - Float company (trigger IPO proceeds)

  FINANCIAL:
    transfer <from> <to> <amount>     - Transfer money between entities

  SYSTEM:
    logs                              - View all game logs
    undo                              - Undo last action
    redo                              - Redo last undone action
    quit                              - Exit the program

Examples:
  list games
  setup game 1830 Alice Bob Charlie Dave
  set par BO 76
  buy ipo Alice BO 2
  revenue BO 200
  pay dividend BO 100
  buy train BO 2 80`

	return CommandResult{
		NewState:      gameState,
		Message:       message,
		Error:         nil,
		NewGameConfig: gameConfig,
	}
}

type StatusCommand struct {
	BaseCommand
}

func (c *StatusCommand) Validate(args []interface{}) error {
	if len(args) > 0 {
		return NewValidationError("status command takes no arguments")
	}
	return nil
}

func (c *StatusCommand) Execute(args []interface{}, gameState state.GameState, gameConfig *config.GameConfig) CommandResult {
	message := formatGameStatus(gameState)
	return CommandResult{
		NewState:      gameState,
		Message:       message,
		Error:         nil,
		NewGameConfig: gameConfig,
	}
}

type AddPlayerCommand struct {
	BaseCommand
}

func (c *AddPlayerCommand) Validate(args []interface{}) error {
	if len(args) != 2 {
		return NewValidationError("requires player name and starting cash amount")
	}

	if _, ok := args[0].(string); !ok {
		return NewValidationError("player name must be a string")
	}

	cash, ok := args[1].(int)
	if !ok {
		return NewValidationError("cash amount must be a number")
	}
	if cash <= 0 {
		return NewValidationError("cash amount must be positive")
	}

	return nil
}

func (c *AddPlayerCommand) Execute(args []interface{}, gameState state.GameState, gameConfig *config.GameConfig) CommandResult {
	playerName := args[0].(string)
	cash := args[1].(int) // Validation already confirmed this works

	newState := gameState.AddPlayer(playerName, playerName, state.Money(cash))
	message := fmt.Sprintf("Added player %s with $%d", playerName, cash)

	return CommandResult{
		NewState:      newState,
		Message:       message,
		Error:         nil,
		NewGameConfig: gameConfig,
	}
}

type AddPlayersCommand struct {
	BaseCommand
}

func (c *AddPlayersCommand) Validate(args []interface{}) error {
	if len(args) < 2 {
		return NewValidationError("requires at least one player name and cash amount")
	}

	// Player names should be strings, last argument should be integer
	for i := 0; i < len(args)-1; i++ {
		if _, ok := args[i].(string); !ok {
			return NewValidationError(fmt.Sprintf("argument %d must be a player name (string)", i+1))
		}
	}

	// Last argument should be a valid number
	cash, ok := args[len(args)-1].(int)
	if !ok {
		return NewValidationError("last argument must be cash amount (number)")
	}
	if cash <= 0 {
		return NewValidationError("cash amount must be positive")
	}

	return nil
}

func (c *AddPlayersCommand) Execute(args []interface{}, gameState state.GameState, gameConfig *config.GameConfig) CommandResult {
	cash := args[len(args)-1].(int) // Validation already confirmed this works
	playerNames := make([]string, 0, len(args)-1)

	for i := 0; i < len(args)-1; i++ {
		playerNames = append(playerNames, args[i].(string))
	}

	newState := gameState
	for _, playerName := range playerNames {
		newState = newState.AddPlayer(playerName, playerName, state.Money(cash))
	}

	message := fmt.Sprintf("Added players: %s (each with $%d)", joinStrings(playerNames, ", "), cash)

	return CommandResult{
		NewState:      newState,
		Message:       message,
		Error:         nil,
		NewGameConfig: gameConfig,
	}
}

type BuyShareCommand struct {
	BaseCommand
}

func (c *BuyShareCommand) Validate(args []interface{}) error {
	if len(args) != 4 {
		return NewValidationError("requires source (ipo/bank), player name, company name, and number of shares")
	}

	source, ok := args[0].(string)
	if !ok {
		return NewValidationError("source must be 'ipo' or 'bank'")
	}
	if source != "ipo" && source != "bank" {
		return NewValidationError("source must be 'ipo' or 'bank'")
	}

	if _, ok := args[1].(string); !ok {
		return NewValidationError("player name must be a string")
	}

	if _, ok := args[2].(string); !ok {
		return NewValidationError("company name must be a string")
	}

	shares, ok := args[3].(int)
	if !ok {
		return NewValidationError("number of shares must be a number")
	}
	if shares <= 0 {
		return NewValidationError("number of shares must be positive")
	}

	return nil
}

func (c *BuyShareCommand) Execute(args []interface{}, gameState state.GameState, gameConfig *config.GameConfig) CommandResult {
	source := args[0].(string)
	playerID := args[1].(string)
	companyID := args[2].(string)
	shares := args[3].(int)

	var newState state.GameState
	var err error
	var message string

	if source == "ipo" {
		newState, err = actions.BuyShareFromIPO(gameState, playerID, companyID, state.ShareCount(shares))
		if err == nil {
			message = fmt.Sprintf("%s bought %d shares of %s from IPO", playerID, shares, companyID)
		}
	} else {
		newState, err = actions.BuyShareFromBankPool(gameState, playerID, companyID, state.ShareCount(shares))
		if err == nil {
			message = fmt.Sprintf("%s bought %d shares of %s from bank pool", playerID, shares, companyID)
		}
	}

	return CommandResult{
		NewState:      newState,
		Message:       message,
		Error:         err,
		NewGameConfig: gameConfig,
	}
}

type SellShareCommand struct {
	BaseCommand
}

func (c *SellShareCommand) Validate(args []interface{}) error {
	if len(args) != 4 {
		return NewValidationError("requires destination (bank), player name, company name, and number of shares")
	}

	dest, ok := args[0].(string)
	if !ok || dest != "bank" {
		return NewValidationError("destination must be 'bank'")
	}

	if _, ok := args[1].(string); !ok {
		return NewValidationError("player name must be a string")
	}

	if _, ok := args[2].(string); !ok {
		return NewValidationError("company name must be a string")
	}

	shares, ok := args[3].(int)
	if !ok {
		return NewValidationError("number of shares must be a number")
	}
	if shares <= 0 {
		return NewValidationError("number of shares must be positive")
	}

	return nil
}

func (c *SellShareCommand) Execute(args []interface{}, gameState state.GameState, gameConfig *config.GameConfig) CommandResult {
	playerID := args[1].(string)
	companyID := args[2].(string)
	shares := args[3].(int)

	newState, err := actions.SellShareToBankPool(gameState, playerID, companyID, state.ShareCount(shares))
	var message string
	if err == nil {
		message = fmt.Sprintf("%s sold %d shares of %s to bank pool", playerID, shares, companyID)
	}

	return CommandResult{
		NewState:      newState,
		Message:       message,
		Error:         err,
		NewGameConfig: gameConfig,
	}
}

type SetCommand struct {
	BaseCommand
}

func (c *SetCommand) Validate(args []interface{}) error {
	if len(args) != 3 {
		return NewValidationError("requires type (par/price), company name, and value")
	}

	setType, ok := args[0].(string)
	if !ok {
		return NewValidationError("type must be 'par' or 'price'")
	}
	if setType != "par" && setType != "price" {
		return NewValidationError("type must be 'par' or 'price'")
	}

	if _, ok := args[1].(string); !ok {
		return NewValidationError("company name must be a string")
	}

	value, ok := args[2].(int)
	if !ok {
		return NewValidationError("value must be a number")
	}
	if value <= 0 {
		return NewValidationError("value must be positive")
	}

	return nil
}

func (c *SetCommand) Execute(args []interface{}, gameState state.GameState, gameConfig *config.GameConfig) CommandResult {
	setType := args[0].(string)
	companyID := args[1].(string)
	value := args[2].(int)

	var newState state.GameState
	var err error
	var message string

	if setType == "par" {
		newState, err = actions.SetParValue(gameState, companyID, state.StockPrice(value))
		if err == nil {
			message = fmt.Sprintf("Set par value for %s to $%d", companyID, value)
		}
	} else {
		newState, err = actions.SetStockPrice(gameState, companyID, state.StockPrice(value))
		if err == nil {
			message = fmt.Sprintf("Set stock price for %s to $%d", companyID, value)
		}
	}

	return CommandResult{
		NewState:      newState,
		Message:       message,
		Error:         err,
		NewGameConfig: gameConfig,
	}
}

// Helper functions
func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	if len(strs) == 1 {
		return strs[0]
	}

	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
}

func formatGameStatus(gameState state.GameState) string {
	status := "=== Game Status ===\n"

	status += "Bank Money: $" + intToString(int(gameState.BankMoney)) + "\n\n"

	status += "Players:\n"
	for id, player := range gameState.Players {
		status += "  " + id + " (" + player.Name + "): $" + intToString(int(player.Cash))
		if len(player.Shares) > 0 {
			status += " - Shares: "
			first := true
			for company, count := range player.Shares {
				if !first {
					status += ", "
				}
				status += company + ":" + intToString(int(count))
				first = false
			}
		}
		status += "\n"
	}

	status += "\nCompanies:\n"
	for id, company := range gameState.Companies {
		status += "  " + id + " (" + company.Name + "):"
		status += " Treasury=$" + intToString(int(company.Treasury))
		if company.ParValue > 0 {
			status += " Par=$" + intToString(int(company.ParValue))
		}
		if company.StockPrice > 0 {
			status += " Price=$" + intToString(int(company.StockPrice))
		}
		status += " IPO:" + intToString(int(company.SharesInIPO))
		status += " Bank:" + intToString(int(company.SharesInBank))
		if company.PresidentID != "" {
			status += " President:" + company.PresidentID
		}
		status += "\n"
	}

	status += "\nActions taken: " + intToString(len(gameState.Log))
	status += "\nUndo available: " + intToString(len(gameState.Log))

	return status
}

func intToString(i int) string {
	if i == 0 {
		return "0"
	}

	if i < 0 {
		return "-" + intToString(-i)
	}

	result := ""
	for i > 0 {
		result = string('0'+byte(i%10)) + result
		i /= 10
	}
	return result
}

// Game Setup Commands

type SetupGameCommand struct {
	BaseCommand
}

func (c *SetupGameCommand) Validate(args []interface{}) error {
	if len(args) < 2 {
		return NewValidationError("requires game name and at least one player name")
	}

	// All arguments should be strings now
	for i, arg := range args {
		if _, ok := arg.(string); !ok {
			if i == 0 {
				return NewValidationError("game name must be a string")
			} else {
				return NewValidationError(fmt.Sprintf("player name %d must be a string", i))
			}
		}
	}

	return nil
}

func (c *SetupGameCommand) Execute(args []interface{}, gameState state.GameState, gameConfig *config.GameConfig) CommandResult {
	gameName := args[0].(string)
	playerNames := make([]string, 0, len(args)-1)

	for i := 1; i < len(args); i++ {
		playerNames = append(playerNames, args[i].(string))
	}

	loadedGameConfig, err := config.LoadGameConfig(gameName)
	if err != nil {
		return CommandResult{
			NewState:      gameState,
			Message:       "",
			Error:         err,
			NewGameConfig: gameConfig,
		}
	}

	newState, err := loadedGameConfig.ApplyGameSetup(gameState, playerNames)
	if err != nil {
		return CommandResult{
			NewState:      gameState,
			Message:       "",
			Error:         err,
			NewGameConfig: gameConfig,
		}
	}

	playerConfig := loadedGameConfig.GetPlayerConfig(len(playerNames))
	message := "Set up " + loadedGameConfig.Title + " for " + intToString(len(playerNames)) + " players: " + joinStrings(playerNames, ", ")
	if playerConfig != nil {
		message += " (each with $" + intToString(playerConfig.StartingMoney) + ", bank: $" + intToString(playerConfig.BankMoney) + ")"
	}

	return CommandResult{
		NewState:      newState,
		Message:       message,
		Error:         nil,
		NewGameConfig: loadedGameConfig,
	}
}

type ListGamesCommand struct {
	BaseCommand
}

func (c *ListGamesCommand) Validate(args []interface{}) error {
	if len(args) > 0 {
		return NewValidationError("list games command takes no arguments")
	}
	return nil
}

func (c *ListGamesCommand) Execute(args []interface{}, gameState state.GameState, gameConfig *config.GameConfig) CommandResult {
	games, err := config.ListAvailableGames()
	if err != nil {
		return CommandResult{
			NewState:      gameState,
			Message:       "",
			Error:         err,
			NewGameConfig: gameConfig,
		}
	}

	if len(games) == 0 {
		return CommandResult{
			NewState:      gameState,
			Message:       "No game configurations found in config/games/",
			Error:         nil,
			NewGameConfig: gameConfig,
		}
	}

	message := "Available games: " + joinStrings(games, ", ")
	return CommandResult{
		NewState:      gameState,
		Message:       message,
		Error:         nil,
		NewGameConfig: gameConfig,
	}
}

type LoadCompaniesCommand struct {
	BaseCommand
}

func (c *LoadCompaniesCommand) Validate(args []interface{}) error {
	if len(args) != 1 {
		return NewValidationError("requires game name")
	}

	if _, ok := args[0].(string); !ok {
		return NewValidationError("game name must be a string")
	}

	return nil
}

func (c *LoadCompaniesCommand) Execute(args []interface{}, gameState state.GameState, gameConfig *config.GameConfig) CommandResult {
	gameName := args[0].(string)

	loadedGameConfig, err := config.LoadGameConfig(gameName)
	if err != nil {
		return CommandResult{
			NewState:      gameState,
			Message:       "",
			Error:         err,
			NewGameConfig: gameConfig,
		}
	}

	newState := loadedGameConfig.ApplyToGameState(gameState)
	companyNames := []string{}
	for _, company := range loadedGameConfig.Companies {
		companyNames = append(companyNames, company.ID)
	}

	message := "Loaded " + loadedGameConfig.Title + " companies: " + joinStrings(companyNames, ", ")
	return CommandResult{
		NewState:      newState,
		Message:       message,
		Error:         nil,
		NewGameConfig: gameConfig,
	}
}

// Company Commands

type AddCompanyCommand struct {
	BaseCommand
}

func (c *AddCompanyCommand) Validate(args []interface{}) error {
	if len(args) != 2 {
		return NewValidationError("requires company name and number of shares")
	}

	if _, ok := args[0].(string); !ok {
		return NewValidationError("company name must be a string")
	}

	shares, ok := args[1].(int)
	if !ok {
		return NewValidationError("number of shares must be a number")
	}
	if shares <= 0 {
		return NewValidationError("number of shares must be positive")
	}

	return nil
}

func (c *AddCompanyCommand) Execute(args []interface{}, gameState state.GameState, gameConfig *config.GameConfig) CommandResult {
	companyName := args[0].(string)
	shares := args[1].(int)

	newState := gameState.AddCompany(companyName, companyName, state.ShareCount(shares))
	message := fmt.Sprintf("Added company %s with %d shares", companyName, shares)

	return CommandResult{
		NewState:      newState,
		Message:       message,
		Error:         nil,
		NewGameConfig: gameConfig,
	}
}

type AddCompaniesCommand struct {
	BaseCommand
}

func (c *AddCompaniesCommand) Validate(args []interface{}) error {
	if len(args) < 2 {
		return NewValidationError("requires at least one company name and number of shares")
	}

	shares, ok := args[len(args)-1].(int)
	if !ok {
		return NewValidationError("last argument must be number of shares (number)")
	}
	if shares <= 0 {
		return NewValidationError("number of shares must be positive")
	}

	for i := 0; i < len(args)-1; i++ {
		if _, ok := args[i].(string); !ok {
			return NewValidationError(fmt.Sprintf("argument %d must be a company name (string)", i+1))
		}
	}

	return nil
}

func (c *AddCompaniesCommand) Execute(args []interface{}, gameState state.GameState, gameConfig *config.GameConfig) CommandResult {
	shares := args[len(args)-1].(int)
	companyNames := make([]string, 0, len(args)-1)

	for i := 0; i < len(args)-1; i++ {
		companyNames = append(companyNames, args[i].(string))
	}

	newState := gameState
	for _, companyName := range companyNames {
		newState = newState.AddCompany(companyName, companyName, state.ShareCount(shares))
	}

	message := fmt.Sprintf("Added companies: %s (each with %d shares)", joinStrings(companyNames, ", "), shares)

	return CommandResult{
		NewState:      newState,
		Message:       message,
		Error:         nil,
		NewGameConfig: gameConfig,
	}
}

// Financial Commands

type RevenueCommand struct {
	BaseCommand
}

func (c *RevenueCommand) Validate(args []interface{}) error {
	if len(args) != 2 {
		return NewValidationError("requires company name and amount")
	}

	if _, ok := args[0].(string); !ok {
		return NewValidationError("company name must be a string")
	}

	amount, ok := args[1].(int)
	if !ok {
		return NewValidationError("amount must be a number")
	}
	if amount <= 0 {
		return NewValidationError("amount must be positive")
	}

	return nil
}

func (c *RevenueCommand) Execute(args []interface{}, gameState state.GameState, gameConfig *config.GameConfig) CommandResult {
	companyID := args[0].(string)
	amount := args[1].(int)

	newState, err := actions.CompanyRevenue(gameState, companyID, state.Money(amount))
	var message string
	if err == nil {
		message = fmt.Sprintf("%s receives $%d revenue", companyID, amount)
	}

	return CommandResult{
		NewState:      newState,
		Message:       message,
		Error:         err,
		NewGameConfig: gameConfig,
	}
}

type WithholdCommand struct {
	BaseCommand
}

func (c *WithholdCommand) Validate(args []interface{}) error {
	if len(args) != 2 {
		return NewValidationError("requires company name and amount")
	}

	if _, ok := args[0].(string); !ok {
		return NewValidationError("company name must be a string")
	}

	amount, ok := args[1].(int)
	if !ok {
		return NewValidationError("amount must be a number")
	}
	if amount <= 0 {
		return NewValidationError("amount must be positive")
	}

	return nil
}

func (c *WithholdCommand) Execute(args []interface{}, gameState state.GameState, gameConfig *config.GameConfig) CommandResult {
	companyID := args[0].(string)
	amount := args[1].(int)

	newState, err := actions.WithholdRevenue(gameState, companyID, state.Money(amount))
	var message string
	if err == nil {
		message = fmt.Sprintf("%s withholds $%d", companyID, amount)
	}

	return CommandResult{
		NewState:      newState,
		Message:       message,
		Error:         err,
		NewGameConfig: gameConfig,
	}
}

type PayDividendCommand struct {
	BaseCommand
}

func (c *PayDividendCommand) Validate(args []interface{}) error {
	if len(args) != 2 {
		return NewValidationError("requires company name and total amount")
	}

	if _, ok := args[0].(string); !ok {
		return NewValidationError("company name must be a string")
	}

	total, ok := args[1].(int)
	if !ok {
		return NewValidationError("total amount must be a number")
	}
	if total <= 0 {
		return NewValidationError("total amount must be positive")
	}

	return nil
}

func (c *PayDividendCommand) Execute(args []interface{}, gameState state.GameState, gameConfig *config.GameConfig) CommandResult {
	companyID := args[0].(string)
	total := args[1].(int)

	if gameConfig == nil {
		return CommandResult{
			NewState:      gameState,
			Message:       "",
			Error:         fmt.Errorf("no game configuration loaded - use 'setup game' first"),
			NewGameConfig: gameConfig,
		}
	}

	newState, err := actions.PayDividendWithConfig(gameState, companyID, state.Money(total), gameConfig.BankSharesPayCompany, gameConfig.IPOSharesPayCompany)
	var message string
	if err == nil {
		settings := ""
		if gameConfig.BankSharesPayCompany || gameConfig.IPOSharesPayCompany {
			settings = " (bank:" + boolToString(gameConfig.BankSharesPayCompany) + " ipo:" + boolToString(gameConfig.IPOSharesPayCompany) + ")"
		}
		message = fmt.Sprintf("%s pays $%d total dividend%s", companyID, total, settings)
	}

	return CommandResult{
		NewState:      newState,
		Message:       message,
		Error:         err,
		NewGameConfig: gameConfig,
	}
}

type TransferCommand struct {
	BaseCommand
}

func (c *TransferCommand) Validate(args []interface{}) error {
	if len(args) != 3 {
		return NewValidationError("requires from entity, to entity, and amount")
	}

	if _, ok := args[0].(string); !ok {
		return NewValidationError("from entity must be a string")
	}

	if _, ok := args[1].(string); !ok {
		return NewValidationError("to entity must be a string")
	}

	amount, ok := args[2].(int)
	if !ok {
		return NewValidationError("amount must be a number")
	}
	if amount <= 0 {
		return NewValidationError("amount must be positive")
	}

	return nil
}

func (c *TransferCommand) Execute(args []interface{}, gameState state.GameState, gameConfig *config.GameConfig) CommandResult {
	from := args[0].(string)
	to := args[1].(string)
	amount := args[2].(int)

	newState, err := actions.TransferMoney(gameState, from, to, state.Money(amount))
	var message string
	if err == nil {
		message = fmt.Sprintf("Transferred $%d from %s to %s", amount, from, to)
	}

	return CommandResult{
		NewState:      newState,
		Message:       message,
		Error:         err,
		NewGameConfig: gameConfig,
	}
}

// Company Operation Commands

type BuyTrainCommand struct {
	BaseCommand
}

func (c *BuyTrainCommand) Validate(args []interface{}) error {
	if len(args) != 3 {
		return NewValidationError("requires company name, train type, and cost")
	}

	if _, ok := args[0].(string); !ok {
		return NewValidationError("company name must be a string")
	}

	if _, ok := args[1].(string); !ok {
		return NewValidationError("train type must be a string")
	}

	cost, ok := args[2].(int)
	if !ok {
		return NewValidationError("cost must be a number")
	}
	if cost <= 0 {
		return NewValidationError("cost must be positive")
	}

	return nil
}

func (c *BuyTrainCommand) Execute(args []interface{}, gameState state.GameState, gameConfig *config.GameConfig) CommandResult {
	companyID := args[0].(string)
	trainType := args[1].(string)
	cost := args[2].(int)

	newState, err := actions.BuyTrain(gameState, companyID, trainType, state.Money(cost))
	var message string
	if err == nil {
		message = fmt.Sprintf("%s buys a %s train for $%d", companyID, trainType, cost)
	}

	return CommandResult{
		NewState:      newState,
		Message:       message,
		Error:         err,
		NewGameConfig: gameConfig,
	}
}

type FloatCommand struct {
	BaseCommand
}

func (c *FloatCommand) Validate(args []interface{}) error {
	if len(args) != 1 {
		return NewValidationError("requires company name")
	}

	if _, ok := args[0].(string); !ok {
		return NewValidationError("company name must be a string")
	}

	return nil
}

func (c *FloatCommand) Execute(args []interface{}, gameState state.GameState, gameConfig *config.GameConfig) CommandResult {
	companyID := args[0].(string)

	newState, err := actions.FloatCompany(gameState, companyID)
	var message string
	if err == nil {
		message = fmt.Sprintf("%s floats", companyID)
	}

	return CommandResult{
		NewState:      newState,
		Message:       message,
		Error:         err,
		NewGameConfig: gameConfig,
	}
}

type SetPresidentCommand struct {
	BaseCommand
}

func (c *SetPresidentCommand) Validate(args []interface{}) error {
	if len(args) != 2 {
		return NewValidationError("requires company name and player name")
	}

	if _, ok := args[0].(string); !ok {
		return NewValidationError("company name must be a string")
	}

	if _, ok := args[1].(string); !ok {
		return NewValidationError("player name must be a string")
	}

	return nil
}

func (c *SetPresidentCommand) Execute(args []interface{}, gameState state.GameState, gameConfig *config.GameConfig) CommandResult {
	companyID := args[0].(string)
	playerID := args[1].(string)

	newState, err := actions.SetPresident(gameState, companyID, playerID)
	var message string
	if err == nil {
		message = fmt.Sprintf("%s becomes president of %s", playerID, companyID)
	}

	return CommandResult{
		NewState:      newState,
		Message:       message,
		Error:         err,
		NewGameConfig: gameConfig,
	}
}

type LogsCommand struct {
	BaseCommand
}

func (c *LogsCommand) Validate(args []interface{}) error {
	if len(args) > 0 {
		return NewValidationError("logs command takes no arguments")
	}
	return nil
}

func (c *LogsCommand) Execute(args []interface{}, gameState state.GameState, gameConfig *config.GameConfig) CommandResult {
	if len(gameState.Log) == 0 {
		return CommandResult{
			NewState:      gameState,
			Message:       "No actions logged yet",
			Error:         nil,
			NewGameConfig: gameConfig,
		}
	}

	message := "=== Game Log ===\n"
	for i, entry := range gameState.Log {
		message += intToString(i+1) + ". " + entry.Timestamp.Format("15:04:05") + " - " + entry.Action

		// Add relevant details based on action type
		if entry.Details != nil {
			switch entry.Action {
			case "dividend_payment":
				if company, ok := entry.Details["company"].(string); ok {
					if recipient, ok := entry.Details["recipient"].(string); ok {
						if amount, ok := entry.Details["amount_received"]; ok {
							if shares, ok := entry.Details["shares_owned"]; ok {
								recipientType, _ := entry.Details["recipient_type"].(string)
								if recipientType == "company" {
									message += ": " + company + " receives $" + formatMoney(amount) + " (self-payout)"
								} else {
									message += ": " + recipient + " receives $" + formatMoney(amount) + " (" + formatMoney(shares) + " shares of " + company + ")"
								}
							}
						}
					}
				}
			case "pay_dividend_with_config", "pay_detailed_dividend":
				if company, ok := entry.Details["company"].(string); ok {
					if total, ok := entry.Details["total_amount"]; ok {
						message += ": " + company + " pays $" + formatMoney(total) + " total dividend"
					}
				}
				if individualPayments, ok := entry.Details["individual_payments"]; ok {
					message += " (" + formatMoney(individualPayments) + " individual payments)"
				}
			case "buy_share_ipo", "buy_share_bank":
				if player, ok := entry.Details["player"].(string); ok {
					if company, ok := entry.Details["company"].(string); ok {
						if shares, ok := entry.Details["shares"]; ok {
							message += ": " + player + " buys " + formatMoney(shares) + " shares of " + company
						}
					}
				}
			case "sell_share_bank":
				if player, ok := entry.Details["player"].(string); ok {
					if company, ok := entry.Details["company"].(string); ok {
						if shares, ok := entry.Details["shares"]; ok {
							message += ": " + player + " sells " + formatMoney(shares) + " shares of " + company
						}
					}
				}
			case "company_revenue", "withhold_revenue":
				if company, ok := entry.Details["company"].(string); ok {
					if amount, ok := entry.Details["amount"]; ok {
						message += ": " + company + " $" + formatMoney(amount)
					}
				}
			case "transfer_money":
				if from, ok := entry.Details["from"].(string); ok {
					if to, ok := entry.Details["to"].(string); ok {
						if amount, ok := entry.Details["amount"]; ok {
							message += ": " + from + " → " + to + " $" + formatMoney(amount)
						}
					}
				}
			default:
				// For other actions, show key details
				if company, ok := entry.Details["company"].(string); ok {
					message += ": " + company
				}
			}
		}
		message += "\n"
	}

	return CommandResult{
		NewState:      gameState,
		Message:       message,
		Error:         nil,
		NewGameConfig: gameConfig,
	}
}

func formatMoney(amount interface{}) string {
	switch v := amount.(type) {
	case int:
		return intToString(v)
	case state.Money:
		return intToString(int(v))
	case state.ShareCount:
		return intToString(int(v))
	default:
		return "0"
	}
}

func boolToString(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

// Helper function to parse string arguments as integers
func parseStringAsInt(arg interface{}) (int, error) {
	str, ok := arg.(string)
	if !ok {
		return 0, fmt.Errorf("argument must be a string")
	}

	num := 0
	for _, char := range str {
		if char >= '0' && char <= '9' {
			num = num*10 + int(char-'0')
		} else {
			return 0, fmt.Errorf("invalid number: %s", str)
		}
	}
	return num, nil
}
