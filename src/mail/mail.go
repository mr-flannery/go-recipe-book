package mail

import (
	"context"
	"fmt"

	"github.com/maileroo/maileroo-go-sdk/maileroo"
)

type MailClient interface {
	SendEmail(recipientEmail, recipientName, subject, plainContent string) error
}

type mailerooClient struct {
	client *maileroo.Client
	domain string
}

func NewMailClient(apiKey, domain string) (MailClient, error) {
	client, err := maileroo.NewClient(apiKey, 30)
	if err != nil {
		return nil, fmt.Errorf("failed to create mail client: %w", err)
	}
	return &mailerooClient{client: client, domain: domain}, nil
}

func (m *mailerooClient) SendEmail(recipientEmail, recipientName, subject, plainContent string) error {
	_, err := m.client.SendBasicEmail(context.Background(), maileroo.BasicEmailData{
		From: maileroo.NewEmail("recipe-book@"+m.domain, "Recipe Book"),
		To: []maileroo.EmailAddress{
			maileroo.NewEmail(recipientEmail, recipientName),
		},
		Subject: subject,
		Plain:   &plainContent,
	})

	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}
	return nil
}

func SendNewRegistrationNotification(mc MailClient, adminEmail, adminName, username, userEmail, approvalURL string) error {
	subject := "New Registration Request - Recipe Book"
	content := fmt.Sprintf(`Hello %s,

A new user has requested to register for the Recipe Book application.

User Details:
- Username: %s
- Email: %s

Please review and approve or deny this registration request by visiting:
%s

Best regards,
Recipe Book System`, adminName, username, userEmail, approvalURL)

	return mc.SendEmail(adminEmail, adminName, subject, content)
}

func SendRegistrationApprovedNotification(mc MailClient, userEmail, username string) error {
	subject := "Registration Approved - Recipe Book"
	content := fmt.Sprintf(`Hello %s,

Great news! Your registration request for the Recipe Book application has been approved.

You can now log in to your account and start using the application.

Best regards,
Recipe Book Team`, username)

	return mc.SendEmail(userEmail, username, subject, content)
}

func SendPasswordResetEmail(mc MailClient, userEmail, username, resetURL string) error {
	subject := "Password Reset Request - Recipe Book"
	content := fmt.Sprintf(`Hello %s,

We received a request to reset your password for the Recipe Book application.

Click the link below to set a new password:
%s

This link will expire in 24 hours.

If you didn't request this, you can safely ignore this email.

Best regards,
Recipe Book`, username, resetURL)

	return mc.SendEmail(userEmail, username, subject, content)
}

type loggingMailClient struct{}

func NewLoggingMailClient() MailClient {
	return &loggingMailClient{}
}

func (l *loggingMailClient) SendEmail(recipientEmail, recipientName, subject, plainContent string) error {
	fmt.Printf("[DEV] Email not sent - To: %s <%s>, Subject: %s\n", recipientName, recipientEmail, subject)
	fmt.Printf("[DEV] Email body:\n%s\n", plainContent)
	return nil
}
