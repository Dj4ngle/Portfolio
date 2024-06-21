import asyncio, typing
from asyncio import Task

from kts_backend.store import Store
from kts_backend.store.vk_api.dataclasses import UpdateObject


class Worker:
    def __init__(self, store: Store, queue: asyncio.Queue):
        self.store = store
        self.queue = queue
        self.is_running = False
        self._task: Task | None = None

    async def worker(self):
        while self.is_running:
            u: UpdateObject = await self.queue.get()
            await self.store.bots_manager.handle_updates(u)
            self.queue.task_done()

    async def start(self):
        print("worker started")
        self.is_running = True
        self._task = asyncio.create_task((self.worker()))

    async def stop(self):
        print("worker stoped")
        self.is_running = False
        self._task.cancel()
