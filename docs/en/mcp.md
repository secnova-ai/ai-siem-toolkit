# Connect an existing MCP server

The MCP starter declares `mcp_discovery: true` and no tools. Credentials supply `mcp_endpoint` (the complete HTTP endpoint) and `api_key`. `auth_strategy.headers` maps that key into a Bearer header. Adapt the schema/strategy if your server uses a different method. Packaging does not contact the server.

After upload, configure a credential in SIEM, authorize it if OAuth is used, and refresh Available Tools. Verify tool names, input descriptions, count and at least one read-only invocation. A successful package validation or a skipped credential probe does not prove authentication or tool discovery works.

For static definitions, remove `mcp_discovery: true`, add `tools/NAME.yaml`, set `runtime_type: mcp` and `runtime_config.tool_name` to the exact remote name. Set the input/output schemas from the remote server. A fixed `runtime_config.endpoint` is optional; otherwise the endpoint comes from credentials. Do not mix static definitions with dynamic discovery in this toolkit.

An MCP provider package does not contain or start an MCP server. Local `tcpkg test` does not simulate MCP sessions, discovery or OAuth. Those flows must be tested against the actual server in SIEM. HTTP is the supported transport; a local stdio server needs an appropriate externally managed HTTP bridge/service.

## Walk-through: a Bearer-authenticated MCP service

### Step 1: Gather the service contract

Obtain the full MCP HTTP endpoint, authentication method and expected tool inventory from the service owner. An ordinary REST endpoint is not an MCP endpoint. This example uses a Bearer token. OAuth services need their actual authorization flow; a client secret is not a Bearer token.

### Step 2: Initialize the project

```sh
tcpkg init my-mcp --runtime mcp
```

Dynamic discovery gets tool definitions from the server, so do not create tools/*.yaml. Replace the root _provider.yaml with this complete file. Enter real tokens only in platform credentials; the source declares fields rather than secret values.

<!-- tutorial-file: _provider.yaml -->
```yaml
provider_id: my-mcp
version: "1.0.0"
vendor: Example
name: my-mcp
description: Discover and call tools from an existing MCP service.
category: utility
supported_locations: [saas]
mcp_discovery: true
credential_schema:
  mcp_endpoint:
    type: url
    required: true
    description: Complete MCP HTTP endpoint, e.g. https://mcp.example.com/mcp.
  api_key:
    type: secret
    required: true
    description: Bearer token accepted by the MCP service.
auth_strategy:
  headers:
    Authorization: "Bearer {{creds.api_key}}"
credential_test:
  type: skip
```

### Step 3: Understand the key fields

- mcp_discovery: true asks the server for tools and their schemas.
- credential_schema.mcp_endpoint declares the full endpoint, including its path.
- credential_schema.api_key holds the Bearer token accepted by this example service.
- auth_strategy.headers substitutes credential fields into request headers.
- credential_test.type: skip disables a separate credential probe; it does not prove authentication succeeded.

For X-API-Key authentication, adapt the header to the server contract. For an unauthenticated service, remove api_key and auth_strategy. Making the credential optional while retaining a header that references a missing value is insufficient.

### Step 4: Validate, package and upload

```sh
tcpkg validate my-mcp
tcpkg pack my-mcp -o my-mcp-1.0.0.tcpkg
tcpkg verify my-mcp-1.0.0.tcpkg
```

Upload the package, enter the endpoint/token in credentials, select the default credential and refresh Available Tools. Compare the discovered inventory with the server, then invoke a read-only tool and inspect its result. Dynamic mode has no local tool fixtures; local test cannot replace this check.

### Step 5: Diagnose connection issues

| Symptom | First checks |
|---|---|
| 401 / 403 | Token validity, header format and server permissions |
| 404 / initialize failure | Complete endpoint path and HTTP MCP protocol support |
| Zero discovered tools | Tools exposed to this account and synchronization details |
| Works locally but not in SIEM | Network access from the execution location; localhost refers to that machine, not your desktop |

## Optional: declare a fixed remote tool

For a fixed inventory, remove mcp_discovery: true from _provider.yaml and create tools/get_device.yaml. The complete tool file below uses an illustrative remote name and schema: replace them with the actual server contract. Do not add it to the dynamic project while retaining mcp_discovery: true.

```yaml
tool_id: my-mcp.get_device
name: Get device
description: Read a single device by immutable ID from the remote MCP server. Does not modify the device. Replace this sample name and schema with the actual server contract.
input_schema:
  type: object
  additionalProperties: false
  required: [device_id]
  properties:
    device_id:
      type: string
      minLength: 1
      description: Exact immutable ID returned by the server's inventory, not a hostname or IP address.
risk_level: low
blast_radius: none
runtime_type: mcp
runtime_config:
  tool_name: get_device
```

Run validate, pack and verify again, then verify the invocation in SIEM. tool_id identifies the local tool; runtime_config.tool_name selects the remote tool.
