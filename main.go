package main

import (
	"context"
	"log/slog"
	"os"

	"fraud-agent/agent"
	"fraud-agent/db"
	"fraud-agent/risk"
)

func main() {
	slog.Info("Autonomous Fraud Detection Agent - Gemini 2.5/3.5 Orchestrator")

	// Verify GEMINI_API_KEY is available in the environment
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		slog.Warn("GEMINI_API_KEY environment variable is not set!")
		slog.Warn("Please run: export GEMINI_API_KEY=\"your-api-key-here\"")
		slog.Warn("Exiting application...")
		os.Exit(1)
	}

	ctx := context.Background()

	// ------------------------------------------------------------------
	// SCENARIO A: Suspicious transaction (Peru Casino, 7200 PLN)
	// ------------------------------------------------------------------
	slog.Info("RUNNING SCENARIO A (Suspicious)", "amount", 7200.00, "merchant", "casino", "location", "Peru")
	
	// Create fresh instances for the scenario
	databaseA := db.NewMockDB()
	calcA := risk.NewCalculator()

	agentA, err := agent.NewAgent(databaseA, calcA)
	if err != nil {
		slog.Error("Failed to initialize Agent", "error", err)
		os.Exit(1)
	}

	// Show initial card status
	profileBeforeA, _ := databaseA.GetCustomerProfile("ACC-7711")
	slog.Info("Initial Card Status", "account_id", profileBeforeA.AccountID, "status", profileBeforeA.CardStatus)

	alertA := agent.TransactionAlert{
		AccountID:    "ACC-7711",
		Amount:       7200.00,
		Currency:     "PLN",
		Country:      "Peru",
		MerchantType: "casino",
	}

	reportA, err := agentA.RunAnalysis(ctx, alertA)
	if err != nil {
		slog.Error("Scenario A failed during agent loop", "error", err)
	} else {
		slog.Info("AGENT INVESTIGATION REPORT (SCENARIO A)", "report", reportA)
	}

	// Show final card status
	profileAfterA, _ := databaseA.GetCustomerProfile("ACC-7711")
	slog.Info("Final Card Status", "account_id", profileAfterA.AccountID, "status", profileAfterA.CardStatus)

	// ------------------------------------------------------------------
	// SCENARIO B: Safe transaction (Poland Local Shop, 50 PLN)
	// ------------------------------------------------------------------
	slog.Info("RUNNING SCENARIO B (Safe)", "amount", 50.00, "merchant", "local shop", "location", "PL")

	// Create fresh instances to isolate databases
	databaseB := db.NewMockDB()
	calcB := risk.NewCalculator()

	agentB, err := agent.NewAgent(databaseB, calcB)
	if err != nil {
		slog.Error("Failed to initialize Agent", "error", err)
		os.Exit(1)
	}

	// Show initial card status
	profileBeforeB, _ := databaseB.GetCustomerProfile("ACC-7711")
	slog.Info("Initial Card Status", "account_id", profileBeforeB.AccountID, "status", profileBeforeB.CardStatus)

	alertB := agent.TransactionAlert{
		AccountID:    "ACC-7711",
		Amount:       50.00,
		Currency:     "PLN",
		Country:      "PL",
		MerchantType: "local shop",
	}

	reportB, err := agentB.RunAnalysis(ctx, alertB)
	if err != nil {
		slog.Error("Scenario B failed during agent loop", "error", err)
	} else {
		slog.Info("AGENT INVESTIGATION REPORT (SCENARIO B)", "report", reportB)
	}

	// Show final card status
	profileAfterB, _ := databaseB.GetCustomerProfile("ACC-7711")
	slog.Info("Final Card Status", "account_id", profileAfterB.AccountID, "status", profileAfterB.CardStatus)
}
