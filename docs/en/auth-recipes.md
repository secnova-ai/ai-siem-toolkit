# Turn authentication documentation into credential configuration

Choose the recipe that matches the API's actual requirements. Blocks are fragments for an existing Provider or tool, not complete files. Merge keys in place; never append a second credential_schema at the end of a YAML document.

## Step 1: Separate declaration from use

credential_schema tells the UI which fields users enter. CEL/WASM tells execution where to send those values. Naming a field api_key does not automatically inject the vendor's header. Secrets belong in credentials, not ordinary AI tool parameters or source code. Service URLs normally belong there too, enabling the platform's existing multiple-credential support.

## Step 2: API key in a header

Provider fragment:

```yaml
credential_schema:
  base_url:
    type: url
    required: true
    description: API origin without a trailing slash.
  api_key:
    type: secret
    required: true
    description: Read-only API key created in the target service.
```

Tool fragment:

```yaml
runtime_config:
  program: 'get_h(creds.base_url + "/health", {"X-API-Key": creds.api_key}).body.decode_json()'
```

This illustrates authentication only; production tools must also handle non-2xx as described in [CEL](cel.md). Use the documented header name, not a guessed universal Authorization header.

## Step 3: Fixed Bearer token

Replace the api_key schema field with token, retaining type: secret. Explicitly use:

```yaml
runtime_config:
  program: 'get_h(creds.base_url + "/health", {"Authorization": "Bearer " + creds.token}).body.decode_json()'
```

Users enter the token without the Bearer prefix. A static token is not automatically OAuth; its field name does not tell the platform how to refresh it.

## Step 4: Basic username and password/API token

Provider fragment:

```yaml
credential_schema:
  base_url:
    type: url
    required: true
  username:
    type: string
    required: true
    description: Username or email as required by the target service.
  password:
    type: secret
    required: true
    description: Password or API token, not a pre-encoded Basic string.
```

Tool fragment:

```yaml
runtime_config:
  program: >-
    get_h(creds.base_url + "/health",
      {"Authorization": "Basic " + base64_encode(creds.username + ":" + creds.password)}).body.decode_json()
```

Users need not calculate Base64. Encoding is not encryption; use the appropriate secure transport.

## Step 5: OAuth client_credentials

Only use this when the vendor supports that grant. Replace the illustrative token path with the documented endpoint:

```yaml
credential_schema:
  base_url:
    type: url
    required: true
  client_id:
    type: string
    required: true
  client_secret:
    type: secret
    required: true
auth_strategy:
  type: oauth2_client_credentials
  token_url: "{{creds.base_url}}/oauth/token"
credential_test:
  type: oauth2_client_credentials
```

The platform's existing flow obtains access_token. Ordinary get(url) injects an available access_token; _h requires the explicit `"Authorization": "Bearer " + creds.access_token` header. CLI tests do not exchange tokens: supply a fake access_token plus required fake client_id/client_secret in mock creds.

Do not relabel browser authorization-code/PKCE as client_credentials. If users must visit an authorization page, use SIEM's supported OAuth credential workflow and verify callback, scopes and refresh there. Adding an arbitrary YAML field does not establish browser authorization.

## Step 6: MCP uses Provider authentication templates

```yaml
credential_schema:
  mcp_endpoint:
    type: url
    required: true
  token:
    type: secret
    required: true
auth_strategy:
  headers:
    Authorization: "Bearer {{creds.token}}"
credential_test:
  type: skip
```

The braces are expanded by platform templates, not CEL. OAuth MCP credentials use the platform authorization workflow. skip does not validate discovery or invocation; refresh tools and invoke a read-only action after installing.

## Step 7: Probe credentials without side effects

Only use this when the documented health endpoint really validates authentication:

```yaml
credential_test:
  type: http_probe
  method: GET
  url: "{{creds.base_url}}/health"
  expected_status: 200
  headers:
    X-API-Key: "{{creds.api_key}}"
```

expected_status is an integer. A public page returning 200 to everyone does not validate an API key. A probe does not save a login cookie. Never use destructive operations for credential testing.

## Step 8: Diagnose in layers

Validate structure/types; mock-test actual headers; configure real credentials and their active/default status in SIEM; then verify addresses, execution location, API permissions, scopes and certificates. For OAuth, test token expiration and renewal too.

WASM receives creds on stdin and constructs these headers explicitly through its SDK. It does not inherit CEL expressions. Implement complex signing in WASM rather than inventing CEL functions.
