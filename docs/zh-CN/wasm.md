# Go WASM 工具开发

```sh
tcpkg init my-wasm --runtime wasm --language go
tcpkg build my-wasm
tcpkg test my-wasm
tcpkg pack my-wasm
```

模板包含独立 Go 模块、`sdk/` 源码、`src/health/main.go`、工具定义和 HTTP 模拟测试，无需公司私有依赖。`artifact_ref: health.wasm` 对应 `src/health/`。多个独立入口使用不同产物名称及源码目录。

编译目标是 `GOOS=wasip1 GOARCH=wasm`，不是浏览器使用的 `GOOS=js`。平台提供 WASI 和 `toolcenter.http_call` 宿主函数。程序从 stdin 读取 `{"params": {...}, "creds": {...}}`，向 stdout 输出一个 JSON 值，日志写 stderr。执行失败以非零状态退出。不要把日志混入 stdout，也不要输出凭证。

HTTP 使用 `sdk.Get`、`sdk.Post`、`sdk.Put`、`sdk.Delete` 或 `sdk.Do(method, url, headers, body)`，宿主检查出站主机白名单。响应提供 Status、Body、OK、Headers、Header(name)。SDK 不自动重试，避免写操作重复执行。先检查状态码，再解析响应。

平台限制为 16 MiB 线性内存、30 秒执行、1 MiB stdout、64 KiB stderr；SDK 响应缓冲区为 4 MiB。本地 WASM 测试应用相同的内存、输出、时间限制及 HTTP ABI。即使编译成功，大型依赖和临时缓冲也可能超限。

沙箱不挂载文件系统，不保留调用状态，不运行本机进程，也不提供任意 socket。使用 SDK 宿主 HTTP，不直接使用 Go 网络客户端。单次调用应是有界工作，持久会话、后台任务应使用外部服务或 MCP。

发布后也可引用 `github.com/secnova-ai/ai-siem-toolkit/sdk/go`。模板直接附带 SDK 源码，因此初始化和示例编译不需要下载私有依赖；升级时保持 SDK 与平台 ABI 一致。

## 实操：从模板到自己的 WASM 动作

### 第 1 步：检查环境并生成项目

执行 `go version`，确认 Go 至少为 1.25.7，再执行页面开头的 init。打开生成的文件：

| 文件 | 你要修改的内容 |
|---|---|
| `_provider.yaml` | 名称、版本、凭证字段、无副作用的凭证探测 |
| `tools/health.yaml` | 工具用途、输入输出 Schema、风险及 artifact_ref |
| `src/health/main.go` | 业务逻辑 |
| `sdk/` | 随模板提供的 HTTP SDK，首次开发无需修改 |
| `tests/health.yaml` | 假参数、假凭证、HTTP 预期和返回断言 |
| `go.mod` | 模板自己的 Go 模块，源码导入 SDK 路径应与它一致 |

### 第 2 步：理解工具定义和源码如何关联

模板中这段配置的含义是：

```yaml
runtime_type: wasm
runtime_config:
  artifact_ref: health.wasm
```

CLI 将 `src/health` 编译为 `wasm/health.wasm`，打包时按 artifact_ref 找到文件。不要把 artifact_ref 写成 `wasm/health.wasm` 或磁盘绝对路径。换成另一个名称时，要同时修改源码目录、artifact_ref 和测试选择的工具 ID（如果也改了 ID）。

### 第 3 步：逐段阅读完整 Go 入口

下面是模板实际生成的入口，保存在 `src/health/main.go`：

```go
package main

import (
	"encoding/json"
	sdk "example.local/tool/sdk"
	"fmt"
	"os"
)

func run() error {
	var in struct {
		Params map[string]any `json:"params"`
		Creds  struct {
			BaseURL string `json:"base_url"`
			APIKey  string `json:"api_key"`
		} `json:"creds"`
	}
	if err := json.NewDecoder(os.Stdin).Decode(&in); err != nil {
		return err
	}
	if in.Creds.BaseURL == "" || in.Creds.APIKey == "" {
		return fmt.Errorf("base_url and api_key are required")
	}
	r, err := sdk.Get(in.Creds.BaseURL+"/health", map[string]string{"X-API-Key": in.Creds.APIKey})
	if err != nil {
		return err
	}
	if !r.OK {
		return fmt.Errorf("health API returned HTTP %d", r.Status)
	}
	var result struct {
		Status string `json:"status"`
	}
	if err = json.Unmarshal(r.Body, &result); err != nil {
		return fmt.Errorf("health API returned invalid JSON")
	}
	if result.Status == "" {
		return fmt.Errorf("health API omitted status")
	}
	return json.NewEncoder(os.Stdout).Encode(result)
}
func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

```

先把 stdin 解码为包含 Params/ Creds 的对象，再校验本次调用必须有的凭证值。`sdk.Get` 通过平台宿主执行 HTTP，不使用普通 Go 网络客户端。先检查 r.OK，再解析业务 JSON，最后仅将结果对象写到 stdout。main 在错误时写 stderr 并非零退出；不能输出秘密。

这个 health 动作没有业务参数，所以 Params 暂时没使用。要查询指定设备时，应同时给 input_schema 增加 device_id，在 Params 中读取并检查它，再按实际 API 正确编码到路径或请求体。输出变更也要同步 output_schema 和测试 expect，不只是改 Go 代码。

### 第 4 步：编译与模拟测试

```sh
tcpkg validate my-wasm
tcpkg build my-wasm
tcpkg test my-wasm
```

预期看到 `built wasm/health.wasm` 和 `PASS health.yaml`。只修改 YAML 时通常不需重新编译，改 Go 源码后必须 build；pack 不会帮你自动重新编译。测试运行的是编译产物，不是原生 Go 程序，因而能发现内存限制或宿主调用问题。

把模拟 response.status 改成 401，原来期望成功的测试应失败。若要保存为失败用例，使用 expect_error 匹配实际执行错误，不把 r.Body 当成功结果返回。运行器不会直接输出客体 stderr，所以排错时不要依靠打印秘密。

### 第 5 步：新增第二个 WASM 工具

例如另一个动作使用 `inventory.wasm`：创建 `src/inventory/main.go` 和 `tools/inventory.yaml`，后者的 artifact_ref 写 inventory.wasm，再创建 tests/inventory.yaml。同一 Provider 下工具 ID 必须唯一。运行 build 会编译各被引用产物；修改了源码但忘记重新 build，测试和打包都会继续使用旧二进制。

### 第 6 步：分发并验证

```sh
tcpkg pack my-wasm -o my-wasm-1.0.0.tcpkg
tcpkg verify my-wasm-1.0.0.tcpkg
```

上传后配置凭证，先运行只读动作，核对结果。开发机能够运行 Go 并不代表平台沙箱能运行任意依赖；保留本地 WASM 测试，发现资源或能力边界时回到 [类型选择](runtimes.md)。
