package m04

import "testing"

func TestAdvisorBoundaryAndReferences(t *testing.T) {
	const valid = `{"state":"ADVISE","reason":"inspect evidence","evidence_ids":["e"],"unknowns":[],"write_tool_requested":false}`
	ev := []AdvisorEvidence{{EvidenceID: "e", ObservedAt: "2026-09-01T00:00:00Z", SourceRef: "fixture:test"}}
	o, status := DecodeAdvisorOutput([]byte(valid))
	if status != "VALID" {
		t.Fatal(status)
	}
	for _, tc := range []struct {
		asOf     string
		maxAge   int
		expected string
	}{
		{"2026-09-01T01:00:00Z", 1, "SUPPORTED"},
		{"2026-09-01T01:00:01Z", 1, "ABSTAIN_STALE"},
		{"2026-08-31T23:59:59Z", 1, "ABSTAIN_FUTURE"},
	} {
		if got := EvaluateAdvisorOutput(o, ev, tc.asOf, tc.maxAge); got != tc.expected {
			t.Fatal(got, tc)
		}
	}
	o.EvidenceIDs = []string{"missing"}
	if got := EvaluateAdvisorOutput(o, ev, "2026-09-01T01:00:00Z", 1); got != "REJECT_UNGROUNDED" {
		t.Fatal(got)
	}
	for _, tc := range []struct{ raw, expected string }{
		{`{"state":"ADVISE","reason":"r","evidence_ids":["e"],"unknowns":[],"write_tool_requested":true}`, "REJECT_WRITE_REQUEST"},
		{`{"state":"ADVISE","reason":" ","evidence_ids":["e"],"unknowns":[],"write_tool_requested":false}`, "INVALID_SCHEMA"},
		{`{"state":"ABSTAIN","state":"ADVISE"}`, "INVALID_SCHEMA"},
		{`{"state":"ADVISE","reason":"r","evidence_ids":["e"],"unknowns":null,"write_tool_requested":false}`, "INVALID_SCHEMA"},
	} {
		if _, got := DecodeAdvisorOutput([]byte(tc.raw)); got != tc.expected {
			t.Fatal(got, tc.expected)
		}
	}
}
