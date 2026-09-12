package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/dvha85/affiliate-expert-learning-roadmap-v2/core/m03"
	"github.com/dvha85/affiliate-expert-learning-roadmap-v2/core/m05"
	corem07 "github.com/dvha85/affiliate-expert-learning-roadmap-v2/core/m07"
	corem10 "github.com/dvha85/affiliate-expert-learning-roadmap-v2/core/m10"
	corem11 "github.com/dvha85/affiliate-expert-learning-roadmap-v2/core/m11"
)

const backupManifestVersion = "affiliate-bot-backup/v3"

type backupFile struct {
	Kind      string `json:"kind"`
	SizeBytes int64  `json:"size_bytes"`
	SHA256    string `json:"sha256"`
}

type backupProfile struct {
	Name           string   `json:"name"`
	RequiredKinds  []string `json:"required_kinds"`
	InventoryKinds []string `json:"inventory_kinds"`
}

type backupManifest struct {
	Version  string                `json:"version"`
	Profile  backupProfile         `json:"profile"`
	Files    map[string]backupFile `json:"files"`
	Required []string              `json:"required"`
}

// backupCopyFault is a test-only seam for an interrupted snapshot. Backup
// publication happens only after staging has a complete, verified manifest.
var backupCopyFault func(relativePath string) error

// restoreCopyFault is a test-only seam for a backup source replacement after
// verification but before staging copy. It cannot be set through command-line
// arguments or environment.
var restoreCopyFault func(relativePath string) error

// backupStagingWriteFault is a test-only seam for the file/directory sync
// boundary while a backup or restore is still private staging. It cannot be set
// through the command line or environment.
var backupStagingWriteFault func(phase, path string) error

// stableRegularFileReadHook is a test-only seam for a same-byte replacement
// after the stable reader has opened its file. It is never configurable from
// the CLI or environment. Recovery journals use the same reader as backups,
// so this seam proves their authority boundary rather than only the backup
// copy path.
var stableRegularFileReadHook func(path string) error

// stableRegularFileContentHook runs in tests only after the shared reader has
// captured its first descriptor snapshot and before it verifies that snapshot.
// It proves a same-inode content rewrite is rejected on the real runtime-store
// read path, without making the timing configurable to production callers.
var stableRegularFileContentHook func(path string) error

// stableRegularFileAppendHook is the write-side equivalent of the reader
// hook. It is test-only: production callers cannot select an append target or
// influence the moment between the name check and open. Keeping this seam at
// the shared primitive lets the M10 and M11 registries prove that a name swap
// cannot turn their append into a write to an external file.
var stableRegularFileAppendHook func(path string) error

func backupStagingFailure(phase, path string) error {
	if backupStagingWriteFault == nil {
		return nil
	}
	return backupStagingWriteFault(phase, path)
}

// syncStagingDirectories persists every directory edge from a copied file back
// to the owned staging root. A file sync alone does not make a newly-created
// nested artifact name durable. This is a local filesystem boundary; it does
// not claim atomic recovery from an interrupted kernel or multi-host writes.
func syncStagingDirectories(root, dir string) error {
	root = filepath.Clean(root)
	relative, err := filepath.Rel(root, dir)
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return fmt.Errorf("staging directory %s is outside root %s", dir, root)
	}
	for {
		if err := syncDirectory(dir); err != nil {
			return err
		}
		if filepath.Clean(dir) == root {
			return nil
		}
		next := filepath.Dir(dir)
		if next == dir {
			return fmt.Errorf("staging directory %s is outside root %s", dir, root)
		}
		dir = next
	}
}

// writeBackupStagingFile writes one new immutable snapshot file and does not
// acknowledge it to the staging caller until file contents and its directory
// lineage are synced. The final backup/restore directory is still unpublished
// until the caller verifies the graph and atomically renames the staging root.
func writeBackupStagingFile(root, path string, data []byte) error {
	if err := backupStagingFailure("before_write", path); err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return err
	}
	if _, err = f.Write(data); err != nil {
		_ = f.Close()
		return err
	}
	if err = backupStagingFailure("after_write", path); err != nil {
		_ = f.Close()
		return err
	}
	if err = f.Sync(); err != nil {
		_ = f.Close()
		return err
	}
	if err = backupStagingFailure("after_file_sync", path); err != nil {
		_ = f.Close()
		return err
	}
	if err = f.Close(); err != nil {
		return err
	}
	if err = backupStagingFailure("before_directory_sync", path); err != nil {
		return err
	}
	if err = syncStagingDirectories(root, filepath.Dir(path)); err != nil {
		return err
	}
	return backupStagingFailure("after_directory_sync", path)
}

// backupTargetGate serializes managed snapshot publication to one destination.
// It is separate from the source runtime gate: two callers can otherwise both
// observe an empty target then race a final rename.
func acquireBackupTargetGate(target string) (func(), error) {
	identity := sha256.Sum256([]byte(filepath.Clean(target)))
	path := filepath.Join(filepath.Dir(target), ".backup-target-"+hex.EncodeToString(identity[:16])+".lock")
	if err := os.Mkdir(path, 0700); err != nil {
		if os.IsExist(err) {
			return nil, fmt.Errorf("backup target is busy or has an unrecovered publisher")
		}
		return nil, err
	}
	return func() { _ = os.Remove(path) }, nil
}

// restoreTargetGate serializes managed restores to one destination. Rename(2)
// can replace an empty directory on POSIX, so the preceding exists check alone
// is not a no-clobber guarantee when two Bot processes restore concurrently.
func acquireRestoreTargetGate(target string) (func(), error) {
	identity := sha256.Sum256([]byte(filepath.Clean(target)))
	path := filepath.Join(filepath.Dir(target), ".restore-target-"+hex.EncodeToString(identity[:16])+".lock")
	if err := os.Mkdir(path, 0700); err != nil {
		if os.IsExist(err) {
			return nil, fmt.Errorf("restore target is busy or has an unrecovered publisher")
		}
		return nil, err
	}
	return func() { _ = os.Remove(path) }, nil
}

// readStableRegularFile reads a file only if the name stayed bound to the same
// regular inode from pre-open through the read. A fully-read descriptor is
// hashed and reread before acknowledgement, so a same-inode content rewrite is
// also rejected. Canonical runtime stores and backup/restore copies share this
// local path-race guard. It is not an atomic multi-host snapshot transaction.
func readStableRegularFile(path string) ([]byte, fs.FileInfo, error) {
	before, err := os.Lstat(path)
	if err != nil {
		return nil, nil, err
	}
	if !before.Mode().IsRegular() {
		return nil, nil, fmt.Errorf("%s is not a regular file", path)
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}
	opened, statErr := f.Stat()
	if statErr != nil {
		_ = f.Close()
		return nil, nil, statErr
	}
	if !opened.Mode().IsRegular() || !os.SameFile(before, opened) {
		_ = f.Close()
		return nil, nil, fmt.Errorf("%s changed while opening stable regular file", path)
	}
	if stableRegularFileReadHook != nil {
		if err := stableRegularFileReadHook(path); err != nil {
			_ = f.Close()
			return nil, nil, err
		}
	}
	data, readErr := io.ReadAll(f)
	if readErr != nil {
		_ = f.Close()
		return nil, nil, readErr
	}
	if stableRegularFileContentHook != nil {
		if err := stableRegularFileContentHook(path); err != nil {
			_ = f.Close()
			return nil, nil, err
		}
	}
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		_ = f.Close()
		return nil, nil, err
	}
	current, readErr := io.ReadAll(f)
	if readErr != nil {
		_ = f.Close()
		return nil, nil, readErr
	}
	contentChanged := sha256.Sum256(data) != sha256.Sum256(current)
	closeErr := f.Close()
	if contentChanged {
		return nil, nil, fmt.Errorf("%s content changed while reading stable regular file", path)
	}
	if closeErr != nil {
		return nil, nil, closeErr
	}
	after, err := os.Lstat(path)
	if err != nil || !after.Mode().IsRegular() || !os.SameFile(opened, after) {
		if err != nil {
			return nil, nil, err
		}
		return nil, nil, fmt.Errorf("%s changed while reading stable regular file", path)
	}
	return data, opened, nil
}

// openStableRegularFileForAppend opens path for an append only when it remains
// the same regular file from the name check through open. A new file is
// published with O_EXCL, so a name created between the absent check and open
// cannot be followed. The caller owns the returned descriptor and must sync,
// close, then call verifyStableRegularFileName before it reports success.
//
// This guards local canonical registries from symlink/name replacement. It is
// intentionally not a multi-file transaction or a multi-host lock.
func openStableRegularFileForAppend(path string) (*os.File, fs.FileInfo, error) {
	before, err := os.Lstat(path)
	if os.IsNotExist(err) {
		if stableRegularFileAppendHook != nil {
			if err := stableRegularFileAppendHook(path); err != nil {
				return nil, nil, err
			}
		}
		f, err := os.OpenFile(path, os.O_WRONLY|os.O_APPEND|os.O_CREATE|os.O_EXCL, 0600)
		if err != nil {
			return nil, nil, err
		}
		opened, statErr := f.Stat()
		if statErr != nil || !opened.Mode().IsRegular() {
			_ = f.Close()
			if statErr != nil {
				return nil, nil, statErr
			}
			return nil, nil, fmt.Errorf("%s is not a regular file", path)
		}
		return f, opened, nil
	}
	if err != nil {
		return nil, nil, err
	}
	if !before.Mode().IsRegular() {
		return nil, nil, fmt.Errorf("%s is not a regular file", path)
	}
	if stableRegularFileAppendHook != nil {
		if err := stableRegularFileAppendHook(path); err != nil {
			return nil, nil, err
		}
	}
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_APPEND, 0)
	if err != nil {
		return nil, nil, err
	}
	opened, statErr := f.Stat()
	if statErr != nil || !opened.Mode().IsRegular() || !os.SameFile(before, opened) {
		_ = f.Close()
		if statErr != nil {
			return nil, nil, statErr
		}
		return nil, nil, fmt.Errorf("%s changed while opening stable regular file for append", path)
	}
	return f, opened, nil
}

