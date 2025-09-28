package ui

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/papes/18xxCli/state"
	"strings"
)

// Style definitions
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 1)

	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#04B575"))

	sectionStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#0969DA"))

	moneyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#CF222E"))

	companyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#8250DF"))

	playerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#1F883D"))

	inputStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFA500"))

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#656D76")).
			Italic(true)

	messageStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#0969DA")).
			Bold(true)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#DA3633")).
			Bold(true)

	separatorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#656D76"))
)

// View implements tea.Model.View()
func (m Model) View() string {
	var sections []string

	// Header
	sections = append(sections, titleStyle.Render("1830 Banking Tool"))
	sections = append(sections, "")

	// Show current message (success/error feedback)
	if m.message != "" {
		var msgStyle lipgloss.Style
		if strings.Contains(strings.ToLower(m.message), "error") {
			msgStyle = errorStyle
		} else {
			msgStyle = messageStyle
		}
		sections = append(sections, msgStyle.Render(m.message))
		sections = append(sections, "")
	}

	// Show game state overview
	sections = append(sections, formatGameOverview(m.gameState))

	// Show history status
	if m.history != nil {
		historyText := "History: " + intToString(m.history.UndoCount()) + " undo, " + intToString(m.history.RedoCount()) + " redo available"
		sections = append(sections, helpStyle.Render(historyText))
		sections = append(sections, "")
	}

	// Command input area
	sections = append(sections, separatorStyle.Render(strings.Repeat("─", 60)))
	inputText := "Command: " + inputStyle.Render(m.input+"█")
	sections = append(sections, inputText)
	sections = append(sections, helpStyle.Render("Type 'help' for commands, 'quit' to exit"))

	return strings.Join(sections, "\n")
}

// formatGameOverview creates a visual overview of the current game state
func formatGameOverview(gameState state.GameState) string {
	var sections []string

	// If no players or companies, show startup message
	if len(gameState.Players) == 0 && len(gameState.Companies) == 0 {
		sections = append(sections, helpStyle.Render("Ready to start! Add players and companies to begin."))
		sections = append(sections, "")
		return strings.Join(sections, "\n")
	}

	// Bank status and quick stats
	bankText := "Bank: " + moneyStyle.Render("$"+intToString(int(gameState.BankMoney)))
	actionsText := "Actions: " + intToString(len(gameState.Log))
	sections = append(sections, bankText+"  "+actionsText)
	sections = append(sections, "")

	// Players section
	if len(gameState.Players) > 0 {
		sections = append(sections, sectionStyle.Render("PLAYERS"))
		for _, player := range gameState.Players {
			totalValue := gameState.CalculatePlayerValue(player.ID)
			playerLine := "  " + playerStyle.Render(player.Name) + ": " + moneyStyle.Render("$"+intToString(int(player.Cash))) +
				" → " + moneyStyle.Render("$"+intToString(int(totalValue)))

			// Show shares owned
			if len(player.Shares) > 0 {
				shareList := ""
				first := true
				for company, count := range player.Shares {
					if count > 0 {
						if !first {
							shareList += ", "
						}
						shareList += company + ":" + intToString(int(count))
						first = false
					}
				}
				if shareList != "" {
					playerLine += " " + helpStyle.Render("["+shareList+"]")
				}
			}
			sections = append(sections, playerLine)
		}
		sections = append(sections, "")
	}

	// Companies section
	if len(gameState.Companies) > 0 {
		sections = append(sections, sectionStyle.Render("COMPANIES"))
		for _, company := range gameState.Companies {
			companyLine := "  " + companyStyle.Render(company.Name) + ": " + moneyStyle.Render("$"+intToString(int(company.Treasury)))

			// Show pricing info
			if company.ParValue > 0 {
				pricingInfo := "Par: $" + intToString(int(company.ParValue))
				if company.StockPrice > 0 && company.StockPrice != company.ParValue {
					pricingInfo += " → $" + intToString(int(company.StockPrice))
				}
				companyLine += " " + helpStyle.Render("("+pricingInfo+")")
			}

			// Share distribution
			shareInfo := "IPO:" + intToString(int(company.SharesInIPO))
			if company.SharesInBank > 0 {
				shareInfo += " Bank:" + intToString(int(company.SharesInBank))
			}
			companyLine += " " + helpStyle.Render(shareInfo)

			if company.PresidentID != "" {
				companyLine += " " + headerStyle.Render("President:"+company.PresidentID)
			}
			sections = append(sections, companyLine)
		}
		sections = append(sections, "")
	}

	return strings.Join(sections, "\n")
}
