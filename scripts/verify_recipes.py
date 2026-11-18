#!/usr/bin/env python3
"""
Memvalidasi seluruh manifes aplikasi di recipes/**/*.yaml terhadap JSON Schema v7.0
serta memeriksa keunikan ID aplikasi.
"""

import json
import sys
from pathlib import Path

import yaml
from jsonschema import ValidationError, validate

ROOT_DIR = Path(__file__).resolve().parent.parent
SCHEMA_PATH = ROOT_DIR / "recipes" / "schema" / "recipe.schema.json"
RECIPES_DIR = ROOT_DIR / "recipes"


def main() -> None:
    schema = json.loads(SCHEMA_PATH.read_text(encoding="utf-8"))
    errors = 0
    seen_ids: set[str] = set()

    for yaml_file in sorted(RECIPES_DIR.rglob("*.yaml")):
        data = yaml.safe_load(yaml_file.read_text(encoding="utf-8"))
        try:
            validate(instance=data, schema=schema)
            app_id = data["id"]
            if app_id in seen_ids:
                raise ValidationError(f"Duplicate recipe id '{app_id}'")
            seen_ids.add(app_id)
            print(f"[VALID] {yaml_file.relative_to(ROOT_DIR)}")
        except ValidationError as exc:
            print(
                f"[ERROR] {yaml_file.relative_to(ROOT_DIR)}: {exc.message}",
                file=sys.stderr,
            )
            errors += 1

    if errors > 0:
        sys.exit(1)
    print(f"Seluruh {len(seen_ids)} resep aplikasi lolos validasi skema v7.0.")


if __name__ == "__main__":
    main()
