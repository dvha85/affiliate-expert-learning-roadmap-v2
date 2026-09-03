package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"reflect"
	"sort"
	"strings"
	"time"
)

const (
	appendAdded        = "APPENDED"
	appendDuplicate    = "EXACT_DUPLICATE"
	replayMatch        = "MATCH"
	replayDrift        = "DRIFT"
	replayUnreplayable = "UNREPLAYABLE"
)

type HistoryRecord struct {
	RecordID       string        `json:"record_id"`
	AsOf           string        `json:"as_of"`
	IngestedAt     string        `json:"ingested_at"`
	FormulaVersion string        `json:"formula_version"`
	InputHash      string        `json:"input_hash"`
	Observations   []Observation `json:"observations"`
	RecordedResult Result        `json:"recorded_result"`
}

type ReplayReport struct {
	RecordID string `json:"record_id"`
	State    string `json:"state"`
	Reason   string `json:"reason"`
}

func cloneObservation(in Observation) Observation {
	out := in
	if in.Price != nil {
		v := *in.Price
		out.Price = &v
	}
	if in.CommissionRate != nil {
		v := *in.CommissionRate
		out.CommissionRate = &v
	}
	return out
}

func canonicalObservations(in []Observation) []Observation {
	out := make([]Observation, len(in))
	for i, observation := range in {
		out[i] = cloneObservation(observation)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].ProductID != out[j].ProductID {
			return out[i].ProductID < out[j].ProductID
		}
		if out[i].ObservationID != out[j].ObservationID {
			return out[i].ObservationID < out[j].ObservationID
		}
		if out[i].ObservedAt != out[j].ObservedAt {
			return out[i].ObservedAt < out[j].ObservedAt
		}
		return out[i].ProductName < out[j].ProductName
	})
	return out
}

func inputHash(observations []Observation) (string, error) {
	encoded, err := json.Marshal(canonicalObservations(observations))
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(encoded)
	return hex.EncodeToString(sum[:]), nil
}

func NewHistoryRecord(recordID, asOf, ingestedAt string, observations []Observation) (HistoryRecord, error) {
	if strings.TrimSpace(recordID) == "" {
		return HistoryRecord{}, errors.New("record_id is required")
	}
	asOfTime, err := time.Parse(time.RFC3339, asOf)
	if err != nil {
		return HistoryRecord{}, fmt.Errorf("invalid as_of: %w", err)
	}
	if _, err := time.Parse(time.RFC3339, ingestedAt); err != nil {
		return HistoryRecord{}, fmt.Errorf("invalid ingested_at: %w", err)
	}
	seenObservationIDs := map[string]bool{}
	for _, observation := range observations {
		if strings.TrimSpace(observation.ObservationID) == "" {
			return HistoryRecord{}, errors.New("observation_id is required for M02 history")
		}
		observedTime, err := time.Parse(time.RFC3339, observation.ObservedAt)
		if err != nil {
			return HistoryRecord{}, fmt.Errorf("invalid observed_at for %s: %w", observation.ObservationID, err)
		}
		if observedTime.After(asOfTime) {
			return HistoryRecord{}, fmt.Errorf("as_of cannot be before observed_at for %s", observation.ObservationID)
		}
		if seenObservationIDs[observation.ObservationID] {
			return HistoryRecord{}, fmt.Errorf("duplicate observation_id %s in one history record", observation.ObservationID)
		}
		seenObservationIDs[observation.ObservationID] = true
	}

	canonical := canonicalObservations(observations)
	hash, err := inputHash(canonical)
	if err != nil {
		return HistoryRecord{}, err
	}
	decision := evaluate(canonical)
	return HistoryRecord{
		RecordID:       recordID,
		AsOf:           asOf,
		IngestedAt:     ingestedAt,
		FormulaVersion: decision.FormulaVersion,
		InputHash:      hash,
		Observations:   canonical,
		RecordedResult: decision,
	}, nil
}

