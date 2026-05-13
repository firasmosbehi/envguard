"""Tests for envguard validator module."""

import unittest

from envguard.validator import ValidationError, ValidationResult


class TestValidatorTypes(unittest.TestCase):
    """Test validator dataclasses."""

    def test_validation_error_creation(self):
        err = ValidationError(key="FOO", message="missing", rule="required")
        self.assertEqual(err.key, "FOO")
        self.assertEqual(err.message, "missing")
        self.assertEqual(err.rule, "required")

    def test_validation_result_creation(self):
        result = ValidationResult(
            valid=True,
            errors=[ValidationError(key="A", message="err", rule="required")],
            warnings=[],
        )
        self.assertTrue(result.valid)
        self.assertEqual(len(result.errors), 1)
        self.assertEqual(len(result.warnings), 0)

    def test_validation_result_invalid(self):
        result = ValidationResult(
            valid=False,
            errors=[
                ValidationError(key="A", message="err1", rule="required"),
                ValidationError(key="B", message="err2", rule="type"),
            ],
            warnings=[
                ValidationError(key="C", message="warn1", rule="strict"),
            ],
        )
        self.assertFalse(result.valid)
        self.assertEqual(len(result.errors), 2)
        self.assertEqual(len(result.warnings), 1)


if __name__ == "__main__":
    unittest.main()
