package main

import (
	"strconv"
	"testing"
	"time"
)

func TestMissionSecondsBounds(t *testing.T) {
	for _, n := range []int{-1, 0, 60} {
		d, ok := checkedMissionSeconds(n)
		if ok != (n >= 0) || ok && d != time.Duration(n)*time.Second {
			t.Fatal(n, d, ok)
		}
	}
	if strconv.IntSize < 64 {
		return
	}
	max := int64((1<<63 - 1) / time.Second)
	for _, n := range []int64{max, max + 1, 1<<63 - 1} {
		_, ok := checkedMissionSeconds(int(n))
		if ok != (n == max) {
			t.Fatal(n, ok)
		}
	}
}

func TestM11HealthExactExpiry(t *testing.T) {
	for _, observed := range []string{"2026-09-03T07:55:00.000000001Z", "2026-09-03T07:55:00Z", "2026-09-03T07:54:59.999999999Z"} {
		s, c := baseM11()
		s.Health.ObservedAt = observed
		refreshHealthTrust(&s, &c)
		g := EvaluateProductionGate(s, c)
		_, _, status := AuthorizeProduction(s, c)
		if observed == "2026-09-03T07:54:59.999999999Z" {
			if g.Reason != "HEALTH_STALE" || status != "DEGRADE" {
				t.Fatal(observed, g, status)
			}
		} else if g.Decision != "ALLOW_PRODUCTION" {
			t.Fatal(observed, g)
		}
		if observed == "2026-09-03T07:55:00Z" && status != "DEGRADE" {
			t.Fatal("authorization must expire at equality", status)
		}
	}
}

func TestMissionWindowTimeBoundaries(t *testing.T) {
	s10, _ := baseM10()
	s11, _ := baseM11()
	start, _ := time.Parse(time.RFC3339, "2026-09-03T08:00:00.5Z")
	s10.Ledger.WindowStartedAt = start.Format(time.RFC3339Nano)
	s11.Ledger.WindowStartedAt = start.Format(time.RFC3339Nano)
	s10.Ledger.ExecutionsInWindow = 1
	s11.Ledger.ExecutionsInWindow = 1
	s10.Grant.WindowSeconds = 60
	s11.Lease.WindowSeconds = 60
	for _, delta := range []time.Duration{60*time.Second - 1, 60 * time.Second, 60*time.Second + 1} {
		a, ok := normalizedCanaryLedger(s10.Ledger, s10.Grant, start.Add(delta))
		b, good := normalizedProductionLedger(s11.Ledger, s11.Lease, start.Add(delta))
		want := 1
		if delta >= 60*time.Second {
			want = 0
		}
		if !ok || !good || a.ExecutionsInWindow != want || b.ExecutionsInWindow != want {
			t.Fatal(delta, a, b)
		}
	}
	if strconv.IntSize < 64 {
		return
	}
	max := int64((1<<63 - 1) / time.Second)
	for _, n := range []int64{max, max + 1, 1<<63 - 1} {
		s10.Grant.WindowSeconds = int(n)
		s11.Lease.WindowSeconds = int(n)
		a, ok := normalizedCanaryLedger(s10.Ledger, s10.Grant, start.Add(time.Hour))
		b, good := normalizedProductionLedger(s11.Ledger, s11.Lease, start.Add(time.Hour))
		if ok != (n == max) || good != (n == max) || a.ExecutionsInWindow != 1 || b.ExecutionsInWindow != 1 {
			t.Fatal(n, a, b, ok, good)
		}
	}
}

func TestM11HealthDurationOverflow(t *testing.T) {
	if strconv.IntSize < 64 {
		t.Skip("large int limits require 64 bits")
	}
	max := int64((1<<63 - 1) / time.Second)
	for _, n := range []int64{max, max + 1, 1<<63 - 1} {
		s, c := baseM11()
		s.Lease.MaxHealthSnapshotAgeSeconds = int(n)
		refreshLeaseTrust(&s, &c)
		a, g, status := AuthorizeProduction(s, c)
		if n == max {
			if status != "AUTHORIZED" {
				t.Fatal(n, g, status)
			}
		} else if status != "DENY" || a != (ProductionExecutionAuthorization{}) || g.Reason != "INVALID_HEALTH_AGE_LIMIT" {
			t.Fatal(n, g, status)
		}
	}
}

func TestMissionWindowOverflowDeniesAuthorization(t *testing.T) {
	if strconv.IntSize < 64 {
		t.Skip("requires 64-bit int")
	}
	n := int64(1<<63 - 1)
	s, c := baseM10()
	s.Grant.WindowSeconds = int(n)
	s.Grant = SealCanaryGrant(s.Grant)
	s.Ledger.GrantHash = s.Grant.GrantHash
	refreshGrantApproval(&c, s.Grant)
	a, g, status := AuthorizeCanary(s, c)
	if status != "DENY" || g.Reason != "LEDGER_MISMATCH" || a != (CanaryExecutionAuthorization{}) {
		t.Fatal(g, status)
	}
	p, pc := baseM11()
	p.Lease.WindowSeconds = int(n)
	refreshLeaseTrust(&p, &pc)
	pa, pg, status := AuthorizeProduction(p, pc)
	if status != "DENY" || pg.Reason != "LEDGER_MISMATCH" || pa != (ProductionExecutionAuthorization{}) {
		t.Fatal(pg, status)
	}
}
