package credential

import (
	"encoding/json"
	"fmt"
	"math"
	"net/url"
	"strings"
)

type schemaField struct {
	Type     string `json:"type"`
	Required bool   `json:"required"`
	Default  any    `json:"default"`
}

var supportedSchemaFieldTypes = map[string]bool{
	"":        true,
	"string":  true,
	"secret":  true,
	"text":    true,
	"url":     true,
	"boolean": true,
	"number":  true,
	"integer": true,
	"select":  true,
}

// ValidateSchemaDefinition validates the provider-side credential schema
// itself. Provider packages should fail during loading instead of deferring an
// invalid field type/default to tenant credential setup.
func ValidateSchemaDefinition(credentialSchema json.RawMessage) error {
	schema, err := decodeSchema(credentialSchema)
	if err != nil {
		return err
	}
	for field, def := range schema {
		if !supportedSchemaFieldTypes[def.Type] {
			return fmt.Errorf("field %q has unsupported type %q", field, def.Type)
		}
		if def.Default != nil {
			if err := validateSchemaValue(field, def.Type, def.Default); err != nil {
				return fmt.Errorf("invalid default: %w", err)
			}
		}
	}
	return nil
}

// ValidateSchema checks that the provided secretJSON satisfies the required
// fields declared in credentialSchema.
// A nil, "null", or "{}" schema is treated as no validation needed.
func ValidateSchema(credentialSchema, secretJSON json.RawMessage) error {
	s := string(credentialSchema)
	if len(credentialSchema) == 0 || s == "null" || s == "{}" {
		return nil
	}

	schema, err := decodeSchema(credentialSchema)
	if err != nil {
		return err
	}

	var secretMap map[string]interface{}
	if err := json.Unmarshal(secretJSON, &secretMap); err != nil {
		return fmt.Errorf("secret is not a valid JSON object: %w", err)
	}

	for field, def := range schema {
		v, exists := secretMap[field]
		if !exists || v == nil {
			if !def.Required {
				continue
			}
			return fmt.Errorf("required field %q is missing or null", field)
		}
		if err := validateSchemaValue(field, def.Type, v); err != nil {
			return err
		}
		if def.Required {
			if s, ok := v.(string); ok && strings.TrimSpace(s) == "" {
				return fmt.Errorf("required field %q must not be empty", field)
			}
		}
	}
	return nil
}

func decodeSchema(credentialSchema json.RawMessage) (map[string]schemaField, error) {
	s := strings.TrimSpace(string(credentialSchema))
	if len(credentialSchema) == 0 || s == "null" || s == "{}" {
		return map[string]schemaField{}, nil
	}
	var schema map[string]schemaField
	if err := json.Unmarshal(credentialSchema, &schema); err != nil {
		return nil, fmt.Errorf("invalid credential_schema: %w", err)
	}
	return schema, nil
}

func validateSchemaValue(field, fieldType string, value any) error {
	switch fieldType {
	case "", "string", "secret", "text", "select":
		if _, ok := value.(string); !ok {
			return fmt.Errorf("field %q must be a string", field)
		}
	case "url":
		raw, ok := value.(string)
		if !ok {
			return fmt.Errorf("field %q must be a string", field)
		}
		parsed, err := url.ParseRequestURI(strings.TrimSpace(raw))
		if err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
			return fmt.Errorf("field %q must be an absolute HTTP or HTTPS URL", field)
		}
	case "boolean":
		if _, ok := value.(bool); !ok {
			return fmt.Errorf("field %q must be a boolean", field)
		}
	case "number":
		if _, ok := value.(float64); !ok {
			return fmt.Errorf("field %q must be a number", field)
		}
	case "integer":
		number, ok := value.(float64)
		if !ok || math.Trunc(number) != number {
			return fmt.Errorf("field %q must be an integer", field)
		}
	default:
		return fmt.Errorf("field %q has unsupported type %q", field, fieldType)
	}
	return nil
}

// ApplySchemaDefaults returns a copy of the secret map with any missing
// optional fields filled in from their default values in credentialSchema.
// This ensures validation and execution see the same credential context.
func ApplySchemaDefaults(credentialSchema json.RawMessage, creds map[string]any) map[string]any {
	result := make(map[string]any, len(creds))
	for k, v := range creds {
		result[k] = v
	}

	s := string(credentialSchema)
	if len(credentialSchema) == 0 || s == "null" || s == "{}" {
		return result
	}

	schema, err := decodeSchema(credentialSchema)
	if err != nil {
		return result
	}

	for field, def := range schema {
		if def.Default != nil {
			if _, exists := result[field]; !exists {
				result[field] = def.Default
			}
		}
	}
	return result
}
