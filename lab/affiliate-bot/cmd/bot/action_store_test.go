package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/dvha85/affiliate-expert-learning-roadmap-v2/core/m03"
)

func TestActionStoreLifecycle(t *testing.T) {
	dir := t.TempDir()
	history := filepath.Join(dir, "history.jsonl")
	actions := filepath.Join(dir, "actions.jsonl")
	input := filepath.Join(dir, "action.json")
	obs, err := loadHistoryObservations("../../data/m02-sample-observations.json")
	if err != nil {
		t.Fatal(err)
	}
	record, err := NewHistoryRecord("decision-1", "2026-09-03T00:00:00Z", "2026-09-03T00:00:00Z", obs)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = AppendHistory(history, record); err != nil {
		t.Fatal(err)
	}
	originalHistory, _ := os.ReadFile(history)
	a := m03.HumanActionRecord{ActionID: "action-1", DecisionID: "decision-1", ActionType: "synthetic_manual_post", Target: "fixture:br10a", PerformedBy: "human", PerformedAt: "2026-09-04T00:00:00Z", MeasurementWindowEnd: "2026-09-05T00:00:00Z", ComplianceReviewed: true}
	write := func(a m03.HumanActionRecord) {
		raw, _ := json.Marshal(a)
		if err := os.WriteFile(input, raw, 0600); err != nil {
			t.Fatal(err)
		}
	}
	call := func(args []string, code int, status string) {
		t.Helper()
		var out, diagnostic bytes.Buffer
		if got := runActionStore(args, &out, &diagnostic); got != code {
			t.Fatalf("code %d: %s", got, diagnostic.String())
		}
		var envelope map[string]any
		if err := json.Unmarshal(out.Bytes(), &envelope); err != nil {
			t.Fatal(err)
		}
		if envelope["status"] != status || envelope["execution_permitted"] != false {
			t.Fatal(envelope)
		}
		if code != 0 && envelope["artifact"] != nil {
			t.Fatal("artifact on error")
		}
	}
	args := []string{"record", history, actions, input}
	missing := a
	missing.DecisionID = "absent"
	write(missing)
	call(args, 1, "DECISION_ERROR")
	if _, err := os.Stat(actions); !os.IsNotExist(err) {
		t.Fatal("created on rejection")
	}
	write(a)
	beforeInput, _ := os.ReadFile(input)
	call(args, 0, "APPENDED")
	saved, _ := os.ReadFile(actions)
	call(args, 0, "EXACT_DUPLICATE")
	call([]string{"list", history, actions}, 0, "VALID")
	for _, tc := range []struct {
		change func(*m03.HumanActionRecord)
		status string
	}{
		{func(a *m03.HumanActionRecord) { a.Target = "other" }, "CONFLICT"},
		{func(a *m03.HumanActionRecord) { a.DecisionID = "absent" }, "DECISION_ERROR"},
		{func(a *m03.HumanActionRecord) { a.PerformedBy = "machine" }, "INVALID_SCHEMA"},
		{func(a *m03.HumanActionRecord) { a.ComplianceReviewed = false }, "HUMAN_REVIEW"},
		{func(a *m03.HumanActionRecord) { a.MeasurementWindowEnd = "2020-01-01T00:00:00Z" }, "INVALID"},
	} {
		b := a
		tc.change(&b)
		write(b)
		call(args, 1, tc.status)
	}
	write(a)
	afterInput, _ := os.ReadFile(input)
	if !bytes.Equal(beforeInput, afterInput) {
		t.Fatal("input changed")
	}
	for _, pair := range [][2]string{{history, history}, {actions, actions}} {
		call([]string{"record", pair[0], pair[1], input}, 1, "PATH_ERROR")
	}
	alias := filepath.Join(dir, "alias")
	if err := os.Link(history, alias); err != nil {
		t.Fatal(err)
	}
	call([]string{"record", history, alias, input}, 1, "PATH_ERROR")
	now, _ := os.ReadFile(actions)
	h, _ := os.ReadFile(history)
	if !bytes.Equal(saved, now) || !bytes.Equal(originalHistory, h) {
		t.Fatal("existing files changed")
	}
	if err := os.WriteFile(actions, []byte("corrupt\n"), 0600); err != nil {
		t.Fatal(err)
	}
	call(args, 1, "STORE_ERROR")
	now, _ = os.ReadFile(actions)
	if string(now) != "corrupt\n" {
		t.Fatal("corrupt store overwritten")
	}
	if err := os.WriteFile(actions, saved[:len(saved)-1], 0600); err != nil {
		t.Fatal(err)
	}
	call(args, 1, "STORE_ERROR")
	now, _ = os.ReadFile(actions)
	if !bytes.Equal(now, saved[:len(saved)-1]) {
		t.Fatal("unterminated action store was changed")
	}
	call([]string{"record"}, 2, "USAGE_ERROR")
}

