package main

import "fmt"

// resolveCanonicalRecord is the learner's single read-only boundary from
// canonical history into downstream M07/M08 adapters. It never accepts a
// caller-provided snapshot in place of the stored, replay-stable record.
func resolveCanonicalRecord(historyPath, recordID string) (HistoryRecord, error) {
	history, err := LoadHistory(historyPath)
	if err != nil {
		return HistoryRecord{}, err
	}
	var found HistoryRecord
	count := 0
	for _, record := range history {
		if record.RecordID == recordID {
			found = record
			count++
		}
	}
	if count != 1 {
		return HistoryRecord{}, fmt.Errorf("record_id must resolve exactly once (found %d)", count)
	}
	if replay := Replay(found); replay.State != replayMatch {
		return HistoryRecord{}, fmt.Errorf("record_id is not replay-stable: %s", replay.State)
	}
	return found, nil
}
