import random, typing, asyncio

from aiohttp import TCPConnector
from aiohttp.client import ClientSession

from kts_backend.base.base_accessor import BaseAccessor
from kts_backend.store.vk_api.dataclasses import (
    Message,
    Update,
    UpdateObject,
    User,
)
from kts_backend.store.vk_api.poller import Poller
from kts_backend.store.vk_api.worker import Worker

if typing.TYPE_CHECKING:
    from kts_backend.web.app import Application

API_PATH = "https://api.vk.com/method/"


class VkApiAccessor(BaseAccessor):
    def __init__(self, app: "Application", *args, **kwargs):
        super().__init__(app, *args, **kwargs)
        self.session: ClientSession | None = None
        self.key: str | None = None
        self.server: str | None = None
        self.poller: Poller | None = None
        self.worker: Worker | None = None
        self.ts: int | None = None
        self.queue = asyncio.Queue()

    async def connect(self, app: "Application"):
        self.session = ClientSession(connector=TCPConnector(verify_ssl=False))
        try:
            await self._get_long_poll_service()
        except Exception as e:
            self.logger.error("Exception", exc_info=e)
        self.worker = Worker(app.store, self.queue)
        self.logger.info("start working")
        self.poller = Poller(app.store, self.queue)
        self.logger.info("start polling")

        await self.poller.start()
        await self.worker.start()

    async def disconnect(self, app: "Application"):
        if self.session:
            await self.session.close()
        if self.poller:
            await self.poller.stop()
            await self.worker.stop()

    @staticmethod
    def _build_query(host: str, method: str, params: dict) -> str:
        url = host + method + "?"
        if "v" not in params:
            params["v"] = "5.131"
        url += "&".join([f"{k}={v}" for k, v in params.items()])
        return url

    async def _get_long_poll_service(self):
        async with self.session.get(
            self._build_query(
                host=API_PATH,
                method="groups.getLongPollServer",
                params={
                    "group_id": self.app.config.bot.group_id,
                    "access_token": self.app.config.bot.token,
                },
            )
        ) as resp:
            data = (await resp.json())["response"]
            self.logger.info(data)
            self.key = data["key"]
            self.server = data["server"]
            self.ts = data["ts"]
            self.logger.info(self.server)

    async def poll(self):
        async with self.session.get(
            self._build_query(
                host=self.server,
                method="",
                params={
                    "act": "a_check",
                    "key": self.key,
                    "wait": 10,
                    "mode": 2,
                    "ts": self.ts,
                },
            )
        ) as resp:
            data = await resp.json()
            self.logger.info(data)
            self.ts = data["ts"]
            raw_updates = data.get("updates", [])
            updates = []
            for update in raw_updates:
                if update["type"] == "message_new":
                    updates.append(
                        Update(
                            type=update["type"],
                            object=UpdateObject(
                                id=update["object"]["message"]["id"],
                                user_id=update["object"]["message"]["from_id"],
                                peer_id=update["object"]["message"]["peer_id"],
                                body=update["object"]["message"],
                            ),
                        )
                    )
            return updates

    async def send_keyboard_message(self, message: Message, keyboard) -> None:
        async with self.session.get(
            self._build_query(
                API_PATH,
                "messages.send",
                params={
                    "random_id": random.randint(1, 2**32),
                    "peer_id": message.peer_id,
                    "message": message.text,
                    "access_token": self.app.config.bot.token,
                    "keyboard": keyboard,
                },
            )
        ) as resp:
            data = await resp.json()
            self.logger.info(data)

    async def send_message(self, message: Message) -> None:
        async with self.session.get(
            self._build_query(
                API_PATH,
                "messages.send",
                params={
                    "random_id": random.randint(1, 2**32),
                    "peer_id": message.peer_id,
                    "message": message.text,
                    "access_token": self.app.config.bot.token,
                },
            )
        ) as resp:
            data = await resp.json()
            self.logger.info(data)

    async def get_all_users_from_chat(self, peer_id) -> None:

        async with self.session.get(
            self._build_query(
                API_PATH,
                "messages.getConversationMembers",
                params={
                    "peer_id": peer_id,
                    "access_token": self.app.config.bot.token,
                    "fields": "first_name",
                },
            )
        ) as resp:
            users = await resp.json()
        self.logger.info(users)
        list_users = []
        for user in users["response"]["profiles"]:

            list_users.append(
                User(
                    id=user["id"],
                    first_name=user["first_name"],
                    last_name=user["last_name"],
                )
            )
        return list_users
