package main

import (
	"os"
	"testing"
	"time"
)

type m11Case struct {
	CaseID string `json:"case_id"`
	Mutation string `json:"mutation"`
	ExpectedDecision string `json:"expected_decision"`
	ExpectedReason string `json:"expected_reason"`
}

func baseM11() (M11State, M11Context) {
	s, c := baseDemoM11()
	s.Intent.IntentID = "i"; s.Intent.DecisionID = "d"; s.Intent.EvidenceIDs = []string{"e"}; s.Intent.IdempotencyKey = "k"; s.Intent.Target = "https://example.com/prod"; s.Intent = SealShadowActionIntent(s.Intent)
	c.PolicyContext.KnownDecisionIDs = []string{"d"}; c.PolicyContext.KnownEvidenceIDs = []string{"e"}; c.PolicyContext.ActionRisk = map[string]string{"PREPARE_LOCAL_DRAFT":"RISK0","UPDATE_DRAFT":"RISK1","PUBLISH_CONTENT":"RISK2","UPDATE_PROFILE":"RISK0"}; c.PolicyContext.AllowedHosts = []string{"example.com","other.example"}
	pc := c.PolicyContext; pc.Now = "2026-09-03T07:05:00Z"; s.Policy = EvaluateShadowPolicy(s.Intent, pc)
	s.Lease.AllowedRiskClasses = []string{"RISK0","RISK1"}; s.Lease.AllowedActionTypes = []string{"PREPARE_LOCAL_DRAFT","UPDATE_DRAFT"}; s.Lease.MaxExecutionsTotal = 3; s.Lease.MaxExecutionsPerWindow = 2; s.Lease.MaxCostMinorTotal = 1000; s.Lease = SealProductionLease(s.Lease)
	s.Ledger = ProductionLedger{LeaseID:s.Lease.LeaseID,LeaseVersion:s.Lease.LeaseVersion,LeaseHash:s.Lease.LeaseHash,ControlMode:"NORMAL",WindowStartedAt:"2026-09-03T07:15:00Z",UpdatedAt:"2026-09-03T07:15:00Z"}
	s.Health.LeaseHash = s.Lease.LeaseHash; s.Health = SealProductionHealth(s.Health)
	s.CostBound = SealCanaryCostBound(CanaryCostBound{CostBoundID:"cost-i",IntentID:s.Intent.IntentID,IntentHash:s.Intent.IntentHash,MaxCostMinor:100,Currency:"USD",SourceRef:"deterministic-cost-registry",ObservedAt:"2026-09-03T07:55:00Z",ExpiresAt:"2026-09-03T09:00:00Z",CorrelationID:s.Intent.CorrelationID})
	approval := ProductionLeaseApproval{ApprovalID:s.Lease.ApprovalRef,LeaseID:s.Lease.LeaseID,LeaseVersion:s.Lease.LeaseVersion,LeaseHash:s.Lease.LeaseHash,PromotionReviewRef:s.Lease.PromotionReviewRef,SourceCanaryGrantID:s.Lease.SourceCanaryGrantID,SourceCanaryGrantVersion:s.Lease.SourceCanaryGrantVersion,SourceCanaryGrantHash:s.Lease.SourceCanaryGrantHash,SourceE5Refs:[]string{"e5-canary-10"},ValidatedRiskClasses:[]string{"RISK0","RISK1"},ReviewedBy:"human",ReviewerID:s.Lease.ReviewerID,ReviewedAt:s.Lease.ReviewedAt,Decision:"APPROVE_PRODUCTION_LEASE"}
	c.TrustedLeaseApprovals = map[string]ProductionLeaseApproval{approval.ApprovalID:approval}; c.KnownCanaryGrantHashes = map[string]string{productionCanaryKey(s.Lease.SourceCanaryGrantID,s.Lease.SourceCanaryGrantVersion):s.Lease.SourceCanaryGrantHash}; c.KnownE5Refs = []string{"e5-canary-10"}
	c.TrustedHealthSnapshots = map[string]string{s.Health.SnapshotID:s.Health.SnapshotHash}; c.TrustedCostBounds = map[string]string{s.CostBound.CostBoundID:s.CostBound.CostBoundHash}
	c.Executor = ExecutorProfile{ExecutorID:"local_sandbox",AllowedActionTypes:[]string{"PREPARE_LOCAL_DRAFT","UPDATE_DRAFT","UPDATE_PROFILE","PUBLISH_CONTENT"},AllowedHosts:[]string{"example.com","other.example"}}; c.AllowedExecutorIDs = []string{"local_sandbox"}
	return s, c
}

