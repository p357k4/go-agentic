package agent

import (
	"context"
	"errors"
	"fmt"

	"google.golang.org/genai"
)

// registerDefaultTools wires up the standard fraud-detection tools.
func registerDefaultTools(a *Agent) {
	a.RegisterTool(getCustomerProfileDecl(), getCustomerProfileHandler(a))
	a.RegisterTool(executeCardSuspensionDecl(), executeCardSuspensionHandler(a))
}

// --- getCustomerProfile ---

func getCustomerProfileDecl() *genai.FunctionDeclaration {
	return &genai.FunctionDeclaration{
		Name:        "getCustomerProfile",
		Description: "Retrieves a customer's profile, home country, typical transaction size, and current card status.",
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
}

func getCustomerProfileHandler(a *Agent) ToolHandler {
	return func(ctx context.Context, args map[string]any) (map[string]any, error) {
		accountID, ok := args["account_id"].(string)
		if !ok {
			return nil, errors.New("missing or invalid account_id parameter")
		}
		profile, err := a.db.GetCustomerProfile(accountID)
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
	}
}

// --- executeCardSuspension ---

func executeCardSuspensionDecl() *genai.FunctionDeclaration {
	return &genai.FunctionDeclaration{
		Name:        "executeCardSuspension",
		Description: "Suspends the card associated with the account ID to prevent unauthorized usage.",
		Parameters: &genai.Schema{
			Type: genai.TypeObject,
			Properties: map[string]*genai.Schema{
				"account_id": {
					Type:        genai.TypeString,
					Description: "The customer account identifier to block (e.g. ACC-7711).",
				},
			},
			Required: []string{"account_id"},
		},
	}
}

func executeCardSuspensionHandler(a *Agent) ToolHandler {
	return func(ctx context.Context, args map[string]any) (map[string]any, error) {
		accountID, ok := args["account_id"].(string)
		if !ok {
			return nil, errors.New("missing or invalid account_id parameter")
		}
		if err := a.db.SuspendCard(accountID); err != nil {
			return nil, err
		}
		return map[string]any{
			"status":  "success",
			"message": fmt.Sprintf("Card suspended for account %s", accountID),
		}, nil
	}
}
