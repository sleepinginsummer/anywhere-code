from typing import Iterable


def build_tmux_command(args: Iterable[str]) -> list[str]:
    """Build a tmux command prefixed with wsl.exe for Windows execution."""
    # Prefix with wsl.exe to run tmux inside WSL.
    return ["wsl.exe", "tmux", *args]
