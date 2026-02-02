from app.wsl_tmux import build_tmux_command


def test_build_tmux_command():
    """Builder should prefix with wsl.exe tmux."""
    cmd = build_tmux_command(["list-sessions"])
    assert cmd[:2] == ["wsl.exe", "tmux"]
