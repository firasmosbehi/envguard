"""EnvGuard validation API."""

import json
import subprocess
from dataclasses import dataclass
from typing import List, Optional, Union

from .install import ensure_binary


@dataclass
class ValidationError:
    key: str
    message: str
    rule: str


@dataclass
class ValidationResult:
    valid: bool
    errors: List[ValidationError]
    warnings: List[ValidationError]


def validate(
    schema_path: Optional[str] = None,
    env_path: Optional[Union[str, List[str]]] = None,
    strict: bool = False,
    env_name: Optional[str] = None,
) -> ValidationResult:
    """Validate a .env file against a schema.

    Args:
        schema_path: Path to the schema YAML file. Defaults to "envguard.yaml".
        env_path: Path to the .env file, or list of paths. Defaults to ".env".
        strict: Fail if .env contains keys not defined in schema.
        env_name: Environment name (e.g. "production", "development") for environment-specific rules.

    Returns:
        ValidationResult with valid flag, errors, and warnings.

    Raises:
        RuntimeError: If EnvGuard binary fails or returns unexpected output.
    """
    binary = ensure_binary()
    args = [str(binary), "validate", "--format", "json"]

    if schema_path:
        args.extend(["--schema", schema_path])
    if env_path:
        if isinstance(env_path, str):
            env_path = [env_path]
        for p in env_path:
            args.extend(["--env", p])
    if strict:
        args.append("--strict")
    if env_name:
        args.extend(["--env-name", env_name])

    result = subprocess.run(
        args,
        capture_output=True,
        text=True,
    )

    # Try to parse JSON from stdout
    stdout = result.stdout.strip()
    if stdout:
        try:
            data = json.loads(stdout)
            return ValidationResult(
                valid=data.get("valid", False),
                errors=[ValidationError(**e) for e in data.get("errors", [])],
                warnings=[ValidationError(**w) for w in data.get("warnings", [])],
            )
        except (json.JSONDecodeError, TypeError):
            pass

    # If we can't parse JSON, raise an error
    stderr = result.stderr.strip()
    raise RuntimeError(
        f"EnvGuard failed (exit code {result.returncode}): "
        f"{stderr or stdout or 'Unknown error'}"
    )
