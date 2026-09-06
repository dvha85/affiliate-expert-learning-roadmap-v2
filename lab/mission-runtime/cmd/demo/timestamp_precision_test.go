package main

import (
	"encoding/json"
	"testing"
	"time"
)

func TestTimestampAuthorizationPrecision(t *testing.T) {
	for _, delta := range []time.Duration{-1, 0, 1} {
		t.Run(delta.String(), func(t *testing.T) {
			s, c := baseM10()
			now, _ := time.Parse(time.RFC3339, c.Now)
			end := now.Add(delta).Format(time.RFC3339Nano)
			s.CostBound.ExpiresAt = end
			s.CostBound = SealCanaryCostBound(s.CostBound)
			c.TrustedCostBounds[s.CostBound.CostBoundID] = s.CostBound.CostBoundHash
			a, _, status := AuthorizeCanary(s, c)
			if delta <= 0 {
				if status == "AUTHORIZED" {
					t.Fatal(status)
				}
			} else {
				if status != "AUTHORIZED" || a.ExpiresAt != end {
					t.Fatal(status, a.ExpiresAt, end)
				}
				b, _ := json.Marshal(a)
				v, valid := DecodeM10Artifact("authorization", b)
				if valid != missionValid {
					t.Fatal(valid)
				}
				a = *v.(*CanaryExecutionAuthorization)
				s.Authorization = &a
				dir := t.TempDir()
				if _, status = ExecuteCanaryLocalSandbox(&s, a, c, dir); status != "EXECUTED" {
					t.Fatal(status)
				}
				if _, err := loadExistingCanaryLedger(canaryLedgerPath(dir, s.Grant.GrantID, s.Grant.GrantVersion), s.Grant); err != nil {
					t.Fatal(err)
				}
			}
			p, pc := baseM11()
			p.Health.ObservedAt = now.Add(-300*time.Second + delta).Format(time.RFC3339Nano)
			refreshHealthTrust(&p, &pc)
			pa, _, ps := AuthorizeProduction(p, pc)
			if delta <= 0 {
				if ps == "AUTHORIZED" {
					t.Fatal(ps)
				}
				return
			}
			if ps != "AUTHORIZED" || pa.ExpiresAt != end {
				t.Fatal(ps, pa.ExpiresAt, end)
			}
			b, _ := json.Marshal(pa)
			v, valid := DecodeM11Artifact("authorization", b)
			if valid != missionValid {
				t.Fatal(valid)
			}
			pa = *v.(*ProductionExecutionAuthorization)
			p.Authorization = &pa
			dir := t.TempDir()
			initM11(t, &p, dir, pc.Now)
			if _, ps = ExecuteProductionLocalSandbox(&p, pa, pc, dir); ps != "EXECUTED" {
				t.Fatal(ps)
			}
			if _, err := loadProductionLedger(productionLedgerPath(dir, p.Lease.LeaseID, p.Lease.LeaseVersion), p.Lease); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestTimestampWindowRestartPrecision(t *testing.T) {
	s, _ := baseM10()
	p, _ := baseM11()
	s.Grant.WindowSeconds = 60
	p.Lease.WindowSeconds = 60
	start, _ := time.Parse(time.RFC3339, "2026-09-03T08:00:00.5Z")
	s.Ledger.WindowStartedAt = start.Format(time.RFC3339Nano)
	p.Ledger.WindowStartedAt = s.Ledger.WindowStartedAt
	for cycle := 0; cycle < 3; cycle++ {
		next := start.Add(time.Minute)
		var ok bool
		s.Ledger, ok = normalizedCanaryLedger(s.Ledger, s.Grant, next)
		if !ok {
			t.Fatal("M10 normalize")
		}
		p.Ledger, ok = normalizedProductionLedger(p.Ledger, p.Lease, next)
		if !ok {
			t.Fatal("M11 normalize")
		}
		if s.Ledger.WindowStartedAt != next.Format(time.RFC3339Nano) || p.Ledger.WindowStartedAt != next.Format(time.RFC3339Nano) {
			t.Fatal("lost fractional window")
		}
		s.Ledger.UpdatedAt = next.Format(time.RFC3339Nano)
		p.Ledger.UpdatedAt = s.Ledger.UpdatedAt
		s.Ledger.ExecutionsTotal = 1
		s.Ledger.ExecutionsInWindow = 1
		p.Ledger.ExecutionsTotal = 1
		p.Ledger.ExecutionsInWindow = 1
		dir := t.TempDir()
		cp := canaryLedgerPath(dir, s.Grant.GrantID, s.Grant.GrantVersion)
		pp := productionLedgerPath(dir, p.Lease.LeaseID, p.Lease.LeaseVersion)
		if err := persistCanaryLedger(cp, s.Ledger); err != nil {
			t.Fatal(err)
		}
		if err := persistProductionLedger(pp, p.Ledger); err != nil {
			t.Fatal(err)
		}
		var err error
		s.Ledger, err = loadExistingCanaryLedger(cp, s.Grant)
		if err != nil {
			t.Fatal(err)
		}
		p.Ledger, err = loadProductionLedger(pp, p.Lease)
		if err != nil {
			t.Fatal(err)
		}
		a, ok := normalizedCanaryLedger(s.Ledger, s.Grant, next.Add(time.Minute-1))
		b, good := normalizedProductionLedger(p.Ledger, p.Lease, next.Add(time.Minute-1))
		if !ok || !good || a.ExecutionsInWindow != 1 || b.ExecutionsInWindow != 1 {
			t.Fatal("early reset after restart")
		}
		start = next
	}
}
