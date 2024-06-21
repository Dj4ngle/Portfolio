from aiohttp.web_app import Application
from aiohttp_cors import CorsConfig

__all__ = ("register_urls",)


# def register_urls(application: Application, cors: CorsConfig):
def register_urls(application: Application):
    import kts_backend.users.urls
    from kts_backend.admin.routes import setup_routes as admin_setup_routes
    from kts_backend.game_info.routes import setup_routes as game_setup_routes

    # kts_backend.users.urls.register_urls(application, cors)
    admin_setup_routes(application)
    game_setup_routes(application)
