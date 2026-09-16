# Definition reference

YAML must be a single document. Aliases, duplicate keys and unknown typed fields are rejected. This prevents misspelled options from silently taking no effect. Tool input/output schemas use self-contained JSON Schema 2020-12; external references are disabled.

## Provider: `_provider.yaml`

| Field | Meaning |
|---|---|
| provider_id | Required lowercase stable ID, up to 64 characters |
| version | Required semantic version, e.g. 1.0.0 |
| vendor, name, category | Required display/vendor/category values |
| description | Service purpose |
| supported_locations | saas and/or edge; actual deployment must provide the selected execution location |
| credential_schema | Map of named credential field definitions; see credentials guide |
| auth_strategy | Platform auth configuration; MCP header/query templates or supported OAuth settings |
| credential_test | Probe configuration: type http_probe/oauth2_client_credentials/skip |
| api | Platform API metadata; does not create extra CEL variables |
| rate_limit | Metadata, not a local runtime rate limiter |
| mcp_discovery | Dynamic discovery; omit static tools when true |
| mcp_endpoint | Optional fixed dynamic-discovery endpoint |
| icon, icon_dark | Bare filenames in _assets/, no path separators |

Credential field definitions support type, required, default, description and UI metadata such as enum/options. The CLI checks recognized field types/default value types; platform UI metadata is not exhaustively linted. Secret defaults should not contain real credentials.

`http_probe` commonly uses method, url, headers and expected_status; templates resolve credential values. `skip` performs no connectivity check. Auth strategy header/query maps use `{{creds.field}}`. `oauth2_client_credentials` uses token_url and credential client_id/client_secret. Actual authorization remains a platform workflow.

## Tool: `tools/NAME.yaml`

| Field | Meaning |
|---|---|
| tool_id | Required provider_id.action; unique in the package |
| provider_id | Optional; must match parent when supplied |
| name | Required display name |
| description, human_description | AI-facing complete usage and optional shorter display text |
| input_schema, output_schema | Input/output structure, descriptions, constraints and examples |
| risk_level | Required low/medium/high/critical |
| runtime_type | Required cel/wasm/mcp |
| runtime_config | Runtime-specific fields below |
| required_permissions | Existing platform permission tags; do not invent new permissions |
| force_approval | Always request platform approval |
| sensitive_params | Parameter names to redact according to platform behavior |
| irreversible | Whether the operation is irreversible |
| blast_radius | none/single/multi/global |
| allowed_caller_types | chat_agent/async_agent/automation/external_api |

CEL config: program (required), allowed_domains (hostnames), timeout_seconds. WASM config: artifact_ref (required bare .wasm filename), optional artifact_checksum and allowed_domains. MCP config: tool_name (required), endpoint (optional). Platform-generated OpenAPI plans, report sources and legacy runtime options are not part of this starter contract.
