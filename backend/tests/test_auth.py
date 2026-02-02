import importlib
import os

from fastapi.testclient import TestClient

from app.auth import issue_token, verify_credentials


def test_verify_credentials():
    """verify_credentials should return True for matching user/pass."""
    assert verify_credentials("u", "p", "u", "p") is True
    assert verify_credentials("u", "p", "u", "x") is False


def test_login_endpoint():
    """Login should return token for valid credentials and 401 otherwise."""
    os.environ["ANYWHERE_AUTH_USERNAME"] = "u"
    os.environ["ANYWHERE_AUTH_PASSWORD"] = "p"

    import app.main as main_module

    importlib.reload(main_module)
    app = main_module.create_app()

    client = TestClient(app)
    res_ok = client.post("/api/login", json={"username": "u", "password": "p"})
    assert res_ok.status_code == 200
    assert res_ok.json() == {"token": issue_token("u")}

    res_fail = client.post("/api/login", json={"username": "u", "password": "x"})
    assert res_fail.status_code == 401