// M03 action records arrive via a caller-owned pathname. A same-byte external
// replacement after descriptor open must fail before the action store appends
// an artifact linked to canonical history.
func TestActionStoreRejectsPortableSymlinkSwapAfterOpen(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink permissions are not portable on Windows")
	}
	dir := t.TempDir()
	history := filepath.Join(dir, "history.jsonl")
	actions := filepath.Join(dir, "actions.jsonl")
	input := filepath.Join(dir, "action.json")
	record, err := NewHistoryRecord("action-portable-swap", "2026-09-03T00:00:00Z", "2026-09-03T00:00:00Z", []Observation{historyObservation("action-portable-swap-o1", "product", "Product", 100, .1, "2026-09-03T00:00:00Z")})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := AppendHistory(history, record); err != nil {
		t.Fatal(err)
	}
	action := m03.HumanActionRecord{ActionID: "action-portable-swap", DecisionID: record.RecordID, ActionType: "synthetic_manual_post", Target: "fixture:portable", PerformedBy: "human", PerformedAt: "2026-09-04T00:00:00Z", MeasurementWindowEnd: "2026-09-05T00:00:00Z", ComplianceReviewed: true}
	inputBytes, err := json.Marshal(action)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(input, inputBytes, 0600); err != nil {
		t.Fatal(err)
	}
	external := filepath.Join(t.TempDir(), "same-byte-external-action.json")
	if err := os.WriteFile(external, inputBytes, 0600); err != nil {
		t.Fatal(err)
	}
	swapped := false
	stableRegularFileReadHook = func(openedPath string) error {
		if filepath.Clean(openedPath) != filepath.Clean(input) || swapped {
			return nil
		}
		swapped = true
		if err := os.Remove(input); err != nil {
			return err
		}
		return os.Symlink(external, input)
	}
	t.Cleanup(func() { stableRegularFileReadHook = nil })
	var out, diagnostic bytes.Buffer
	if code := runActionStore([]string{"record", history, actions, input}, &out, &diagnostic); code == 0 {
		t.Fatalf("action store accepted a post-open input swap: response=%s diagnostic=%s", out.String(), diagnostic.String())
	}
	var envelope map[string]any
	if err := json.Unmarshal(out.Bytes(), &envelope); err != nil || envelope["status"] != "IO_ERROR" {
		t.Fatalf("action store did not reject portable swap as input error: response=%s err=%v diagnostic=%s", out.String(), err, diagnostic.String())
	}
	if !swapped {
		t.Fatal("action store did not use the shared portable-input reader")
	}
	if _, err := os.Stat(actions); !os.IsNotExist(err) {
		t.Fatalf("action store persisted after portable input swap: %v", err)
	}
	if got, err := os.ReadFile(external); err != nil || !bytes.Equal(got, inputBytes) {
		t.Fatalf("swap rejection changed external action bytes: %q err=%v", got, err)
	}
}

func TestActionDecisionRequiresUniqueReplay(t *testing.T) {
	obs, err := loadHistoryObservations("../../data/m02-sample-observations.json")
	if err != nil {
		t.Fatal(err)
	}
	r, err := NewHistoryRecord("d", "2026-09-03T00:00:00Z", "2026-09-03T00:00:00Z", obs)
	if err != nil {
		t.Fatal(err)
	}
	if err := resolveActionDecision([]HistoryRecord{r}, "d"); err != nil {
		t.Fatal(err)
	}
	if err := resolveActionDecision([]HistoryRecord{r, r}, "d"); err == nil {
		t.Fatal("ambiguous accepted")
	}
	r.RecordedResult.State = "HUMAN_REVIEW"
	if err := resolveActionDecision([]HistoryRecord{r}, "d"); err == nil {
		t.Fatal("drift accepted")
	}
}
