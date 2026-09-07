package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/dvha85/affiliate-expert-learning-roadmap-v2/core/m06"
)

const watcherPinnedURL = "https://raw.githubusercontent.com/dvha85/affiliate-expert-learning-roadmap-v2/9732ea141b789b945b8b9c82677659476c86d13f/examples/watcher/offer-valid.json"
const watcherPinnedHash = "f70e18b84cf036e4973e8545f61f7ba0a111b8d69ecd25858ea60ceac877817b"

func watcherPublicIP(ip net.IP) bool {
	return ip.IsGlobalUnicast() && !ip.IsPrivate() && !ip.IsLoopback() && !ip.IsLinkLocalUnicast() && !ip.IsUnspecified()
}

func newWatcherHTTPClient() *http.Client {
	transport := &http.Transport{Proxy: nil, DisableCompression: true, TLSHandshakeTimeout: 5 * time.Second, ResponseHeaderTimeout: 10 * time.Second, MaxResponseHeaderBytes: 16 << 10}
	transport.DialContext = func(ctx context.Context, network, address string) (net.Conn, error) {
		host, port, err := net.SplitHostPort(address)
		if err != nil || host != "raw.githubusercontent.com" || port != "443" {
			return nil, fmt.Errorf("unexpected fetch destination")
		}
		ips, err := net.DefaultResolver.LookupIPAddr(ctx, host)
		if err != nil || len(ips) == 0 {
			return nil, fmt.Errorf("fetch DNS unavailable")
		}
		for _, ip := range ips {
			if !watcherPublicIP(ip.IP) {
				return nil, fmt.Errorf("fetch DNS includes forbidden address")
			}
		}
		// Dial exactly the checked address; no second DNS resolution/rebinding.
		return (&net.Dialer{Timeout: 5 * time.Second}).DialContext(ctx, "tcp", net.JoinHostPort(ips[0].IP.String(), port))
	}
	return &http.Client{Transport: transport, Timeout: 20 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
}

func fetchPinnedWatcher(ctx context.Context, client *http.Client) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, watcherPinnedURL, nil)
	if err != nil {
		return nil, fmt.Errorf("fetch request invalid")
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "affiliate-learning-br13-fixture/1")
	response, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fixture fetch failed; no automatic retry")
	}
	defer response.Body.Close()
	if response.StatusCode != 200 || response.Header.Get("Content-Encoding") != "" {
		return nil, fmt.Errorf("fixture HTTP status/encoding rejected")
	}
	raw, err := io.ReadAll(io.LimitReader(response.Body, (64<<10)+1))
	if err != nil || len(raw) > 64<<10 {
		return nil, fmt.Errorf("fixture response unreadable/too large")
	}
	if m06.ContentHash(string(raw)) != watcherPinnedHash {
		return nil, fmt.Errorf("fixture digest mismatch")
	}
	return raw, nil
}

func runWatcherFetch(args []string, stdout, stderr io.Writer) int {
	attempted := false
	emit := func(status string, artifact any, err error, code int) int {
		if err != nil {
			fmt.Fprintln(stderr, err)
		}
		env := map[string]any{"command": "watcher fetch-fixture", "status": status, "execution_permitted": false, "network_fetch_attempted": attempted, "persisted": artifact != nil}
		if artifact != nil {
			env["artifact"] = artifact
		}
		if err := json.NewEncoder(stdout).Encode(env); err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		return code
	}
	if len(args) != 2 {
		return emit("USAGE_ERROR", nil, fmt.Errorf("usage: bot watcher fetch-fixture HISTORY (no URL/key/options)"), 2)
	}
	// Preflight store and parent before making any request. One trusted writer.
	parent, err := os.Stat(filepath.Dir(args[1]))
	if err != nil || !parent.IsDir() {
		return emit("PATH_ERROR", nil, fmt.Errorf("history parent missing/not directory"), 1)
	}
	info, err := os.Lstat(args[1])
	if err == nil && !info.Mode().IsRegular() {
		return emit("PATH_ERROR", nil, fmt.Errorf("history must be regular, not symlink"), 1)
	}
	if err != nil && !os.IsNotExist(err) {
		return emit("PATH_ERROR", nil, err, 1)
	}
	if err == nil {
		if _, err := LoadHistory(args[1]); err != nil {
			return emit("HISTORY_ERROR", nil, err, 1)
		}
	}
	client := newWatcherHTTPClient()
	defer client.CloseIdleConnections()
	started := time.Now().UTC().Format(time.RFC3339Nano)
	attempted = true
	raw, err := fetchPinnedWatcher(context.Background(), client)
	if err != nil {
		return emit("FETCH_ERROR", nil, err, 1)
	}
	record, err := watcherRecordSource(raw, watcherPinnedURL)
	if err != nil {
		return emit("FIXTURE_ERROR", nil, err, 1)
	}
	status, err := AppendHistory(args[1], record)
	if err != nil {
		return emit("HANDOFF_ERROR", nil, err, 1)
	}
	return emit(status, map[string]any{"record_id": record.RecordID, "state": record.RecordedResult.State, "source_url": watcherPinnedURL, "response_sha256": watcherPinnedHash, "fetch_started_at": started, "fetch_completed_at": time.Now().UTC().Format(time.RFC3339Nano), "evidence_kind": "synthetic", "scenario_observed_at": record.AsOf}, nil, 0)
}
