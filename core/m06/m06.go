package m06

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/url"
	"strings"
	"time"
)

const missionValid = "VALID"
const missionInvalid = "INVALID"

type WatchRequest struct {
	Method        string   `json:"method"`
	URL           string   `json:"url"`
	AllowHosts    []string `json:"allow_hosts"`
	ObservedAt    string   `json:"observed_at"`
	CorrelationID string   `json:"correlation_id"`
	Body          string   `json:"body"`
	PreviousHash  string   `json:"previous_hash"`
}

func ContentHash(body string) string {
	s := sha256.Sum256([]byte(body))
	return hex.EncodeToString(s[:])
}
func EvaluateWatchRequest(r WatchRequest) string {
	m := strings.ToUpper(strings.TrimSpace(r.Method))
	if m != "GET" && m != "HEAD" {
		return "REJECT_WRITE_METHOD"
	}
	u, e := url.Parse(r.URL)
	if e != nil || u.Scheme != "https" || u.Hostname() == "" {
		return "REJECT_SOURCE"
	}
	allowed := false
	for _, h := range r.AllowHosts {
		if strings.EqualFold(strings.TrimSpace(h), u.Hostname()) {
			allowed = true
		}
	}
	if !allowed {
		return "REJECT_SOURCE"
	}
	if _, e := time.Parse(time.RFC3339, r.ObservedAt); e != nil || strings.TrimSpace(r.CorrelationID) == "" {
		return missionInvalid
	}
	current := ContentHash(r.Body)
	if r.PreviousHash == "" {
		return "NEW"
	}
	if r.PreviousHash == current {
		return "UNCHANGED"
	}
	return "CHANGED"
}

type CanonicalObservation struct {
	ObservationID string `json:"observation_id"`
	SubjectID     string `json:"subject_id"`
	SourceURL     string `json:"source_url"`
	ObservedAt    string `json:"observed_at"`
	AccessMethod  string `json:"access_method"`
	EvidenceKind  string `json:"evidence_kind"`
	UseContext    string `json:"use_context"`
	ClaimKind     string `json:"claim_kind"`
	State         string `json:"state"`
	Limitation    string `json:"limitation"`
	CorrelationID string `json:"correlation_id"`
	ContentHash   string `json:"content_hash"`
}

func NormalizeWatchObservation(r WatchRequest, subjectID string) (CanonicalObservation, string) {
	state := EvaluateWatchRequest(r)
	if state == "REJECT_WRITE_METHOD" || state == "REJECT_SOURCE" || state == missionInvalid {
		return CanonicalObservation{}, state
	}
	if strings.TrimSpace(subjectID) == "" {
		return CanonicalObservation{}, missionInvalid
	}
	hash := ContentHash(r.Body)
	method := strings.ToUpper(strings.TrimSpace(r.Method))
	at, _ := time.Parse(time.RFC3339, r.ObservedAt)
	identity, _ := json.Marshal([]string{subjectID, r.URL, at.UTC().Format(time.RFC3339Nano), method, r.CorrelationID, hash})
	observationID := "obs-" + ContentHash(string(identity))
	o := CanonicalObservation{ObservationID: observationID, SubjectID: subjectID, SourceURL: r.URL, ObservedAt: r.ObservedAt, AccessMethod: method, EvidenceKind: "synthetic", UseContext: "test", ClaimKind: "fact", State: "observed", Limitation: "offline supplied response fixture; no network fetch or business truth verified; content hash only", CorrelationID: r.CorrelationID, ContentHash: hash}
	if method == "HEAD" || strings.TrimSpace(r.Body) == "" {
		o.ClaimKind = "unknown"
		o.State = "missing"
	}
	raw, err := json.Marshal(o)
	if err != nil || ValidateM06Observation(raw) != missionValid {
		return CanonicalObservation{}, "INVALID_SCHEMA"
	}
	return o, state
}
