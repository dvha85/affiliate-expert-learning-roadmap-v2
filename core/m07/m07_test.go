package m07

import (
	"bytes"
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
	base := AgentOutput{State: "HUMAN_REVIEW", Claims: []Claim{{FieldOrClaim: "price", Value: json.RawMessage("100"), EvidenceIDs: []string{"e1"}}}, EvidenceIDs: []string{"e1"}, ToolCalls: []ToolRequest{}, Authority: "A2-RO", WritePermission: false}
	base.Claims[0].Text = renderClaim(base.Claims[0])
	base.Answer = RenderGroundedAnswer(base.Claims)
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
	output := AgentOutput{State: "HUMAN_REVIEW", Answer: "price=1000000 [evidence:e1]", Claims: []Claim{{FieldOrClaim: "price", Value: json.RawMessage("1000000"), EvidenceIDs: []string{"e1"}}}, EvidenceIDs: []string{"e1"}, ToolCalls: []ToolRequest{}, Authority: "A2-RO", WritePermission: false}
	output.Claims[0].Text = renderClaim(output.Claims[0])
	raw, _ := json.Marshal(output)
	if _, err := ValidateAgentOutput(raw, []Evidence{{EvidenceID: "e1", FieldOrClaim: "price", Value: 100, ClaimKind: "assumption", Limitation: "synthetic"}}, registry()); err == nil {
		t.Fatal("claim with a known ID but forged value accepted")
	}
}

func TestValidateAgentOutputPreservesExactJSONNumbers(t *testing.T) {
	claim := Claim{FieldOrClaim: "amount", Value: json.RawMessage("9007199254740992"), EvidenceIDs: []string{"e1"}}
	claim.Text = renderClaim(claim)
	output := AgentOutput{State: "HUMAN_REVIEW", Claims: []Claim{claim}, EvidenceIDs: []string{"e1"}, ToolCalls: []ToolRequest{}, Authority: "A2-RO", WritePermission: false}
	output.Answer = RenderGroundedAnswer(output.Claims)
	raw, _ := json.Marshal(output)
	if _, err := ValidateAgentOutput(raw, []Evidence{{EvidenceID: "e1", FieldOrClaim: "amount", Value: json.Number("9007199254740993"), ClaimKind: "fact", Limitation: "exact fixture"}}, registry()); err == nil {
		t.Fatal("rounded large number was accepted as grounded evidence")
	}
	claim.Value = json.RawMessage("9007199254740993")
	claim.Text = renderClaim(claim)
	output.Claims, output.Answer = []Claim{claim}, RenderGroundedAnswer([]Claim{claim})
	raw, _ = json.Marshal(output)
	if _, err := ValidateAgentOutput(raw, []Evidence{{EvidenceID: "e1", FieldOrClaim: "amount", Value: json.Number("9007199254740993"), ClaimKind: "fact", Limitation: "exact fixture"}}, registry()); err != nil {
		t.Fatal(err)
	}
	if output.Answer != "amount=9007199254740993 [evidence:e1]" {
		t.Fatalf("large number was rendered imprecisely: %s", output.Answer)
	}
}

func TestValidateAgentOutputRequiresLiteralFalseWritePermission(t *testing.T) {
	raw := []byte(`{"state":"ABSTAIN","answer":"insufficient","evidence_ids":[],"claims":[],"tool_calls":[],"authority":"A2-RO","write_permission":null}`)
	if _, err := ValidateAgentOutput(raw, []Evidence{{EvidenceID: "e1", ClaimKind: "assumption", Limitation: "synthetic"}}, registry()); err == nil {
		t.Fatal("null write_permission accepted")
	}
}

func TestValidateAgentOutputRejectsNonCanonicalJSONShape(t *testing.T) {
	valid := `{"state":"ABSTAIN","answer":"insufficient","evidence_ids":[],"claims":[],"tool_calls":[],"authority":"A2-RO","write_permission":false}`
	claim := `{"text":"price=100 [evidence:e1]","field_or_claim":"price","value":100,"evidence_ids":["e1"]}`
	cases := []struct {
		name string
		raw  string
	}{
		{"duplicate top-level write permission", `{"state":"ABSTAIN","answer":"insufficient","evidence_ids":[],"claims":[],"tool_calls":[],"authority":"A2-RO","write_permission":true,"write_permission":false}`},
		{"duplicate nested claim field", `{"state":"HUMAN_REVIEW","answer":"price=100 [evidence:e1]","evidence_ids":["e1"],"claims":[{"text":"price=100 [evidence:e1]","field_or_claim":"price","field_or_claim":"other","value":100,"evidence_ids":["e1"]}],"tool_calls":[],"authority":"A2-RO","write_permission":false}`},
		{"case-variant unknown field", valid[:len(valid)-1] + `,"Write_Permission":false}`},
		{"unknown top-level field", valid[:len(valid)-1] + `,"unreviewed_instruction":"write now"}`},
		{"trailing JSON", valid + ` {}`},
		{"unknown nested claim field", `{"state":"HUMAN_REVIEW","answer":"price=100 [evidence:e1]","evidence_ids":["e1"],"claims":[` + claim[:len(claim)-1] + `,"unreviewed_instruction":"write now"}],"tool_calls":[],"authority":"A2-RO","write_permission":false}`},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			if _, err := ValidateAgentOutput([]byte(test.raw), []Evidence{{EvidenceID: "e1", FieldOrClaim: "price", Value: 100, ClaimKind: "assumption", Limitation: "synthetic"}}, registry()); err == nil {
				t.Fatalf("non-canonical model JSON was accepted: %s", test.raw)
			}
		})
	}
}

