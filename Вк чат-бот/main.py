import pathlib
from pathlib import Path

from kts_backend.web.app import setup_app
from aiohttp.web import run_app

if __name__ == "__main__":
    run_app(setup_app(config_path=pathlib.Path("local").resolve() / "config.yml"))
