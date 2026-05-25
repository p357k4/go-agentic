package agent

import (
	"context"
	"os"

	"fraud-agent/db"
	"fraud-agent/risk"
	"google.golang.org/genai"
)

// Agent wraps the GenAI client, the mock DB, the risk calculator, and registered tools.
type Agent struct {
	Client     *genai.Client
	ModelName  string
	DB         *db.MockDB
	Calculator *risk.Calculator
	Tools      []*genai.Tool
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

	// getCustomerProfile tool schema
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

	// executeCardSuspension tool schema
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

	tools := []*genai.Tool{
		{
			FunctionDeclarations: []*genai.FunctionDeclaration{
				getCustomerProfileDecl,
				executeCardSuspensionDecl,
			},
		},
	}

	return &Agent{
		Client:     client,
		ModelName:  modelName,
		DB:         database,
		Calculator: calculator,
		Tools:      tools,
	}, nil
}
