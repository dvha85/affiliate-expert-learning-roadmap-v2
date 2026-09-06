// Package m00 converts an explicit M00 JSON transcription into M02 input.
// It does not fetch sources, infer missing values, or write a store.
package m00

import (
	"encoding/json"
	"fmt"
	"math/big"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/dvha85/affiliate-expert-learning-roadmap-v2/contracts"
)

type Packet struct {
	Version  string    `json:"version"`
	Question string    `json:"question"`
	Products []Product `json:"products"`
}

// SourceFields validates the complete derived projection before exposing its
// original field IDs for history conflict detection. Legacy inputs return none.
func SourceFields(raw []byte) ([]Field, error) {
	value, err := contracts.Decode(raw)
	if err != nil {
		return nil, err
	}
	object, ok := value.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("observation must be object")
	}
	if object["access_method"] != "local_packet_conversion" {
		return nil, nil
	}
	encoded, ok := object["transformation_or_method"].(string)
	if !ok {
		return nil, fmt.Errorf("missing import provenance")
	}
	var source struct {
		Version  string  `json:"version"`
		Question string  `json:"question"`
		Product  Product `json:"product"`
	}
	if err := contracts.DecodeStrict([]byte(encoded), &source); err != nil {
		return nil, err
	}
	packet, _ := json.Marshal(Packet{source.Version, source.Question, []Product{source.Product}})
	converted, err := Convert(packet)
	if err != nil {
		return nil, err
	}
	expectedRaw, _ := json.Marshal(converted[0])
	expected, err := contracts.Decode(expectedRaw)
	if err != nil {
		return nil, err
	}
	if !reflect.DeepEqual(value, expected) {
		return nil, fmt.Errorf("import projection/provenance mismatch")
	}
	return source.Product.Fields, nil
}

type Product struct {
	ObservationID string  `json:"observation_id"`
	SubjectID     string  `json:"subject_id"`
	ProductName   string  `json:"product_name"`
	Currency      string  `json:"currency"`
	Fields        []Field `json:"fields"`
}

type Field struct {
	ObservationID string   `json:"observation_id"`
	SubjectID     string   `json:"subject_id"`
	SourceURL     string   `json:"source_url,omitempty"`
	SourceRef     string   `json:"source_ref,omitempty"`
	ObservedAt    string   `json:"observed_at"`
	AccessMethod  string   `json:"access_method"`
	EvidenceKind  string   `json:"evidence_kind"`
	UseContext    string   `json:"use_context,omitempty"`
	Field         string   `json:"field_or_claim"`
	ClaimKind     string   `json:"claim_kind"`
	Value         *float64 `json:"value"`
	State         string   `json:"state"`
	Role          string   `json:"source_authority_or_role"`
	Method        string   `json:"transformation_or_method"`
	Limitation    string   `json:"limitation"`
}

