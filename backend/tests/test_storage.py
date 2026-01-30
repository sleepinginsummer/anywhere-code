from app.storage import SessionStore


def test_create_and_list_session(tmp_path):
    """Creating a session should persist and list it back."""
    store = SessionStore(tmp_path)
    session = store.create(name="session-1")
    items = store.list_sessions()
    assert session["id"] in [item["id"] for item in items]
