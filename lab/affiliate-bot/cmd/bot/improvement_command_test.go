package main

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestImprovementCommandIdentifiesOperation(t *testing.T) {
	for _, kind := range []string{"proposal", "review"} {
		for _, operation := range []string{"import", "list"} {
			var out, diagnostic bytes.Buffer
			// Deliberately incomplete invocation: diagnostics must identify the requested
			// operation without opening stores or authorizing anything.
			code := runImprovementStore(kind, []string{operation}, &out, &diagnostic)
			var env map[string]any
			if err := json.Unmarshal(out.Bytes(), &env); err != nil {
				t.Fatal(err)
			}
			if code != 2 || env["command"] != kind+" "+operation {
				t.Fatalf("operation label: got %v, want %s %s", env["command"], kind, operation)
			}
			if env["execution_permitted"] != false || env["auto_apply"] != false {
				t.Fatal(env)
			}
		}
	}
}