// Convert requires one explicit price and commission field per product snapshot.
// Non-observed/unknown fields remain in provenance, but project to null.
// The serialized source packet is retained in transformation_or_method so the
// existing M02 typed projection and hash preserve it without a store migration.
func Convert(raw []byte) ([]map[string]any, error) {
	var p Packet
	if err := contracts.DecodeStrict(raw, &p); err != nil {
		return nil, err
	}
	// M02 uses float64. Reject values whose decimal meaning would change during
	// that projection (including underflow), instead of silently rewriting source.
	var numbers struct {
		Products []struct {
			Fields []struct {
				Value json.RawMessage `json:"value"`
			} `json:"fields"`
		} `json:"products"`
	}
	if err := json.Unmarshal(raw, &numbers); err != nil {
		return nil, err
	}
	for i, product := range p.Products {
		for j, field := range product.Fields {
			if field.Value == nil {
				continue
			}
			decimal := strings.TrimSpace(string(numbers.Products[i].Fields[j].Value))
			// Bound decimal work before big.Rat handles adversarial exponents.
			if len(decimal) > 128 {
				return nil, fmt.Errorf("numeric literal exceeds import profile limit")
			}
			if at := strings.IndexAny(decimal, "eE"); at >= 0 {
				exponent, err := strconv.Atoi(decimal[at+1:])
				if err != nil || exponent < -400 || exponent > 400 {
					return nil, fmt.Errorf("numeric exponent exceeds import profile limit")
				}
			}
			original, ok := new(big.Rat).SetString(decimal)
			if !ok {
				return nil, fmt.Errorf("invalid decimal value")
			}
			projectedRaw, err := json.Marshal(*field.Value)
			if err != nil {
				return nil, err
			}
			projected, ok := new(big.Rat).SetString(string(projectedRaw))
			if !ok || original.Cmp(projected) != 0 {
				return nil, fmt.Errorf("value loses precision in M02 numeric projection")
			}
		}
	}
	if p.Version != "m00-input/v1" || strings.TrimSpace(p.Question) == "" || len(p.Products) == 0 {
		return nil, fmt.Errorf("version, question and products required")
	}
	ids, subjects := map[string]bool{}, map[string]bool{}
	for _, product := range p.Products {
		if strings.TrimSpace(product.ObservationID) == "" || ids[product.ObservationID] {
			return nil, fmt.Errorf("duplicate/empty aggregate observation_id")
		}
		ids[product.ObservationID] = true
	}
	out := []map[string]any{}
	for _, product := range p.Products {
		if strings.TrimSpace(product.SubjectID) == "" || subjects[product.SubjectID] || strings.TrimSpace(product.ProductName) == "" || strings.TrimSpace(product.Currency) == "" {
			return nil, fmt.Errorf("unique subject, product_name and currency required")
		}
		subjects[product.SubjectID] = true
		if len(product.Fields) != 2 {
			return nil, fmt.Errorf("explicit price and commission_rate fields required (use missing/null when absent)")
		}
		seen := map[string]bool{}
		values := map[string]any{"price": nil, "commission_rate": nil}
		kind, latest := "", ""
		var latestTime time.Time
		incomplete := false
		for _, f := range product.Fields {
			b, err := json.Marshal(f)
			if err != nil {
				return nil, err
			}
			if err := contracts.ValidateRaw("observation.schema.json", b); err != nil {
				return nil, err
			}
			if strings.TrimSpace(f.ObservationID) == "" || ids[f.ObservationID] || f.SubjectID != product.SubjectID {
				return nil, fmt.Errorf("duplicate ID or mismatched field subject")
			}
			ids[f.ObservationID] = true
			if (f.Field != "price" && f.Field != "commission_rate") || seen[f.Field] {
				return nil, fmt.Errorf("unsupported or duplicate field")
			}
			seen[f.Field] = true
			if strings.TrimSpace(f.Role) == "" || strings.TrimSpace(f.Method) == "" || strings.TrimSpace(f.Limitation) == "" || strings.TrimSpace(f.AccessMethod) == "" || (strings.TrimSpace(f.SourceURL) == "" && strings.TrimSpace(f.SourceRef) == "") {
				return nil, fmt.Errorf("field provenance required")
			}
			if f.EvidenceKind == "real" && strings.TrimSpace(f.SourceURL) == "" {
				return nil, fmt.Errorf("M00 real field requires source_url; this does not verify the source")
			}
			if kind != "" && kind != f.EvidenceKind {
				return nil, fmt.Errorf("mixed origins within product")
			}
			kind = f.EvidenceKind
			instant, err := time.Parse(time.RFC3339, f.ObservedAt)
			if err != nil {
				return nil, err
			}
			if latest == "" || instant.After(latestTime) {
				latest, latestTime = f.ObservedAt, instant
			}
			if f.Value != nil && (*f.Value < 0 || (f.Field == "commission_rate" && *f.Value > 1)) {
				return nil, fmt.Errorf("numeric value out of range")
			}
			if f.State == "observed" && f.ClaimKind != "unknown" {
				if f.Value == nil {
					return nil, fmt.Errorf("observed known field requires numeric value")
				}
				values[f.Field] = *f.Value
			} else {
				incomplete = true
			}
		}
		sort.Slice(product.Fields, func(i, j int) bool { return product.Fields[i].Field < product.Fields[j].Field })
		provenance, err := json.Marshal(struct {
			Version  string  `json:"version"`
			Question string  `json:"question"`
			Product  Product `json:"product"`
		}{p.Version, p.Question, product})
		if err != nil {
			return nil, err
		}
		state := "observed"
		if incomplete {
			state = "inconclusive"
		}
		out = append(out, map[string]any{
			"observation_id": product.ObservationID, "subject_id": product.SubjectID, "product_id": product.SubjectID, "product_name": product.ProductName, "currency": product.Currency,
			"price": values["price"], "commission_rate": values["commission_rate"], "observed_at": latest, "source_ref": "derived:m00-input/v1:" + product.ObservationID,
			"access_method": "local_packet_conversion", "evidence_kind": kind, "claim_kind": "assumption", "state": state,
			"transformation_or_method": string(provenance), "limitation": "Derived scenario input; product_name/currency are explicit caller context. Exact field provenance retained; latest timestamp does not refresh older fields. Not source verification or approval.",
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i]["product_id"].(string) < out[j]["product_id"].(string) })
	b, err := json.Marshal(out)
	if err != nil {
		return nil, err
	}
	if err := contracts.ValidateRaw("history-record.schema.json#/properties/observations", b); err != nil {
		return nil, err
	}
	return out, nil
}
