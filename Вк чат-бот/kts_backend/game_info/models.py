from dataclasses import dataclass

from sqlalchemy.orm import relationship

from kts_backend.store.database.sqlalchemy_base import db
from sqlalchemy import Column, BigInteger, String, Boolean, ForeignKey


@dataclass
class Question:
    id: int
    title: str
    answer: str


@dataclass
class GamesHistory:
    id: int | None
    session_id: int | None
    guestion_id: int | None
    player_id: int | None
    is_answer_right: bool


@dataclass
class Sessions:
    id: int | None
    group_id: int | None
    status: str | None
    capitan_id: int | None


@dataclass
class TeamPlayer:
    id: int
    session_id: int
    player_id: int
    vk_id: int
    first_name: str
    last_name: str


class QuestionModel(db):
    __tablename__ = "questions"
    id = Column(BigInteger, primary_key=True)
    title = Column(String, unique=True)
    answer = Column(String)


class GamesHistoryModel(db):
    __tablename__ = "gameshistory"
    id = Column(BigInteger, primary_key=True)
    session_id = Column(BigInteger, ForeignKey("sessions.id"))
    guestion_id = Column(BigInteger, ForeignKey("game_info.id"))
    player_id = Column(BigInteger)
    is_answer_right = Column(Boolean)


class SessionsModel(db):
    __tablename__ = "sessions"
    id = Column(BigInteger, primary_key=True)
    group_id = Column(BigInteger)
    status = Column(String)
    capitan_id = Column(BigInteger)


class TeamPlayerModel(db):
    __tablename__ = "teamplayer"
    id = Column(BigInteger, primary_key=True)
    session_id = Column(BigInteger, ForeignKey("sessions.id"))
    player_id = Column(BigInteger)
    first_name = Column(String)
    last_name = Column(String)
