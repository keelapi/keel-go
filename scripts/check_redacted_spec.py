#!/usr/bin/env python3
"""Fail if the vendored contract contains publication prose or is empty."""

from __future__ import annotations

import json
import sys
from pathlib import Path

PROSE_KEYS = {"description", "summary", "example", "examples", "externalDocs", "tags"}
INFO_KEYS = {"contact", "termsOfService", "license"}


def main() -> int:
    spec = json.loads(Path(sys.argv[1]).read_text(encoding="utf-8"))
    if not spec.get("paths") or not spec.get("components", {}).get("schemas"):
        print("vendored spec has no paths or schemas", file=sys.stderr)
        return 1
    violations: list[str] = []

    def walk(value: object, path: str, parent_key: str = "") -> None:
        if isinstance(value, dict):
            for key, child in value.items():
                if (parent_key != "properties" and key in PROSE_KEYS) or (
                    path == "info" and key in INFO_KEYS
                ):
                    violations.append(f"{path}.{key}")
                walk(child, f"{path}.{key}" if path else key, key)
        elif isinstance(value, list):
            for index, child in enumerate(value):
                walk(child, f"{path}[{index}]", parent_key)

    walk(spec, "")
    if violations:
        print("vendored spec contains publication prose: " + ", ".join(violations[:10]), file=sys.stderr)
        return 1
    print(f"redacted spec: {len(spec['paths'])} paths, {len(spec['components']['schemas'])} schemas")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
