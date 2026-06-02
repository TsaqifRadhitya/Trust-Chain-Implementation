# Entry point untuk uvicorn (CMD: uvicorn main:app)
# Re-export `app` dari app package.
from app.main import app  # noqa: F401