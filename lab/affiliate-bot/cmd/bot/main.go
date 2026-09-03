package main

import (
    "encoding/json"
    "fmt"
    "os"
    "sort"
)

type Observation struct {
    ProductID      string   `json:"product_id"`
    ProductName    string   `json:"product_name"`
    Price          *float64 `json:"price"`
    CommissionRate *float64 `json:"commission_rate"`
    Currency       string   `json:"currency"`
    EvidenceKind   string   `json:"evidence_kind"`
}

type Ranked struct {
    ProductName string
    Score       float64
}

type Result struct {
    State           string
    Ranked          []Ranked
    MissingEvidence []string
    EvidenceMode    string
}

func evaluate(records []Observation) Result {
    result := Result{State: "RANK_SCENARIO"}
    if len(records) == 0 {
        result.State = "GET_MORE_DATA"
        result.MissingEvidence = append(result.MissingEvidence, "không có observation")
        return result
    }

    kinds := map[string]bool{}
    currencies := map[string]bool{}
    seen := map[string]bool{}

    for _, r := range records {
        if r.ProductID == "" || r.ProductName == "" {
            result.State = "GET_MORE_DATA"
            result.MissingEvidence = append(result.MissingEvidence, "thiếu identity")
            continue
        }
        if seen[r.ProductID] {
            result.State = "HUMAN_REVIEW"
        }
        seen[r.ProductID] = true
        kinds[r.EvidenceKind] = true
        currencies[r.Currency] = true
        if r.Price == nil || r.CommissionRate == nil {
            result.State = "GET_MORE_DATA"
            result.MissingEvidence = append(result.MissingEvidence, r.ProductID+": thiếu price/commission_rate")
            continue
        }
        result.Ranked = append(result.Ranked, Ranked{ProductName: r.ProductName, Score: *r.Price * *r.CommissionRate})
    }

    if len(kinds) == 1 {
        for k := range kinds {
            result.EvidenceMode = k
        }
    } else {
        result.EvidenceMode = "mixed"
        result.State = "HUMAN_REVIEW"
    }
    if len(currencies) > 1 {
        result.State = "HUMAN_REVIEW"
    }

    sort.SliceStable(result.Ranked, func(i, j int) bool {
        if result.Ranked[i].Score == result.Ranked[j].Score {
            return result.Ranked[i].ProductName < result.Ranked[j].ProductName
        }
        return result.Ranked[i].Score > result.Ranked[j].Score
    })
    return result
}

func main() {
    path := "data/sample-observations.json"
    if len(os.Args) > 1 {
        path = os.Args[1]
    }
    raw, err := os.ReadFile(path)
    if err != nil {
        panic(err)
    }
    var records []Observation
    if err := json.Unmarshal(raw, &records); err != nil {
        panic(err)
    }
    result := evaluate(records)

    fmt.Println("Affiliate Bot v0.1 — deterministic baseline")
    fmt.Printf("Evidence mode: %s\n", result.EvidenceMode)
    for i, item := range result.Ranked {
        fmt.Printf("%d. %s | score=%.2f\n", i+1, item.ProductName, item.Score)
    }
    fmt.Printf("Decision state: %s\n", result.State)
    fmt.Println("Baseline limitation: score hiện chỉ dùng price × commission_rate; chưa chứng minh product tốt nhất cho Affiliate.")
    fmt.Println("Authority boundary: real evidence không tự nâng RANK_SCENARIO thành RECOMMEND; output không phải approval hay execution permission.")
}
