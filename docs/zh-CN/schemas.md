# 逐步设计输入输出参数

这一页的 YAML 都是工具文件内的片段，不能单独作为完整工具上传。完整组合见 [教程](tutorial.md)。

## 1. 从调用值反推类型

希望用户传入 `{"device_id":"device-42","limit":10}`，对应 Schema：

```yaml
input_schema:
  type: object
  additionalProperties: false
  required: [device_id, limit]
  properties:
    device_id:
      type: string
      minLength: 1
      description: 资产查询接口返回的精确设备 ID，不是 IP 或主机名。
    limit:
      type: integer
      minimum: 1
      maximum: 100
      description: 本次最多返回的记录数，单位为条，范围 1 到 100。
```

type: object 表示入参是对象；properties 声明字段；required 列出必须出现的字段名。`additionalProperties: false` 拒绝拼错或多余的参数。最外层 required 不会自动让嵌套对象的字段必填。

| JSON 值 | Schema type |
|---|---|
| `"device-42"` | string |
| `10` | integer |
| `0.5` | number |
| `true` | boolean |
| `{"id":"a"}` | object |
| `["a","b"]` | array |
| `null` | null，或允许 null 的联合类型 |

type 不能写 int、float、bool。`secret`、`url` 是凭证字段类型，也不能直接作为 JSON Schema type。

## 2. 给枚举值和可选值明确语义

```yaml
input_schema:
  type: object
  properties:
    state:
      type: string
      enum: [open, closed]
      description: 可选过滤状态，open 为未关闭，closed 为已关闭；省略时查询全部状态。
    note:
      type: [string, "null"]
      description: 可选备注；省略表示不更新，null 表示清除，空字符串表示写入空备注。
```

可选意味着不在父对象 required 中，不意味着可以传 null。`default` 在参数 Schema 中也不会自动填进 params；要么显式要求用户提供，要么在 CEL 中处理缺省，例如 `has(params.state) ? params.state : "open"`。不要直接读取可能不存在的字段，也不要直接把 null 拼成 URL 或请求头。

## 3. 对象里的字段再写一层 properties

```yaml
input_schema:
  type: object
  required: [target]
  properties:
    target:
      type: object
      description: 本次操作的单个目标设备。
      additionalProperties: false
      required: [device_id]
      properties:
        device_id:
          type: string
          description: 查询工具返回的设备 ID。
        reason:
          type: string
          description: 可选审计原因，省略时使用设备服务的默认原因。
      examples:
        - device_id: device-42
          reason: Confirmed incident
```

外层 required: [target] 只要求 target 对象存在；内层 required: [device_id] 要求对象内的 ID。正确调用值是 `{"target":{"device_id":"device-42"}}`，不是把这段 JSON 再包成字符串。

## 4. 对象数组使用 items

```yaml
input_schema:
  type: object
  required: [targets]
  properties:
    targets:
      type: array
      minItems: 1
      maxItems: 20
      description: 待查询的设备，1 至 20 项，每项使用精确 ID，不会自动分页。
      items:
        type: object
        required: [device_id]
        additionalProperties: false
        properties:
          device_id:
            type: string
            minLength: 1
            description: 资产查询返回的不可变设备标识。
      examples:
        - [{device_id: device-42}, {device_id: device-43}]
```

properties 属于对象，items 属于数组。minLength 限制字符串长度，minItems 限制数组元素数，不要混用。嵌套 JSON 每层都要写清字段含义，而不是只写“JSON 对象”。确实允许任意字典时，解释键值语义并用 additionalProperties 描述值类型。

## 5. 描述实际输出，不照抄原始接口所有字段

如果 CEL 提取后只返回 `{"devices":[{"id":"device-42"}]}`，输出 Schema 应描述这个提取后的值。HTTP body 默认是字符串，必须 decode_json 后才是对象；用对象 Schema 声明一个未经解析的字符串会导致本地测试失败。

有成功和错误两种返回时，可用 oneOf 描述互斥分支；确保各分支的 required 能区分它们。测试同时覆盖成功、空结果和错误，不能只验证一个“正常样本”。

## 6. 三次检查

1. `tcpkg validate` 检查 Schema 定义是否合法，但不能证明说明与真实 API 一致。
2. `tcpkg test` 检查输入实例、执行结果和输出 Schema；增加错误类型、遗漏必填、越界值的用例。
3. 在 SIEM 对真实只读接口核对字段含义和结果，不把模拟测试当作真实接入证明。

凭证的 `required: true` 与参数 Schema 的 `required: [name]` 是两种不同结构，详见 [认证配方](auth-recipes.md)。
