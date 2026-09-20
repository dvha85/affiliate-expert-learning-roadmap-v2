package canonical

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/dvha85/affiliate-expert-learning-roadmap-v2/core/m00"
	corem07 "github.com/dvha85/affiliate-expert-learning-roadmap-v2/core/m07"
)

// ContextVersion identifies the stable envelope shared by learner M07/M08
// adapters. The envelope is evidence context only; it grants no execution
// authority.
const ContextVersion = "canonical-evidence-context/v1"

// ObservationInput is the canonical projection boundary. Raw is the exact
// serialized learner observation so M00 can re-derive original field IDs and
// provenance instead of accepting caller-supplied field identifiers.
type ObservationInput struct {
	Raw                   json.RawMessage
	ObservationID         string
	SubjectID             string
	ClaimKind             string
	SourceAuthorityOrRole string
	Limitation            string
	Value                 any
}

// Context is consumed by M07 grounding and M08 proposal/policy adapters.
// EvidenceIDs starts with aggregate observation IDs and then contains the
// original M00 field IDs in deterministic observation/field order.
type Context struct {
	Version     string           `json:"version"`
	RecordID    string           `json:"record_id"`
	DecisionID  string           `json:"decision_id"`
	EvidenceIDs []string         `json:"evidence_ids"`
	Evidence    []corem07.Evidence `json:"evidence"`
	Authority   string           `json:"authority"`
}

// BuildEvidenceContext creates one shared evidence envelope for every
// downstream adapter. It fails closed when the recorded aggregate IDs do not
// exactly match the observations, when a raw source projection is malformed,
// or when an original field ID collides with another evidence ID.
func BuildEvidenceContext(recordID, decisionID string, aggregateIDs []string, observations []ObservationInput) (Context, error) {
	if strings.TrimSpace(recordID) == "" || strings.TrimSpace(decisionID) == "" {
		return Context{}, fmt.Errorf("record_id and decision_id are required")
	}
	if len(aggregateIDs) != len(observations) {
		return Context{}, fmt.Errorf("aggregate evidence_ids must match observations")
	}

	ctx := Context{
		Version:    ContextVersion,
		RecordID:   recordID,
		DecisionID: decisionID,
		Authority:  "canonical_history_store",
	}
	seen := map[string]bool{}
	for i, observation := range observations {
		if strings.TrimSpace(observation.ObservationID) == "" || strings.TrimSpace(observation.SubjectID) == "" {
			return Context{}, fmt.Errorf("observation identity is required")
		}
		if aggregateIDs[i] != observation.ObservationID || seen[observation.ObservationID] {
			return Context{}, fmt.Errorf("aggregate evidence_ids do not match canonical observations")
		}
		seen[observation.ObservationID] = true
		ctx.EvidenceIDs = append(ctx.EvidenceIDs, observation.ObservationID)
		ctx.Evidence = append(ctx.Evidence, corem07.Evidence{
			EvidenceID: observation.ObservationID, SubjectID: observation.SubjectID,
			FieldOrClaim: "product snapshot", Value: observation.Value,
			ClaimKind: observation.ClaimKind, SourceAuthorityOrRole: observation.SourceAuthorityOrRole,
			Limitation: observation.Limitation,
		})

		fields, err := m00.SourceFields(observation.Raw)
		if err != nil {
			return Context{}, fmt.Errorf("source fields for %s: %w", observation.ObservationID, err)
		}
		for _, field := range fields {
			if strings.TrimSpace(field.ObservationID) == "" || seen[field.ObservationID] {
				return Context{}, fmt.Errorf("field evidence id collides or is empty: %s", field.ObservationID)
			}
			seen[field.ObservationID] = true
			ctx.EvidenceIDs = append(ctx.EvidenceIDs, field.ObservationID)
			ctx.Evidence = append(ctx.Evidence, corem07.Evidence{
				EvidenceID: field.ObservationID, SubjectID: field.SubjectID,
				FieldOrClaim: field.Field, Value: field.Value, ClaimKind: field.ClaimKind,
				SourceAuthorityOrRole: field.Role, Limitation: field.Limitation,
			})
		}
	}
	return ctx, nil
}
