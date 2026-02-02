import importlib
import os

from fastapi.testclient import TestClient

from app.auth import issue_token


def _create_app(tmp_path):
    os.environ["ANYWHERE_AUTH_USERNAME"] = "u"
    os.environ["ANYWHERE_AUTH_PASSWORD"] = "p"
    os.environ["ANYWHERE_DATA_DIR"] = str(tmp_path)

    import app.main as main_module

    importlib.reload(main_module)
    return main_module.create_app()


def test_list_sessions_requires_auth(tmp_path):
    """Sessions endpoint should reject missing auth header."""
    app = _create_app(tmp_path)
    client = TestClient(app)

    res = client.get("/api/sessions")
    assert res.status_code == 401


def test_list_sessions_with_auth(tmp_path):
    """Sessions endpoint should return data with valid auth header."""
    app = _create_app(tmp_path)
    client = TestClient(app)

    token = issue_token("u")
    res = client.get("/api/sessions", headers={"Authorization": f"Bearer {token}"})
    assert res.status_code == 200
    assert res.json() == []
