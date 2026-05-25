package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"fraud-agent/agent"
	"fraud-agent/db"
	"fraud-agent/risk"
)

func main() {
	log.Println("==================================================================")
	log.Println("  Autonomous Fraud Detection Agent - Gemini 2.5/3.5 Orchestrator  ")
	log.Println("==================================================================")

	// Verify GEMINI_API_KEY is available in the environment
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		log.Println("[WARNING] GEMINI_API_KEY environment variable is not set!")
		log.Println("Please run: export GEMINI_API_KEY=\"your-api-key-here\"")
		log.Println("Exiting application...")
		os.Exit(1)
	}

	ctx := context.Background()

	// ------------------------------------------------------------------
	// SCENARIO A: Suspicious transaction (Peru Casino, 7200 PLN)
	// ------------------------------------------------------------------
	log.Println("\n>>> RUNNING SCENARIO A (Podejrzany): Peru Casino, 7200 PLN")
	
	// Create fresh instances for the scenario
	databaseA := db.NewMockDB()
	calcA := risk.NewCalculator()

	agentA, err := agent.NewAgent(databaseA, calcA)
	if err != nil {
		log.Fatalf("Failed to initialize Agent: %v", err)
	}

	// Show initial card status
	profileBeforeA, _ := databaseA.GetCustomerProfile("ACC-7711")
	fmt.Printf("[DB Before] Card Status for Jan Kowalski (%s): %s\n", profileBeforeA.AccountID, profileBeforeA.CardStatus)

	alertA := agent.TransactionAlert{
		AccountID:    "ACC-7711",
		Amount:       7200.00,
		Currency:     "PLN",
		Country:      "Peru",
		MerchantType: "casino",
	}

	reportA, err := agentA.RunAnalysis(ctx, alertA)
	if err != nil {
		log.Printf("[Error] Scenario A failed during agent loop: %v\n", err)
	} else {
		log.Println("\n--- AGENT INVESTIGATION REPORT (SCENARIO A) ---")
		fmt.Println(reportA)
		log.Println("-----------------------------------------------")
	}

	// Show final card status
	profileAfterA, _ := databaseA.GetCustomerProfile("ACC-7711")
	fmt.Printf("[DB After] Card Status for Jan Kowalski (%s): %s\n", profileAfterA.AccountID, profileAfterA.CardStatus)

	// ------------------------------------------------------------------
	// SCENARIO B: Safe transaction (Poland Local Shop, 50 PLN)
	// ------------------------------------------------------------------
	log.Println("\n>>> RUNNING SCENARIO B (Bezpieczny): Poland Local Shop, 50 PLN")

	// Create fresh instances to isolate databases
	databaseB := db.NewMockDB()
	calcB := risk.NewCalculator()

	agentB, err := agent.NewAgent(databaseB, calcB)
	if err != nil {
		log.Fatalf("Failed to initialize Agent: %v", err)
	}

	// Show initial card status
	profileBeforeB, _ := databaseB.GetCustomerProfile("ACC-7711")
	fmt.Printf("[DB Before] Card Status for Jan Kowalski (%s): %s\n", profileBeforeB.AccountID, profileBeforeB.CardStatus)

	alertB := agent.TransactionAlert{
		AccountID:    "ACC-7711",
		Amount:       50.00,
		Currency:     "PLN",
		Country:      "PL",
		MerchantType: "local shop",
	}

	reportB, err := agentB.RunAnalysis(ctx, alertB)
	if err != nil {
		log.Printf("[Error] Scenario B failed during agent loop: %v\n", err)
	} else {
		log.Println("\n--- AGENT INVESTIGATION REPORT (SCENARIO B) ---")
		fmt.Println(reportB)
		log.Println("-----------------------------------------------")
	}

	// Show final card status
	profileAfterB, _ := databaseB.GetCustomerProfile("ACC-7711")
	fmt.Printf("[DB After] Card Status for Jan Kowalski (%s): %s\n", profileAfterB.AccountID, profileAfterB.CardStatus)
}
