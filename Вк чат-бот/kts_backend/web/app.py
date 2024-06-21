from typing import Sequence, Callable

from aiohttp_apispec import setup_aiohttp_apispec
from aiohttp_cors import CorsConfig
from pyparsing import Optional


from aiohttp.web import (
    Application as AiohttpApplication,
    Request as AiohttpRequest,
    View as AiohttpView,
)
from aiohttp_session import setup as session_setup
from aiohttp_session.cookie_storage import EncryptedCookieStorage


# from kts_backend import __appname__, __version__
from kts_backend.web.config import Config, setup_config
from kts_backend.web.logger import setup_logging
from .middlewares import setup_middlewares
from .urls import register_urls


# __all__ = ("ApiApplication",)

from kts_backend.store import Store, setup_store, Database
from kts_backend.web.urls import register_urls
from ..admin.models import Admin


class Application(AiohttpApplication):
    config: Config | None = None
    store: Store | None = None
    database: Database | None = None


class Request(AiohttpRequest):
    admin: Admin | None = None

    @property
    def app(self) -> Application:
        return super().app()


class View(AiohttpView):
    @property
    def request(self) -> Request:
        return super().request

    @property
    def database(self):
        return self.request.app.database

    @property
    def store(self) -> Store:
        return self.request.app.store

    @property
    def data(self) -> dict:
        return self.request.get("data", {})


app = Application()


def setup_app(config_path: str) -> Application:
    setup_logging(app)
    setup_config(app, config_path)
    session_setup(app, EncryptedCookieStorage(app.config.session.key))
    register_urls(app)
    setup_aiohttp_apispec(
        app, title="Vk Quiz Bot", url="/docs/json", swagger_path="/docs"
    )
    setup_middlewares(app)
    setup_store(app)
    return app
