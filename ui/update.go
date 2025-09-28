package ui

import (
	tea "github.com/charmbracelet/bubbletea"
)

// Update implements tea.Model.Update()
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit

		case "enter":
			// Execute the command
			if m.input == "quit" || m.input == "exit" {
				return m, tea.Quit
			}

			if m.input == "undo" {
				return m.undo(), nil
			}

			if m.input == "redo" {
				return m.redo(), nil
			}

			return m.executeCommand(m.input), nil

		case "backspace":
			// Remove last character from input
			if len(m.input) > 0 {
				m.input = m.input[:len(m.input)-1]
			}
			return m, nil

		default:
			// Add character to input (if it's a printable character)
			if len(msg.String()) == 1 && msg.String()[0] >= 32 && msg.String()[0] <= 126 {
				m.input += msg.String()
			}
			return m, nil
		}

	case tea.WindowSizeMsg:
		// Handle window resize if needed
		return m, nil

	default:
		return m, nil
	}
}
