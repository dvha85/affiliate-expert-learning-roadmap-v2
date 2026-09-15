//go:build linux || darwin

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func canaryTestCampaign(t *testing.T) string {
	t.Helper()
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if err := initializeFixedCampaign(root); err != nil {
		t.Fatal(err)
	}
	return filepath.Join(root, "affiliate-expert-learning-roadmap-v2", "deepseek-br11-v1")
}

func TestCanaryPreflightNoReservation(t *testing.T) {
	for _, kind := range []string{"missing-key", "newline-key", "locked", "corrupt", "existing"} {
		t.Run(kind, func(t *testing.T) {
			path := canaryTestCampaign(t)
			t.Setenv("DEEPSEEK_API_KEY", "fixture-key")
			want := ""
			switch kind {
			case "missing-key":
				t.Setenv("DEEPSEEK_API_KEY", "")
				want = "CONFIG_ERROR"
			case "newline-key":
				t.Setenv("DEEPSEEK_API_KEY", "key\n")
				want = "CONFIG_ERROR"
			case "locked":
				if err := os.Mkdir(filepath.Join(filepath.Dir(path), "deepseek-br11-canary.lock"), 0700); err != nil {
					t.Fatal(err)
				}
				want = "LOCKED"
			case "corrupt":
				if err := os.WriteFile(filepath.Join(path, "manifest"), []byte("broken"), 0600); err != nil {
					t.Fatal(err)
				}
				want = "REPORT_ERROR"
			case "existing":
				if _, err := reserveAdvisorAttempt(path); err != nil {
					t.Fatal(err)
				}
				want = "REVIEW_REQUIRED"
			}
			// No network client: any accidental provider call fails the test by panic.
			n, status, err := runCampaignCanary(context.Background(), path, &deepSeekProvider{})
			if n != 0 || status != want || err == nil {
				t.Fatal(n, status, err)
			}
			file := "attempt-001.json"
			if kind == "existing" {
				file = "attempt-002.json"
			}
			if _, err := os.Lstat(filepath.Join(path, file)); !os.IsNotExist(err) {
				t.Fatal("unexpected reservation", err)
			}
		})
	}
	var out bytes.Buffer
	if code := runCampaignCanaryCLI([]string{"campaign-canary", "arbitrary-input"}, &out, &out); code != 2 {
		t.Fatal(code)
	}
}

func TestCanarySingleRequestAndNoRetry(t *testing.T) {
	for _, httpStatus := range []int{200, 503} {
		t.Run(http.StatusText(httpStatus), func(t *testing.T) {
			path := canaryTestCampaign(t)
			t.Setenv("DEEPSEEK_API_KEY", "fixture-key")
			calls := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls++
				if _, err := os.Stat(filepath.Join(path, "attempt-001.json")); err != nil {
					t.Error(err)
				}
				// Simulate another invocation while this one still owns the canary lock.
				if _, status, _ := runCampaignCanary(context.Background(), path, &deepSeekProvider{}); status != "LOCKED" {
					t.Error(status)
				}
				w.WriteHeader(httpStatus)
				_ = json.NewEncoder(w).Encode(map[string]any{"model": deepSeekModel, "usage": map[string]int{"prompt_tokens": 100, "completion_tokens": 50, "total_tokens": 150}, "choices": []any{map[string]any{"finish_reason": "stop", "message": map[string]any{"role": "assistant", "content": `{"state":"ABSTAIN","recommendation":"","reason":"Chưa đủ bằng chứng.","evidence_ids":[],"unknowns":["Chưa có kết quả đo"],"write_tool_requested":false}`}}}})
			}))
			defer server.Close()
			p := newDeepSeekProvider()
			p.endpoint = server.URL
			n, status, err := runCampaignCanary(context.Background(), path, p)
			want := "ABSTAIN"
			if httpStatus == 503 {
				want = "PROVIDER_ERROR"
			}
			if n != 1 || status != want || err != nil || calls != 1 {
				t.Fatal(n, status, err, calls)
			}
			if _, status, _ := runCampaignCanary(context.Background(), path, p); status != "REVIEW_REQUIRED" || calls != 1 {
				t.Fatal(status, calls)
			}
			report, err := readCampaignReport(path)
			if err != nil || report.Attempts != 1 || len(report.Results) != 1 {
				t.Fatal(report, err)
			}
		})
	}
}

