# CLI 命令参考

选项放在目录参数后。不认识或重复的选项会报错；成功退出码为 0，失败为 1。

| 命令 | 行为 |
|---|---|
| `tcpkg init DIR --runtime cel` | 创建 CEL 项目，默认类型也是 CEL |
| `tcpkg init DIR --runtime wasm --language go` | 创建 Go WASM 项目，带 SDK 源码 |
| `tcpkg init DIR --runtime mcp` | 创建动态发现 MCP Provider |
| `tcpkg validate DIR` | 离线校验定义、Schema、CEL，输出描述警告；允许尚未编译的 WASM 缺失 |
| `tcpkg build DIR` | 从 `src/NAME` 编译 `wasm/NAME.wasm`，需要 Go |
| `tcpkg test DIR [--tool ID]` | 用模拟 HTTP 执行匹配的 `tests/*.yaml` |
| `tcpkg test DIR --live --credentials FILE [--tool ID]` | 从本地 JSON 文件读取凭证，发出真实 HTTP 请求 |
| `tcpkg pack DIR [-o FILE]` | 校验、打包并验证；默认在当前目录输出 `PROVIDER-VERSION.tcpkg` |
| `tcpkg verify FILE [FILE...]` | 检查 ZIP CRC、路径、体积、定义及引用文件 |
| `tcpkg version` | 显示版本和创作协议基线 |

`init` 拒绝覆盖已有目录。目录名成为 Provider ID：小写字母开头，后面只包含小写字母、数字和短横线，最多 64 字符。生成后修改显示名称和厂商。

`build` 调用本地 Go，设置 `GOOS=wasip1 GOARCH=wasm CGO_ENABLED=0`，编译成功后替换 WASM 产物，不自动安装编译器。`pack` 不自动编译或运行测试，也不覆盖已有包；更换输出名称或自行移除废弃产物。分发修改后的工具前递增 `_provider.yaml` 中的版本。

旧内部 CLI 的 `--wasm-dir` 不在 v0.1 范围内，编译产物统一放入项目的 `wasm/`。
