# YAML essentials for tool authors

This page explains file syntax. See the [field reference](reference.md) for semantics and the [tutorial](tutorial.md) for a complete project. Use UTF-8 .yaml files and enable whitespace display in your editor.

## 1. Indentation defines ownership

Align sibling keys and normally use two spaces per level, never tabs:

```yaml
credential_schema:
  base_url:
    type: url
    required: true
  api_key:
    type: secret
    required: true
```

Both fields belong to credential_schema; type/required belong to each field. Moving api_key to the same column as credential_schema creates an invalid top-level Provider field.

## 2. Preserve value types

```yaml
version: "1.0.0"
force_approval: false
runtime_config:
  timeout_seconds: 30
```

Version is a string, approval is boolean, timeout is integer. Do not quote booleans/numbers when their declared type requires actual boolean/number values. Quote identifiers such as "00123" to preserve leading zeros. Quote text containing colon-space, space-hash, leading braces/brackets, or ambiguous values such as "null", "true" and "on".

null, an empty string, {}, [], and an omitted field are different values and may have different API meanings.

## 3. Lists have two equivalent forms

```yaml
supported_locations:
  - saas
  - edge
```

```yaml
supported_locations: [saas, edge]
```

For a list of objects, align properties within each item:

```yaml
filters:
  - field: hostname
    value: web-01
  - field: ip
    value: "192.0.2.10"
```

This has two objects, not four elements. It is invocation data, not a top-level tool-definition fragment.

## 4. Write multiline CEL as a block string

```yaml
runtime_config:
  program: |
    get_h(
      creds.base_url + "/health",
      {"X-API-Key": creds.api_key}
    ).body.decode_json()
```

`|` preserves line breaks. `>-` folds ordinary adjacent lines and strips the final newline; more-indented lines can retain breaks. Validate the resulting expression rather than depending on folding to concatenate strings or preserve line comments. Indent the program body deeper than its key.

## 5. Three syntaxes, three interpreters

| Location | Correct form | Interpreter |
|---|---|---|
| Provider auth_strategy.headers | `Authorization: "Bearer {{creds.token}}"` | Credential template substitution |
| CEL program | `{"Authorization": "Bearer " + creds.token}` | CEL |
| Test creds | `token: example-token` | Test values |

Similarly, input_schema describes input; test params contains actual input. Do not interchange them.

## 6. Common mistakes

These are deliberately invalid examples, shown as text:

```text
credential_schema:
api_key:                 # Wrong indentation: now a top-level field
  type: secret

risk_level: low
risk_level: high         # Duplicate key

input_schema:
  type: int              # Use integer in JSON Schema
  required: true         # Here required is an array of property names
```

Use one YAML document per file, one tool per tools/*.yaml. Do not join documents with --- or use YAML anchors/aliases. The toolkit rejects aliases, duplicate keys and unknown typed fields.

## 7. Validate the actual project

Run `tcpkg validate DIRECTORY` after saving. Editor syntax highlighting cannot prove Provider fields, JSON Schema and CEL functions are correct. Fix the first reported error and validate again, then continue the [tutorial](tutorial.md).
