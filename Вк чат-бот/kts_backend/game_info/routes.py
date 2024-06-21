import typing

from kts_backend.game_info.views import (
    QuestionAddView,
    QuestionListView,
    QuestionEditeView,
    SessionsView,
)

if typing.TYPE_CHECKING:
    from kts_backend.web.app import Application


def setup_routes(app: "Application"):
    app.router.add_view("/add_question", QuestionAddView)
    app.router.add_view("/list_questions", QuestionListView)
    app.router.add_view("/edite_question", QuestionEditeView)
    app.router.add_view("/list_sessions", SessionsView)
