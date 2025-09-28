package ui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestUpdateKeyHandling(t *testing.T) {
	tests := []struct {
		name           string
		keyMsg         tea.KeyMsg
		expectQuit     bool
		expectedInput  string
		expectUndo     bool
		expectRedo     bool
		expectExecute  bool
	}{
		{
			name:       "ctrl+c quits",
			keyMsg:     tea.KeyMsg{Type: tea.KeyCtrlC},
			expectQuit: true,
		},
		{
			name:          "quit command quits",
			keyMsg:        tea.KeyMsg{Type: tea.KeyEnter},
			expectQuit:    true,
			expectedInput: "",
		},
		{
			name:       "undo command",
			keyMsg:     tea.KeyMsg{Type: tea.KeyEnter},
			expectUndo: true,
		},
		{
			name:       "redo command",
			keyMsg:     tea.KeyMsg{Type: tea.KeyEnter},
			expectRedo: true,
		},
		{
			name:          "regular command",
			keyMsg:        tea.KeyMsg{Type: tea.KeyEnter},
			expectExecute: true,
		},
		{
			name:          "backspace removes character",
			keyMsg:        tea.KeyMsg{Type: tea.KeyBackspace},
			expectedInput: "hell",
		},
		{
			name:          "printable character added",
			keyMsg:        tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}},
			expectedInput: "helloa",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testModelPtr := NewModel()
			testModel := *testModelPtr

			// Set up input for specific tests
			switch tt.name {
			case "quit command quits":
				testModel.input = "quit"
			case "undo command":
				testModel.input = "undo"
			case "redo command":
				testModel.input = "redo"
			case "regular command":
				testModel.input = "help"
			case "backspace removes character":
				testModel.input = "hello"
			case "printable character added":
				testModel.input = "hello"
			}

			newModel, cmd := testModel.Update(tt.keyMsg)

			if tt.expectQuit {
				if cmd == nil {
					t.Error("Expected quit command but got nil")
				}
				// Note: We can't easily test tea.Quit here, but we can verify a command was returned
			}

			if tt.expectedInput != "" {
				if newModel.(Model).input != tt.expectedInput {
					t.Errorf("Expected input %q, got %q", tt.expectedInput, newModel.(Model).input)
				}
			}

			// For undo/redo/execute, we mainly test that the model changed appropriately
			if tt.expectUndo || tt.expectRedo || tt.expectExecute {
				if newModel.(Model).input != "" {
					t.Error("Input should be cleared after command execution")
				}
			}
		})
	}
}

func TestUpdateNonPrintableKeys(t *testing.T) {
	testModelPtr := NewModel()
	testModel := *testModelPtr
	testModel.input = "test"

	// Test non-printable characters (should be ignored)
	nonPrintableKeys := []tea.KeyMsg{
		{Type: tea.KeyTab},
		{Type: tea.KeyEsc},
		{Type: tea.KeyF1},
		{Type: tea.KeyUp},
		{Type: tea.KeyDown},
		{Type: tea.KeyLeft},
		{Type: tea.KeyRight},
	}

	for _, keyMsg := range nonPrintableKeys {
		newModel, _ := testModel.Update(keyMsg)
		if newModel.(Model).input != "test" {
			t.Errorf("Non-printable key %v should not change input, but input changed to %q", keyMsg.Type, newModel.(Model).input)
		}
	}
}

func TestUpdatePrintableCharacters(t *testing.T) {
	testModelPtr := NewModel()
	testModel := *testModelPtr

	printableChars := []string{"a", "B", "1", "!", "@", " ", "~"}

	for _, char := range printableChars {
		keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(char)}
		newModel, _ := testModel.Update(keyMsg)
		expectedInput := testModel.input + char

		if newModel.(Model).input != expectedInput {
			t.Errorf("Expected input %q after adding %q, got %q", expectedInput, char, newModel.(Model).input)
		}

		testModel = newModel.(Model)
	}
}

func TestUpdateBackspace(t *testing.T) {
	tests := []struct {
		name          string
		initialInput  string
		expectedInput string
	}{
		{
			name:          "backspace on empty input",
			initialInput:  "",
			expectedInput: "",
		},
		{
			name:          "backspace on single character",
			initialInput:  "a",
			expectedInput: "",
		},
		{
			name:          "backspace on multiple characters",
			initialInput:  "hello",
			expectedInput: "hell",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testModelPtr := NewModel()
			testModel := *testModelPtr
			testModel.input = tt.initialInput

			keyMsg := tea.KeyMsg{Type: tea.KeyBackspace}
			newModel, _ := testModel.Update(keyMsg)

			if newModel.(Model).input != tt.expectedInput {
				t.Errorf("Expected input %q after backspace, got %q", tt.expectedInput, newModel.(Model).input)
			}
		})
	}
}

func TestUpdateWindowSize(t *testing.T) {
	testModelPtr := NewModel()
	testModel := *testModelPtr
	windowMsg := tea.WindowSizeMsg{Width: 80, Height: 24}

	newModel, cmd := testModel.Update(windowMsg)

	if cmd != nil {
		t.Error("Window size message should not return a command")
	}

	// Model should be unchanged by window size message
	if newModel.(Model).input != testModel.input {
		t.Error("Window size message should not change model input")
	}
}

func TestUpdateUnknownMessage(t *testing.T) {
	testModelPtr := NewModel()
	testModel := *testModelPtr

	// Test with a custom message type
	type CustomMsg struct{}
	customMsg := CustomMsg{}

	newModel, cmd := testModel.Update(customMsg)

	if cmd != nil {
		t.Error("Unknown message should not return a command")
	}

	// Model should be unchanged
	if newModel.(Model).input != testModel.input {
		t.Error("Unknown message should not change model")
	}
}

func TestUpdateModelImmutability(t *testing.T) {
	originalModelPtr := NewModel()
	originalModel := *originalModelPtr
	originalModel.input = "test input"
	originalInput := originalModel.input

	// Update with a character
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}}
	newModel, _ := originalModel.Update(keyMsg)

	// Original model should be unchanged
	if originalModel.input != originalInput {
		t.Error("Original model was modified during update")
	}

	// New model should have the change
	if newModel.(Model).input != originalInput+"x" {
		t.Error("New model does not have expected changes")
	}
}

func TestUpdateEnterWithExitCommand(t *testing.T) {
	testModelPtr := NewModel()
	testModel := *testModelPtr
	testModel.input = "exit"

	keyMsg := tea.KeyMsg{Type: tea.KeyEnter}
	_, cmd := testModel.Update(keyMsg)

	if cmd == nil {
		t.Error("Exit command should return a quit command")
	}
}