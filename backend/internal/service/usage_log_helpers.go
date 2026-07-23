package service

import "strings"

func optionalTrimmedStringPtr(raw string) *string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

// optionalNonEqualStringPtr returns a pointer to value if it is non-empty and
// differs from compare; otherwise nil. Used to store upstream_model only when
// it differs from the requested model.
func optionalNonEqualStringPtr(value, compare string) *string {
	if value == "" || value == compare {
		return nil
	}
	return &value
}

func forwardResultBillingModel(requestedModel, upstreamModel string) string {
	if trimmed := strings.TrimSpace(requestedModel); trimmed != "" {
		return trimmed
	}
	return strings.TrimSpace(upstreamModel)
}

// isAccountMappedOpenAIBillingResult reports whether BillingModel was produced
// by the selected account's model mapping. The requested model remains the
// customer-facing billable model; BillingModel and UpstreamModel are retained
// for upstream audit data.
func isAccountMappedOpenAIBillingResult(account *Account, result *OpenAIForwardResult) bool {
	if account == nil || result == nil {
		return false
	}

	requestedModel := strings.TrimSpace(result.Model)
	billingModel := strings.TrimSpace(result.BillingModel)
	if requestedModel == "" || billingModel == "" || strings.EqualFold(requestedModel, billingModel) {
		return false
	}

	mappedModel, matched := account.ResolveMappedModel(requestedModel)
	return matched && strings.EqualFold(strings.TrimSpace(mappedModel), billingModel)
}

func optionalInt64Ptr(v int64) *int64 {
	if v == 0 {
		return nil
	}
	return &v
}
