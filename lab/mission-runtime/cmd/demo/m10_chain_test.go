package main

import (
	"bytes"
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"testing"
)

func chainRaw(t *testing.T, mutate func(*M10State, *M10Context)) [][]byte {
	t.Helper()
	s, c := baseM10()
	if mutate != nil {
		mutate(&s, &c)
	}
	a, gate, status := AuthorizeCanary(s, c)
	if status != "AUTHORIZED" {
		t.Fatal(status, gate)
	}
	ap := CanaryGrantApproval{s.Grant.ApprovalRef, s.Grant.GrantID, s.Grant.GrantVersion, s.Grant.GrantHash, "APPROVE", "human", s.Grant.ApproverID, s.Grant.ApprovedAt}
	r := canaryExecutionRecord(s, a, c.Now, "CANCELLED", "NOT_PERFORMED", "", "")
	raw := [][]byte{}
	for _, v := range []any{s.Intent, s.Policy, s.Grant, ap, s.CostBound, s.Ledger, gate, a, r} {
		b, e := json.Marshal(v)
		if e != nil {
			t.Fatal(e)
		}
		raw = append(raw, b)
	}
	return raw
}

func changeChain(t *testing.T, raw [][]byte, index int, key string, value any) {
	t.Helper()
	var fields map[string]json.RawMessage
	if e := json.Unmarshal(raw[index], &fields); e != nil {
		t.Fatal(e)
	}
	b, e := json.Marshal(value)
	if e != nil {
		t.Fatal(e)
	}
	fields[key] = b
	raw[index], e = json.Marshal(fields)
	if e != nil {
		t.Fatal(e)
	}
}

func TestM10ChainBindings(t *testing.T) {
	for _, tc := range []struct {
		index int
		key   string
		value any
		want  string
	}{
		{0, "target", "https://example.com/changed", "TAMPERED_INTENT"},
		{1, "intent_id", "wrong", "BROKEN_LINK"}, {1, "policy_version", "wrong", "BROKEN_LINK"},
		{3, "approval_ref", "wrong", "BROKEN_LINK"}, {3, "grant_id", "wrong", "BROKEN_LINK"}, {3, "grant_version", "wrong", "BROKEN_LINK"}, {3, "approver_id", "wrong", "BROKEN_LINK"},
		{5, "grant_id", "wrong", "BROKEN_LINK"}, {6, "cost_bound_id", "wrong", "BROKEN_LINK"}, {6, "risk_class", "RISK1", "BROKEN_LINK"},
		{7, "canary_gate_id", "wrong", "BROKEN_LINK"}, {7, "idempotency_key", "wrong", "BROKEN_LINK"}, {7, "correlation_id", "wrong", "BROKEN_LINK"},
		{8, "authorization_id", "wrong", "BROKEN_LINK"}, {8, "canary_cost_bound_minor", 101, "BROKEN_LINK"}, {8, "executor_id", "wrong", "BROKEN_LINK"},
		{6, "decision", "DENY", "INVALID_GATE_STATE"}, {6, "reason", "wrong", "INVALID_GATE_STATE"},
		{3, "approved_at", "2026-09-03T07:06:00Z", "INVALID_TIME_BINDING"},
		{7, "authorized_at", "2026-09-03T08:01:00Z", "INVALID_TIME_BINDING"},
		{7, "expires_at", "2026-09-03T09:30:00Z", "INVALID_TIME_BINDING"},
		{8, "attempted_at", "2026-09-03T07:59:59Z", "INVALID_TIME_BINDING"},
		{5, "updated_at", "2026-09-03T08:01:00Z", "INVALID_TIME_BINDING"},
		{5, "reconciliation_required", true, "LEDGER_BLOCKED"},
		{6, "executions_total_before", 1, "LEDGER_SNAPSHOT_MISMATCH"},
		{7, "execution_mode", "GOVERNED_PRODUCTION", "INVALID_SCHEMA"}, {8, "approval_id", "m09", "INVALID_SCHEMA"},
	} {
		t.Run(tc.key, func(t *testing.T) {
			raw := chainRaw(t, nil)
			changeChain(t, raw, tc.index, tc.key, tc.value)
			s, state := CheckM10Chain(raw)
			if state != tc.want || s != (M10ChainSummary{}) {
				t.Fatal(state, tc.want, s)
			}
		})
	}
	// Each input must pass its own raw boundary; zero values cannot hide absent files.
	for n := 0; n < 9; n++ {
		raw := chainRaw(t, nil)
		raw[n] = []byte(`{}`)
		if _, s := CheckM10Chain(raw); s != "INVALID_SCHEMA" {
			t.Fatal(n, s)
		}
	}
}

