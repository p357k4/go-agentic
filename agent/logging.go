package agent

import (
	"fmt"
	"log/slog"

	"google.golang.org/genai"
)

// logRequest logs the content being sent to the model in the current turn.
func logRequest(turn int, content *genai.Content) {
	slog.Info("Request sent", "turn", turn, "role", content.Role)
	for i, part := range content.Parts {
		switch {
		case part.Text != "":
			slog.Debug("Request part", "turn", turn, "index", i, "type", "text", "content", part.Text)
		case part.FunctionResponse != nil:
			slog.Info("Request part",
				"turn", turn, "index", i, "type", "tool_response",
				"name", part.FunctionResponse.Name,
				"id", part.FunctionResponse.ID,
				"response", fmt.Sprintf("%v", part.FunctionResponse.Response),
			)
		}
	}
}

// logResponse logs the content received from the model in the current turn.
func logResponse(turn int, content *genai.Content) {
	slog.Info("Response received", "turn", turn, "role", content.Role)
	for i, part := range content.Parts {
		switch {
		case part.Text != "":
			slog.Debug("Response part", "turn", turn, "index", i, "type", "text", "content", part.Text)
		case part.FunctionCall != nil:
			slog.Info("Response part",
				"turn", turn, "index", i, "type", "tool_call",
				"name", part.FunctionCall.Name,
				"id", part.FunctionCall.ID,
				"args", fmt.Sprintf("%v", part.FunctionCall.Args),
			)
		}
	}
}
