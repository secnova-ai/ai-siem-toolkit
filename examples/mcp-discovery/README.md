# mcp-discovery

This provider connects an existing MCP HTTP server using dynamic tool discovery.

1. `tcpkg validate .`
2. `tcpkg pack .`
3. Upload the package in SIEM, configure the MCP endpoint and token in Credentials.
4. Refresh Available Tools and execute a read-only tool to verify the connection.

No tools/ directory or local execution test is required. A skipped credential probe
is not proof that discovery or invocation works. OAuth setup must be verified in SIEM.
