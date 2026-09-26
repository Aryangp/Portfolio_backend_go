package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"portfolio-backend/internal/config"
	"portfolio-backend/internal/repository"
	"portfolio-backend/internal/router"
	"portfolio-backend/internal/tools"

	"go.mongodb.org/mongo-driver/mongo"
)

func main() {
	// 1. Load configuration
	cfg := config.Load()

	// 2. Initialize repositories (Cached MongoDB backed with fallback to in-memory)
	var projectRepo repository.ProjectRepository
	var mongoClient *mongo.Client

	if cfg.MongoURI != "" {
		mongoCtx, mongoCancel := context.WithTimeout(context.Background(), 15*time.Second)
		client, err := repository.ConnectMongo(mongoCtx, cfg.MongoURI)
		mongoCancel()

		if err != nil {
			log.Printf("⚠️ Could not connect to MongoDB at %s (%v)", cfg.MongoURI, err)
			log.Println("ℹ️ Falling back to In-Memory Project Repository")

		} else {
			mongoClient = client
			coll := client.Database(cfg.MongoDBName).Collection(cfg.MongoProjectsColl)
			cachedRepo, err := repository.NewCachedMongoProjectRepository(coll, 10*time.Minute)
			if err != nil {
				log.Printf("⚠️ Failed to init cached mongo repository: %v", err)

			} else {
				log.Printf("🗄️ Connected to MongoDB (%s / %s)", cfg.MongoDBName, cfg.MongoProjectsColl)
				projectRepo = cachedRepo
			}
		}
	}

	// 3. Initialize Dynamic Tool Registry
	toolRegistry := tools.NewRegistry()
	ghClient := tools.NewGitHubClient(cfg.GitHubToken)
	toolRegistry.RegisterAll(tools.NewGitHubTools(ghClient, cfg.GitHubUsername)...)
	log.Printf("🛠️ Registered %d AI tools: %v", len(toolRegistry.ListNames()), toolRegistry.ListNames())

	// 4. Initialize Gemini Chat Repository if API Key is configured
	var chatRepo repository.ChatRepository
	if cfg.GeminiAPIKey != "" {
		geminiRepo, err := repository.NewGeminiChatRepository(context.Background(), cfg.GeminiAPIKey, cfg.GeminiModel, toolRegistry)
		if err != nil {
			log.Printf("⚠️ Failed to initialize Gemini chat client: %v", err)
		} else {
			log.Printf("🤖 Gemini AI Chatbot ready (model: %s)", cfg.GeminiModel)
			chatRepo = geminiRepo
		}
	} else {
		log.Println("ℹ️ Gemini API key not set; chatbot running in offline mode")
	}

	// 3. Initialize HTTP router
	appRouter := router.New(cfg, projectRepo, chatRepo)

	// 4. Configure HTTP Server
	serverAddr := fmt.Sprintf(":%s", cfg.Port)
	srv := &http.Server{
		Addr:         serverAddr,
		Handler:      appRouter,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 60 * time.Second, // Increased for long-lived SSE streaming responses
		IdleTimeout:  60 * time.Second,
	}

	// 5. Start server in background goroutine
	go func() {
		log.Printf("==================================================")
		log.Printf("🚀 Portfolio Backend API Server")
		log.Printf("📡 Listening on: http://localhost:%s", cfg.Port)
		log.Printf("🌐 Environment:  %s", cfg.AppEnv)
		log.Printf("🩺 Health Check: http://localhost:%s/api/v1/health", cfg.Port)
		log.Printf("📂 Projects API: http://localhost:%s/api/v1/projects", cfg.Port)
		log.Printf("🤖 AI Chatbot:   http://localhost:%s/api/v1/chat/stream", cfg.Port)
		log.Printf("==================================================")

		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("❌ Server failed to start: %v", err)
		}
	}()

	// 6. Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	<-quit

	log.Println("\n🛑 Shutting down server gracefully...")

	// 7. Context with timeout for shutdown completion
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if chatRepo != nil {
		_ = chatRepo.Close()
	}

	if mongoClient != nil {
		if err := mongoClient.Disconnect(ctx); err != nil {
			log.Printf("⚠️ Error disconnecting MongoDB client: %v", err)
		} else {
			log.Println("🔌 MongoDB connection closed.")
		}
	}

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("❌ Server forced to shutdown: %v", err)
	}

	log.Println("✅ Server gracefully stopped.")
}
