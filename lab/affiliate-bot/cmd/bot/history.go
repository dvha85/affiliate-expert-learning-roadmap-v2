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

	"github.com/dvha85/affiliate-expert-learning-roadmap-v2/contracts"
	"github.com/dvha85/affiliate-expert-learning-roadmap-v2/core/m00"
	"github.com/dvha85/affiliate-expert-learning-roadmap-v2/lab/affiliate-bot/internal/store"
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

func canonicalEvidenceIDs(observations []Observation) []string {
	canonical := canonicalObservations(observations)
	ids := make([]string, 0, len(canonical))
	for _, observation := range canonical {
		ids = append(ids, observation.ObservationID)
	}
	return ids
}

func inputHash(observations []Observation) (string, error) {
	encoded, err := json.Marshal(canonicalObservations(observations))
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(encoded)
	return hex.EncodeToString(sum[:]), nil
}

func validateCanonicalHistoryObservation(observation Observation, asOfTime time.Time) error {
	raw, err := json.Marshal(observation)
	if err != nil {
		return err
	}
	if _, err := m00.SourceFields(raw); err != nil {
		return err
	}
	if strings.TrimSpace(observation.ObservationID) == "" {
		return errors.New("observation_id is required for M02 history")
	}
	if strings.TrimSpace(observation.SubjectID) == "" {
		return fmt.Errorf("subject_id is required for %s", observation.ObservationID)
	}
	if strings.TrimSpace(observation.ProductID) == "" {
		return fmt.Errorf("product_id is required for %s", observation.ObservationID)
	}
	if observation.SubjectID != observation.ProductID {
		return fmt.Errorf("subject_id must equal product_id in M02 learner runtime for %s", observation.ObservationID)
	}
	if strings.TrimSpace(observation.SourceURL) == "" && strings.TrimSpace(observation.SourceRef) == "" {
		return fmt.Errorf("source_url or source_ref is required for %s", observation.ObservationID)
	}
	if strings.TrimSpace(observation.AccessMethod) == "" {
		return fmt.Errorf("access_method is required for %s", observation.ObservationID)
	}
	switch observation.ClaimKind {
	case "fact", "estimate", "assumption", "unknown":
	default:
		return fmt.Errorf("claim_kind is invalid for %s", observation.ObservationID)
	}
	switch observation.State {
	case "observed", "missing", "pending", "not_yet_observable", "inconclusive":
	default:
		return fmt.Errorf("state is invalid for %s", observation.ObservationID)
	}
	if strings.TrimSpace(observation.Limitation) == "" {
		return fmt.Errorf("limitation is required for %s", observation.ObservationID)
	}
	observedTime, err := time.Parse(time.RFC3339, observation.ObservedAt)
	if err != nil {
		return fmt.Errorf("invalid observed_at for %s: %w", observation.ObservationID, err)
	}
	if observedTime.After(asOfTime) {
		return fmt.Errorf("as_of cannot be before observed_at for %s", observation.ObservationID)
	}
	return nil
}

