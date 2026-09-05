package contracts

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"testing"
	"time"
)

// This is an example verifier, not a production importer or authorization gate.
type manualFixture struct {
	Synthetic  bool            `json:"synthetic"`
	Version    string          `json:"format_version"`
	Limitation string          `json:"limitation"`
	Action     json.RawMessage `json:"action"`
	Reports    []struct {
		ID         string  `json:"snapshot_id"`
		Source     string  `json:"source_ref"`
		Action     string  `json:"action_id"`
		Start      string  `json:"period_start"`
		End        string  `json:"period_end"`
		Observed   string  `json:"observed_at"`
		Status     string  `json:"status"`
		Clicks     *int    `json:"clicks"`
		Orders     *int    `json:"valid_orders"`
		Paid       *int    `json:"commission_paid_vnd"`
		Supersedes *string `json:"supersedes"`
	} `json:"reports"`
	Outcomes []json.RawMessage `json:"outcomes"`
}

func verifyManualFixture(raw []byte) error {
	var f manualFixture
	if err := DecodeStrict(raw, &f); err != nil {
		return err
	}
	if !f.Synthetic || f.Version != "br06-intermediate-v1" || f.Limitation == "" || len(f.Reports) != 3 || len(f.Outcomes) != 3 {
		return fmt.Errorf("fixture label/version/count mismatch")
	}
	if err := ValidateRaw("action-record.schema.json", f.Action); err != nil {
		return err
	}
	var action struct {
		ID       string `json:"action_id"`
		Start    string `json:"performed_at"`
		End      string `json:"measurement_window_end"`
		Reviewed bool   `json:"compliance_reviewed"`
	}
	if err := json.Unmarshal(f.Action, &action); err != nil {
		return err
	}
	start, _ := time.Parse(time.RFC3339, action.Start)
	end, _ := time.Parse(time.RFC3339, action.End)
	if action.Reviewed || !end.After(start) {
		return fmt.Errorf("fixture must not claim review; window must be positive")
	}
	seen := map[string]bool{}
	for i, report := range f.Reports {
		if report.ID == "" || seen[report.ID] || report.Action != action.ID || report.Start != action.Start || report.End != action.End {
			return fmt.Errorf("snapshot identity/window mismatch")
		}
		if (i == 0 && report.Supersedes != nil) || (i > 0 && (report.Supersedes == nil || *report.Supersedes != f.Reports[i-1].ID)) {
			return fmt.Errorf("snapshot update chain mismatch")
		}
		seen[report.ID] = true
		observed, err := time.Parse(time.RFC3339, report.Observed)
		if err != nil || observed.Before(start) || (report.Status == "NO_OBSERVED_OUTCOME" && observed.Before(end)) {
			return fmt.Errorf("observation outside allowed window")
		}
		if err := ValidateRaw("outcome-record.schema.json", f.Outcomes[i]); err != nil {
			return err
		}
		metrics := map[string]int{}
		for key, value := range map[string]*int{"clicks": report.Clicks, "valid_orders": report.Orders, "commission_paid_vnd": report.Paid} {
			if value != nil {
				if *value < 0 {
					return fmt.Errorf("negative metric")
				}
				metrics[key] = *value
			}
		}
		expected, err := json.Marshal(map[string]any{
			"outcome_id":  fmt.Sprintf("syn-out-%02d", i+1),
			"effect_ref":  map[string]string{"effect_kind": "HUMAN_ACTION", "effect_id": action.ID},
			"observed_at": report.Observed, "status": report.Status,
			"metrics": metrics, "source_ref": report.Source,
		})
		if err != nil {
			return err
		}
		want, err := Decode(expected)
		if err != nil {
			return err
		}
		got, err := Decode(f.Outcomes[i])
		if err != nil {
			return err
		}
		if !reflect.DeepEqual(got, want) {
			return fmt.Errorf("report %s: mapping mismatch (including missing versus zero)", report.ID)
		}
	}
	return nil
}

func TestManualAffiliateFixture(t *testing.T) {
	raw, err := os.ReadFile("../examples/affiliate-manual/manual-loop.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := verifyManualFixture(raw); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"missing_to_zero", "early_zero", "orphan_action", "bad_update", "unknown_field", "negative", "duplicate_id"} {
		t.Run(name, func(t *testing.T) {
			var value map[string]any
			if err := json.Unmarshal(raw, &value); err != nil {
				t.Fatal(err)
			}
			reports := value["reports"].([]any)
			outcomes := value["outcomes"].([]any)
			switch name {
			case "missing_to_zero":
				outcomes[0].(map[string]any)["metrics"].(map[string]any)["valid_orders"] = 0
			case "early_zero":
				reports[1].(map[string]any)["observed_at"] = "2026-09-03T12:00:00+07:00"
				outcomes[1].(map[string]any)["observed_at"] = "2026-09-03T12:00:00+07:00"
			case "orphan_action":
				reports[0].(map[string]any)["action_id"] = "missing"
			case "bad_update":
				reports[2].(map[string]any)["supersedes"] = "missing"
			case "unknown_field":
				outcomes[0].(map[string]any)["extra"] = true
			case "negative":
				reports[0].(map[string]any)["clicks"] = -1
			case "duplicate_id":
				reports[1].(map[string]any)["snapshot_id"] = "syn-report-01"
			}
			mutated, err := json.Marshal(value)
			if err != nil {
				t.Fatal(err)
			}
			if verifyManualFixture(mutated) == nil {
				t.Fatal("invalid mutation accepted")
			}
		})
	}
}
