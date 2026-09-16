package credential

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestValidateSchemaDefinitionAcceptsTypedDefaults(t *testing.T) {
	schema := json.RawMessage(`{
		"base_url":{"type":"url","required":false,"default":"https://example.test"},
		"verify_ssl":{"type":"boolean","required":false,"default":false},
		"port":{"type":"integer","required":false,"default":443}
	}`)
	if err := ValidateSchemaDefinition(schema); err != nil {
		t.Fatalf("ValidateSchemaDefinition() error = %v", err)
	}
}

func TestValidateSchemaRejectsInvalidURLBeforeConnectivity(t *testing.T) {
	schema := json.RawMessage(`{"base_url":{"type":"url","required":true}}`)
	secret := json.RawMessage(`{"base_url":"not-a-url"}`)
	err := ValidateSchema(schema, secret)
	if err == nil || !strings.Contains(err.Error(), "absolute HTTP or HTTPS URL") {
		t.Fatalf("ValidateSchema() error = %v", err)
	}
}

func TestValidateSchemaChecksBooleanAndNumberTypes(t *testing.T) {
	schema := json.RawMessage(`{
		"verify_ssl":{"type":"boolean","required":true},
		"retries":{"type":"integer","required":true}
	}`)
	if err := ValidateSchema(schema, json.RawMessage(`{"verify_ssl":false,"retries":3}`)); err != nil {
		t.Fatalf("valid typed credential rejected: %v", err)
	}
	if err := ValidateSchema(schema, json.RawMessage(`{"verify_ssl":"false","retries":3}`)); err == nil {
		t.Fatal("string boolean should be rejected")
	}
}

func TestApplySchemaDefaultsPreservesFalseZeroAndEmptyString(t *testing.T) {
	schema := json.RawMessage(`{
		"verify_ssl":{"type":"boolean","default":false},
		"retries":{"type":"integer","default":0},
		"scope":{"type":"string","default":""}
	}`)
	got := ApplySchemaDefaults(schema, map[string]any{})
	if got["verify_ssl"] != false || got["retries"] != float64(0) || got["scope"] != "" {
		t.Fatalf("ApplySchemaDefaults() = %#v", got)
	}
}