func TestM10ChainScopeAndTime(t *testing.T) {
	for _, tc := range []struct {
		name   string
		mutate func(*M10State, *M10Context)
	}{
		{"risk0", nil},
		{"risk1", func(s *M10State, c *M10Context) {
			refreshM10Intent(s, c, "UPDATE_DRAFT", "https://example.com/draft", "r1", "r1-key", 100)
		}},
		{"window_reset", func(s *M10State, c *M10Context) {
			s.Ledger.ExecutionsTotal = 2
			s.Ledger.ExecutionsInWindow = 2
			c.Now = "2026-09-03T08:10:00Z"
		}},
		{"large_window", func(s *M10State, c *M10Context) {
			s.Grant.WindowSeconds = math.MaxInt64
			s.Grant = SealCanaryGrant(s.Grant)
			s.Ledger.GrantHash = s.Grant.GrantHash
			refreshGrantApproval(c, s.Grant)
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			raw := chainRaw(t, tc.mutate)
			s, state := CheckM10Chain(raw)
			if state != missionValid || s.Result != "CONSISTENT_UNVERIFIED" || s.ProvenanceAuthenticated || s.ExecutionPermitted {
				t.Fatal(state, s)
			}
		})
	}
	for _, field := range []string{"allowed_hosts", "allowed_action_types", "executor_ids", "allowed_risk_classes"} {
		raw := chainRaw(t, nil)
		var g CanaryGrant
		json.Unmarshal(raw[2], &g)
		switch field {
		case "allowed_hosts":
			g.AllowedHosts = []string{"other.example"}
		case "allowed_action_types":
			g.AllowedActionTypes = []string{"OTHER"}
		case "executor_ids":
			g.ExecutorIDs = []string{"other"}
		case "allowed_risk_classes":
			g.AllowedRiskClasses = []string{"RISK1"}
		}
		g = SealCanaryGrant(g)
		raw[2], _ = json.Marshal(g)
		for _, idx := range []int{3, 5, 6} {
			changeChain(t, raw, idx, "grant_hash", g.GrantHash)
		}
		for _, idx := range []int{7, 8} {
			changeChain(t, raw, idx, "canary_grant_hash", g.GrantHash)
		}
		if _, state := CheckM10Chain(raw); state != "SCOPE_NOT_DELEGATED" {
			t.Fatal(field, state)
		}
	}
	for _, tc := range []struct{ at, want string }{{"2026-09-03T08:59:59Z", missionValid}, {"2026-09-03T15:59:59+07:00", missionValid}, {"2026-09-03T09:00:00Z", "EXPIRED_AUTHORIZATION"}} {
		raw := chainRaw(t, nil)
		changeChain(t, raw, 8, "status", "SUCCEEDED")
		changeChain(t, raw, 8, "side_effect_state", "PERFORMED")
		changeChain(t, raw, 8, "attempted_at", tc.at)
		if _, s := CheckM10Chain(raw); s != tc.want {
			t.Fatal(s, tc.want)
		}
	}
}

