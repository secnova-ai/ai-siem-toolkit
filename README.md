# AI-SIEM Toolkit

[简体中文](README.zh-CN.md) · [Authoring guide](docs/en/README.md) · [CLI reference](docs/en/cli.md)

Create, validate, test and package custom tools for SecNova AI-SIEM.
The `tcpkg` CLI supports CEL tools, Go WASM tools, and providers for existing MCP HTTP servers.
Packages use the existing `.tcpkg` format and are uploaded through SIEM's Tools page.

## Install

Download an executable for your OS/architecture from [Releases](https://github.com/secnova-ai/ai-siem-toolkit/releases) when a release is available, or build from source with Go 1.25.7 or newer:

```sh
git clone https://github.com/secnova-ai/ai-siem-toolkit.git
cd ai-siem-toolkit
go build -o tcpkg ./cmd/tcpkg
```

Move the executable into a directory on your PATH. CEL and MCP users need no Go compiler when using a prebuilt CLI. Go WASM authors need Go 1.25.7+.

## First tool

```sh
tcpkg init my-api --runtime cel
tcpkg validate my-api
tcpkg test my-api
tcpkg pack my-api -o my-api-1.0.0.tcpkg
tcpkg verify my-api-1.0.0.tcpkg
```

The generated `/health` example uses simulated HTTP. Replace it with your documented API before using it against a real service. Upload the package via **Tools → Add Tool → Upload Package**, configure a credential, and verify a read-only tool.

## Contents

| Directory | Purpose |
|---|---|
| `cmd/tcpkg` | CLI entry point |
| `pkg` | Reusable archive, definition, schema, credential and execution helpers |
| `sdk/go` | Go WASM HTTP SDK for the Tool Center host ABI |
| `templates` | Embedded, offline CEL / WASM / MCP starters |
| `examples` | Runnable example projects and test cases |
| `docs/en`, `docs/zh-CN` | End-to-end authoring documentation |
| `skills/create-custom-tool` | AI-assisted authoring skill |

The v0.1 authoring contract targets Tool Center dev/3.0.7. Local tests check tool logic; they do not simulate tenant RBAC, approvals, OAuth authorization, or deployment networking. See [compatibility](docs/en/compatibility.md).

## AI-assisted authoring

Give your assistant the [skill](skills/create-custom-tool/SKILL.md), your API documentation and intended operations. For agents with local skill discovery, copy the complete skill directory into their skill location. Its references are self-contained.

## Development

```sh
go test ./...
go vet ./...
python scripts/check_docs.py
```

Tests compile and execute the Go WASM starter under the same 16 MiB memory limit as Tool Center. `go test -short ./...` skips that compilation test. No real external API calls are made by default.

## License

The CLI, SDK, templates, examples, documentation and AI skill in this repository are licensed under [Apache-2.0](LICENSE). Third-party dependencies retain their own licenses.

Using `tcpkg` to create or package your own tools does not require you to open-source them. If you redistribute copied or modified toolkit code, templates or SDK code, retain the applicable license and notices and identify modified files as required by Apache-2.0. Generated source projects include a LICENSE copy for the supplied template/SDK code; it does not automatically license your independently authored code. When distributing a `.tcpkg` containing such code, provide the license alongside the package.
