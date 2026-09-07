package main

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestAdvisorFixtureBundle(t *testing.T) {
	parent := t.TempDir()
	seen := map[string]bool{}
	for i := 0; i < 2; i++ {
		var output bytes.Buffer
		if code := runAdvisor([]string{"fixture-run", parent}, &output, io.Discard); code != 0 {
			t.Fatal(code, output.String())
		}
		var envelope struct {
			Status string
			Path   string `json:"bundle_path"`
		}
		if err := json.Unmarshal(output.Bytes(), &envelope); err != nil {
			t.Fatal(err)
		}
		if envelope.Status != "SUPPORTED" || seen[envelope.Path] {
			t.Fatal(envelope)
		}
		seen[envelope.Path] = true
		raw, err := os.ReadFile(filepath.Join(envelope.Path, "advisor.json"))
		if err != nil {
			t.Fatal(err)
		}
		var bundle struct {
			Context   advisorContext
			Execution bool `json:"execution_permitted"`
		}
		if err = json.Unmarshal(raw, &bundle); err != nil || bundle.Execution {
			t.Fatal(err)
		}
		for _, id := range []string{"br11-decision", "br11-observation", "br11-action", "br11-outcome"} {
			if _, ok := bundle.Context.Payload[id]; !ok {
				t.Fatal(id)
			}
		}
		c, err := buildAdvisorContext(filepath.Join(envelope.Path, "history.jsonl"), filepath.Join(envelope.Path, "actions.jsonl"), filepath.Join(envelope.Path, "outcomes.jsonl"), bundle.Context.Config)
		if err != nil {
			t.Fatal(err)
		}
		if len(c.Evidence) != 4 {
			t.Fatal(c.Evidence)
		}
	}
}
