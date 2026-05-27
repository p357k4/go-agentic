package main

import (
	"context"
	"log/slog"
	"os"

	"fraud-agent/agent"
	"fraud-agent/db"
	"fraud-agent/risk"
)

// scenario bundles an alert with a human-readable label for logging.
type scenario struct {
	name  string
	alert agent.TransactionAlert
}

func main() {
	slog.Info("Autonomous Fraud Detection Agent")

	if os.Getenv("GEMINI_API_KEY") == "" {
		slog.Error("GEMINI_API_KEY environment variable is required")
		os.Exit(1)
	}

	scenarios := []scenario{
		{
			name: "Suspicious",
			alert: agent.TransactionAlert{
				AccountID:    "ACC-7711",
				Amount:       7200.00,
				Currency:     "PLN",
				Country:      "Peru",
				MerchantType: "casino",
			},
		},
		{
			name: "Safe",
			alert: agent.TransactionAlert{
				AccountID:    "ACC-7711",
				Amount:       50.00,
				Currency:     "PLN",
				Country:      "PL",
				MerchantType: "local shop",
			},
		},
	}

	ctx := context.Background()

	for _, s := range scenarios {
		runScenario(ctx, s)
	}
}

// runScenario creates an isolated agent with its own database and runs
// a single investigation, logging card status before and after.
func runScenario(ctx context.Context, s scenario) {
	slog.Info("Running scenario", "name", s.name, "account_id", s.alert.AccountID,
		"amount", s.alert.Amount, "country", s.alert.Country, "merchant", s.alert.MerchantType)

	database := db.NewMockDB()
	calculator := risk.NewCalculator()

	a, err := agent.NewAgent(database, calculator)
	if err != nil {
		slog.Error("Failed to initialize agent", "scenario", s.name, "error", err)
		return
	}

	logCardStatus(database, s.alert.AccountID, "before")

	report, err := a.RunAnalysis(ctx, s.alert)
	if err != nil {
		slog.Error("Scenario failed", "scenario", s.name, "error", err)
	} else {
		slog.Info("Investigation report", "scenario", s.name, "report", report)
	}

	logCardStatus(database, s.alert.AccountID, "after")
}

func logCardStatus(database db.Database, accountID, phase string) {
	profile, err := database.GetCustomerProfile(accountID)
	if err != nil {
		slog.Warn("Could not read card status", "account_id", accountID, "phase", phase, "error", err)
		return
	}
	slog.Info("Card status", "account_id", accountID, "phase", phase, "status", profile.CardStatus)
}
