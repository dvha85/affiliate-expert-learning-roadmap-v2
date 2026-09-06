package app

import (
	"encoding/json"
	"fmt"
	"github.com/dvha85/affiliate-expert-learning-roadmap-v2/core/m00"
	"io"
	"os"
)

// RunEvidence imports into an output artifact only; persistence remains an
// explicit history capture operation. Never redirect stdout over the input.
func RunEvidence(args []string, stdout, stderr io.Writer) int {
	emit := func(status string, artifact any, code int) int {
		envelope := map[string]any{"command": "evidence import", "status": status, "execution_permitted": false}
		if artifact != nil {
			envelope["artifact"] = artifact
		}
		if err := json.NewEncoder(stdout).Encode(envelope); err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		return code
	}
	if len(args) != 2 || args[0] != "import" {
		fmt.Fprintln(stderr, "Cách dùng: bot evidence import PACKET.json")
		return emit("USAGE_ERROR", nil, 2)
	}
	raw, err := os.ReadFile(args[1])
	if err != nil {
		fmt.Fprintln(stderr, err)
		return emit("IO_ERROR", nil, 1)
	}
	observations, err := m00.Convert(raw)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return emit("INVALID_INPUT", nil, 1)
	}
	return emit("VALID", observations, 0)
}
