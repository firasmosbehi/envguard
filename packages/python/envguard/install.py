"""Download and install the EnvGuard binary for the current platform."""

import os
import platform
import urllib.request
from pathlib import Path

VERSION = "0.1.2"
REPO = "firasmosbehi/envguard"


def _get_platform() -> str:
    system = platform.system().lower()
    machine = platform.machine().lower()

    system_map = {
        "darwin": "darwin",
        "linux": "linux",
        "windows": "windows",
    }

    arch_map = {
        "x86_64": "amd64",
        "amd64": "amd64",
        "arm64": "arm64",
        "aarch64": "arm64",
    }

    sys_name = system_map.get(system)
    arch_name = arch_map.get(machine)

    if not sys_name or not arch_name:
        raise RuntimeError(f"Unsupported platform: {system}/{machine}")

    return f"{sys_name}-{arch_name}"


def _get_binary_dir() -> Path:
    """Return the directory where the binary should be stored."""
    # Store in user's home directory under .envguard
    home = Path.home()
    binary_dir = home / ".envguard" / "bin"
    binary_dir.mkdir(parents=True, exist_ok=True)
    return binary_dir


def _get_binary_path() -> Path:
    binary_dir = _get_binary_dir()
    binary_name = "envguard.exe" if platform.system().lower() == "windows" else "envguard"
    return binary_dir / binary_name


def ensure_binary() -> Path:
    """Ensure the EnvGuard binary is installed. Download if missing."""
    binary_path = _get_binary_path()

    if binary_path.exists():
        return binary_path

    plat = _get_platform()
    suffix = ".exe" if platform.system().lower() == "windows" else ""
    url = f"https://github.com/{REPO}/releases/download/v{VERSION}/envguard-{plat}{suffix}"

    print(f"Downloading EnvGuard {VERSION} for {plat}...")
    print(f"URL: {url}")

    req = urllib.request.Request(
        url,
        headers={"User-Agent": "envguard-python-installer"},
    )

    try:
        with urllib.request.urlopen(req, timeout=60) as response:
            if response.status in (301, 302, 307, 308):
                # Follow redirect manually if needed
                redirect_url = response.headers.get("Location")
                if redirect_url:
                    with urllib.request.urlopen(redirect_url, timeout=60) as redirect_response:
                        binary_path.write_bytes(redirect_response.read())
                else:
                    raise RuntimeError("Redirect without location header")
            else:
                binary_path.write_bytes(response.read())
    except Exception as e:
        raise RuntimeError(
            f"Failed to download EnvGuard binary: {e}\n"
            f"You can manually download it from: "
            f"https://github.com/{REPO}/releases/tag/v{VERSION}"
        ) from e

    # Make executable on Unix
    if platform.system().lower() != "windows":
        os.chmod(binary_path, 0o755)

    print(f"EnvGuard binary installed at {binary_path}")
    return binary_path
