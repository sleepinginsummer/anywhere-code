from dataclasses import dataclass
from typing import Final


TOKEN_PREFIX: Final[str] = "token-"


@dataclass(frozen=True)
class AuthConfig:
    """Auth configuration for the application."""

    username: str
    password: str


def verify_credentials(username: str, password: str, expected_user: str, expected_pass: str) -> bool:
    """Return True when provided credentials match expected values."""
    return username == expected_user and password == expected_pass


def issue_token(username: str) -> str:
    """Issue a simple deterministic token for the given username."""
    # Keep it deterministic to simplify testing and initial auth flow.
    return f"{TOKEN_PREFIX}{username}"


def verify_token(token: str, username: str) -> bool:
    """Validate a token for a given username."""
    return token == issue_token(username)
