# 写工具 YAML 前需要掌握的语法

本页讲文件怎么写；字段的业务含义见 [字段参考](reference.md)，完整项目见 [逐步教程](tutorial.md)。使用 UTF-8 文本和 `.yaml` 扩展名，编辑器建议开启空格及缩进显示。

## 1. 层级由缩进决定

同层字段左对齐，下一层通常缩进两个空格，不使用 Tab：

```yaml
credential_schema:
  base_url:
    type: url
    required: true
  api_key:
    type: secret
    required: true
```

base_url 和 api_key 都属于 credential_schema；type/required 分别属于各自的凭证字段。若把 api_key 与 credential_schema 对齐，它就变成了不存在的 Provider 顶层字段，校验会拒绝。

## 2. 字符串、数字、布尔值不是同一种值

```yaml
version: "1.0.0"
force_approval: false
runtime_config:
  timeout_seconds: 30
```

版本是字符串，开关是布尔值，超时是整数。不要把 `false` 写成 `"false"` 或把 30 写成 `"30"`。设备编号如 `"00123"` 应加引号保留前导零。包含 `: `、` #` 的描述、以 `{` 或 `[` 开始的字符串、容易被当成特殊值的 `"null"`、`"true"`、`"on"`，建议加引号。

`null` 表示空值，`""` 表示空字符串，`{}` 表示空对象，`[]` 表示空数组，省略字段表示没有提供。这些值在接口中可能有不同含义，不能随意互换。

## 3. 数组可以竖写或横写

下面是等价写法：

```yaml
supported_locations:
  - saas
  - edge
```

```yaml
supported_locations: [saas, edge]
```

对象数组要保持同一个元素内的字段对齐：

```yaml
filters:
  - field: hostname
    value: web-01
  - field: ip
    value: "192.0.2.10"
```

这里有两个对象，不是四个数组元素。这个片段是调用数据，不是可以直接放进工具顶层的字段。

## 4. 多行 CEL 使用块字符串

```yaml
runtime_config:
  program: |
    get_h(
      creds.base_url + "/health",
      {"X-API-Key": creds.api_key}
    ).body.decode_json()
```

`|` 保留换行，适合表达式和注释。`>-` 通常折叠相邻普通行的换行，并去掉末尾换行；更深缩进仍可能保留换行。两种形式都应让完整表达式通过 `tcpkg validate`，不要依赖折叠来拼接字符串或使用行尾注释。表达式内容必须比 program 多缩进一层。

## 5. 区分 YAML、CEL 和模板占位符

| 所在位置 | 正确写法 | 谁解释它 |
|---|---|---|
| Provider auth_strategy.headers | `Authorization: "Bearer {{creds.token}}"` | 平台的凭证模板替换 |
| CEL program | `{"Authorization": "Bearer " + creds.token}` | CEL 表达式 |
| 测试 creds | `token: example-token` | 测试运行器读取实际值 |

同样地，input_schema 描述参数格式，tests 中的 params 才是参数值。不要把两者复制错位置。

## 6. 不能这样写

以下是反例，用文本块展示，不要复制到文件：

```text
credential_schema:
api_key:                 # 缩进错了，变成顶层字段
  type: secret

risk_level: low
risk_level: high         # 同一对象里重复键

input_schema:
  type: int              # JSON Schema 应使用 integer
  required: true         # 此处应为字段名数组，不是布尔值
```

每个文件只放一个 YAML 文档，不使用 `---` 拼接多个工具，也不使用 `&anchor` / `*alias` 复用节点。本工具集拒绝别名、重复键和未知的结构字段。一个工具一个 `tools/*.yaml`。

## 7. 用 CLI 确认，不只看编辑器颜色

保存后运行 `tcpkg validate 项目目录`。编辑器能检查 YAML 语法，但不一定知道 Provider 字段、JSON Schema 和 CEL 函数是否正确。先修复首个错误再运行，错误修复后继续 [完整教程](tutorial.md)。
