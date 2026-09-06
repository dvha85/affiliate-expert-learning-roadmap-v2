package main

import (
	"encoding/json"
	"testing"
)

func TestGovernedDiagnosticExport(t *testing.T) {
	for _, mode := range []string{"empty", "cost", "hash", "time", "persistence", "nil"} {
		t.Run(mode, func(t *testing.T) {
			s, c := baseM10()
			p, pc := baseM11()
			switch mode {
			case "empty":
				s = M10State{}
				p = M11State{}
			case "cost":
				s.CostBound = CanaryCostBound{}
				p.CostBound = CanaryCostBound{}
			case "hash":
				s.Grant.GrantHash = ""
				p.Lease.LeaseHash = ""
			case "time":
				c.Now = "invalid"
				pc.Now = "invalid"
			}
			g := EvaluateCanaryGate(s, c)
			pg := EvaluateProductionGate(p, pc)
			if mode == "persistence" {
				pg, _ = EnforceProductionGate(&p, pc, t.TempDir())
			}
			if mode == "nil" {
				pg, _ = EnforceProductionGate(nil, pc, t.TempDir())
			}
			inputs := map[string]any{"M11": pg}
			if mode != "persistence" && mode != "nil" {
				inputs["M10"] = g
			}
			for mission, gate := range inputs {
				b, e := marshalMissionDemo(mission, map[string]any{"gate": gate, "authorization": "must not leak", "authority": "must not leak"})
				if e != nil {
					t.Fatal(e)
				}
				var out struct {
					Result map[string]json.RawMessage `json:"result"`
				}
				if e = json.Unmarshal(b, &out); e != nil {
					t.Fatal(e)
				}
				if len(out.Result) != 1 || out.Result["diagnostic"] == nil {
					t.Fatal(string(b))
				}
				var d gateDiagnostic
				if e = json.Unmarshal(out.Result["diagnostic"], &d); e != nil {
					t.Fatal(e)
				}
				if d.Kind != "LOCAL_GATE_DIAGNOSTIC" || d.ExecutionPermitted || d.Validation == missionValid {
					t.Fatal(d)
				}
			}
		})
	}
}

func TestGovernedCanonicalExport(t *testing.T) {
	for _, mission := range []string{"M10", "M11"} {
		s, c := baseM10()
		a, g, _ := AuthorizeCanary(s, c)
		p, pc := baseM11()
		pa, pg, _ := AuthorizeProduction(p, pc)
		input := map[string]any{"gate": g, "authorization": a, "ledger": s.Ledger}
		decode := DecodeM10Artifact
		if mission == "M11" {
			input = map[string]any{"gate": pg, "authorization": pa, "ledger": p.Ledger}
			decode = DecodeM11Artifact
		}
		b, e := marshalMissionDemo(mission, input)
		if e != nil {
			t.Fatal(e)
		}
		var out struct {
			Result map[string]json.RawMessage `json:"result"`
		}
		json.Unmarshal(b, &out)
		for kind, raw := range out.Result {
			if _, status := decode(kind, raw); status != missionValid {
				t.Fatal(kind, status)
			}
		}
		input["authorization"] = nil
		if b, e = marshalMissionDemo(mission, input); e == nil || len(b) != 0 {
			t.Fatal("invalid sibling exported")
		}
	}
}
