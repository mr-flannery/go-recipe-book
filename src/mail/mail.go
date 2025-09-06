package mail

import (
	"context"
	"fmt"

	"github.com/maileroo/maileroo-go-sdk/maileroo"
	"github.com/mr-flannery/go-recipe-book/src/config"
)

var (
	client *maileroo.Client
)

func getClient() *maileroo.Client {
	if client == nil {
		var err error

		conf := config.GetConfig()

		client, err = maileroo.NewClient(conf.Mail.ApiKey, 30)
		if err != nil {
			panic("Failed to create Maileroo client: " + err.Error())
		}
	}
	return client
}

func sendEmail(recipientEmailAddress string, recipientName string, subject string, plainContent string) error {
	client := getClient()

	config := config.GetConfig()

	// TODO: learn what exactly a context is in Go
	_, err := client.SendBasicEmail(context.Background(), maileroo.BasicEmailData{
		From: maileroo.NewEmail("recipe-book@"+config.Mail.Domain, "Recipe Book"),
		To: []maileroo.EmailAddress{
			maileroo.NewEmail(recipientEmailAddress, recipientName),
		},
		Subject: subject,
		Plain:   &plainContent,
	})

	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}
	return nil
}

// SendNewRegistrationNotification sends an email to admin when a new registration is pending
func SendNewRegistrationNotification(adminEmail, adminName, username, userEmail, approvalURL string) error {
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

	return sendEmail(adminEmail, adminName, subject, content)
}

// SendRegistrationApprovedNotification sends an email to user when their registration is approved
func SendRegistrationApprovedNotification(userEmail, username string) error {
	subject := "Registration Approved - Recipe Book"
	content := fmt.Sprintf(`Hello %s,

Great news! Your registration request for the Recipe Book application has been approved.

You can now log in to your account and start using the application.

Best regards,
Recipe Book Team`, username)

	return sendEmail(userEmail, username, subject, content)
}
