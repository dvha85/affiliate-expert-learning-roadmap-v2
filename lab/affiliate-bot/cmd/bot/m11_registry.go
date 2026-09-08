package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	corem11 "github.com/dvha85/affiliate-expert-learning-roadmap-v2/core/m11"
)

func m11ArtifactRegistryPath(dir string) string { return filepath.Join(dir, "m11-artifacts.jsonl") }

func loadM11ArtifactRegistry(dir string) ([]corem11.ArtifactEntry, error) {
	raw, err := os.ReadFile(m11ArtifactRegistryPath(dir))
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	entries := []corem11.ArtifactEntry{}
	seen := map[string]string{}
	for _, line := range bytes.Split(raw, []byte{'\n'}) {
		line = bytes.TrimSpace(line)
		if len(line) == 0 {
			continue
		}
		entry, err := corem11.ValidateArtifactEntry(line)
		if err != nil {
			return nil, fmt.Errorf("invalid M11 artifact registry entry: %w", err)
		}
		key := entry.ArtifactKind + "\x00" + entry.ArtifactID
		if prior, exists := seen[key]; exists {
			if prior == entry.ContentHash {
				return nil, fmt.Errorf("duplicate M11 artifact registry entry")
			}
			return nil, fmt.Errorf("M11 artifact ID reused with different content")
		}
		seen[key] = entry.ContentHash
		entries = append(entries, entry)
	}
	if err := corem11.ValidateArtifactGraph(entries); err != nil {
		return nil, err
	}
	return entries, nil
}

func registerM11Artifact(dir, kind string, raw []byte) (corem11.ArtifactEntry, string, error) {
	entry, err := corem11.NewArtifactEntry(kind, raw)
	if err != nil {
		return entry, "", err
	}
	entries, err := loadM11ArtifactRegistry(dir)
	if err != nil {
		return entry, "", err
	}
	for _, registered := range entries {
		if registered.ArtifactKind != entry.ArtifactKind || registered.ArtifactID != entry.ArtifactID {
			continue
		}
		if registered.ContentHash == entry.ContentHash {
			return entry, appendDuplicate, nil
		}
		return entry, "", fmt.Errorf("M11 artifact ID reused with different content")
	}
	if err := corem11.ValidateArtifactGraph(append(entries, entry)); err != nil {
		return entry, "", err
	}
	line, err := json.Marshal(entry)
	if err != nil {
		return entry, "", err
	}
	f, err := os.OpenFile(m11ArtifactRegistryPath(dir), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return entry, "", err
	}
	_, err = f.Write(append(line, '\n'))
	if syncErr := f.Sync(); err == nil {
		err = syncErr
	}
	if closeErr := f.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return entry, "", err
	}
	return entry, appendAdded, nil
}

func resolveM11Artifact(dir, kind, id, hash string) (corem11.ArtifactEntry, error) {
	entries, err := loadM11ArtifactRegistry(dir)
	if err != nil {
		return corem11.ArtifactEntry{}, err
	}
	for _, entry := range entries {
		if entry.ArtifactKind == kind && entry.ArtifactID == id && (hash == "" || hash == entry.ContentHash) {
			return entry, nil
		}
	}
	return corem11.ArtifactEntry{}, fmt.Errorf("M11 artifact does not resolve")
}

func activateM11Lease(dir, leaseID, activatedAt string) (corem11.ProductionActivationRecord, string, error) {
	state, err := loadMissionState(dir)
	if err != nil {
		return corem11.ProductionActivationRecord{}, "", err
	}
	if state.Stop {
		return corem11.ProductionActivationRecord{}, "", fmt.Errorf("durable STOP: %s", state.StopReason)
	}
	entries, err := loadM11ArtifactRegistry(dir)
	if err != nil {
		return corem11.ProductionActivationRecord{}, "", err
	}
	var lease *corem11.ProductionLease
	approvalExists := false
	for _, entry := range entries {
		profile := ""
		if entry.ArtifactKind == corem11.ArtifactKindLease {
			profile = "lease"
		}
		if entry.ArtifactKind == corem11.ArtifactKindLeaseApproval {
			profile = "approval"
		}
		if profile == "" {
			continue
		}
		value, status := corem11.DecodeArtifact(profile, entry.Artifact)
		if status != corem11.Valid {
			return corem11.ProductionActivationRecord{}, "", fmt.Errorf("invalid M11 registry artifact")
		}
		if candidate, ok := value.(*corem11.ProductionLease); ok && candidate.LeaseID == leaseID {
			copied := *candidate
			lease = &copied
		}
		if candidate, ok := value.(*corem11.ProductionLeaseApproval); ok && candidate.LeaseID == leaseID {
			approvalExists = true
		}
	}
	if lease == nil || !approvalExists {
		return corem11.ProductionActivationRecord{}, "", fmt.Errorf("activation requires registered production lease and approval")
	}
	activated, errActivated := time.Parse(time.RFC3339, activatedAt)
	validFrom, errValid := time.Parse(time.RFC3339, lease.ValidFrom)
	expires, errExpires := time.Parse(time.RFC3339, lease.ExpiresAt)
	if errActivated != nil || errValid != nil || errExpires != nil || activated.Before(validFrom) || !activated.Before(expires) {
		return corem11.ProductionActivationRecord{}, "", fmt.Errorf("activation time is outside the registered lease")
	}
	record := corem11.ProductionActivationRecord{LeaseID: lease.LeaseID, LeaseVersion: lease.LeaseVersion, LeaseHash: lease.LeaseHash, ActivatedAt: activatedAt}
	raw, err := json.Marshal(record)
	if err != nil {
		return record, "", err
	}
	_, status, err := registerM11Artifact(dir, corem11.ArtifactKindActivation, raw)
	return record, status, err
}
