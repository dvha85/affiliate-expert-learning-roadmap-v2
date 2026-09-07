package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDeepSeekProtocol(t *testing.T) {
	for _, tc := range []struct {
		name                   string
		status                 int
		finish, content, model string
		tools                  any
		missingUsage           bool
		want                   string
	}{
		{"valid", 200, "stop", string(mockAdvisor(providerTestContext())), deepSeekModel, nil, false, "SUPPORTED"},
		{"truncated", 200, "length", "{}", deepSeekModel, nil, false, "PROVIDER_ERROR"},
		{"empty", 200, "stop", "", deepSeekModel, nil, false, "PROVIDER_ERROR"},
		{"wrong model", 200, "stop", "{}", "other", nil, false, "PROVIDER_ERROR"},
		{"tools", 200, "stop", "{}", deepSeekModel, []any{map[string]string{"id": "x"}}, false, "PROVIDER_ERROR"},
		{"missing usage", 200, "stop", "{}", deepSeekModel, nil, true, "PROVIDER_ERROR"},
		{"auth", 401, "stop", "fixture-key-secret", deepSeekModel, nil, false, "PROVIDER_ERROR"},
		{"no retry", 503, "stop", "{}", deepSeekModel, nil, false, "PROVIDER_ERROR"},
		{"bad schema", 200, "stop", "{}", deepSeekModel, nil, false, "INVALID_SCHEMA"},
		{"secret", 200, "stop", "fixture-key-secret", deepSeekModel, nil, false, "PROVIDER_ERROR"},
		{"oversize", 200, "stop", strings.Repeat("x", 65537), deepSeekModel, nil, false, "PROVIDER_ERROR"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("DEEPSEEK_API_KEY", "fixture-key-secret")
			calls := 0
			s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls++
				var request map[string]any
				if json.NewDecoder(r.Body).Decode(&request) != nil {
					t.Error("request JSON")
				}
				if request["model"] != deepSeekModel || request["stream"] != false || request["max_tokens"] != float64(1024) || request["tools"] != nil || request["thinking"].(map[string]any)["type"] != "disabled" || request["response_format"].(map[string]any)["type"] != "json_object" {
					t.Error("unsafe request")
				}
				if r.Header.Get("Authorization") != "Bearer fixture-key-secret" {
					t.Error("auth")
				}
				out := map[string]any{"model": tc.model, "choices": []any{map[string]any{"finish_reason": tc.finish, "message": map[string]any{"role": "assistant", "content": tc.content, "tool_calls": tc.tools}}}}
				if !tc.missingUsage {
					out["usage"] = map[string]int{"prompt_tokens": 100, "completion_tokens": 50, "total_tokens": 150}
				}
				w.WriteHeader(tc.status)
				_ = json.NewEncoder(w).Encode(out)
			}))
			defer s.Close()
			p := newDeepSeekProvider()
			p.endpoint = s.URL
			out, status := evaluateAdvisorProvider(context.Background(), p, providerTestContext())
			if status != tc.want || calls != 1 {
				t.Fatal(status, calls)
			}
			if status != "SUPPORTED" && out.Recommendation != "" {
				t.Fatal("rejected content escaped")
			}
		})
	}
}

func TestDeepSeekPreflightAndSecret(t *testing.T) {
	t.Setenv("DEEPSEEK_API_KEY", "")
	p := newDeepSeekProvider()
	if _, err := p.generate(context.Background(), providerTestContext()); err == nil {
		t.Fatal("missing secret accepted")
	}
	t.Setenv("DEEPSEEK_API_KEY", "fixture-key-secret")
	c := providerTestContext()
	c.Payload["e"] = strings.Repeat("x", 8193)
	if _, err := p.generate(context.Background(), c); err == nil {
		t.Fatal("oversize context accepted")
	}
}
