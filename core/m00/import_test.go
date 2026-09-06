package m00

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func fixture(t *testing.T) Packet {
	t.Helper()
	b, e := os.ReadFile("../../examples/m00-import/packet-t2.json")
	if e != nil {
		t.Fatal(e)
	}
	var p Packet
	if e = json.Unmarshal(b, &p); e != nil {
		t.Fatal(e)
	}
	return p
}

func TestNumericProjectionDoesNotRewriteEvidence(t *testing.T) {
	p := fixture(t)
	raw, _ := json.Marshal(p)
	for _, value := range []string{"1e-1000", "1e-999999999", "9007199254740993", "100.00000000000000001"} {
		changed := strings.Replace(string(raw), `"value":100`, `"value":`+value, 1)
		if _, err := Convert([]byte(changed)); err == nil {
			t.Fatal("rounded source accepted", value)
		}
	}
}
func TestMissingAndProvenance(t *testing.T) {
	for _, state := range []string{"observed", "missing", "pending", "not_yet_observable", "inconclusive"} {
		t.Run(state, func(t *testing.T) {
			p := fixture(t)
			f := &p.Products[0].Fields[1]
			f.State = state
			zero := 0.0
			f.Value = &zero
			raw, _ := json.Marshal(p)
			out, e := Convert(raw)
			if e != nil {
				t.Fatal(e)
			}
			if state == "observed" {
				if out[0]["commission_rate"] != float64(0) {
					t.Fatal(out)
				}
			} else if out[0]["commission_rate"] != nil {
				t.Fatal("uncertain became numeric", out)
			}
			var provenance struct {
				Product Product `json:"product"`
			}
			if e = json.Unmarshal([]byte(out[0]["transformation_or_method"].(string)), &provenance); e != nil {
				t.Fatal(e)
			}
			if provenance.Product.Fields[0].State != state || *provenance.Product.Fields[0].Value != 0 {
				t.Fatal("provenance lost")
			}
		})
	}
	p := fixture(t)
	p.Products[0].Fields[1].ClaimKind = "unknown"
	raw, _ := json.Marshal(p)
	out, e := Convert(raw)
	if e != nil || out[0]["commission_rate"] != nil {
		t.Fatal(out, e)
	}
}
func TestRejectAmbiguousInputs(t *testing.T) {
	for name, mutate := range map[string]func(*Packet){
		"duplicate field":        func(p *Packet) { p.Products[0].Fields[1].Field = "price" },
		"duplicate id":           func(p *Packet) { p.Products[0].Fields[1].ObservationID = p.Products[0].Fields[0].ObservationID },
		"aggregate id collision": func(p *Packet) { p.Products[0].Fields[0].ObservationID = p.Products[0].ObservationID },
		"mismatched subject":     func(p *Packet) { p.Products[0].Fields[0].SubjectID = "other" },
		"absent field":           func(p *Packet) { p.Products[0].Fields = p.Products[0].Fields[:1] },
		"mixed origin": func(p *Packet) {
			p.Products[0].Fields[0].EvidenceKind = "real"
			p.Products[0].Fields[0].SourceURL = "https://example.org/source"
		},
		"known null":   func(p *Packet) { p.Products[0].Fields[0].Value = nil },
		"bad number":   func(p *Packet) { v := 1.1; p.Products[0].Fields[1].Value = &v },
		"bad time":     func(p *Packet) { p.Products[0].Fields[0].ObservedAt = "yesterday" },
		"blank source": func(p *Packet) { p.Products[0].Fields[0].SourceRef = " " },
	} {
		t.Run(name, func(t *testing.T) {
			p := fixture(t)
			mutate(&p)
			raw, _ := json.Marshal(p)
			if out, e := Convert(raw); e == nil || out != nil {
				t.Fatal(out, e)
			}
		})
	}
	for _, raw := range []string{`{"version":"m00-input/v1","version":"m00-input/v1"}`, `{"Version":"m00-input/v1"}`, `null`, `{} {}`, `{"unexpected":true}`} {
		if _, e := Convert([]byte(raw)); e == nil {
			t.Fatal(raw)
		}
	}
}
