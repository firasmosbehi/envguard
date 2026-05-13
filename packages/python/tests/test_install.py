"""Tests for envguard install module."""

import unittest
from unittest.mock import patch

from envguard.install import _get_binary_dir, _get_binary_path, _get_platform


class TestInstall(unittest.TestCase):
    """Test install utilities."""

    def test_get_platform_linux_amd64(self):
        with patch("platform.system", return_value="Linux"), patch(
            "platform.machine", return_value="x86_64"
        ):
            self.assertEqual(_get_platform(), "linux-amd64")

    def test_get_platform_darwin_arm64(self):
        with patch("platform.system", return_value="Darwin"), patch(
            "platform.machine", return_value="arm64"
        ):
            self.assertEqual(_get_platform(), "darwin-arm64")

    def test_get_platform_windows_amd64(self):
        with patch("platform.system", return_value="Windows"), patch(
            "platform.machine", return_value="AMD64"
        ):
            self.assertEqual(_get_platform(), "windows-amd64")

    def test_get_platform_linux_aarch64(self):
        with patch("platform.system", return_value="Linux"), patch(
            "platform.machine", return_value="aarch64"
        ):
            self.assertEqual(_get_platform(), "linux-arm64")

    def test_get_platform_unsupported_system(self):
        with patch("platform.system", return_value="FreeBSD"), patch(
            "platform.machine", return_value="x86_64"
        ):
            with self.assertRaises(RuntimeError) as ctx:
                _get_platform()
            self.assertIn("Unsupported platform", str(ctx.exception))

    def test_get_platform_unsupported_arch(self):
        with patch("platform.system", return_value="Linux"), patch(
            "platform.machine", return_value="mips"
        ):
            with self.assertRaises(RuntimeError) as ctx:
                _get_platform()
            self.assertIn("Unsupported platform", str(ctx.exception))

    def test_get_binary_dir_returns_path(self):
        binary_dir = _get_binary_dir()
        self.assertTrue(binary_dir.exists())
        self.assertTrue(binary_dir.is_dir())

    def test_get_binary_path_returns_path(self):
        binary_path = _get_binary_path()
        self.assertIn("envguard", binary_path.name)


if __name__ == "__main__":
    unittest.main()
