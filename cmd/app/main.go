package main

import (
	"database/sql"
	"log"
	"net/http"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	httpSwagger "github.com/swaggo/http-swagger"

	_ "github.com/SniperMOCs/MPT_Project/docs"

	"github.com/SniperMOCs/MPT_Project/internal/auth"
	"github.com/SniperMOCs/MPT_Project/internal/comment"
	"github.com/SniperMOCs/MPT_Project/internal/config"
	"github.com/SniperMOCs/MPT_Project/internal/content"
	"github.com/SniperMOCs/MPT_Project/internal/cors"
	"github.com/SniperMOCs/MPT_Project/internal/follow"
	"github.com/SniperMOCs/MPT_Project/internal/like"
	"github.com/SniperMOCs/MPT_Project/internal/notification"
	"github.com/SniperMOCs/MPT_Project/internal/user"
)

// @title 			Zapretgram Backend API
// @version         1.0
// @description     Backend-приложение на Go, реализующее функциональность социальной сети, аналогичной Instagram.
// @description     Предоставляет REST API для управления пользователями, публикациями, комментариями, лайками, подписками и уведомлениями в реальном времени через WebSocket.

// @contact.name   API Support
// @contact.email  snipermocs@gmail.com

// @host      localhost:8080
// @BasePath  /api

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Введите токен в формате: Bearer {token}

