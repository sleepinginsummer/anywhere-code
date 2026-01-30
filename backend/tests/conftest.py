import sys
from pathlib import Path

# Ensure backend/app is importable for tests.
ROOT = Path(__file__).resolve().parents[1]
APP_DIR = ROOT / "app"
if str(ROOT) not in sys.path:
    sys.path.insert(0, str(ROOT))
