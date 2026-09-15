package main

import (
	"testing"

	corem08 "github.com/dvha85/affiliate-expert-learning-roadmap-v2/core/m08"
	corem09 "github.com/dvha85/affiliate-expert-learning-roadmap-v2/core/m09"
)

// Keep the learner persistence envelope tied to the same types that own the
// canonical M08/M09 decoders. These assignments intentionally require type
// identity, rather than merely matching JSON fields: if a local struct is
// reintroduced, this test stops the learner and harness schemas drifting
// apart again.
func TestLearnerMissionStateUsesCanonicalM08M09Types(t *testing.T) {
	state := LearnerMissionState{
		Intent:   &corem08.Intent{},
		Policy:   &corem08.PolicyDecision{},
		Approval: &corem09.ApprovalRecord{},
	}
	if state.Intent == nil || state.Policy == nil || state.Approval == nil {
		t.Fatal("canonical learner schema fields were not retained")
	}
}
