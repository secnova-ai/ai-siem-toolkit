# 兼容范围

初始基线为 Tool Center dev/3.0.7，保留 ZIP 格式的 `.tcpkg` 目录协议。面向客户编写 CEL、Go WASM、MCP，不等于支持导出平台全部内部特性。Toolkit 版本与 Provider 版本独立。

包解析、凭证类型/默认值校验、JSON Schema 校验、CEL 函数声明及 WASM HTTP SDK 来自对应 Tool Center 实现；公开代码不依赖数据库和内部服务。CLI 增加更严格的创作检查，可能拒绝旧服务曾接收的宽松定义，并不替代服务端校验。

公共入口包括 `pkg/tcpkg.ParsePackage`、`Load`、`ValidateFiles`、`Pack`、`Verify`，以及 schema、credential、cel 包的校验方法。服务端复用时应固定经过评审的版本，保留租户级校验。本仓库的发布不会自动修改现有服务器依赖。

模拟测试运行真实 CEL 和编译后的 WASM，仅替换 HTTP 响应；不覆盖 OAuth 续期、RBAC、审批、租户隔离、凭证激活、部署级 SSRF 策略和 MCP 协议会话。live 模式显式读取凭证、严格 TLS、拒绝重定向，不能宣传为完整模拟生产环境。

发布前运行全部测试、跨平台编译，检查文档链接和 SDK 模板同步，并通过目标 Tool Center 的解析/加载器检查生成的 CEL/WASM/MCP 包。部署上的上传和实际调用仍需单独验收。

平台新增字段时先核对实现，再补支持、测试和文档，不要为兼容一个字段直接取消所有未知字段校验。
