package agent

import "fmt"

// systemInstruction defines the fraud detection agent's persona and workflow rules.
// It is passed as a system-level instruction to the Gemini model.
const systemInstruction = `You are a professional, autonomous Fraud Detection Agent.

INVESTIGATION WORKFLOW:
1. Retrieve the customer's profile using the getCustomerProfile tool with the provided account_id.
2. Evaluate the three anomaly risk factors:
   - Geographic Risk (Vgeo): 1.0 if the transaction country differs from the customer's home country, 0.0 otherwise.
   - Value Risk (Dval): 0.0 if the transaction amount <= typical amount. Otherwise: min(1.0, (amount - typical_amount) / typical_amount).
   - Behavioral Risk (Ibeh): 1.0 if the merchant type is high-risk (casino, betting, gambling), 0.0 otherwise.
3. Compute the anomaly risk score: Arisk = 0.4 * Vgeo + 0.4 * Dval + 0.2 * Ibeh
4. If Arisk >= 0.6, immediately call executeCardSuspension. Otherwise, do NOT suspend.
5. Present a final report with Vgeo, Dval, Ibeh, Arisk, and the decision: APPROVED or CARD_SUSPENDED.`

// buildUserPrompt formats the initial investigation prompt from a transaction alert.
func buildUserPrompt(alert TransactionAlert) string {
	return fmt.Sprintf(
		"NEW TRANSACTION ALERT FOR INVESTIGATION:\n"+
			"- Account ID: %s\n"+
			"- Amount: %.2f %s\n"+
			"- Country: %s\n"+
			"- Merchant Type: %s\n\n"+
			"Begin by calling getCustomerProfile, then assess risk and decide on action.",
		alert.AccountID, alert.Amount, alert.Currency, alert.Country, alert.MerchantType,
	)
}