func TestCanaryDisclosesVisibleResultAcknowledgementWithoutProviderRetry(t *testing.T) {
	path := canaryTestCampaign(t)
	t.Setenv("DEEPSEEK_API_KEY", "fixture-key")
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		_ = json.NewEncoder(w).Encode(map[string]any{"model": deepSeekModel, "usage": map[string]int{"prompt_tokens": 100, "completion_tokens": 50, "total_tokens": 150}, "choices": []any{map[string]any{"finish_reason": "stop", "message": map[string]any{"role": "assistant", "content": `{"state":"ABSTAIN","recommendation":"","reason":"Chưa đủ bằng chứng.","evidence_ids":[],"unknowns":["Chưa có kết quả đo"],"write_tool_requested":false}`}}}})
	}))
	defer server.Close()
	p := newDeepSeekProvider()
	p.endpoint = server.URL
	campaignResultWriteFault = func(phase string) error {
		if phase == "after_result_file_sync" {
			return errors.New("simulated campaign result acknowledgement loss")
		}
		return nil
	}
	t.Cleanup(func() { campaignResultWriteFault = nil })
	n, status, err := runCampaignCanary(context.Background(), path, p)
	if n != 1 || status != "PUBLISHED_RECOVERY_REQUIRED" || !isPublishedAppendUncertainty(err) || calls != 1 {
		t.Fatalf("n=%d status=%s err=%v calls=%d", n, status, err, calls)
	}
	result, err := resolvedCampaignResult(path, n)
	if err != nil || result.Status != "ABSTAIN" || result.ExecutionPermitted {
		t.Fatalf("visible result=%+v err=%v", result, err)
	}
	campaignResultWriteFault = nil
	if n, status, err := runCampaignCanary(context.Background(), path, p); n != 0 || status != "REVIEW_REQUIRED" || err == nil || calls != 1 {
		t.Fatalf("unsafe retry n=%d status=%s err=%v calls=%d", n, status, err, calls)
	}
}

func TestCanaryCLIDisclosesVisibleResultAcknowledgement(t *testing.T) {
	t.Setenv("DEEPSEEK_API_KEY", "fixture-key")
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(root, "deepseek-br11-v1")
	if err := os.Chmod(filepath.Dir(path), 0700); err != nil {
		t.Fatal(err)
	}
	if err := initAdvisorCampaign(path); err != nil {
		t.Fatal(err)
	}
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if _, err := os.Stat(filepath.Join(path, "attempt-001.json")); err != nil {
			t.Error(err)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"model": deepSeekModel, "usage": map[string]int{"prompt_tokens": 100, "completion_tokens": 50, "total_tokens": 150}, "choices": []any{map[string]any{"finish_reason": "stop", "message": map[string]any{"role": "assistant", "content": `{"state":"ABSTAIN","recommendation":"","reason":"Chưa đủ bằng chứng.","evidence_ids":[],"unknowns":["Chưa có kết quả đo"],"write_tool_requested":false}`}}}})
	}))
	defer server.Close()
	p := newDeepSeekProvider()
	p.endpoint = server.URL
	campaignCanaryProvider = func() *deepSeekProvider { return p }
	campaignCanaryPath = func() (string, error) { return path, nil }
	campaignResultWriteFault = func(phase string) error {
		if phase == "after_result_file_sync" {
			return errors.New("simulated campaign result acknowledgement loss")
		}
		return nil
	}
	t.Cleanup(func() {
		campaignCanaryProvider = newDeepSeekProvider
		campaignCanaryPath = advisorCampaignPath
		campaignResultWriteFault = nil
	})
	var stdout, stderr bytes.Buffer
	if code := runCampaignCanaryCLI([]string{"campaign-canary"}, &stdout, &stderr); code != 1 {
		t.Fatalf("code=%d stderr=%s", code, stderr.String())
	}
	var envelope map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &envelope); err != nil || envelope["status"] != "PUBLISHED_RECOVERY_REQUIRED" || envelope["execution_permitted"] != false {
		t.Fatalf("envelope=%s stderr=%s err=%v", stdout.String(), stderr.String(), err)
	}
	artifact, ok := envelope["artifact"].(map[string]any)
	if !ok || artifact["attempt"] != float64(1) || artifact["execution_permitted"] != false || calls != 1 {
		t.Fatalf("artifact=%+v calls=%d", artifact, calls)
	}
	campaignResultWriteFault = nil
	stdout.Reset()
	stderr.Reset()
	if code := runCampaignCanaryCLI([]string{"campaign-canary"}, &stdout, &stderr); code != 1 || calls != 1 {
		t.Fatalf("retry code=%d calls=%d stderr=%s", code, calls, stderr.String())
	}
	envelope = map[string]any{}
	if err := json.Unmarshal(stdout.Bytes(), &envelope); err != nil || envelope["status"] != "REVIEW_REQUIRED" || envelope["artifact"] != nil {
		t.Fatalf("retry envelope=%s err=%v", stdout.String(), err)
	}
}
