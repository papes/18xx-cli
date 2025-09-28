package logging

import (
	"fmt"
	"time"

	"github.com/papes/18xxCli/state"
)

// Logger provides structured logging for game actions with consistent formatting
type Logger struct {
	timezone *time.Location
}

// NewLogger creates a new logger instance with UTC timezone by default
func NewLogger() *Logger {
	return &Logger{
		timezone: time.UTC,
	}
}

// SetTimezone sets the timezone for log timestamps
func (l *Logger) SetTimezone(tz *time.Location) {
	l.timezone = tz
}

// LogPlayerAction creates a structured log entry for player-related actions
func (l *Logger) LogPlayerAction(action string, playerID, playerName string, details map[string]interface{}) state.LogEntry {
	if details == nil {
		details = make(map[string]interface{})
	}
	details["player_id"] = playerID
	details["player_name"] = playerName

	return state.LogEntry{
		Timestamp: time.Now().In(l.timezone),
		Action:    action,
		Details:   details,
	}
}

// LogCompanyAction creates a structured log entry for company-related actions
func (l *Logger) LogCompanyAction(action string, companyID string, details map[string]interface{}) state.LogEntry {
	if details == nil {
		details = make(map[string]interface{})
	}
	details["company"] = companyID

	return state.LogEntry{
		Timestamp: time.Now().In(l.timezone),
		Action:    action,
		Details:   details,
	}
}

// LogShareTransaction creates a structured log entry for share transactions
func (l *Logger) LogShareTransaction(action string, playerID, companyID string, shares state.ShareCount, price state.StockPrice, total state.Money) state.LogEntry {
	return state.LogEntry{
		Timestamp: time.Now().In(l.timezone),
		Action:    action,
		Details: map[string]interface{}{
			"player":  playerID,
			"company": companyID,
			"shares":  shares,
			"price":   price,
			"total":   total,
		},
	}
}

// LogMoneyTransfer creates a structured log entry for money transfers
func (l *Logger) LogMoneyTransfer(fromID, toID string, amount state.Money, transferType string) state.LogEntry {
	return state.LogEntry{
		Timestamp: time.Now().In(l.timezone),
		Action:    "transfer_money",
		Details: map[string]interface{}{
			"from":   fromID,
			"to":     toID,
			"amount": amount,
			"type":   transferType,
		},
	}
}

// LogDividendPayment creates a structured log entry for dividend payments
func (l *Logger) LogDividendPayment(companyID, recipientID string, recipientType string, shares state.ShareCount, dividendPerShare, amountReceived, totalDividend state.Money) state.LogEntry {
	return state.LogEntry{
		Timestamp: time.Now().In(l.timezone),
		Action:    "dividend_payment",
		Details: map[string]interface{}{
			"company":            companyID,
			"recipient":          recipientID,
			"recipient_type":     recipientType,
			"shares_owned":       shares,
			"dividend_per_share": dividendPerShare,
			"amount_received":    amountReceived,
			"total_dividend":     totalDividend,
		},
	}
}

// LogRevenue creates a structured log entry for company revenue
func (l *Logger) LogRevenue(companyID string, amount state.Money, revenueType string) state.LogEntry {
	return state.LogEntry{
		Timestamp: time.Now().In(l.timezone),
		Action:    revenueType + "_revenue",
		Details: map[string]interface{}{
			"company": companyID,
			"amount":  amount,
		},
	}
}

// LogPriceChange creates a structured log entry for stock price changes
func (l *Logger) LogPriceChange(action string, companyID string, oldPrice, newPrice state.StockPrice) state.LogEntry {
	return state.LogEntry{
		Timestamp: time.Now().In(l.timezone),
		Action:    action,
		Details: map[string]interface{}{
			"company":   companyID,
			"old_price": oldPrice,
			"new_price": newPrice,
		},
	}
}

// LogTrainPurchase creates a structured log entry for train purchases
func (l *Logger) LogTrainPurchase(companyID, trainType string, cost state.Money) state.LogEntry {
	return state.LogEntry{
		Timestamp: time.Now().In(l.timezone),
		Action:    "buy_train",
		Details: map[string]interface{}{
			"company":    companyID,
			"train_type": trainType,
			"cost":       cost,
		},
	}
}

// LogPresidentChange creates a structured log entry for president changes
func (l *Logger) LogPresidentChange(companyID, oldPresidentID, newPresidentID string) state.LogEntry {
	return state.LogEntry{
		Timestamp: time.Now().In(l.timezone),
		Action:    "set_president",
		Details: map[string]interface{}{
			"company":       companyID,
			"old_president": oldPresidentID,
			"new_president": newPresidentID,
		},
	}
}

// LogGameSetup creates a structured log entry for game setup actions
func (l *Logger) LogGameSetup(action string, details map[string]interface{}) state.LogEntry {
	return state.LogEntry{
		Timestamp: time.Now().In(l.timezone),
		Action:    action,
		Details:   details,
	}
}