func refreshM11Intent(s *M11State,c *M11Context,action,target,id,idem string,cost int64) {
	s.Intent.IntentID=id; s.Intent.ActionType=action; s.Intent.Target=target; s.Intent.IdempotencyKey=idem; s.Intent=SealShadowActionIntent(s.Intent)
	pc:=c.PolicyContext; pc.Now="2026-09-03T07:05:00Z"; s.Policy=EvaluateShadowPolicy(s.Intent,pc)
	s.CostBound=SealCanaryCostBound(CanaryCostBound{CostBoundID:"cost-"+id,IntentID:s.Intent.IntentID,IntentHash:s.Intent.IntentHash,MaxCostMinor:cost,Currency:"USD",SourceRef:"deterministic-cost-registry",ObservedAt:"2026-09-03T07:55:00Z",ExpiresAt:"2026-09-03T09:00:00Z",CorrelationID:s.Intent.CorrelationID}); c.TrustedCostBounds[s.CostBound.CostBoundID]=s.CostBound.CostBoundHash
}
func refreshLeaseTrust(s *M11State,c *M11Context) {
	s.Lease=SealProductionLease(s.Lease); s.Ledger.LeaseHash=s.Lease.LeaseHash; s.Health.LeaseHash=s.Lease.LeaseHash; s.Health=SealProductionHealth(s.Health)
	a:=c.TrustedLeaseApprovals[s.Lease.ApprovalRef]; a.LeaseHash=s.Lease.LeaseHash; a.LeaseID=s.Lease.LeaseID; a.LeaseVersion=s.Lease.LeaseVersion; a.PromotionReviewRef=s.Lease.PromotionReviewRef; a.SourceCanaryGrantID=s.Lease.SourceCanaryGrantID; a.SourceCanaryGrantVersion=s.Lease.SourceCanaryGrantVersion; a.SourceCanaryGrantHash=s.Lease.SourceCanaryGrantHash; a.ReviewerID=s.Lease.ReviewerID; a.ReviewedAt=s.Lease.ReviewedAt; c.TrustedLeaseApprovals[s.Lease.ApprovalRef]=a
	c.KnownCanaryGrantHashes[productionCanaryKey(s.Lease.SourceCanaryGrantID,s.Lease.SourceCanaryGrantVersion)]=s.Lease.SourceCanaryGrantHash; c.TrustedHealthSnapshots[s.Health.SnapshotID]=s.Health.SnapshotHash
}
func refreshHealthTrust(s *M11State,c *M11Context){s.Health=SealProductionHealth(s.Health);c.TrustedHealthSnapshots[s.Health.SnapshotID]=s.Health.SnapshotHash}
func initM11(t *testing.T,s *M11State,dir,now string){t.Helper();if got:=InitializeProductionLedger(s,dir,now);got!="INITIALIZED"{t.Fatalf("activation failed: %s",got)}}

