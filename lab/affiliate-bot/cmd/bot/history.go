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
	"time"
)

const (
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

func canonicalObservations(records []Observation) []Observation {
	out := append([]Observation(nil), records...)
	sort.Slice(out, func(i, j int) bool {
		if out[i].ProductID != out[j].ProductID {
			return out[i].ProductID < out[j].ProductID
		}
		return out[i].ProductName < out[j].ProductName
	})
	return out
}

func hashObservations(records []Observation) (string, error) {
	raw, err := json.Marshal(canonicalObservations(records))
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:]), nil
}

func NewHistoryRecord(recordID, asOf, ingestedAt string, records []Observation) (HistoryRecord, error) {
	if recordID == "" {
		return HistoryRecord{}, errors.New("record_id is required")
	}
	if _, err := time.Parse(time.RFC3339, asOf); err != nil {
		return HistoryRecord{}, fmt.Errorf("invalid as_of: %w", err)
	}
	if _, err := time.Parse(time.RFC3339, ingestedAt); err != nil {
		return HistoryRecord{}, fmt.Errorf("invalid ingested_at: %w", err)
	}
	canonical := canonicalObservations(records)
	hash, err := hashObservations(canonical)
	if err != nil {
		return HistoryRecord{}, err
	}
	return HistoryRecord{RecordID: recordID, AsOf: asOf, IngestedAt: ingestedAt, FormulaVersion: formulaVersion, InputHash: hash, Observations: canonical, RecordedResult: evaluate(canonical)}, nil
}

func LoadHistory(path string) ([]HistoryRecord, error) {
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return []HistoryRecord{}, nil
		}
		return nil, err
	}
	defer file.Close()

	var records []HistoryRecord
	scanner := bufio.NewScanner(file)
	line := 0
	for scanner.Scan() {
		line++
		var record HistoryRecord
		if err := json.Unmarshal(scanner.Bytes(), &record); err != nil {
			return nil, fmt.Errorf("corrupt history line %d: %w", line, err)
		}
		records = append(records, record)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return records, nil
}

func AppendHistory(path string, candidate HistoryRecord) (string, error) {
	existing, err := LoadHistory(path)
	if err != nil {
		return "", err
	}
	for _, record := range existing {
		if record.RecordID != candidate.RecordID {
			continue
		}
		if reflect.DeepEqual(record, candidate) {
			return "EXACT_DUPLICATE", nil
		}
		return "CONFLICT", fmt.Errorf("record_id %s already exists with different content", candidate.RecordID)
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return "", err
	}
	defer file.Close()
	raw, err := json.Marshal(candidate)
	if err != nil {
		return "", err
	}
	if _, err := file.Write(append(raw, '\n')); err != nil {
		return "", err
	}
	return "APPENDED", nil
}

func QueryHistory(records []HistoryRecord) ([]HistoryRecord, error) {
	out := append([]HistoryRecord(nil), records...)
	for _, record := range out {
		if _, err := time.Parse(time.RFC3339, record.AsOf); err != nil {
			return nil, fmt.Errorf("record %s has invalid as_of: %w", record.RecordID, err)
		}
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].AsOf != out[j].AsOf {
			return out[i].AsOf < out[j].AsOf
		}
		return out[i].RecordID < out[j].RecordID
	})
	return out, nil
}

func Replay(record HistoryRecord) ReplayReport {
	if record.FormulaVersion != formulaVersion {
		return ReplayReport{RecordID: record.RecordID, State: replayUnreplayable, Reason: "formula_version is not registered"}
	}
	hash, err := hashObservations(record.Observations)
	if err != nil || hash != record.InputHash {
		return ReplayReport{RecordID: record.RecordID, State: "INTEGRITY_ERROR", Reason: "input hash mismatch"}
	}
	got := evaluate(record.Observations)
	if reflect.DeepEqual(got, record.RecordedResult) {
		return ReplayReport{RecordID: record.RecordID, State: replayMatch, Reason: "recorded result reproduced with the same formula version"}
	}
	return ReplayReport{RecordID: record.RecordID, State: replayDrift, Reason: "replayed result differs from recorded result"}
}
