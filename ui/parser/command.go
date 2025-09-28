package parser

import (
	"fmt"
	"strings"

	"github.com/papes/18xxCli/config"
	"github.com/papes/18xxCli/state"
)

// ValidationError represents an error in command validation
type ValidationError struct {
	message string
}

// NewValidationError creates a new validation error
func NewValidationError(message string) *ValidationError {
	return &ValidationError{message: message}
}

// Error implements the error interface
func (e *ValidationError) Error() string {
	return e.message
}

// CommandResult represents the result of executing a command
type CommandResult struct {
	NewState      state.GameState
	Message       string
	Error         error
	NewGameConfig *config.GameConfig
}

// Command interface defines the contract for all commands
type Command interface {
	// Name returns the primary name of the command
	Name() string

	// Aliases returns alternative names for the command
	Aliases() []string

	// Description returns a description of what the command does
	Description() string

	// Usage returns usage information for the command
	Usage() string

	// Validate checks if the provided arguments are valid for this command
	Validate(args []interface{}) error

	// Execute runs the command with the given arguments and game state
	Execute(args []interface{}, gameState state.GameState, gameConfig *config.GameConfig) CommandResult
}

// CommandRegistry manages all available commands
type CommandRegistry struct {
	commands map[string]Command
	aliases  map[string]string // maps alias to primary command name
}

// NewCommandRegistry creates a new command registry
func NewCommandRegistry() *CommandRegistry {
	return &CommandRegistry{
		commands: make(map[string]Command),
		aliases:  make(map[string]string),
	}
}

// Register adds a command to the registry
func (r *CommandRegistry) Register(cmd Command) {
	name := strings.ToLower(cmd.Name())
	r.commands[name] = cmd

	// Register aliases
	for _, alias := range cmd.Aliases() {
		alias = strings.ToLower(alias)
		r.aliases[alias] = name
	}
}

// Get retrieves a command by name or alias
func (r *CommandRegistry) Get(name string) (Command, bool) {
	name = strings.ToLower(name)

	// Try direct command name first
	if cmd, exists := r.commands[name]; exists {
		return cmd, true
	}

	// Try alias
	if primaryName, exists := r.aliases[name]; exists {
		if cmd, exists := r.commands[primaryName]; exists {
			return cmd, true
		}
	}

	return nil, false
}

// GetSuggestions returns command names that are similar to the given input
func (r *CommandRegistry) GetSuggestions(input string) []string {
	input = strings.ToLower(input)
	var suggestions []string

	// Exact prefix matches first
	for name := range r.commands {
		if strings.HasPrefix(name, input) {
			suggestions = append(suggestions, name)
		}
	}

	// Alias prefix matches
	for alias, primaryName := range r.aliases {
		if strings.HasPrefix(alias, input) {
			// Only add if not already in suggestions
			found := false
			for _, existing := range suggestions {
				if existing == primaryName {
					found = true
					break
				}
			}
			if !found {
				suggestions = append(suggestions, alias)
			}
		}
	}

	return suggestions
}

// List returns all registered command names
func (r *CommandRegistry) List() []string {
	var names []string
	for name := range r.commands {
		names = append(names, name)
	}
	return names
}

// CommandExecutor handles parsing and executing commands
type CommandExecutor struct {
	registry *CommandRegistry
}

// NewCommandExecutor creates a new command executor with the given registry
func NewCommandExecutor(registry *CommandRegistry) *CommandExecutor {
	return &CommandExecutor{
		registry: registry,
	}
}

// Execute parses a command string and executes the appropriate command
func (e *CommandExecutor) Execute(commandStr string, gameState state.GameState, gameConfig *config.GameConfig) CommandResult {
	// Parse the command
	parsedCmd, err := ParseCommand(commandStr)
	if err != nil {
		return CommandResult{
			NewState:      gameState,
			Message:       "",
			Error:         fmt.Errorf("failed to parse command: %w", err),
			NewGameConfig: gameConfig,
		}
	}

	// Try to find command with all arguments as part of the command
	// This handles multi-word commands like "add player", "buy ipo", etc.
	fullCommand := parsedCmd.Command
	args := parsedCmd.Args

	// Try progressively longer command names
	for i := 0; i < len(args) && i < 3; i++ { // Limit to prevent overly long command names
		if argStr, ok := args[i].(string); ok {
			testCommand := fullCommand + " " + argStr
			if cmd, exists := e.registry.Get(testCommand); exists {
				// Found a match - use remaining args
				remainingArgs := args[i+1:]

				// Validate arguments
				if err := cmd.Validate(remainingArgs); err != nil {
					return CommandResult{
						NewState:      gameState,
						Message:       "",
						Error:         fmt.Errorf("invalid arguments for '%s': %w\nUsage: %s", testCommand, err, cmd.Usage()),
						NewGameConfig: gameConfig,
					}
				}

				// Execute the command
				return cmd.Execute(remainingArgs, gameState, gameConfig)
			}
			fullCommand = testCommand
		} else {
			break // Can't extend command name with non-string arg
		}
	}

	// Try the base command
	if cmd, exists := e.registry.Get(parsedCmd.Command); exists {
		// Validate arguments
		if err := cmd.Validate(parsedCmd.Args); err != nil {
			return CommandResult{
				NewState:      gameState,
				Message:       "",
				Error:         fmt.Errorf("invalid arguments for '%s': %w\nUsage: %s", parsedCmd.Command, err, cmd.Usage()),
				NewGameConfig: gameConfig,
			}
		}

		// Execute the command
		return cmd.Execute(parsedCmd.Args, gameState, gameConfig)
	}

	// Command not found
	suggestions := e.registry.GetSuggestions(parsedCmd.Command)
	if len(suggestions) > 0 {
		return CommandResult{
			NewState:      gameState,
			Message:       "",
			Error:         fmt.Errorf("unknown command '%s'. Did you mean: %s?", parsedCmd.Command, strings.Join(suggestions, ", ")),
			NewGameConfig: gameConfig,
		}
	}
	return CommandResult{
		NewState:      gameState,
		Message:       "",
		Error:         fmt.Errorf("unknown command '%s'. Type 'help' for available commands", parsedCmd.Command),
		NewGameConfig: gameConfig,
	}
}

// BaseCommand provides common functionality for commands
type BaseCommand struct {
	name        string
	aliases     []string
	description string
	usage       string
}

// NewBaseCommand creates a new base command
func NewBaseCommand(name, description, usage string, aliases ...string) BaseCommand {
	return BaseCommand{
		name:        name,
		aliases:     aliases,
		description: description,
		usage:       usage,
	}
}

// Name returns the command name
func (b BaseCommand) Name() string {
	return b.name
}

// Aliases returns the command aliases
func (b BaseCommand) Aliases() []string {
	return b.aliases
}

// Description returns the command description
func (b BaseCommand) Description() string {
	return b.description
}

// Usage returns the command usage
func (b BaseCommand) Usage() string {
	return b.usage
}