func TestM11EvalPack(t *testing.T){
	for _,x:=range loadCases[m11Case](t,"M11-production-closed-loop") { t.Run(x.CaseID,func(t *testing.T){
		s,c:=baseM11()
		switch x.Mutation {
		case "risk1": refreshM11Intent(&s,&c,"UPDATE_DRAFT","https://example.com/prod","i-r1","k-r1",100)
		case "risk2": refreshM11Intent(&s,&c,"PUBLISH_CONTENT","https://example.com/post","i-r2","k-r2",100)
		case "risk1_not_promoted": refreshM11Intent(&s,&c,"UPDATE_DRAFT","https://example.com/prod","i-r1","k-r1",100); a:=c.TrustedLeaseApprovals[s.Lease.ApprovalRef];a.ValidatedRiskClasses=[]string{"RISK0"};c.TrustedLeaseApprovals[s.Lease.ApprovalRef]=a
		case "risk_not_delegated": refreshM11Intent(&s,&c,"UPDATE_DRAFT","https://example.com/prod","i-r1","k-r1",100); s.Lease.AllowedRiskClasses=[]string{"RISK0"}; refreshLeaseTrust(&s,&c)
		case "scope_action": refreshM11Intent(&s,&c,"UPDATE_PROFILE","https://example.com/profile","i-scope","k-scope",100)
		case "expired_lease": s.Lease.ExpiresAt="2026-09-03T07:30:00Z"; refreshLeaseTrust(&s,&c)
		case "revoked_lease": c.RevokedLeaseIDs=[]string{s.Lease.LeaseID}
		case "kill_switch": c.KillSwitch=true
		case "lease_approval_mismatch": a:=c.TrustedLeaseApprovals[s.Lease.ApprovalRef];a.LeaseHash="sha256:0000000000000000000000000000000000000000000000000000000000000000";c.TrustedLeaseApprovals[s.Lease.ApprovalRef]=a
		case "promotion_review_missing": delete(c.TrustedLeaseApprovals,s.Lease.ApprovalRef)
		case "tampered_lease": s.Lease.MaxExecutionsTotal=99
		case "tampered_health": s.Health.TelemetryComplete=false
		case "health_untrusted": delete(c.TrustedHealthSnapshots,s.Health.SnapshotID)
		case "health_stale": s.Health.ObservedAt="2026-09-03T07:40:00Z";refreshHealthTrust(&s,&c)
		case "telemetry_incomplete": s.Health.TelemetryComplete=false;refreshHealthTrust(&s,&c)
		case "dependency_degraded": s.Health.DependencyState="DEGRADED";refreshHealthTrust(&s,&c)
		case "compliance_alert": s.Health.ComplianceAlertCount=1;refreshHealthTrust(&s,&c)
		case "failure_threshold": s.Health.ConsecutiveFailures=s.Lease.MaxConsecutiveFailures;refreshHealthTrust(&s,&c)
		case "reconciliation": s.Health.ReconciliationRequired=true;refreshHealthTrust(&s,&c)
		case "outcome_stale": s.Health.OldestPendingOutcomeAgeSeconds=s.Lease.MaxOutcomeAgeSeconds+1;refreshHealthTrust(&s,&c)
		case "sticky_stop": s.Ledger.ControlMode="STOPPED";s.Ledger.StopReason="COMPLIANCE_ALERT"
		case "rate_limit": s.Ledger.ExecutionsInWindow=s.Lease.MaxExecutionsPerWindow
		case "total_budget": s.Ledger.ExecutionsTotal=s.Lease.MaxExecutionsTotal
		case "cost_budget": s.Ledger.CostMinorTotal=950
		case "pending_outcome": s.Ledger.PendingOutcomes=s.Lease.MaxPendingOutcomes;s.Ledger.PendingExecutionIDs=[]string{"exec-old"}
		case "stale_policy": s.Policy.Reason="tampered"
		case "wrong_executor": c.Executor.ExecutorID="other";c.AllowedExecutorIDs=[]string{"local_sandbox","other"}
		case "missing_cost_bound": s.CostBound=CanaryCostBound{}
		case "governance_unavailable": c.TrustedLeaseApprovals=nil
		}
		g:=EvaluateProductionGate(s,c);if g.Decision!=x.ExpectedDecision||g.Reason!=x.ExpectedReason{t.Fatalf("expected %s/%s got %s/%s",x.ExpectedDecision,x.ExpectedReason,g.Decision,g.Reason)};if g.ExecutionAuthorized{t.Fatal("ProductionGateDecision must not authorize execution")}
	})}
}

