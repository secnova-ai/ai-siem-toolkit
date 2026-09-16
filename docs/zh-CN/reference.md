# 定义字段参考

YAML 只能有一个文档，拒绝别名、重复键和结构体中的未知字段，避免拼错选项后静默无效。输入输出使用自包含 JSON Schema 2020-12，禁止外部引用。

## `_provider.yaml`

| 字段 | 含义 |
|---|---|
| provider_id | 必填，小写稳定 ID，最多 64 字符 |
| version | 必填，语义版本，例如 1.0.0 |
| vendor、name、category | 必填，厂商、显示名称、分类 |
| description | 服务用途 |
| supported_locations | saas / edge，实际部署需具备对应执行环境 |
| credential_schema | 凭证字段映射，见凭证文档 |
| auth_strategy | 平台认证配置，例如 MCP 请求头/查询模板或支持的 OAuth 设置 |
| credential_test | type 为 http_probe / oauth2_client_credentials / skip |
| api | 平台 API 元信息，不会增加 CEL 可用变量 |
| rate_limit | 元信息，不是本地执行限流器 |
| mcp_discovery | 动态发现，开启时不提供静态工具 |
| mcp_endpoint | 动态发现固定地址，可选 |
| icon、icon_dark | _assets/ 中的纯文件名，不含路径分隔符 |

凭证字段包含 type、required、default、description，以及 enum/options 等 UI 元信息。CLI 检查支持的字段类型和默认值类型，不穷举检查全部 UI 元信息。秘密字段的默认值不能放真实凭证。

`http_probe` 常用 method、url、headers、expected_status，模板引用凭证值；`skip` 不检查连通性。认证请求头和查询参数用 `{{creds.field}}`。`oauth2_client_credentials` 使用 token_url 和相应 client_id/client_secret，实际授权属于平台流程。

## `tools/NAME.yaml`

| 字段 | 含义 |
|---|---|
| tool_id | 必填，provider_id.action，包内唯一 |
| provider_id | 可省略，填写时必须匹配父 Provider |
| name | 必填，显示名称 |
| description、human_description | 面向 AI 的完整说明、可选界面短说明 |
| input_schema、output_schema | 输入输出结构、说明、约束和示例 |
| risk_level | 必填，low / medium / high / critical |
| runtime_type | 必填，cel / wasm / mcp |
| runtime_config | 运行时配置 |
| required_permissions | 已有平台权限标签，不能凭空创造新权限 |
| force_approval | 强制平台审批 |
| sensitive_params | 由平台按现有行为脱敏的参数名 |
| irreversible | 是否不可逆 |
| blast_radius | none / single / multi / global |
| allowed_caller_types | chat_agent / async_agent / automation / external_api |

CEL 配置包含 program（必填）、allowed_domains（主机名）、timeout_seconds。WASM 包含 artifact_ref（必填、纯 .wasm 文件名）、可选 artifact_checksum、allowed_domains。MCP 包含 tool_name（必填）、endpoint（可选）。平台生成的 OpenAPI 计划、报告数据源和旧运行时配置不在此模板协议内。