func TestM10ChainBudget(t *testing.T) {
	// Both operands fit int64 but their sum would overflow. Audit uses subtraction.
	overflow := chainRaw(t, nil)
	var grant CanaryGrant
	var cost CanaryCostBound
	json.Unmarshal(overflow[2], &grant)
	json.Unmarshal(overflow[4], &cost)
	grant.MaxCostMinorTotal = math.MaxInt64
	grant = SealCanaryGrant(grant)
	overflow[2], _ = json.Marshal(grant)
	cost.MaxCostMinor = math.MaxInt64
	cost = SealCanaryCostBound(cost)
	overflow[4], _ = json.Marshal(cost)
	for _, idx := range []int{3, 5, 6} {
		changeChain(t, overflow, idx, "grant_hash", grant.GrantHash)
	}
	for _, idx := range []int{7, 8} {
		changeChain(t, overflow, idx, "canary_grant_hash", grant.GrantHash)
		changeChain(t, overflow, idx, "canary_cost_bound_hash", cost.CostBoundHash)
		changeChain(t, overflow, idx, "canary_cost_bound_minor", cost.MaxCostMinor)
	}
	changeChain(t, overflow, 6, "cost_bound_hash", cost.CostBoundHash)
	changeChain(t, overflow, 6, "cost_bound_minor", cost.MaxCostMinor)
	changeChain(t, overflow, 5, "cost_minor_total", 1)
	changeChain(t, overflow, 6, "cost_minor_total_before", 1)
	if _, s := CheckM10Chain(overflow); s != "BUDGET_EXCEEDED" {
		t.Fatal("overflow", s)
	}
	for _, tc := range []struct {
		key, gateKey string
		value        any
		want         string
	}{
		{"executions_total", "executions_total_before", 3, "BUDGET_EXCEEDED"},
		{"cost_minor_total", "cost_minor_total_before", int64(math.MaxInt64), "BUDGET_EXCEEDED"},
		{"cost_minor_total", "cost_minor_total_before", 901, "BUDGET_EXCEEDED"},
		{"cost_minor_total", "cost_minor_total_before", 900, missionValid},
	} {
		raw := chainRaw(t, nil)
		changeChain(t, raw, 5, tc.key, tc.value)
		changeChain(t, raw, 6, tc.gateKey, tc.value)
		if _, s := CheckM10Chain(raw); s != tc.want {
			t.Fatal(tc, s)
		}
	}
	raw := chainRaw(t, nil)
	changeChain(t, raw, 5, "executions_total", 2)
	changeChain(t, raw, 6, "executions_total_before", 2)
	changeChain(t, raw, 5, "executions_in_window", 2)
	changeChain(t, raw, 6, "executions_in_window_before", 2)
	if _, s := CheckM10Chain(raw); s != "BUDGET_EXCEEDED" {
		t.Fatal(s)
	}
	raw = chainRaw(t, nil)
	changeChain(t, raw, 5, "executions_total", 1)
	changeChain(t, raw, 6, "executions_total_before", 1)
	changeChain(t, raw, 5, "successful_idempotency_keys", []string{"k"})
	if _, s := CheckM10Chain(raw); s != "LEDGER_BLOCKED" {
		t.Fatal(s)
	}
}

func TestM10ChainCLI(t *testing.T) {
	raw := chainRaw(t, nil)
	dir := t.TempDir()
	paths := []string{}
	for n, name := range []string{"intent", "policy", "grant", "approval", "cost", "ledger", "gate", "authorization", "execution"} {
		p := filepath.Join(dir, name+".json")
		if e := os.WriteFile(p, raw[n], 0600); e != nil {
			t.Fatal(e)
		}
		paths = append(paths, p)
	}
	var out bytes.Buffer
	for n := 0; n < 2; n++ {
		out.Reset()
		if e := runM10ChainCheck(&out, paths); e != nil {
			t.Fatal(e)
		}
		var s M10ChainSummary
		if e := json.Unmarshal(out.Bytes(), &s); e != nil {
			t.Fatal(e)
		}
		if s.Result != "CONSISTENT_UNVERIFIED" || s.ProvenanceAuthenticated || s.ExecutionPermitted {
			t.Fatal(s)
		}
	}
	if e := runM10ChainCheck(m03BrokenWriter{}, paths); e == nil {
		t.Fatal("writer")
	}
	for n, p := range paths {
		b, e := os.ReadFile(p)
		if e != nil || !bytes.Equal(b, raw[n]) {
			t.Fatal("changed file")
		}
	}
	for _, args := range [][]string{nil, paths[:8], append([]string{"missing"}, paths[1:]...)} {
		out.Reset()
		if e := runM10ChainCheck(&out, args); e == nil || out.Len() != 0 {
			t.Fatal("bad args")
		}
	}
	os.WriteFile(paths[8], []byte(`{}`), 0600)
	out.Reset()
	if e := runM10ChainCheck(&out, paths); e == nil || out.Len() != 0 {
		t.Fatal("bad input")
	}
	entries, e := os.ReadDir(dir)
	if e != nil || len(entries) != 9 {
		t.Fatal("persisted state")
	}
}
