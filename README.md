# 18xx CLI 

A simple, clean command-line banking tool for 18xx board games. 

Note: this was a project used to pilot Claude Code for professional purpues. Outside of of small edits to this readme and select other files, all code in this repo is ai generated. 

## Purpose

This CLI tool serves as a **smart ledger** for 18xx, handling:
- Stock transactions (buy/sell shares)
- Money management (player cash, company treasuries)
- Banking operations (IPO, bank pool, dividends)
- Complete transaction logging
- Undo/redo functionality for mistake correction

## What This Tool Does

✅ **Financial Operations**
- Buy shares from IPO or bank pool
- Sell shares to bank pool
- Set company par values
- Adjust stock prices manually
- Pay dividends
- Transfer money between entities
- Track player cash and company treasuries

✅ **Game Support**
- Complete action logging with timestamps
- Undo/redo for correcting mistakes
- Simple command-based interface
- Save/load game state

## What This Tool Does NOT Do

❌ **Game Rules Enforcement**
- No tile laying or route validation
- No train management
- No operating round rules
- No turn order enforcement
- No certificate limits
- No complex game logic

**Philosophy: Trust the players, just track the money.**

## Architecture

### Functional Design
```
18xxCli/
├── main.go              # Bubble Tea app entry point
├── state/               # Game state management
│   ├── types.go        # Core immutable types
│   ├── history.go      # Undo/redo system
│   └── *_test.go       # Comprehensive tests
├── actions/             # Pure transaction functions
│   ├── transactions.go # Buy/sell, money transfers
│   ├── log.go         # Action logging
│   └── *_test.go      # Transaction tests
└── ui/                  # Bubble Tea interface
    ├── model.go        # TUI model
    ├── view.go         # Display logic
    └── update.go       # Command handling
```

### Core Principles

1. **Immutable State** - All operations return new state, never mutate
2. **Functional Style** - Pure functions for all game logic
3. **Complete Undo** - Every action can be undone/redone
4. **Everything Logged** - Complete audit trail of all actions
5. **Simple Commands** - Easy to use during physical gameplay

### Key Types

```go
type GameState struct {
    Players   map[string]*Player
    Companies map[string]*Company
    BankMoney Money
    Log       []LogEntry
}

type Player struct {
    ID     string
    Name   string
    Cash   Money
    Shares map[string]ShareCount  // company -> share count
}

type Company struct {
    ID           string
    Name         string
    Treasury     Money
    StockPrice   StockPrice  // Current market price
    ParValue     StockPrice  // IPO price
    SharesInIPO  ShareCount  // Available for purchase
    SharesInBank ShareCount  // In bank pool
    TotalShares  ShareCount  // Usually 10
}
```

## Usage Examples

### Basic Operations
```bash
# Set up game
> add player alice "Alice Smith" 600
> add player bob "Bob Jones" 600
> add company BO "Baltimore & Ohio" 10

# Stock transactions
> buy ipo BO 2 alice
"Alice buys 2 B&O shares from IPO for $120"

> sell bank PRR 1 bob
"Bob sells 1 PRR share to Bank for $90"

# Banking operations
> set par BO 76
> adjust price BO up 2
> pay dividend BO 10
> transfer BO alice 200

# Mistake correction
> undo
> undo 3
> redo
```

### Complex Operations via Simple Steps
Instead of complex "operating round" commands, use sequences:
```bash
# Company operating round:
> pay dividend BO half  # Pay half to shareholders
> withhold BO 115       # Rest stays in treasury
> buy train BO 4 100    # Purchase 4-train

# IPO opening:
> set par NYC 67
> move shares NYC ipo 10
> buy ipo NYC 2 alice   # President's certificate
```

## Building and Running

```bash
# Build
go build -o 1830-banker

# Run
./1830-banker

# Test
go test ./...
```

## Game State

The tool maintains complete game state including:
- All player cash and share holdings
- All company treasuries and stock prices
- Complete transaction log with timestamps
- Full undo/redo history

State can be saved/loaded for game sessions spanning multiple days.

## Design Philosophy

This tool follows the principle of **"simple operations, complex workflows"**. Rather than encoding complex 1830 or any title rules, it provides simple, reliable building blocks that players can use to construct any game situation.

The tool acts as a smart calculator and ledger, leaving game rules and decisions to the human players while ensuring perfect record-keeping and easy mistake correction.
