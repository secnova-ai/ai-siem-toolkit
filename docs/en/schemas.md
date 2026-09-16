# Design input and output schemas step by step

These YAML blocks are fragments inside a tool file, not standalone tools. See the [tutorial](tutorial.md) for the complete combination.

## 1. Derive types from actual input

For `{"device_id":"device-42","limit":10}`:

```yaml
input_schema:
  type: object
  additionalProperties: false
  required: [device_id, limit]
  properties:
    device_id:
      type: string
      minLength: 1
      description: Exact immutable ID returned by inventory, not hostname or IP.
    limit:
      type: integer
      minimum: 1
      maximum: 100
      description: Maximum number of records in this request, 1 through 100.
```

object declares a JSON object; properties declares fields; required lists mandatory property names. additionalProperties: false rejects misspellings/extra fields. Outer required does not make nested fields mandatory.

Use string for text, integer for whole numbers, number for decimals, boolean for true/false, object for maps, array for lists, and null when allowed. int/float/bool are not JSON Schema type names. Credential types secret/url are not JSON Schema types either.

## 2. Define enums and optional/null semantics

```yaml
input_schema:
  type: object
  properties:
    state:
      type: string
      enum: [open, closed]
      description: Optional state filter; omission selects all states.
    note:
      type: [string, "null"]
      description: Omit to leave unchanged, null to clear, or an empty string to write an empty note.
```

Optional means absent from the parent's required array; it does not imply nullable. Input-schema default does not automatically populate params. Require the value or handle omission explicitly, e.g. `has(params.state) ? params.state : "open"`. Do not read missing properties or concatenate null into URLs/headers.

## 3. Nest properties for an object

```yaml
input_schema:
  type: object
  required: [target]
  properties:
    target:
      type: object
      description: One target device for this operation.
      additionalProperties: false
      required: [device_id]
      properties:
        device_id:
          type: string
          description: Exact device ID returned by inventory.
        reason:
          type: string
          description: Optional audit reason; omission uses the device service default.
      examples:
        - device_id: device-42
          reason: Confirmed incident
```

Outer required requires target; inner required requires target.device_id. Pass `{"target":{"device_id":"device-42"}}`, not a JSON-encoded string.

## 4. Use items for an array of objects

```yaml
input_schema:
  type: object
  required: [targets]
  properties:
    targets:
      type: array
      minItems: 1
      maxItems: 20
      description: Between 1 and 20 devices selected by exact ID; no automatic pagination.
      items:
        type: object
        required: [device_id]
        additionalProperties: false
        properties:
          device_id:
            type: string
            minLength: 1
            description: Immutable device identifier from inventory.
      examples:
        - [{device_id: device-42}, {device_id: device-43}]
```

properties belongs to an object, items to an array. minLength measures a string, minItems counts list elements. Describe every nested field rather than just saying 'JSON object'. For a genuinely arbitrary dictionary, explain key/value semantics and constrain values with additionalProperties.

## 5. Describe the transformed output

If CEL returns only `{"devices":[{"id":"device-42"}]}`, describe that result rather than every raw API field. HTTP body is a string until decode_json converts it. Declaring an object schema for an unparsed string fails local output validation.

Use oneOf for mutually exclusive success/error shapes and distinguish them through required properties. Test success, empty and error results, not only one happy-path sample.

## 6. Check definition, behavior and real semantics

1. validate checks schema syntax, not correspondence with the real API.
2. test checks input values, execution results and output schema. Include missing/wrong-type/out-of-range inputs.
3. Verify the real read-only API in SIEM. Simulated tests do not prove integration.

Credential `required: true` and input-schema `required: [name]` are distinct structures; see [authentication recipes](auth-recipes.md).
