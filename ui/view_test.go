package ui

import (
	"strings"
	"testing"

	"github.com/papes/18xxCli/state"
)

func TestView(t *testing.T) {
	model := NewModel()
	view := model.View()

	// Basic checks for view structure
	if view == "" {
		t.Error("View should not be empty")
	}

	// Check for essential components in the view
	expectedComponents := []string{
		"1830 Banking Tool",
		"Welcome to 1830 Banking Tool!",
		"Command:",
	}

	for _, component := range expectedComponents {
		if !strings.Contains(view, component) {
			t.Errorf("View should contain %q", component)
		}
	}
}

func TestViewWithPlayers(t *testing.T) {
	model := NewModel()

	// Add some players to the game state
	newState := model.gameState.AddPlayer("alice", "Alice", state.Money(600))
	newState = newState.AddPlayer("bob", "Bob", state.Money(600))
	model.gameState = newState

	view := model.View()

	// Check that players are displayed
	if !strings.Contains(view, "Alice") {
		t.Error("View should contain player Alice")
	}

	if !strings.Contains(view, "Bob") {
		t.Error("View should contain player Bob")
	}

	if !strings.Contains(view, "600") {
		t.Error("View should contain player cash amounts")
	}
}

func TestViewWithCompanies(t *testing.T) {
	model := NewModel()

	// Add some companies to the game state
	newState := model.gameState.AddCompany("BO", "Baltimore & Ohio", state.ShareCount(10))
	newState = newState.AddCompany("PRR", "Pennsylvania Railroad", state.ShareCount(10))
	model.gameState = newState

	view := model.View()

	// Check that companies are displayed
	if !strings.Contains(view, "Baltimore & Ohio") {
		t.Error("View should contain company Baltimore & Ohio")
	}

	if !strings.Contains(view, "Pennsylvania Railroad") {
		t.Error("View should contain company Pennsylvania Railroad")
	}
}

func TestViewWithInput(t *testing.T) {
	model := NewModel()
	model.input = "test command"

	view := model.View()

	// Check that current input is displayed
	if !strings.Contains(view, "test command") {
		t.Error("View should contain current input")
	}
}

func TestViewWithMessage(t *testing.T) {
	model := NewModel()
	model.message = "Test message"

	view := model.View()

	// Check that message is displayed
	if !strings.Contains(view, "Test message") {
		t.Error("View should contain message")
	}
}

func TestViewWithError(t *testing.T) {
	model := NewModel()
	model.message = "Error: Something went wrong"

	view := model.View()

	// Check that error message is displayed
	if !strings.Contains(view, "Error: Something went wrong") {
		t.Error("View should contain error message")
	}
}

func TestViewGameStateDisplay(t *testing.T) {
	model := NewModel()

	// Set up a more complex game state
	newState := model.gameState.AddPlayer("alice", "Alice", state.Money(500))
	newState = newState.AddPlayer("bob", "Bob", state.Money(700))
	newState = newState.AddCompany("BO", "Baltimore & Ohio", state.ShareCount(10))

	// Set bank money to a specific value
	newState.BankMoney = state.Money(15000)
	model.gameState = newState

	view := model.View()

	// Check that bank money is displayed
	if !strings.Contains(view, "15000") {
		t.Error("View should contain bank money")
	}

	// Check that both players are shown with their cash
	if !strings.Contains(view, "Alice") || !strings.Contains(view, "500") {
		t.Error("View should show Alice with 500 cash")
	}

	if !strings.Contains(view, "Bob") || !strings.Contains(view, "700") {
		t.Error("View should show Bob with 700 cash")
	}
}

func TestViewNoPlayers(t *testing.T) {
	model := NewModel()
	view := model.View()

	// With no players, should still render without crashing
	if view == "" {
		t.Error("View should not be empty even with no players")
	}

	// Should contain bank information
	if !strings.Contains(view, "Bank") {
		t.Error("View should contain bank information")
	}
}

func TestViewLongInput(t *testing.T) {
	model := NewModel()
	model.input = "this is a very long command that might cause display issues if not handled properly"

	view := model.View()

	// Should still render without issues
	if view == "" {
		t.Error("View should not be empty with long input")
	}

	// Should contain the long input
	if !strings.Contains(view, "very long command") {
		t.Error("View should contain long input text")
	}
}

func TestViewConsistency(t *testing.T) {
	model := NewModel()

	// Generate view multiple times - should be consistent
	view1 := model.View()
	view2 := model.View()

	if view1 != view2 {
		t.Error("View should be consistent when called multiple times with same state")
	}
}

func TestViewStateChangeReflection(t *testing.T) {
	model := NewModel()
	view1 := model.View()

	// Change state
	newState := model.gameState.AddPlayer("alice", "Alice", state.Money(600))
	model.gameState = newState
	view2 := model.View()

	// Views should be different
	if view1 == view2 {
		t.Error("View should reflect state changes")
	}

	// New view should contain Alice
	if !strings.Contains(view2, "Alice") {
		t.Error("Updated view should contain new player Alice")
	}
}