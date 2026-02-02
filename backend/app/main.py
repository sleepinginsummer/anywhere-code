from dataclasses import dataclass

from fastapi import FastAPI, HTTPException
from pydantic import BaseModel

from app.auth import issue_token, verify_credentials
from app.config import Settings


@dataclass(frozen=True)
class AuthConfig:
    """Auth configuration for the application."""

    username: str
    password: str


class LoginRequest(BaseModel):
    """Login request payload."""

    username: str
    password: str


def _load_auth_config(auth_username: str | None, auth_password: str | None) -> AuthConfig:
    """Load auth config from overrides or environment settings."""
    if auth_username is not None and auth_password is not None:
        return AuthConfig(username=auth_username, password=auth_password)

    settings = Settings()
    return AuthConfig(username=settings.auth_username, password=settings.auth_password)


def create_app(auth_username: str | None = None, auth_password: str | None = None) -> FastAPI:
    """Create and configure the FastAPI application."""
    app = FastAPI(title="Anywhere Code")
    auth_config = _load_auth_config(auth_username, auth_password)

    @app.get("/health")
    def health() -> dict:
        """Health check endpoint used for basic liveness testing."""
        return {"status": "ok"}

    @app.post("/api/login")
    def login(payload: LoginRequest) -> dict:
        """Validate credentials and issue a token."""
        if not verify_credentials(payload.username, payload.password, auth_config.username, auth_config.password):
            # Invalid credentials; reject the login.
            raise HTTPException(status_code=401, detail="invalid credentials")
        return {"token": issue_token(payload.username)}

    return app


app = create_app()