// FormatLogEntry formats a log entry for display with consistent formatting
func (l *Logger) FormatLogEntry(entry state.LogEntry) string {
	timestamp := entry.Timestamp.Format("15:04:05")
	action := entry.Action

	// Format details based on action type
	var details string
	if entry.Details != nil {
		switch entry.Action {
		case "buy_share_ipo", "buy_share_bank":
			if player, ok := entry.Details["player"].(string); ok {
				if company, ok := entry.Details["company"].(string); ok {
					if shares, ok := entry.Details["shares"]; ok {
						if total, ok := entry.Details["total"]; ok {
							details = fmt.Sprintf("%s buys %v shares of %s for $%v", player, shares, company, total)
						}
					}
				}
			}
		case "sell_share_bank":
			if player, ok := entry.Details["player"].(string); ok {
				if company, ok := entry.Details["company"].(string); ok {
					if shares, ok := entry.Details["shares"]; ok {
						if total, ok := entry.Details["total"]; ok {
							details = fmt.Sprintf("%s sells %v shares of %s for $%v", player, shares, company, total)
						}
					}
				}
			}
		case "dividend_payment":
			if recipient, ok := entry.Details["recipient"].(string); ok {
				if company, ok := entry.Details["company"].(string); ok {
					if amount, ok := entry.Details["amount_received"]; ok {
						if shares, ok := entry.Details["shares_owned"]; ok {
							recipientType, _ := entry.Details["recipient_type"].(string)
							if recipientType == "company" {
								details = fmt.Sprintf("%s receives $%v (self-payout)", company, amount)
							} else {
								details = fmt.Sprintf("%s receives $%v (%v shares of %s)", recipient, amount, shares, company)
							}
						}
					}
				}
			}
		case "transfer_money":
			if from, ok := entry.Details["from"].(string); ok {
				if to, ok := entry.Details["to"].(string); ok {
					if amount, ok := entry.Details["amount"]; ok {
						details = fmt.Sprintf("%s → %s: $%v", from, to, amount)
					}
				}
			}
		case "company_revenue", "withhold_revenue":
			if company, ok := entry.Details["company"].(string); ok {
				if amount, ok := entry.Details["amount"]; ok {
					details = fmt.Sprintf("%s: $%v", company, amount)
				}
			}
		case "set_par_value":
			if company, ok := entry.Details["company"].(string); ok {
				if parValue, ok := entry.Details["par_value"]; ok {
					details = fmt.Sprintf("%s par value set to $%v", company, parValue)
				}
			}
		case "set_stock_price":
			if company, ok := entry.Details["company"].(string); ok {
				if newPrice, ok := entry.Details["new_price"]; ok {
					details = fmt.Sprintf("%s price set to $%v", company, newPrice)
				}
			}
		default:
			// For other actions, show key details
			if company, ok := entry.Details["company"].(string); ok {
				details = company
			}
		}
	}

	if details != "" {
		return fmt.Sprintf("%s - %s: %s", timestamp, action, details)
	}
	return fmt.Sprintf("%s - %s", timestamp, action)
}

// ExportLog exports the game log in a structured format
func (l *Logger) ExportLog(gameState state.GameState, format string) (string, error) {
	switch format {
	case "json":
		return l.exportJSON(gameState.Log)
	case "csv":
		return l.exportCSV(gameState.Log)
	case "text":
		return l.exportText(gameState.Log)
	default:
		return "", fmt.Errorf("unsupported export format: %s", format)
	}
}

func (l *Logger) exportJSON(log []state.LogEntry) (string, error) {
	// In a real implementation, we would use encoding/json
	// For now, return a simple JSON-like format
	result := "[\n"
	for i, entry := range log {
		result += fmt.Sprintf("  {\n    \"timestamp\": \"%s\",\n    \"action\": \"%s\"",
			entry.Timestamp.Format(time.RFC3339), entry.Action)
		if entry.Details != nil {
			result += ",\n    \"details\": " + fmt.Sprintf("%v", entry.Details)
		}
		result += "\n  }"
		if i < len(log)-1 {
			result += ","
		}
		result += "\n"
	}
	result += "]"
	return result, nil
}

func (l *Logger) exportCSV(log []state.LogEntry) (string, error) {
	result := "Timestamp,Action,Details\n"
	for _, entry := range log {
		details := ""
		if entry.Details != nil {
			details = fmt.Sprintf("%v", entry.Details)
		}
		result += fmt.Sprintf("%s,%s,\"%s\"\n",
			entry.Timestamp.Format(time.RFC3339), entry.Action, details)
	}
	return result, nil
}

func (l *Logger) exportText(log []state.LogEntry) (string, error) {
	result := "=== Game Log ===\n"
	for i, entry := range log {
		result += fmt.Sprintf("%d. %s\n", i+1, l.FormatLogEntry(entry))
	}
	return result, nil
}