package main

import (
	"bytes"
	"math"
	"os"
	"reflect"
	"testing"
)

func TestM10CostBudgetOverflow(t *testing.T) {
	for _, tc := range []struct {
		name               string
		spent, cost, limit int64
		allow              bool
	}{
		{"overflow_small_limit", 1, math.MaxInt64, 1000, false},
		{"overflow_max_limit", 1, math.MaxInt64, math.MaxInt64, false},
		{"overflow_both_max", math.MaxInt64, math.MaxInt64, math.MaxInt64, false},
		{"already_over_budget", 1001, 0, 1000, false},
		{"one_over", 901, 100, 1000, false},
		{"exact_limit", 900, 100, 1000, true},
		{"below_limit", 899, 100, 1000, true},
		{"zero_budget", 0, 0, 0, true},
		{"zero_budget_positive_cost", 0, 1, 0, false},
		{"exact_max", math.MaxInt64 - 1, 1, math.MaxInt64, true},
		{"max_with_zero", math.MaxInt64, 0, math.MaxInt64, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s, c := baseM10()
			s.Grant.MaxCostMinorTotal = tc.limit
			s.Grant = SealCanaryGrant(s.Grant)
			s.Ledger.GrantHash = s.Grant.GrantHash
			refreshGrantApproval(&c, s.Grant)
			s.CostBound.MaxCostMinor = tc.cost
			s.CostBound = SealCanaryCostBound(s.CostBound)
			c.TrustedCostBounds[s.CostBound.CostBoundID] = s.CostBound.CostBoundHash
			s.Ledger.CostMinorTotal = tc.spent
			before := s
			g := EvaluateCanaryGate(s, c)
			want, reason := "REQUIRE_APPROVAL", "CANARY_COST_BUDGET_EXHAUSTED"
			if tc.allow {
				want, reason = "ALLOW_CANARY", "CANARY_ELIGIBLE"
			}
			if g.Decision != want || g.Reason != reason || g.ExecutionAuthorized {
				t.Fatalf("got %s/%s; want %s/%s", g.Decision, g.Reason, want, reason)
			}
			a, _, status := AuthorizeCanary(s, c)
			if tc.allow {
				if status != "AUTHORIZED" {
					t.Fatal(status)
				}
			} else if status != "REQUIRE_APPROVAL" || a != (CanaryExecutionAuthorization{}) {
				t.Fatal("overbudget authorization", status, a)
			}
			if !reflect.DeepEqual(before, s) {
				t.Fatal("gate mutated state")
			}
		})
	}
}

func TestM10ExecutorRejectsOverflowAfterLedgerReload(t *testing.T) {
	s, c := baseM10()
	s.CostBound.MaxCostMinor = 1
	s.CostBound = SealCanaryCostBound(s.CostBound)
	c.TrustedCostBounds[s.CostBound.CostBoundID] = s.CostBound.CostBoundHash
	a, _, status := AuthorizeCanary(s, c)
	if status != "AUTHORIZED" {
		t.Fatal(status)
	}
	dir := t.TempDir()
	path := canaryLedgerPath(dir, s.Grant.GrantID, s.Grant.GrantVersion)
	stored := s.Ledger
	stored.CostMinorTotal = math.MaxInt64
	if e := persistCanaryLedger(path, stored); e != nil {
		t.Fatal(e)
	}
	before, e := os.ReadFile(path)
	if e != nil {
		t.Fatal(e)
	}
	r, status := ExecuteCanaryLocalSandbox(&s, a, c, dir)
	if status != "REQUIRE_APPROVAL" || r != (CanaryExecutionRecord{}) {
		t.Fatal(status, r)
	}
	after, e := os.ReadFile(path)
	if e != nil || !bytes.Equal(before, after) {
		t.Fatal("ledger changed on rejection")
	}
	entries, e := os.ReadDir(dir)
	if e != nil || len(entries) != 1 {
		t.Fatal("unexpected side effect or marker", entries, e)
	}
}