func TestM11AgentCostHintDoesNotOwnBudget(t *testing.T){s,c:=baseM11();s.Intent.Parameters["estimated_cost_minor"]=0;s.Intent=SealShadowActionIntent(s.Intent);pc:=c.PolicyContext;pc.Now="2026-09-03T07:05:00Z";s.Policy=EvaluateShadowPolicy(s.Intent,pc);s.CostBound=SealCanaryCostBound(CanaryCostBound{CostBoundID:"trusted",IntentID:s.Intent.IntentID,IntentHash:s.Intent.IntentHash,MaxCostMinor:100,Currency:"USD",SourceRef:"registry",ObservedAt:"2026-09-03T07:55:00Z",ExpiresAt:"2026-09-03T09:00:00Z",CorrelationID:s.Intent.CorrelationID});c.TrustedCostBounds[s.CostBound.CostBoundID]=s.CostBound.CostBoundHash;s.Ledger.CostMinorTotal=950;g:=EvaluateProductionGate(s,c);if g.Reason!="PRODUCTION_COST_BUDGET_EXHAUSTED"{t.Fatalf("intent hint bypassed trusted budget: %+v",g)}}

func TestM11StopOutranksDegrade(t *testing.T){s,c:=baseM11();s.Health.TelemetryComplete=false;s.Health.ComplianceAlertCount=1;refreshHealthTrust(&s,&c);g:=EvaluateProductionGate(s,c);if g.Decision!="STOP"||g.Reason!="COMPLIANCE_ALERT"{t.Fatalf("severe STOP signal must outrank DEGRADE: %+v",g)}}

func TestM11AuthorizationExpiresAtHealthFreshnessBoundary(t *testing.T){s,c:=baseM11();s.Health.ObservedAt="2026-09-03T07:56:00Z";refreshHealthTrust(&s,&c);a,_,status:=AuthorizeProduction(s,c);if status!="AUTHORIZED"{t.Fatal(status)};if a.ExpiresAt!="2026-09-03T08:01:00Z"{t.Fatalf("authorization must expire at observed_at + health TTL, got %s",a.ExpiresAt)}}

func TestM11PromotionBindsCanaryAndE5Evidence(t *testing.T){s,c:=baseM11();c.KnownCanaryGrantHashes[productionCanaryKey(s.Lease.SourceCanaryGrantID,s.Lease.SourceCanaryGrantVersion)]="sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb";g:=EvaluateProductionGate(s,c);if g.Reason!="PROMOTION_SOURCE_MISMATCH"{t.Fatalf("wrong CanaryGrant provenance accepted: %+v",g)};s,c=baseM11();c.KnownE5Refs=nil;g=EvaluateProductionGate(s,c);if g.Reason!="PROMOTION_E5_EVIDENCE_MISSING"{t.Fatalf("missing E5 evidence accepted: %+v",g)}}

func TestM11ExecutionOutcomeAndBackpressure(t *testing.T){s,c:=baseM11();dir:=t.TempDir();initM11(t,&s,dir,c.Now);a,_,status:=AuthorizeProduction(s,c);if status!="AUTHORIZED"{t.Fatal(status)};s.Authorization=&a;r,status:=ExecuteProductionLocalSandbox(&s,a,c,dir);if status!="EXECUTED"||r.Status!="SUCCEEDED"{t.Fatalf("%+v %s",r,status)};refreshM11Intent(&s,&c,"PREPARE_LOCAL_DRAFT","https://example.com/prod","i2","k2",100);s.Authorization=nil;if g:=EvaluateProductionGate(s,c);g.Decision!="WAIT"||g.Reason!="OUTCOME_BACKPRESSURE"{t.Fatalf("expected backpressure: %+v",g)};if got:=RecordProductionOutcome(&s,dir,"bad","exec-other","2026-09-03T08:10:00Z");got!="DENY_OUTCOME_LINK"{t.Fatal(got)};if got:=RecordProductionOutcome(&s,dir,"out1",r.ExecutionID,"2026-09-03T08:10:00Z");got!="OUTCOME_RECORDED"{t.Fatal(got)};if g:=EvaluateProductionGate(s,c);g.Decision!="ALLOW_PRODUCTION"{t.Fatalf("outcome should release backpressure: %+v",g)}}

