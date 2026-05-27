package agent

import (
	"context"
	"fmt"
	"log/slog"

	"google.golang.org/genai"
)

// maxTurns limits conversation rounds to prevent runaway loops.
const maxTurns = 5

// RunAnalysis orchestrates a multi-turn investigation of a transaction alert.
// The model calls tools as needed; the loop dispatches them via registered handlers
// and feeds results back until the model produces a final text verdict.
func (a *Agent) RunAnalysis(ctx context.Context, alert TransactionAlert) (string, error) {
	slog.Info("Starting investigation",
		"account_id", alert.AccountID,
		"amount", alert.Amount,
		"currency", alert.Currency,
		"country", alert.Country,
		"merchant_type", alert.MerchantType,
	)

	contents := []*genai.Content{
		genai.NewContentFromText(buildUserPrompt(alert), "user"),
	}

	config := &genai.GenerateContentConfig{
		SystemInstruction: genai.NewContentFromText(systemInstruction, ""),
		Tools:             a.tools,
	}

	for turn := 1; turn <= maxTurns; turn++ {
		logRequest(turn, contents[len(contents)-1])

		resp, err := a.client.Models.GenerateContent(ctx, a.modelName, contents, config)
		if err != nil {
			return "", fmt.Errorf("gemini call failed on turn %d: %w", turn, err)
		}

		candidate, err := extractCandidate(resp)
		if err != nil {
			return "", err
		}

		logResponse(turn, candidate.Content)
		contents = append(contents, candidate.Content)

		calls, text := classifyParts(candidate.Content.Parts)
		if len(calls) == 0 {
			slog.Info("Investigation complete")
			return text, nil
		}

		responseParts, err := a.dispatchTools(ctx, calls)
		if err != nil {
			return "", err
		}

		contents = append(contents, &genai.Content{Role: "user", Parts: responseParts})
	}

	return "", fmt.Errorf("investigation did not complete within %d turns", maxTurns)
}

// extractCandidate validates and returns the first candidate from the response.
func extractCandidate(resp *genai.GenerateContentResponse) (*genai.Candidate, error) {
	if len(resp.Candidates) == 0 || resp.Candidates[0].Content == nil {
		return nil, fmt.Errorf("received empty response from model")
	}
	return resp.Candidates[0], nil
}

// classifyParts separates function calls from text in a response.
func classifyParts(parts []*genai.Part) ([]*genai.FunctionCall, string) {
	var calls []*genai.FunctionCall
	var text string

	for _, p := range parts {
		if p.FunctionCall != nil {
			calls = append(calls, p.FunctionCall)
		}
		if p.Text != "" {
			if text != "" {
				text += "\n"
			}
			text += p.Text
		}
	}
	return calls, text
}

// dispatchTools executes each function call via the registered handler
// and returns the response parts with matching IDs for the Gemini API.
func (a *Agent) dispatchTools(ctx context.Context, calls []*genai.FunctionCall) ([]*genai.Part, error) {
	parts := make([]*genai.Part, 0, len(calls))

	for _, fc := range calls {
		slog.Info("Dispatching tool", "name", fc.Name, "call_id", fc.ID)

		handler, ok := a.handlers[fc.Name]
		if !ok {
			return nil, fmt.Errorf("no handler registered for tool %q", fc.Name)
		}

		result, err := handler(ctx, fc.Args)
		if err != nil {
			result = map[string]any{"error": err.Error()}
		}

		parts = append(parts, &genai.Part{
			FunctionResponse: &genai.FunctionResponse{
				Name:     fc.Name,
				ID:       fc.ID, // Must match the call ID for Gemini API validation
				Response: result,
			},
		})
	}
	return parts, nil
}
