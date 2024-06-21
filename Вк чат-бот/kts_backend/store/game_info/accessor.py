from sqlalchemy import insert, select, update
from sqlalchemy.exc import IntegrityError

from kts_backend.base.base_accessor import BaseAccessor
from kts_backend.game_info.models import (
    Question,
    QuestionModel,
    Sessions,
    SessionsModel,
)


class QuizAccessor(BaseAccessor):
    async def create_question(self, title: str, answer: str) -> Question:
        async with self.app.database.session() as session:
            try:
                res = await session.execute(
                    insert(QuestionModel)
                    .values(title=title, answer=answer)
                    .returning(QuestionModel.id)
                )
                await session.commit()
            except IntegrityError:
                return None
            question_id = res.scalars().first()
        return Question(id=question_id, title=title, answer=answer)

    async def list_questions(self) -> list[Question]:
        async with self.app.database.session() as session:
            res = await session.execute(select(QuestionModel))
        return res.scalars().all()

    async def edite_question(
        self, _id: int, title: str | None = None, answer: str | None = None
    ) -> Question:

        update_values = {}
        if title is not None:
            update_values["title"] = title
        if answer is not None:
            update_values["answer"] = answer

        if update_values == {}:
            return None

        async with self.app.database.session() as session:

            res = await session.execute(
                update(QuestionModel)
                .where(QuestionModel.id == _id)
                .values(**update_values)
                .returning(
                    QuestionModel.id, QuestionModel.title, QuestionModel.answer
                )
            )
            await session.commit()

        question = res.fetchall()
        if question == []:
            return None

        return Question(
            id=question[0].id,
            title=question[0].title,
            answer=question[0].answer,
        )

    async def list_session(
        self,
        _id: int | None = None,
        group_id: int | None = None,
        status: str | None = None,
        capitan_id: int | None = None,
    ) -> Sessions:
        update_values = {}
        if _id is not None:
            update_values["_id"] = _id
        if group_id is not None:
            update_values["group_id"] = group_id
        if status is not None:
            update_values["status"] = status
        if capitan_id is not None:
            update_values["capitan_id"] = capitan_id

        async with self.app.database.session() as session:
            res = await session.execute(
                select(SessionsModel).where(**update_values)
            )
        return res.scalars().all()

    async def views_parametrs(self, params: list, data):
        update_values = {}
        for par in params:
            if par in data and data[par] is not None:
                update_values[par] = data[par]
        return update_values
