#!/usr/bin/env python3
"""Keep generator-required OpenAPI structure without publication prose.

The source document's key order is preserved so a regenerated vendored spec
shows the redactions clearly in review.
"""

from __future__ import annotations

import argparse
import json
from pathlib import Path

PROSE_KEYS = {"description", "summary", "example", "examples", "externalDocs", "tags"}
INFO_KEYS = {"contact", "termsOfService", "license"}


def redact(value: object, parent_key: str = "") -> None:
    if isinstance(value, dict):
        for key in tuple(value):
            # Under a schema's "properties", these are field names, not
            # OpenAPI prose keywords.
            if parent_key != "properties" and key in PROSE_KEYS:
                del value[key]
                continue
            if key == "info" and isinstance(value[key], dict):
                for info_key in INFO_KEYS:
                    value[key].pop(info_key, None)
            redact(value[key], key)
    elif isinstance(value, list):
        for item in value:
            redact(item, parent_key)


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("--input", required=True, type=Path)
    parser.add_argument("--output", required=True, type=Path)
    args = parser.parse_args()
    with args.input.open(encoding="utf-8") as source:
        document = json.load(source)
    if not all(key in document for key in ("openapi", "paths", "components")):
        parser.error("source is missing OpenAPI structure")
    redact(document)
    args.output.write_text(
        json.dumps(document, indent=2, ensure_ascii=False, allow_nan=False) + "\n",
        encoding="utf-8",
    )


if __name__ == "__main__":
    main()
