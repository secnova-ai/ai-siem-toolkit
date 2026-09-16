# Design useful tools

Start from user tasks, not a complete list of vendor endpoints. Group endpoints into meaningful, independently usable actions such as finding devices or isolating a specified device. Avoid tools that only return an intermediate login token. If target selection, permissions, behavior or destructive scope is ambiguous, clarify it before implementation. Confirm the proposed tool list and side effects before creation; an explicit prior approval of that exact list is sufficient.

Use a short stable provider ID such as `firewall`, and tool IDs such as `firewall.block_ip`. Descriptions must state when to use the action, what it changes, prerequisites and the meaning of its result. Human-readable text can be concise; AI-facing descriptions must retain required technical detail.

For every parameter describe format, units, valid values, where its value comes from, and the difference between omission, null and empty values. Use enums and bounds when the API defines them. Do not hide structured input in an undocumented string.

```yaml
input_schema:
  type: object
  required: [targets]
  properties:
    targets:
      type: array
      minItems: 1
      description: Devices to isolate, selected from the inventory result. Every entry affects one device.
      items:
        type: object
        additionalProperties: false
        required: [device_id]
        properties:
          device_id:
            type: string
            description: Exact immutable device ID returned by inventory; not hostname or IP.
          reason:
            type: string
            description: Optional audit reason. Omit to let the service use its default; do not include secrets.
      examples:
        - [{device_id: "device-42", reason: "Confirmed incident"}]
```

Risk follows the real effect, not just HTTP method. A read-only GET is normally low, configuration writes normally medium, deletion or isolation normally high. A POST search may be low and a destructive GET high. Use `force_approval`, `irreversible`, `blast_radius` and `sensitive_params` where their actual meaning applies. Do not change credential or platform RBAC systems to implement a tool.
