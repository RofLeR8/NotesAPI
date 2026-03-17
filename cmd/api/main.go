package main

import (
	"NotesAPI/internal/config"
	"NotesAPI/internal/handlers"
	"NotesAPI/internal/middleware"
	"NotesAPI/internal/models"
	"NotesAPI/internal/repository"
	"NotesAPI/internal/utils"
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()

	db, err := config.NewPostgresDB(cfg.DSN())
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	if err := db.AutoMigrate(&models.User{}, &models.Note{}); err != nil {
		log.Fatalf("Failed to  migrate database: %v", err)
	}

	jwtConfig := utils.JWTConfig{
		SecretKey:      "asdasdasdasddssfgdsfg1235",
		AccessExpHours: 12,
	}

	userRepo := repository.NewUserRepository(db)
	noteRepo := repository.NewNoteRepository(db)

	userHandler := handlers.NewUserHandler(userRepo, jwtConfig)
	noteHandler := handlers.NewNoteHandler(noteRepo)

	r := gin.Default()

	users := r.Group("/api/users")
	{
		users.POST("/register", userHandler.Register)
		users.POST("/login", userHandler.Login)
	}
	protected := r.Group("/api")
	protected.Use(middleware.JWTMiddleware(jwtConfig.SecretKey))
	{
		protected.GET("/users/:id", userHandler.GetUser)

		notes := protected.Group("/notes")
		{
			notes.POST("", noteHandler.CreateNote)
			notes.GET("", noteHandler.GetNotes)
			notes.GET("/:id", noteHandler.GetNote)
			notes.PUT("/:id", noteHandler.UpdateNote)
			notes.DELETE("/:id", noteHandler.DeleteNote)
		}
	}

	log.Printf("Server starting on port %s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

}
