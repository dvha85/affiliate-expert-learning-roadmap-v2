package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestAdvisorFixtureBundle(t *testing.T) {
	parent := t.TempDir()
	seen := map[string]bool{}
	for i := 0; i < 2; i++ {
		var output bytes.Buffer
		if code := runAdvisor([]string{"fixture-run", parent}, &output, &output); code != 0 {
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

func TestAdvisorFixtureBundleRejectsSymlinkOutputParent(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink permissions are not portable on Windows")
	}
	outside := t.TempDir()
	link := filepath.Join(t.TempDir(), "fixture-output-parent")
	if err := os.Symlink(outside, link); err != nil {
		t.Fatal(err)
	}
	var output bytes.Buffer
	var stderr bytes.Buffer
	if code := runAdvisor([]string{"fixture-run", link}, &output, &stderr); code == 0 {
		t.Fatalf("fixture-run accepted symlink output parent: %s", output.String())
	}
	var envelope struct {
		Status string `json:"status"`
	}
	if err := json.Unmarshal(output.Bytes(), &envelope); err != nil || envelope.Status != "PATH_ERROR" {
		t.Fatalf("fixture-run did not reject symlink parent before writing: envelope=%s err=%v", output.String(), err)
	}
	entries, err := os.ReadDir(outside)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Fatalf("fixture-run created a bundle through a rejected symlink parent: %+v", entries)
	}
}
