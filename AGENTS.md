# Working on AI-SIEM Toolkit

Read docs/en/compatibility.md before changing the package format or execution contract.
Keep CLI validation, embedded templates, examples, bilingual documentation and skill references consistent.
Run go test ./..., go vet ./..., and python scripts/check_docs.py for runtime or contract changes.
Do not claim local tests prove platform authentication, tenant isolation or approval behavior.
Never include customer credentials, internal service addresses or private application dependencies in templates, packages or documentation.
Do not introduce a new package format. Maintain Tool Center compatibility and cover changes with meaningful contract fixtures.