// verifyStableRegularFileName confirms that a successful append remains named
// by the same canonical regular file before its caller acknowledges success.
func verifyStableRegularFileName(path string, opened fs.FileInfo) error {
	after, err := os.Lstat(path)
	if err != nil {
		return err
	}
	if !after.Mode().IsRegular() || !os.SameFile(opened, after) {
		return fmt.Errorf("%s changed while appending stable regular file", path)
	}
	return nil
}

func fileDigest(path string) (string, error) {
	b, _, err := readStableRegularFile(path)
	if err != nil {
		return "", err
	}
	s := sha256.Sum256(b)
	return hex.EncodeToString(s[:]), nil
}

func backupArtifactKind(name string) (string, error) {
	if strings.HasPrefix(name, "history.jsonl.m07/tool-results/") && strings.HasSuffix(name, ".json") {
		return "M07_TOOL_RESULT", nil
	}
	if strings.HasPrefix(name, "history.jsonl.m07/proposals/") && strings.HasSuffix(name, ".json") {
		return "M07_AGENT_PROPOSAL", nil
	}
	kinds := map[string]string{
		"history.jsonl":              "M02_HISTORY",
		"mission-state.json":         "MISSION_STATE",
		"actions.jsonl":              "M03_ACTION_STORE",
		"outcomes.jsonl":             "M03_OUTCOME_STORE",
		"evaluations.jsonl":          "M05_EVALUATION_STORE",
		"proposals.jsonl":            "M05_PROPOSAL_STORE",
		"reviews.jsonl":              "M05_REVIEW_STORE",
		"accesstrade-receipts.jsonl": "ACCESSTRADE_RECEIPT_STORE",
		"trusted-cost-bounds.jsonl":  "M10_COST_BOUND_STORE",
		"m10-artifacts.jsonl":        "M10_ARTIFACT_REGISTRY",
		"m10-outcomes.jsonl":         "M10_OUTCOME_STORE",
		"m11-artifacts.jsonl":        "M11_ARTIFACT_REGISTRY",
		"m11-outcomes.jsonl":         "M11_OUTCOME_STORE",
		"STOP":                       "DURABLE_STOP",
	}
	kind, ok := kinds[name]
	if !ok {
		return "", fmt.Errorf("unsupported backup artifact layout %s", name)
	}
	return kind, nil
}

func sortedKinds(files []string) ([]string, error) {
	seen := map[string]bool{}
	for _, name := range files {
		kind, err := backupArtifactKind(name)
		if err != nil {
			return nil, err
		}
		seen[kind] = true
	}
	kinds := make([]string, 0, len(seen))
	for kind := range seen {
		kinds = append(kinds, kind)
	}
	sort.Strings(kinds)
	return kinds, nil
}

func backupProfileFor(required, files []string) (backupProfile, error) {
	requiredKinds, err := sortedKinds(required)
	if err != nil {
		return backupProfile{}, err
	}
	inventoryKinds, err := sortedKinds(files)
	if err != nil {
		return backupProfile{}, err
	}
	return backupProfile{Name: "learner-runtime/offline-v3", RequiredKinds: requiredKinds, InventoryKinds: inventoryKinds}, nil
}

func backupFileMetadata(path, name string) (backupFile, error) {
	kind, err := backupArtifactKind(name)
	if err != nil {
		return backupFile{}, err
	}
	data, info, err := readStableRegularFile(path)
	if err != nil {
		return backupFile{}, err
	}
	digestBytes := sha256.Sum256(data)
	digest := hex.EncodeToString(digestBytes[:])
	return backupFile{Kind: kind, SizeBytes: info.Size(), SHA256: digest}, nil
}

func backupFileDataAndMetadata(path, name string) ([]byte, backupFile, error) {
	kind, err := backupArtifactKind(name)
	if err != nil {
		return nil, backupFile{}, err
	}
	data, info, err := readStableRegularFile(path)
	if err != nil {
		return nil, backupFile{}, err
	}
	digestBytes := sha256.Sum256(data)
	return data, backupFile{Kind: kind, SizeBytes: info.Size(), SHA256: hex.EncodeToString(digestBytes[:])}, nil
}

func backupSourceInventory(source string, files []string) (map[string]backupFile, error) {
	inventory := make(map[string]backupFile, len(files))
	for _, name := range files {
		clean, err := backupRelativePath(name)
		if err != nil {
			return nil, err
		}
		metadata, err := backupFileMetadata(filepath.Join(source, clean), name)
		if err != nil {
			return nil, err
		}
		inventory[name] = metadata
	}
	return inventory, nil
}

func sameBackupInventory(left, right map[string]backupFile) bool {
	if len(left) != len(right) {
		return false
	}
	for name, metadata := range left {
		if right[name] != metadata {
			return false
		}
	}
	return true
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
func backupRelativePath(name string) (string, error) {
	if name == "" || filepath.IsAbs(name) {
		return "", fmt.Errorf("unsafe backup path")
	}
	clean := filepath.Clean(filepath.FromSlash(name))
	if clean == "." || clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("unsafe backup path")
	}
	canonical := filepath.ToSlash(clean)
	if canonical != name {
		return "", fmt.Errorf("backup path is not normalized")
	}
	return clean, nil
}

func backupFiles(source string) ([]string, error) {
	if info, err := os.Lstat(source); err != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("runtime source must be a non-symlink directory")
	}
	out := []string{}
	err := filepath.WalkDir(source, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == source {
			return nil
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("backup does not follow symlink %s", path)
		}
		if entry.IsDir() {
			if entry.Name() == runtimeGateName {
				return filepath.SkipDir
			}
			if entry.Name() == ".mission.lock" {
				return fmt.Errorf("runtime has an active writer lock")
			}
			return nil
		}
		if !entry.Type().IsRegular() {
			return fmt.Errorf("backup only accepts regular files")
		}
		rel, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		name := filepath.ToSlash(rel)
		if name == "manifest.json" {
			return nil
		}
		if _, err := backupRelativePath(name); err != nil {
			return err
		}
		if _, err := backupArtifactKind(name); err != nil {
			return err
		}
		out = append(out, name)
		return nil
	})
	if err != nil {
		return nil, err
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("runtime created no backup artifacts")
	}
	sort.Strings(out)
	return out, nil
}

func requiredBackupFiles(source string) ([]string, error) {
	required := []string{"history.jsonl", "mission-state.json"}
	m00ToM05Files, err := m00ToM05BackupFiles(source)
	if err != nil {
		return nil, err
	}
	required = append(required, m00ToM05Files...)
	m07Files, err := m07BackupFiles(source)
	if err != nil {
		return nil, err
	}
	required = append(required, m07Files...)
	accesstradeReceiptRequired, err := accesstradeBackupReceiptRequired(source)
	if err != nil {
		return nil, err
	}
	if accesstradeReceiptRequired {
		required = append(required, filepath.Base(accesstradeReceiptPath(filepath.Join(source, "outcomes.jsonl"))))
	}
	state, err := loadMissionState(source)
	if err != nil {
		return nil, err
	}
	if state.Canary != nil {
		required = append(required, filepath.Base(m10ArtifactRegistryPath(source)))
		entries, err := loadM10ArtifactRegistry(source)
		if err != nil {
			return nil, err
		}
		for _, entry := range entries {
			if entry.ArtifactKind != corem10.ArtifactKindExecutionRecord {
				continue
			}
			record, err := corem10.ValidateExecutionRecord(entry.Artifact)
			if err != nil {
				return nil, err
			}
			if record.Status == "FAILED" {
				required = append(required, filepath.Base(m10OutcomeStorePath(source)))
				break
			}
		}
	}
	if _, err := os.Stat(m11ArtifactRegistryPath(source)); err == nil {
		required = append(required, filepath.Base(m11ArtifactRegistryPath(source)))
		entries, err := readM11ArtifactRegistry(source)
		if err != nil {
			return nil, err
		}
		for _, entry := range entries {
			if entry.ArtifactKind != corem11.ArtifactKindExecution {
				continue
			}
			record, status := corem11.DecodeArtifact("execution", entry.Artifact)
			if status != corem11.Valid {
				return nil, fmt.Errorf("invalid M11 execution artifact")
			}
			if record.(*corem11.ProductionExecutionRecord).Status == "FAILED" {
				required = append(required, filepath.Base(m11OutcomeStorePath(source)))
				break
			}
		}
	} else if !os.IsNotExist(err) {
		return nil, err
	}
	sort.Strings(required)
	return required, nil
}

