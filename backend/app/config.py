from pydantic_settings import BaseSettings, SettingsConfigDict


class Settings(BaseSettings):
    """App settings loaded from environment or defaults."""

    max_sessions: int = 10
    auth_username: str
    auth_password: str
    data_dir: str = "data"

    # Use env prefix for all settings keys.
    model_config = SettingsConfigDict(env_prefix="ANYWHERE_")
