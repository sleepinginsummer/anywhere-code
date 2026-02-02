from datetime import datetime

from fastapi import APIRouter, Header, HTTPException, status
from pydantic import BaseModel

from app.auth import AuthConfig, issue_token
from app.storage import SessionStore


class CreateSessionRequest(BaseModel):
    """Request payload for creating a session."""

    name: str | None = None


class RenameSessionRequest(BaseModel):
    """Request payload for renaming a session."""

    name: str


def _require_token(authorization: str | None, expected_token: str) -> None:
    """Raise HTTP 401 if the bearer token is missing or invalid."""
    if not authorization:
        raise HTTPException(status_code=status.HTTP_401_UNAUTHORIZED, detail="missing token")
    if not authorization.startswith("Bearer "):
        raise HTTPException(status_code=status.HTTP_401_UNAUTHORIZED, detail="invalid token")

    token = authorization.removeprefix("Bearer ")
    if token != expected_token:
        raise HTTPException(status_code=status.HTTP_401_UNAUTHORIZED, detail="invalid token")


def _default_session_name() -> str:
    """Generate a default session name based on timestamp."""
    return datetime.now().strftime("session-%Y%m%d-%H%M%S")


def create_sessions_router(auth_config: AuthConfig, store: SessionStore, max_sessions: int) -> APIRouter:
    """Create a router for session management endpoints."""
    router = APIRouter(prefix="/api/sessions")
    expected_token = issue_token(auth_config.username)

    @router.get("")
    def list_sessions(authorization: str | None = Header(default=None)) -> list[dict]:
        """List all sessions for the current user."""
        _require_token(authorization, expected_token)
        return store.list_sessions()

    @router.post("")
    def create_session(payload: CreateSessionRequest, authorization: str | None = Header(default=None)) -> dict:
        """Create a new session if the limit allows."""
        _require_token(authorization, expected_token)
        if len(store.list_sessions()) >= max_sessions:
            raise HTTPException(status_code=status.HTTP_409_CONFLICT, detail="session limit reached")

        name = payload.name or _default_session_name()
        return store.create(name=name)

    @router.post("/{session_id}/rename")
    def rename_session(session_id: str, payload: RenameSessionRequest, authorization: str | None = Header(default=None)) -> dict:
        """Rename a session by id."""
        _require_token(authorization, expected_token)
        try:
            return store.rename(session_id, payload.name)
        except KeyError:
            raise HTTPException(status_code=status.HTTP_404_NOT_FOUND, detail="session not found")

    @router.delete("/{session_id}")
    def delete_session(session_id: str, authorization: str | None = Header(default=None)) -> dict:
        """Delete a session by id."""
        _require_token(authorization, expected_token)
        store.delete(session_id)
        return {"status": "deleted"}

    return router
