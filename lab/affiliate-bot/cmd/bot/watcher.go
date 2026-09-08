package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/dvha85/affiliate-expert-learning-roadmap-v2/contracts"
	"github.com/dvha85/affiliate-expert-learning-roadmap-v2/core/m00"
	"github.com/dvha85/affiliate-expert-learning-roadmap-v2/core/m06"
	corem07 "github.com/dvha85/affiliate-expert-learning-roadmap-v2/core/m07"
)

const watcherFixtureURL = "https://example.com/br13/offer"
const m07MaxToolResponseBytes = 256 << 10

type watcherFixture = m06.OfferFixture

func watcherRecord(raw []byte) (HistoryRecord, error) {
	return watcherRecordSource(raw, "")
}

// Remote source is supplied only by the fixed, hash-verified fetch adapter.
func watcherRecordSource(raw []byte, remote string) (HistoryRecord, error) {
	profile := m06.OfferFixtureProfile{
		FixtureURL: watcherFixtureURL,
		SourceURL:  watcherFixtureURL,
		AllowHost:  "example.com",
		Access:     "local_fixture",
		Role:       "synthetic_fixture",
		Limitation: "Fixture supplied locally; no fetch, seller authority or business truth verified.",
	}
	if remote != "" {
		if remote != watcherPinnedURL {
			return HistoryRecord{}, fmt.Errorf("watcher: unsupported remote source")
		}
		profile.SourceURL = remote
		profile.AllowHost = "raw.githubusercontent.com"
		profile.Access = "GET"
		profile.Limitation = "Fetched pinned synthetic fixture over HTTPS; scenario observed_at is not fetch time; no seller authority or business truth verified."
	}
	built, err := m06.BuildOfferFixture(raw, profile)
	if err != nil {
		return HistoryRecord{}, fmt.Errorf("watcher: %w", err)
	}
	converted, err := m00.Convert(built.Packet)
	if err != nil {
		return HistoryRecord{}, fmt.Errorf("watcher: invalid offer fields or numeric precision: %w", err)
	}
	projection, err := json.Marshal(converted)
	if err != nil {
		return HistoryRecord{}, fmt.Errorf("watcher: projection encoding: %w", err)
	}
	var observations []Observation
	if err := json.Unmarshal(projection, &observations); err != nil {
		return HistoryRecord{}, fmt.Errorf("watcher: projection decoding: %w", err)
	}
	// In fixture mode both timestamps are declared scenario time, not wall clock.
	return NewHistoryRecord(built.RecordID, built.ObservedAt, built.ObservedAt, observations)
}

func runWatcher(args []string, stdout, stderr io.Writer) int {
	if len(args) > 0 && args[0] == "serve" {
		return runWatcherServer(args[1:], stdout, stderr)
	}
	if len(args) > 0 && (args[0] == "history-handoff" || args[0] == "handoff") {
		return runWatcherHistoryHandoff(args[1:], stdout, stderr)
	}
	if len(args) > 0 && args[0] == "fetch-fixture" {
		return runWatcherFetch(args, stdout, stderr)
	}
	emit := func(status string, artifact any, err error, code int) int {
		if err != nil {
			fmt.Fprintln(stderr, err)
		}
		env := map[string]any{"command": "watcher fixture-import", "status": status, "execution_permitted": false, "network_fetch_performed": false, "persisted": false}
		if artifact != nil {
			env["artifact"] = artifact
			env["persisted"] = true
		}
		if err := json.NewEncoder(stdout).Encode(env); err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		return code
	}
	if len(args) != 3 || args[0] != "fixture-import" {
		return emit("USAGE_ERROR", nil, fmt.Errorf("usage: bot watcher fixture-import HISTORY FIXTURE"), 2)
	}
	if err := distinctActionPaths(args[1:]...); err != nil {
		return emit("PATH_ERROR", nil, err, 1)
	}
	info, err := os.Lstat(args[1])
	if err != nil && !os.IsNotExist(err) {
		return emit("HISTORY_ERROR", nil, err, 1)
	}
	if err == nil && !info.Mode().IsRegular() {
		return emit("PATH_ERROR", nil, fmt.Errorf("history must be regular, not symlink"), 1)
	}
	raw, err := readCampaignFile(args[2], 64<<10)
	if err != nil {
		return emit("INPUT_ERROR", nil, err, 1)
	}
	record, err := watcherRecord(raw)
	if err != nil {
		return emit("FIXTURE_ERROR", nil, err, 1)
	}
	status, resolved, err := appendResolvedHistory(args[1], record)
	if err != nil {
		return emit("HANDOFF_ERROR", nil, err, 1)
	}
	return emit(status, map[string]any{"record_id": resolved.RecordID, "decision_id": resolved.RecordedResult.DecisionID, "state": resolved.RecordedResult.State, "observation_ids": resolved.RecordedResult.EvidenceIDs}, nil, 0)
}

