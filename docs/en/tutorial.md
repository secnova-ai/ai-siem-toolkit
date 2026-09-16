# Build an installable tool from an API request

You will build `device-api.search`: search inventory with structured filters, return device IDs, test locally and package for SIEM. All endpoints and keys are illustrative. Do not send requests to the example.com endpoint; mock tests require no device or real credential.

Install the CLI and check `tcpkg version`. Read [YAML basics](yaml-basics.md) if indentation or block strings are unfamiliar. Run commands from the parent directory of the new project.

## Step 1: Understand the API

Assume the vendor documents this request (for reading, not execution):

```sh
curl -X POST 'https://inventory.example.com/devices/search' -H 'X-API-Key: example-only' -H 'Content-Type: application/json' -d '{"filters":[{"field":"hostname","value":"web-01"}],"limit":10}'
```

A success returns `{"devices":[{"id":"device-42"}]}`, an empty result returns `{"devices":[]}`, and invalid authentication returns HTTP 401. POST is read-only in this example.

| API information | Definition location | Reason |
|---|---|---|
| Service origin | Credential base_url | Varies by environment |
| X-API-Key value | Credential api_key | Secret, not an invocation parameter |
| filters and limit | input_schema | Selected for each invocation |
| POST and /devices/search | CEL program | Fixed operation contract |
| devices/id result | output_schema | Explains the result to callers |
| Read-only effect | risk_level: low | Risk follows behavior, not verb |

Choose provider ID `device-api` and action `search`. The complete ID is `device-api.search`. Start with one meaningful action, not every endpoint in the vendor catalog.

## Step 2: Initialize and replace the placeholder example

```sh
tcpkg init device-api --runtime cel
```

Using your editor, remove the generated `tools/health.yaml` and original health test files under `tests/` in this newly created project. Keep .gitignore, README and LICENSE. If the older v0.1.0 CLI does not generate LICENSE, copy the release LICENSE attachment into the project root. Do not remove tools from an existing project. The completed layout will be:

```text
device-api/
  _provider.yaml
  tools/search.yaml
  tests/search.yaml
  tests/empty.yaml
  tests/unauthorized.yaml
  .gitignore
  README.md
  LICENSE
```

The tool_id inside YAML is the identifier; the filename is organizational.

## Step 3: Replace `_provider.yaml` completely

<!-- tutorial-file: _provider.yaml -->
```yaml
provider_id: device-api
version: "1.0.0"
vendor: Example
name: Device inventory example
description: Illustrative inventory API; adapt endpoints to your vendor documentation.
category: utility
supported_locations: [saas]
credential_schema:
  base_url:
    type: url
    required: true
    description: API origin without a trailing slash.
  api_key:
    type: secret
    required: true
    description: API key with read-only inventory access.
credential_test:
  type: skip
```

The block is a complete file. provider_id is stable; name is a display label. version is the provider package version, not CLI/device version. vendor/category should match the actual service. supported_locations selects where execution happens; browser reachability does not prove the platform can reach the API.

credential_schema declares fields, not their values. base_url's url type validates the address and contributes its hostname to the runtime allowlist. api_key is secret. required: true is a credential-field flag. credential_test: skip means this example has no separate probe; it does not prove authentication works. Replace it with a harmless documented probe using [authentication recipes](auth-recipes.md) when available.

## Step 4: Create the complete `tools/search.yaml`

<!-- tutorial-file: tools/search.yaml -->
```yaml
tool_id: device-api.search
name: Search devices
description: Search device inventory using exact-match filters. Read-only despite using POST. Returns up to limit devices; an empty devices array means no match. Does not isolate or modify devices.
risk_level: low
blast_radius: none
runtime_type: cel
input_schema:
  type: object
  additionalProperties: false
  required: [filters, limit]
  properties:
    filters:
      type: array
      minItems: 1
      maxItems: 10
      description: Exact-match filters combined with AND. Each entry supplies a supported field and a nonempty value.
      items:
        type: object
        additionalProperties: false
        required: [field, value]
        properties:
          field:
            type: string
            enum: [hostname, ip]
            description: Inventory field to match; hostname is the exact registered name, ip is the device address.
          value:
            type: string
            minLength: 1
            description: Exact value for the selected field, without wildcard characters; for example web-01.
      examples:
        - [{field: hostname, value: web-01}]
    limit:
      type: integer
      minimum: 1
      maximum: 100
      description: Maximum returned device count, 1 through 100. Required explicitly; this tool does not fetch additional pages.
output_schema:
  type: object
  properties:
    devices:
      type: array
      description: Matched device records; empty when nothing matches.
      items:
        type: object
        required: [id]
        properties:
          id:
            type: string
            description: Immutable device identifier.
    error:
      type: string
      description: Sanitized API error explanation.
    http_status:
      type: integer
      description: HTTP error status.
  oneOf:
    - required: [devices]
    - required: [error, http_status]
runtime_config:
  program: >-
    [post_h(creds.base_url + "/devices/search",
      {"filters": params.filters, "limit": params.limit}.encode_json(),
      {"X-API-Key": creds.api_key})]
      .map(r, r.ok ? {"devices": r.body.decode_json().devices} :
        {"error": "Inventory request failed", "http_status": r.status})[0]
```

Read the file in four parts: identity/risk; input schema; output schema; execution program. filters is an array of objects; items.properties defines each object's field/value. limit is a bounded integer. See [parameter schemas](schemas.md) for how nested required, enum and items work.