func NewHistoryRecord(recordID, asOf, ingestedAt string, observations []Observation) (HistoryRecord, error) {
	if err := validateImportIdentities(observations); err != nil {
		return HistoryRecord{}, err
	}
	raw, err := json.Marshal(observations)
	if err != nil {
		return HistoryRecord{}, err
	}
	if err := contracts.ValidateRaw("history-record.schema.json#/properties/observations", raw); err != nil {
		return HistoryRecord{}, err
	}
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
		if err := validateCanonicalHistoryObservation(observation, asOfTime); err != nil {
			return HistoryRecord{}, err
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
	if decision.Ranked == nil {
		decision.Ranked = []Ranked{}
	}
	decision.DecisionID = recordID
	decision.EvidenceIDs = canonicalEvidenceIDs(canonical)
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
	if err := validateImportIdentities(record.Observations); err != nil {
		return err
	}
	raw, err := json.Marshal(record)
	if err != nil {
		return err
	}
	if err := validateHistoryJSON(raw); err != nil {
		return err
	}
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
	if record.RecordedResult.DecisionID != record.RecordID {
		return errors.New("recorded decision_id must equal record_id in M02 learner runtime")
	}
	seenObservationIDs := map[string]bool{}
	for _, observation := range record.Observations {
		if err := validateCanonicalHistoryObservation(observation, asOfTime); err != nil {
			return err
		}
		if seenObservationIDs[observation.ObservationID] {
			return fmt.Errorf("duplicate observation_id %s in one history record", observation.ObservationID)
		}
		seenObservationIDs[observation.ObservationID] = true
	}
	expectedEvidenceIDs := canonicalEvidenceIDs(record.Observations)
	if !reflect.DeepEqual(record.RecordedResult.EvidenceIDs, expectedEvidenceIDs) {
		return errors.New("recorded evidence_ids must exactly match canonical observation_ids")
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

// Raw field IDs and derived observation IDs share one namespace per snapshot.
func validateImportIdentities(observations []Observation) error {
	seen := map[string]bool{}
	for _, observation := range observations {
		if seen[observation.ObservationID] {
			return fmt.Errorf("duplicate observation_id %s", observation.ObservationID)
		}
		seen[observation.ObservationID] = true
	}
	for _, observation := range observations {
		raw, err := json.Marshal(observation)
		if err != nil {
			return err
		}
		fields, err := m00.SourceFields(raw)
		if err != nil {
			return err
		}
		for _, field := range fields {
			if seen[field.ObservationID] {
				return fmt.Errorf("duplicate source observation_id %s", field.ObservationID)
			}
			seen[field.ObservationID] = true
		}
	}
	return nil
}

func LoadHistory(path string) ([]HistoryRecord, error) {
	return loadHistoryWith(store.JSONL{}, path)
}

func loadHistoryWith(storage store.History, path string) ([]HistoryRecord, error) {
	file, err := storage.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	// Allow the full payload plus CRLF framing, with an explicit payload check.
	scanner.Buffer(make([]byte, 4096), store.MaxHistoryRecordBytes+2)
	var records []HistoryRecord
	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		if len(scanner.Bytes()) > store.MaxHistoryRecordBytes {
			return nil, fmt.Errorf("history line %d exceeds %d bytes", lineNumber, store.MaxHistoryRecordBytes)
		}
		raw := strings.TrimSpace(scanner.Text())
		if raw == "" {
			return nil, fmt.Errorf("history line %d is empty", lineNumber)
		}
		var record HistoryRecord
		if err := validateHistoryJSON([]byte(raw)); err != nil {
			return nil, fmt.Errorf("history line %d corrupt: %w", lineNumber, err)
		}
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
	release, err := acquireHistoryRuntimeGate(path)
	if err != nil {
		return "", err
	}
	defer release()
	return appendHistoryWith(store.JSONL{}, path, record)
}

func appendHistoryWith(storage store.History, path string, record HistoryRecord) (string, error) {
	if err := validateHistoryRecord(record); err != nil {
		return "", err
	}
	encoded, err := json.Marshal(record)
	if err != nil {
		return "", err
	}
	if len(encoded) > store.MaxHistoryRecordBytes {
		return "", fmt.Errorf("history record exceeds %d bytes", store.MaxHistoryRecordBytes)
	}
	existing, err := loadHistoryWith(storage, path)
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
				previousRaw, _ := json.Marshal(previousObservation)
				candidateRaw, _ := json.Marshal(candidateObservation)
				previousFields, err := m00.SourceFields(previousRaw)
				if err != nil {
					return "", err
				}
				candidateFields, err := m00.SourceFields(candidateRaw)
				if err != nil {
					return "", err
				}
				for _, field := range previousFields {
					if field.ObservationID == candidateObservation.ObservationID {
						return "", fmt.Errorf("CONFLICT: source ID reused as aggregate ID")
					}
					for _, candidate := range candidateFields {
						if field.ObservationID == candidate.ObservationID && !reflect.DeepEqual(field, candidate) {
							return "", fmt.Errorf("CONFLICT: source observation_id %s reused with different content", field.ObservationID)
						}
					}
				}
				for _, field := range candidateFields {
					if field.ObservationID == previousObservation.ObservationID {
						return "", fmt.Errorf("CONFLICT: aggregate ID reused as source ID")
					}
				}
				if previousObservation.ObservationID == candidateObservation.ObservationID && !reflect.DeepEqual(previousObservation, candidateObservation) {
					return "", fmt.Errorf("CONFLICT: observation_id %s reused with different content", candidateObservation.ObservationID)
				}
			}
		}
	}

	if err := storage.AppendLine(path, encoded); err != nil {
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
	actual.DecisionID = record.RecordID
	actual.EvidenceIDs = canonicalEvidenceIDs(record.Observations)
	// Legacy projection encoded an empty ranking as null; preserve file bytes
	// while comparing its empty-array meaning with newly emitted records.
	recorded := record.RecordedResult
	if actual.Ranked == nil {
		actual.Ranked = []Ranked{}
	}
	if recorded.Ranked == nil {
		recorded.Ranked = []Ranked{}
	}
	if reflect.DeepEqual(actual, recorded) {
		return ReplayReport{RecordID: record.RecordID, State: replayMatch, Reason: "same input + same formula_version tái tạo cùng decision"}
	}
	return ReplayReport{RecordID: record.RecordID, State: replayDrift, Reason: "replayed decision khác recorded decision"}
}
