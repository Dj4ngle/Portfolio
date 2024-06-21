from kts_backend.base.base_accessor import BaseAccessor


class UserAccessor(BaseAccessor):
    async def create_user(self, name: str):
        raise NotImplemented
