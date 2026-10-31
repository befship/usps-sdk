#!/usr/bin/env python3
"""Light preprocess for USPS OpenAPI before oapi-codegen.

Normalizes servers (TEM + production) and ensures bearer security scheme.
"""
from __future__ import annotations

import argparse
import json
import sys
from pathlib import Path
from typing import Any

try:
    import yaml
except ImportError:
    yaml = None  # type: ignore


def load(path: Path) -> dict[str, Any]:
    text = path.read_text(encoding="utf-8")
    if path.suffix.lower() == ".json":
        doc = json.loads(text)
    else:
        if yaml is None:
            raise SystemExit("PyYAML required for YAML specs: pip3 install pyyaml")
        doc = yaml.safe_load(text)
    if not isinstance(doc, dict):
        raise SystemExit(f"invalid OpenAPI: {path}")
    return doc


def dump(path: Path, doc: dict[str, Any]) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    if path.suffix.lower() == ".json":
        path.write_text(json.dumps(doc, indent=2, ensure_ascii=False) + "\n", encoding="utf-8")
        return
    if yaml is None:
        raise SystemExit("PyYAML required")
    path.write_text(yaml.safe_dump(doc, sort_keys=False, allow_unicode=True), encoding="utf-8")


def set_servers(doc: dict[str, Any], tem: str, prod: str) -> None:
    doc["servers"] = [
        {"url": tem.rstrip("/"), "description": "USPS TEM (test)"},
        {"url": prod.rstrip("/"), "description": "USPS production"},
    ]


def ensure_bearer(doc: dict[str, Any]) -> None:
    comps = doc.setdefault("components", {})
    schemes = comps.setdefault("securitySchemes", {})
    if "bearerAuth" not in schemes and "BearerAuth" not in schemes:
        schemes["bearerAuth"] = {"type": "http", "scheme": "bearer", "bearerFormat": "JWT"}
    if "security" not in doc:
        key = "bearerAuth" if "bearerAuth" in schemes else "BearerAuth"
        doc["security"] = [{key: []}]


def main() -> int:
    ap = argparse.ArgumentParser()
    ap.add_argument("--in", dest="inp", required=True)
    ap.add_argument("--out", required=True)
    ap.add_argument("--server-tem", required=True)
    ap.add_argument("--server-prod", required=True)
    args = ap.parse_args()

    doc = load(Path(args.inp))
    set_servers(doc, args.server_tem, args.server_prod)
    ensure_bearer(doc)
    dump(Path(args.out), doc)
    print(f"→ {args.out}", file=sys.stderr)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
