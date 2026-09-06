package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func checkMatrixOutput(t *testing.T, mission, kind string, value any) {
	t.Helper()
	b, e := json.Marshal(value)
	if e != nil {
		t.Fatal(e)
	}
	var status string
	switch mission {
	case "09":
		if kind == "authorization" {
			_, status = DecodeM09Authorization(b)
		} else {
			_, status = DecodeM09Execution(b)
		}
	case "10":
		_, status = DecodeM10Artifact(kind, b)
	case "11":
		_, status = DecodeM11Artifact(kind, b)
	}
	if status != missionValid {
		t.Fatalf("M%s %s: %s: %s", mission, kind, status, b)
	}
}

// Local diagnostic returns are deliberately not canonical artifact evidence.
func TestOutputMatrixDiagnosticGateIsNotCanonical(t *testing.T) {
	for _, mode := range []string{"empty", "missing_cost"} {
		s, c := baseM10()
		p, pc := baseM11()
		if mode == "empty" {
			s = M10State{}
			p = M11State{}
		} else {
			s.CostBound = CanaryCostBound{}
			p.CostBound = CanaryCostBound{}
		}
		g := EvaluateCanaryGate(s, c)
		pg := EvaluateProductionGate(p, pc)
		b, _ := json.Marshal(g)
		if _, status := DecodeM10Artifact("gate", b); status == missionValid {
			t.Fatal("diagnostic claimed canonical", mode)
		}
		b, _ = json.Marshal(pg)
		if _, status := DecodeM11Artifact("gate", b); status == missionValid {
			t.Fatal("diagnostic claimed canonical", mode)
		}
	}
}

func TestOutputMatrixExecutors(t *testing.T) {
	for _, mode := range []string{"success", "unknown", "failure"} {
		t.Run(mode, func(t *testing.T) {
			dir := t.TempDir()
			s, c := baseM09()
			a, status := AuthorizeM09(s, c)
			if status != "AUTHORIZED" {
				t.Fatal(status)
			}
			s.Authorization = &a
			checkMatrixOutput(t, "09", "authorization", a)
			if mode == "unknown" {
				if e := os.WriteFile(sandboxIdempotencyPath(dir, a.IdempotencyKey), []byte("unknown"), 0600); e != nil {
					t.Fatal(e)
				}
			}
			if mode == "failure" {
				dir = filepath.Join(dir, "file")
				if e := os.WriteFile(dir, []byte("blocked"), 0600); e != nil {
					t.Fatal(e)
				}
			}
			r, status := ExecuteLocalSandbox(&s, a, c, dir)
			want := "EXECUTED"
			if mode == "unknown" {
				want = "WAIT_RECONCILIATION"
			}
			if mode == "failure" {
				want = "EXECUTION_FAILED"
			}
			if status != want {
				t.Fatal(status, want)
			}
			checkMatrixOutput(t, "09", "execution", r)
			if mode == "failure" {
				return
			} // M10/M11 pre-effect I/O failures return zero sentinels, not records.
			cs, cc := baseM10()
			ca, cg, status := AuthorizeCanary(cs, cc)
			if status != "AUTHORIZED" {
				t.Fatal(status)
			}
			cs.Authorization = &ca
			checkMatrixOutput(t, "10", "authorization", ca)
			checkMatrixOutput(t, "10", "gate", cg)
			cd := t.TempDir()
			if mode == "unknown" {
				if e := os.WriteFile(sandboxIdempotencyPath(cd, ca.IdempotencyKey), []byte("unknown"), 0600); e != nil {
					t.Fatal(e)
				}
			}
			cr, status := ExecuteCanaryLocalSandbox(&cs, ca, cc, cd)
			if status != want {
				t.Fatal(status, want)
			}
			checkMatrixOutput(t, "10", "execution", cr)
			ps, pc := baseM11()
			pd := t.TempDir()
			initM11(t, &ps, pd, pc.Now)
			pa, pg, status := AuthorizeProduction(ps, pc)
			if status != "AUTHORIZED" {
				t.Fatal(status)
			}
			ps.Authorization = &pa
			checkMatrixOutput(t, "11", "authorization", pa)
			checkMatrixOutput(t, "11", "gate", pg)
			if mode == "unknown" {
				if e := os.WriteFile(sandboxIdempotencyPath(pd, pa.IdempotencyKey), []byte("unknown"), 0600); e != nil {
					t.Fatal(e)
				}
				want = "STOP_RECONCILIATION"
			}
			pr, status := ExecuteProductionLocalSandbox(&ps, pa, pc, pd)
			if status != want {
				t.Fatal(status, want)
			}
			checkMatrixOutput(t, "11", "execution", pr)
		})
	}
}

func TestOutputMatrixGateDecisions(t *testing.T) {
	for _, mode := range []string{"allow", "deny", "wait", "approval", "stop", "degrade"} {
		t.Run(mode, func(t *testing.T) {
			s, c := baseM10()
			p, pc := baseM11()
			want10, want11 := "ALLOW_CANARY", "ALLOW_PRODUCTION"
			switch mode {
			case "deny":
				c.AllowedExecutorIDs = nil
				pc.AllowedExecutorIDs = nil
				want10 = "DENY"
				want11 = "DENY"
			case "wait":
				s.Ledger.ExecutionsTotal = 2
				s.Ledger.ExecutionsInWindow = s.Grant.MaxExecutionsPerWindow
				p.Ledger.ExecutionsTotal = 2
				p.Ledger.ExecutionsInWindow = p.Lease.MaxExecutionsPerWindow
				want10 = "WAIT"
				want11 = "WAIT"
			case "approval":
				s.Ledger.ExecutionsTotal = s.Grant.MaxExecutionsTotal
				p.Ledger.ExecutionsTotal = p.Lease.MaxExecutionsTotal
				want10 = "REQUIRE_APPROVAL"
				want11 = "REQUIRE_APPROVAL"
			case "stop":
				p.Ledger.ControlMode = "STOPPED"
				p.Ledger.StopReason = "test"
				want11 = "STOP"
			case "degrade":
				p.Health.TelemetryComplete = false
				refreshHealthTrust(&p, &pc)
				want11 = "DEGRADE"
			}
			a, g, status := AuthorizeCanary(s, c)
			if g.Decision != want10 {
				t.Fatal(mode, g)
			}
			checkMatrixOutput(t, "10", "gate", g)
			if status != "AUTHORIZED" && a != (CanaryExecutionAuthorization{}) {
				t.Fatal("nonzero denied auth")
			}
			pa, pg, ps := AuthorizeProduction(p, pc)
			if pg.Decision != want11 {
				t.Fatal(mode, pg)
			}
			checkMatrixOutput(t, "11", "gate", pg)
			if ps != "AUTHORIZED" && pa != (ProductionExecutionAuthorization{}) {
				t.Fatal("nonzero denied auth")
			}
		})
	}
}
