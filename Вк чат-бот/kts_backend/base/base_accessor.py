import typing
from logging import getLogger

from sqlalchemy.orm import sessionmaker
from sqlalchemy.ext.asyncio import AsyncSession, create_async_engine

if typing.TYPE_CHECKING:
    from kts_backend.web.app import Application


class BaseAccessor:
    def __init__(self, app: "Application", *args, **kwargs):
        self.app = app
        self.logger = getLogger("accessor")
        app.on_startup.append(self.connect)
        app.on_cleanup.append(self.disconnect)
        self.session: typing.Optional[AsyncSession] = None

    async def connect(self, app: "Application"):
        db_url = (
            f"postgresql+asyncpg://{self.app.config.database.user}:"
            f"{self.app.config.database.password}@{self.app.config.database.host}/"
            f"{self.app.config.database.database}"
        )
        self.session = sessionmaker(
            create_async_engine(db_url, echo=True),
            expire_on_commit=False,
            class_=AsyncSession,
        )
        return self.session

    async def disconnect(self, app: "Application"):
        return