// m06AdapterRequest deliberately carries only a local synthetic fixture. The
// adapter, rather than n8n JavaScript, builds the M00 input and HistoryRecord
// from the shared M06 profile before it can append to canonical history.
type m06AdapterRequest struct {
	Fixture json.RawMessage `json:"fixture"`
}

type m07AdapterRequest struct {
	RecordID     string               `json:"record_id"`
	Registry     []corem07.ToolSpec   `json:"registry"`
	ToolRequest  *corem07.ToolRequest `json:"tool_request,omitempty"`
	ToolResult   json.RawMessage      `json:"tool_result,omitempty"`
	ModelOutput  json.RawMessage      `json:"model_output,omitempty"`
	ToolResultID string               `json:"tool_result_id,omitempty"`
}

func decodeM06AdapterRequest(r *http.Request) (m06AdapterRequest, error) {
	var request m06AdapterRequest
	raw, err := io.ReadAll(io.LimitReader(r.Body, 64<<10))
	if err != nil {
		return request, err
	}
	if err := contracts.DecodeStrict(raw, &request); err != nil || len(request.Fixture) == 0 {
		return request, fmt.Errorf("invalid M06 adapter request")
	}
	return request, nil
}

func appendResolvedHistory(historyPath string, record HistoryRecord) (string, HistoryRecord, error) {
	status, err := AppendHistory(historyPath, record)
	if err != nil {
		return "", HistoryRecord{}, err
	}
	resolved, err := resolveCanonicalRecord(historyPath, record.RecordID)
	if err != nil {
		return "", HistoryRecord{}, err
	}
	return status, resolved, nil
}

func m06AdapterHandler(historyPath string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			_ = json.NewEncoder(w).Encode(map[string]any{"status": "REJECT_METHOD", "canonical_history_ack": false, "execution_permitted": false})
			return
		}
		request, err := decodeM06AdapterRequest(r)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]any{"status": "INVALID_M06_INPUT", "canonical_history_ack": false, "execution_permitted": false})
			return
		}
		record, err := watcherRecord(request.Fixture)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]any{"status": "FIXTURE_ERROR", "canonical_history_ack": false, "execution_permitted": false})
			return
		}
		status, resolved, err := appendResolvedHistory(historyPath, record)
		if err != nil {
			w.WriteHeader(http.StatusConflict)
			_ = json.NewEncoder(w).Encode(map[string]any{"status": "HANDOFF_ERROR", "canonical_history_ack": false, "execution_permitted": false})
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"status": status, "record_id": resolved.RecordID, "decision_id": resolved.RecordedResult.DecisionID, "state": resolved.RecordedResult.State, "evidence_ids": resolved.RecordedResult.EvidenceIDs, "record": resolved, "canonical_history_ack": true, "canonical_history_persisted": true, "execution_permitted": false})
	}
}

func m07ArtifactPath(historyPath, kind, id string) (string, error) {
	if !strings.HasPrefix(id, "sha256:") || len(id) != len("sha256:")+64 {
		return "", fmt.Errorf("invalid M07 artifact id")
	}
	for _, r := range strings.TrimPrefix(id, "sha256:") {
		if !((r >= '0' && r <= '9') || (r >= 'a' && r <= 'f')) {
			return "", fmt.Errorf("invalid M07 artifact id")
		}
	}
	if kind != "tool-results" && kind != "proposals" {
		return "", fmt.Errorf("invalid M07 artifact kind")
	}
	return filepath.Clean(historyPath) + ".m07/" + kind + "/" + strings.TrimPrefix(id, "sha256:") + ".json", nil
}

func decodeM07AdapterRequest(r *http.Request) (m07AdapterRequest, error) {
	var request m07AdapterRequest
	raw, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		return request, err
	}
	if err := contracts.DecodeStrict(raw, &request); err != nil {
		return request, err
	}
	if strings.TrimSpace(request.RecordID) == "" || corem07.ValidateRegistry(request.Registry) != nil {
		return request, fmt.Errorf("invalid M07 adapter record or registry")
	}
	return request, nil
}

