"""EnvGuard — validate .env files against a YAML schema."""

from .validator import ValidationError, ValidationResult, validate

__version__ = "2.0.0"
__all__ = ["validate", "ValidationResult", "ValidationError"]
