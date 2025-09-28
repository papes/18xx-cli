package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/papes/18xxCli/config"
	"github.com/papes/18xxCli/state"
	"github.com/papes/18xxCli/ui/parser"
)

// Model represents the Bubble Tea application model
type Model struct {
	gameState       state.GameState
	history         *state.History
	input           string
	message         string
	err             error
	gameConfig      *config.GameConfig
	commandExecutor *parser.CommandExecutor
}

// NewModel creates a new UI model with initial game state
func NewModel() *Model {
	initialState := state.NewGameState()
	history := state.NewHistory(initialState, 100) // Keep last 100 states

	// Create command registry and executor
	registry := parser.CreateDefaultRegistry()
	executor := parser.NewCommandExecutor(registry)

	return &Model{
		gameState:       initialState,
		history:         history,
		input:           "",
		message:         "Welcome to 1830 Banking Tool! Type 'help' for commands.",
		err:             nil,
		gameConfig:      nil,
		commandExecutor: executor,
	}
}

// Init implements tea.Model
func (m Model) Init() tea.Cmd {
	return nil
}

// executeCommand processes user commands and returns updated model
func (m Model) executeCommand(command string) Model {
	newModel := m

	if command == "" {
		return newModel
	}

	// Execute command using the new parser
	result := m.commandExecutor.Execute(command, m.gameState, m.gameConfig)

	if result.Error != nil {
		newModel.err = result.Error
		newModel.message = "Error: " + result.Error.Error()
	} else {
		// Command succeeded - update state through history
		newModel.history.Push(result.NewState)
		newModel.gameState = result.NewState
		newModel.message = result.Message
		newModel.err = nil
		if result.NewGameConfig != nil {
			newModel.gameConfig = result.NewGameConfig
		}
	}

	// Clear input
	newModel.input = ""

	return newModel
}

// undo performs an undo operation
func (m Model) undo() Model {
	newModel := m

	if undoState, ok := newModel.history.Undo(); ok {
		newModel.gameState = undoState
		newModel.message = "Undid last action"
		newModel.err = nil
	} else {
		newModel.message = "Nothing to undo"
	}

	newModel.input = ""
	return newModel
}

// redo performs a redo operation
func (m Model) redo() Model {
	newModel := m

	if redoState, ok := newModel.history.Redo(); ok {
		newModel.gameState = redoState
		newModel.message = "Redid action"
		newModel.err = nil
	} else {
		newModel.message = "Nothing to redo"
	}

	newModel.input = ""
	return newModel
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

func boolToString(b bool) string {
	if b {
		return "true"
	}
	return "false"
}
