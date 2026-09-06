package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dvha85/affiliate-expert-learning-roadmap-v2/core/m00"
)

func importedReviewRecord(t *testing.T, count int, question string) HistoryRecord {
	t.Helper()
	raw, err := os.ReadFile("../../../../examples/m00-import/packet-t2.json")
	if err != nil {
		t.Fatal(err)
	}
	var packet m00.Packet
	if err = json.Unmarshal(raw, &packet); err != nil {
		t.Fatal(err)
	}
	original := packet.Products[0]
	packet.Products = nil
	if question != "" {
		packet.Question = question
	}
	for i := 0; i < count; i++ {
		product := original
		product.Fields = append([]m00.Field(nil), original.Fields...)
		product.SubjectID = fmt.Sprintf("product-%d", i)
		product.ObservationID = fmt.Sprintf("aggregate-%d", i)
		for j := range product.Fields {
			product.Fields[j].SubjectID = product.SubjectID
			product.Fields[j].ObservationID = fmt.Sprintf("field-%d-%d", i, j)
		}
		product.Fields[1].ObservedAt = "2026-09-02T07:00:00+07:00"
		packet.Products = append(packet.Products, product)
	}
	raw, _ = json.Marshal(packet)
	converted, err := m00.Convert(raw)
	if err != nil {
		t.Fatal(err)
	}
	raw, _ = json.Marshal(converted)
	var observations []Observation
	if err = json.Unmarshal(raw, &observations); err != nil {
		t.Fatal(err)
	}
	record, err := NewHistoryRecord("review-record", "2026-09-03T00:00:00Z", "2026-09-03T00:00:00Z", observations)
	if err != nil {
		t.Fatal(err)
	}
	return record
}

func TestImportedLargeHistoryRoundTrip(t *testing.T) {
	record := importedReviewRecord(t, 40, "")
	encoded, _ := json.Marshal(record)
	if len(encoded) <= 65536 {
		t.Fatal("probe too small")
	}
	path := filepath.Join(t.TempDir(), "history.jsonl")
	if state, err := AppendHistory(path, record); err != nil || state != appendAdded {
		t.Fatal(state, err)
	}
	before, _ := os.ReadFile(path)
	loaded, err := LoadHistory(path)
	if err != nil || len(loaded) != 1 {
		t.Fatal(err)
	}
	if Replay(loaded[0]).State != replayMatch {
		t.Fatal("replay mismatch")
	}
	if state, err := AppendHistory(path, record); err != nil || state != appendDuplicate {
		t.Fatal(state, err)
	}
	after, _ := os.ReadFile(path)
	if !bytes.Equal(before, after) {
		t.Fatal("history rewritten")
	}
}

func TestOversizedHistoryRejectedWithoutWrite(t *testing.T) {
	large := importedReviewRecord(t, 1, strings.Repeat("q", 1<<20))
	for _, existing := range []bool{false, true} {
		t.Run(fmt.Sprint(existing), func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "history.jsonl")
			if existing {
				seed := importedReviewRecord(t, 1, "")
				seed.RecordID = "seed"
				seed.RecordedResult.DecisionID = "seed"
				if _, err := AppendHistory(path, seed); err != nil {
					t.Fatal(err)
				}
			}
			before, _ := os.ReadFile(path)
			if _, err := AppendHistory(path, large); err == nil || !strings.Contains(err.Error(), "exceeds") {
				t.Fatal("expected size rejection", err)
			}
			after, err := os.ReadFile(path)
			if !existing && !os.IsNotExist(err) {
				t.Fatal("created file on rejection")
			}
			if !bytes.Equal(before, after) {
				t.Fatal("existing bytes changed")
			}
		})
	}
}

func TestHistoryPayloadBoundary(t *testing.T) {
	const limit = 1 << 20
	observations, err := loadHistoryObservations("../../data/m02-sample-observations.json")
	if err != nil {
		t.Fatal(err)
	}
	observations = observations[:1]
	observations[0].Limitation = "x"
	makeRecord := func() HistoryRecord {
		r, e := NewHistoryRecord("boundary", "2026-09-03T00:00:00Z", "2026-09-03T00:00:00Z", observations)
		if e != nil {
			t.Fatal(e)
		}
		return r
	}
	initial, _ := json.Marshal(makeRecord())
	for _, size := range []int{limit - 1, limit, limit + 1} {
		t.Run(fmt.Sprint(size), func(t *testing.T) {
			observations[0].Limitation = strings.Repeat("x", 1+size-len(initial))
			record := makeRecord()
			raw, _ := json.Marshal(record)
			if len(raw) != size {
				t.Fatal(len(raw), size)
			}
			path := filepath.Join(t.TempDir(), "append.jsonl")
			_, err := AppendHistory(path, record)
			if size > limit {
				if err == nil {
					t.Fatal("oversize accepted")
				}
			} else if err != nil {
				t.Fatal(err)
			}
			for _, ending := range []string{"\n", "\r\n", ""} {
				file := filepath.Join(t.TempDir(), "read.jsonl")
				if e := os.WriteFile(file, append(append([]byte(nil), raw...), []byte(ending)...), 0600); e != nil {
					t.Fatal(e)
				}
				records, e := LoadHistory(file)
				if size > limit {
					if e == nil {
						t.Fatal("oversize read accepted")
					}
				} else if e != nil || len(records) != 1 || Replay(records[0]).State != replayMatch {
					t.Fatal("boundary read", ending, e)
				}
			}
		})
	}
}
