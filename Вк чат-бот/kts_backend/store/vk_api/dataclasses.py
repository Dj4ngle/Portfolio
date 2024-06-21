from dataclasses import dataclass


@dataclass
class UpdateObject:
    id: int
    user_id: int
    peer_id: int
    body: str


@dataclass
class Update:
    type: str
    object: UpdateObject


@dataclass
class Message:
    peer_id: int
    text: str


@dataclass
class User:
    id: int
    first_name: str
    last_name: str
