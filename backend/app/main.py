from fastapi import FastAPI


def create_app() -> FastAPI:
    """Create and configure the FastAPI application."""
    app = FastAPI(title="Anywhere Code")

    @app.get("/health")
    def health() -> dict:
        """Health check endpoint used for basic liveness testing."""
        return {"status": "ok"}

    return app


app = create_app()
