"""EnvGuard — validate .env files against a YAML schema."""

from .validator import validate, ValidationResult, ValidationError

__version__ = "0.1.1"
__all__ = ["validate", "ValidationResult", "ValidationError"]
