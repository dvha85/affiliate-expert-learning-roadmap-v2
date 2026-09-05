package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
)

const (
	formulaVersion   = "commission-per-order/v1"
	stateRank        = "RANK_SCENARIO"
	stateGetMoreData = "GET_MORE_DATA"
	stateHumanReview = "HUMAN_REVIEW"
)

type Observation struct {
	ObservationID          string   `json:"observation_id,omitempty"`
	SubjectID              string   `json:"subject_id,omitempty"`
	SourceURL              string   `json:"source_url,omitempty"`
	SourceRef              string   `json:"source_ref,omitempty"`
	ObservedAt             string   `json:"observed_at,omitempty"`
	AccessMethod           string   `json:"access_method,omitempty"`
	EvidenceKind           string   `json:"evidence_kind"`
	UseContext             string   `json:"use_context,omitempty"`
	ClaimKind              string   `json:"claim_kind,omitempty"`
	State                  string   `json:"state,omitempty"`
	SourceAuthorityOrRole  string   `json:"source_authority_or_role,omitempty"`
	TransformationOrMethod string   `json:"transformation_or_method,omitempty"`
	CorrelationID          string   `json:"correlation_id,omitempty"`
	Limitation             string   `json:"limitation,omitempty"`
	ProductID              string   `json:"product_id"`
	ProductName            string   `json:"product_name"`
	Price                  *float64 `json:"price"`
	CommissionRate         *float64 `json:"commission_rate"`
	Currency               string   `json:"currency"`
}

type Ranked struct {
	ProductID   string  `json:"product_id"`
	ProductName string  `json:"product_name"`
	Score       float64 `json:"score"`
}

type Result struct {
	DecisionID      string   `json:"decision_id,omitempty"`
	EvidenceIDs     []string `json:"evidence_ids,omitempty"`
	FormulaVersion  string   `json:"formula_version"`
	State           string   `json:"state"`
	Ranked          []Ranked `json:"ranked"`
	MissingEvidence []string `json:"missing_evidence"`
	Reasons         []string `json:"reasons"`
	EvidenceMode    string   `json:"evidence_mode"`
}

func evaluate(records []Observation) Result {
	result := Result{FormulaVersion: formulaVersion, State: stateRank, EvidenceMode: "unknown"}
	if len(records) == 0 {
		result.State = stateGetMoreData
		result.MissingEvidence = append(result.MissingEvidence, "không có observation")
		result.Reasons = append(result.Reasons, "không thể xếp hạng khi chưa có input")
		return result
	}

	kinds := map[string]bool{}
	currencies := map[string]bool{}
	seen := map[string]bool{}
	needsData := false
	needsReview := false

	for _, r := range records {
		id := strings.TrimSpace(r.ProductID)
		name := strings.TrimSpace(r.ProductName)
		currency := strings.TrimSpace(r.Currency)
		kind := strings.TrimSpace(r.EvidenceKind)

		if id == "" || name == "" {
			needsData = true
			result.MissingEvidence = append(result.MissingEvidence, "thiếu product_id hoặc product_name")
			continue
		}
		if seen[id] {
			needsReview = true
			result.Reasons = append(result.Reasons, id+": product_id bị lặp; cần kiểm tra identity")
		}
		seen[id] = true

		switch kind {
		case "real", "synthetic":
			kinds[kind] = true
		default:
			needsData = true
			result.MissingEvidence = append(result.MissingEvidence, id+": evidence_kind phải là real hoặc synthetic")
		}

		if currency == "" {
			needsData = true
			result.MissingEvidence = append(result.MissingEvidence, id+": thiếu currency")
		} else {
			currencies[currency] = true
		}

		validNumeric := true
		if r.Price == nil {
			needsData = true
			validNumeric = false
			result.MissingEvidence = append(result.MissingEvidence, id+": thiếu price")
		} else if *r.Price < 0 {
			needsData = true
			validNumeric = false
			result.MissingEvidence = append(result.MissingEvidence, id+": price không được âm")
		}

		if r.CommissionRate == nil {
			needsData = true
			validNumeric = false
			result.MissingEvidence = append(result.MissingEvidence, id+": thiếu commission_rate")
		} else if *r.CommissionRate < 0 || *r.CommissionRate > 1 {
			needsData = true
			validNumeric = false
			result.MissingEvidence = append(result.MissingEvidence, id+": commission_rate phải nằm trong [0,1]")
		}

		if validNumeric {
			result.Ranked = append(result.Ranked, Ranked{ProductID: id, ProductName: name, Score: *r.Price * *r.CommissionRate})
		}
	}

	if len(kinds) == 1 {
		for kind := range kinds {
			result.EvidenceMode = kind
		}
	} else if len(kinds) > 1 {
		result.EvidenceMode = "mixed"
		needsReview = true
		result.Reasons = append(result.Reasons, "đang trộn real và synthetic evidence trong cùng ranking context")
	}

	if len(currencies) > 1 {
		needsReview = true
		result.Reasons = append(result.Reasons, "currency không đồng nhất; score chưa so sánh được trực tiếp")
	}

	sort.Slice(result.Ranked, func(i, j int) bool {
		if result.Ranked[i].Score != result.Ranked[j].Score {
			return result.Ranked[i].Score > result.Ranked[j].Score
		}
		if result.Ranked[i].ProductName != result.Ranked[j].ProductName {
			return result.Ranked[i].ProductName < result.Ranked[j].ProductName
		}
		return result.Ranked[i].ProductID < result.Ranked[j].ProductID
	})

	switch {
	case needsReview:
		result.State = stateHumanReview
	case needsData:
		result.State = stateGetMoreData
		result.Reasons = append(result.Reasons, "input còn thiếu hoặc không hợp lệ; cần bổ sung trước khi tin vào ranking")
	default:
		result.State = stateRank
		result.Reasons = append(result.Reasons, "baseline chỉ xếp hạng scenario bằng price × commission_rate")
	}
	return result
}

