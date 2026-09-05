package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dvha85/affiliate-expert-learning-roadmap-v2/contracts"
)

func m06Request() WatchRequest {
	return WatchRequest{Method: "GET", URL: "https://example.com/offer", AllowHosts: []string{"example.com"}, ObservedAt: "2026-09-03T01:00:00Z", CorrelationID: "syn-c", Body: "price=100"}
}

func TestM06CommittedFixture(t *testing.T) {
	var output bytes.Buffer
	if err := runM06Check(&output, []string{"../../testdata/m06-response.json"}); err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(output.Bytes(), []byte(`"evidence_kind":"synthetic"`)) || !bytes.Contains(output.Bytes(), []byte(`"external_side_effects":false`)) {
		t.Fatal("fixture provenance/authority changed")
	}
}

func TestM06NormalizerSemantics(t *testing.T) {
	r := m06Request()
	base, _ := NormalizeWatchObservation(r, "syn-product")
	r.PreviousHash = contentHash(r.Body)
	same, state := NormalizeWatchObservation(r, "syn-product")
	if state != "UNCHANGED" || same.ObservationID != base.ObservationID {
		t.Fatal(state)
	}
	r.Body = "changed"
	changed, state := NormalizeWatchObservation(r, "syn-product")
	if state != "CHANGED" || changed.ObservationID == base.ObservationID {
		t.Fatal(state)
	}
	r = m06Request()
	r.ObservedAt = "2026-09-03T08:00:00+07:00"
	same, _ = NormalizeWatchObservation(r, "syn-product")
	if same.ObservationID != base.ObservationID {
		t.Fatal("equivalent instant changed ID")
	}
	for _, kind := range []string{"HEAD", "empty"} {
		r = m06Request()
		if kind == "HEAD" {
			r.Method = " head "
		} else {
			r.Body = ""
		}
		o, state := NormalizeWatchObservation(r, "syn-product")
		if state != "NEW" || o.State != "missing" || o.ClaimKind != "unknown" {
			t.Fatal(kind, state, o)
		}
	}
	for _, field := range []string{"subject", "time", "host", "write"} {
		r = m06Request()
		subject := "syn-product"
		switch field {
		case "subject":
			subject = " "
		case "time":
			r.ObservedAt = "bad"
		case "host":
			r.URL = "https://other.invalid"
		case "write":
			r.Method = "POST"
		}
		o, state := NormalizeWatchObservation(r, subject)
		if state == "NEW" || o.ObservationID != "" {
			t.Fatal(field, state)
		}
	}
}

func TestM06OutputSchema(t *testing.T) {
	o, _ := NormalizeWatchObservation(m06Request(), "syn-product")
	raw, err := json.Marshal(o)
	if err != nil {
		t.Fatal(err)
	}
	var original map[string]json.RawMessage
	if err := json.Unmarshal(raw, &original); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"observation_id", "subject_id", "observed_at", "access_method", "evidence_kind", "claim_kind", "state", "limitation"} {
		for _, mutation := range []string{"missing", "null"} {
			fields := map[string]json.RawMessage{}
			for k, v := range original {
				fields[k] = v
			}
			if mutation == "missing" {
				delete(fields, key)
			} else {
				fields[key] = json.RawMessage(`null`)
			}
			bad, err := json.Marshal(fields)
			if err != nil {
				t.Fatal(err)
			}
			if state := ValidateM06Observation(bad); state == missionValid {
				t.Fatal(key, mutation)
			}
		}
	}
	extension := strings.Replace(string(raw), `{`, `{"unsupported_extension":1,`, 1)
	if err := contracts.ValidateRaw("observation.schema.json", []byte(extension)); err != nil {
		t.Fatal("schema extensions unexpectedly forbidden", err)
	}
	if ValidateM06Observation([]byte(extension)) == missionValid {
		t.Fatal("projection silently dropped extension")
	}
	for _, bad := range []string{string(raw) + ` {}`, strings.Replace(string(raw), `"subject_id":`, `"subject_id":"other","subject_id":`, 1), strings.Replace(string(raw), `"synthetic"`, `"real"`, 1), strings.Replace(string(raw), o.ContentHash, "bad-hash", 1)} {
		if ValidateM06Observation([]byte(bad)) == missionValid {
			t.Fatal("bad output accepted")
		}
	}
}

