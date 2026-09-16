# Authoring workflow

## First-time authors: follow this learning path

1. [YAML essentials](yaml-basics.md): indentation, types, block strings and common mistakes.
2. [Complete API-to-tool tutorial](tutorial.md): create five complete files, test success/empty/auth failure, package and install.
3. [Parameter schemas](schemas.md): objects, arrays, nested required, enums, null and defaults.
4. [Authentication recipes](auth-recipes.md): API Key, Bearer, Basic, OAuth, MCP and probes in the right locations.

Then use the focused guides below and the [structure contract](structure.md). CI extracts and validates complete YAML files from the tutorials, and executes the main tutorial's tests and packaging flow.


1. [Quick start](quickstart.md): create and install your first tool.
2. [Choose a runtime](runtimes.md): CEL, Go WASM or an existing MCP server.
3. [Design tools](design.md): meaningful actions, parameter descriptions and risk.
4. Implement with [CEL](cel.md), [WASM](wasm.md), or [MCP](mcp.md).
5. Define [credentials and authentication](credentials.md).
6. [Test and troubleshoot](testing.md), then [package and install](packaging.md).
7. Consult the [field reference](reference.md), [CLI reference](cli.md) and [compatibility contract](compatibility.md).

Run commands from the directory containing your project, or pass its path explicitly. `<directory>` always means the provider source directory, not a .tcpkg archive.
