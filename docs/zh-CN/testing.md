# 测试与排错

每个 `tests/*.yaml` 是一个用例：指定工具、params、creds、按顺序排列的 HTTP 预期，以及完整输出。未填写 expect 表示期望 JSON null，不是跳过比较。少发或多发请求都失败。请求头检查列出的字段，URL 和请求体精确比较。模拟测试绝不回退到真实网络。

```yaml
name: Read health
tool: my-api.health
params: {}
creds:
  base_url: https://api.example.com
  api_key: example-not-a-real-secret
http:
  - request:
      method: GET
      url: https://api.example.com/health
      headers: {X-API-Key: example-not-a-real-secret}
    response:
      status: 200
      body: '{"status":"ok"}'
expect: {status: ok}
```

用 `expect_error: 子串` 断言输入校验或执行错误，此时不比较输出；非预期请求不会被 expect_error 放过。至少覆盖成功、空结果、错误输入、认证拒绝、响应格式错误。测试会校验输入输出 Schema；输出校验是开发检查，不代表平台每次调用都强制校验输出。

```sh
tcpkg test my-api
tcpkg test my-api --tool my-api.health
tcpkg test my-api --live --credentials my-api/.local/credentials.json
```

live 模式使用真实凭证和 HTTP，保留输入与输出断言。先筛选只读工具，确认测试动作符合本次意图。CLI 不刷新 OAuth，拒绝重定向、严格校验证书，响应上限 1 MiB。断言失败不打印实际输出和秘密；WASM stderr 不直接展示，避免客体程序泄露凭证。

| 错误 | 检查方向 |
|---|---|
| CEL 函数未声明 | 使用支持的函数，或选择 WASM |
| 主机不在白名单 | 将地址声明为 URL 凭证，或配置 allowed_domains |
| 请求不匹配 | 方法、完整 URL、精确请求体及指定请求头 |
| 输出不匹配 | 响应解析、成功/错误 Schema、预期值类型 |
| 缺失 WASM | 执行 build，核对 src/NAME 和 artifact_ref |
| WASM 内存错误 | 减少依赖及缓冲，编译通过不等于能在沙箱运行 |
| stdout 非法 | 仅输出一个 JSON，日志写 stderr |
| MCP 不支持本地执行测试 | 校验打包后在 SIEM 中发现并调用 |

最终仍需验证租户访问、权限与审批、凭证状态与默认值、OAuth 过期、目标网络可达性。
