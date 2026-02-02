from app.terminal_ws import normalize_input


def test_normalize_input():
    """normalize_input should preserve newline input."""
    assert normalize_input("ls\n") == "ls\n"
