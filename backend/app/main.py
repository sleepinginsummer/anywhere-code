from pathlib import Path

from fastapi import FastAPI, HTTPException, WebSocket
from pydantic import BaseModel

from app.auth import AuthConfig, issue_token, verify_credentials
from app.config import Settings
from app.routes_sessions import create_sessions_router
from app.storage import SessionStore
from app.terminal_ws import handle_terminal_ws


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


def _load_data_dir(data_dir: Path | str | None, settings: Settings | None) -> Path:
    """Load data dir from override or settings."""
    if data_dir is not None:
        return Path(data_dir)
    if settings is None:
        settings = Settings()
    return Path(settings.data_dir)


def _load_max_sessions(max_sessions: int | None, settings: Settings | None, has_auth_override: bool) -> int:
    """Load max sessions from override or settings."""
    if max_sessions is not None:
        return max_sessions
    if has_auth_override and settings is None:
        # Use the documented default when auth is overridden in tests.
        return 10
    if settings is None:
        settings = Settings()
    return settings.max_sessions


def create_app(
    auth_username: str | None = None,
    auth_password: str | None = None,
    data_dir: Path | str | None = None,
    max_sessions: int | None = None,
) -> FastAPI:
    """Create and configure the FastAPI application."""
    app = FastAPI(title="Anywhere Code")
    has_auth_override = auth_username is not None and auth_password is not None
    settings = Settings() if not has_auth_override else None
    auth_config = _load_auth_config(auth_username, auth_password)
    store = SessionStore(_load_data_dir(data_dir, settings))
    max_sessions_value = _load_max_sessions(max_sessions, settings, has_auth_override)

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

    @app.websocket("/ws/terminal/{session_id}")
    async def terminal_ws(websocket: WebSocket, session_id: str) -> None:
        """Websocket endpoint for tmux session streaming."""
        await handle_terminal_ws(websocket, session_id)

    app.include_router(create_sessions_router(auth_config, store, max_sessions=max_sessions_value))

    return app


app = create_app()
