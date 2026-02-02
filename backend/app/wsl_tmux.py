from typing import Iterable


def build_tmux_command(args: Iterable[str]) -> list[str]:
    """Build a tmux command prefixed with wsl.exe for Windows execution."""
    # Prefix with wsl.exe to run tmux inside WSL.
    return ["wsl.exe", "tmux", *args]


def build_tmux_attach_command(session_id: str) -> list[str]:
    """Build command to attach to a tmux session."""
    return build_tmux_command(["attach", "-t", session_id])


def build_tmux_new_command(session_id: str) -> list[str]:
    """Build command to create a detached tmux session."""
    return build_tmux_command(["new-session", "-d", "-s", session_id])


def build_tmux_list_command() -> list[str]:
    """Build command to list tmux sessions."""
    return build_tmux_command(["list-sessions"])
