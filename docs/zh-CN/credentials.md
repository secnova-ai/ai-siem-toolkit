# 凭证与认证

`credential_schema` 是字段映射，不是 JSON Schema 对象。定义用户安装后要填写的内容，支持 string、text、secret、url、boolean、number、integer、select。`required: true` 拒绝缺失和空值，可选字段默认值必须符合类型。URL 类型字段的主机也用于出站白名单。

```yaml
credential_schema:
  base_url:
    type: url
    required: true
    description: 服务根地址，不带结尾斜杠。
  api_key:
    type: secret
    required: true
    description: 在目标服务管理页面创建的只读 API Key。
```

包中只放字段定义，不放客户的秘密。不同账号或地址通过平台已有的多凭证能力管理。上传包归属当前租户，不要把租户 ID 拼进 Provider ID。

| 类型 | 认证行为 |
|---|---|
| CEL | 从 `creds` 取值；普通 HTTP 函数会将已有 access_token 作为 Bearer，`_h` 函数使用显式请求头 |
| WASM | stdin 提供 creds，程序自行构造请求头 |
| MCP | 平台 MCP 客户端应用认证配置，管理所支持的 OAuth 流程 |

API Key、Basic 必须引用正确字段。CEL Provider 中声明模板请求头，不等于 `_h` 表达式会自动使用它；WASM 也不会自动继承请求头。

平台 CEL 支持现有 client-credentials 换 token 机制，需要提供 token URL 和相应客户端凭证。授权码 OAuth 在 SIEM 中按部署支持的凭证流程配置授权。CLI 不交换、刷新或保存 token：模拟测试用假 access_token，真实联调使用已有有效 token。续期和回调必须在平台实际验证。

`credential_test.type` 可选 `http_probe`、`oauth2_client_credentials`、`skip`。探测使用已知无副作用接口，它是连通性检查，不是保存登录状态的脚本。不能通过删除、封禁等动作测试凭证。

真实联调凭证放 `.local/credentials.json`，该目录被忽略且不会打包。本地 live 测试始终校验证书并拒绝重定向，不模拟平台其它可选证书策略。不要将真实 token 放进样例、命令行、截图和日志。
