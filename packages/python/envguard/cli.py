#!/usr/bin/env python3
"""CLI entrypoint for the Python EnvGuard wrapper."""

import argparse
import json
import sys

from .validator import validate


def main() -> None:
    """Run the envguard-py CLI."""
    parser = argparse.ArgumentParser(
        description="Validate .env files against a YAML schema",
        prog="envguard-py",
    )
    parser.add_argument(
        "--schema",
        "-s",
        default="envguard.yaml",
        help="Path to schema YAML file (default: envguard.yaml)",
    )
    parser.add_argument(
        "--env",
        "-e",
        default=".env",
        help="Path to .env file (default: .env)",
    )
    parser.add_argument(
        "--format",
        "-f",
        choices=["text", "json"],
        default="text",
        help="Output format",
    )
    parser.add_argument(
        "--strict",
        action="store_true",
        help="Fail if .env contains keys not defined in schema",
    )

    args = parser.parse_args()

    try:
        result = validate(
            schema_path=args.schema,
            env_path=args.env,
            strict=args.strict,
        )
    except RuntimeError as e:
        print(f"Error: {e}", file=sys.stderr)
        sys.exit(2)

    if args.format == "json":
        output = {
            "valid": result.valid,
            "errors": [{"key": e.key, "message": e.message, "rule": e.rule} for e in result.errors],
            "warnings": [
                {"key": w.key, "message": w.message, "rule": w.rule} for w in result.warnings
            ],
        }
        print(json.dumps(output, indent=2))
    else:
        if result.valid and not result.warnings:
            print("✓ All environment variables validated.")
        else:
            if not result.valid:
                print(f"✗ Environment validation failed ({len(result.errors)} error(s))\n")
                for err in result.errors:
                    print(f"  • {err.key}")
                    print(f"    └─ {err.rule}: {err.message}")
            if result.warnings:
                if not result.valid:
                    print()
                print(f"⚠ Warnings ({len(result.warnings)}):\n")
                for warn in result.warnings:
                    print(f"  • {warn.key}")
                    print(f"    └─ {warn.rule}: {warn.message}")

    sys.exit(0 if result.valid else 1)


if __name__ == "__main__":
    main()
