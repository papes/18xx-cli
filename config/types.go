package config

import (
	"fmt"
	"github.com/papes/18xxCli/state"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
)

// Company represents a company configuration
type Company struct {
	ID     string `yaml:"id"`
	Name   string `yaml:"name"`
	Shares int    `yaml:"shares"`
}

// PlayerConfig represents player count specific configuration
type PlayerConfig struct {
	PlayerCount   int `yaml:"player_count"`
	StartingMoney int `yaml:"starting_money"`
	BankMoney     int `yaml:"bank_money"`
}

// GameConfig represents the configuration for a specific 18xx game
type GameConfig struct {
	Title                string         `yaml:"title"`
	Description          string         `yaml:"description"`
	Companies            []Company      `yaml:"companies"`
	PlayerConfigs        []PlayerConfig `yaml:"player_configs"`
	BankSharesPayCompany bool           `yaml:"bank_shares_pay_company"`
	IPOSharesPayCompany  bool           `yaml:"ipo_shares_pay_company"`
}

// LoadGameConfig loads a game configuration from a YAML file
func LoadGameConfig(gameName string) (*GameConfig, error) {
	// Look for config file in config/games/ directory
	configPath := filepath.Join("config", "games", gameName+".yaml")

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var config GameConfig
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

// ApplyToGameState creates companies from the config and adds them to the game state
func (gc *GameConfig) ApplyToGameState(gameState state.GameState) state.GameState {
	newState := gameState

	for _, company := range gc.Companies {
		// Use company ID as both ID and name for consistency with our simplified naming
		newState = newState.AddCompany(company.ID, company.ID, state.ShareCount(company.Shares))
	}

	return newState
}

// ApplyGameSetup creates a complete game setup with players, companies, and proper bank amount
func (gc *GameConfig) ApplyGameSetup(gameState state.GameState, playerNames []string) (state.GameState, error) {
	newState := gameState
	playerCount := len(playerNames)

	// Find the player configuration for this player count
	var playerConfig *PlayerConfig
	for _, config := range gc.PlayerConfigs {
		if config.PlayerCount == playerCount {
			playerConfig = &config
			break
		}
	}

	if playerConfig == nil {
		return gameState, fmt.Errorf("no configuration found for %d players in %s", playerCount, gc.Title)
	}

	// Set bank money
	newState.BankMoney = state.Money(playerConfig.BankMoney)

	// Add players with correct starting money
	for _, playerName := range playerNames {
		newState = newState.AddPlayer(playerName, playerName, state.Money(playerConfig.StartingMoney))
	}

	// Add companies
	for _, company := range gc.Companies {
		newState = newState.AddCompany(company.ID, company.ID, state.ShareCount(company.Shares))
	}

	return newState, nil
}

// GetPlayerConfig returns the player configuration for a specific player count
func (gc *GameConfig) GetPlayerConfig(playerCount int) *PlayerConfig {
	for _, config := range gc.PlayerConfigs {
		if config.PlayerCount == playerCount {
			return &config
		}
	}
	return nil
}

// ListAvailableGames returns a list of available game configuration files
func ListAvailableGames() ([]string, error) {
	configDir := filepath.Join("config", "games")

	files, err := os.ReadDir(configDir)
	if err != nil {
		return nil, err
	}

	var games []string
	for _, file := range files {
		if filepath.Ext(file.Name()) == ".yaml" {
			gameName := file.Name()[:len(file.Name())-5] // Remove .yaml extension
			games = append(games, gameName)
		}
	}

	return games, nil
}