func TestValidateAgentOutputRejectsForgedProseAndUnregisteredToolCall(t *testing.T) {
	base := AgentOutput{State: "HUMAN_REVIEW", Answer: "guaranteed profit", Claims: []Claim{{FieldOrClaim: "price", Value: json.RawMessage("100"), EvidenceIDs: []string{"e1"}}}, EvidenceIDs: []string{"e1"}, Authority: "A2-RO", WritePermission: false}
	base.Claims[0].Text = renderClaim(base.Claims[0])
	raw, _ := json.Marshal(base)
	if _, err := ValidateAgentOutput(raw, []Evidence{{EvidenceID: "e1", FieldOrClaim: "price", Value: 100, ClaimKind: "assumption", Limitation: "synthetic"}}, registry()); err == nil {
		t.Fatal("free model prose was labeled grounded")
	}
	base.Answer = RenderGroundedAnswer(base.Claims)
	base.ToolCalls = []ToolRequest{{ToolName: "public_http", Method: "GET", Target: "https://example.com/a"}}
	raw, _ = json.Marshal(base)
	if _, err := ValidateAgentOutput(raw, []Evidence{{EvidenceID: "e1", FieldOrClaim: "price", Value: 100, ClaimKind: "assumption", Limitation: "synthetic"}}, registry()); err == nil {
		t.Fatal("model-declared tool call was accepted without a registered trace")
	}
}

func TestRegisteredToolResultIsBoundToRequestAndRecord(t *testing.T) {
	raw := []byte(`{"record_id":"r1","tool_call":{"tool_name":"public_http","method":"GET","target":"https://example.com/a"},"status_code":200,"received_at":"2026-09-08T00:00:00Z","redirected":false,"body":{"price":100}}`)
	registered, err := RegisterToolResult(raw, registry())
	if err != nil {
		t.Fatal(err)
	}
	stored, _ := json.Marshal(registered)
	resolved, err := ValidateRegisteredToolResult(stored, registry(), "r1")
	if err != nil || resolved.Evidence().EvidenceID != registered.TraceID+"#body" {
		t.Fatalf("registered result did not resolve: %v", err)
	}
	tampered := string(stored)
	tampered = string(bytes.Replace([]byte(tampered), []byte(`"trace_id":"`+registered.TraceID), []byte(`"trace_id":"sha256:forged`), 1))
	if _, err := ValidateRegisteredToolResult([]byte(tampered), registry(), "r1"); err == nil {
		t.Fatal("forged trace id accepted")
	}
	if _, err := ValidateRegisteredToolResult(stored, registry(), "other"); err == nil {
		t.Fatal("cross-record result accepted")
	}
}

func TestRegisteredAgentProposalRerunsGroundingAndDigest(t *testing.T) {
	claim := Claim{FieldOrClaim: "price", Value: json.RawMessage("100"), EvidenceIDs: []string{"e1"}}
	claim.Text = renderClaim(claim)
	output := AgentOutput{State: "HUMAN_REVIEW", Claims: []Claim{claim}, EvidenceIDs: []string{"e1"}, Authority: "A2-RO", WritePermission: false, ProposedAction: &ProposedAction{ActionType: "DRAFT", Target: "https://example.com/draft", Parameters: json.RawMessage(`{"id":9007199254740993}`)}}
	output.Answer = RenderGroundedAnswer(output.Claims)
	raw, _ := json.Marshal(output)
	registered, err := RegisterAgentProposal(raw, []Evidence{{EvidenceID: "e1", FieldOrClaim: "price", Value: 100, ClaimKind: "assumption", Limitation: "synthetic"}}, registry(), "r1")
	if err != nil {
		t.Fatal(err)
	}
	stored, _ := json.Marshal(registered)
	if _, _, err := ValidateRegisteredAgentProposal(stored, []Evidence{{EvidenceID: "e1", FieldOrClaim: "price", Value: 100, ClaimKind: "assumption", Limitation: "synthetic"}}, registry(), "r1"); err != nil {
		t.Fatal(err)
	}
	registered.OutputDigest = "sha256:forged"
	tampered, _ := json.Marshal(registered)
	if _, _, err := ValidateRegisteredAgentProposal(tampered, []Evidence{{EvidenceID: "e1", FieldOrClaim: "price", Value: 100, ClaimKind: "assumption", Limitation: "synthetic"}}, registry(), "r1"); err == nil {
		t.Fatal("forged proposal digest accepted")
	}
}
