import typing, random, time
from logging import getLogger

from sqlalchemy import select, insert, desc, func
from sqlalchemy import update as alchemy_update

from kts_backend.game_info.models import (
    QuestionModel,
    SessionsModel,
    GamesHistoryModel,
    Sessions,
    TeamPlayerModel,
    GamesHistory,
)
from kts_backend.store.vk_api.dataclasses import Message, Update

if typing.TYPE_CHECKING:
    from kts_backend.web.app import Application

from vk_api.keyboard import VkKeyboard, VkKeyboardColor


class BotManager:
    def __init__(self, app: "Application"):
        self.app = app
        self.bot = None
        self.logger = getLogger("handler")
        self.question_limit = 13
        self.start_option = "/start"
        self.stop_option = "/stop"

    async def handle_updates(self, update):

        if not update:
            return
        else:
            # Команда для запуска бота
            if update.object.body["text"] == self.start_option:

                # Выбор капитана команды
                await self.app.store.game_manager.choose_the_capitan(update)

            # Досрочное окончание игры
            elif update.object.body["text"] == self.stop_option:

                bot_points = 0
                users_points = 0

                async with self.app.database.session() as session:
                    res = await session.execute(
                        select(SessionsModel).where(
                            SessionsModel.group_id == update.object.peer_id,
                            SessionsModel.status != "Closed",
                        )
                    )
                    cur_session = res.scalars().first()
                    if not cur_session:
                        pass
                    else:
                        res = await session.execute(
                            select(GamesHistoryModel).where(
                                GamesHistoryModel.session_id == cur_session.id
                            )
                        )
                        all_rounds = res.scalars().all()

                        for round in all_rounds:
                            if round.is_answer_right:
                                users_points += 1
                            else:
                                bot_points += 1

                        if bot_points == 0:
                            bot_points += 1

                        await self.app.store.game_manager.game_end(
                            bot_points - 1, users_points, update, cur_session.id
                        )

            else:

                async with self.app.database.session() as session:
                    res = await session.execute(
                        select(SessionsModel).where(
                            SessionsModel.group_id == update.object.peer_id,
                            SessionsModel.status != "Closed",
                        )
                    )
                    cur_session = res.scalars().first()
                    if cur_session == None:
                        cur_session = Sessions(
                            id=None, group_id=None, status=None, capitan_id=None
                        )

                    res = await session.execute(
                        select(GamesHistoryModel)
                        .where(GamesHistoryModel.session_id == cur_session.id)
                        .order_by(desc(GamesHistoryModel.id))
                    )
                    cur_round_info = res.scalars().first()
                    if cur_round_info == None:
                        cur_round_info = GamesHistory(
                            id=None,
                            session_id=None,
                            guestion_id=None,
                            player_id=None,
                            is_answer_right=False,
                        )

                # Проверка, что собщение является командой боту
                if update.object.body["text"].startswith(
                    f"[club{self.app.config.bot.group_id}|@"
                ):

                    index = update.object.body["text"].index(" ")
                    command = update.object.body["text"][index + 1 :]

                    if update.object.user_id == cur_session.capitan_id:
                        if command == "Старт игры!!!":

                            async with self.app.database.session() as session:

                                res = await session.execute(
                                    select(QuestionModel).order_by(
                                        func.random()
                                    )
                                )
                                new_question = res.scalars().first()

                                await session.execute(
                                    insert(GamesHistoryModel).values(
                                        session_id=cur_session.id,
                                        guestion_id=new_question.id,
                                        is_answer_right=False,
                                    )
                                )

                                await session.execute(
                                    alchemy_update(SessionsModel)
                                    .where(
                                        SessionsModel.group_id
                                        == update.object.peer_id,
                                        SessionsModel.status != "Closed",
                                    )
                                    .values(status="Game started")
                                )

                                await session.commit()

                            # Старт вопросов после решения капитана о начале игры
                            await self.app.store.game_manager.game_round(
                                update, new_question.title, cur_session.id
                            )

                        elif cur_session.status == "Game started":
                            # Информирование всех, кто выбран отвечающим для данного вопроса
                            async with self.app.database.session() as session:
                                res = await session.execute(
                                    select(TeamPlayerModel).where(
                                        TeamPlayerModel.first_name
                                        == command.split(" ")[0],
                                        TeamPlayerModel.last_name
                                        == command.split(" ")[1],
                                    )
                                )
                                answerer = res.scalars().first()

                                await session.execute(
                                    alchemy_update(GamesHistoryModel)
                                    .where(
                                        GamesHistoryModel.session_id
                                        == cur_session.id
                                    )
                                    .values(player_id=answerer.player_id)
                                )

                                await session.commit()

                            message_to_user = f"@id{answerer.player_id} ({command})  выбран отвечающим"

                            await self.app.store.vk_api.send_message(
                                Message(
                                    peer_id=update.object.peer_id,
                                    text=message_to_user,
                                ),
                            )

                elif (
                    cur_session.status == "Game started"
                    and cur_round_info.player_id == update.object.user_id
                ):

                    async with self.app.database.session() as session:
                        res = await session.execute(
                            select(QuestionModel).where(
                                QuestionModel.id == cur_round_info.guestion_id
                            )
                        )
                        cur_question = res.scalars().first()

                        # Проверка правильности ответа на вопрос
                        if (
                            update.object.body["text"].lower()
                            == cur_question.answer
                        ):
                            await self.app.store.vk_api.send_message(
                                Message(
                                    peer_id=update.object.peer_id,
                                    text="Ответ правильный!!!!!!",
                                )
                            )

                            async with self.app.database.session() as session:
                                await session.execute(
                                    alchemy_update(GamesHistoryModel)
                                    .where(
                                        GamesHistoryModel.id
                                        == cur_round_info.id
                                    )
                                    .values(is_answer_right=True)
                                )
                                await session.commit()

                        else:
                            await self.app.store.vk_api.send_message(
                                Message(
                                    peer_id=update.object.peer_id,
                                    text=f"Ответ неправильный!!! Правильный ответ: {cur_question.answer}",
                                )
                            )

                        bot_points = 0
                        users_points = 0

                        async with self.app.database.session() as session:
                            res = await session.execute(
                                select(GamesHistoryModel).where(
                                    GamesHistoryModel.session_id
                                    == cur_session.id
                                )
                            )
                            all_rounds = res.scalars().all()

                            all_questions_list = []
                        for round in all_rounds:
                            if round.is_answer_right:
                                users_points += 1
                            else:
                                bot_points += 1
                            all_questions_list.append(round.guestion_id)

                        await self.app.store.vk_api.send_message(
                            Message(
                                peer_id=update.object.peer_id,
                                text=f"Текущий счёт: Бот: {bot_points} Игроки: {users_points}",
                            )
                        )

                        if len(all_rounds) == self.question_limit:
                            # Конец игры
                            await self.app.store.game_manager.game_end(
                                bot_points, users_points, update, cur_session.id
                            )

                        else:

                            async with self.app.database.session() as session:
                                res = await session.execute(
                                    select(QuestionModel).order_by(
                                        func.random()
                                    )
                                )
                                new_questions = res.scalars().all()

                                for question in new_questions:
                                    if question.id not in all_questions_list:
                                        new_question = question
                                        break

                                await session.execute(
                                    insert(GamesHistoryModel).values(
                                        session_id=cur_session.id,
                                        guestion_id=new_question.id,
                                        is_answer_right=False,
                                    )
                                )

                                await session.commit()

                            await self.app.store.game_manager.game_round(
                                update, new_question.title, cur_session.id
                            )
