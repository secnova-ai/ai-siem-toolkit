package schema

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidatorRejectsWrongNestedEvidenceBindingShape(t *testing.T) {
	schema := json.RawMessage(`{
		"type":"object",
		"additionalProperties":false,
		"required":["edges"],
		"properties":{"edges":{"type":"array","items":{"type":"object","required":["evidence_bindings"],"properties":{"evidence_bindings":{"type":"array","minItems":1,"items":{"type":"object","required":["object"],"properties":{"object":{"type":"object","required":["field_path","value"],"properties":{"field_path":{"type":"string"},"value":{"type":"string"}}}}}}}}}}
	}`)

	err := new(Validator).Validate("detection.apply_incident_ai_graph_patch", schema, json.RawMessage(`{
		"edges":[{"evidence_bindings":[{"object":"198.18.95.95:8443"}]}]
	}`))

	var validationErr *ValidationError
	require.ErrorAs(t, err, &validationErr)
	require.Contains(t, err.Error(), "/edges/0/evidence_bindings/0/object")
	require.Contains(t, err.Error(), "want object")
}

func TestValidatorAcceptsNestedEvidenceBindingObject(t *testing.T) {
	schema := json.RawMessage(`{
		"type":"object",
		"required":["object"],
		"properties":{"object":{"type":"object","required":["field_path","value"],"properties":{"field_path":{"type":"string"},"value":{"type":"string"}}}}
	}`)

	err := new(Validator).Validate("test.tool", schema, json.RawMessage(`{
		"object":{"field_path":"dst_endpoint.ip","value":"198.18.95.95"}
	}`))
	require.NoError(t, err)
}

func TestValidatorRejectsUnknownPropertiesAndBadFormats(t *testing.T) {
	schema := json.RawMessage(`{
		"type":"object",
		"additionalProperties":false,
		"required":["round_uid"],
		"properties":{"round_uid":{"type":"string","format":"uuid"}}
	}`)
	validator := new(Validator)

	for name, input := range map[string]string{
		"unknown property": `{"round_uid":"0d29e73a-98ca-4cb3-aa10-f6ae15e54bc2","revision":1}`,
		"bad UUID":         `{"round_uid":"current-round"}`,
	} {
		t.Run(name, func(t *testing.T) {
			var validationErr *ValidationError
			err := validator.Validate("test.tool", schema, json.RawMessage(input))
			require.True(t, errors.As(err, &validationErr), "error: %v", err)
		})
	}
}

func TestValidateDefinitionRejectsBrokenReference(t *testing.T) {
	err := ValidateDefinition(json.RawMessage(`{"$ref":"#/$defs/missing"}`))
	require.Error(t, err)
}

func TestValidateDefinitionRejectsExternalReference(t *testing.T) {
	err := ValidateDefinition(json.RawMessage(`{"$ref":"https://example.invalid/external-schema.json"}`))
	require.Error(t, err)
}

func TestValidatorAllowsMissingSchema(t *testing.T) {
	require.NoError(t, new(Validator).Validate("test.tool", json.RawMessage(`null`), json.RawMessage(`{"anything":true}`)))
}
