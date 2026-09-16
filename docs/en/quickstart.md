# Quick start

This page runs the starter quickly. To understand each authoring decision, start with the [complete tutorial](tutorial.md).

Install `tcpkg` as described in the repository README. Create a CEL project:

```sh
tcpkg init my-api --runtime cel
tcpkg validate my-api
tcpkg test my-api
```

`_provider.yaml` defines the service, its version and credentials. `tools/health.yaml` defines a read-only action. `tests/health.yaml` supplies fake credentials, the expected HTTP request and a simulated response. The template has no real connection to a vendor API.

To adapt it, locate a documented read-only API, replace `/health` and its authentication header, describe the actual output, and update the test. Every request parameter must have an input schema; every secret must come from `creds`. Do not put tokens in YAML programs.

```sh
tcpkg pack my-api -o my-api-1.0.0.tcpkg
tcpkg verify my-api-1.0.0.tcpkg
```

In SIEM, choose Tools → Add Tool → Upload Package. Open the installed provider's Credentials tab and create a credential with the real service URL and API key. Select an active default credential when appropriate. Invoke the read-only tool and compare its response with the source API.

For WASM, use `tcpkg init my-wasm --runtime wasm --language go`, then `tcpkg build my-wasm` before test/pack. For MCP, use `tcpkg init my-mcp --runtime mcp`; validate and pack directly, then configure the MCP endpoint and use Refresh tools in SIEM. MCP discovery does not have local execution fixtures.

See [testing](testing.md) before opting into real HTTP requests.
