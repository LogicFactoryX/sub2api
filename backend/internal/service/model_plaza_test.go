package service

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestModelPlazaPricingFromModelAppliesPerMillionFactor(t *testing.T) {
	pricing := &ModelPricing{
		InputPricePerToken:         5e-6,
		OutputPricePerToken:        30e-6,
		CacheCreationPricePerToken: 6.25e-6,
		CacheReadPricePerToken:     0.5e-6,
	}

	actual := modelPlazaPricingFromModel(pricing, 0.1/5)

	require.NotNil(t, actual.InputPerMillion)
	require.NotNil(t, actual.OutputPerMillion)
	require.NotNil(t, actual.CacheWritePerMillion)
	require.NotNil(t, actual.CacheReadPerMillion)
	require.InDelta(t, 0.1, *actual.InputPerMillion, 1e-12)
	require.InDelta(t, 0.6, *actual.OutputPerMillion, 1e-12)
	require.InDelta(t, 0.125, *actual.CacheWritePerMillion, 1e-12)
	require.InDelta(t, 0.01, *actual.CacheReadPerMillion, 1e-12)
}

func TestNormalizeModelPlazaTagsTrimsDeduplicatesAndLimits(t *testing.T) {
	tags := normalizeModelPlazaTags([]string{
		" code ", "CODE", "reasoning", "one", "two", "three", "four", "five", "six", "seven",
	})

	require.Equal(t, []string{"code", "reasoning", "one", "two", "three", "four", "five", "six"}, tags)
}
