package main

import (
	"encoding/json"
	"fmt"
)

func decodeProductionLedger(raw []byte, lease ProductionLease) (ProductionLedger, error) {
	v, status := DecodeM11Artifact("ledger", raw)
	if status != missionValid {
		return ProductionLedger{}, fmt.Errorf("M11 persisted ledger: %s", status)
	}
	l := *v.(*ProductionLedger)
	if l.LeaseID != lease.LeaseID || l.LeaseVersion != lease.LeaseVersion || l.LeaseHash != lease.LeaseHash {
		return ProductionLedger{}, fmt.Errorf("M11 ledger lease mismatch")
	}
	return l, nil
}

func encodeProductionArtifact(kind string, value any) ([]byte, error) {
	b, e := json.MarshalIndent(value, "", "  ")
	if e != nil {
		return nil, e
	}
	if _, status := DecodeM11Artifact(kind, b); status != missionValid {
		return nil, fmt.Errorf("M11 persisted %s: %s", kind, status)
	}
	return b, nil
}
