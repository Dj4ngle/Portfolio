import asyncio
import random

from sqlalchemy import select, insert, update
from vk_api.keyboard import VkKeyboard, VkKeyboardColor

from kts_backend.game_info.models import TeamPlayerModel, SessionsModel
from kts_backend.store.vk_api.dataclasses import Message


class GameManager:
    def __init__(self, app: "Application", *args, **kwargs):
        self.app = app

    async def game_round(self, _update, question, session_id):

        await self.app.store.vk_api.send_message(
            Message(
                peer_id=_update.object.body["peer_id"],
                text=question,
            ),
        )

        # await self.app.store.vk_api.send_message(
        #     Message(
        #         peer_id=_update.object.body["peer_id"],
        #         text="На размышление одна минута",
        #     ),
        # )
        #
        # await asyncio.sleep(50)
        #
        # await self.app.store.vk_api.send_message(
        #     Message(
        #         peer_id=_update.object.body["peer_id"],
        #         text="Осталось 10 секунд",
        #     ),
        # )
        #
        # await asyncio.sleep(10)

        async with self.app.database.session() as session:
            res = await session.execute(
                select(TeamPlayerModel).where(
                    TeamPlayerModel.session_id == session_id
                )
            )
            answerers = res.scalars().all()

        keyboard = VkKeyboard(inline=True)

        if len(answerers)>6:

            # Игроков больше 6
            await self.app.store.vk_api.send_message(
                Message(
                    peer_id=_update.object.body["peer_id"],
                    text="Капитан, выберите кто отвечает",
                ),
            )
            i=0
            for ans in answerers:
                keyboard.add_button(
                    ans.first_name + " " + ans.last_name, VkKeyboardColor.PRIMARY
                )
                keyboard.add_line()
                i+=1
                if i >= 5:
                    keyboard.lines.pop()
                    i=0
                    await self.app.store.vk_api.send_keyboard_message(
                        Message(
                            peer_id=_update.object.body["peer_id"],
                            text="_",
                        ),
                        keyboard.get_keyboard(),
                    )
                    keyboard.lines.clear()
                    keyboard = VkKeyboard(inline=True)
            keyboard.lines.pop()

            await self.app.store.vk_api.send_keyboard_message(
                Message(
                    peer_id=_update.object.body["peer_id"],
                    text="_",
                ),
                keyboard.get_keyboard(),
            )

        else:

            #Игроков меньше 6
            for ans in answerers:
                keyboard.add_button(
                    ans.first_name + " " + ans.last_name, VkKeyboardColor.PRIMARY
                )
                keyboard.add_line()
            keyboard.lines.pop()

            await self.app.store.vk_api.send_keyboard_message(
                Message(
                    peer_id=_update.object.body["peer_id"],
                    text="Капитан, выберите кто отвечает",
                ),
                keyboard.get_keyboard(),
            )

    async def choose_the_capitan(self, _update):

        async with self.app.database.session() as session:
            res = await session.execute(
                select(SessionsModel).where(
                    SessionsModel.group_id == _update.object.peer_id
                )
            )
            is_session_opened = res.scalars().all()

            if is_session_opened == []:
                await session.execute(
                    insert(SessionsModel).values(
                        group_id=_update.object.peer_id,
                        status="Started",
                        capitan_id=None,
                    )
                )
                await session.commit()
            else:
                await session.execute(
                    update(SessionsModel)
                    .where(SessionsModel.group_id == _update.object.peer_id)
                    .values(status="Closed")
                )
                await session.execute(
                    insert(SessionsModel).values(
                        group_id=_update.object.peer_id,
                        status="Started",
                        capitan_id=None,
                    )
                )
                await session.commit()

        users_list = await self.app.store.vk_api.get_all_users_from_chat(
            _update.object.body["peer_id"]
        )
        random_user = random.choice(users_list)

        async with self.app.database.session() as session:
            res = await session.execute(
                update(SessionsModel)
                .where(
                    SessionsModel.group_id == _update.object.peer_id,
                    SessionsModel.status != "Closed",
                )
                .values(capitan_id=random_user.id)
                .returning(SessionsModel.id)
            )
            session_id = res.scalars().first()

            for user in users_list:
                await session.execute(
                    insert(TeamPlayerModel).values(
                        session_id=session_id,
                        player_id=user.id,
                        first_name=user.first_name,
                        last_name=user.last_name,
                    )
                )
            await session.commit()

        message_to_user = f"Капитан команды @id{random_user.id} ({random_user.first_name} {random_user.last_name}!)"

        keyboard = VkKeyboard(inline=True)
        keyboard.add_button("Старт игры!!!", VkKeyboardColor.PRIMARY)

        await self.app.store.vk_api.send_keyboard_message(
            Message(
                peer_id=_update.object.peer_id,
                text=message_to_user,
            ),
            keyboard.get_keyboard(),
        )

    async def game_end(self, bot_points, users_points, _update, session_id):
        await self.app.store.vk_api.send_message(
            Message(
                peer_id=_update.object.peer_id,
                text=f"Игра окончена! "
                f"Бот: {bot_points} "
                f"Игроки: {users_points}",
            )
        )

        if users_points > bot_points:
            await self.app.store.vk_api.send_message(
                Message(
                    peer_id=_update.object.peer_id,
                    text="Победили игроки!!!!",
                )
            )
        else:
            await self.app.store.vk_api.send_message(
                Message(
                    peer_id=_update.object.peer_id,
                    text="Победил бот!!!!",
                )
            )

        async with self.app.database.session() as session:
            await session.execute(
                update(SessionsModel)
                .where(SessionsModel.id == session_id)
                .values(status="Closed")
            )
            await session.commit()
