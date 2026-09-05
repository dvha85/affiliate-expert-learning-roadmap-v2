package main

import (
	"fmt"
	"os"

	"github.com/dvha85/affiliate-expert-learning-roadmap-v2/contracts"
)

// Legacy Result encodes a nil ranking as null. Validate an equivalent empty
// array view, without modifying persisted bytes or the replay representation.
func validateHistoryJSON(raw []byte) error {
	value, err := contracts.Decode(raw)
	if err != nil {
		return err
	}
	if object, ok := value.(map[string]any); ok {
		if result, ok := object["recorded_result"].(map[string]any); ok {
			if ranked, exists := result["ranked"]; exists && ranked == nil {
				result["ranked"] = []any{}
			}
		}
	}
	if err := contracts.Validate("history-record.schema.json", value); err != nil {
		return err
	}
	// The Go projection cannot preserve arbitrary extension fields. Reject them
	// rather than silently dropping them and hashing/replaying a different record.
	var record HistoryRecord
	return contracts.DecodeStrict(raw, &record)
}

func loadHistoryObservations(path string) ([]Observation, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if err := contracts.ValidateRaw("history-record.schema.json#/properties/observations", raw); err != nil {
		return nil, fmt.Errorf("invalid M02 observations: %w", err)
	}
	var observations []Observation
	if err := contracts.DecodeStrict(raw, &observations); err != nil {
		return nil, err
	}
	return observations, nil
}
