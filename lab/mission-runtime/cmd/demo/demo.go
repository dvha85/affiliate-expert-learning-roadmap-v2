package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

type missionDemoEnvelope struct { Mission string `json:"mission"`; Result any `json:"result"` }
func runMissionDemo(args []string) error {
	if len(args)!=1{return errors.New("usage: mission-demo O00|M03|M04|M05|M06|M07|M08|M09|M10")}; id:=strings.ToUpper(args[0]); var result any
	switch id{
	case "O00": result=RunSyntheticWalkthrough()
	case "M03": result=ValidateHumanActionRecord(HumanActionRecord{ActionID:"a1",DecisionID:"d1",ActionType:"PUBLISH_REVIEW",Target:"https://example.com/review",PerformedBy:"human",PerformedAt:"2026-09-03T01:00:00Z",MeasurementWindowEnd:"2026-09-10T01:00:00Z",ComplianceReviewed:true})
	case "M04": result=EvaluateAdvisorOutput(AdvisorOutput{State:"ADVISE",Reason:"weak ranking",EvidenceIDs:[]string{"e1"}},[]AdvisorEvidence{{EvidenceID:"e1",ObservedAt:"2026-09-03T00:00:00Z",SourceRef:"public"}},"2026-09-03T01:00:00Z",24)
	case "M05": result=EvaluateImprovementProposal(ImprovementProposal{ProposalID:"p1",EvaluationIDs:[]string{"ev1"},CurrentVersion:"v1",ProposedVersion:"v2",ChangeSummary:"add evidence",ExpectedBenefit:"reduce error",Rollback:"restore v1"})
	case "M06": result=EvaluateWatchRequest(WatchRequest{Method:"GET",URL:"https://example.com/offer",AllowHosts:[]string{"example.com"},ObservedAt:"2026-09-03T01:00:00Z",CorrelationID:"c1",Body:"price=100"})
	case "M07": result=EvaluateAgentProposal(AgentProposal{State:"PROPOSE",EvidenceIDs:[]string{"e1"},ToolCalls:[]AgentToolCall{{ToolName:"public_http",Method:"GET",Target:"https://example.com/offer"}}},[]ToolSpec{{Name:"public_http",ReadOnly:true,AllowedMethods:[]string{"GET"},AllowedHosts:[]string{"example.com"}}},[]string{"e1"})
	case "M08": result=demoM08Decision()
	case "M09": r,e:=demoM09();if e!=nil{return e};result=r
	case "M10": r,e:=demoM10();if e!=nil{return e};result=r
	default:return fmt.Errorf("unknown mission %s",id)}
	b,e:=json.MarshalIndent(missionDemoEnvelope{id,result},"","  ");if e!=nil{return e};fmt.Println(string(b));return nil
}
