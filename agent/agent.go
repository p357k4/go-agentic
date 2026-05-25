package agent

import (
	"context"
	"errors"
	"fmt"
	"os"

	"fraud-agent/db"
	"fraud-agent/risk"
	"google.golang.org/genai"
)

// ToolHandler defines the execution signature for a model tool.
type ToolHandler func(ctx context.Context, args map[string]any) (map[string]any, error)

// Agent wraps the GenAI client, mock DB, risk calculator, schemas, and execution handlers.
type Agent struct {
	Client     *genai.Client
	ModelName  string
	DB         *db.MockDB
	Calculator *risk.Calculator
	Tools      []*genai.Tool
	Handlers   map[string]ToolHandler
}

// NewAgent initializes the Gemini client and builds the Function Calling tools config.
func NewAgent(database *db.MockDB, calculator *risk.Calculator) (*Agent, error) {
	ctx := context.Background()

	// NewClient automatically retrieves the GEMINI_API_KEY from environment variables.
	client, err := genai.NewClient(ctx, nil)
	if err != nil {
		return nil, err
	}

	// Allow model customization, default to "gemini-2.5-flash"
	modelName := os.Getenv("GEMINI_MODEL")
	if modelName == "" {
		modelName = "gemini-2.5-flash"
	}

	a := &Agent{
		Client:     client,
		ModelName:  modelName,
		DB:         database,
		Calculator: calculator,
		Tools:      []*genai.Tool{{}}, // Initialize empty tool container
		Handlers:   make(map[string]ToolHandler),
	}

	// Register default tools
	a.registerDefaultTools()

	return a, nil
}

// RegisterTool registers both the tool schema and its execution handler.
func (a *Agent) RegisterTool(decl *genai.FunctionDeclaration, handler ToolHandler) {
	if len(a.Tools) == 0 {
		a.Tools = append(a.Tools, &genai.Tool{})
	}
	a.Tools[0].FunctionDeclarations = append(a.Tools[0].FunctionDeclarations, decl)
	a.Handlers[decl.Name] = handler
}

// registerDefaultTools configures getCustomerProfile and executeCardSuspension.
func (a *Agent) registerDefaultTools() {
	// 1. getCustomerProfile
	getCustomerProfileDecl := &genai.FunctionDeclaration{
		Name:        "getCustomerProfile",
		Description: "Retrieves a customer's profile, home country, typical transaction size, and current card status by account ID.",
		Parameters: &genai.Schema{
			Type: genai.TypeObject,
			Properties: map[string]*genai.Schema{
				"account_id": {
					Type:        genai.TypeString,
					Description: "The customer account identifier (e.g. ACC-7711).",
				},
			},
			Required: []string{"account_id"},
		},
	}
	a.RegisterTool(getCustomerProfileDecl, func(ctx context.Context, args map[string]any) (map[string]any, error) {
		accountID, ok := args["account_id"].(string)
		if !ok {
			return nil, errors.New("missing or invalid account_id parameter")
		}
		profile, err := a.DB.GetCustomerProfile(accountID)
		if err != nil {
			return nil, err
		}
		return map[string]any{
			"account_id":     profile.AccountID,
			"customer_name":  profile.CustomerName,
			"home_country":   profile.HomeCountry,
			"typical_amount": profile.TypicalAmount,
			"card_status":    profile.CardStatus,
		}, nil
	})

	// 2. executeCardSuspension
	executeCardSuspensionDecl := &genai.FunctionDeclaration{
		Name:        "executeCardSuspension",
		Description: "Suspends the card associated with the account ID to prevent further unauthorized usage.",
		Parameters: &genai.Schema{
			Type: genai.TypeObject,
			Properties: map[string]*genai.Schema{
				"account_id": {
					Type:        genai.TypeString,
					Description: "The customer account identifier for which the card should be blocked (e.g. ACC-7711).",
				},
			},
			Required: []string{"account_id"},
		},
	}
	a.RegisterTool(executeCardSuspensionDecl, func(ctx context.Context, args map[string]any) (map[string]any, error) {
		accountID, ok := args["account_id"].(string)
		if !ok {
			return nil, errors.New("missing or invalid account_id parameter")
		}
		err := a.DB.SuspendCard(accountID)
		if err != nil {
			return nil, err
		}
		return map[string]any{
			"status":  "success",
			"message": fmt.Sprintf("Card suspended successfully for account %s", accountID),
		}, nil
	})
}