func validateHistoryRecord(record HistoryRecord) error {
	if strings.TrimSpace(record.RecordID) == "" {
		return errors.New("record_id is required")
	}
	asOfTime, err := time.Parse(time.RFC3339, record.AsOf)
	if err != nil {
		return fmt.Errorf("invalid as_of: %w", err)
	}
	if _, err := time.Parse(time.RFC3339, record.IngestedAt); err != nil {
		return fmt.Errorf("invalid ingested_at: %w", err)
	}
	if strings.TrimSpace(record.FormulaVersion) == "" {
		return errors.New("formula_version is required")
	}
	if record.RecordedResult.FormulaVersion != record.FormulaVersion {
		return errors.New("record formula_version must match recorded decision formula_version")
	}
	seenObservationIDs := map[string]bool{}
	for _, observation := range record.Observations {
		if strings.TrimSpace(observation.ObservationID) == "" {
			return errors.New("observation_id is required for M02 history")
		}
		observedTime, err := time.Parse(time.RFC3339, observation.ObservedAt)
		if err != nil {
			return fmt.Errorf("invalid observed_at for %s: %w", observation.ObservationID, err)
		}
		if observedTime.After(asOfTime) {
			return fmt.Errorf("as_of cannot be before observed_at for %s", observation.ObservationID)
		}
		if seenObservationIDs[observation.ObservationID] {
			return fmt.Errorf("duplicate observation_id %s in one history record", observation.ObservationID)
		}
		seenObservationIDs[observation.ObservationID] = true
	}
	hash, err := inputHash(record.Observations)
	if err != nil {
		return err
	}
	if hash != record.InputHash {
		return fmt.Errorf("input_hash mismatch for %s", record.RecordID)
	}
	return nil
}

func LoadHistory(path string) ([]HistoryRecord, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var records []HistoryRecord
	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		raw := strings.TrimSpace(scanner.Text())
		if raw == "" {
			return nil, fmt.Errorf("history line %d is empty", lineNumber)
		}
		var record HistoryRecord
		if err := json.Unmarshal([]byte(raw), &record); err != nil {
			return nil, fmt.Errorf("history line %d corrupt: %w", lineNumber, err)
		}
		if err := validateHistoryRecord(record); err != nil {
			return nil, fmt.Errorf("history line %d invalid: %w", lineNumber, err)
		}
		records = append(records, record)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	sort.Slice(records, func(i, j int) bool {
		left, _ := time.Parse(time.RFC3339, records[i].AsOf)
		right, _ := time.Parse(time.RFC3339, records[j].AsOf)
		if !left.Equal(right) {
			return left.Before(right)
		}
		return records[i].RecordID < records[j].RecordID
	})
	return records, nil
}

func AppendHistory(path string, record HistoryRecord) (string, error) {
	if err := validateHistoryRecord(record); err != nil {
		return "", err
	}
	existing, err := LoadHistory(path)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return "", err
	}
	for _, old := range existing {
		if old.RecordID == record.RecordID {
			if reflect.DeepEqual(old, record) {
				return appendDuplicate, nil
			}
			return "", fmt.Errorf("CONFLICT: record_id %s already exists with different content", record.RecordID)
		}
		for _, previousObservation := range old.Observations {
			for _, candidateObservation := range record.Observations {
				if previousObservation.ObservationID == candidateObservation.ObservationID && !reflect.DeepEqual(previousObservation, candidateObservation) {
					return "", fmt.Errorf("CONFLICT: observation_id %s reused with different content", candidateObservation.ObservationID)
				}
			}
		}
	}

	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return "", err
	}
	defer file.Close()
	encoded, err := json.Marshal(record)
	if err != nil {
		return "", err
	}
	if _, err := file.Write(append(encoded, '\n')); err != nil {
		return "", err
	}
	return appendAdded, nil
}

func Replay(record HistoryRecord) ReplayReport {
	if err := validateHistoryRecord(record); err != nil {
		return ReplayReport{RecordID: record.RecordID, State: "INTEGRITY_ERROR", Reason: err.Error()}
	}
	if record.FormulaVersion != formulaVersion {
		return ReplayReport{RecordID: record.RecordID, State: replayUnreplayable, Reason: "formula_version không được runtime hiện tại hỗ trợ"}
	}
	actual := evaluate(record.Observations)
	if reflect.DeepEqual(actual, record.RecordedResult) {
		return ReplayReport{RecordID: record.RecordID, State: replayMatch, Reason: "same input + same formula_version tái tạo cùng decision"}
	}
	return ReplayReport{RecordID: record.RecordID, State: replayDrift, Reason: "replayed decision khác recorded decision"}
}
