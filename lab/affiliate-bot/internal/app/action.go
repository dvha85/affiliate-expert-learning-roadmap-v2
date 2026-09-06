package app

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/dvha85/affiliate-expert-learning-roadmap-v2/contracts"
	"github.com/dvha85/affiliate-expert-learning-roadmap-v2/core/m03"
)

type actionEnvelope struct {
	Command            string              `json:"command"`
	Status             string              `json:"status"`
	Artifact           *m03.M03CheckResult `json:"artifact,omitempty"`
	ExecutionPermitted bool                `json:"execution_permitted"`
}

// RunAction is read-only: no directory creation, store writes or network calls.
// VALID means local pair validation, not evidence provenance or store linkage.
func RunAction(args []string, stdout, stderr io.Writer) int {
	emit := func(status string, artifact *m03.M03CheckResult, code int) int {
		b, e := json.Marshal(actionEnvelope{"action validate", status, artifact, false})
		if e == nil {
			_, e = stdout.Write(append(b, '\n'))
		}
		if e != nil {
			fmt.Fprintln(stderr, "Không thể xuất kết quả JSON:", e)
			return 1
		}
		return code
	}
	if len(args) != 3 || args[0] != "validate" {
		fmt.Fprintln(stderr, "Cách dùng: bot action validate ACTION.json OUTCOME.json")
		return emit("USAGE_ERROR", nil, 2)
	}
	a, e := os.ReadFile(args[1])
	if e != nil {
		fmt.Fprintln(stderr, "Không đọc được action:", e)
		return emit("IO_ERROR", nil, 1)
	}
	o, e := os.ReadFile(args[2])
	if e != nil {
		fmt.Fprintln(stderr, "Không đọc được outcome:", e)
		return emit("IO_ERROR", nil, 1)
	}
	r, status := m03.CheckM03Pair(a, o)
	if status != "VALID" {
		fmt.Fprintln(stderr, "Cặp action/outcome không đạt:", status)
		return emit(status, nil, 1)
	}
	for _, item := range []struct {
		schema string
		value  any
	}{{"action-record.schema.json", r.Action}, {"outcome-record.schema.json", r.Outcome}} {
		b, err := json.Marshal(item.value)
		if err != nil {
			fmt.Fprintln(stderr, err)
			return emit("OUTPUT_INVALID", nil, 1)
		}
		if err = contracts.ValidateRaw(item.schema, b); err != nil {
			fmt.Fprintln(stderr, err)
			return emit("OUTPUT_INVALID", nil, 1)
		}
	}
	return emit("VALID", &r, 0)
}
