package main

import (
	"encoding/json"
	"errors"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type ApprovalRecord struct {
	ApprovalID string `json:"approval_id"`; IntentID string `json:"intent_id"`; IntentHash string `json:"intent_hash"`; PolicyVersion string `json:"policy_version"`; Decision string `json:"decision"`; ApprovedBy string `json:"approved_by"`; ApproverID string `json:"approver_id"`; ApprovedAt string `json:"approved_at"`; ExpiresAt string `json:"expires_at"`; CorrelationID string `json:"correlation_id"`; OneTime bool `json:"one_time"`
}
type ExecutorProfile struct { ExecutorID string `json:"executor_id"`; AllowedActionTypes []string `json:"allowed_action_types"`; AllowedHosts []string `json:"allowed_hosts"` }
type ExecutionAuthorization struct { AuthorizationID string `json:"authorization_id"`; IntentID string `json:"intent_id"`; IntentHash string `json:"intent_hash"`; PolicyVersion string `json:"policy_version"`; ApprovalID string `json:"approval_id"`; ExecutorID string `json:"executor_id"`; AuthorizedAt string `json:"authorized_at"`; ExpiresAt string `json:"expires_at"`; IdempotencyKey string `json:"idempotency_key"`; CorrelationID string `json:"correlation_id"`; ExecutionAuthorized bool `json:"execution_authorized"` }
type ExecutionRecord struct { ExecutionID string `json:"execution_id"`; AuthorizationID string `json:"authorization_id"`; ApprovalID string `json:"approval_id"`; IntentID string `json:"intent_id"`; IntentHash string `json:"intent_hash"`; ExecutorID string `json:"executor_id"`; AttemptedAt string `json:"attempted_at"`; Status string `json:"status"`; SideEffectPerformed bool `json:"side_effect_performed"`; ExternalRef string `json:"external_ref,omitempty"`; Error string `json:"error,omitempty"`; CorrelationID string `json:"correlation_id"` }
type M09State struct { Intent ShadowActionIntent `json:"intent"`; Policy ShadowPolicyDecision `json:"policy"`; Approval *ApprovalRecord `json:"approval,omitempty"`; Authorization *ExecutionAuthorization `json:"authorization,omitempty"`; Execution *ExecutionRecord `json:"execution,omitempty"` }
type M09Context struct { Now string; KillSwitch bool; Executor ExecutorProfile; AlreadySucceeded map[string]bool }

func containsFold(xs []string, v string) bool { for _, x := range xs { if strings.EqualFold(strings.TrimSpace(x), strings.TrimSpace(v)) { return true } }; return false }
func executorAllows(p ExecutorProfile, i ShadowActionIntent) bool {
	if p.ExecutorID == "" || !containsFold(p.AllowedActionTypes, i.ActionType) { return false }
	u, err := url.Parse(i.Target); if err != nil || u.Scheme != "https" || u.Hostname()=="" { return false }
	return containsFold(p.AllowedHosts, u.Hostname())
}
func minTime(a,b time.Time) time.Time { if a.Before(b) { return a }; return b }

func AuthorizeM09(state M09State, ctx M09Context) (ExecutionAuthorization, string) {
	i, p := state.Intent, state.Policy
	if i.IntentHash=="" || i.IntentHash!=ComputeShadowIntentHash(i) { return ExecutionAuthorization{}, "DENY_TAMPERED_INTENT" }
	now, e1:=time.Parse(time.RFC3339,ctx.Now); ie, e2:=time.Parse(time.RFC3339,i.ExpiresAt); if e1!=nil||e2!=nil||!ie.After(now) { return ExecutionAuthorization{}, "DENY_EXPIRED_INTENT" }
	if p.IntentID!=i.IntentID || p.IntentHash!=i.IntentHash || p.PolicyVersion=="" || (p.Decision!="ALLOW" && p.Decision!="HUMAN_REVIEW") { return ExecutionAuthorization{}, "DENY_POLICY_STATE" }
	if state.Approval==nil { return ExecutionAuthorization{}, "WAIT_APPROVAL" }
	a:=*state.Approval
	if a.ApprovedBy!="human" || strings.TrimSpace(a.ApproverID)=="" || !a.OneTime { return ExecutionAuthorization{}, "DENY_INVALID_APPROVER" }
	if a.Decision=="REJECT" { return ExecutionAuthorization{}, "DENY_REJECTED" }; if a.Decision!="APPROVE" { return ExecutionAuthorization{}, "DENY_INVALID_APPROVER" }
	if a.IntentID!=i.IntentID || a.IntentHash!=i.IntentHash || a.PolicyVersion!=p.PolicyVersion || a.CorrelationID!=i.CorrelationID { return ExecutionAuthorization{}, "DENY_APPROVAL_MISMATCH" }
	aa,e3:=time.Parse(time.RFC3339,a.ApprovedAt); ae,e4:=time.Parse(time.RFC3339,a.ExpiresAt); if e3!=nil||e4!=nil||aa.After(now)||!ae.After(now)||!ae.After(aa) { return ExecutionAuthorization{}, "DENY_EXPIRED_APPROVAL" }
	if ctx.KillSwitch { return ExecutionAuthorization{}, "DENY_KILL_SWITCH" }
	if !executorAllows(ctx.Executor,i) { return ExecutionAuthorization{}, "DENY_EXECUTOR" }
	if ctx.AlreadySucceeded[i.IdempotencyKey] { return ExecutionAuthorization{}, "WAIT_ALREADY_EXECUTED" }
	expires:=minTime(ie,ae)
	return ExecutionAuthorization{AuthorizationID:"auth-"+a.ApprovalID,IntentID:i.IntentID,IntentHash:i.IntentHash,PolicyVersion:p.PolicyVersion,ApprovalID:a.ApprovalID,ExecutorID:ctx.Executor.ExecutorID,AuthorizedAt:ctx.Now,ExpiresAt:expires.Format(time.RFC3339),IdempotencyKey:i.IdempotencyKey,CorrelationID:i.CorrelationID,ExecutionAuthorized:true}, "AUTHORIZED"
}

func PersistM09State(path string, s M09State) error { b,e:=json.MarshalIndent(s,"","  "); if e!=nil{return e}; if e=os.MkdirAll(filepath.Dir(path),0755);e!=nil{return e}; tmp:=path+".tmp"; if e=os.WriteFile(tmp,b,0600);e!=nil{return e}; return os.Rename(tmp,path) }
func LoadM09State(path string) (M09State,error) { var s M09State; b,e:=os.ReadFile(path); if e!=nil{return s,e}; e=json.Unmarshal(b,&s); return s,e }

func ExecuteLocalSandbox(state M09State, auth ExecutionAuthorization, ctx M09Context, dir string) (ExecutionRecord,string) {
	if ctx.KillSwitch { return ExecutionRecord{}, "DENY_KILL_SWITCH" }
	if !auth.ExecutionAuthorized || auth.IntentID!=state.Intent.IntentID || auth.IntentHash!=state.Intent.IntentHash || auth.ApprovalID=="" || auth.ExecutorID!=ctx.Executor.ExecutorID { return ExecutionRecord{}, "DENY_AUTHORIZATION" }
	now,e1:=time.Parse(time.RFC3339,ctx.Now); exp,e2:=time.Parse(time.RFC3339,auth.ExpiresAt); if e1!=nil||e2!=nil||!exp.After(now) { return ExecutionRecord{}, "DENY_EXPIRED_AUTHORIZATION" }
	if ctx.AlreadySucceeded[auth.IdempotencyKey] { return ExecutionRecord{}, "WAIT_ALREADY_EXECUTED" }
	if !executorAllows(ctx.Executor,state.Intent) || ctx.Executor.ExecutorID!="local_sandbox" { return ExecutionRecord{}, "DENY_EXECUTOR" }
	if e:=os.MkdirAll(dir,0755); e!=nil{return ExecutionRecord{},"EXECUTION_FAILED"}
	path:=filepath.Join(dir,auth.AuthorizationID+".json")
	payload:=map[string]any{"intent_id":state.Intent.IntentID,"intent_hash":state.Intent.IntentHash,"action_type":state.Intent.ActionType,"target":state.Intent.Target,"parameters":state.Intent.Parameters,"approved_by":"human"}
	b,e:=json.MarshalIndent(payload,"","  "); if e!=nil{return ExecutionRecord{},"EXECUTION_FAILED"}; if e=os.WriteFile(path,b,0600);e!=nil{return ExecutionRecord{},"EXECUTION_FAILED"}
	r:=ExecutionRecord{ExecutionID:"exec-"+auth.AuthorizationID,AuthorizationID:auth.AuthorizationID,ApprovalID:auth.ApprovalID,IntentID:auth.IntentID,IntentHash:auth.IntentHash,ExecutorID:auth.ExecutorID,AttemptedAt:ctx.Now,Status:"SUCCEEDED",SideEffectPerformed:true,ExternalRef:path,CorrelationID:auth.CorrelationID}
	return r,"EXECUTED"
}

func demoM09() (map[string]any,error) {
	i:=SealShadowActionIntent(ShadowActionIntent{IntentID:"i9",DecisionID:"d9",EvidenceIDs:[]string{"e9"},ActionType:"UPDATE_DRAFT",Target:"https://example.com/draft",Parameters:map[string]any{"title":"approved sandbox write"},ProposedBy:"human",CreatedAt:"2026-09-03T07:00:00Z",ExpiresAt:"2026-09-03T09:00:00Z",CorrelationID:"c9",IdempotencyKey:"k9"})
	p:=ShadowPolicyDecision{PolicyVersion:"m09-v1",IntentID:i.IntentID,IntentHash:i.IntentHash,Decision:"HUMAN_REVIEW",RiskClass:"RISK1",Reason:"RISK1_REQUIRES_REVIEW",ApprovalRequired:true,ShadowOnly:true,ExecutionAuthorized:false,PolicyCheckedAt:"2026-09-03T07:10:00Z"}
	a:=ApprovalRecord{ApprovalID:"ap9",IntentID:i.IntentID,IntentHash:i.IntentHash,PolicyVersion:p.PolicyVersion,Decision:"APPROVE",ApprovedBy:"human",ApproverID:"learner",ApprovedAt:"2026-09-03T07:20:00Z",ExpiresAt:"2026-09-03T08:30:00Z",CorrelationID:i.CorrelationID,OneTime:true}
	s:=M09State{Intent:i,Policy:p,Approval:&a}; ctx:=M09Context{Now:"2026-09-03T08:00:00Z",Executor:ExecutorProfile{ExecutorID:"local_sandbox",AllowedActionTypes:[]string{"UPDATE_DRAFT"},AllowedHosts:[]string{"example.com"}},AlreadySucceeded:map[string]bool{}}
	auth,status:=AuthorizeM09(s,ctx); if status!="AUTHORIZED" {return nil,errors.New(status)}; s.Authorization=&auth
	dir,e:=os.MkdirTemp("","m09-sandbox-"); if e!=nil{return nil,e}; rec,execStatus:=ExecuteLocalSandbox(s,auth,ctx,dir); if execStatus!="EXECUTED"{return nil,errors.New(execStatus)}
	return map[string]any{"authorization":auth,"execution":rec,"authority":"HUMAN_APPROVAL_REQUIRED"},nil
}
