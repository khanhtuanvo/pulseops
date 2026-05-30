package alerting

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
)

type AlertPayload struct {
	Source      string                 `json:"source"`
	AlertName   string                 `json:"alertName"`
	Severity    string                 `json:"severity"`
	Environment string                 `json:"environment"`
	Labels      map[string]string      `json:"labels"`
	Payload     map[string]interface{} `json:"payload"`
}

func Fingerprint(a AlertPayload) string {
	parts := []string{
		normalizeFingerprintPart(a.Source),
		normalizeFingerprintPart(a.AlertName),
		normalizeFingerprintPart(a.Severity),
		normalizeFingerprintPart(a.Environment),
	}
	sum := sha256.Sum256([]byte(strings.Join(parts, "::")))

	return hex.EncodeToString(sum[:])[:16]
}

func NormalizeAlertPayload(raw map[string]interface{}) (AlertPayload, error) {
	payload := AlertPayload{
		Labels:  map[string]string{},
		Payload: raw,
	}

	var err error
	if payload.Source, err = requiredString(raw, "source"); err != nil {
		return AlertPayload{}, err
	}
	if payload.AlertName, err = requiredString(raw, "alertName"); err != nil {
		return AlertPayload{}, err
	}
	if payload.Severity, err = requiredString(raw, "severity"); err != nil {
		return AlertPayload{}, err
	}
	if payload.Environment, err = requiredString(raw, "environment"); err != nil {
		return AlertPayload{}, err
	}

	if labels, ok := raw["labels"].(map[string]interface{}); ok {
		for key, value := range labels {
			if stringValue, ok := value.(string); ok {
				payload.Labels[key] = stringValue
			}
		}
	}

	return payload, nil
}

func requiredString(raw map[string]interface{}, field string) (string, error) {
	value, ok := raw[field]
	if !ok {
		return "", fmt.Errorf("missing required field %s", field)
	}

	stringValue, ok := value.(string)
	if !ok || strings.TrimSpace(stringValue) == "" {
		return "", fmt.Errorf("missing required field %s", field)
	}

	return stringValue, nil
}

func normalizeFingerprintPart(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}
