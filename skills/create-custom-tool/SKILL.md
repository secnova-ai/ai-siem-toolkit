---
name: create-custom-tool
description: Create, modify, test and package custom tools for SecNova AI-SIEM. Use when a user wants to turn device or service APIs, API documentation, or request examples into tools, develop CEL or Go WASM tools, or connect an existing MCP server through a Provider package. Guides meaningful tool selection, clear parameter descriptions, credential handling and the tcpkg workflow.
---

# Create custom tools

Deliver a usable provider project, tests and verified `.tcpkg` package that implement the user's intended operations. Use the installed `tcpkg` CLI and the documented Tool Center contract. This skill does not deploy an MCP server or grant permission to invoke production write operations.

## Understand and confirm the tool set

Read the user's actual API documentation/request examples before writing code. Identify the intended outcome, real endpoints, authentication, required permissions and response shapes. Do not infer an undocumented endpoint or make up successful response fields. Record missing evidence and ask for it when correctness depends on it.

Select independently meaningful actions; avoid copying an entire API catalog or exposing intermediate login steps with no standalone user value. Clarify ambiguous targets, semantics and side effects. Before creating the project, present the proposed tool names, purposes, key inputs and effect/risk for confirmation. Reuse an explicit approval of that same list instead of asking again. Read [design](references/design.md) for description and nested-parameter requirements.

## Choose an implementable runtime

- CEL: simple HTTP calls and transformations expressible with the documented functions. Read [CEL](references/cel.md).
- Go WASM: complex parsing/signing or bounded multi-step calls that fit the WASI sandbox. Read [WASM](references/wasm.md).
- MCP: an existing HTTP MCP service, or recommend a separately developed service when persistent sessions/background tasks or capabilities outside the sandbox are required. Read [MCP](references/mcp.md).

Do not force an unsupported requirement into CEL. CEL is not JavaScript/Python and does not have arbitrary networking, HMAC, mutable session storage or invented helpers. WASM also has limits: no native processes, arbitrary sockets or cross-invocation persistence. Explain an actual capability gap and guide the user toward WASM or an external MCP service as appropriate.

## Author against the real contract

Read [credentials](references/credentials.md) and the relevant sections of the [field reference](references/reference.md). Initialize with `tcpkg init DIR --runtime cel|wasm|mcp`; Go WASM uses `--language go`. Read [CLI](references/cli.md) for exact command syntax. Modify an existing project without overwriting unrelated user files.

Treat tool execution as stateless across invocations. A login tool does not authenticate subsequent tools, and executing code does not write values back into the credential store. Platform-managed OAuth is a separate supported credential workflow. For other authentication use values supplied in `creds`; CEL `_h` helpers and WASM construct headers explicitly. Never embed real credentials in definitions, source or fixtures.

Describe each tool so a human or AI can determine when to use it and interpret its outcome. Describe each parameter's meaning, source, format, units, valid values and omission/null behavior. For objects and arrays, define nested properties/items with descriptions and a realistic example; never use an unexplained JSON string as a substitute. State side effects and choose risk based on actual behavior, including misleading HTTP methods.

Handle non-2xx and malformed responses deliberately. Do not return an API error as a successful operation. Do not fabricate unsupported platform fields, functions, automatic header injection or credential refresh behavior.

## Verify and deliver

Read [testing](references/testing.md). Add fixtures for success and relevant failure/empty/input cases, with fake credentials. Run `tcpkg validate DIR`, `tcpkg build DIR` for WASM, and `tcpkg test DIR`. Correct concrete failures; do not weaken checks merely to make tests pass. Mock tests must not contact the real service.

Real calls require a user-authorized target and action. Use `--live --credentials FILE` only within that scope; do not treat package creation approval as approval to execute destructive operations. MCP protocol, discovery and OAuth verification happens in SIEM against the real server; local validate is not an invocation test.

Run `tcpkg pack DIR -o FILE` and `tcpkg verify FILE`. Report implemented tools, artifact location, tests actually run, known capability gaps and the remaining SIEM credential/install verification steps. Read [compatibility](references/compatibility.md) when discussing what local checks prove. Do not claim installation or runtime success from a package build alone.
