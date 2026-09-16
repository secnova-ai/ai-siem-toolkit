package schema

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	jsonschema "github.com/santhosh-tekuri/jsonschema/v6"
)

const inMemorySchemaURL = "https://schemas.secnova.example/tool-input.json"

// ValidationError reports caller-supplied tool parameters that do not match
// the Tool Catalog input_schema. Callers may correct the parameters and retry.
type ValidationError struct {
	ToolID string
	Cause  error
}

func (e *ValidationError) Error() string {
	if e == nil {
		return "invalid tool parameters"
	}
	if e.ToolID == "" {
		return "invalid tool parameters: " + e.Cause.Error()
	}
	return fmt.Sprintf("invalid parameters for tool %s: %v", e.ToolID, e.Cause)
}

func (e *ValidationError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Cause
}

// SchemaError reports a malformed input_schema stored in the Tool Catalog.
// This is a server configuration problem, not a caller validation failure.
type SchemaError struct {
	ToolID string
	Cause  error
}

func (e *SchemaError) Error() string {
	if e == nil {
		return "invalid tool input schema"
	}
	if e.ToolID == "" {
		return "invalid tool input schema: " + e.Cause.Error()
	}
	return fmt.Sprintf("invalid input schema for tool %s: %v", e.ToolID, e.Cause)
}

func (e *SchemaError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Cause
}

type cacheKey struct {
	toolID string
	digest [sha256.Size]byte
}

type cacheEntry struct {
	schema *jsonschema.Schema
	err    error
}

// Validator compiles Tool Catalog JSON Schemas once per tool/schema digest and
// validates every invocation before policy evaluation or runtime dispatch.
// Its zero value is ready for use.
type Validator struct {
	cache sync.Map
}

// ValidateDefinition compiles a schema without validating an instance. Catalog
// loading uses it to reject malformed provider packages before they are stored.
func ValidateDefinition(raw json.RawMessage) error {
	if schemaDisabled(raw) {
		return nil
	}
	_, err := compile(raw)
	return err
}

// Validate checks a JSON invocation payload against the tool input schema.
func (v *Validator) Validate(toolID string, rawSchema, rawInput json.RawMessage) error {
	instance, err := decodeJSON(rawInput)
	if err != nil {
		return &ValidationError{ToolID: toolID, Cause: fmt.Errorf("invalid JSON: %w", err)}
	}
	if schemaDisabled(rawSchema) {
		return nil
	}

	key := cacheKey{toolID: toolID, digest: sha256.Sum256(rawSchema)}
	entryValue, ok := v.cache.Load(key)
	if !ok {
		schema, compileErr := compile(rawSchema)
		entryValue, _ = v.cache.LoadOrStore(key, cacheEntry{schema: schema, err: compileErr})
	}
	entry := entryValue.(cacheEntry)
	if entry.err != nil {
		return &SchemaError{ToolID: toolID, Cause: entry.err}
	}
	if err := entry.schema.Validate(instance); err != nil {
		return &ValidationError{ToolID: toolID, Cause: err}
	}
	return nil
}

func schemaDisabled(raw json.RawMessage) bool {
	trimmed := strings.TrimSpace(string(raw))
	return trimmed == "" || trimmed == "null"
}

func compile(raw json.RawMessage) (*jsonschema.Schema, error) {
	document, err := decodeJSON(raw)
	if err != nil {
		return nil, fmt.Errorf("decode schema JSON: %w", err)
	}
	compiler := jsonschema.NewCompiler()
	compiler.DefaultDraft(jsonschema.Draft2020)
	compiler.AssertFormat()
	// Tool schemas are self-contained catalog declarations. External $ref
	// loading would make catalog import and invocation depend on arbitrary
	// network or filesystem resources supplied by a provider package.
	compiler.UseLoader(jsonschema.SchemeURLLoader{})
	if err := compiler.AddResource(inMemorySchemaURL, document); err != nil {
		return nil, fmt.Errorf("add schema resource: %w", err)
	}
	schema, err := compiler.Compile(inMemorySchemaURL)
	if err != nil {
		return nil, fmt.Errorf("compile schema: %w", err)
	}
	return schema, nil
}

func decodeJSON(raw json.RawMessage) (any, error) {
	if len(raw) == 0 {
		raw = json.RawMessage(`{}`)
	}
	return jsonschema.UnmarshalJSON(bytes.NewReader(raw))
}
