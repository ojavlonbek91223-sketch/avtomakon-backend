package service

import (
	"context"
	"fmt"

	"go.uber.org/zap"
)

// SMSProvider — SMS yuborish interfeysi.
// Production'da Eskiz yoki Playmobile implementatsiyasi ishlatiladi.
type SMSProvider interface {
	Send(ctx context.Context, phone, message string) error
}

// MockSMSProvider — dev rejimi uchun. SMS yuborish o'rniga logga yozadi.
type MockSMSProvider struct {
	logger *zap.Logger
}

func NewMockSMSProvider(logger *zap.Logger) *MockSMSProvider {
	return &MockSMSProvider{logger: logger}
}

func (m *MockSMSProvider) Send(ctx context.Context, phone, message string) error {
	m.logger.Info("[DEV SMS]",
		zap.String("phone", phone),
		zap.String("message", message),
	)
	fmt.Printf("\n📱 SMS → %s: %s\n\n", phone, message)
	return nil
}
