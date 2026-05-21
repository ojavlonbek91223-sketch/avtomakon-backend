package server

import (
	"context"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	fiberlogger "github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"go.uber.org/zap"

	"github.com/avtomakon/backend/internal/config"
	"github.com/avtomakon/backend/internal/handler"
	"github.com/avtomakon/backend/internal/pkg/jwt"
	ws "github.com/avtomakon/backend/internal/websocket"
)

// Deps — server'ga zarur barcha bog'liqliklar.
type Deps struct {
	Config                     *config.Config
	Logger                     *zap.Logger
	JWTManager                 *jwt.Manager
	AuthHandler                *handler.AuthHandler
	UserHandler                *handler.UserHandler
	PostHandler                *handler.PostHandler
	BusinessHandler            *handler.BusinessHandler
	BusinessApplicationHandler *handler.BusinessApplicationHandler
	MarketHandler              *handler.MarketHandler
	ChatHandler                *handler.ChatHandler
	UploadHandler              *handler.UploadHandler
	FileHandler                *handler.FileHandler
	CommentHandler             *handler.CommentHandler
	NotificationHandler        *handler.NotificationHandler
	CartHandler                *handler.CartHandler
	OrderHandler               *handler.OrderHandler
	ReviewHandler              *handler.ReviewHandler
	VideoHandler               *handler.VideoHandler
	WSHub                      *ws.Hub
}

type Server struct {
	app    *fiber.App
	cfg    *config.Config
	logger *zap.Logger
}

func New(d *Deps) (*Server, error) {
	app := fiber.New(fiber.Config{
		AppName:               d.Config.AppName,
		ServerHeader:          "",
		DisableStartupMessage: true,
		ErrorHandler:          errorHandler(d.Logger),
		ReadTimeout:           30 * time.Second,
		WriteTimeout:          30 * time.Second,
		BodyLimit:             50 * 1024 * 1024, // 50 MB
	})

	app.Use(recover.New())
	app.Use(requestid.New())
	app.Use(fiberlogger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins:     d.Config.CORSAllowedOrigins,
		AllowMethods:     "GET,POST,PUT,DELETE,PATCH,OPTIONS",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowCredentials: false,
		MaxAge:           300,
	}))

	s := &Server{app: app, cfg: d.Config, logger: d.Logger}
	s.registerRoutes(d)

	return s, nil
}

func (s *Server) Start() error {
	return s.app.Listen(fmt.Sprintf(":%s", s.cfg.AppPort))
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.app.ShutdownWithContext(ctx)
}

func errorHandler(logger *zap.Logger) fiber.ErrorHandler {
	return func(c *fiber.Ctx, err error) error {
		code := fiber.StatusInternalServerError
		message := "Ichki server xatosi"

		if fe, ok := err.(*fiber.Error); ok {
			code = fe.Code
			message = fe.Message
		}

		if code >= 500 {
			logger.Error("server xatosi",
				zap.String("path", c.Path()),
				zap.String("method", c.Method()),
				zap.Error(err))
		}

		return c.Status(code).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    code,
				"message": message,
			},
		})
	}
}
