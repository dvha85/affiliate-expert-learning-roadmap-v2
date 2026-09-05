// Package learning is an offline exercise, not a platform adapter or OutcomeRecord.
package learning

import (
	"fmt"
	"os"
	"strings"

	"github.com/dvha85/affiliate-expert-learning-roadmap-v2/contracts"
)

type Report struct {
	SourceRef   string `json:"source_ref"`
	Status      string `json:"status"`
	ValidOrders *int   `json:"valid_orders"`
}

func NormalizeStatus(value string) (string, error) {
	status := strings.ToUpper(strings.TrimSpace(value))
	switch status {
	case "PENDING", "VALID", "CANCELLED", "REFUNDED", "PAID":
		return status, nil
	default:
		return "", fmt.Errorf("unknown fixture status %q", value)
	}
}

func ReadReport(path string) (Report, error) {
	var report Report
	raw, err := os.ReadFile(path)
	if err != nil {
		return report, fmt.Errorf("read report: %w", err)
	}
	if err := contracts.DecodeStrict(raw, &report); err != nil {
		return report, fmt.Errorf("decode report: %w", err)
	}
	if strings.TrimSpace(report.SourceRef) == "" {
		return report, fmt.Errorf("source_ref is required")
	}
	report.Status, err = NormalizeStatus(report.Status)
	if err != nil {
		return Report{}, err
	}
	if report.ValidOrders != nil && *report.ValidOrders < 0 {
		return Report{}, fmt.Errorf("valid_orders cannot be negative")
	}
	return report, nil
}

// Missing/null is unknown, not an observed zero. Omit unknown metrics entirely.
func Metrics(report Report) map[string]int {
	metrics := map[string]int{}
	if report.ValidOrders != nil {
		metrics["valid_orders"] = *report.ValidOrders
	}
	return metrics
}
