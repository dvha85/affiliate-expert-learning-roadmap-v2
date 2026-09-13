package m06

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/dvha85/affiliate-expert-learning-roadmap-v2/contracts"
)

// AccesstradeShopeeCampaignCapture is a deliberately small, sanitized
// transcription of the selected authenticated campaign page. It is not HTML,
// a credential export, a report export, or an affiliate-link request. The
// capture deliberately excludes commissions, EPC/CVR, publisher counts,
// reviews, promotional copy and any personal data.
type AccesstradeShopeeCampaignCapture struct {
	Version             string `json:"version"`
	Method              string `json:"method"`
	ObservedAt          string `json:"observed_at"`
	CorrelationID       string `json:"correlation_id"`
	StatusCode          int    `json:"status_code"`
	Redirected          bool   `json:"redirected"`
	SourcePageSHA256    string `json:"source_page_sha256"`
	CampaignTitle       string `json:"campaign_title"`
	MerchantLabel       string `json:"merchant_label"`
	CampaignCategory    string `json:"campaign_category"`
	CampaignStatusLabel string `json:"campaign_status_label"`
	CampaignPeriodLabel string `json:"campaign_period_label"`
}

// AccesstradeShopeeCampaignProfile is code/config reviewed before a capture
// can enter canonical history. In particular, the source URL is never taken
// from n8n input, a browser response, or a model output.
type AccesstradeShopeeCampaignProfile struct {
	SourceURL string
}

const AccesstradeShopeeSmartlinkURL = "https://pub2.accesstrade.vn/campaign-v2/Shopee%20Vi%E1%BB%87t%20Nam%20Smartlink%20cho%20t%E1%BA%A5t%20c%E1%BA%A3%20thi%E1%BA%BFt%20b%E1%BB%8B--128"

const accesstradeShopeeTitle = "Shopee Việt Nam Smartlink cho tất cả thiết bị"

// BuildAccesstradeShopeeCampaign builds the same M00 input packet used by the
// watcher fixture path. It represents campaign metadata only. Price and
// commission are explicitly unknown, so the downstream learner cannot rank,
// recommend, promise revenue, generate a link, or treat a campaign claim as a
// business outcome.
func BuildAccesstradeShopeeCampaign(raw []byte, profile AccesstradeShopeeCampaignProfile) (OfferFixtureBuild, error) {
	fail := func(message string) (OfferFixtureBuild, error) {
		return OfferFixtureBuild{}, fmt.Errorf("M06 ACCESSTRADE Shopee campaign: %s", message)
	}
	parsedURL, err := url.Parse(profile.SourceURL)
	if err != nil || profile.SourceURL != AccesstradeShopeeSmartlinkURL || parsedURL.Scheme != "https" || parsedURL.Hostname() != "pub2.accesstrade.vn" || parsedURL.RawQuery != "" || parsedURL.Fragment != "" {
		return fail("unsupported fixed selected-source profile")
	}
	var capture AccesstradeShopeeCampaignCapture
	if err := contracts.DecodeStrict(raw, &capture); err != nil {
		return fail("invalid sanitized capture")
	}
	if capture.Version != "accesstrade-shopee-campaign-capture/v1" || capture.Method != "GET" || capture.StatusCode != 200 || capture.Redirected {
		return fail("unsupported capture method/status/redirect")
	}
	observedAt, err := time.Parse(time.RFC3339, capture.ObservedAt)
	if err != nil {
		return fail("invalid observed_at")
	}
	capture.ObservedAt = observedAt.UTC().Format(time.RFC3339Nano)
	if !stableToken(capture.CorrelationID) {
		return fail("stable correlation_id required")
	}
	if len(capture.SourcePageSHA256) != 64 || strings.ToLower(capture.SourcePageSHA256) != capture.SourcePageSHA256 {
		return fail("source_page_sha256 must be lowercase SHA-256")
	}
	for _, r := range capture.SourcePageSHA256 {
		if !((r >= '0' && r <= '9') || (r >= 'a' && r <= 'f')) {
			return fail("source_page_sha256 must be lowercase SHA-256")
		}
	}
	if capture.CampaignTitle != accesstradeShopeeTitle || capture.MerchantLabel != "Shopee" || capture.CampaignCategory != "Thương Mại Điện Tử" || !boundedLabel(capture.CampaignStatusLabel) || !boundedLabel(capture.CampaignPeriodLabel) {
		return fail("capture does not match the reviewed selected-source contract")
	}
	semantic, err := json.Marshal(struct {
		Title, Merchant, Category, Status, Period, SourcePageSHA256 string
	}{capture.CampaignTitle, capture.MerchantLabel, capture.CampaignCategory, capture.CampaignStatusLabel, capture.CampaignPeriodLabel, capture.SourcePageSHA256})
	if err != nil {
		return fail("semantic capture encoding")
	}
	request := WatchRequest{Method: "GET", URL: profile.SourceURL, AllowHosts: []string{"pub2.accesstrade.vn"}, ObservedAt: capture.ObservedAt, CorrelationID: capture.CorrelationID, Body: string(semantic)}
	normalized, state := NormalizeWatchObservation(request, "accesstrade-shopee-smartlink-128")
	if state != "NEW" {
		return fail("normalization rejected")
	}
	metadata := "campaign_metadata=" + string(semantic) + "; source_page_sha256=" + capture.SourcePageSHA256
	limitation := "Observed authenticated ACCESSTRADE campaign metadata, not independent business truth. Campaign/operator claims are volatile seller claims; commission, EPC/CVR, eligibility, approval and earnings are unknown here. No affiliate link, publisher action, execution authority, PII or business outcome is present. " + metadata
	fields := make([]map[string]any, 0, 2)
	for _, field := range []string{"price", "commission_rate"} {
		fields = append(fields, map[string]any{
			"observation_id":           normalized.ObservationID + "-" + field,
			"subject_id":               "accesstrade-shopee-smartlink-128",
			"source_url":               profile.SourceURL,
			"observed_at":              normalized.ObservedAt,
			"access_method":            "GET",
			"evidence_kind":            "real",
			"field_or_claim":           field,
			"claim_kind":               "unknown",
			"value":                    nil,
			"state":                    "missing",
			"source_authority_or_role": "accesstrade_campaign_page",
			"transformation_or_method": "accesstrade-shopee-campaign-capture/v1; " + metadata,
			"limitation":               limitation,
		})
	}
	packet, err := json.Marshal(map[string]any{
		"version":  "m00-input/v1",
		"question": "Read-only selected campaign metadata; no ranking, recommendation, affiliate link, publisher action, execution, or business-outcome claim.",
		"products": []any{map[string]any{
			"observation_id": normalized.ObservationID,
			"subject_id":     "accesstrade-shopee-smartlink-128",
			"product_name":   capture.CampaignTitle,
			"currency":       "N/A",
			"fields":         fields,
		}},
	})
	if err != nil {
		return fail("projection encoding")
	}
	identity, err := json.Marshal([]string{capture.Version, profile.SourceURL, capture.CorrelationID, normalized.ContentHash})
	if err != nil {
		return fail("record identity encoding")
	}
	return OfferFixtureBuild{RecordID: "watch-" + ContentHash(string(identity)), ObservedAt: normalized.ObservedAt, Packet: packet}, nil
}

func stableToken(value string) bool {
	return strings.TrimSpace(value) != "" && strings.TrimSpace(value) == value && len(value) <= 128
}

func boundedLabel(value string) bool {
	return strings.TrimSpace(value) != "" && strings.TrimSpace(value) == value && len(value) <= 256
}
