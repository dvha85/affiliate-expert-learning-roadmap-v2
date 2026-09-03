package main

import "encoding/json"

// UnmarshalJSON restores internal compatibility aliases from the canonical
// proposal-only ActionIntent representation after persisted state is reloaded.
func (i *ShadowActionIntent) UnmarshalJSON(data []byte) error {
	type wire ShadowActionIntent
	var decoded wire
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}
	*i = ShadowActionIntent(decoded)
	i.ShadowOnly = i.IntentMode == "PROPOSAL_ONLY" && !i.ExecutionAuthorized
	i.DryRun = i.ShadowOnly
	return nil
}

// UnmarshalJSON restores internal compatibility aliases used by M09-M11
// revalidation without re-introducing obsolete fields into canonical JSON.
func (p *ShadowPolicyDecision) UnmarshalJSON(data []byte) error {
	type wire ShadowPolicyDecision
	var decoded wire
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}
	*p = ShadowPolicyDecision(decoded)
	p.ShadowOnly = p.PolicyMode == "NON_AUTHORIZING" && !p.ExecutionAuthorized
	p.ApprovalRequired = p.PolicyReviewRequired
	return nil
}
