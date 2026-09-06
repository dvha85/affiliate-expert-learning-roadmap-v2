package main

import (
	"encoding/json"
	"fmt"
)

// Diagnostics are local status, never a canonical gate or an authority grant.
type gateDiagnostic struct {
	Kind               string `json:"kind"`
	Decision           string `json:"decision"`
	Reason             string `json:"reason"`
	Validation         string `json:"validation"`
	ExecutionPermitted bool   `json:"execution_permitted"`
}

// exportGovernedResult snapshots validated bytes, not mutable caller objects.
// Invalid gates suppress the whole artifact bundle (including authorization).
func exportGovernedResult(mission string, input map[string]any) (map[string]any, error) {
	decode := DecodeM10Artifact
	if mission == "M11" {
		decode = DecodeM11Artifact
	} else if mission != "M10" {
		return nil, fmt.Errorf("unsupported governed mission")
	}
	gate, exists := input["gate"]
	if !exists {
		return nil, fmt.Errorf("missing gate")
	}
	b, e := json.Marshal(gate)
	if e != nil {
		return nil, e
	}
	_, status := decode("gate", b)
	if status != missionValid {
		var d struct {
			Decision string `json:"decision"`
			Reason   string `json:"reason"`
		}
		if e = json.Unmarshal(b, &d); e != nil {
			return nil, e
		}
		return map[string]any{"diagnostic": gateDiagnostic{"LOCAL_GATE_DIAGNOSTIC", d.Decision, d.Reason, status, false}}, nil
	}
	out := map[string]any{"gate": json.RawMessage(b)}
	for kind, value := range input {
		if kind == "gate" {
			continue
		}
		if kind == "authority" {
			want := "GOVERNED_CANARY"
			if mission == "M11" {
				want = "GOVERNED_PRODUCTION"
			}
			if value != want {
				return nil, fmt.Errorf("invalid authority label")
			}
			out[kind] = want
			continue
		}
		if kind != "authorization" && kind != "execution" && kind != "ledger" {
			return nil, fmt.Errorf("unexpected output %s", kind)
		}
		raw, err := json.Marshal(value)
		if err != nil {
			return nil, err
		}
		if _, s := decode(kind, raw); s != missionValid {
			return nil, fmt.Errorf("%s output %s: %s", mission, kind, s)
		}
		out[kind] = json.RawMessage(raw)
	}
	return out, nil
}

func marshalMissionDemo(id string, result any) ([]byte, error) {
	if id == "M10" || id == "M11" {
		input, ok := result.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("invalid governed result")
		}
		out, e := exportGovernedResult(id, input)
		if e != nil {
			return nil, e
		}
		result = out
	}
	return json.MarshalIndent(missionDemoEnvelope{id, result}, "", "  ")
}
