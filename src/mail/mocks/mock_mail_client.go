package mocks

import "context"

type MockMailClient struct {
	SendEmailFunc func(ctx context.Context, recipientEmail, recipientName, subject, plainContent string) error
}

func (m *MockMailClient) SendEmail(ctx context.Context, recipientEmail, recipientName, subject, plainContent string) error {
	if m.SendEmailFunc != nil {
		return m.SendEmailFunc(ctx, recipientEmail, recipientName, subject, plainContent)
	}
	return nil
}
