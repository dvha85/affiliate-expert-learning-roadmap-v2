package main

import "testing"

func f(v float64) *float64 { return &v }

func TestRealEvidenceNeverAutoRecommends(t *testing.T) {
    records := []Observation{
        {ProductID: "a", ProductName: "A", Price: f(100), CommissionRate: f(0.10), Currency: "USD", EvidenceKind: "real"},
        {ProductID: "b", ProductName: "B", Price: f(120), CommissionRate: f(0.08), Currency: "USD", EvidenceKind: "real"},
    }
    got := evaluate(records)
    if got.State != "RANK_SCENARIO" {
        t.Fatalf("expected RANK_SCENARIO, got %s", got.State)
    }
}

func TestMissingEvidenceGetsMoreData(t *testing.T) {
    records := []Observation{{ProductID: "a", ProductName: "A", Price: nil, CommissionRate: f(0.10), Currency: "USD", EvidenceKind: "synthetic"}}
    got := evaluate(records)
    if got.State != "GET_MORE_DATA" {
        t.Fatalf("expected GET_MORE_DATA, got %s", got.State)
    }
}

func TestMixedEvidenceNeedsHumanReview(t *testing.T) {
    records := []Observation{
        {ProductID: "a", ProductName: "A", Price: f(100), CommissionRate: f(0.10), Currency: "USD", EvidenceKind: "real"},
        {ProductID: "b", ProductName: "B", Price: f(100), CommissionRate: f(0.10), Currency: "USD", EvidenceKind: "synthetic"},
    }
    got := evaluate(records)
    if got.State != "HUMAN_REVIEW" {
        t.Fatalf("expected HUMAN_REVIEW, got %s", got.State)
    }
}

func TestStableTieBreak(t *testing.T) {
    records := []Observation{
        {ProductID: "b", ProductName: "B", Price: f(100), CommissionRate: f(0.10), Currency: "USD", EvidenceKind: "synthetic"},
        {ProductID: "a", ProductName: "A", Price: f(100), CommissionRate: f(0.10), Currency: "USD", EvidenceKind: "synthetic"},
    }
    got := evaluate(records)
    if len(got.Ranked) != 2 || got.Ranked[0].ProductName != "A" {
        t.Fatalf("expected deterministic A-first tie break, got %#v", got.Ranked)
    }
}
