from app.config import Settings


def test_load_settings_defaults():
    """Settings should provide default values for optional fields."""
    settings = Settings(auth_username="u", auth_password="p")
    assert settings.max_sessions == 10
    assert settings.data_dir == "data"
