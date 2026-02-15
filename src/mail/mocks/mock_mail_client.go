package mocks

type MockMailClient struct {
	SendEmailFunc func(recipientEmail, recipientName, subject, plainContent string) error
}

func (m *MockMailClient) SendEmail(recipientEmail, recipientName, subject, plainContent string) error {
	if m.SendEmailFunc != nil {
		return m.SendEmailFunc(recipientEmail, recipientName, subject, plainContent)
	}
	return nil
}