func openDB(dsn string) (*sql.DB, error) {
	log.Println(dsn)
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(time.Hour)

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println(".env file not found")
	}
	cfg := config.Load()
	db, err := openDB(cfg.DBUrl)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	// User
	userRepo := user.NewUserRepository(db)
	userService := user.NewUserService(userRepo)
	userHandler := user.NewUserHandler(userService)

	// Auth
	jwtManager := auth.NewJWTManager([]byte(cfg.JwtSecret))
	authService := auth.NewAuthService(
		userService,
		jwtManager,
		cfg,
	)
	authHandler := auth.NewAuthHandler(authService)

	// Notification
	notificationRepo := notification.NewNotificationRepository(db)
	hub := notification.NewWebSocketHub()
	notificationService := notification.NewNotificationService(notificationRepo, hub)
	notificationHandler := notification.NewNotificationHandler(notificationService, hub)

	// Content
	contentRepo := content.NewContentRepository(db)
	contentService := content.NewContentService(contentRepo)
	contentHandler := content.NewContentHandler(contentService)

	// Follows
	followRepo := follow.NewFollowRepository(db)
	followService := follow.NewFollowService(followRepo, notificationService)
	followHandler := follow.NewFollowHandler(followService)

	// Likes
	likeRepo := like.NewLikeRepository(db)
	likeService := like.NewLikeService(likeRepo, notificationService, contentService)
	likeHandler := like.NewLikeHandler(likeService)

	// Comments
	commentRepo := comment.NewCommentRepository(db)
	commentService := comment.NewCommentService(commentRepo, notificationService, contentService)
	commentHandler := comment.NewCommentHandler(commentService)

	// HTTP mux
	mux := http.NewServeMux()

	// Route register
	// USERS
	mux.HandleFunc("/api/users/login", authHandler.Login)
	mux.HandleFunc("/api/users/register", userHandler.HandleRegister)
	mux.HandleFunc("/api/users/id/", userHandler.HandleGetByID)
	mux.HandleFunc("/api/users/username/", userHandler.HandleGetByUsername)
	mux.HandleFunc("/api/users/search", userHandler.HandleSearchByFullName)
	mux.Handle("/api/users/me", auth.Middleware(jwtManager, cfg)(http.HandlerFunc(userHandler.HandleMe)))
	mux.Handle("/api/users/update", auth.Middleware(jwtManager, cfg)(http.HandlerFunc(userHandler.HandleUpdate)))
	mux.Handle("/api/users/setPassword", auth.Middleware(jwtManager, cfg)(http.HandlerFunc(userHandler.HandleSetPassword)))
	mux.Handle("/api/users/setAvatar", auth.Middleware(jwtManager, cfg)(http.HandlerFunc(userHandler.HandleUploadAvatar)))

	// CONTENT
	mux.HandleFunc("/api/content/{id}", contentHandler.GetByID)
	mux.HandleFunc("/api/content/user/{id}", contentHandler.GetByAuthor)
	mux.Handle("/api/content/subscriptions", auth.Middleware(jwtManager, cfg)(http.HandlerFunc(contentHandler.GetSubscriptions)))
	mux.Handle("/api/content/create", auth.Middleware(jwtManager, cfg)(http.HandlerFunc(contentHandler.Create)))
	mux.Handle("/api/content/delete/{id}", auth.Middleware(jwtManager, cfg)(http.HandlerFunc(contentHandler.Delete)))
	mux.Handle("/api/content/update/{id}", auth.Middleware(jwtManager, cfg)(http.HandlerFunc(contentHandler.Update)))
	mux.Handle("/api/content/media/{id}", auth.Middleware(jwtManager, cfg)(http.HandlerFunc(contentHandler.UploadMedia)))

	// FOLLOW
	mux.HandleFunc("/api/followers/{id}", followHandler.GetFollowers)
	mux.HandleFunc("/api/following/{id}", followHandler.GetFollowing)
	mux.HandleFunc("/api/follow/info/{id}", followHandler.FollowInfo)
	mux.Handle("/api/follower/{id}", auth.Middleware(jwtManager, cfg)(http.HandlerFunc(followHandler.IsFollower)))
	mux.Handle("/api/follow/{id}", auth.Middleware(jwtManager, cfg)(http.HandlerFunc(followHandler.Create)))
	mux.Handle("/api/unfollow/{id}", auth.Middleware(jwtManager, cfg)(http.HandlerFunc(followHandler.Delete)))

	// LIKE
	mux.HandleFunc("/api/likes/{id}", likeHandler.GetLikesCount)
	mux.Handle("/api/liked/{id}", auth.Middleware(jwtManager, cfg)(http.HandlerFunc(likeHandler.Exists)))
	mux.Handle("/api/like/{id}", auth.Middleware(jwtManager, cfg)(http.HandlerFunc(likeHandler.Create)))
	mux.Handle("/api/unlike/{id}", auth.Middleware(jwtManager, cfg)(http.HandlerFunc(likeHandler.Delete)))

	// COMMENT
	mux.HandleFunc("/api/comment/content/{id}", commentHandler.GetByContentId)
	mux.Handle("/api/comment", auth.Middleware(jwtManager, cfg)(http.HandlerFunc(commentHandler.Create)))
	mux.Handle("/api/comment/delete/{id}", auth.Middleware(jwtManager, cfg)(http.HandlerFunc(commentHandler.Delete)))

	// NOTIFICATION
	mux.Handle("/api/notifications", auth.Middleware(jwtManager, cfg)(http.HandlerFunc(notificationHandler.GetMy)))
	mux.Handle("/api/notifications/read", auth.Middleware(jwtManager, cfg)(http.HandlerFunc(notificationHandler.MarkAsRead)))
	mux.Handle("/api/notifications/count", auth.Middleware(jwtManager, cfg)(http.HandlerFunc(notificationHandler.GetUnreadCount)))

	// STATIC
	mux.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir("./uploads"))))

	// WEBSOCKET
	mux.Handle("/ws/notifications", auth.Middleware(jwtManager, cfg)(http.HandlerFunc(notificationHandler.ServeWS)))

	// SWAGGER
	mux.HandleFunc("/swagger/", httpSwagger.WrapHandler)

	// Middleware
	corsMux := cors.CORSHandler(mux)

	// Server settings
	server := &http.Server{
		Addr:         ":8080",
		Handler:      corsMux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("Server started on %v\n", server.Addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server error: %v", err)
	}
}
