package agent_test

import (
	"context"
	"os"
	"testing"

	"fraud-agent/agent"
	"fraud-agent/db"
	"fraud-agent/risk"
)

// TestCalculator verifies that the deterministic anomaly risk calculation
// accurately follows the formula: Arisk = 0.4 * Vgeo + 0.4 * Dval + 0.2 * Ibeh
func TestCalculator(t *testing.T) {
	calc := risk.NewCalculator()

	tests := []struct {
		name     string
		vGeo     float64
		dVal     float64
		iBeh     float64
		expected float64
	}{
		{"No Risk Factors", 0.0, 0.0, 0.0, 0.0},
		{"Maximum Risk Factors", 1.0, 1.0, 1.0, 1.0},
		{"Geographical Only", 1.0, 0.0, 0.0, 0.4},
		{"Value Only", 0.0, 1.0, 0.0, 0.4},
		{"Behavioral Only", 0.0, 0.0, 1.0, 0.2},
		{"Combined High Risk", 1.0, 1.0, 0.0, 0.8},
		{"Combined Low Risk", 0.0, 0.2, 0.0, 0.08},
		{"Out of Bounds High", 2.0, 1.5, 1.2, 1.0},  // checks clamping
		{"Out of Bounds Low", -1.0, -0.5, 0.0, 0.0}, // checks clamping
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calc.Calculate(tt.vGeo, tt.dVal, tt.iBeh)
			// Using small delta for float comparison
			diff := got - tt.expected
			if diff < -0.0001 || diff > 0.0001 {
				t.Errorf("Calculate(%.2f, %.2f, %.2f) = %.4f; expected %.4f",
					tt.vGeo, tt.dVal, tt.iBeh, got, tt.expected)
			}
		})
	}
}

// TestMockDB verifies the basic CRUD operations and card suspension on the mock database.
func TestMockDB(t *testing.T) {
	database := db.NewMockDB()

	// 1. Test retrieval of seeded profile
	profile, err := database.GetCustomerProfile("ACC-7711")
	if err != nil {
		t.Fatalf("Failed to retrieve existing account ACC-7711: %v", err)
	}

	if profile.CustomerName != "John Doe" {
		t.Errorf("Expected name 'John Doe', got '%s'", profile.CustomerName)
	}

	if profile.CardStatus != "ACTIVE" {
		t.Errorf("Expected initial status 'ACTIVE', got '%s'", profile.CardStatus)
	}

	// 2. Test retrieving non-existent account
	_, err = database.GetCustomerProfile("ACC-UNKNOWN")
	if err == nil {
		t.Error("Expected error retrieving non-existent account, got nil")
	}

	// 3. Test suspending a card
	err = database.SuspendCard("ACC-7711")
	if err != nil {
		t.Fatalf("Failed to suspend card: %v", err)
	}

	profileAfter, err := database.GetCustomerProfile("ACC-7711")
	if err != nil {
		t.Fatalf("Failed to retrieve profile after suspension: %v", err)
	}

	if profileAfter.CardStatus != "SUSPENDED" {
		t.Errorf("Expected status 'SUSPENDED', got '%s'", profileAfter.CardStatus)
	}

	// 4. Test suspending an already suspended card
	err = database.SuspendCard("ACC-7711")
	if err == nil {
		t.Error("Expected error suspending already suspended card, got nil")
	}
}

// TestAgentToolsDeclaration verifies the tool registration details.
func TestAgentToolsDeclaration(t *testing.T) {
	// Setup dummy DB and calculator
	database := db.NewMockDB()
	calc := risk.NewCalculator()

	// Momentarily set standard API key so initialization does not crash if validation occurs
	os.Setenv("GEMINI_API_KEY", "dummy-key-for-init-test")
	defer os.Unsetenv("GEMINI_API_KEY")

	a, err := agent.NewAgent(database, calc)
	if err != nil {
		t.Fatalf("Failed to initialize Agent: %v", err)
	}

	decls := a.ToolDeclarations()
	if len(decls) != 2 {
		t.Fatalf("Expected 2 function declarations, got %d", len(decls))
	}

	names := map[string]bool{
		decls[0].Name: true,
		decls[1].Name: true,
	}

	if !names["getCustomerProfile"] || !names["executeCardSuspension"] {
		t.Errorf("Registered tool names mismatch. Got: %v", names)
	}
}

// TestLiveAgentWorkflow executes the agent orchestrator end-to-end loop.
// It will be run only if a valid GEMINI_API_KEY environment variable is provided.
func TestLiveAgentWorkflow(t *testing.T) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" || apiKey == "dummy-key-for-init-test" {
		t.Skip("Skipping live Gemini API integration tests as GEMINI_API_KEY is not set")
	}

	database := db.NewMockDB()
	calc := risk.NewCalculator()

	a, err := agent.NewAgent(database, calc)
	if err != nil {
		t.Fatalf("Failed to initialize Agent: %v", err)
	}

	ctx := context.Background()

	// 1. Run Analysis on Safe transaction (Scenario B)
	alertSafe := agent.TransactionAlert{
		AccountID:    "ACC-7711",
		Amount:       50.00,
		Currency:     "PLN",
		Country:      "PL",
		MerchantType: "local shop",
	}

	reportSafe, err := a.RunAnalysis(ctx, alertSafe)
	if err != nil {
		t.Errorf("Safe transaction analysis failed: %v", err)
	}

	if reportSafe == "" {
		t.Error("Expected an investigation report, got empty string")
	}

	// Verify card remains ACTIVE
	profileSafe, _ := database.GetCustomerProfile("ACC-7711")
	if profileSafe.CardStatus != "ACTIVE" {
		t.Errorf("Expected card to remain ACTIVE, got '%s'", profileSafe.CardStatus)
	}

	// 2. Run Analysis on Suspicious transaction (Scenario A)
	alertSuspicious := agent.TransactionAlert{
		AccountID:    "ACC-7711",
		Amount:       7200.00,
		Currency:     "PLN",
		Country:      "Peru",
		MerchantType: "casino",
	}

	reportSuspicious, err := a.RunAnalysis(ctx, alertSuspicious)
	if err != nil {
		t.Errorf("Suspicious transaction analysis failed: %v", err)
	}

	if reportSuspicious == "" {
		t.Error("Expected an investigation report, got empty string")
	}

	// Verify card is SUSPENDED
	profileSuspicious, _ := database.GetCustomerProfile("ACC-7711")
	if profileSuspicious.CardStatus != "SUSPENDED" {
		t.Errorf("Expected card to be SUSPENDED, got '%s'", profileSuspicious.CardStatus)
	}
}
