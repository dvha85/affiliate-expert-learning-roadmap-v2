package learning

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func reportFile(t *testing.T, raw string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "report.json")
	if err := os.WriteFile(path, []byte(raw), 0600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestNullIsNotZero(t *testing.T) {
	cases := []struct {
		name, raw string
		observed  bool
		value     int
	}{
		{"null", `{"source_ref":"synthetic","status":"pending","valid_orders":null}`, false, 0},
		{"absent", `{"source_ref":"synthetic","status":"pending"}`, false, 0},
		{"zero", `{"source_ref":"synthetic","status":"pending","valid_orders":0}`, true, 0},
		{"one", `{"source_ref":"synthetic","status":"valid","valid_orders":1}`, true, 1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			report, err := ReadReport(reportFile(t, tc.raw))
			if err != nil {
				t.Fatal(err)
			}
			value, observed := Metrics(report)["valid_orders"]
			if value != tc.value || observed != tc.observed {
				t.Fatalf("want value=%d observed=%v; got value=%d observed=%v", tc.value, tc.observed, value, observed)
			}
		})
	}
}

func TestReportRejectsBadInput(t *testing.T) {
	for _, raw := range []string{
		`{"source_ref":"synthetic","status":"pending","valid_orders":"0"}`,
		`{"source_ref":"synthetic","status":"pending","valid_orders":-1}`,
		`{"source_ref":"synthetic","status":"pending","valid_orders":1.5}`,
		`{"source_ref":"synthetic","status":"pending","clicks":0}`,
		`{"source_ref":"synthetic","status":"success"}`,
		`{"source_ref":"","status":"pending"}`,
		`{"source_ref":"synthetic","status":"pending","Status":"paid"}`,
		`{"source_ref":"synthetic","status":"pending","status":"paid"}`,
		`{"source_ref":"synthetic","status":"pending"} {}`,
	} {
		if _, err := ReadReport(reportFile(t, raw)); err == nil {
			t.Fatalf("bad input accepted: %s", raw)
		}
	}
	if _, err := ReadReport(filepath.Join(t.TempDir(), "missing.json")); err == nil || !strings.Contains(err.Error(), "read report") {
		t.Fatal("file error must be returned")
	}
}

func TestNormalizeStatus(t *testing.T) {
	for _, value := range []string{"pending", " PENDING ", "Pending"} {
		got, err := NormalizeStatus(value)
		if err != nil || got != "PENDING" {
			t.Fatalf("normalize %q: %q %v", value, got, err)
		}
	}
	if _, err := NormalizeStatus("success"); err == nil {
		t.Fatal("unknown status became success")
	}
}

func TestCommittedReportFixture(t *testing.T) {
	report, err := ReadReport("testdata/report.json")
	if err != nil {
		t.Fatal(err)
	}
	value, observed := Metrics(report)["valid_orders"]
	if report.Status != "PENDING" || !observed || value != 0 {
		t.Fatalf("fixture must show observed zero without window conclusion: %+v", report)
	}
}
