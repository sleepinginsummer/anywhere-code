from app.storage import SessionStore


def test_create_and_list_session(tmp_path):
    """Creating a session should persist and list it back."""
    store = SessionStore(tmp_path)
    session = store.create(name="session-1")
    items = store.list_sessions()
    assert session["id"] in [item["id"] for item in items]


def test_rename_and_delete_session(tmp_path):
    """Rename and delete should update persisted records."""
    store = SessionStore(tmp_path)
    session = store.create(name="session-1")

    updated = store.rename(session["id"], "session-2")
    assert updated["name"] == "session-2"

    store.delete(session["id"])
    items = store.list_sessions()
    assert session["id"] not in [item["id"] for item in items]
