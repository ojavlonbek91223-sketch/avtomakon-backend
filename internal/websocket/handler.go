package websocket

import (
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/avtomakon/backend/internal/pkg/jwt"
)

// UpgradeMiddleware — Fiber'da WebSocket ulanish uchun.
func UpgradeMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	}
}

// Handler — Token query parameter orqali autentifikatsiya qilingan WS handler.
func Handler(hub *Hub, jwtMgr *jwt.Manager, logger *zap.Logger) fiber.Handler {
	return websocket.New(func(c *websocket.Conn) {
		// Token query parametr'dan
		token := c.Query("token")
		if token == "" {
			logger.Debug("ws: token yo'q")
			c.Close()
			return
		}

		claims, err := jwtMgr.ParseToken(token)
		if err != nil {
			logger.Debug("ws: yaroqsiz token", zap.Error(err))
			c.Close()
			return
		}

		client := &Client{
			UserID: claims.UserID,
			Send:   make(chan []byte, 64),
		}
		hub.Register(client)
		defer hub.Unregister(client)

		// Sync: kirgan zahoti welcome event
		welcome, _ := jsonMarshal(Event{
			Event: "connected",
			Data:  map[string]any{"user_id": claims.UserID},
		})
		client.Send <- welcome

		// Writer goroutine
		done := make(chan struct{})
		go func() {
			defer close(done)
			for msg := range client.Send {
				if err := c.WriteMessage(websocket.TextMessage, msg); err != nil {
					return
				}
			}
		}()

		// Reader loop (klientdan kelgan eventlarni o'qish — typing, ping va h.k.)
		for {
			_, _, err := c.ReadMessage()
			if err != nil {
				break
			}
			// Hozircha mijoz eventlarini e'tibordan chiqaramiz
		}

		<-done
		_ = uuid.Nil
	})
}

// jsonMarshal — kichik wrapper (importni minimal qilish uchun).
func jsonMarshal(v any) ([]byte, error) {
	return jsonMarshalFn(v)
}
