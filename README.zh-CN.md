# AI-SIEM Toolkit

[English](README.md) · [开发指南](docs/zh-CN/README.md) · [CLI 命令](docs/zh-CN/cli.md)

用于 SecNova AI-SIEM 自定义工具开发的工具集：初始化项目、校验定义、测试逻辑、编译 Go WASM、生成 `.tcpkg` 工具包，并指导 AI 完成这些工作。

## 安装

在有发行版本后，从 [Releases](https://github.com/secnova-ai/ai-siem-toolkit/releases) 下载对应系统和架构的可执行文件；也可使用 Go 1.25.7 或更新版本构建：

```sh
git clone https://github.com/secnova-ai/ai-siem-toolkit.git
cd ai-siem-toolkit
go build -o tcpkg ./cmd/tcpkg
```

将可执行文件所在目录加入 PATH。使用预编译 CLI 开发 CEL/MCP 不需要 Go；Go WASM 编译需要 Go 1.25.7+。

## 创建第一个工具

```sh
tcpkg init my-api --runtime cel
tcpkg validate my-api
tcpkg test my-api
tcpkg pack my-api -o my-api-1.0.0.tcpkg
tcpkg verify my-api-1.0.0.tcpkg
```

模板中的 `/health` 是教学接口，测试使用模拟响应。请根据目标服务的真实文档修改接口、认证和描述，再上传至 **Tools → Add Tool → Upload Package**，配置凭证并执行只读工具验证。

仓库包含 CLI、公共包处理与校验库、Go WASM SDK、三种项目模板、可运行示例、中英文文档，以及 [AI 创建工具 skill](skills/create-custom-tool/SKILL.md)。

v0.1 对齐 Tool Center dev/3.0.7 的工具创作协议。本地测试验证工具逻辑，不模拟平台的租户权限、审批、OAuth 授权或部署网络。详见 [兼容说明](docs/zh-CN/compatibility.md)。

## 使用 AI

将完整 `skills/create-custom-tool` 目录安装到所用 Agent 的 skill 目录，或让 AI 阅读其中的 `SKILL.md`，提供接口文档和期望操作。skill 包含独立参考资料，不需要访问公司内部仓库。

## 开发检查

```sh
go test ./...
go vet ./...
python scripts/check_docs.py
```

测试会真实编译 Go WASM 模板，并在 16 MiB 内存限制下执行。`go test -short ./...` 跳过编译测试。默认不会访问真实设备 API。
