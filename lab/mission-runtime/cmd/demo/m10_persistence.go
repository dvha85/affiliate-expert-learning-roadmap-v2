package main

import (
	"encoding/json"
	"fmt"
	"os"
)

func decodeCanaryLedger(raw []byte) (CanaryLedger, error) {
	v, status := DecodeM10Artifact("ledger", raw)
	if status != missionValid {
		return CanaryLedger{}, fmt.Errorf("M10 persisted ledger: %s", status)
	}
	return *v.(*CanaryLedger), nil
}

func encodeCanaryLedger(l CanaryLedger) ([]byte, error) {
	b, e := json.MarshalIndent(l, "", "  ")
	if e != nil {
		return nil, e
	}
	if _, e = decodeCanaryLedger(b); e != nil {
		return nil, e
	}
	return b, nil
}

// The existence marker prevents reinitializing a newly managed ledger after its
// file disappears. It is not provenance or protection against deleting both files.
func canaryInitializationPath(path string) string { return path + ".initialized" }

func reserveCanaryInitialization(path string) error {
	if _, e := os.Lstat(path); e == nil {
		return os.ErrExist
	} else if !os.IsNotExist(e) {
		return e
	}
	f, e := os.OpenFile(canaryInitializationPath(path), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if e != nil {
		return e
	}
	if e = f.Sync(); e != nil {
		_ = f.Close()
		return e
	}
	return f.Close()
}
