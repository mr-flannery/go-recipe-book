package mail

import (
	"context"
	"fmt"

	"github.com/maileroo/maileroo-go-sdk/maileroo"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

var tracer = otel.Tracer("mail")

type MailClient interface {
	SendEmail(ctx context.Context, recipientEmail, recipientName, subject, plainContent string) error
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

func (m *mailerooClient) SendEmail(ctx context.Context, recipientEmail, recipientName, subject, plainContent string) error {
	ctx, span := tracer.Start(ctx, "mail.send",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			attribute.String("mail.recipient", recipientEmail),
			attribute.String("mail.subject", subject),
		),
	)
	defer span.End()

	_, err := m.client.SendBasicEmail(ctx, maileroo.BasicEmailData{
		From: maileroo.NewEmail("recipe-book@"+m.domain, "Recipe Book"),
		To: []maileroo.EmailAddress{
			maileroo.NewEmail(recipientEmail, recipientName),
		},
		Subject: subject,
		Plain:   &plainContent,
	})

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return fmt.Errorf("failed to send email: %w", err)
	}
	span.SetStatus(codes.Ok, "email sent")
	return nil
}

func SendNewRegistrationNotification(ctx context.Context, mc MailClient, adminEmail, adminName, username, userEmail, approvalURL string) error {
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

	return mc.SendEmail(ctx, adminEmail, adminName, subject, content)
}

func SendRegistrationApprovedNotification(ctx context.Context, mc MailClient, userEmail, username, loginURL string) error {
	subject := "Registration Approved - Recipe Book"
	content := fmt.Sprintf(`Hello %s,

Great news! Your registration request for the Recipe Book application has been approved.

You can now log in to your account:
%s

Best regards,
Recipe Book Team`, username, loginURL)

	return mc.SendEmail(ctx, userEmail, username, subject, content)
}

func SendPasswordResetEmail(ctx context.Context, mc MailClient, userEmail, username, resetURL string) error {
	subject := "Password Reset Request - Recipe Book"
	content := fmt.Sprintf(`Hello %s,

We received a request to reset your password for the Recipe Book application.

Click the link below to set a new password:
%s

This link will expire in 24 hours.

If you didn't request this, you can safely ignore this email.

Best regards,
Recipe Book`, username, resetURL)

	return mc.SendEmail(ctx, userEmail, username, subject, content)
}

type loggingMailClient struct{}

func NewLoggingMailClient() MailClient {
	return &loggingMailClient{}
}

func (l *loggingMailClient) SendEmail(ctx context.Context, recipientEmail, recipientName, subject, plainContent string) error {
	_, span := tracer.Start(ctx, "mail.send",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			attribute.String("mail.recipient", recipientEmail),
			attribute.String("mail.subject", subject),
		),
	)
	defer span.End()

	fmt.Printf("[DEV] Email not sent - To: %s <%s>, Subject: %s\n", recipientName, recipientEmail, subject)
	fmt.Printf("[DEV] Email body:\n%s\n", plainContent)
	span.SetStatus(codes.Ok, "email logged (dev mode)")
	return nil
}