func optionalRuntimeFile(dir, name string) (bool, error) {
	info, err := os.Lstat(filepath.Join(dir, name))
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
		return false, fmt.Errorf("runtime artifact %s must be a regular file", name)
	}
	return true, nil
}

// m00ToM05BackupFiles resolves every optional M03–M05 store from the prior
// store. A backup may legitimately contain history alone, but once a later
// store exists its complete upstream lineage must be present and replayable.
func m00ToM05BackupFiles(dir string) ([]string, error) {
	history, err := LoadHistory(filepath.Join(dir, "history.jsonl"))
	if err != nil {
		return nil, err
	}
	files := []string{}
	hasActions, err := optionalRuntimeFile(dir, "actions.jsonl")
	if err != nil {
		return nil, err
	}
	hasOutcomes, err := optionalRuntimeFile(dir, "outcomes.jsonl")
	if err != nil {
		return nil, err
	}
	hasEvaluations, err := optionalRuntimeFile(dir, "evaluations.jsonl")
	if err != nil {
		return nil, err
	}
	hasProposals, err := optionalRuntimeFile(dir, "proposals.jsonl")
	if err != nil {
		return nil, err
	}
	hasReviews, err := optionalRuntimeFile(dir, "reviews.jsonl")
	if err != nil {
		return nil, err
	}
	if (hasOutcomes || hasEvaluations || hasProposals || hasReviews) && !hasActions {
		return nil, fmt.Errorf("M03+ store requires actions.jsonl")
	}
	if (hasEvaluations || hasProposals || hasReviews) && !hasOutcomes {
		return nil, fmt.Errorf("M05 store requires outcomes.jsonl")
	}
	if (hasProposals || hasReviews) && !hasEvaluations {
		return nil, fmt.Errorf("improvement store requires evaluations.jsonl")
	}
	if hasReviews && !hasProposals {
		return nil, fmt.Errorf("review store requires proposals.jsonl")
	}
	if !hasActions {
		return files, nil
	}
	actions, err := loadActions(filepath.Join(dir, "actions.jsonl"), history)
	if err != nil {
		return nil, err
	}
	files = append(files, "actions.jsonl")
	if !hasOutcomes {
		return files, nil
	}
	outcomes, err := loadOutcomes(filepath.Join(dir, "outcomes.jsonl"), actions)
	if err != nil {
		return nil, err
	}
	files = append(files, "outcomes.jsonl")
	if !hasEvaluations {
		return files, nil
	}
	evaluations, err := loadEvaluations(filepath.Join(dir, "evaluations.jsonl"), history, actions, outcomes)
	if err != nil {
		return nil, err
	}
	files = append(files, "evaluations.jsonl")
	if !hasProposals {
		return files, nil
	}
	decodeProposal := func(raw []byte) (m05.ImprovementProposal, error) { return linkedProposal(raw, evaluations) }
	proposals, err := loadImprovementRecords(filepath.Join(dir, "proposals.jsonl"), decodeProposal, func(p m05.ImprovementProposal) string { return p.ProposalID })
	if err != nil {
		return nil, err
	}
	files = append(files, "proposals.jsonl")
	if !hasReviews {
		return files, nil
	}
	_, err = loadImprovementRecords(filepath.Join(dir, "reviews.jsonl"), func(raw []byte) (m05.ReviewRecord, error) { return linkedReview(raw, proposals, evaluations) }, func(r m05.ReviewRecord) string { return r.ReviewID })
	if err != nil {
		return nil, err
	}
	return append(files, "reviews.jsonl"), nil
}

// m07BackupFiles lists durable adapter artifacts only when history is stored
// inside this runtime root. A history sidecar outside the root is not silently
// skipped: it cannot be restored as part of this runtime snapshot.
func m07BackupFiles(source string) ([]string, error) {
	history := filepath.Join(source, "history.jsonl")
	sidecar := filepath.Clean(history) + ".m07"
	info, err := os.Lstat(sidecar)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return nil, fmt.Errorf("M07 artifact sidecar must be a non-symlink directory")
	}
	files := []string{}
	err = filepath.WalkDir(sidecar, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == sidecar {
			return nil
		}
		if entry.Type()&os.ModeSymlink != 0 || entry.IsDir() && entry.Name() == ".mission.lock" {
			return fmt.Errorf("unsafe M07 artifact path")
		}
		if entry.IsDir() {
			return nil
		}
		if !entry.Type().IsRegular() {
			return fmt.Errorf("M07 artifact must be a regular file")
		}
		rel, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		name := filepath.ToSlash(rel)
		if _, err := backupRelativePath(name); err != nil {
			return err
		}
		files = append(files, name)
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(files)
	return files, nil
}

func validM07ArtifactFile(kind, name, id string) bool {
	if kind != "tool-results" && kind != "proposals" || len(id) != 64 || name != id+".json" {
		return false
	}
	for _, r := range id {
		if !(r >= '0' && r <= '9' || r >= 'a' && r <= 'f') {
			return false
		}
	}
	return true
}

func validateM07BackupGraph(dir string) error {
	sidecar := filepath.Join(dir, "history.jsonl.m07")
	if _, err := os.Lstat(sidecar); os.IsNotExist(err) {
		state, stateErr := loadMissionState(dir)
		if stateErr != nil {
			return stateErr
		}
		if state.Intent != nil && state.Intent.ProposedBy == "agent" {
			return fmt.Errorf("agent intent is missing its persisted M07 proposal")
		}
		return nil
	} else if err != nil {
		return err
	}
	records, err := LoadHistory(filepath.Join(dir, "history.jsonl"))
	if err != nil {
		return err
	}
	byID := map[string]HistoryRecord{}
	for _, record := range records {
		byID[record.RecordID] = record
	}
	proposalPaths := []string{}
	toolEvidence := map[string][]corem07.Evidence{}
	err = filepath.WalkDir(sidecar, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == sidecar {
			return nil
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("unsafe M07 artifact")
		}
		rel, err := filepath.Rel(sidecar, path)
		if err != nil {
			return err
		}
		parts := strings.Split(filepath.ToSlash(rel), "/")
		if entry.IsDir() {
			if len(parts) != 1 || parts[0] != "tool-results" && parts[0] != "proposals" {
				return fmt.Errorf("invalid M07 artifact layout")
			}
			return nil
		}
		if !entry.Type().IsRegular() {
			return fmt.Errorf("unsafe M07 artifact")
		}
		if len(parts) != 2 || !validM07ArtifactFile(parts[0], parts[1], strings.TrimSuffix(parts[1], ".json")) {
			return fmt.Errorf("invalid M07 artifact layout")
		}
		raw, _, err := readStableRegularFile(path)
		if err != nil {
			return err
		}
		if parts[0] == "tool-results" {
			var stored corem07.RegisteredToolResult
			if err := json.Unmarshal(raw, &stored); err != nil || !strings.HasPrefix(stored.TraceID, "sha256:") || strings.TrimPrefix(stored.TraceID, "sha256:") != strings.TrimSuffix(parts[1], ".json") {
				return fmt.Errorf("M07 tool result filename does not bind its trace")
			}
			registered, err := corem07.ValidateStoredRegisteredToolResult(raw, stored.RecordID)
			if err != nil || byID[registered.RecordID].RecordID == "" {
				return fmt.Errorf("M07 tool result is not valid for canonical history")
			}
			toolEvidence[registered.RecordID] = append(toolEvidence[registered.RecordID], registered.Evidence())
			return nil
		}
		proposalPaths = append(proposalPaths, path)
		return nil
	})
	if err != nil {
		return err
	}
	proposals := map[string]corem07.AgentOutput{}
	for _, path := range proposalPaths {
		raw, _, err := readStableRegularFile(path)
		if err != nil {
			return err
		}
		var stored corem07.RegisteredAgentProposal
		if err := json.Unmarshal(raw, &stored); err != nil || !strings.HasPrefix(stored.ProposalID, "sha256:") || strings.TrimPrefix(stored.ProposalID, "sha256:") != strings.TrimSuffix(filepath.Base(path), ".json") {
			return fmt.Errorf("M07 proposal filename does not bind its proposal")
		}
		record, ok := byID[stored.RecordID]
		if !ok {
			return fmt.Errorf("M07 proposal references missing canonical history")
		}
		ctx, err := m07EvidenceContext(record)
		if err != nil {
			return err
		}
		ctx.Evidence = append(ctx.Evidence, toolEvidence[record.RecordID]...)
		proposal, output, err := corem07.ValidateRegisteredAgentProposal(raw, ctx.Evidence, nil, record.RecordID)
		if err != nil {
			return fmt.Errorf("M07 proposal is not grounded after restore: %w", err)
		}
		proposals[proposal.ProposalID] = output
	}
	state, err := loadMissionState(dir)
	if err != nil {
		return err
	}
	if state.Intent == nil || state.Intent.ProposedBy != "agent" {
		return nil
	}
	output, ok := proposals[state.Intent.ProposalRef]
	if !ok || output.ProposedAction == nil || output.ProposedAction.ActionType != state.Intent.ActionType || output.ProposedAction.Target != state.Intent.Target || !sameParameters(output.ProposedAction.Parameters, state.Intent.Parameters) {
		return fmt.Errorf("agent intent does not resolve to restored M07 proposal")
	}
	for _, id := range state.Intent.EvidenceIDs {
		found := false
		for _, proposalID := range output.EvidenceIDs {
			found = found || id == proposalID
		}
		if !found {
			return fmt.Errorf("agent intent has evidence absent from restored M07 proposal")
		}
	}
	return nil
}