func TestM11StickyStopPersistsAcrossRestartAndOutcomeStillCloses(t *testing.T){s,c:=baseM11();dir:=t.TempDir();initM11(t,&s,dir,c.Now);a,_,status:=AuthorizeProduction(s,c);if status!="AUTHORIZED"{t.Fatal(status)};s.Authorization=&a;r,status:=ExecuteProductionLocalSandbox(&s,a,c,dir);if status!="EXECUTED"{t.Fatal(status)};s.Health.ComplianceAlertCount=1;refreshHealthTrust(&s,&c);refreshM11Intent(&s,&c,"PREPARE_LOCAL_DRAFT","https://example.com/prod","i2","k2",100);a2,_,status:=AuthorizeProduction(s,c);if status!="STOP"{t.Fatalf("expected STOP got %s",status)};s.Authorization=&a2;if got:=RecordProductionOutcome(&s,dir,"out-stop",r.ExecutionID,"2026-09-03T08:10:00Z");got!="OUTCOME_RECORDED"{t.Fatalf("STOP must not prevent outcome observation: %s",got)}}

func TestM11ExecutorNeverCreatesMissingProductionLedger(t *testing.T){s,c:=baseM11();a,_,status:=AuthorizeProduction(s,c);if status!="AUTHORIZED"{t.Fatal(status)};s.Authorization=&a;dir:=t.TempDir();if _,status=ExecuteProductionLocalSandbox(&s,a,c,dir);status!="STOP_ACTIVATION_STATE_MISSING"{t.Fatalf("executor must not bootstrap production state: %s",status)};initM11(t,&s,dir,c.Now);if err:=os.Remove(productionLedgerPath(dir,s.Lease.LeaseID,s.Lease.LeaseVersion));err!=nil{t.Fatal(err)};if _,status=ExecuteProductionLocalSandbox(&s,a,c,dir);status!="STOP_LEDGER_MISSING"{t.Fatalf("missing durable ledger must fail closed: %s",status)}}

func TestM11UnknownEffectRequiresHumanResolutionAndNewLease(t *testing.T){s,c:=baseM11();dir:=t.TempDir();initM11(t,&s,dir,c.Now);a,_,status:=AuthorizeProduction(s,c);if status!="AUTHORIZED"{t.Fatal(status)};s.Authorization=&a;if err:=os.WriteFile(sandboxIdempotencyPath(dir,a.IdempotencyKey),[]byte("unknown"),0600);err!=nil{t.Fatal(err)};r,status:=ExecuteProductionLocalSandbox(&s,a,c,dir);if status!="STOP_RECONCILIATION"||r.Status!="RECONCILIATION_REQUIRED"||s.Ledger.ControlMode!="STOPPED"{t.Fatalf("%+v %s %+v",r,status,s.Ledger)};resolution:=ProductionReconciliationResolution{ResolutionID:"rr1",LeaseID:s.Lease.LeaseID,LeaseVersion:s.Lease.LeaseVersion,LeaseHash:s.Lease.LeaseHash,ExecutionID:r.ExecutionID,ResolvedBy:"human",ResolverID:"learner",ResolvedAt:"2026-09-03T08:20:00Z",EffectState:"NOT_PERFORMED",Reason:"provider audit confirmed no side effect"};if got:=ResolveProductionReconciliation(&s,dir,resolution);got!="RECONCILIATION_RESOLVED_STOPPED"{t.Fatal(got)};healthy:=s.Health;healthy.ReconciliationRequired=false;healthy.ComplianceAlertCount=0;healthy.ObservedAt="2026-09-03T08:20:00Z";healthy=SealProductionHealth(healthy);c.TrustedHealthSnapshots[healthy.SnapshotID]=healthy.SnapshotHash;s.Health=healthy;s.Authorization=nil;if g:=EvaluateProductionGate(s,c);g.Decision!="STOP"||g.Reason!="STICKY_STOP"{t.Fatalf("resolution must not resume old lease: %+v",g)}}

