from hashlib import sha256

from aiohttp.web import HTTPForbidden, HTTPUnauthorized
from aiohttp_apispec import request_schema, response_schema
from aiohttp_session import new_session

from kts_backend.admin.schemes import AdminSchema
from kts_backend.web.app import View
from kts_backend.web.utils import json_response


class AdminLoginView(View):
    @request_schema(AdminSchema)
    @response_schema(AdminSchema, 200)
    async def post(self):

        admin = await self.store.admins.get_by_email(self.data["email"])
        if admin:
            if (
                sha256(self.data["password"].encode()).hexdigest()
                == admin.password
            ):
                raw_admin = AdminSchema().dump(admin)
                session = await new_session(self.request)
                session["admin"] = raw_admin
                return json_response(raw_admin)
        raise HTTPForbidden()


class AdminCurrentView(View):
    @response_schema(AdminSchema, 200)
    async def get(self):
        if not self.request.cookies.get("AIOHTTP_SESSION"):
            raise HTTPUnauthorized
        else:
            return json_response({"id": 1, "email": "admin@admin.com"})
