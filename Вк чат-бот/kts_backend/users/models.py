from dataclasses import dataclass

from kts_backend.store.database.sqlalchemy_base import db
from sqlalchemy import Column, BigInteger, String


class Users:
    id: int
    vk_id: int
    first_name: str
    last_name: str
    right_answers: int
    answers: int


class UsersModel(db):
    __tablename__ = "users"
    id = Column(BigInteger, primary_key=True)
    vk_id = Column(BigInteger, unique=True)
    first_name = Column(String)
    last_name = Column(String)
    right_answers = Column(BigInteger)
    answers = Column(BigInteger)
