package db

import (
	"errors"
	"fmt"
	"sync"
)

// CustomerProfile contains the customer details used for fraud detection.
type CustomerProfile struct {
	AccountID     string  `json:"account_id"`
	CustomerName  string  `json:"customer_name"`
	HomeCountry   string  `json:"home_country"`
	TypicalAmount float64 `json:"typical_amount"`
	CardStatus    string  `json:"card_status"` // "ACTIVE", "SUSPENDED"
}

// MockDB is a thread-safe simulation of a customer database.
type MockDB struct {
	mu        sync.RWMutex
	customers map[string]*CustomerProfile
}

// NewMockDB initializes and seeds the database with test profiles.
func NewMockDB() *MockDB {
	db := &MockDB{
		customers: make(map[string]*CustomerProfile),
	}
	db.seed()
	return db
}

// seed populates initial data for testing.
func (db *MockDB) seed() {
	db.customers["ACC-7711"] = &CustomerProfile{
		AccountID:     "ACC-7711",
		CustomerName:  "Jan Kowalski",
		HomeCountry:   "PL",
		TypicalAmount: 100.0,
		CardStatus:    "ACTIVE",
	}
	db.customers["ACC-8822"] = &CustomerProfile{
		AccountID:     "ACC-8822",
		CustomerName:  "John Smith",
		HomeCountry:   "US",
		TypicalAmount: 50.0,
		CardStatus:    "ACTIVE",
	}
}

// GetCustomerProfile retrieves a copy of the customer profile by ID.
func (db *MockDB) GetCustomerProfile(accountID string) (*CustomerProfile, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	profile, exists := db.customers[accountID]
	if !exists {
		return nil, fmt.Errorf("customer profile not found for account %s", accountID)
	}

	// Return a deep copy to ensure thread safety when modifying
	return &CustomerProfile{
		AccountID:     profile.AccountID,
		CustomerName:  profile.CustomerName,
		HomeCountry:   profile.HomeCountry,
		TypicalAmount: profile.TypicalAmount,
		CardStatus:    profile.CardStatus,
	}, nil
}

// SuspendCard suspends the card associated with the account.
func (db *MockDB) SuspendCard(accountID string) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	profile, exists := db.customers[accountID]
	if !exists {
		return fmt.Errorf("customer profile not found for account %s", accountID)
	}

	if profile.CardStatus == "SUSPENDED" {
		return errors.New("card is already suspended")
	}

	profile.CardStatus = "SUSPENDED"
	return nil
}
