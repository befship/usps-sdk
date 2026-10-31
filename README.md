# usps-sdk

English | [中文](./README.zh-CN.md)

Unofficial Go SDK for [USPS REST APIs v3](https://developer.usps.com/), generated from OpenAPI specs and wrapped with OAuth2 v3, Payments auth, retries, and a stable facade.

> **Not affiliated with USPS.** You need your own Developer Portal credentials. See [DISCLAIMER.md](./DISCLAIMER.md).
> **Do not use legacy eVS / Web Tools XML.**

## Why this layout

| Path | Purpose |
|------|---------|
| `usps/` | Hand-written facade (OAuth, Payments, typed API accessors) |
| `internal/generated/{addresses,labels,prices,tracking}/` | `oapi-codegen` output (committed so `go get` works) |
| `scripts/` + `generate.sh` | Resolve YAML → preprocess → generate |
| `.cache/specs/` | OpenAPI seeds (**gitignored**, never commit) |

Official OpenAPI files are **not** stored in git. Portal YAML is gated — use `SPEC_DIR` or `SPEC_LOCAL_DIR`.

## Install

```bash
go get github.com/befship/usps-sdk@latest
```

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
	fmt.Println("ready", usps.BaseURL(client.Environment()), api != nil)
}
```

## Endpoints

| Environment | Base URL |
|---|---|
| TEM / test (`tem`) | `https://apis-tem.usps.com` |
| Production | `https://apis.usps.com` |

- Auth: OAuth 2.0 V3 (`POST /oauth2/v3/token`)
- Labels need Payments (`POST /payments/v3/payment-authorization`)

## Regenerate

```bash
# Portal downloads (filename globs in scripts/manifest.yaml)
SPEC_DIR=/path/to/usps-openapi-downloads make generate

# Focused / local YAML seeds (addresses|labels|prices|tracking).openapi.yml
SPEC_LOCAL_DIR=/path/to/yamls make generate
```

## License

Hand-written code: MIT. Generated clients: subject to USPS API terms. See `LICENSE` and `DISCLAIMER.md`.
