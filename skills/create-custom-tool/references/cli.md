# CLI reference

Options follow the directory argument. Unknown or repeated options fail. Failures exit 1; success exits 0.

| Command | Behavior |
|---|---|
| `tcpkg init DIR --runtime cel` | CEL starter; CEL is the default |
| `tcpkg init DIR --runtime wasm --language go` | Go WASM starter with vendored SDK source |
| `tcpkg init DIR --runtime mcp` | Dynamic MCP discovery provider |
| `tcpkg validate DIR` | Offline definition/schema/CEL checks, description warnings; missing not-yet-built WASM is allowed |
| `tcpkg build DIR` | Build each referenced `wasm/NAME.wasm` from `src/NAME`; Go toolchain required |
| `tcpkg test DIR [--tool ID]` | Run matching `tests/*.yaml` against simulated HTTP |
| `tcpkg test DIR --live --credentials FILE [--tool ID]` | Execute real HTTP calls, with credentials from a local JSON file |
| `tcpkg pack DIR [-o FILE]` | Validate, package and verify; default filename is `PROVIDER-VERSION.tcpkg` in current directory |
| `tcpkg verify FILE [FILE...]` | Verify ZIP CRC, paths, size, definitions and referenced artifacts |
| `tcpkg version` | CLI version and authoring contract baseline |

`init` refuses any existing destination. The directory basename becomes the provider ID: lowercase letter first, followed by lowercase letters, digits or hyphens, at most 64 characters. Edit the display name/vendor afterwards.

`build` invokes the installed Go compiler with `GOOS=wasip1 GOARCH=wasm CGO_ENABLED=0`, replacing generated WASM artifacts only after compilation succeeds. It does not install compilers. `pack` does not run builds/tests and never overwrites an existing archive; choose another output or remove an obsolete artifact yourself. Source projects keep their version in `_provider.yaml`; bump it before distributing a changed package.

The legacy internal CLI's `--wasm-dir` option is not part of v0.1. Place compiled artifacts in the project's `wasm/` directory.
