// Package m10 owns canonical cost-bound validation shared by governed stages.
package m10

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strings"
	"time"

	"github.com/dvha85/affiliate-expert-learning-roadmap-v2/contracts"
)

type TrustedCostBound struct {
	CostBoundID   string `json:"cost_bound_id"`
	IntentID      string `json:"intent_id"`
	IntentHash    string `json:"intent_hash"`
	MaxCostMinor  int64  `json:"max_cost_minor"`
	Currency      string `json:"currency"`
	SourceRef     string `json:"source_ref"`
	ObservedAt    string `json:"observed_at"`
	ExpiresAt     string `json:"expires_at"`
	CorrelationID string `json:"correlation_id"`
	HashVersion   string `json:"hash_version"`
	CostBoundHash string `json:"cost_bound_hash"`
}

type costBoundHashPayload struct {
	CostBoundID   string `json:"cost_bound_id"`
	IntentID      string `json:"intent_id"`
	IntentHash    string `json:"intent_hash"`
	MaxCostMinor  int64  `json:"max_cost_minor"`
	Currency      string `json:"currency"`
	SourceRef     string `json:"source_ref"`
	ObservedAt    string `json:"observed_at"`
	ExpiresAt     string `json:"expires_at"`
	CorrelationID string `json:"correlation_id"`
	HashVersion   string `json:"hash_version"`
}

func ComputeTrustedCostBoundHash(c TrustedCostBound) string {
	raw, err := json.Marshal(costBoundHashPayload{c.CostBoundID, c.IntentID, c.IntentHash, c.MaxCostMinor, strings.ToUpper(strings.TrimSpace(c.Currency)), c.SourceRef, c.ObservedAt, c.ExpiresAt, c.CorrelationID, c.HashVersion})
	if err != nil {
		return ""
	}
	sum := sha256.Sum256(raw)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func DecodeTrustedCostBound(raw []byte) (TrustedCostBound, string) {
	var c TrustedCostBound
	if contracts.ValidateRaw("trusted-cost-bound.schema.json", raw) != nil || contracts.DecodeStrict(raw, &c) != nil {
		return c, "INVALID_SCHEMA"
	}
	if c.HashVersion != "go-json-v1" || c.CostBoundHash != ComputeTrustedCostBoundHash(c) {
		return c, "TAMPERED_COST_BOUND"
	}
	observed, observedErr := time.Parse(time.RFC3339, c.ObservedAt)
	expires, expiresErr := time.Parse(time.RFC3339, c.ExpiresAt)
	if observedErr != nil || expiresErr != nil || !expires.After(observed) {
		return c, "INVALID_TIME_BINDING"
	}
	return c, "VALID"
}

// ValidFor reserves only a currently valid bound that exactly matches the
// active proposal and grant currency. Registry membership is checked by the
// adapter, not inferred from a self-supplied hash.
func ValidFor(c TrustedCostBound, intentID, intentHash, correlationID, currency string, now time.Time) string {
	if c.IntentID != intentID || c.IntentHash != intentHash || c.CorrelationID != correlationID || c.Currency != strings.ToUpper(strings.TrimSpace(currency)) || strings.TrimSpace(c.SourceRef) == "" {
		return "COST_BOUND_MISMATCH"
	}
	expires, err := time.Parse(time.RFC3339, c.ExpiresAt)
	if err != nil || !expires.After(now) {
		return "COST_BOUND_EXPIRED"
	}
	return "VALID"
}
