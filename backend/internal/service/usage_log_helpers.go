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

// forwardedUsageModel returns the model that was actually sent upstream when
// an account-level mapping changed the request. The original model remains in
// requested_model for auditability.
func forwardedUsageModel(model, upstreamModel string) string {
	model = strings.TrimSpace(model)
	upstreamModel = strings.TrimSpace(upstreamModel)
	if upstreamModel != "" && !strings.EqualFold(upstreamModel, model) {
		return upstreamModel
	}
	return model
}

// openAIForwardedUsageModel prefers the account mapping target retained by the
// OpenAI forwarder. Some accounts normalize that target before sending it
// upstream, so BillingModel is the stable model used for pricing and display.
func openAIForwardedUsageModel(result *OpenAIForwardResult) string {
	if result == nil {
		return ""
	}
	model := strings.TrimSpace(result.Model)
	billingModel := strings.TrimSpace(result.BillingModel)
	if billingModel != "" && !strings.EqualFold(billingModel, model) {
		return billingModel
	}
	return forwardedUsageModel(model, result.UpstreamModel)
}

func optionalInt64Ptr(v int64) *int64 {
	if v == 0 {
		return nil
	}
	return &v
}