func validateM10BackupGraph(dir string) error {
	state, err := loadMissionState(dir)
	if err != nil || state.Canary == nil {
		return err
	}
	entries, err := loadM10ArtifactRegistry(dir)
	if err != nil {
		return err
	}
	// The mutable mission state is never sufficient evidence of a delegation.
	// On restore its active canary must resolve to the exact immutable registry
	// entry; otherwise a checksum-valid backup could retain counters while
	// silently dropping the grant that bounds them.
	grantRaw, err := json.Marshal(state.Canary.CanaryGrant)
	if err != nil {
		return err
	}
	grantEntry, err := corem10.NewArtifactEntry(corem10.ArtifactKindCanaryGrant, grantRaw)
	if err != nil {
		return fmt.Errorf("restored state has invalid canary grant: %w", err)
	}
	grantFound := false
	bounds := map[string]corem10.TrustedCostBound{}
	authorizations := map[string]corem10.ExecutionAuthorization{}
	executions := map[string]corem10.ExecutionRecord{}
	for _, entry := range entries {
		if entry.ArtifactKind == corem10.ArtifactKindCanaryGrant && entry.ArtifactID == grantEntry.ArtifactID && entry.ContentHash == grantEntry.ContentHash && bytes.Equal(entry.Artifact, grantEntry.Artifact) {
			grantFound = true
		}
		if entry.ArtifactKind == corem10.ArtifactKindExecutionAuthorization {
			authorization, err := corem10.ValidateExecutionAuthorization(entry.Artifact)
			if err != nil {
				return fmt.Errorf("restored authorization is invalid: %w", err)
			}
			authorizations[authorization.AuthorizationID] = authorization
		}
		if entry.ArtifactKind == corem10.ArtifactKindTrustedCostBound {
			bound, status := corem10.DecodeTrustedCostBound(entry.Artifact)
			if status != "VALID" {
				return fmt.Errorf("restored cost bound is invalid")
			}
			bounds[bound.CostBoundID] = bound
		}
		if entry.ArtifactKind != corem10.ArtifactKindExecutionRecord {
			continue
		}
		record, err := corem10.ValidateExecutionRecord(entry.Artifact)
		if err != nil || !reservationForExecution(state, record.ExecutionID) {
			return fmt.Errorf("M10 execution record is orphaned from restored reservation")
		}
		executions[record.ExecutionID] = record
	}
	if !grantFound {
		return fmt.Errorf("active canary grant is absent from restored registry")
	}
	// A pending reservation has already consumed mutable canary budget. It must
	// continue to resolve to the immutable authorization that established its
	// binding, even before an execution record exists to expose the orphan.
	for _, reservation := range state.Reservations {
		if reservation.CostBoundID != "" || reservation.CostBoundHash != "" {
			bound, found := bounds[reservation.CostBoundID]
			if !found || reservation.CostBoundID == "" || reservation.CostBoundHash == "" || bound.CostBoundHash != reservation.CostBoundHash || bound.IntentID != reservation.IntentID || bound.IntentHash != reservation.IntentHash || bound.MaxCostMinor != reservation.CostMinor {
				return fmt.Errorf("reservation is orphaned from restored cost bound")
			}
		}
		if reservation.ReservationMode != "GOVERNED_AUTHORIZATION" {
			continue
		}
		authorization, found := authorizations[reservation.AuthorizationID]
		if !found || !authorizationBindsMissionState(authorization, state) || authorization.CanaryGrantID != reservation.GrantID || authorization.IntentID != reservation.IntentID || authorization.IntentHash != reservation.IntentHash || authorization.CanaryCostBoundMinor != reservation.CostMinor || authorization.CanaryCostBoundID != reservation.CostBoundID || authorization.CanaryCostBoundHash != reservation.CostBoundHash {
			return fmt.Errorf("governed reservation is orphaned from restored authorization")
		}
		reservedAt, reservedAtErr := time.Parse(time.RFC3339, reservation.ReservedAt)
		authorizedAt, authorizedAtErr := time.Parse(time.RFC3339, authorization.AuthorizedAt)
		expiresAt, expiresAtErr := time.Parse(time.RFC3339, authorization.ExpiresAt)
		if reservedAtErr != nil || authorizedAtErr != nil || expiresAtErr != nil || reservedAt.Before(authorizedAt) || !reservedAt.Before(expiresAt) {
			return fmt.Errorf("governed reservation falls outside authorization lifetime")
		}
		if reservation.ExecutionID == "" {
			continue
		}
		record, found := executions[reservation.ExecutionID]
		if !found || record.AuthorizationID != reservation.AuthorizationID {
			return fmt.Errorf("governed reservation execution is orphaned or mismatched")
		}
		attemptedAt, attemptedAtErr := time.Parse(time.RFC3339, record.AttemptedAt)
		if attemptedAtErr != nil || attemptedAt.Before(reservedAt) {
			return fmt.Errorf("governed reservation execution predates its reservation")
		}
	}
	outcomes, err := loadM10FixtureOutcomes(dir, state)
	if err != nil {
		return err
	}
	failedOutcomeIDs := map[string]bool{}
	for _, outcome := range outcomes {
		failedOutcomeIDs[outcome.EffectRef.EffectID] = true
	}
	for executionID, record := range executions {
		if record.Status == "FAILED" && !failedOutcomeIDs[executionID] {
			return fmt.Errorf("failed execution is missing restored fixture outcome")
		}
	}
	return nil
}

