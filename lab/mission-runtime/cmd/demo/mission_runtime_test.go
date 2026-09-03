package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
)

func evalPath(t *testing.T, dir string) string { t.Helper(); _,f,_,ok:=runtime.Caller(0); if !ok { t.Fatal("cannot resolve test file path") }; return filepath.Clean(filepath.Join(filepath.Dir(f),"..","..","..","..","evals",dir,"cases.json")) }
func loadCases[T any](t *testing.T, dir string) []T { t.Helper(); raw,err:=os.ReadFile(evalPath(t,dir)); if err!=nil { t.Fatal(err) }; var cases []T; if err:=json.Unmarshal(raw,&cases); err!=nil { t.Fatal(err) }; return cases }

func TestO00EvalPack(t *testing.T) {
	type tc struct { CaseID string `json:"case_id"`; ExpectedFinalState string `json:"expected_final_state"`; ExpectedExternalSideEffects *bool `json:"expected_external_side_effects"`; ExpectedValidation string `json:"expected_validation"`; RequiredKind string `json:"required_kind"` }
	report:=RunSyntheticWalkthrough()
	for _,c:=range loadCases[tc](t,"O00-safe-walkthrough") { t.Run(c.CaseID,func(t *testing.T){
		if c.ExpectedFinalState!="" && report.FinalState!=c.ExpectedFinalState { t.Fatalf("expected %s got %s",c.ExpectedFinalState,report.FinalState) }
		if c.ExpectedExternalSideEffects!=nil && report.ExternalSideEffects!=*c.ExpectedExternalSideEffects { t.Fatal("side effect mismatch") }
		if c.ExpectedValidation!="" && ValidateSyntheticWalkthrough(report)!=c.ExpectedValidation { t.Fatalf("walkthrough validation failed: %s",ValidateSyntheticWalkthrough(report)) }
		if c.RequiredKind!="" { found:=false; for _,a:=range report.Artifacts { if a.Kind==c.RequiredKind { found=true } }; if !found { t.Fatalf("missing artifact kind %s",c.RequiredKind) } }
	}) }
}

func TestM03EvalPack(t *testing.T) {
	type tc struct { CaseID string `json:"case_id"`; Mode string `json:"mode"`; Expected string `json:"expected"`; Record HumanActionRecord `json:"record"`; Outcome OutcomeRecord `json:"outcome"` }
	for _,c:=range loadCases[tc](t,"M03-tracked-human-action") { t.Run(c.CaseID,func(t *testing.T){ var got string; switch c.Mode { case "action": got=ValidateHumanActionRecord(c.Record); case "outcome": got=ValidateOutcomeRecord(c.Outcome); case "link": got=ValidateActionOutcomeLink(c.Record,c.Outcome); default: t.Fatalf("unknown mode %s",c.Mode) }; if got!=c.Expected { t.Fatalf("expected %s got %s",c.Expected,got) } }) }
}

func TestM04EvalPack(t *testing.T) {
	type tc struct { CaseID string `json:"case_id"`; Output AdvisorOutput `json:"output"`; Evidence []AdvisorEvidence `json:"evidence"`; AsOf string `json:"as_of"`; MaxAgeHours int `json:"max_age_hours"`; Expected string `json:"expected"` }
	for _,c:=range loadCases[tc](t,"M04-grounded-ai-advisor") { t.Run(c.CaseID,func(t *testing.T){ if got:=EvaluateAdvisorOutput(c.Output,c.Evidence,c.AsOf,c.MaxAgeHours); got!=c.Expected { t.Fatalf("expected %s got %s",c.Expected,got) } }) }
}

func TestM05EvalPack(t *testing.T) {
	type tc struct { CaseID string `json:"case_id"`; Mode string `json:"mode"`; Proposal ImprovementProposal `json:"proposal"`; Evaluation EvaluationRecord `json:"evaluation"`; Evaluations []EvaluationRecord `json:"evaluations"`; Action HumanActionRecord `json:"action"`; Outcomes []OutcomeRecord `json:"outcomes"`; Review ReviewRecord `json:"review"`; Expected string `json:"expected"` }
	for _,c:=range loadCases[tc](t,"M05-reviewed-improvement") { t.Run(c.CaseID,func(t *testing.T){ var got string; switch c.Mode { case "proposal": got=EvaluateImprovementProposal(c.Proposal); case "evaluation": got=ValidateEvaluationRecord(c.Evaluation,c.Action,c.Outcomes); case "proposal-link": got=ValidateProposalEvaluationLink(c.Proposal,c.Evaluations); case "review": got=ValidateReviewRecord(c.Review,c.Proposal); default: t.Fatalf("unknown mode %s",c.Mode) }; if got!=c.Expected { t.Fatalf("expected %s got %s",c.Expected,got) } }) }
}

func TestM06EvalPack(t *testing.T) {
	type tc struct { CaseID string `json:"case_id"`; Request WatchRequest `json:"request"`; Expected string `json:"expected"` }
	for _,c:=range loadCases[tc](t,"M06-readonly-watcher") { t.Run(c.CaseID,func(t *testing.T){ if got:=EvaluateWatchRequest(c.Request); got!=c.Expected { t.Fatalf("expected %s got %s",c.Expected,got) } }) }
}
func TestM06NormalizesCanonicalObservation(t *testing.T) {
	r:=WatchRequest{Method:"GET",URL:"https://example.com/offer",AllowHosts:[]string{"example.com"},ObservedAt:"2026-09-03T01:00:00Z",CorrelationID:"c1",Body:"price=100"}
	o,state:=NormalizeWatchObservation(r,"product-a"); if state!="NEW" { t.Fatalf("expected NEW got %s",state) }; if o.SubjectID=="" || o.SourceURL=="" || o.EvidenceKind!="real" || o.ClaimKind!="fact" || o.Limitation=="" || o.ContentHash=="" { t.Fatalf("not canonical enough: %+v",o) }
}

func TestM07EvalPack(t *testing.T) {
	type tc struct { CaseID string `json:"case_id"`; Proposal AgentProposal `json:"proposal"`; Registry []ToolSpec `json:"registry"`; AvailableEvidenceIDs []string `json:"available_evidence_ids"`; ToolResultText string `json:"tool_result_text"`; Expected string `json:"expected"` }
	seenInjection:=false
	for _,c:=range loadCases[tc](t,"M07-readonly-evidence-agent") { t.Run(c.CaseID,func(t *testing.T){ if c.ToolResultText!="" { seenInjection=true }; got:=EvaluateAgentProposal(c.Proposal,c.Registry,c.AvailableEvidenceIDs,c.ToolResultText); if got!=c.Expected { t.Fatalf("expected %s got %s",c.Expected,got) }; if strings.Contains(c.CaseID,"prompt-injection") && c.ToolResultText=="" { t.Fatal("prompt-injection case must carry malicious tool text") } }) }
	if !seenInjection { t.Fatal("M07 eval pack must exercise untrusted tool text") }
}

func TestSortedStringsDoesNotMutateInput(t *testing.T) { in:=[]string{"b","a"}; got:=sortedStrings(in); if !reflect.DeepEqual(got,[]string{"a","b"}) || !reflect.DeepEqual(in,[]string{"b","a"}) { t.Fatal("sort helper must be deterministic and non-mutating") } }
