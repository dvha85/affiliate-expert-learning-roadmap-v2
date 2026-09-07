package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
)

type backupManifest struct {
	Version string            `json:"version"`
	Files   map[string]string `json:"files"`
}

func fileDigest(path string) (string, error) {
	b, e := os.ReadFile(path)
	if e != nil {
		return "", e
	}
	s := sha256.Sum256(b)
	return hex.EncodeToString(s[:]), nil
}
func regularFile(path string) error {
	i, e := os.Lstat(path)
	if e != nil {
		return e
	}
	if !i.Mode().IsRegular() {
		return fmt.Errorf("%s must be a regular file", path)
	}
	return nil
}
func backupFiles(source string) ([]string, error) {
	out := []string{}
	entries, err := os.ReadDir(source)
	if err != nil {
		return nil, err
	}
	for _, entry := range entries {
		if entry.Name() == "manifest.json" || entry.IsDir() {
			continue
		}
		name := entry.Name()
		p := filepath.Join(source, name)
		if err := regularFile(p); err != nil {
			return nil, err
		}
		out = append(out, name)
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("runtime created no backup artifacts")
	}
	sort.Strings(out)
	return out, nil
}
func verifyBackup(dir string) (backupManifest, error) {
	var m backupManifest
	if err := readJSON(filepath.Join(dir, "manifest.json"), &m); err != nil {
		return m, err
	}
	if m.Version != "affiliate-bot-backup/v1" || len(m.Files) == 0 {
		return m, fmt.Errorf("invalid backup manifest")
	}
	for name, want := range m.Files {
		if filepath.Base(name) != name {
			return m, fmt.Errorf("unsafe backup path")
		}
		p := filepath.Join(dir, name)
		if err := regularFile(p); err != nil {
			return m, err
		}
		got, e := fileDigest(p)
		if e != nil || got != want {
			return m, fmt.Errorf("backup checksum mismatch for %s", name)
		}
	}
	for _, required := range []string{"history.jsonl", "mission-state.json"} {
		if _, ok := m.Files[required]; !ok {
			return m, fmt.Errorf("backup is missing required %s", required)
		}
	}
	return m, nil
}

func runBackupCommand(args []string, stdout, stderr io.Writer) int {
	emit := func(status string, artifact any, err error, code int) int {
		if err != nil {
			fmt.Fprintln(stderr, err)
		}
		o := map[string]any{"command": "backup", "status": status, "execution_permitted": false}
		if artifact != nil {
			o["artifact"] = artifact
		}
		if e := json.NewEncoder(stdout).Encode(o); e != nil {
			return 1
		}
		return code
	}
	if len(args) != 3 || (args[0] != "create" && args[0] != "restore") {
		return emit("USAGE_ERROR", nil, fmt.Errorf("usage: bot backup create RUNTIME_DIR BACKUP_DIR | bot backup restore BACKUP_DIR EMPTY_TARGET_DIR"), 2)
	}
	if args[0] == "create" {
		files, e := backupFiles(args[1])
		if e != nil {
			return emit("INPUT_ERROR", nil, e, 1)
		}
		if e = os.MkdirAll(args[2], 0700); e != nil {
			return emit("STORE_ERROR", nil, e, 1)
		}
		m := backupManifest{Version: "affiliate-bot-backup/v1", Files: map[string]string{}}
		for _, name := range files {
			b, e := os.ReadFile(filepath.Join(args[1], name))
			if e != nil {
				return emit("INPUT_ERROR", nil, e, 1)
			}
			target := filepath.Join(args[2], name)
			if e = os.WriteFile(target, b, 0600); e != nil {
				return emit("STORE_ERROR", nil, e, 1)
			}
			m.Files[name], _ = fileDigest(target)
		}
		if e = writeJSON(filepath.Join(args[2], "manifest.json"), m); e != nil {
			return emit("STORE_ERROR", nil, e, 1)
		}
		return emit("BACKED_UP", m, nil, 0)
	}
	m, e := verifyBackup(args[1])
	if e != nil {
		return emit("VERIFY_FAILED", nil, e, 1)
	}
	if _, e = os.Stat(args[2]); e == nil {
		entries, _ := os.ReadDir(args[2])
		if len(entries) > 0 {
			return emit("TARGET_NOT_EMPTY", nil, fmt.Errorf("restore target must be empty"), 1)
		}
	} else if os.IsNotExist(e) {
		if e = os.MkdirAll(args[2], 0700); e != nil {
			return emit("STORE_ERROR", nil, e, 1)
		}
	} else {
		return emit("TARGET_ERROR", nil, e, 1)
	}
	for name := range m.Files {
		b, _ := os.ReadFile(filepath.Join(args[1], name))
		if e = os.WriteFile(filepath.Join(args[2], name), b, 0600); e != nil {
			return emit("STORE_ERROR", nil, e, 1)
		}
	}
	if _, ok := m.Files["history.jsonl"]; ok {
		var records []HistoryRecord
		if records, e = LoadHistory(filepath.Join(args[2], "history.jsonl")); e != nil {
			return emit("REPLAY_FAILED", nil, e, 1)
		}
		for _, record := range records {
			if replay := Replay(record); replay.State != replayMatch {
				return emit("REPLAY_FAILED", nil, fmt.Errorf("record %s replayed as %s: %s", record.RecordID, replay.State, replay.Reason), 1)
			}
		}
	}
	if _, ok := m.Files["mission-state.json"]; ok {
		if _, e = loadMissionState(args[2]); e != nil {
			return emit("STATE_FAILED", nil, e, 1)
		}
	}
	return emit("RESTORED", m, nil, 0)
}