func m07AdapterContext(historyPath, recordID string) (m07Context, error) {
	record, err := resolveCanonicalRecord(historyPath, recordID)
	if err != nil {
		return m07Context{}, err
	}
	return m07EvidenceContext(record)
}

func loadM07ToolArtifact(historyPath, id string, registry []corem07.ToolSpec, recordID string) (corem07.RegisteredToolResult, error) {
	path, err := m07ArtifactPath(historyPath, "tool-results", id)
	if err != nil {
		return corem07.RegisteredToolResult{}, err
	}
	info, err := os.Lstat(path)
	if err != nil || !info.Mode().IsRegular() {
		return corem07.RegisteredToolResult{}, fmt.Errorf("registered M07 tool artifact not found")
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return corem07.RegisteredToolResult{}, err
	}
	return corem07.ValidateRegisteredToolResult(raw, registry, recordID)
}

type m07LookupIPAddr func(context.Context, string) ([]net.IPAddr, error)

func m07PublicAddress(ip net.IP) bool {
	address, ok := netip.AddrFromSlice(ip)
	if !ok {
		return false
	}
	address = address.Unmap()
	sharedAddressSpace := netip.MustParsePrefix("100.64.0.0/10")
	if !address.IsValid() || !address.IsGlobalUnicast() || address.IsPrivate() || address.IsLoopback() || address.IsLinkLocalUnicast() || address.IsMulticast() || address.IsUnspecified() || sharedAddressSpace.Contains(address) {
		return false
	}
	return true
}

func m07ResolvePublicHost(ctx context.Context, host string, lookup m07LookupIPAddr) ([]net.IPAddr, error) {
	addresses, err := lookup(ctx, host)
	if err != nil || len(addresses) == 0 {
		return nil, fmt.Errorf("M07 DNS resolution failed")
	}
	for _, address := range addresses {
		if !m07PublicAddress(address.IP) {
			return nil, fmt.Errorf("M07 host resolves to a non-public address")
		}
	}
	return addresses, nil
}

func m07ToolResponseBody(reader io.Reader) (json.RawMessage, error) {
	body, err := io.ReadAll(io.LimitReader(reader, m07MaxToolResponseBytes+1))
	if err != nil {
		return nil, err
	}
	if len(body) > m07MaxToolResponseBytes {
		return nil, fmt.Errorf("M07 tool response exceeds size limit")
	}
	// Preserve arbitrary read-only content as a JSON string when it is not a
	// JSON value; it remains unknown/untrusted at the grounding boundary.
	trimmed := bytes.TrimSpace(body)
	if !json.Valid(trimmed) {
		trimmed, err = json.Marshal(string(body))
		if err != nil {
			return nil, err
		}
	}
	return trimmed, nil
}

