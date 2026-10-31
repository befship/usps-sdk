# usps-sdk

[English](./README.md) | 中文

非官方 Go SDK，面向 [USPS REST APIs v3](https://developer.usps.com/)。从 OpenAPI 生成客户端，并提供 OAuth2 v3、Payments 鉴权、重试与稳定 facade。

> **与 USPS 无隶属关系。** 需自备开发者账号。见 [DISCLAIMER.md](./DISCLAIMER.md)。
> **不要使用旧版 eVS / Web Tools XML。**

## 布局

| 路径 | 作用 |
|------|------|
| `usps/` | 手写 facade |
| `internal/generated/...` | `oapi-codegen` 产物（已提交） |
| `scripts/` + `generate.sh` | 解析 YAML → 预处理 → 生成 |
| `.cache/specs/` | OpenAPI 种子（**不入库**） |

## 重新生成

```bash
SPEC_DIR=/path/to/portal-downloads make generate
SPEC_LOCAL_DIR=/path/to/yamls make generate
```
