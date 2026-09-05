package contracts

import "testing"

func TestAllCanonicalSchemasCompileOffline(t *testing.T) {
	once.Do(compileSchemas)
	if compileErr != nil {
		t.Fatal(compileErr)
	}
	entries, err := schemas.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if compiled[entry.Name()] == nil {
			t.Fatal(entry.Name())
		}
	}
}

func TestSchemaAssertions(t *testing.T) {
	valid := `{"decision_id":"d","question":"q","evidence_ids":["e"],"state":"GET_MORE_DATA","reason":"r","supported_facts":[],"assumptions":[],"unknowns":[],"missing_evidence":[],"next_measurement":"","action":null}`
	if err := ValidateRaw("decision-packet.schema.json", []byte(valid)); err != nil {
		t.Fatal(err)
	}
	for _, raw := range []string{
		`{"effect_kind":"HUMAN_ACTION","effect_id":"a","extra":1}`,
		`{"effect_kind":"HUMAN_ACTION","effect_id":""}`,
		`{"effect_kind":"BOGUS","effect_id":"a"}`,
		`{"effect_kind":"HUMAN_ACTION"}`,
	} {
		if ValidateRaw("effect-ref.schema.json", []byte(raw)) == nil {
			t.Fatal(raw)
		}
	}
	raw := `{"outcome_id":"o","effect_ref":{"effect_kind":"HUMAN_ACTION","effect_id":"a"},"observed_at":"invalid","status":"PENDING","metrics":{},"source_ref":"fixture"}`
	if ValidateRaw("outcome-record.schema.json", []byte(raw)) == nil {
		t.Fatal("date-time format not asserted")
	}
	if ValidateRaw("not-embedded.schema.json", []byte(`{}`)) == nil {
		t.Fatal("unknown schema must fail offline")
	}
}

func TestDecodeRejectsAmbiguousJSON(t *testing.T) {
	for _, raw := range []string{`{"a":1,"a":2}`, `{"x":[{"a":1,"a":2}]}`, `{} {}`, `{"a":`, `[1,]`} {
		if _, err := Decode([]byte(raw)); err == nil {
			t.Fatal(raw)
		}
	}
	var out struct {
		ID string `json:"id"`
	}
	if err := DecodeStrict([]byte(`{"ID":"a"}`), &out); err == nil {
		t.Fatal("case-insensitive alias accepted")
	}
}
