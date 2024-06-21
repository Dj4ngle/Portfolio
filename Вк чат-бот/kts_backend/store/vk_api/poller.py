import asyncio
from asyncio import Task

from kts_backend.store import Store


class Poller:
    def __init__(self, store: Store, queue: asyncio.Queue):
        self.queue = queue
        self.store = store
        self.is_running = False
        self.poll_task: Task | None = None

    async def poll(self):
        while self.is_running:
            updates = await self.store.vk_api.poll()
            for u in updates:
                self.queue.put_nowait(u)

    async def start(self):
        self.is_running = True
        self.poll_task = asyncio.create_task(self.poll())

    async def stop(self):
        self.is_running = False
        self.poll_task.cancel()