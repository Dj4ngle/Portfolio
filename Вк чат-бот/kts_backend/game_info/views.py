from aiohttp.web_exceptions import HTTPBadRequest, HTTPConflict, HTTPNotFound
from aiohttp_apispec import querystring_schema, request_schema, response_schema

from kts_backend.game_info.schemes import (
    ListQuestionSchema,
    QuestionSchema,
    QuestionEditeSchema,
    SessionSchema,
    ListQSessionSchema,
)
from kts_backend.web.app import View
from kts_backend.web.mixins import AuthRequiredMixin
from kts_backend.web.utils import json_response


class QuestionAddView(AuthRequiredMixin, View):
    @request_schema(QuestionSchema)
    @response_schema(QuestionSchema)
    async def post(self):
        question = await self.store.game_info.create_question(
            self.data["title"], self.data["answer"]
        )
        if not question:
            raise HTTPConflict
        return json_response(data=QuestionSchema().dump(question))


class QuestionListView(AuthRequiredMixin, View):
    @response_schema(ListQuestionSchema)
    async def get(self):
        questions = await self.store.game_info.list_questions()
        return json_response(
            data=ListQuestionSchema().dump({"questions": questions})
        )


class QuestionEditeView(AuthRequiredMixin, View):
    @request_schema(QuestionEditeSchema)
    @response_schema(QuestionSchema)
    async def post(self):
        _id = self.data["id"]

        keys = ["title", "answer"]
        params = await self.store.game_info.views_parametrs(keys, self.data)

        question = await self.store.game_info.edite_question(_id, **params)

        if not question:
            raise HTTPBadRequest
        return json_response(data=QuestionSchema().dump(question))


class SessionsView(AuthRequiredMixin, View):
    async def get(self):

        keys = ["id", "group_id", "status", "capitan_id"]
        params = await self.store.game_info.views_parametrs(keys, self.data)

        sessions = await self.store.game_info.list_session(**params)

        return json_response(
            data=ListQSessionSchema().dump({"sessions": sessions})
        )