func TestM11ClosedLoopRequiresReviewedNonAutoAppliedImprovement(t *testing.T){s,c:=baseM11();dir:=t.TempDir();initM11(t,&s,dir,c.Now);a,g,status:=AuthorizeProduction(s,c);if status!="AUTHORIZED"{t.Fatal(status)};s.Authorization=&a;r,status:=ExecuteProductionLocalSandbox(&s,a,c,dir);if status!="EXECUTED"{t.Fatal(status)};out:=OutcomeRecord{OutcomeID:"out11",ActionID:r.ExecutionID,ObservedAt:"2026-09-03T08:10:00Z",Status:"VALID",Metrics:map[string]float64{"clicks":1},SourceRef:"real-analytics"};ev:=EvaluationRecord{EvaluationID:"eval11",DecisionID:s.Intent.DecisionID,ActionID:r.ExecutionID,OutcomeIDs:[]string{out.OutcomeID},EvaluatedAt:"2026-09-03T08:15:00Z",Result:"SUPPORTED",EvidenceIDs:[]string{"e11"}};p:=ImprovementProposal{ProposalID:"p11",EvaluationIDs:[]string{ev.EvaluationID},CurrentVersion:"v1",ProposedVersion:"v2",ChangeSummary:"narrow retry window",ExpectedBenefit:"reduce uncertainty",Rollback:"restore v1",AutoApply:false};review:=ReviewRecord{ReviewID:"review11",ProposalID:p.ProposalID,ReviewedBy:"human",ReviewedAt:"2026-09-03T08:20:00Z",Decision:"APPROVE_FOR_MANUAL_CHANGE",Reason:"reviewed production evidence"};cycle:=ProductionCycleRecord{CycleID:"cycle11",LeaseID:s.Lease.LeaseID,LeaseVersion:s.Lease.LeaseVersion,LeaseHash:s.Lease.LeaseHash,ObservationIDs:[]string{"e11"},DecisionID:s.Intent.DecisionID,IntentID:s.Intent.IntentID,IntentHash:s.Intent.IntentHash,GateID:g.GateID,AuthorizationID:a.AuthorizationID,ExecutionID:r.ExecutionID,OutcomeID:out.OutcomeID,EvaluationID:ev.EvaluationID,ImprovementProposalID:p.ProposalID,ReviewID:review.ReviewID,Status:"CLOSED",OpenedAt:"2026-09-03T08:00:00Z",ClosedAt:"2026-09-03T08:20:00Z",CorrelationID:s.Intent.CorrelationID};if got:=ValidateProductionClosedCycle(cycle,s,g,a,r,out,ev,&p,&review);got!=missionValid{t.Fatalf("closed loop invalid: %s",got)};p.AutoApply=true;if got:=ValidateProductionClosedCycle(cycle,s,g,a,r,out,ev,&p,&review);got!="REJECT_AUTO_APPLY"{t.Fatalf("self-applying improvement must be rejected: %s",got)}}

func TestM11ActivationRecordDoesNotPermitLedgerRecreation(t *testing.T){s,c:=baseM11();dir:=t.TempDir();initM11(t,&s,dir,c.Now);if err:=os.Remove(productionLedgerPath(dir,s.Lease.LeaseID,s.Lease.LeaseVersion));err!=nil{t.Fatal(err)};if got:=InitializeProductionLedger(&s,dir,c.Now);got!="ALREADY_INITIALIZED"{t.Fatalf("activation record must prevent budget reset: %s",got)}}

func TestM11HealthBoundaryUsesObservedTime(t *testing.T){s,c:=baseM11();s.Health.ObservedAt="2026-09-03T07:55:00Z";refreshHealthTrust(&s,&c);c.Now="2026-09-03T08:00:00Z";g:=EvaluateProductionGate(s,c);if g.Decision!="ALLOW_PRODUCTION"{t.Fatalf("exact freshness boundary should still be valid: %+v",g)};c.Now=time.Date(2026,9,3,8,0,1,0,time.UTC).Format(time.RFC3339);g=EvaluateProductionGate(s,c);if g.Decision!="DEGRADE"||g.Reason!="HEALTH_STALE"{t.Fatalf("snapshot past observed TTL must degrade: %+v",g)}}
