#!/usr/bin/env python3
"""Light preprocess for USPS OpenAPI before oapi-codegen.

- Sets TEM + production servers (preserves API path suffix from official servers)
- Ensures bearer security scheme
- Normalizes kebab-case / dotted operationIds to PascalCase for stable Go names
"""
from __future__ import annotations

import argparse
import json
import re
import sys
from pathlib import Path
from typing import Any
from urllib.parse import urlparse

try:
    import yaml
except ImportError:
    yaml = None  # type: ignore

METHOD_VERB = {
    "get": "Get",
    "post": "Create",
    "put": "Update",
    "patch": "Patch",
    "delete": "Delete",
    "head": "Head",
    "options": "Options",
}

NON_ALNUM_RE = re.compile(r"[^A-Za-z0-9]+")
PATH_PARAM_RE = re.compile(r"\{([^}]+)\}")


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


def api_path_suffix(doc: dict[str, Any]) -> str:
    """Extract path suffix from first official server (e.g. /labels/v3)."""
    for s in doc.get("servers") or []:
        if not isinstance(s, dict):
            continue
        url = str(s.get("url") or "")
        parsed = urlparse(url)
        if parsed.path and parsed.path not in ("", "/"):
            return parsed.path.rstrip("/")
    return ""


def set_servers(doc: dict[str, Any], tem: str, prod: str) -> None:
    suffix = api_path_suffix(doc)

    def with_suffix(base: str) -> str:
        base = base.rstrip("/")
        if not suffix:
            return base
        # If caller already passed a path-qualified URL, keep it.
        parsed = urlparse(base)
        if parsed.path and parsed.path not in ("", "/"):
            return base
        return base + suffix

    doc["servers"] = [
        {"url": with_suffix(tem), "description": "USPS TEM (test)"},
        {"url": with_suffix(prod), "description": "USPS production"},
    ]


def ensure_bearer(doc: dict[str, Any]) -> None:
    comps = doc.setdefault("components", {})
    schemes = comps.setdefault("securitySchemes", {})
    if "bearerAuth" not in schemes and "BearerAuth" not in schemes:
        schemes["bearerAuth"] = {"type": "http", "scheme": "bearer", "bearerFormat": "JWT"}
    if "security" not in doc:
        key = "bearerAuth" if "bearerAuth" in schemes else "BearerAuth"
        doc["security"] = [{key: []}]


def fix_param_schema_name_collisions(doc: dict[str, Any]) -> int:
    """oapi-codegen errors when a parameter name matches a component schema name (e.g. ZIPCode)."""
    comps = doc.setdefault("components", {})
    schemas = comps.setdefault("schemas", {})
    schema_names = set(schemas.keys())
    fixed = 0
    colliding: set[str] = set()

    def collect_param_name(p: dict[str, Any]) -> None:
        if "$ref" in p:
            return
        name = p.get("name")
        if isinstance(name, str) and name in schema_names:
            colliding.add(name)

    for p in (comps.get("parameters") or {}).values():
        if isinstance(p, dict):
            collect_param_name(p)

    for item in (doc.get("paths") or {}).values():
        if not isinstance(item, dict):
            continue
        for key, val in item.items():
            if key == "parameters" and isinstance(val, list):
                for p in val:
                    if isinstance(p, dict):
                        collect_param_name(p)
            elif isinstance(val, dict):
                for p in val.get("parameters") or []:
                    if isinstance(p, dict):
                        collect_param_name(p)

    # Prefer renaming the shared schema; keep parameter names stable for query tags.
    for name in colliding:
        sch = schemas.get(name)
        if isinstance(sch, dict) and "x-go-name" not in sch:
            sch["x-go-name"] = to_pascal(name) + "Value"
            fixed += 1

    def fix_param(p: dict[str, Any]) -> None:
        nonlocal fixed
        if "$ref" in p:
            return
        name = p.get("name")
        if not isinstance(name, str) or name not in colliding:
            return
        schema = p.get("schema")
        if not isinstance(schema, dict):
            return
        if "x-go-name" in schema:
            return
        schema["x-go-name"] = to_pascal(name) + "Param"
        fixed += 1

    for p in (comps.get("parameters") or {}).values():
        if isinstance(p, dict):
            fix_param(p)

    for item in (doc.get("paths") or {}).values():
        if not isinstance(item, dict):
            continue
        for key, val in item.items():
            if key == "parameters" and isinstance(val, list):
                for p in val:
                    if isinstance(p, dict):
                        fix_param(p)
            elif isinstance(val, dict):
                for p in val.get("parameters") or []:
                    if isinstance(p, dict):
                        fix_param(p)
    return fixed


def to_pascal(s: str) -> str:
    parts = [p for p in NON_ALNUM_RE.split(s) if p]
    return "".join(p[:1].upper() + p[1:] for p in parts) or "Op"


def path_to_pascal(path: str) -> str:
    parts: list[str] = []
    for seg in path.strip("/").split("/"):
        if not seg:
            continue
        m = PATH_PARAM_RE.fullmatch(seg)
        if m:
            parts.append("By" + to_pascal(m.group(1)))
            continue
        parts.append(to_pascal(seg))
    return "".join(parts) or "Root"


def needs_rewrite(oid: str | None) -> bool:
    """Rewrite portal kebab/snake ids; keep camelCase / PascalCase as-is."""
    if not oid:
        return True
    if "-" in oid or "." in oid or "_" in oid:
        return True
    return False


def normalize_operation_ids(doc: dict[str, Any]) -> int:
    paths = doc.get("paths") or {}
    used: set[str] = set()
    rewritten = 0
    for path, item in paths.items():
        if not isinstance(item, dict):
            continue
        for method, op in item.items():
            if method.startswith("x-") or not isinstance(op, dict):
                continue
            if method.lower() not in METHOD_VERB and method.lower() not in {
                "get",
                "post",
                "put",
                "patch",
                "delete",
                "head",
                "options",
                "trace",
            }:
                continue
            oid = op.get("operationId")
            if isinstance(oid, str) and not needs_rewrite(oid):
                # Normalize leading lowercase camelCase → PascalCase for Go.
                name = oid[:1].upper() + oid[1:] if oid[:1].islower() else oid
                if name != oid:
                    rewritten += 1
            elif isinstance(oid, str):
                name = to_pascal(oid)
                rewritten += 1
            else:
                verb = METHOD_VERB.get(method.lower(), method.capitalize())
                name = verb + path_to_pascal(path)
                rewritten += 1
            base = name
            n = 2
            while name in used:
                name = f"{base}{n}"
                n += 1
            used.add(name)
            op["operationId"] = name
    return rewritten


def main() -> int:
    ap = argparse.ArgumentParser()
    ap.add_argument("--in", dest="inp", required=True)
    ap.add_argument("--out", required=True)
    ap.add_argument("--server-tem", required=True)
    ap.add_argument("--server-prod", required=True)
    args = ap.parse_args()

    doc = load(Path(args.inp))
    # Prefer explicit manifest servers (already path-qualified when provided).
    doc["servers"] = [
        {"url": args.server_tem.rstrip("/"), "description": "USPS TEM (test)"},
        {"url": args.server_prod.rstrip("/"), "description": "USPS production"},
    ]
    ensure_bearer(doc)
    n_coll = fix_param_schema_name_collisions(doc)
    n_ops = normalize_operation_ids(doc)
    dump(Path(args.out), doc)
    print(
        f"fixed {n_coll} name collision(s), rewrote {n_ops} operationId(s) → {args.out}",
        file=sys.stderr,
    )
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
