# Package, install and update

Source projects contain `_provider.yaml`, `tools/*.yaml`, optional `_assets/` icons and `wasm/` artifacts, plus authoring-only source/tests/README files. The archive preserves the platform layout at its root; it must not add an outer project directory.

```text
_provider.yaml
tools/health.yaml
wasm/health.wasm       # when referenced
_assets/icon.svg       # when referenced
```

`pack` includes only definitions and referenced icons/WASM. It excludes source, tests and `.local` credentials. It performs validation and then verifies the produced ZIP, writing the destination only when complete. Existing output files are not overwritten. Multiple tools can reference the same WASM/icon without duplicate ZIP entries.

The archive parser limits compressed/expanded total size to 256 MiB and individual entries to 32 MiB. Deployment reverse proxies/UI may impose lower limits; check your deployment. Verify checksums/CRC, rejects traversal/duplicate paths and validates definitions. A valid archive does not imply a service is reachable or credentials work.

Upload through SIEM → Tools → Add Tool → Upload Package, configure credentials, then verify execution. Use a new semantic version for changed releases. Keep old source and packages for a controlled restoration using the platform's supported update procedure; the CLI does not promise an automatic rollback mechanism.

Distribute only the `.tcpkg` to tool users. Developers can separately receive the source repository. Package publication does not grant target-service permissions; users still configure appropriate credentials in their tenant.