// fetchM07ToolResult owns the network transport used by M07. The hostname is
// resolved before the request and again for every dial; every returned address
// must be public. It disables redirects, proxies and connection reuse so an
// n8n workflow cannot turn an allowlisted hostname into an SSRF hop.
func fetchM07ToolResult(ctx context.Context, recordID string, request corem07.ToolRequest, registry []corem07.ToolSpec, lookup m07LookupIPAddr) (corem07.ToolResult, error) {
	if err := corem07.ValidateToolRequest(request, registry); err != nil {
		return corem07.ToolResult{}, err
	}
	parsed, err := url.Parse(request.Target)
	if err != nil {
		return corem07.ToolResult{}, err
	}
	var timeout time.Duration
	for _, tool := range registry {
		if tool.Name == request.ToolName {
			timeout = time.Duration(tool.TimeoutMS) * time.Millisecond
			break
		}
	}
	if timeout <= 0 {
		return corem07.ToolResult{}, fmt.Errorf("M07 tool timeout is required")
	}
	if _, err := m07ResolvePublicHost(ctx, parsed.Hostname(), lookup); err != nil {
		return corem07.ToolResult{}, err
	}
	dialer := &net.Dialer{Timeout: timeout}
	transport := &http.Transport{
		Proxy:                 nil,
		DisableKeepAlives:     true,
		ForceAttemptHTTP2:     false,
		TLSHandshakeTimeout:   timeout,
		ResponseHeaderTimeout: timeout,
		DialContext: func(dialCtx context.Context, network, address string) (net.Conn, error) {
			host, port, err := net.SplitHostPort(address)
			if err != nil || !strings.EqualFold(host, parsed.Hostname()) || port != "443" {
				return nil, fmt.Errorf("M07 dial target changed")
			}
			addresses, err := m07ResolvePublicHost(dialCtx, host, lookup)
			if err != nil {
				return nil, err
			}
			var lastErr error
			for _, candidate := range addresses {
				connection, err := dialer.DialContext(dialCtx, network, net.JoinHostPort(candidate.IP.String(), port))
				if err == nil {
					return connection, nil
				}
				lastErr = err
			}
			return nil, lastErr
		},
	}
	client := &http.Client{Transport: transport, Timeout: timeout, CheckRedirect: func(_ *http.Request, _ []*http.Request) error { return http.ErrUseLastResponse }}
	httpRequest, err := http.NewRequestWithContext(ctx, strings.ToUpper(request.Method), request.Target, nil)
	if err != nil {
		return corem07.ToolResult{}, err
	}
	response, err := client.Do(httpRequest)
	if err != nil {
		return corem07.ToolResult{}, err
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode > 299 {
		return corem07.ToolResult{}, fmt.Errorf("M07 tool returned non-success status")
	}
	body, err := m07ToolResponseBody(response.Body)
	if err != nil {
		return corem07.ToolResult{}, err
	}
	return corem07.ToolResult{RecordID: recordID, ToolCall: request, StatusCode: response.StatusCode, ReceivedAt: time.Now().UTC().Format(time.RFC3339), Redirected: false, Body: body}, nil
}

func m07AdapterHandler(historyPath string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			_ = json.NewEncoder(w).Encode(map[string]any{"status": "REJECT_METHOD", "execution_permitted": false})
			return
		}
		request, err := decodeM07AdapterRequest(r)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]any{"status": "INVALID_REQUEST", "execution_permitted": false})
			return
		}
		ctx, err := m07AdapterContext(historyPath, request.RecordID)
		if err != nil {
			w.WriteHeader(http.StatusConflict)
			_ = json.NewEncoder(w).Encode(map[string]any{"status": "HISTORY_ERROR", "execution_permitted": false})
			return
		}
		switch r.URL.Path {
		case "/v1/m07/context":
			_ = json.NewEncoder(w).Encode(map[string]any{"status": "VALID", "artifact": ctx, "execution_permitted": false})
		case "/v1/m07/preflight":
			if request.ToolRequest == nil {
				w.WriteHeader(http.StatusBadRequest)
				_ = json.NewEncoder(w).Encode(map[string]any{"status": "TOOL_REQUEST_REQUIRED", "execution_permitted": false})
				return
			}
			if err := corem07.ValidateToolRequest(*request.ToolRequest, request.Registry); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				_ = json.NewEncoder(w).Encode(map[string]any{"status": "TOOL_REQUEST_REJECTED", "execution_permitted": false})
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"status": "ALLOW_READ_ONLY", "artifact": *request.ToolRequest, "execution_permitted": false})
		case "/v1/m07/fetch-and-register":
			if request.ToolRequest == nil {
				w.WriteHeader(http.StatusBadRequest)
				_ = json.NewEncoder(w).Encode(map[string]any{"status": "TOOL_REQUEST_REQUIRED", "execution_permitted": false})
				return
			}
			result, err := fetchM07ToolResult(r.Context(), ctx.RecordID, *request.ToolRequest, request.Registry, net.DefaultResolver.LookupIPAddr)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				_ = json.NewEncoder(w).Encode(map[string]any{"status": "TOOL_TRANSPORT_REJECTED", "execution_permitted": false})
				return
			}
			rawResult, err := json.Marshal(result)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				_ = json.NewEncoder(w).Encode(map[string]any{"status": "SERIALIZATION_ERROR", "execution_permitted": false})
				return
			}
			registered, err := corem07.RegisterToolResult(rawResult, request.Registry)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				_ = json.NewEncoder(w).Encode(map[string]any{"status": "TOOL_RESULT_REJECTED", "execution_permitted": false})
				return
			}
			path, err := m07ArtifactPath(historyPath, "tool-results", registered.TraceID)
			if err == nil {
				_, err = writeNewJSON(path, registered)
			}
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				_ = json.NewEncoder(w).Encode(map[string]any{"status": "PERSISTENCE_ERROR", "execution_permitted": false})
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"status": "ACK", "artifact_id": registered.TraceID, "artifact": registered, "evidence": registered.Evidence(), "execution_permitted": false})
		case "/v1/m07/register-tool-result":
			if len(request.ToolResult) == 0 {
				w.WriteHeader(http.StatusBadRequest)
				_ = json.NewEncoder(w).Encode(map[string]any{"status": "TOOL_RESULT_REQUIRED", "execution_permitted": false})
				return
			}
			registered, err := corem07.RegisterToolResult(request.ToolResult, request.Registry)
			if err != nil || registered.RecordID != ctx.RecordID {
				w.WriteHeader(http.StatusBadRequest)
				_ = json.NewEncoder(w).Encode(map[string]any{"status": "TOOL_RESULT_REJECTED", "execution_permitted": false})
				return
			}
			path, err := m07ArtifactPath(historyPath, "tool-results", registered.TraceID)
			if err == nil {
				_, err = writeNewJSON(path, registered)
			}
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				_ = json.NewEncoder(w).Encode(map[string]any{"status": "PERSISTENCE_ERROR", "execution_permitted": false})
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"status": "ACK", "artifact_id": registered.TraceID, "artifact": registered, "evidence": registered.Evidence(), "execution_permitted": false})
		case "/v1/m07/validate", "/v1/m07/register-proposal":
			if request.ToolResultID != "" {
				registered, err := loadM07ToolArtifact(historyPath, request.ToolResultID, request.Registry, ctx.RecordID)
				if err != nil {
					w.WriteHeader(http.StatusBadRequest)
					_ = json.NewEncoder(w).Encode(map[string]any{"status": "TOOL_RESULT_REJECTED", "execution_permitted": false})
					return
				}
				ctx.Evidence = append(ctx.Evidence, registered.Evidence())
			}
			if len(request.ModelOutput) == 0 {
				w.WriteHeader(http.StatusBadRequest)
				_ = json.NewEncoder(w).Encode(map[string]any{"status": "MODEL_OUTPUT_REQUIRED", "execution_permitted": false})
				return
			}
			if r.URL.Path == "/v1/m07/validate" {
				output, err := corem07.ValidateAgentOutput(request.ModelOutput, ctx.Evidence, request.Registry)
				if err != nil {
					w.WriteHeader(http.StatusBadRequest)
					_ = json.NewEncoder(w).Encode(map[string]any{"status": "ABSTAIN", "execution_permitted": false})
					return
				}
				_ = json.NewEncoder(w).Encode(map[string]any{"status": "VALID", "artifact": output, "execution_permitted": false})
				return
			}
			proposal, err := corem07.RegisterAgentProposal(request.ModelOutput, ctx.Evidence, request.Registry, ctx.RecordID)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				_ = json.NewEncoder(w).Encode(map[string]any{"status": "PROPOSAL_REJECTED", "execution_permitted": false})
				return
			}
			path, err := m07ArtifactPath(historyPath, "proposals", proposal.ProposalID)
			if err == nil {
				_, err = writeNewJSON(path, proposal)
			}
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				_ = json.NewEncoder(w).Encode(map[string]any{"status": "PERSISTENCE_ERROR", "execution_permitted": false})
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"status": "ACK", "artifact_id": proposal.ProposalID, "artifact": proposal, "execution_permitted": false})
		default:
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]any{"status": "NOT_FOUND", "execution_permitted": false})
		}
	}
}

