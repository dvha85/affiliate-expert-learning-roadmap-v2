package main

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

func TestActionBinary(t *testing.T) {
	dir := t.TempDir()
	binary := filepath.Join(dir, "bot")
	if runtime.GOOS == "windows" {
		binary += ".exe"
	}
	build := exec.Command("go", "build", "-o", binary, ".")
	if b, e := build.CombinedOutput(); e != nil {
		t.Fatalf("build: %v %s", e, b)
	}
	a := filepath.Join(dir, "action.json")
	o := filepath.Join(dir, "outcome.json")
	for path, source := range map[string]string{a: "../../../mission-runtime/testdata/m03-action.json", o: "../../../mission-runtime/testdata/m03-outcome.json"} {
		b, e := os.ReadFile(source)
		if e != nil {
			t.Fatal(e)
		}
		if e = os.WriteFile(path, b, 0600); e != nil {
			t.Fatal(e)
		}
	}
	beforeA, _ := os.ReadFile(a)
	beforeO, _ := os.ReadFile(o)
	for _, tc := range []struct {
		name   string
		args   []string
		code   int
		status string
	}{
		{"valid", []string{"action", "validate", a, o}, 0, "VALID"},
		{"usage", []string{"action", "validate"}, 2, "USAGE_ERROR"},
		{"missing", []string{"action", "validate", a, filepath.Join(dir, "missing.json")}, 1, "IO_ERROR"},
		{"invalid", []string{"action", "validate", o, a}, 1, "INVALID_SCHEMA"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cmd := exec.Command(binary, tc.args...)
			var out, errout bytes.Buffer
			cmd.Stdout = &out
			cmd.Stderr = &errout
			e := cmd.Run()
			code := 0
			if e != nil {
				if x, ok := e.(*exec.ExitError); ok {
					code = x.ExitCode()
				} else {
					t.Fatal(e)
				}
			}
			if code != tc.code {
				t.Fatal(code, errout.String())
			}
			var r struct {
				Status             string          `json:"status"`
				Artifact           json.RawMessage `json:"artifact"`
				ExecutionPermitted bool            `json:"execution_permitted"`
			}
			if e = json.Unmarshal(out.Bytes(), &r); e != nil {
				t.Fatal(e, out.String())
			}
			if r.Status != tc.status || r.ExecutionPermitted || (len(r.Artifact) > 0) != (code == 0) {
				t.Fatal(out.String())
			}
			if (errout.Len() == 0) != (code == 0) {
				t.Fatal(errout.String())
			}
		})
	}
	afterA, _ := os.ReadFile(a)
	afterO, _ := os.ReadFile(o)
	if !bytes.Equal(beforeA, afterA) || !bytes.Equal(beforeO, afterO) {
		t.Fatal("input mutated")
	}
	entries, _ := os.ReadDir(dir)
	if len(entries) != 3 {
		t.Fatal("unexpected store write")
	}
	// Existing file-input entrypoint and history dispatch must retain their behavior.
	for _, args := range [][]string{{"../../data/sample-observations.json"}, {"history", "list", "../../data/absent-history-for-cli-test.json"}} {
		cmd := exec.Command(binary, args...)
		out, e := cmd.CombinedOutput()
		if args[0] == "history" {
			if e == nil {
				t.Fatal("missing history unexpectedly succeeded")
			}
		} else if e != nil || !bytes.Contains(out, []byte("deterministic baseline")) {
			t.Fatal(e, string(out))
		}
	}
}
