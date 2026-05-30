package alerting

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFingerprintIsStable(t *testing.T) {
	payload := AlertPayload{Source: "prometheus", AlertName: "HighCPU", Severity: "critical", Environment: "prod"}

	require.Equal(t, Fingerprint(payload), Fingerprint(payload))
}

func TestFingerprintChangesForDifferentSource(t *testing.T) {
	first := AlertPayload{Source: "prometheus", AlertName: "HighCPU", Severity: "critical", Environment: "prod"}
	second := AlertPayload{Source: "datadog", AlertName: "HighCPU", Severity: "critical", Environment: "prod"}

	require.NotEqual(t, Fingerprint(first), Fingerprint(second))
}

func TestFingerprintIsCaseInsensitive(t *testing.T) {
	first := AlertPayload{Source: "PROMETHEUS", AlertName: "HighCPU", Severity: "CRITICAL", Environment: "PROD"}
	second := AlertPayload{Source: "prometheus", AlertName: "highcpu", Severity: "critical", Environment: "prod"}

	require.Equal(t, Fingerprint(first), Fingerprint(second))
}

func TestNormalizeAlertPayloadRequiresFields(t *testing.T) {
	_, err := NormalizeAlertPayload(map[string]interface{}{
		"source":      "prometheus",
		"severity":    "critical",
		"environment": "prod",
	})

	require.ErrorContains(t, err, "alertName")
}
