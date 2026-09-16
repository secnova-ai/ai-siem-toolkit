# Connect an existing MCP server

The MCP starter declares `mcp_discovery: true` and no tools. Credentials supply `mcp_endpoint` (the complete HTTP endpoint) and `api_key`. `auth_strategy.headers` maps that key into a Bearer header. Adapt the schema/strategy if your server uses a different method. Packaging does not contact the server.

After upload, configure a credential in SIEM, authorize it if OAuth is used, and refresh Available Tools. Verify tool names, input descriptions, count and at least one read-only invocation. A successful package validation or a skipped credential probe does not prove authentication or tool discovery works.

For static definitions, remove `mcp_discovery: true`, add `tools/NAME.yaml`, set `runtime_type: mcp` and `runtime_config.tool_name` to the exact remote name. Set the input/output schemas from the remote server. A fixed `runtime_config.endpoint` is optional; otherwise the endpoint comes from credentials. Do not mix static definitions with dynamic discovery in this toolkit.

An MCP provider package does not contain or start an MCP server. Local `tcpkg test` does not simulate MCP sessions, discovery or OAuth. Those flows must be tested against the actual server in SIEM. HTTP is the supported transport; a local stdio server needs an appropriate externally managed HTTP bridge/service.
