# Autonomous Fraud Investigation Agent in Go

An autonomous cybersecurity and transaction risk-assessment agent implemented in **Go** using the modern, official Google GenAI SDK (`google.golang.org/genai`).

This agent evaluates incoming transaction alerts, executes dynamic tool calls (Function Calling) to fetch customer profiles and suspend cards, and leverages a mathematical model to assess transaction risk.

---

## 🏗️ System Architecture & Orchestration Flow

```mermaid
sequenceDiagram
    autonumber
    participant App as Main Application
    participant Loop as Agent Orchestrator Loop
    participant Gemini as Gemini 2.5/3.5 Model
    participant DB as Thread-safe Mock DB

    App->>Loop: RunAnalysis(alert details)
    Loop->>Gemini: Init prompt with Tools schema & System Instructions
    Gemini->>Loop: Tool Call: getCustomerProfile(account_id)
    Loop->>DB: Fetch profile for John Doe (ACC-7711)
    DB-->>Loop: Profile: Poland, Typical Amount: 100 PLN, status ACTIVE
    Loop->>Gemini: FunctionResponse: Profile details (matching ID)
    Note over Gemini: Gemini evaluates risk factors:<br/>Vgeo (Peru vs PL = 1.0)<br/>Dval (7200 vs 100 = 1.0)<br/>Ibeh (casino = 1.0)<br/>Arisk = 0.4*1.0 + 0.4*1.0 + 0.2*1.0 = 1.0
    Gemini->>Loop: Tool Call: executeCardSuspension(account_id)
    Loop->>DB: Suspend card in database
    DB-->>Loop: Status SUCCESS
    Loop->>Gemini: FunctionResponse: Confirmation (matching ID)
    Gemini->>Loop: Final report text
    Loop-->>App: Detailed investigation summary & final decision
```

---

## 📈 Anomaly Risk Calculator Model ($A_{risk}$)

Risk calculation is handled by a deterministic multi-factor formula defined in `risk/calculator.go`:

$$A_{risk} = \phi \cdot V_{geo} + \psi \cdot D_{val} + \omega \cdot I_{beh}$$

Where standard weights are configured as:
*   $\phi = 0.4$ (Geographical anomaly factor weight)
*   $\psi = 0.4$ (Transaction amount anomaly factor weight)
*   $\omega = 0.2$ (Behavioral category factor weight)

### Rule Set
*   **Geographic Risk ($V_{geo}$):** $1.0$ if the transaction country differs from the customer's home country, $0.0$ otherwise.
*   **Value Risk ($D_{val}$):** $0.0$ if amount $\le$ typical amount. Otherwise, calculated as:
    $$\min\left(1.0, \frac{\text{amount} - \text{typical\_amount}}{\text{typical\_amount}}\right)$$
*   **Behavioral Risk ($I_{beh}$):** $1.0$ if merchant type is high-risk (e.g. `casino`, `betting`, `gambling`), $0.0$ otherwise.
*   **Decision:** Card suspension is executed automatically by the agent if $A_{risk} \ge 0.6$.

---

## 📂 Project Structure

```
.
├── db/
│   └── mock_db.go      # Thread-safe database mock with customer profiles and suspension tool.
├── risk/
│   └── calculator.go   # Mathematical multi-factor risk calculator implementation.
├── agent/
│   ├── agent.go        # Configuration of the official GenAI client & parameters definitions.
│   ├── loop.go         # Multi-turn agent loop with robust FunctionCall ID matching.
│   └── agent_test.go   # Test suite for risk calculations, database, and live workflows.
├── go.mod              # Go module descriptor.
├── main.go             # Simulation run scenarios.
└── README.md           # Documentation (this file).
```

---

## 🚀 Getting Started

### 📋 Prerequisites
- Go installed (version 1.20+ recommended)
- A valid Gemini API Key from Google AI Studio.

### 🔑 Set your API Key
Before running the tests or simulation, export your Gemini API Key in the environment:
```bash
export GEMINI_API_KEY="your-actual-api-key-here"
```

*Optional:* If you want to use a specific model, export `GEMINI_MODEL`. By default, the agent targets `gemini-2.5-flash`:
```bash
export GEMINI_MODEL="gemini-2.5-flash"
```

### 🧪 Running Unit Tests
You can verify the correctness of the risk calculator, thread-safe database mutations, tool declarations, and agent loop execution by running:
```bash
go test -v ./...
```
*Note: If `GEMINI_API_KEY` is not present, live integration tests will be skipped automatically.*

### 🖥️ Running the Scenario Simulation
The main program demonstrates the agent in action on two scenarios:
*   **Scenario A (Suspicious):** Withdrawal of 7200 PLN in a casino in Peru on account `ACC-7711`. (Should trigger automatic card blocking as $A_{risk} = 1.0$).
*   **Scenario B (Safe):** Local payment of 50 PLN in Poland on account `ACC-7711`. (Should be approved automatically as $A_{risk} = 0.0$).

Run the simulation:
```bash
go run main.go
```