func m06InputJSON(t *testing.T) []byte {
	t.Helper()
	raw, err := json.Marshal(M06FileInput{SubjectID: "syn-product", StatusCode: 200, Request: m06Request()})
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func TestM06InputAndCLI(t *testing.T) {
	raw := m06InputJSON(t)
	for _, bad := range []string{`{}`, `null`, strings.Replace(string(raw), `"body":"price=100"`, `"body":null`, 1), strings.Replace(string(raw), `"status_code":200`, `"status_code":null`, 1), strings.Replace(string(raw), `{`, `{"extra":true,`, 1)} {
		if _, s := DecodeM06Input([]byte(bad)); s != "INVALID_SCHEMA" {
			t.Fatal(s)
		}
	}
	dir := t.TempDir()
	path := filepath.Join(dir, "input.json")
	if err := os.WriteFile(path, raw, 0600); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	if err := runM06Check(&out, []string{path}); err != nil {
		t.Fatal(err)
	}
	var envelope struct {
		Observation json.RawMessage `json:"observation"`
		Result      string          `json:"result"`
		External    bool            `json:"external_side_effects"`
	}
	if err := contracts.DecodeStrict(out.Bytes(), &envelope); err != nil {
		t.Fatal(err)
	}
	if envelope.Result != "NEW" || envelope.External || ValidateM06Observation(envelope.Observation) != missionValid {
		t.Fatal("invalid output")
	}
	if err := runM06Check(m03BrokenWriter{}, []string{path}); err == nil {
		t.Fatal("writer error ignored")
	}
	for _, tc := range []struct{ old, new, want string }{{`200`, `500`, `REJECT_RESPONSE_STATUS`}, {`200`, `302`, `REJECT_RESPONSE_STATUS`}, {`"GET"`, `"POST"`, `REJECT_WRITE_METHOD`}, {`"body":"price=100"`, `"body":null`, `INVALID_SCHEMA`}} {
		bad := strings.Replace(string(raw), tc.old, tc.new, 1)
		if err := os.WriteFile(path, []byte(bad), 0600); err != nil {
			t.Fatal(err)
		}
		out.Reset()
		err := runM06Check(&out, []string{path})
		if err == nil || !strings.Contains(err.Error(), tc.want) || out.Len() != 0 {
			t.Fatal(tc.want, err)
		}
		got, err := os.ReadFile(path)
		if err != nil || string(got) != bad {
			t.Fatal("input changed")
		}
	}
	for _, args := range [][]string{nil, {"missing"}, {path, "extra"}} {
		out.Reset()
		if err := runM06Check(&out, args); err == nil || out.Len() != 0 {
			t.Fatal("bad invocation accepted")
		}
	}
	entries, err := os.ReadDir(dir)
	if err != nil || len(entries) != 1 {
		t.Fatal("unexpected persist")
	}
}

func TestM06NormalizerEvalPack(t *testing.T) {
	type testCase struct {
		ID       string       `json:"case_id"`
		Request  WatchRequest `json:"request"`
		Expected string       `json:"expected"`
	}
	for _, c := range loadCases[testCase](t, "M06-readonly-watcher") {
		t.Run(c.ID, func(t *testing.T) {
			o, state := NormalizeWatchObservation(c.Request, "syn-product")
			if state != c.Expected {
				t.Fatal(state, c.Expected)
			}
			if state == "NEW" || state == "UNCHANGED" || state == "CHANGED" {
				raw, err := json.Marshal(o)
				if err != nil {
					t.Fatal(err)
				}
				if ValidateM06Observation(raw) != missionValid {
					t.Fatal("output invalid")
				}
			}
		})
	}
}

func TestM06OfflineProvenance(t *testing.T) {
	r := m06Request()
	o, state := NormalizeWatchObservation(r, "syn-product")
	if state != "NEW" {
		t.Fatal(state)
	}
	if o.EvidenceKind != "synthetic" {
		t.Fatal("offline body claimed real evidence")
	}
	raw, err := json.Marshal(o)
	if err != nil {
		t.Fatal(err)
	}
	if state := ValidateM06Observation(raw); state != missionValid {
		t.Fatal(state)
	}
}

func TestM06IdentityBinding(t *testing.T) {
	r := m06Request()
	base, _ := NormalizeWatchObservation(r, "syn-product")
	for _, field := range []string{"url", "time", "method"} {
		changed := r
		switch field {
		case "url":
			changed.URL = "https://example.com/other"
		case "time":
			changed.ObservedAt = "2026-09-04T01:00:00Z"
		case "method":
			changed.Method = "HEAD"
		}
		o, _ := NormalizeWatchObservation(changed, "syn-product")
		if o.ObservationID == base.ObservationID {
			t.Errorf("%s not bound to identity", field)
		}
	}
}
