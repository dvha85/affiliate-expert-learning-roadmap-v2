package m06

import (
	"encoding/hex"
	"github.com/dvha85/affiliate-expert-learning-roadmap-v2/contracts"
)

// M06FileInput describes an offline response fixture, not an HTTP fetch API.
type M06FileInput struct {
	SubjectID  string       `json:"subject_id"`
	StatusCode int          `json:"status_code"`
	Request    WatchRequest `json:"request"`
}

func DecodeM06Input(raw []byte) (M06FileInput, string) {
	var input M06FileInput
	value, err := contracts.Decode(raw)
	if err != nil {
		return input, "INVALID_SCHEMA"
	}
	object, ok := value.(map[string]any)
	if !ok {
		return input, "INVALID_SCHEMA"
	}
	for _, key := range []string{"subject_id", "status_code", "request"} {
		if object[key] == nil {
			return input, "INVALID_SCHEMA"
		}
	}
	request, ok := object["request"].(map[string]any)
	if !ok {
		return input, "INVALID_SCHEMA"
	}
	for _, key := range []string{"method", "url", "allow_hosts", "observed_at", "correlation_id", "body"} {
		if request[key] == nil {
			return input, "INVALID_SCHEMA"
		}
	}
	if v, exists := request["previous_hash"]; exists && v == nil {
		return input, "INVALID_SCHEMA"
	}
	if hosts, ok := request["allow_hosts"].([]any); ok {
		for _, host := range hosts {
			if host == nil {
				return input, "INVALID_SCHEMA"
			}
		}
	}
	if contracts.DecodeStrict(raw, &input) != nil {
		return M06FileInput{}, "INVALID_SCHEMA"
	}
	return input, missionValid
}

// This profile supports the normalizer's fields only. The canonical schema
// allows extensions, but this typed projection must not silently discard them.
func ValidateM06Observation(raw []byte) string {
	if contracts.ValidateRaw("observation.schema.json", raw) != nil {
		return "INVALID_SCHEMA"
	}
	var observation CanonicalObservation
	if contracts.DecodeStrict(raw, &observation) != nil {
		return "INVALID_SCHEMA"
	}
	if observation.EvidenceKind != "synthetic" || observation.UseContext != "test" || len(observation.ContentHash) != 64 {
		return "INVALID_PROFILE"
	}
	if _, err := hex.DecodeString(observation.ContentHash); err != nil {
		return "INVALID_PROFILE"
	}
	return missionValid
}
