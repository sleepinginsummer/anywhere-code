import json
import uuid
from datetime import datetime, timezone
from pathlib import Path
from typing import Any


class SessionStore:
    """Persist and retrieve session metadata in a JSON file."""

    def __init__(self, data_dir: Path) -> None:
        """Initialize storage with a data directory path."""
        self._data_dir = Path(data_dir)
        self._data_dir.mkdir(parents=True, exist_ok=True)
        self._file_path = self._data_dir / "sessions.json"

    def create(self, name: str) -> dict[str, Any]:
        """Create a new session record and persist it."""
        sessions = self._load()
        now = datetime.now(timezone.utc).isoformat()
        session = {
            "id": uuid.uuid4().hex,
            "name": name,
            "createdAt": now,
            "lastInputSummary": "",
            "lastActiveAt": now,
        }
        sessions.append(session)
        self._save(sessions)
        return session

    def list_sessions(self) -> list[dict[str, Any]]:
        """Return all persisted session records."""
        return self._load()

    def _load(self) -> list[dict[str, Any]]:
        """Load session records from disk or return an empty list."""
        if not self._file_path.exists():
            # No data yet; start with an empty list.
            return []
        return json.loads(self._file_path.read_text(encoding="utf-8"))

    def _save(self, sessions: list[dict[str, Any]]) -> None:
        """Write session records to disk atomically."""
        # Write to a temp file first to avoid partial writes.
        tmp_path = self._file_path.with_suffix(".tmp")
        tmp_path.write_text(json.dumps(sessions, ensure_ascii=False, indent=2), encoding="utf-8")
        tmp_path.replace(self._file_path)
