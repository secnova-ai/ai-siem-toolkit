# Go WASM tools

```sh
tcpkg init my-wasm --runtime wasm --language go
tcpkg build my-wasm
tcpkg test my-wasm
tcpkg pack my-wasm
```

The starter includes a standalone Go module, SDK source in `sdk/`, `src/health/main.go`, a tool definition and a mocked HTTP test. No private Go modules are required. `artifact_ref: health.wasm` maps to `src/health/`. Add another tool with another artifact name and matching source directory for independent entry points.

Build target: `GOOS=wasip1 GOARCH=wasm`, not JavaScript's `GOOS=js`. Tool Center supplies WASI and a `toolcenter.http_call` import. The module reads one JSON object `{"params": {...}, "creds": {...}}` from stdin, writes one JSON value to stdout, and writes logs to stderr. Return a nonzero exit status on execution failure. Never print logs to stdout or credentials to either stream.

Use `sdk.Get`, `sdk.Post`, `sdk.Put`, `sdk.Delete`, or `sdk.Do(method, url, headers, body)` for HTTP. The host enforces allowed domains. `HTTPResponse` exposes Status, Body, OK, Headers and Header(name). Calls are not implicitly retried: retrying a write may duplicate an action. Check status before decoding JSON.

The runtime provides 16 MiB linear memory, a 30-second execution budget, 1 MiB stdout and 64 KiB stderr. The SDK response buffer is 4 MiB. Local tests apply the same WASM memory/output/time limits and the HTTP host ABI. Large dependency trees and large temporary buffers may exceed the budget even when compilation succeeds.

The sandbox has no mounted filesystem, persistent state, native process execution or general socket networking. Use host HTTP rather than Go's ordinary network client. Only bounded work belongs in a single invocation. Stateful or long-running workflows should use an external service/MCP.

The public SDK is also importable as `github.com/secnova-ai/ai-siem-toolkit/sdk/go` once a revision is published. Starter projects vendor its source so initialization and example builds work offline. Keep the SDK and platform ABI compatible when upgrading.

## Hands-on: turn the starter into your WASM action

### Step 1: Check the compiler and inspect the project

Run go version (1.25.7 or newer), then init as above. Edit _provider.yaml for identity/credentials, tools/health.yaml for schemas/risk, src/health/main.go for logic, and tests/health.yaml for fake HTTP and output assertions. The sdk directory is provided code; leave it unchanged initially. Its import path must agree with the generated go.mod.

### Step 2: Connect YAML to source

```yaml
runtime_type: wasm
runtime_config:
  artifact_ref: health.wasm
```

The CLI builds src/health into wasm/health.wasm. artifact_ref is a bare filename, not wasm/health.wasm or an absolute path. Rename the source directory and artifact_ref together; update test tool IDs if the action ID also changes.

### Step 3: Read the complete generated entry point

File src/health/main.go:

```go
package main

import (
	"encoding/json"
	sdk "example.local/tool/sdk"
	"fmt"
	"os"
)

func run() error {
	var in struct {
		Params map[string]any `json:"params"`
		Creds  struct {
			BaseURL string `json:"base_url"`
			APIKey  string `json:"api_key"`
		} `json:"creds"`
	}
	if err := json.NewDecoder(os.Stdin).Decode(&in); err != nil {
		return err
	}
	if in.Creds.BaseURL == "" || in.Creds.APIKey == "" {
		return fmt.Errorf("base_url and api_key are required")
	}
	r, err := sdk.Get(in.Creds.BaseURL+"/health", map[string]string{"X-API-Key": in.Creds.APIKey})
	if err != nil {
		return err
	}
	if !r.OK {
		return fmt.Errorf("health API returned HTTP %d", r.Status)
	}
	var result struct {
		Status string `json:"status"`
	}
	if err = json.Unmarshal(r.Body, &result); err != nil {
		return fmt.Errorf("health API returned invalid JSON")
	}
	if result.Status == "" {
		return fmt.Errorf("health API omitted status")
	}
	return json.NewEncoder(os.Stdout).Encode(result)
}
func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

```

Decode the stdin envelope, check credentials, call the host through sdk.Get, check HTTP status, decode business JSON, and write exactly one result to stdout. main writes errors to stderr and exits nonzero. Never print secrets. The SDK is used instead of an ordinary Go network client.

Health has no business parameters, so Params is unused. A device lookup must add device_id to input_schema, read/check it in Params and correctly encode it into the actual request. Update output_schema and test expect when changing the returned shape; modifying Go alone is insufficient.

### Step 4: Compile and test the module

```sh
tcpkg validate my-wasm
tcpkg build my-wasm
tcpkg test my-wasm
```

Expect built wasm/health.wasm and PASS health.yaml. Rebuild after Go changes; pack does not compile automatically. Tests execute the compiled WASM rather than the native Go program, exposing sandbox/host constraints.

Change the simulated status to 401: a success test should fail. To keep it as a negative fixture, use expect_error for the actual execution failure. Do not return an error body as success. Guest stderr is not directly printed by the runner; do not depend on logging secrets for diagnosis.

### Step 5: Add another action

For inventory.wasm, create src/inventory/main.go, tools/inventory.yaml with that artifact_ref, and tests/inventory.yaml. Keep tool IDs unique within the Provider. build compiles every referenced artifact. Without rebuilding, tests and pack keep using the old binary after source edits.

### Step 6: Package and verify in SIEM

```sh
tcpkg pack my-wasm -o my-wasm-1.0.0.tcpkg
tcpkg verify my-wasm-1.0.0.tcpkg
```

Upload, configure credentials and invoke a read-only action. Native Go support does not imply every dependency works in the sandbox. Keep WASM tests and revisit the documented runtime constraints when requirements exceed available capabilities.
