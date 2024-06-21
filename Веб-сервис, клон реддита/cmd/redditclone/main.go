package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
	"log"
	"net/http"
	"os"
	"redditclone/configs"
	"redditclone/internal/handlers"
	"redditclone/internal/middleware"
	"redditclone/internal/posts"
	"redditclone/internal/sessions"
	"redditclone/internal/user"

	"github.com/gomodule/redigo/redis"
)

func homeHandler(w http.ResponseWriter, r *http.Request) {

	html, err := os.ReadFile("../../static/html/index.html")
	if err != nil {
		// Обработка ошибки, если файл не найден
		http.Error(w, "File not found", 404)
		return
	}

	// Установка Content-Type
	w.Header().Set("Content-Type", "text/html")

	// Отправка содержимого файла
	_, err = w.Write(html)
	if err != nil {
		log.Println(err.Error())
	}
}

func main() {
	config, err := configs.LoadConfig()
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	// Настраиваем подключение к mongodb.
	ctx := context.Background()
	mongoAddr := fmt.Sprintf("mongodb://%s", config.MongoDB.Host)
	mongoDB, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoAddr))
	if err != nil {
		log.Fatalf("Error opening connection to database: %v", err)
	}
	defer func() {
		if disconnectErr := mongoDB.Disconnect(ctx); disconnectErr != nil {
			log.Fatalf("Failed to disconnect from MongoDB: %v", disconnectErr)
		} else {
			fmt.Println("Disconnected from MongoDB.")
		}
	}()

	if err = mongoDB.Ping(ctx, nil); err != nil {
		log.Printf("Error connecting to database: %v", err)
	}
	log.Println("Успешное подключение к MongoDB!")
	collection := mongoDB.Database("golang").Collection("posts")

	// Настраиваем подключение к mysql.
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s",
		config.MySQL.User,
		config.MySQL.Password,
		config.MySQL.Host,
		config.MySQL.Port,
		config.MySQL.Name)

	mysql, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Printf("Error opening connection to database: %v", err)
	}

	err = mysql.Ping()
	if err != nil {
		log.Printf("Error connecting to database: %v", err)
	}
	log.Println("Успешное подключение к MySQL!")

	// Настраиваем подключение к redis
	redisAddr := fmt.Sprintf("redis://%s:@%s:%d/0", config.Redis.User, config.Redis.Host, config.Redis.Port)
	addr := flag.String("addr", redisAddr, "help message for flagname")
	redisConn, err := redis.DialURL(*addr)
	if err != nil {
		log.Printf("Error connecting to redis: %v", err)
	}

	sessManager := sessions.NewSessionManager(redisConn)

	zapLogger, err := zap.NewProduction()
	if err != nil {
		log.Printf("Error making new logger: %v", err)
	}
	defer func() {
		if err = zapLogger.Sync(); err != nil {
			log.Fatalf("Failed to sync logger: %v", err)
		}
	}()
	logger := zapLogger.Sugar()

	userRepo := user.NewMysqlRepo(mysql)
	postsRepo := posts.NewMongoRepo(collection)

	userHandler := &handlers.UserHandler{
		UserRepo: userRepo,
		Logger:   logger,
		Sessions: sessManager,
	}

	postsHandler := &handlers.PostsHandler{
		PostsRepo: postsRepo,
		Logger:    logger,
		Sessions:  sessManager,
	}

	r := mux.NewRouter()

	r.HandleFunc("/", homeHandler)
	r.PathPrefix("/a/").HandlerFunc(homeHandler)

	staticHandler := http.StripPrefix("/static/", http.FileServer(http.Dir("../../static/")))
	r.PathPrefix("/static/").Handler(staticHandler)

	r.HandleFunc("/api/login", userHandler.Login).Methods("POST")
	r.HandleFunc("/api/register", userHandler.Register).Methods("POST")

	r.HandleFunc("/api/posts/", postsHandler.GetAllPosts).Methods("GET")
	r.HandleFunc("/api/posts/{CATEGORY_NAME}", postsHandler.GetCategoryPosts).Methods("GET")
	r.HandleFunc("/api/user/{USER_LOGIN}", postsHandler.GetUserPosts).Methods("GET")
	r.HandleFunc("/api/post/{POST_ID}", postsHandler.GetPost).Methods("GET")
	r.HandleFunc("/api/post/{POST_ID}/upvote", postsHandler.UpVotePost).Methods("GET")
	r.HandleFunc("/api/post/{POST_ID}/downvote", postsHandler.DownVotePost).Methods("GET")
	r.HandleFunc("/api/post/{POST_ID}/unvote", postsHandler.UnVotePost).Methods("GET")

	r.HandleFunc("/api/posts", postsHandler.MakePost).Methods("POST")
	r.HandleFunc("/api/post/{POST_ID}", postsHandler.DeletePost).Methods("DELETE")

	r.HandleFunc("/api/post/{POST_ID}", postsHandler.MakeComment).Methods("POST")
	r.HandleFunc("/api/post/{POST_ID}/{COMMENT_ID}", postsHandler.DeleteComment).Methods("DELETE")

	middleWares := middleware.AccessLog(logger, r)

	log.Println("starting server at :8080")
	err = http.ListenAndServe(":8080", middleWares)
	if err != nil {
		panic(err)
	}
}
