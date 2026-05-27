package server

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"

	"github.com/avtomakon/backend/internal/handler"
	"github.com/avtomakon/backend/internal/middleware"
	ws "github.com/avtomakon/backend/internal/websocket"
)

func (s *Server) registerRoutes(d *Deps) {
	s.app.Get("/health", handler.Health)
	s.app.Get("/ready", handler.Ready)

	api := s.app.Group("/api/v1")

	requireAuth := middleware.RequireAuth(d.JWTManager)
	optionalAuth := middleware.OptionalAuth(d.JWTManager)

	// ----- AUTH -----
	auth := api.Group("/auth")

	loginLimiter := limiter.New(limiter.Config{
		Max:        5,
		Expiration: time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP() + ":" + c.Path()
		},
	})
	otpLimiter := limiter.New(limiter.Config{
		Max:        3,
		Expiration: 5 * time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string { return c.IP() },
	})

	auth.Post("/otp/send", otpLimiter, d.AuthHandler.SendOTP)
	auth.Post("/register", loginLimiter, d.AuthHandler.Register)
	auth.Post("/login", loginLimiter, d.AuthHandler.Login)
	auth.Post("/refresh", d.AuthHandler.Refresh)
	auth.Post("/logout", d.AuthHandler.Logout)
	auth.Post("/logout-all", requireAuth, d.AuthHandler.LogoutAll)

	// ----- POSTS (Feed) -----
	posts := api.Group("/posts")
	posts.Get("/", optionalAuth, d.PostHandler.ListFeed)
	posts.Get("/saved", requireAuth, d.PostHandler.SavedPosts)
	posts.Post("/", requireAuth, d.PostHandler.Create)
	posts.Delete("/:id", requireAuth, d.PostHandler.Delete)
	posts.Post("/:id/react", requireAuth, d.PostHandler.SetReaction)
	posts.Delete("/:id/react", requireAuth, d.PostHandler.RemoveReaction)
	posts.Post("/:id/save", requireAuth, d.PostHandler.Save)
	posts.Delete("/:id/save", requireAuth, d.PostHandler.Unsave)

	// ----- BUSINESSES (Xarita) -----
	bizGroup := api.Group("/businesses")
	bizGroup.Get("/", d.BusinessHandler.Nearby)
	bizGroup.Get("/:id", d.BusinessHandler.Get)

	// ----- USERS (Profil) -----
	users := api.Group("/users")
	users.Get("/me", requireAuth, d.UserHandler.Me)
	users.Patch("/me", requireAuth, d.UserHandler.UpdateMe)
	users.Get("/:id", optionalAuth, d.UserHandler.Get)
	users.Get("/:id/posts", optionalAuth, d.PostHandler.UserPosts)
	users.Get("/:id/followers", optionalAuth, d.UserHandler.Followers)
	users.Get("/:id/following", optionalAuth, d.UserHandler.Following)
	users.Post("/:id/follow", requireAuth, d.UserHandler.Follow)
	users.Delete("/:id/follow", requireAuth, d.UserHandler.Unfollow)

	// ----- BIZNES ARIZA -----
	api.Post("/business-applications", requireAuth, d.BusinessApplicationHandler.Apply)
	api.Get("/business-applications/mine", requireAuth, d.BusinessApplicationHandler.Mine)

	// ----- MARKET -----
	api.Get("/categories", d.MarketHandler.Categories)
	api.Get("/promotions", d.MarketHandler.Promotions)
	products := api.Group("/products")
	products.Get("/", d.MarketHandler.Products)
	products.Get("/featured", d.MarketHandler.Featured)
	products.Get("/:id", d.MarketHandler.GetProduct)

	// ----- CHAT -----
	chat := api.Group("/conversations", requireAuth)
	chat.Get("/", d.ChatHandler.List)
	chat.Post("/", d.ChatHandler.Start)
	chat.Get("/:id/messages", d.ChatHandler.Messages)
	chat.Post("/:id/messages", d.ChatHandler.Send)
	chat.Post("/:id/read", d.ChatHandler.MarkRead)

	// ----- WebSocket -----
	s.app.Use("/ws", ws.UpgradeMiddleware())
	s.app.Get("/ws", ws.Handler(d.WSHub, d.JWTManager, s.logger))

	// ----- UPLOAD -----
	api.Post("/uploads", requireAuth, d.UploadHandler.Upload)

	// ----- FILES (MinIO proxy — tunnel orqali rasm/video ko'rsatish) -----
	s.app.Get("/files/*", d.FileHandler.Serve)

	// ----- COMMENTS -----
	api.Get("/posts/:id/comments", optionalAuth, d.CommentHandler.List)
	api.Post("/posts/:id/comments", requireAuth, d.CommentHandler.Create)
	api.Delete("/comments/:comment_id", requireAuth, d.CommentHandler.Delete)
	api.Post("/comments/:comment_id/like", requireAuth, d.CommentHandler.Like)

	// ----- NOTIFICATIONS -----
	api.Get("/notifications", requireAuth, d.NotificationHandler.List)
	api.Get("/notifications/unread-count", requireAuth, d.NotificationHandler.UnreadCount)
	api.Post("/notifications/read-all", requireAuth, d.NotificationHandler.MarkAllRead)
	api.Post("/push-tokens", requireAuth, d.NotificationHandler.RegisterPushToken)

	// ----- CART -----
	cart := api.Group("/cart", requireAuth)
	cart.Get("/", d.CartHandler.Get)
	cart.Post("/items", d.CartHandler.Add)
	cart.Patch("/items/:id", d.CartHandler.Update)
	cart.Delete("/items/:id", d.CartHandler.Remove)

	// ----- ORDERS -----
	orders := api.Group("/orders", requireAuth)
	orders.Get("/", d.OrderHandler.List)
	orders.Post("/", d.OrderHandler.Create)
	orders.Post("/:id/cancel", d.OrderHandler.Cancel)

	// ----- REVIEWS -----
	api.Get("/businesses/:id/reviews", d.ReviewHandler.ListBusinessReviews)
	api.Post("/businesses/:id/reviews", requireAuth, d.ReviewHandler.CreateBusinessReview)
	api.Get("/products/:id/reviews", d.ReviewHandler.ListProductReviews)
	api.Post("/products/:id/reviews", requireAuth, d.ReviewHandler.CreateProductReview)
	api.Get("/users/me/reviews", requireAuth, d.ReviewHandler.MyReviews)
	api.Post("/reviews/:id/reply", requireAuth, d.ReviewHandler.Reply)
	api.Delete("/reviews/:id", requireAuth, d.ReviewHandler.Delete)
}

func registerPlaceholderRoutes(api fiber.Router, requireAuth fiber.Handler) {
	// Hozircha placeholder qolmagan
	_ = api
	_ = requireAuth
}

func placeholder(name string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{
			"message": "endpoint hali yozilmagan",
			"handler": name,
		})
	}
}