func runWatcherServer(args []string, stdout, stderr io.Writer) int {
	if len(args) < 1 || len(args) > 2 {
		fmt.Fprintln(stderr, "usage: bot watcher serve HISTORY [127.0.0.1:8787]")
		return 2
	}
	historyPath := args[0]
	address := "127.0.0.1:8787"
	if len(args) == 2 {
		address = args[1]
	}
	if !strings.HasPrefix(address, "127.0.0.1:") {
		fmt.Fprintln(stderr, "canonical adapter must bind to loopback 127.0.0.1")
		return 1
	}
	if err := distinctActionPaths(historyPath); err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	if info, err := os.Lstat(historyPath); err == nil && !info.Mode().IsRegular() {
		fmt.Fprintln(stderr, "history must be regular, not symlink")
		return 1
	}
	if _, err := os.Stat(historyPath); err == nil {
		if _, err := LoadHistory(historyPath); err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/m06/fixture-import", m06AdapterHandler(historyPath))
	mux.HandleFunc("/v1/m07/context", m07AdapterHandler(historyPath))
	mux.HandleFunc("/v1/m07/preflight", m07AdapterHandler(historyPath))
	mux.HandleFunc("/v1/m07/fetch-and-register", m07AdapterHandler(historyPath))
	mux.HandleFunc("/v1/m07/register-tool-result", m07AdapterHandler(historyPath))
	mux.HandleFunc("/v1/m07/validate", m07AdapterHandler(historyPath))
	mux.HandleFunc("/v1/m07/register-proposal", m07AdapterHandler(historyPath))
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"status":"OK","execution_permitted":false}`+"\n")
	})
	mux.HandleFunc("/v1/history/append", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			_, _ = io.WriteString(w, `{"status":"REJECT_METHOD","canonical_history_ack":false}`+"\n")
			return
		}
		body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = io.WriteString(w, `{"status":"INPUT_ERROR","canonical_history_ack":false}`+"\n")
			return
		}
		var record HistoryRecord
		if err := json.Unmarshal(body, &record); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = io.WriteString(w, `{"status":"INVALID_SCHEMA","canonical_history_ack":false}`+"\n")
			return
		}
		if err := validateHistoryRecord(record); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = io.WriteString(w, `{"status":"INVALID_HISTORY","canonical_history_ack":false}`+"\n")
			return
		}
		status, err := AppendHistory(historyPath, record)
		if err != nil {
			w.WriteHeader(http.StatusConflict)
			_, _ = io.WriteString(w, `{"status":"HANDOFF_ERROR","canonical_history_ack":false}`+"\n")
			return
		}
		resolved, err := resolveCanonicalRecord(historyPath, record.RecordID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]any{"status": "RESOLUTION_ERROR", "canonical_history_ack": false, "execution_permitted": false})
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"status": status, "record_id": resolved.RecordID, "canonical_history_ack": true, "canonical_history_persisted": true, "execution_permitted": false, "artifact": resolved})
	})
	mux.HandleFunc("/v1/history", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			_, _ = io.WriteString(w, `{"status":"REJECT_METHOD"}`+"\n")
			return
		}
		recordID := strings.TrimSpace(r.URL.Query().Get("record_id"))
		if recordID == "" {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = io.WriteString(w, `{"status":"RECORD_ID_REQUIRED"}`+"\n")
			return
		}
		found, err := resolveCanonicalRecord(historyPath, recordID)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			_, _ = io.WriteString(w, `{"status":"NOT_FOUND"}`+"\n")
			return
		}
		_ = json.NewEncoder(w).Encode(found)
	})
	server := &http.Server{Addr: address, Handler: mux, ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second, WriteTimeout: 15 * time.Second, IdleTimeout: 30 * time.Second}
	fmt.Fprintf(stdout, "watcher canonical adapter listening on http://%s\n", address)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		fmt.Fprintln(stderr, err)
		return 1
	}
	return 0
}

// runWatcherHistoryHandoff is the local BR-13 adapter used by the n8n
// blueprint. It accepts a complete HistoryRecord, validates it with the same
// loader used by list/replay, and reports persistence only after AppendHistory
// returns APPENDED or EXACT_DUPLICATE.
func runWatcherHistoryHandoff(args []string, stdout, stderr io.Writer) int {
	emit := func(status string, artifact any, err error, code int) int {
		if err != nil {
			fmt.Fprintln(stderr, err)
		}
		out := map[string]any{"command": "watcher history-handoff", "status": status, "execution_permitted": false, "canonical_history_ack": false}
		if status == appendAdded || status == appendDuplicate || status == "ACK" {
			out["canonical_history_ack"] = true
		}
		if artifact != nil {
			out["artifact"] = artifact
		}
		if err := json.NewEncoder(stdout).Encode(out); err != nil {
			return 1
		}
		return code
	}
	if len(args) != 2 {
		return emit("USAGE_ERROR", nil, fmt.Errorf("usage: bot watcher history-handoff HISTORY RECORD.json"), 2)
	}
	if err := distinctActionPaths(args...); err != nil {
		return emit("PATH_ERROR", nil, err, 1)
	}
	raw, err := os.ReadFile(args[1])
	if err != nil {
		return emit("INPUT_ERROR", nil, err, 1)
	}
	var record HistoryRecord
	if err := json.Unmarshal(raw, &record); err != nil {
		return emit("INVALID_SCHEMA", nil, err, 1)
	}
	if err := validateHistoryRecord(record); err != nil {
		return emit("INVALID_HISTORY", nil, err, 1)
	}
	status, resolved, err := appendResolvedHistory(args[0], record)
	if err != nil {
		return emit("HANDOFF_ERROR", nil, err, 1)
	}
	return emit(status, map[string]any{"record_id": resolved.RecordID, "decision_id": resolved.RecordedResult.DecisionID, "state": resolved.RecordedResult.State, "evidence_ids": resolved.RecordedResult.EvidenceIDs, "record": resolved}, nil, 0)
}
