package main

import (
	"path/filepath"
	"testing"
)

type m09Case struct { CaseID string `json:"case_id"`; Mutation string `json:"mutation"`; Expected string `json:"expected"` }
func baseM09() (M09State,M09Context) {
	i:=SealShadowActionIntent(ShadowActionIntent{IntentID:"i",DecisionID:"d",EvidenceIDs:[]string{"e"},ActionType:"UPDATE_DRAFT",Target:"https://example.com/draft",Parameters:map[string]any{"x":1},ProposedBy:"human",CreatedAt:"2026-09-03T07:00:00Z",ExpiresAt:"2026-09-03T09:00:00Z",CorrelationID:"c",IdempotencyKey:"k"})
	p:=ShadowPolicyDecision{PolicyVersion:"v1",IntentID:i.IntentID,IntentHash:i.IntentHash,Decision:"HUMAN_REVIEW",RiskClass:"RISK1",Reason:"review",ApprovalRequired:true,ShadowOnly:true,ExecutionAuthorized:false,PolicyCheckedAt:"2026-09-03T07:05:00Z"}
	a:=ApprovalRecord{ApprovalID:"a",IntentID:i.IntentID,IntentHash:i.IntentHash,PolicyVersion:"v1",Decision:"APPROVE",ApprovedBy:"human",ApproverID:"u",ApprovedAt:"2026-09-03T07:10:00Z",ExpiresAt:"2026-09-03T08:30:00Z",CorrelationID:"c",OneTime:true}
	return M09State{Intent:i,Policy:p,Approval:&a},M09Context{Now:"2026-09-03T08:00:00Z",Executor:ExecutorProfile{ExecutorID:"local_sandbox",AllowedActionTypes:[]string{"UPDATE_DRAFT"},AllowedHosts:[]string{"example.com"}},AllowedExecutorIDs:[]string{"local_sandbox"},AlreadySucceeded:map[string]bool{}}
}
func TestM09EvalPack(t *testing.T){
	cases:=loadCases[m09Case](t,"M09-approval-execution")
	for _,c:=range cases{t.Run(c.CaseID,func(t *testing.T){s,ctx:=baseM09(); switch c.Mutation{
	case "no_approval": s.Approval=nil
	case "rejected": s.Approval.Decision="REJECT"
	case "machine_approver": s.Approval.ApprovedBy="agent"
	case "approval_hash": s.Approval.IntentHash="sha256:0000000000000000000000000000000000000000000000000000000000000000"
	case "approval_policy": s.Approval.PolicyVersion="v0"
	case "expired_approval": s.Approval.ExpiresAt="2026-09-03T07:30:00Z"
	case "kill_switch": ctx.KillSwitch=true
	case "wrong_executor": ctx.Executor.ExecutorID="other"
	case "bad_target": s.Intent.Target="https://evil.example/draft"; s.Intent.IntentHash=ComputeShadowIntentHash(s.Intent); s.Policy.IntentHash=s.Intent.IntentHash; s.Approval.IntentHash=s.Intent.IntentHash
	case "policy_deny": s.Policy.Decision="DENY"
	case "already_executed": ctx.AlreadySucceeded["k"]=true
	case "tampered_intent": s.Intent.Target="https://evil.example/draft"
	case "expired_intent": ctx.Now="2026-09-03T09:30:00Z"
	}; _,got:=AuthorizeM09(s,ctx); if got!=c.Expected{t.Fatalf("expected %s got %s",c.Expected,got)} })}
}
func TestM09DurableResumeRevalidates(t *testing.T){s,ctx:=baseM09(); p:=filepath.Join(t.TempDir(),"state.json"); if e:=PersistM09State(p,s);e!=nil{t.Fatal(e)}; loaded,e:=LoadM09State(p);if e!=nil{t.Fatal(e)}; loaded.Policy.PolicyVersion="v2"; _,got:=AuthorizeM09(loaded,ctx);if got!="DENY_APPROVAL_MISMATCH"{t.Fatalf("resume must revalidate approval binding, got %s",got)}}
func TestM09ControlledExecutorAndIdempotency(t *testing.T){s,ctx:=baseM09(); auth,status:=AuthorizeM09(s,ctx);if status!="AUTHORIZED"{t.Fatal(status)}; r,status:=ExecuteLocalSandbox(s,auth,ctx,t.TempDir());if status!="EXECUTED"||!r.SideEffectPerformed||r.Status!="SUCCEEDED"{t.Fatalf("unexpected execution %+v %s",r,status)};ctx.AlreadySucceeded[auth.IdempotencyKey]=true;if _,status=ExecuteLocalSandbox(s,auth,ctx,t.TempDir());status!="WAIT_ALREADY_EXECUTED"{t.Fatalf("duplicate must not execute: %s",status)}}
func TestM09KillSwitchRecheckedAtExecution(t *testing.T){s,ctx:=baseM09();auth,status:=AuthorizeM09(s,ctx);if status!="AUTHORIZED"{t.Fatal(status)};ctx.KillSwitch=true;if _,status=ExecuteLocalSandbox(s,auth,ctx,t.TempDir());status!="DENY_KILL_SWITCH"{t.Fatalf("kill switch must stop execution: %s",status)}}
func TestM09ExecutorProfileCannotSelfAuthorize(t *testing.T){s,ctx:=baseM09();ctx.Executor.ExecutorID="other";ctx.Executor.AllowedActionTypes=[]string{"UPDATE_DRAFT"};ctx.Executor.AllowedHosts=[]string{"example.com"};if _,status:=AuthorizeM09(s,ctx);status!="DENY_EXECUTOR"{t.Fatalf("unregistered executor must be denied: %s",status)}}
