// Package agent implements the Gemini-powered fraud detection agent.
//
// Architecture:
//   - agent.go:   Agent struct, client init, tool registry
//   - tools.go:   Tool schema declarations and handler implementations
//   - prompt.go:  System instruction and prompt templates
//   - loop.go:    Multi-turn orchestration loop
//   - logging.go: Structured request/response audit logging
package agent

import (
	"context"
	"os"

	"fraud-agent/db"
	"fraud-agent/risk"

	"google.golang.org/genai"
)

// ToolHandler is a callback that executes a tool and returns a JSON-serializable result.
type ToolHandler func(ctx context.Context, args map[string]any) (map[string]any, error)

// TransactionAlert describes an incoming transaction to be investigated.
type TransactionAlert struct {
	AccountID    string  `json:"account_id"`
	Amount       float64 `json:"amount"`
	Currency     string  `json:"currency"`
	Country      string  `json:"country"`
	MerchantType string  `json:"merchant_type"`
}

// Agent orchestrates fraud investigations using Gemini function calling.
type Agent struct {
	client    *genai.Client
	modelName string
	db        db.Database
	calc      *risk.Calculator
	tools     []*genai.Tool
	handlers  map[string]ToolHandler
}

// NewAgent creates and returns an Agent ready for use.
// It reads GEMINI_API_KEY and optionally GEMINI_MODEL from the environment.
func NewAgent(database db.Database, calculator *risk.Calculator) (*Agent, error) {
	client, err := genai.NewClient(context.Background(), nil)
	if err != nil {
		return nil, err
	}

	modelName := os.Getenv("GEMINI_MODEL")
	if modelName == "" {
		modelName = "gemini-2.5-flash"
	}

	a := &Agent{
		client:    client,
		modelName: modelName,
		db:        database,
		calc:      calculator,
		handlers:  make(map[string]ToolHandler),
	}

	registerDefaultTools(a)

	return a, nil
}

// RegisterTool adds a tool declaration and its handler to the agent.
func (a *Agent) RegisterTool(decl *genai.FunctionDeclaration, handler ToolHandler) {
	if len(a.tools) == 0 {
		a.tools = append(a.tools, &genai.Tool{})
	}
	a.tools[0].FunctionDeclarations = append(a.tools[0].FunctionDeclarations, decl)
	a.handlers[decl.Name] = handler
}

// ToolDeclarations returns the registered tool declarations for testing.
func (a *Agent) ToolDeclarations() []*genai.FunctionDeclaration {
	if len(a.tools) == 0 {
		return nil
	}
	return a.tools[0].FunctionDeclarations
}
