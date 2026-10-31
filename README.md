# usps-sdk

English | [中文](./README.zh-CN.md)

Unofficial Go SDK for [USPS REST APIs v3](https://developer.usps.com/), generated from OpenAPI specs and wrapped with OAuth2 v3, Payments auth, retries, and a stable facade.

> **Not affiliated with USPS.** You need your own Developer Portal credentials. See [DISCLAIMER.md](./DISCLAIMER.md).
> **Do not use legacy eVS / Web Tools XML.**

## Why this layout

| Path | Purpose |
|------|---------|
| `usps/` | Hand-written facade (OAuth, Payments, typed API accessors) |
| `specs/` | Portal OpenAPI YAML (**committed**) |
| `internal/generated/<api>/` | `oapi-codegen` output (committed so `go get` works) |
| `scripts/` + `generate.sh` | `specs/` → preprocess → generate |
| `.cache/specs/` | Local preprocess cache (gitignored) |

## Install

```bash
go get github.com/befship/usps-sdk@latest
```

## Specs in this repo

Official portal downloads live under [`specs/`](./specs/). See [`specs/README.md`](./specs/README.md).

完整对照见 [`specs/README.md`](./specs/README.md)。本次新增并已 codegen：

| Spec | Package |
|------|---------|
| `oauth2.yaml` | `oauth2` |
| `payments.yaml` | `payments` |
| `international-prices.yaml` | `internationalprices` |
| `locations.yaml` | `locations` |
| `service-standards.yaml` | `servicestandards` |
| `service-standards-files.yaml` | `servicestandardsfiles` |
| `shipment-tracking.yaml` | `shipmenttracking` |

## Quick start

```go
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/befship/usps-sdk/usps"
)

func main() {
	client := usps.New(usps.Config{
		ClientID:     os.Getenv("USPS_CLIENT_ID"),
		ClientSecret: os.Getenv("USPS_CLIENT_SECRET"),
		Environment:  usps.EnvTEM,
	})
	api, err := client.Addresses(context.Background(), usps.Creds{})
	if err != nil {
		panic(err)
	}
	fmt.Println("ready", usps.ServerURL(client.Environment(), "addresses"), api != nil)
}
```

## Endpoints

| Environment | Host |
|---|---|
| TEM / test (`tem`) | `https://apis-tem.usps.com` |
| Production | `https://apis.usps.com` |

Per-API bases (e.g. `.../labels/v3`) come from `ServerURL(env, package)`.

- Auth: OAuth 2.0 V3 (`POST /oauth2/v3/token`)
- Labels need Payments (`POST /payments/v3/payment-authorization`)

## Regenerate

```bash
make generate

# Optional: use another directory of portal YAMLs
SPEC_DIR=/path/to/usps-openapi-downloads make generate
```

## License

Hand-written code: MIT. Generated clients / committed OpenAPI: subject to USPS API terms. See `LICENSE` and `DISCLAIMER.md`.
