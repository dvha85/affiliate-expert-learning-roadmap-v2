package m07

import (
	"encoding/json"
	"testing"
)

func registry() []ToolSpec {
	return []ToolSpec{{Name: "public_http", ReadOnly: true, AllowedMethods: []string{"GET"}, AllowedHosts: []string{"example.com"}, TimeoutMS: 10000, FollowRedirects: false}}
}

func TestValidateToolRequestIsExactAndReadOnly(t *testing.T) {
	for _, request := range []ToolRequest{
		{ToolName: "public_http", Method: "POST", Target: "https://example.com/a"},
		{ToolName: "public_http", Method: "GET", Target: "https://example.com.evil/a"},
		{ToolName: "public_http", Method: "GET", Target: "https://example.com:8443/a"},
	} {
		if err := ValidateToolRequest(request, registry()); err == nil {
			t.Fatalf("unsafe request accepted: %+v", request)
		}
	}
	if err := ValidateToolRequest(ToolRequest{ToolName: "public_http", Method: "GET", Target: "https://example.com/a"}, registry()); err != nil {
		t.Fatal(err)
	}
}

func TestValidateAgentOutputRequiresModelCitations(t *testing.T) {
	base := AgentOutput{State: "HUMAN_REVIEW", Answer: "limited", Claims: []Claim{{Text: "supported", FieldOrClaim: "price", Value: json.RawMessage("100"), EvidenceIDs: []string{"e1"}}}, EvidenceIDs: []string{"e1"}, ToolCalls: []ToolRequest{}, Authority: "A2-RO", WritePermission: false}
	raw, _ := json.Marshal(base)
	if _, err := ValidateAgentOutput(raw, []Evidence{{EvidenceID: "e1", FieldOrClaim: "price", Value: 100, ClaimKind: "assumption", Limitation: "synthetic"}}, registry()); err != nil {
		t.Fatal(err)
	}
	base.Claims[0].EvidenceIDs = []string{"forged"}
	raw, _ = json.Marshal(base)
	if _, err := ValidateAgentOutput(raw, []Evidence{{EvidenceID: "e1", ClaimKind: "assumption", Limitation: "synthetic"}}, registry()); err == nil {
		t.Fatal("forged claim citation accepted")
	}
}

func TestValidateAgentOutputRejectsForgedMeaningWithKnownID(t *testing.T) {
	output := AgentOutput{State: "HUMAN_REVIEW", Answer: "guaranteed profit", Claims: []Claim{{Text: "guaranteed profit", FieldOrClaim: "price", Value: json.RawMessage("1000000"), EvidenceIDs: []string{"e1"}}}, EvidenceIDs: []string{"e1"}, ToolCalls: []ToolRequest{}, Authority: "A2-RO", WritePermission: false}
	raw, _ := json.Marshal(output)
	if _, err := ValidateAgentOutput(raw, []Evidence{{EvidenceID: "e1", FieldOrClaim: "price", Value: 100, ClaimKind: "assumption", Limitation: "synthetic"}}, registry()); err == nil {
		t.Fatal("claim with a known ID but forged value accepted")
	}
}

func TestValidateAgentOutputRequiresLiteralFalseWritePermission(t *testing.T) {
	raw := []byte(`{"state":"ABSTAIN","answer":"insufficient","evidence_ids":[],"claims":[],"tool_calls":[],"authority":"A2-RO","write_permission":null}`)
	if _, err := ValidateAgentOutput(raw, []Evidence{{EvidenceID: "e1", ClaimKind: "assumption", Limitation: "synthetic"}}, registry()); err == nil {
		t.Fatal("null write_permission accepted")
	}
}