func validateM11BackupGraph(dir string) error {
	entries, err := loadM11ArtifactRegistry(dir)
	if err != nil {
		return fmt.Errorf("M11 artifact graph is invalid: %w", err)
	}
	outcomes, err := loadM11FixtureOutcomes(dir)
	if err != nil {
		return fmt.Errorf("M11 outcome store is invalid: %w", err)
	}
	history, err := LoadHistory(filepath.Join(dir, "history.jsonl"))
	if err != nil {
		return fmt.Errorf("M11 canonical history is invalid: %w", err)
	}
	historyByID := map[string]HistoryRecord{}
	for _, record := range history {
		if replay := Replay(record); replay.State != replayMatch {
			return fmt.Errorf("M11 canonical history does not replay: %s", record.RecordID)
		}
		if _, exists := historyByID[record.RecordID]; exists {
			return fmt.Errorf("M11 canonical history has an ambiguous record ID")
		}
		historyByID[record.RecordID] = record
	}
	linked := map[string]bool{}
	outcomesByID := map[string]m03.OutcomeRecord{}
	outcomeExecutionIDs := map[string]string{}
	leases := map[string]corem11.ProductionLease{}
	leaseApprovals := map[string]corem11.ProductionLeaseApproval{}
	activations := map[string]corem11.ProductionActivationRecord{}
	healthSnapshots := map[string]corem11.ProductionHealthSnapshot{}
	gates := map[string]corem11.ProductionGateDecision{}
	ledgers := []corem11.ProductionLedger{}
	ledgersByArtifactID := map[string]corem11.ProductionLedger{}
	resolutions := map[string]corem11.ProductionReconciliationResolution{}
	authorizations := map[string]corem11.ProductionExecutionAuthorization{}
	executions := []corem11.ProductionExecutionRecord{}
	executionsByID := map[string]corem11.ProductionExecutionRecord{}
	evaluations := map[string]corem11.ProductionOutcomeEvaluation{}
	cycles := []corem11.ProductionCycleRecord{}
	admissions := []corem11.ProductionRecoveryAdmission{}
	admissionNewLeases := map[string]bool{}
	admissionPriorResolutions := map[string]bool{}
	for _, outcome := range outcomes {
		if prior, exists := outcomeExecutionIDs[outcome.EffectRef.EffectID]; exists && prior != outcome.OutcomeID {
			return fmt.Errorf("M11 execution has more than one restored fixture outcome")
		}
		outcomeExecutionIDs[outcome.EffectRef.EffectID] = outcome.OutcomeID
		linked[outcome.EffectRef.EffectID] = true
		outcomesByID[outcome.OutcomeID] = outcome
	}
	for _, entry := range entries {
		switch entry.ArtifactKind {
		case corem11.ArtifactKindLease:
			value, status := corem11.DecodeArtifact("lease", entry.Artifact)
			if status != corem11.Valid {
				return fmt.Errorf("M11 lease artifact is invalid")
			}
			lease := *value.(*corem11.ProductionLease)
			leases[lease.LeaseID] = lease
		case corem11.ArtifactKindLeaseApproval:
			value, status := corem11.DecodeArtifact("approval", entry.Artifact)
			if status != corem11.Valid {
				return fmt.Errorf("M11 lease approval artifact is invalid")
			}
			approval := *value.(*corem11.ProductionLeaseApproval)
			leaseApprovals[approval.ApprovalID] = approval
		case corem11.ArtifactKindActivation:
			value, status := corem11.DecodeArtifact("activation", entry.Artifact)
			if status != corem11.Valid {
				return fmt.Errorf("M11 activation artifact is invalid")
			}
			activation := *value.(*corem11.ProductionActivationRecord)
			activations[activation.LeaseID] = activation
		case corem11.ArtifactKindHealth:
			value, status := corem11.DecodeArtifact("health", entry.Artifact)
			if status != corem11.Valid {
				return fmt.Errorf("M11 health artifact is invalid")
			}
			snapshot := *value.(*corem11.ProductionHealthSnapshot)
			healthSnapshots[snapshot.SnapshotID] = snapshot
		case corem11.ArtifactKindLedger:
			value, status := corem11.DecodeArtifact("ledger", entry.Artifact)
			if status != corem11.Valid {
				return fmt.Errorf("M11 ledger artifact is invalid")
			}
			ledger := *value.(*corem11.ProductionLedger)
			ledgers = append(ledgers, ledger)
			ledgersByArtifactID[entry.ArtifactID] = ledger
		case corem11.ArtifactKindReconciliation:
			value, status := corem11.DecodeArtifact("resolution", entry.Artifact)
			if status != corem11.Valid {
				return fmt.Errorf("M11 reconciliation artifact is invalid")
			}
			resolution := *value.(*corem11.ProductionReconciliationResolution)
			if _, exists := resolutions[resolution.ExecutionID]; exists {
				return fmt.Errorf("M11 execution has more than one restored reconciliation resolution")
			}
			resolutions[resolution.ExecutionID] = resolution
		case corem11.ArtifactKindAuthorization:
			value, status := corem11.DecodeArtifact("authorization", entry.Artifact)
			if status != corem11.Valid {
				return fmt.Errorf("M11 authorization artifact is invalid")
			}
			authorization := *value.(*corem11.ProductionExecutionAuthorization)
			authorizations[authorization.AuthorizationID] = authorization
		case corem11.ArtifactKindGate:
			value, status := corem11.DecodeArtifact("gate", entry.Artifact)
			if status != corem11.Valid {
				return fmt.Errorf("M11 gate artifact is invalid")
			}
			gate := *value.(*corem11.ProductionGateDecision)
			gates[gate.GateID] = gate
		case corem11.ArtifactKindExecution:
			value, status := corem11.DecodeArtifact("execution", entry.Artifact)
			if status != corem11.Valid {
				return fmt.Errorf("M11 execution artifact is invalid")
			}
			record := *value.(*corem11.ProductionExecutionRecord)
			executions = append(executions, record)
			executionsByID[record.ExecutionID] = record
		case corem11.ArtifactKindEvaluation:
			value, status := corem11.DecodeArtifact("evaluation", entry.Artifact)
			if status != corem11.Valid {
				return fmt.Errorf("M11 outcome evaluation artifact is invalid")
			}
			evaluation := *value.(*corem11.ProductionOutcomeEvaluation)
			evaluations[evaluation.EvaluationID] = evaluation
		case corem11.ArtifactKindCycle:
			value, status := corem11.DecodeArtifact("cycle", entry.Artifact)
			if status != corem11.Valid {
				return fmt.Errorf("M11 cycle artifact is invalid")
			}
			cycles = append(cycles, *value.(*corem11.ProductionCycleRecord))
		case corem11.ArtifactKindRecoveryAdmission:
			value, status := corem11.DecodeArtifact("recovery_admission", entry.Artifact)
			if status != corem11.Valid {
				return fmt.Errorf("M11 recovery admission artifact is invalid")
			}
			admissions = append(admissions, *value.(*corem11.ProductionRecoveryAdmission))
			admission := *value.(*corem11.ProductionRecoveryAdmission)
			priorKey := admission.PriorRuntimeDir + "\x00" + admission.ResolutionID
			if admissionNewLeases[admission.NewLeaseID] || admissionPriorResolutions[priorKey] {
				return fmt.Errorf("M11 recovery admission lineage is ambiguous")
			}
			admissionNewLeases[admission.NewLeaseID], admissionPriorResolutions[priorKey] = true, true
		}
	}
	// The append-only registry permits a lease to be recorded before its human
	// approval arrives, but a backup is a complete runtime snapshot. Every
	// persisted lease must therefore retain the exact approval it names; an
	// otherwise checksum-valid snapshot must not restore a delegation whose
	// review evidence has been removed.
	for _, lease := range leases {
		approval, found := leaseApprovals[lease.ApprovalRef]
		if !found || approval.LeaseID != lease.LeaseID || approval.LeaseVersion != lease.LeaseVersion || approval.LeaseHash != lease.LeaseHash || approval.PromotionReviewRef != lease.PromotionReviewRef || approval.SourceCanaryGrantID != lease.SourceCanaryGrantID || approval.SourceCanaryGrantVersion != lease.SourceCanaryGrantVersion || approval.SourceCanaryGrantHash != lease.SourceCanaryGrantHash || approval.ReviewerID != lease.ReviewerID || approval.ReviewedAt != lease.ReviewedAt {
			return fmt.Errorf("M11 lease is orphaned from its restored approval")
		}
	}
	for _, outcome := range outcomes {
		if outcome.EffectRef.EffectKind != "MACHINE_EXECUTION" {
			return fmt.Errorf("M11 fixture outcome has an unsupported effect kind")
		}
		if _, found := executionsByID[outcome.EffectRef.EffectID]; !found {
			return fmt.Errorf("M11 fixture outcome is orphaned from its restored execution")
		}
	}
	// Fixture outcomes are not merely an auxiliary store: completing one also
	// commits the corresponding lease ledger transition. Validate that link in
	// both directions so a checksum-valid snapshot cannot retain an outcome
	// while silently erasing its accounting/audit record (or invent one).
	ledgerOutcomeLinks := map[string]bool{}
	for _, ledger := range ledgers {
		for _, link := range ledger.OutcomeLinks {
			outcome, outcomeOK := outcomesByID[link.OutcomeID]
			execution, executionOK := executionsByID[link.ExecutionID]
			if !outcomeOK || !executionOK || outcome.EffectRef.EffectKind != "MACHINE_EXECUTION" || outcome.EffectRef.EffectID != link.ExecutionID || outcome.ObservedAt != link.ObservedAt || execution.ProductionLeaseID != ledger.LeaseID || execution.ProductionLeaseVersion != ledger.LeaseVersion || execution.ProductionLeaseHash != ledger.LeaseHash {
				return fmt.Errorf("M11 ledger outcome link is orphaned or mismatched")
			}
			ledgerOutcomeLinks[link.OutcomeID] = true
		}
	}
	for _, admission := range admissions {
		lease, leaseOK := leases[admission.NewLeaseID]
		approval, approvalOK := leaseApprovals[admission.NewApprovalID]
		activation, activationOK := activations[admission.NewLeaseID]
		reviewedAt, reviewedErr := time.Parse(time.RFC3339, admission.ReviewedAt)
		approvalAt, approvalErr := time.Parse(time.RFC3339, approval.ReviewedAt)
		normalLedger := false
		for _, ledger := range ledgers {
			normalLedger = normalLedger || ledger.LeaseID == admission.NewLeaseID && ledger.LeaseVersion == admission.NewLeaseVersion && ledger.LeaseHash == admission.NewLeaseHash && ledger.ControlMode == "NORMAL" && !ledger.ReconciliationRequired
		}
		if !leaseOK || !approvalOK || !activationOK || reviewedErr != nil || approvalErr != nil || !reviewedAt.After(approvalAt) || admission.ExecutionPermitted || lease.LeaseVersion != admission.NewLeaseVersion || lease.LeaseHash != admission.NewLeaseHash || lease.ApprovalRef != admission.NewApprovalID || approval.LeaseID != lease.LeaseID || approval.LeaseVersion != lease.LeaseVersion || approval.LeaseHash != lease.LeaseHash || activation.LeaseVersion != lease.LeaseVersion || activation.LeaseHash != lease.LeaseHash || !normalLedger {
			return fmt.Errorf("M11 recovery admission lacks a restored new-runtime admission boundary")
		}
	}
	for outcomeID := range outcomesByID {
		if !ledgerOutcomeLinks[outcomeID] {
			return fmt.Errorf("M11 fixture outcome is absent from its restored ledger")
		}
	}
	// A lease may remain registered but inactive. Once a ledger exists, however,
	// it is a lifecycle state created after activation; restore must retain that
	// exact activation rather than accepting a checksum-valid ledger detached
	// from its admission boundary.
	for _, ledger := range ledgers {
		activation, found := activations[ledger.LeaseID]
		activatedAt, activatedAtErr := time.Parse(time.RFC3339, activation.ActivatedAt)
		ledgerAt, ledgerAtErr := time.Parse(time.RFC3339, ledger.UpdatedAt)
		if !found || activatedAtErr != nil || ledgerAtErr != nil || activation.LeaseVersion != ledger.LeaseVersion || activation.LeaseHash != ledger.LeaseHash || activatedAt.After(ledgerAt) {
			return fmt.Errorf("M11 ledger is orphaned from its restored activation")
		}
	}
	// A gate carries a budget snapshot for audit and later authorization. Its
	// immutable ledger reference is insufficient if those copied counters are
	// allowed to drift from the referenced ledger in a checksum-valid backup.
	for _, gate := range gates {
		ledger, found := ledgersByArtifactID[gate.LedgerArtifactID]
		if !found || gate.ExecutionsTotalBefore != ledger.ExecutionsTotal || gate.ExecutionsInWindowBefore != ledger.ExecutionsInWindow || gate.CostMinorTotalBefore != ledger.CostMinorTotal || gate.PendingOutcomesBefore != ledger.PendingOutcomes {
			return fmt.Errorf("M11 gate budget snapshot does not match its restored ledger")
		}
		activation, activationOK := activations[gate.LeaseID]
		health, healthOK := healthSnapshots[gate.HealthSnapshotID]
		gateAt, gateAtErr := time.Parse(time.RFC3339, gate.EvaluatedAt)
		activatedAt, activatedAtErr := time.Parse(time.RFC3339, activation.ActivatedAt)
		healthAt, healthAtErr := time.Parse(time.RFC3339, health.ObservedAt)
		if !activationOK || !healthOK || gateAtErr != nil || activatedAtErr != nil || healthAtErr != nil || activation.LeaseVersion != gate.LeaseVersion || activation.LeaseHash != gate.LeaseHash || health.LeaseID != gate.LeaseID || health.LeaseVersion != gate.LeaseVersion || health.LeaseHash != gate.LeaseHash || gateAt.Before(activatedAt) || healthAt.Before(activatedAt) {
			return fmt.Errorf("M11 gate or health predates its restored activation")
		}
	}
	// Authorization is a historical decision, so restore does not compare it
	// with the current clock. It must nevertheless retain the same antecedent
	// ALLOW gate and health snapshot that the command path observed.
	for _, authorization := range authorizations {
		gate, gateOK := gates[authorization.ProductionGateID]
		health, healthOK := healthSnapshots[authorization.ProductionHealthSnapshotID]
		authorizedAt, authorizedAtErr := time.Parse(time.RFC3339, authorization.AuthorizedAt)
		gateAt, gateAtErr := time.Parse(time.RFC3339, gate.EvaluatedAt)
		healthAt, healthAtErr := time.Parse(time.RFC3339, health.ObservedAt)
		if !gateOK || !healthOK || authorizedAtErr != nil || gateAtErr != nil || healthAtErr != nil || gate.Decision != "ALLOW_PRODUCTION" || authorizedAt.Before(gateAt) || authorizedAt.Before(healthAt) {
			return fmt.Errorf("M11 authorization lacks a prior restored allow gate and health snapshot")
		}
	}
	// Reconciliation is a narrowly scoped human review of an UNKNOWN effect.
	// The command path already enforces this before advancing a stopped ledger;
	// mirror the boundary here so a checksum-valid backup cannot smuggle a
	// generic (or pre-attempt) resolution into the restored audit history.
	for executionID, resolution := range resolutions {
		execution, found := executionsByID[executionID]
		resolvedAt, resolvedAtErr := time.Parse(time.RFC3339, resolution.ResolvedAt)
		attemptedAt, attemptedAtErr := time.Parse(time.RFC3339, execution.AttemptedAt)
		if !found || resolution.ResolvedBy != "human" || resolution.EffectState != "NOT_PERFORMED" || execution.Status != "RECONCILIATION_REQUIRED" || execution.SideEffectState != "UNKNOWN" || execution.ProductionLeaseID != resolution.LeaseID || execution.ProductionLeaseVersion != resolution.LeaseVersion || execution.ProductionLeaseHash != resolution.LeaseHash || resolvedAtErr != nil || attemptedAtErr != nil || resolvedAt.Before(attemptedAt) {
			return fmt.Errorf("M11 reconciliation does not bind its restored unknown execution")
		}
	}
	// Historical artifacts may be restored after their authority expires, but
	// their recorded attempt cannot have happened outside that authority's
	// lifetime. Preserve the command-path boundary in the semantic restore
	// graph; no current-clock decision is made here.
	for _, execution := range executions {
		authorization, found := authorizations[execution.AuthorizationID]
		attemptedAt, attemptedAtErr := time.Parse(time.RFC3339, execution.AttemptedAt)
		authorizedAt, authorizedAtErr := time.Parse(time.RFC3339, authorization.AuthorizedAt)
		expiresAt, expiresAtErr := time.Parse(time.RFC3339, authorization.ExpiresAt)
		if !found || attemptedAtErr != nil || authorizedAtErr != nil || expiresAtErr != nil || attemptedAt.Before(authorizedAt) || !attemptedAt.Before(expiresAt) {
			return fmt.Errorf("M11 execution falls outside its restored authorization lifetime")
		}
		// The learner records only a governed execution that was first charged
		// into a prior, normal ledger. Keep that predecessor link in the backup
		// graph; the later outcome ledger alone cannot prove an execution was
		// ever reserved within its lease budget.
		reserved := false
		for _, ledger := range ledgers {
			ledgerAt, ledgerAtErr := time.Parse(time.RFC3339, ledger.UpdatedAt)
			if ledgerAtErr != nil || ledger.LeaseID != execution.ProductionLeaseID || ledger.LeaseVersion != execution.ProductionLeaseVersion || ledger.LeaseHash != execution.ProductionLeaseHash || ledger.ControlMode != "NORMAL" || ledgerAt.After(attemptedAt) {
				continue
			}
			for _, pendingID := range ledger.PendingExecutionIDs {
				reserved = reserved || pendingID == execution.ExecutionID
			}
		}
		if !reserved {
			return fmt.Errorf("M11 execution is orphaned from its restored reservation ledger")
		}
	}
	for _, record := range executions {
		if record.Status == "FAILED" && !linked[record.ExecutionID] {
			return fmt.Errorf("failed M11 execution is missing restored fixture outcome")
		}
		if record.Status != "RECONCILIATION_REQUIRED" || record.SideEffectState != "UNKNOWN" {
			continue
		}
		matched := false
		for _, ledger := range ledgers {
			if ledger.LeaseID != record.ProductionLeaseID || ledger.LeaseVersion != record.ProductionLeaseVersion || ledger.LeaseHash != record.ProductionLeaseHash || ledger.ControlMode != "STOPPED" {
				continue
			}
			if resolution, resolved := resolutions[record.ExecutionID]; resolved {
				if !ledger.ReconciliationRequired && ledger.StopReason == "RECOVERY_REVIEW_REQUIRED" {
					for _, id := range ledger.ReconciliationResolutionIDs {
						matched = matched || id == resolution.ResolutionID
					}
				}
			} else if ledger.ReconciliationRequired && ledger.StopReason == "RECONCILIATION_REQUIRED" {
				matched = true
			}
		}
		if !matched {
			return fmt.Errorf("unknown M11 execution lacks matching stopped reconciliation ledger")
		}
	}
	for _, evaluation := range evaluations {
		outcome, outcomeOK := outcomesByID[evaluation.OutcomeID]
		execution, executionOK := executionsByID[evaluation.ExecutionID]
		outcomeAt, outcomeTimeErr := time.Parse(time.RFC3339, outcome.ObservedAt)
		evaluatedAt, evaluatedTimeErr := time.Parse(time.RFC3339, evaluation.EvaluatedAt)
		// The only offline evaluation evidence is the fixture outcome that it
		// evaluates. Do not accept a checksum-valid evaluation which keeps the
		// outcome ID field but silently swaps its cited evidence.
		if !outcomeOK || !executionOK || outcomeTimeErr != nil || evaluatedTimeErr != nil || evaluatedAt.Before(outcomeAt) || outcome.EffectRef.EffectID != execution.ExecutionID || execution.ProductionLeaseID != evaluation.LeaseID || execution.ProductionLeaseVersion != evaluation.LeaseVersion || execution.ProductionLeaseHash != evaluation.LeaseHash || len(evaluation.EvidenceIDs) != 1 || evaluation.EvidenceIDs[0] != outcome.OutcomeID {
			return fmt.Errorf("M11 outcome evaluation does not resolve its fixture outcome and execution")
		}
	}
	for _, cycle := range cycles {
		evaluation, evaluationOK := evaluations[cycle.EvaluationID]
		execution, executionOK := executionsByID[cycle.ExecutionID]
		closedAt, closedTimeErr := time.Parse(time.RFC3339, cycle.ClosedAt)
		evaluatedAt, evaluatedTimeErr := time.Parse(time.RFC3339, evaluation.EvaluatedAt)
		record, recordOK := historyByID[cycle.DecisionID]
		observations := map[string]bool{}
		for _, observation := range record.Observations {
			observations[observation.ObservationID] = true
		}
		cycleObservations := map[string]bool{}
		for _, id := range cycle.ObservationIDs {
			if id == "" || cycleObservations[id] {
				return fmt.Errorf("M11 cycle has duplicate or empty canonical observation link")
			}
			cycleObservations[id] = true
		}
		if !evaluationOK || !executionOK || !recordOK || closedTimeErr != nil || evaluatedTimeErr != nil || cycle.Status != "CLOSED" || cycle.OpenedAt != execution.AttemptedAt || closedAt.Before(evaluatedAt) || evaluation.ExecutionID != cycle.ExecutionID || evaluation.OutcomeID != cycle.OutcomeID || evaluation.LeaseID != cycle.LeaseID || evaluation.LeaseVersion != cycle.LeaseVersion || evaluation.LeaseHash != cycle.LeaseHash || cycle.LeaseID != execution.ProductionLeaseID || cycle.LeaseVersion != execution.ProductionLeaseVersion || cycle.LeaseHash != execution.ProductionLeaseHash || cycle.IntentID != execution.IntentID || cycle.IntentHash != execution.IntentHash || cycle.GateID != execution.ProductionGateID || cycle.AuthorizationID != execution.AuthorizationID || cycle.CorrelationID != execution.CorrelationID || !reflect.DeepEqual(observations, cycleObservations) {
			return fmt.Errorf("M11 cycle does not resolve its outcome evaluation")
		}
	}
	return nil
}

