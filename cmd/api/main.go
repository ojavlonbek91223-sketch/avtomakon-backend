package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"go.uber.org/zap"

	"github.com/avtomakon/backend/internal/config"
	"github.com/avtomakon/backend/internal/database"
	"github.com/avtomakon/backend/internal/handler"
	"github.com/avtomakon/backend/internal/pkg/jwt"
	"github.com/avtomakon/backend/internal/repository/postgres"
	"github.com/avtomakon/backend/internal/server"
	"github.com/avtomakon/backend/internal/service"
	"github.com/avtomakon/backend/internal/storage"
	ws "github.com/avtomakon/backend/internal/websocket"
)

func main() {
	_ = godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	logger, err := newLogger(cfg.AppEnv)
	if err != nil {
		log.Fatalf("logger: %v", err)
	}
	defer logger.Sync()

	ctx := context.Background()

	// Ma'lumotlar bazasi
	pool, err := database.NewPostgresPool(ctx, cfg.DB.DSN())
	if err != nil {
		logger.Fatal("postgres", zap.Error(err))
	}
	defer pool.Close()
	logger.Info("postgres ulandi")

	// Redis (ixtiyoriy — bo'lmasa ham backend ishlaydi)
	redisClient, err := database.NewRedisClient(
		ctx, cfg.Redis.Addr(), cfg.Redis.Password, cfg.Redis.DB,
	)
	if err != nil {
		logger.Warn("redis ulanmadi (Redis o'rnatilmagan bo'lishi mumkin)", zap.Error(err))
		redisClient = nil
	} else {
		defer redisClient.Close()
		logger.Info("redis ulandi")
	}
	_ = redisClient

	// JWT
	jwtMgr := jwt.NewManager(cfg.JWT.Secret, cfg.JWT.AccessTTL)

	// Repositories
	userRepo := postgres.NewUserRepository(pool)
	refreshRepo := postgres.NewRefreshTokenRepository(pool)
	otpRepo := postgres.NewOTPRepository(pool)
	postRepo := postgres.NewPostRepository(pool)
	bizRepo := postgres.NewBusinessRepository(pool)
	bizAppRepo := postgres.NewBusinessApplicationRepository(pool)
	marketRepo := postgres.NewMarketRepository(pool)
	chatRepo := postgres.NewChatRepository(pool)
	commentRepo := postgres.NewCommentRepository(pool)
	notifRepo := postgres.NewNotificationRepository(pool)
	cartRepo := postgres.NewCartRepository(pool)
	orderRepo := postgres.NewOrderRepository(pool)
	reviewRepo := postgres.NewReviewRepository(pool)
	videoRepo := postgres.NewVideoRepository(pool)

	// WebSocket hub
	hub := ws.NewHub(logger)

	// MinIO (ixtiyoriy — bo'lmasa ham backend ishlaydi, faqat upload yo'q)
	var minioClient *storage.MinIOClient
	if cfg.MinIO.AccessKey != "" && cfg.MinIO.SecretKey != "" {
		minioClient, err = storage.NewMinIOClient(
			cfg.MinIO.Endpoint, cfg.MinIO.AccessKey, cfg.MinIO.SecretKey,
			cfg.MinIO.Bucket, cfg.MinIO.UseSSL, cfg.MinIO.Region, cfg.MinIO.PublicURL,
		)
		if err != nil {
			logger.Warn("minio ulanmadi (upload ishlamaydi)", zap.Error(err))
			minioClient = nil
		} else {
			logger.Info("minio ulandi")
		}
	} else {
		logger.Warn("minio sozlanmagan — fayl yuklash o'chirilgan (R2/MinIO kalitlari yo'q)")
	}

	// SMS provider (dev'da mock, production'da real provayder)
	var smsProvider service.SMSProvider
	isDev := cfg.AppEnv != "production"
	if isDev {
		smsProvider = service.NewMockSMSProvider(logger)
	} else {
		// TODO: Eskiz yoki Playmobile implementatsiyasi
		smsProvider = service.NewMockSMSProvider(logger)
	}

	// Services
	otpSvc := service.NewOTPService(otpRepo, smsProvider, isDev)
	authSvc := service.NewAuthService(userRepo, refreshRepo, otpSvc, jwtMgr, cfg.JWT.RefreshTTL)
	postSvc := service.NewPostService(postRepo)
	bizSvc := service.NewBusinessService(bizRepo)
	userSvc := service.NewUserService(userRepo)
	bizAppSvc := service.NewBusinessApplicationService(bizAppRepo)
	marketSvc := service.NewMarketService(marketRepo)
	chatSvc := service.NewChatService(chatRepo, hub)
	uploadSvc := service.NewUploadService(pool, minioClient)
	commentSvc := service.NewCommentService(commentRepo)
	notifSvc := service.NewNotificationService(notifRepo, hub)
	cartSvc := service.NewCartService(cartRepo, orderRepo)
	orderSvc := service.NewOrderService(orderRepo)
	reviewSvc := service.NewReviewService(reviewRepo)
	videoSvc := service.NewVideoService(videoRepo)

	// Bildirishnoma DI (circular bog'liqlikni oldini olish)
	postSvc.SetNotificationService(notifSvc)
	userSvc.SetNotificationService(notifSvc)

	// Handlers
	authHandler := handler.NewAuthHandler(authSvc)
	postHandler := handler.NewPostHandler(postSvc)
	bizHandler := handler.NewBusinessHandler(bizSvc)
	userHandler := handler.NewUserHandler(userSvc)
	bizAppHandler := handler.NewBusinessApplicationHandler(bizAppSvc)
	marketHandler := handler.NewMarketHandler(marketSvc)
	chatHandler := handler.NewChatHandler(chatSvc)
	uploadHandler := handler.NewUploadHandler(uploadSvc)
	fileHandler := handler.NewFileHandler(minioClient)
	commentHandler := handler.NewCommentHandler(commentSvc)
	notifHandler := handler.NewNotificationHandler(notifSvc)
	cartHandler := handler.NewCartHandler(cartSvc)
	orderHandler := handler.NewOrderHandler(orderSvc)
	reviewHandler := handler.NewReviewHandler(reviewSvc)
	videoHandler := handler.NewVideoHandler(videoSvc)

	// Server
	srv, err := server.New(&server.Deps{
		Config:                     cfg,
		Logger:                     logger,
		JWTManager:                 jwtMgr,
		AuthHandler:                authHandler,
		UserHandler:                userHandler,
		PostHandler:                postHandler,
		BusinessHandler:            bizHandler,
		BusinessApplicationHandler: bizAppHandler,
		MarketHandler:              marketHandler,
		ChatHandler:                chatHandler,
		UploadHandler:              uploadHandler,
		FileHandler:                fileHandler,
		CommentHandler:             commentHandler,
		NotificationHandler:        notifHandler,
		CartHandler:                cartHandler,
		OrderHandler:               orderHandler,
		ReviewHandler:              reviewHandler,
		VideoHandler:               videoHandler,
		WSHub:                      hub,
	})
	if err != nil {
		logger.Fatal("server", zap.Error(err))
	}

	// Graceful shutdown
	stopCtx, stop := signal.NotifyContext(context.Background(),
		os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		if err := srv.Start(); err != nil {
			logger.Fatal("server start", zap.Error(err))
		}
	}()

	logger.Info("server ishga tushdi",
		zap.String("port", cfg.AppPort),
		zap.String("env", cfg.AppEnv))

	<-stopCtx.Done()
	logger.Info("server to'xtatilmoqda...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("shutdown", zap.Error(err))
	}
	logger.Info("server to'xtatildi")
}

func newLogger(env string) (*zap.Logger, error) {
	if env == "production" {
		return zap.NewProduction()
	}
	return zap.NewDevelopment()
}
