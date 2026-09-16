# Authoring structure contract

Read this before generating definitions. These are two complete, matching files for a CEL provider. The endpoint is illustrative. Copy both, replace the provider ID consistently, adapt the API contract and add tests. Do not copy fenced fragments from other pages as if they were complete files.

<!-- tutorial-file: _provider.yaml -->
```yaml
provider_id: sample-api
version: "1.0.0"
vendor: Example
name: sample-api
description: Read service health through a documented HTTP API.
category: utility
supported_locations: [saas]
credential_schema:
  base_url:
    type: url
    required: true
    description: Service origin, e.g. https://api.example.com, without a trailing slash.
  api_key:
    type: secret
    required: true
    description: API key with permission to read service health.
credential_test:
  type: http_probe
  method: GET
  url: "{{creds.base_url}}/health"
  headers:
    X-API-Key: "{{creds.api_key}}"
```

<!-- tutorial-file: tools/health.yaml -->
```yaml
tool_id: sample-api.health
name: Read service health
description: Read the service health without changing configuration. Returns the service status.
input_schema:
  type: object
  properties: {}
  additionalProperties: false
output_schema:
  type: object
  properties:
    status:
      type: string
      description: Service-reported health status, for example ok.
    error:
      type: string
      description: Sanitized explanation when the HTTP request failed.
    http_status:
      type: integer
      description: Non-success HTTP status reported by the service.
  oneOf:
    - required: [status]
    - required: [error, http_status]
risk_level: low
blast_radius: none
runtime_type: cel
runtime_config:
  program: >-
    [get_h(creds.base_url + "/health", {"X-API-Key": creds.api_key})]
      .map(r, r.ok ? {"status": r.body.decode_json().status} :
        {"error": "Health request failed", "http_status": r.status})[0]
```

## Structural invariants

- `_provider.yaml` is at the project root, tools are immediate `tools/*.yaml` files, referenced artifacts are `wasm/NAME.wasm`, icons are `_assets/NAME`. No enclosing directory inside the archive.
- Provider requires provider_id, version, vendor, name and category. Use a short lowercase ID and quoted semantic version. Tool requires tool_id, name, risk_level, runtime_type and the runtime's required config. tool_id begins with provider_id plus a dot.
- credential_schema is a field map. Its fields use `required: true` and credential types such as secret/url. It does not use a top-level properties map.
- input_schema/output_schema are JSON Schema. Objects use properties and `required: [names]`, arrays use items, JSON integer is spelled integer. Do not use credential types in a parameter schema.
- Nested required applies to its own object. Optional is not nullable. Parameter defaults do not auto-populate params; implement omission behavior explicitly.
- CEL config requires program. Valid variables are params/creds; header placeholders `{{creds.x}}` belong to Provider templates, not CEL. `_h` functions need explicit auth headers.
- WASM config requires a bare artifact_ref. Use src/NAME for Go source, wasm/NAME.wasm for output, and the provided HTTP SDK. Do not insert program into a WASM config.
- Static MCP config requires tool_name and optionally endpoint; otherwise credentials supply mcp_endpoint. Dynamic MCP sets mcp_discovery: true on the Provider and has no static tools/ definitions. Neither mode runs a server from the package.
- Tests use tool (not tool_id), params, creds, ordered http expectations and expect or expect_error. Test creds contains fake values, not credential schema entries.
- Use one YAML document per file, no aliases or duplicate keys. Validate the assembled project, not just individual fragments. Unknown typed fields fail. Read [reference](reference.md) for the allowed fields and [testing](testing.md) for fixtures.

A YAML parser proving syntax is not enough: run tcpkg validate, tests, pack and verify. Do not claim credentials or MCP work based on offline checks.