func sameFileSet(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}

func containsAll(files, required []string) bool {
	present := map[string]bool{}
	for _, file := range files {
		present[file] = true
	}
	for _, file := range required {
		if !present[file] {
			return false
		}
	}
	return true
}

func verifyBackup(dir string) (backupManifest, error) {
	var m backupManifest
	// The manifest selects the inventory and expected digests for every later
	// restore read. Treat it as a backup-owned artifact boundary rather than a
	// portable input: a symlink or name/content race must fail before inventory
	// validation can make the external bytes authoritative.
	manifestRaw, _, err := readStableRegularFile(filepath.Join(dir, "manifest.json"))
	if err != nil {
		return m, err
	}
	decoder := json.NewDecoder(bytes.NewReader(manifestRaw))
	decoder.UseNumber()
	if err := decoder.Decode(&m); err != nil {
		return m, err
	}
	if m.Version != backupManifestVersion || len(m.Files) == 0 {
		return m, fmt.Errorf("invalid backup manifest")
	}
	actualFiles, err := backupFiles(dir)
	if err != nil {
		return m, fmt.Errorf("backup inventory is invalid: %w", err)
	}
	manifestFiles := make([]string, 0, len(m.Files))
	for name := range m.Files {
		manifestFiles = append(manifestFiles, name)
	}
	sort.Strings(manifestFiles)
	if !sameFileSet(manifestFiles, actualFiles) {
		return m, fmt.Errorf("backup artifact inventory does not exactly match manifest")
	}
	for name, want := range m.Files {
		clean, err := backupRelativePath(name)
		if err != nil {
			return m, err
		}
		kind, err := backupArtifactKind(name)
		if err != nil {
			return m, err
		}
		if want.Kind != kind || want.SizeBytes < 0 || want.SHA256 == "" {
			return m, fmt.Errorf("backup artifact metadata is invalid for %s", name)
		}
		p := filepath.Join(dir, clean)
		if err := regularFile(p); err != nil {
			return m, err
		}
		info, err := os.Stat(p)
		if err != nil || info.Size() != want.SizeBytes {
			return m, fmt.Errorf("backup size mismatch for %s", name)
		}
		got, e := fileDigest(p)
		if e != nil || got != want.SHA256 {
			return m, fmt.Errorf("backup checksum mismatch for %s", name)
		}
	}
	expectedRequired, err := requiredBackupFiles(dir)
	if err != nil {
		return m, fmt.Errorf("backup required artifact graph is invalid: %w", err)
	}
	sort.Strings(m.Required)
	if !sameFileSet(m.Required, expectedRequired) {
		return m, fmt.Errorf("backup required artifact inventory does not match runtime graph")
	}
	for _, required := range expectedRequired {
		if _, ok := m.Files[required]; !ok {
			return m, fmt.Errorf("backup is missing required %s", required)
		}
	}
	expectedProfile, err := backupProfileFor(expectedRequired, actualFiles)
	if err != nil {
		return m, err
	}
	if !reflect.DeepEqual(m.Profile, expectedProfile) {
		return m, fmt.Errorf("backup profile does not match artifact inventory")
	}
	if err := validateM07BackupGraph(dir); err != nil {
		return m, fmt.Errorf("backup M07 graph is invalid: %w", err)
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
		source, sourceErr := filepath.Abs(args[1])
		target, targetErr := filepath.Abs(args[2])
		if sourceErr != nil || targetErr != nil || source == target || strings.HasPrefix(target, source+string(filepath.Separator)) {
			return emit("INPUT_ERROR", nil, fmt.Errorf("backup target must not be the runtime or a child of it"), 1)
		}
		if info, statErr := os.Lstat(args[2]); statErr == nil {
			if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
				return emit("TARGET_ERROR", nil, fmt.Errorf("backup target must be a non-symlink directory"), 1)
			}
			entries, readErr := os.ReadDir(args[2])
			if readErr != nil {
				return emit("TARGET_ERROR", nil, readErr, 1)
			}
			if len(entries) != 0 {
				return emit("TARGET_NOT_EMPTY", nil, fmt.Errorf("backup target must be empty"), 1)
			}
		} else if !os.IsNotExist(statErr) {
			return emit("TARGET_ERROR", nil, statErr, 1)
		}
		release, lockErr := acquireRuntimeGate(args[1])
		if lockErr != nil {
			return emit("BUSY", nil, lockErr, 1)
		}
		defer release()
		if e := recoverM10ExecutionJournal(args[1]); e != nil {
			return emit("RECOVERY_REQUIRED", nil, e, 1)
		}
		if e := recoverM11FailedExecutionJournal(args[1]); e != nil {
			return emit("RECOVERY_REQUIRED", nil, e, 1)
		}
		if e := recoverM11UnknownStopJournal(args[1]); e != nil {
			return emit("RECOVERY_REQUIRED", nil, e, 1)
		}
		if e := recoverM11OutcomeJournal(args[1]); e != nil {
			return emit("RECOVERY_REQUIRED", nil, e, 1)
		}
		if e := validateAccesstradeBackupGraph(args[1]); e != nil {
			return emit("INPUT_ERROR", nil, fmt.Errorf("runtime ACCESSTRADE receipt graph is invalid: %w", e), 1)
		}
		if e := validateM10BackupGraph(args[1]); e != nil {
			return emit("INPUT_ERROR", nil, fmt.Errorf("runtime M10 graph is invalid: %w", e), 1)
		}
		files, e := backupFiles(args[1])
		if e != nil {
			return emit("INPUT_ERROR", nil, e, 1)
		}
		sourceInventory, e := backupSourceInventory(args[1], files)
		if e != nil {
			return emit("INPUT_ERROR", nil, e, 1)
		}
		parent := filepath.Dir(args[2])
		if e = os.MkdirAll(parent, 0700); e != nil {
			return emit("STORE_ERROR", nil, e, 1)
		}
		parentInfo, statErr := os.Lstat(parent)
		if statErr != nil || !parentInfo.IsDir() || parentInfo.Mode()&os.ModeSymlink != 0 {
			return emit("TARGET_ERROR", nil, fmt.Errorf("backup target parent must be a non-symlink directory"), 1)
		}
		releaseTargetGate, gateErr := acquireBackupTargetGate(args[2])
		if gateErr != nil {
			return emit("BUSY", nil, gateErr, 1)
		}
		defer releaseTargetGate()
		// Recheck after claiming the cooperative target publisher gate. A caller
		// may intentionally pre-create an empty target, but never a populated or
		// symlink target.
		if info, targetErr := os.Lstat(args[2]); targetErr == nil {
			if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
				return emit("TARGET_ERROR", nil, fmt.Errorf("backup target must be a non-symlink directory"), 1)
			}
			entries, readErr := os.ReadDir(args[2])
			if readErr != nil {
				return emit("TARGET_ERROR", nil, readErr, 1)
			}
			if len(entries) != 0 {
				return emit("TARGET_NOT_EMPTY", nil, fmt.Errorf("backup target must be empty"), 1)
			}
		} else if !os.IsNotExist(targetErr) {
			return emit("TARGET_ERROR", nil, targetErr, 1)
		}
		staging, stageErr := os.MkdirTemp(parent, ".backup-staging-")
		if stageErr != nil {
			return emit("STORE_ERROR", nil, stageErr, 1)
		}
		if e = syncDirectory(parent); e != nil {
			_ = os.RemoveAll(staging)
			return emit("STORE_ERROR", nil, e, 1)
		}
		published := false
		defer func() {
			if !published {
				_ = os.RemoveAll(staging)
			}
		}()
		required, e := requiredBackupFiles(args[1])
		if e != nil {
			return emit("INPUT_ERROR", nil, e, 1)
		}
		if !containsAll(files, required) {
			return emit("INPUT_ERROR", nil, fmt.Errorf("runtime is missing required backup artifacts"), 1)
		}
		if e = validateM10BackupGraph(args[1]); e != nil {
			return emit("INPUT_ERROR", nil, fmt.Errorf("runtime M10 graph is invalid: %w", e), 1)
		}
		if e = validateM07BackupGraph(args[1]); e != nil {
			return emit("INPUT_ERROR", nil, fmt.Errorf("runtime M07 graph is invalid: %w", e), 1)
		}
		if e = validateM11BackupGraph(args[1]); e != nil {
			return emit("INPUT_ERROR", nil, e, 1)
		}
		verifiedSourceInventory, inventoryErr := backupSourceInventory(args[1], files)
		if inventoryErr != nil || !sameBackupInventory(sourceInventory, verifiedSourceInventory) {
			return emit("SNAPSHOT_CONFLICT", nil, fmt.Errorf("runtime changed while backup snapshot was prepared"), 1)
		}
		profile, e := backupProfileFor(required, files)
		if e != nil {
			return emit("INPUT_ERROR", nil, e, 1)
		}
		m := backupManifest{Version: backupManifestVersion, Profile: profile, Files: map[string]backupFile{}, Required: required}
		for _, name := range files {
			if backupCopyFault != nil {
				if e = backupCopyFault(name); e != nil {
					return emit("STORE_ERROR", nil, e, 1)
				}
			}
			b, sourceMetadata, readErr := backupFileDataAndMetadata(filepath.Join(args[1], name), name)
			if readErr != nil {
				return emit("INPUT_ERROR", nil, readErr, 1)
			}
			if sourceMetadata != sourceInventory[name] {
				return emit("SNAPSHOT_CONFLICT", nil, fmt.Errorf("runtime changed while reading %s", name), 1)
			}
			clean, e := backupRelativePath(name)
			if e != nil {
				return emit("INPUT_ERROR", nil, e, 1)
			}
			target := filepath.Join(staging, clean)
			if e = os.MkdirAll(filepath.Dir(target), 0700); e != nil {
				return emit("STORE_ERROR", nil, e, 1)
			}
			if e = writeBackupStagingFile(staging, target, b); e != nil {
				return emit("STORE_ERROR", nil, e, 1)
			}
			metadata, metadataErr := backupFileMetadata(target, name)
			if metadataErr != nil {
				return emit("STORE_ERROR", nil, metadataErr, 1)
			}
			if metadata != sourceInventory[name] {
				return emit("SNAPSHOT_CONFLICT", nil, fmt.Errorf("runtime changed while copying %s", name), 1)
			}
			m.Files[name] = metadata
		}
		if e = writeJSONAtomic(filepath.Join(staging, "manifest.json"), m); e != nil {
			return emit("STORE_ERROR", nil, e, 1)
		}
		if _, e = verifyBackup(staging); e != nil {
			return emit("STORE_ERROR", nil, fmt.Errorf("staged backup verification failed: %w", e), 1)
		}
		if _, targetErr := os.Lstat(args[2]); targetErr == nil {
			entries, readErr := os.ReadDir(args[2])
			if readErr != nil || len(entries) != 0 {
				if readErr != nil {
					return emit("TARGET_ERROR", nil, readErr, 1)
				}
				return emit("TARGET_NOT_EMPTY", nil, fmt.Errorf("backup target changed while staging"), 1)
			}
			if e = os.Remove(args[2]); e != nil {
				return emit("STORE_ERROR", nil, fmt.Errorf("prepare empty backup target for publish: %w", e), 1)
			}
		} else if !os.IsNotExist(targetErr) {
			return emit("TARGET_ERROR", nil, targetErr, 1)
		}
		if e = os.Rename(staging, args[2]); e != nil {
			return emit("STORE_ERROR", nil, fmt.Errorf("publish backup snapshot: %w", e), 1)
		}
		parentFile, parentErr := os.Open(parent)
		if parentErr != nil {
			return emit("STORE_ERROR", nil, parentErr, 1)
		}
		syncErr := parentFile.Sync()
		closeErr := parentFile.Close()
		if syncErr != nil {
			return emit("STORE_ERROR", nil, syncErr, 1)
		}
		if closeErr != nil {
			return emit("STORE_ERROR", nil, closeErr, 1)
		}
		published = true
		return emit("BACKED_UP", m, nil, 0)
	}
	m, e := verifyBackup(args[1])
	if e != nil {
		return emit("VERIFY_FAILED", nil, e, 1)
	}
	// A restore is not a usable runtime until every loader and cross-store
	// validator has accepted it. Materialize into an owned sibling directory so
	// a failed replay/graph check never publishes a partial target.
	if _, statErr := os.Lstat(args[2]); statErr == nil {
		return emit("TARGET_NOT_EMPTY", nil, fmt.Errorf("restore target must not already exist"), 1)
	} else if !os.IsNotExist(statErr) {
		return emit("TARGET_ERROR", nil, statErr, 1)
	}
	parent := filepath.Dir(args[2])
	if e = os.MkdirAll(parent, 0700); e != nil {
		return emit("STORE_ERROR", nil, e, 1)
	}
	parentInfo, statErr := os.Lstat(parent)
	if statErr != nil || !parentInfo.IsDir() || parentInfo.Mode()&os.ModeSymlink != 0 {
		return emit("TARGET_ERROR", nil, fmt.Errorf("restore target parent must be a non-symlink directory"), 1)
	}
	releaseTargetGate, gateErr := acquireRestoreTargetGate(args[2])
	if gateErr != nil {
		return emit("BUSY", nil, gateErr, 1)
	}
	defer releaseTargetGate()
	// Claiming the per-target gate precedes staging and the final existence
	// check, so two managed restores cannot race into a clobbering Rename.
	if _, statErr := os.Lstat(args[2]); statErr == nil {
		return emit("TARGET_NOT_EMPTY", nil, fmt.Errorf("restore target appeared while acquiring its gate"), 1)
	} else if !os.IsNotExist(statErr) {
		return emit("TARGET_ERROR", nil, statErr, 1)
	}
	staging, stageErr := os.MkdirTemp(parent, ".restore-staging-")
	if stageErr != nil {
		return emit("STORE_ERROR", nil, stageErr, 1)
	}
	if e = syncDirectory(parent); e != nil {
		_ = os.RemoveAll(staging)
		return emit("STORE_ERROR", nil, e, 1)
	}
	published := false
	defer func() {
		if !published {
			_ = os.RemoveAll(staging)
		}
	}()
	fileNames := make([]string, 0, len(m.Files))
	for name := range m.Files {
		fileNames = append(fileNames, name)
	}
	sort.Strings(fileNames)
	for _, name := range fileNames {
		clean, pathErr := backupRelativePath(name)
		if pathErr != nil {
			return emit("VERIFY_FAILED", nil, pathErr, 1)
		}
		if restoreCopyFault != nil {
			if e = restoreCopyFault(name); e != nil {
				return emit("STORE_ERROR", nil, e, 1)
			}
		}
		b, sourceMetadata, readErr := backupFileDataAndMetadata(filepath.Join(args[1], clean), name)
		if readErr != nil {
			return emit("VERIFY_FAILED", nil, readErr, 1)
		}
		if sourceMetadata != m.Files[name] {
			return emit("VERIFY_FAILED", nil, fmt.Errorf("backup source changed while restoring %s", name), 1)
		}
		target := filepath.Join(staging, clean)
		if e = os.MkdirAll(filepath.Dir(target), 0700); e != nil {
			return emit("STORE_ERROR", nil, e, 1)
		}
		if e = writeBackupStagingFile(staging, target, b); e != nil {
			return emit("STORE_ERROR", nil, e, 1)
		}
	}
	if _, ok := m.Files["history.jsonl"]; ok {
		var records []HistoryRecord
		if records, e = LoadHistory(filepath.Join(staging, "history.jsonl")); e != nil {
			return emit("REPLAY_FAILED", nil, e, 1)
		}
		for _, record := range records {
			if replay := Replay(record); replay.State != replayMatch {
				return emit("REPLAY_FAILED", nil, fmt.Errorf("record %s replayed as %s: %s", record.RecordID, replay.State, replay.Reason), 1)
			}
		}
	}
	if _, ok := m.Files["mission-state.json"]; ok {
		if _, e = loadMissionState(staging); e != nil {
			return emit("STATE_FAILED", nil, e, 1)
		}
	}
	if _, e = m00ToM05BackupFiles(staging); e != nil {
		return emit("GRAPH_FAILED", nil, e, 1)
	}
	if e = validateM10BackupGraph(staging); e != nil {
		return emit("GRAPH_FAILED", nil, e, 1)
	}
	if e = validateM07BackupGraph(staging); e != nil {
		return emit("GRAPH_FAILED", nil, e, 1)
	}
	if e = validateM11BackupGraph(staging); e != nil {
		return emit("GRAPH_FAILED", nil, e, 1)
	}
	if e = validateAccesstradeBackupGraph(staging); e != nil {
		return emit("GRAPH_FAILED", nil, fmt.Errorf("restored ACCESSTRADE receipt graph is invalid: %w", e), 1)
	}
	if _, statErr := os.Lstat(args[2]); statErr == nil {
		return emit("TARGET_NOT_EMPTY", nil, fmt.Errorf("restore target appeared while staging"), 1)
	} else if !os.IsNotExist(statErr) {
		return emit("TARGET_ERROR", nil, statErr, 1)
	}
	if e = os.Rename(staging, args[2]); e != nil {
		return emit("STORE_ERROR", nil, fmt.Errorf("publish restored runtime: %w", e), 1)
	}
	published = true
	return emit("RESTORED", m, nil, 0)
}