oneOf describes either successful devices or an explicit error object. Returning an error object is not automatically a platform execution failure: downstream consumers must check error/http_status.

| Expression | Meaning |
|---|---|
| `creds.base_url + "/devices/search"` | Selected credential origin plus fixed path |
| `params.filters`, `params.limit` | This invocation's input values |
| `{...}.encode_json()` | Serialize the request body rather than manually interpolate JSON |
| `{"X-API-Key": creds.api_key}` | Explicit authentication header |
| `[post_h(...)]` | Perform the request once and store its response in a one-element list |
| `.map(r, ...)[0]` | Transform that response and return the single result, without another request |
| `r.ok ? ... : ...` | Decode devices on 2xx, otherwise return a recognizable error |

post_h defaults to JSON Content-Type. _h helpers do not auto-inject Bearer. Provider template syntax `{{creds.api_key}}` is not CEL syntax; use `creds.api_key` in an expression.

## Step 5: Validate before testing

```sh
tcpkg validate device-api
```

Expect `valid device-api: 1 static tools, discovery=false`. Fix unknown fields, schema errors and CEL compilation failures before packaging. Validation performs no HTTP requests and cannot authenticate an API key.

## Step 6: Test a successful request

Create `tests/search.yaml`:

<!-- tutorial-file: tests/search.yaml -->
```yaml
name: Nested filters produce the correct request
tool: device-api.search
params:
  filters:
    - field: hostname
      value: web-01
  limit: 10
creds:
  base_url: https://inventory.example.com
  api_key: example-only
http:
  - request:
      method: POST
      url: https://inventory.example.com/devices/search
      headers:
        X-API-Key: example-only
        Content-Type: application/json
      body: '{"filters":[{"field":"hostname","value":"web-01"}],"limit":10}'
    response:
      status: 200
      body: '{"devices":[{"id":"device-42"}]}'
expect:
  devices:
    - id: device-42
```

`tool` selects an ID; params contains input values, not a schema; creds contains fake credential values, not credential_schema. http lists ordered request expectations and simulated responses. expect is the full output value. URLs and request bodies compare exactly; JSON body strings must use the serializer's key ordering. Headers compare only fields listed in the expectation.

```sh
tcpkg test device-api
```

Expect `PASS search.yaml` and `1 tests passed`. The actual CEL expression runs, but HTTP transport is simulated; no request is sent to inventory.example.com.

## Step 7: Cover empty results and authentication rejection

Create `tests/empty.yaml`:

<!-- tutorial-file: tests/empty.yaml -->
```yaml
name: No matching device is a valid result
tool: device-api.search
params:
  filters:
    - field: hostname
      value: web-01
  limit: 10
creds:
  base_url: https://inventory.example.com
  api_key: example-only
http:
  - request:
      method: POST
      url: https://inventory.example.com/devices/search
      headers:
        X-API-Key: example-only
        Content-Type: application/json
      body: '{"filters":[{"field":"hostname","value":"web-01"}],"limit":10}'
    response:
      status: 200
      body: '{"devices":[]}'
expect:
  devices: []
```

Create `tests/unauthorized.yaml`:

<!-- tutorial-file: tests/unauthorized.yaml -->
```yaml
name: Authentication failure is explicit
tool: device-api.search
params:
  filters:
    - field: hostname
      value: web-01
  limit: 10
creds:
  base_url: https://inventory.example.com
  api_key: example-only
http:
  - request:
      method: POST
      url: https://inventory.example.com/devices/search
      headers:
        X-API-Key: example-only
        Content-Type: application/json
      body: '{"filters":[{"field":"hostname","value":"web-01"}],"limit":10}'
    response:
      status: 401
      body: '{"devices":[{"id":"device-42"}]}'
expect:
  error: Inventory request failed
  http_status: 401
```

Run `tcpkg test device-api` again: all three tests should pass. The 401 test uses expect because CEL returns an error object; expect_error applies to actual validation/execution failures.

As an exercise, change the successful test's limit to 0. Input validation should fail before HTTP runs. Restore 10 and confirm success. Do not relax a real API constraint merely to pass a test.

## Step 8: Package and verify

```sh
tcpkg pack device-api -o device-api-1.0.0.tcpkg
tcpkg verify device-api-1.0.0.tcpkg
```

The archive contains _provider.yaml and tools/search.yaml. Source tests, secrets, README and LICENSE are excluded. Distribute the project LICENSE alongside packages containing copied template code. Existing archives are not overwritten; increment the provider version and use a new filename when distributing changes.

## Step 9: Verify in SIEM against a real service

Adapt the illustrative API contract to an actual vendor first. Upload through Tools → Add Tool → Upload Package. Create a credential with the real base_url/api_key and select the default credential where appropriate. Invoke the read-only tool with:

```json
{"filters":[{"field":"hostname","value":"web-01"}],"limit":10}
```

Check that the IDs belong to the intended environment. An error/http_status is not an empty search result. Separately verify network access, permissions, approvals and credential state. Do not use --live against the teaching domain. For developer-machine live tests, follow [testing](testing.md) and keep secrets in .local/credentials.json.

## Step 10: Extend the project

Add another tools/*.yaml with the same provider prefix, add tests and increment the package version. Do not assume a login tool establishes a shared session for later tools. Use [WASM](wasm.md) for unsupported signing/encoding or bounded complex logic; consider [MCP](mcp.md) for persistent state or background services.