func loadObservations(path string) ([]Observation, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var records []Observation
	if err := json.Unmarshal(raw, &records); err != nil {
		return nil, err
	}
	return records, nil
}

func runHistory(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: history capture|list|replay ...")
	}
	switch args[1] {
	case "decision":
		return exportHistoryDecision(os.Stdout, args[2:])
	case "capture":
		if len(args) != 7 {
			return fmt.Errorf("usage: history capture <history.jsonl> <observations.json> <record_id> <as_of> <ingested_at>")
		}
		observations, err := loadHistoryObservations(args[3])
		if err != nil {
			return err
		}
		record, err := NewHistoryRecord(args[4], args[5], args[6], observations)
		if err != nil {
			return err
		}
		state, err := AppendHistory(args[2], record)
		if err != nil {
			return err
		}
		fmt.Printf("History append: %s | record_id=%s | input_hash=%s\n", state, record.RecordID, record.InputHash)
		return nil
	case "list":
		if len(args) != 3 {
			return fmt.Errorf("usage: history list <history.jsonl>")
		}
		records, err := LoadHistory(args[2])
		if err != nil {
			return err
		}
		for _, record := range records {
			fmt.Printf("%s | as_of=%s | ingested_at=%s | formula=%s | state=%s\n", record.RecordID, record.AsOf, record.IngestedAt, record.FormulaVersion, record.RecordedResult.State)
		}
		return nil
	case "replay":
		if len(args) != 3 {
			return fmt.Errorf("usage: history replay <history.jsonl>")
		}
		records, err := LoadHistory(args[2])
		if err != nil {
			return err
		}
		for _, record := range records {
			report := Replay(record)
			fmt.Printf("%s | replay=%s | %s\n", report.RecordID, report.State, report.Reason)
		}
		return nil
	default:
		return fmt.Errorf("unknown history command %q", args[1])
	}
}

func main() {
	var err error
	if len(os.Args) > 1 && os.Args[1] == "history" {
		err = runHistory(os.Args[1:])
	} else {
		path := "data/sample-observations.json"
		if len(os.Args) > 1 {
			path = os.Args[1]
		}
		records, loadErr := loadObservations(path)
		if loadErr != nil {
			err = loadErr
		} else {
			result := evaluate(records)
			fmt.Println("Affiliate Bot v0.1 — deterministic baseline")
			fmt.Printf("Phiên bản công thức (Formula version): %s\n", result.FormulaVersion)
			fmt.Printf("Loại bằng chứng (Evidence mode): %s\n", result.EvidenceMode)
			for i, item := range result.Ranked {
				fmt.Printf("%d. %s [%s] | điểm (score)=%.2f\n", i+1, item.ProductName, item.ProductID, item.Score)
			}
			fmt.Printf("Trạng thái quyết định (Decision state): %s\n", result.State)
			for _, reason := range result.Reasons {
				fmt.Printf("- Lý do (Reason): %s\n", reason)
			}
			for _, missing := range result.MissingEvidence {
				fmt.Printf("- Bằng chứng còn thiếu/không hợp lệ (Missing/invalid evidence): %s\n", missing)
			}
			fmt.Println("Giới hạn baseline: score hiện chỉ dùng price × commission_rate; chưa chứng minh product tốt nhất cho Affiliate.")
			fmt.Println("Authority boundary: real evidence không tự nâng RANK_SCENARIO thành RECOMMEND; output không phải approval hay execution permission.")
		}
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "Bot không thể tiếp tục: %v\n", err)
		os.Exit(1)
	}
}
