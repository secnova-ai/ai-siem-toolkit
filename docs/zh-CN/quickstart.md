# 快速开始

按仓库 README 安装 `tcpkg` 后，创建 CEL 项目：

```sh
tcpkg init my-api --runtime cel
tcpkg validate my-api
tcpkg test my-api
```

`_provider.yaml` 描述服务、版本和凭证；`tools/health.yaml` 定义只读操作；`tests/health.yaml` 提供假凭证、预期请求和模拟响应。模板没有连接任何真实厂商 API。

根据真实接口文档替换 `/health`、认证请求头和输出结构，并修改测试。调用参数放在输入 Schema 中，秘密从 `creds` 读取，不要把 token 写入 CEL 或源码。

```sh
tcpkg pack my-api -o my-api-1.0.0.tcpkg
tcpkg verify my-api-1.0.0.tcpkg
```

在 SIEM 中选择 Tools → Add Tool → Upload Package。安装后打开工具详情的 Credentials 页签，创建真实的服务地址和 API Key 凭证，按需将有效凭证设为默认。执行只读工具，与源接口结果核对。

WASM 使用 `tcpkg init my-wasm --runtime wasm --language go`，先执行 `tcpkg build my-wasm`，再测试打包。MCP 使用 `tcpkg init my-mcp --runtime mcp`，直接校验打包，安装配置凭证后用 Refresh tools 发现工具。动态 MCP 不需要本地执行用例。

调用真实 API 前阅读 [测试说明](testing.md)。
