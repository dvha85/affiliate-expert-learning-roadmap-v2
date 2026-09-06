package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/dvha85/affiliate-expert-learning-roadmap-v2/core/m04"
)

func providerTestContext() advisorContext {
	return advisorContext{Version: "br11a-context/v1", Config: advisorConfig{DecisionID: "d", Question: "q", AsOf: "2026-09-06T00:00:00Z", MaxAgeHours: ptr(24)}, Evidence: []m04.AdvisorEvidence{{EvidenceID: "e", ObservedAt: "2026-09-06T00:00:00Z", SourceRef: "fixture"}}, Payload: map[string]any{"e": "untrusted"}}
}
func fixtureConfig(endpoint string) fixtureProviderConfig {
	return fixtureProviderConfig{endpoint, "fixture-model", "BR11_TEST_KEY", time.Second, 2, 4096}
}
func TestFixtureProviderBoundaries(t *testing.T) {
	for _, tc := range []struct {
		name  string
		code  int
		body  string
		want  string
		calls int
	}{
		{"valid", 200, string(mockAdvisor(providerTestContext())), "SUPPORTED", 1},
		{"advice", 200, strings.ReplaceAll(string(mockAdvisor(providerTestContext())), "HUMAN_REVIEW", "ADVISE"), "REJECT_ADVICE", 1},
		{"abstain", 200, strings.ReplaceAll(string(mockAdvisor(providerTestContext())), "HUMAN_REVIEW", "ABSTAIN"), "ABSTAIN", 1},
		{"abstain unknown", 200, strings.ReplaceAll(strings.ReplaceAll(string(mockAdvisor(providerTestContext())), "HUMAN_REVIEW", "ABSTAIN"), `"e"`, `"orphan"`), "REJECT_UNGROUNDED", 1},
		{"schema", 200, `{"bad":true}`, "INVALID_SCHEMA", 1},
		{"unknown", 200, strings.ReplaceAll(string(mockAdvisor(providerTestContext())), `"e"`, `"orphan"`), "REJECT_UNGROUNDED", 1},
		{"write", 200, strings.ReplaceAll(string(mockAdvisor(providerTestContext())), `"write_tool_requested":false`, `"write_tool_requested":true`), "REJECT_WRITE_REQUEST", 1},
		{"oversize", 200, strings.Repeat("x", 4097), "PROVIDER_ERROR", 1},
		{"secret echo", 200, "fixture-secret-123", "PROVIDER_ERROR", 1},
		{"auth", 401, "fixture-secret-123", "PROVIDER_ERROR", 1},
		{"retry exhausted", 503, "fixture-secret-123", "PROVIDER_ERROR", 3},
		{"rate limit", 429, "busy", "PROVIDER_ERROR", 3},
		{"redirect", 302, "redirect", "PROVIDER_ERROR", 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("BR11_TEST_KEY", "fixture-secret-123")
			calls := 0
			s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls++
				if r.Header.Get("Authorization") != "Bearer fixture-secret-123" {
					t.Error("credential header")
				}
				var request struct {
					Identity    providerIdentity
					Instruction string
					Context     advisorContext
				}
				if err := json.NewDecoder(r.Body).Decode(&request); err != nil || request.Identity.PromptVersion != advisorPromptVersion || request.Identity.Model != "fixture-model" || request.Context.Version != "br11a-context/v1" {
					t.Error("request provenance/context")
				}
				w.Header().Set("Location", "https://example.com")
				w.WriteHeader(tc.code)
				_, _ = w.Write([]byte(tc.body))
			}))
			defer s.Close()
			p, err := newFixtureHTTPProvider(fixtureConfig(s.URL))
			if err != nil {
				t.Fatal(err)
			}
			out, status := evaluateAdvisorProvider(context.Background(), p, providerTestContext())
			if status != tc.want || calls != tc.calls {
				t.Fatalf("status %s calls %d", status, calls)
			}
			if status != "SUPPORTED" && status != "ABSTAIN" && out.Recommendation != "" {
				t.Fatal("rejected output escaped")
			}
		})
	}
}
func TestFixtureRetryRecoveryAndContextPreflight(t *testing.T) {
	t.Setenv("BR11_TEST_KEY", "fixture-secret-123")
	calls := 0
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if calls == 1 {
			w.WriteHeader(503)
			return
		}
		_, _ = w.Write(mockAdvisor(providerTestContext()))
	}))
	defer s.Close()
	p, _ := newFixtureHTTPProvider(fixtureConfig(s.URL))
	if _, status := evaluateAdvisorProvider(context.Background(), p, providerTestContext()); status != "SUPPORTED" || calls != 2 {
		t.Fatal(status, calls)
	}
	c := providerTestContext()
	c.Config.AsOf = "2026-09-08T00:00:00Z"
	if _, status := evaluateAdvisorProvider(context.Background(), p, c); status != "ABSTAIN_STALE" || calls != 2 {
		t.Fatal(status, calls)
	}
	c.Config.AsOf = "2026-09-05T00:00:00Z"
	if _, status := evaluateAdvisorProvider(context.Background(), p, c); status != "ABSTAIN_FUTURE" || calls != 2 {
		t.Fatal(status, calls)
	}
	t.Setenv("BR11_TEST_KEY", "")
	if _, err := p.generate(context.Background(), providerTestContext()); err == nil || calls != 2 {
		t.Fatal("missing credential sent request")
	}
}
func TestFixtureTimeoutCancelAndNetworkFailure(t *testing.T) {
	t.Setenv("BR11_TEST_KEY", "fixture-secret-123")
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-r.Context().Done():
		case <-time.After(100 * time.Millisecond):
		}
	}))
	c := fixtureConfig(s.URL)
	c.Timeout = 30 * time.Millisecond
	p, _ := newFixtureHTTPProvider(c)
	start := time.Now()
	if _, err := p.generate(context.Background(), providerTestContext()); err == nil {
		t.Fatal("timeout accepted")
	}
	if time.Since(start) > time.Second {
		t.Fatal("unbounded timeout")
	}
	s.Close()
	if _, err := p.generate(context.Background(), providerTestContext()); err == nil {
		t.Fatal("network failure accepted")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := p.generate(ctx, providerTestContext()); err == nil {
		t.Fatal("cancellation accepted")
	}
}
func TestFixtureConfigBounds(t *testing.T) {
	for _, endpoint := range []string{"https://example.com", "http://localhost", "http://127.0.0.1.evil.test", "http://user:pass@127.0.0.1", "http://127.0.0.1?key=secret"} {
		if _, err := newFixtureHTTPProvider(fixtureConfig(endpoint)); err == nil {
			t.Fatal(endpoint)
		}
	}
	for _, mutate := range []func(*fixtureProviderConfig){func(c *fixtureProviderConfig) { c.Timeout = 0 }, func(c *fixtureProviderConfig) { c.Timeout = 31 * time.Second }, func(c *fixtureProviderConfig) { c.Retries = -1 }, func(c *fixtureProviderConfig) { c.Retries = 3 }, func(c *fixtureProviderConfig) { c.MaxResponseBytes = 0 }, func(c *fixtureProviderConfig) { c.MaxResponseBytes = 1<<20 + 1 }, func(c *fixtureProviderConfig) { c.Model = "" }} {
		c := fixtureConfig("http://127.0.0.1")
		mutate(&c)
		if _, err := newFixtureHTTPProvider(c); err == nil {
			t.Fatal("invalid bounds accepted")
		}
	}
}

func TestFixtureExactSizeAndRequestBound(t *testing.T) {
	t.Setenv("BR11_TEST_KEY", "fixture-secret-123")
	raw := mockAdvisor(providerTestContext())
	calls := 0
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		_, _ = w.Write(raw)
	}))
	defer s.Close()
	cfg := fixtureConfig(s.URL)
	cfg.MaxResponseBytes = int64(len(raw))
	p, _ := newFixtureHTTPProvider(cfg)
	if _, status := evaluateAdvisorProvider(context.Background(), p, providerTestContext()); status != "SUPPORTED" {
		t.Fatal(status)
	}
	c := providerTestContext()
	c.Payload["e"] = strings.Repeat("x", 1<<20)
	if _, err := p.generate(context.Background(), c); err == nil || calls != 1 {
		t.Fatal("oversized request sent")
	}
	c.Config.MaxAgeHours = nil
	if _, status := evaluateAdvisorProvider(context.Background(), p, c); status != "CONFIG_ERROR" || calls != 1 {
		t.Fatal(status, calls)
	}
}
