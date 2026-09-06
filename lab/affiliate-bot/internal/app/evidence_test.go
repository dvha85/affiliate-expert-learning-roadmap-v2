package app

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestEvidenceBoundary(t *testing.T) {
	raw, e := os.ReadFile("../../../../examples/m00-import/packet-t1.json")
	if e != nil {
		t.Fatal(e)
	}
	path := filepath.Join(t.TempDir(), "packet.json")
	if e = os.WriteFile(path, raw, 0600); e != nil {
		t.Fatal(e)
	}
	for _, tc := range []struct {
		args   []string
		code   int
		status string
	}{{[]string{"import", path}, 0, "VALID"}, {nil, 2, "USAGE_ERROR"}, {[]string{"import", path + "absent"}, 1, "IO_ERROR"}} {
		var out, errout bytes.Buffer
		if code := RunEvidence(tc.args, &out, &errout); code != tc.code {
			t.Fatal(code, errout.String())
		}
		var envelope map[string]any
		if e = json.Unmarshal(out.Bytes(), &envelope); e != nil {
			t.Fatal(e)
		}
		if envelope["status"] != tc.status || envelope["execution_permitted"] != false {
			t.Fatal(envelope)
		}
		if tc.code != 0 && envelope["artifact"] != nil {
			t.Fatal(envelope)
		}
	}
	after, _ := os.ReadFile(path)
	if !bytes.Equal(raw, after) {
		t.Fatal("input changed")
	}
	if e = os.WriteFile(path, []byte(`{"unexpected":true}`), 0600); e != nil {
		t.Fatal(e)
	}
	var out, errout bytes.Buffer
	if RunEvidence([]string{"import", path}, &out, &errout) != 1 || bytes.Contains(out.Bytes(), []byte(`"artifact"`)) {
		t.Fatal(out.String())
	}
}
