package service

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestForwardedUsageModelPrefersFinalUpstreamModel(t *testing.T) {
	require.Equal(t, "target-model", forwardedUsageModel("requested-model", "target-model"))
	require.Equal(t, "requested-model", forwardedUsageModel("requested-model", "requested-model"))
	require.Equal(t, "requested-model", forwardedUsageModel("requested-model", ""))
}

func TestOpenAIForwardedUsageModelPrefersAccountMappingTarget(t *testing.T) {
	result := &OpenAIForwardResult{
		Model:         "requested-model",
		BillingModel:  "target-model",
		UpstreamModel: "target-model-2026-01-01",
	}

	require.Equal(t, "target-model", openAIForwardedUsageModel(result))
}
