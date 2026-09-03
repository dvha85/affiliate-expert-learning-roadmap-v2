package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"
)

func f(v float64) *float64 { return &v }

func TestRealEvidenceNeverAutoRecommends(t *testing.T) {
	records := []Observation{{ProductID: "a", ProductName: "A", Price: f(100), CommissionRate: f(0.10), Currency: "USD", EvidenceKind: "real"}}
	if got := evaluate(records); got.State != stateRank {
		t.Fatalf("expected %s, got %s", stateRank, got.State)
	}
}

func TestZeroIsNotMissing(t *testing.T) {
	records := []Observation{{ProductID: "a", ProductName: "A", Price: f(0), CommissionRate: f(0), Currency: "USD", EvidenceKind: "synthetic"}}
	if got := evaluate(records); got.State != stateRank {
		t.Fatalf("zero is observed data; expected %s, got %s", stateRank, got.State)
	}
}

func TestInvalidEvidenceGetsMoreData(t *testing.T) {
	records := []Observation{{ProductID: "a", ProductName: "A", Price: f(100), CommissionRate: f(0.1), Currency: "USD", EvidenceKind: "replay"}}
	if got := evaluate(records); got.State != stateGetMoreData {
		t.Fatalf("expected %s, got %s", stateGetMoreData, got.State)
	}
}

func TestConflictPrecedenceOverMissing(t *testing.T) {
	records := []Observation{
		{ProductID: "a", ProductName: "A", Price: f(100), CommissionRate: f(0.10), Currency: "USD", EvidenceKind: "real"},
		{ProductID: "a", ProductName: "A duplicate", Price: nil, CommissionRate: f(0.10), Currency: "USD", EvidenceKind: "real"},
	}
	if got := evaluate(records); got.State != stateHumanReview {
		t.Fatalf("identity conflict must dominate missing data; expected %s, got %s", stateHumanReview, got.State)
	}
}

func TestSameInputSameOutput(t *testing.T) {
	records := []Observation{
		{ProductID: "b", ProductName: "B", Price: f(100), CommissionRate: f(0.10), Currency: "USD", EvidenceKind: "synthetic"},
		{ProductID: "a", ProductName: "A", Price: f(100), CommissionRate: f(0.10), Currency: "USD", EvidenceKind: "synthetic"},
	}
	first, second := evaluate(records), evaluate(records)
	if !reflect.DeepEqual(first, second) {
		t.Fatalf("same input/version must produce same output: %#v != %#v", first, second)
	}
	if len(first.Ranked) != 2 || first.Ranked[0].ProductID != "a" {
		t.Fatalf("expected deterministic A-first tie break, got %#v", first.Ranked)
	}
}

type evalCase struct {
	CaseID                   string        `json:"case_id"`
	Observations             []Observation `json:"observations"`
	ExpectedState            string        `json:"expected_state"`
	ExpectedRankedProductIDs []string      `json:"expected_ranked_product_ids"`
}

func TestEvalPack(t *testing.T) {
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot resolve test file path")
	}
	path := filepath.Clean(filepath.Join(filepath.Dir(currentFile), "..", "..", "..", "..", "evals", "M01-deterministic-bot", "cases.json"))
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read eval pack: %v", err)
	}
	var cases []evalCase
	if err := json.Unmarshal(raw, &cases); err != nil {
		t.Fatalf("decode eval pack: %v", err)
	}
	if len(cases) < 8 {
		t.Fatalf("expected at least 8 eval cases, got %d", len(cases))
	}
	for _, tc := range cases {
		t.Run(tc.CaseID, func(t *testing.T) {
			got := evaluate(tc.Observations)
			if got.State != tc.ExpectedState {
				t.Fatalf("expected state %s, got %s", tc.ExpectedState, got.State)
			}
			ids := make([]string, 0, len(got.Ranked))
			for _, item := range got.Ranked {
				ids = append(ids, item.ProductID)
			}
			if !reflect.DeepEqual(ids, tc.ExpectedRankedProductIDs) {
				t.Fatalf("expected ranking %v, got %v", tc.ExpectedRankedProductIDs, ids)
			}
		})
	}
}
