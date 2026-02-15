package mail

import (
	"errors"
	"strings"
	"testing"

	"github.com/mr-flannery/go-recipe-book/src/mail/mocks"
)

func TestSendNewRegistrationNotification_SendsCorrectEmailContent(t *testing.T) {
	var capturedEmail, capturedName, capturedSubject, capturedContent string
	mockClient := &mocks.MockMailClient{
		SendEmailFunc: func(recipientEmail, recipientName, subject, plainContent string) error {
			capturedEmail = recipientEmail
			capturedName = recipientName
			capturedSubject = subject
			capturedContent = plainContent
			return nil
		},
	}

	err := SendNewRegistrationNotification(mockClient, "admin@test.com", "Admin", "newuser", "new@test.com", "http://example.com/approve")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if capturedEmail != "admin@test.com" {
		t.Errorf("expected recipient email 'admin@test.com', got '%s'", capturedEmail)
	}

	if capturedName != "Admin" {
		t.Errorf("expected recipient name 'Admin', got '%s'", capturedName)
	}

	if capturedSubject != "New Registration Request - Recipe Book" {
		t.Errorf("expected subject 'New Registration Request - Recipe Book', got '%s'", capturedSubject)
	}

	if !strings.Contains(capturedContent, "newuser") {
		t.Error("expected content to contain username 'newuser'")
	}

	if !strings.Contains(capturedContent, "new@test.com") {
		t.Error("expected content to contain user email 'new@test.com'")
	}

	if !strings.Contains(capturedContent, "http://example.com/approve") {
		t.Error("expected content to contain approval URL")
	}
}

func TestSendNewRegistrationNotification_ReturnsErrorWhenSendFails(t *testing.T) {
	mockClient := &mocks.MockMailClient{
		SendEmailFunc: func(recipientEmail, recipientName, subject, plainContent string) error {
			return errors.New("send failed")
		},
	}

	err := SendNewRegistrationNotification(mockClient, "admin@test.com", "Admin", "newuser", "new@test.com", "http://example.com/approve")

	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestSendRegistrationApprovedNotification_SendsCorrectEmailContent(t *testing.T) {
	var capturedEmail, capturedName, capturedSubject, capturedContent string
	mockClient := &mocks.MockMailClient{
		SendEmailFunc: func(recipientEmail, recipientName, subject, plainContent string) error {
			capturedEmail = recipientEmail
			capturedName = recipientName
			capturedSubject = subject
			capturedContent = plainContent
			return nil
		},
	}

	err := SendRegistrationApprovedNotification(mockClient, "user@test.com", "testuser")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if capturedEmail != "user@test.com" {
		t.Errorf("expected recipient email 'user@test.com', got '%s'", capturedEmail)
	}

	if capturedName != "testuser" {
		t.Errorf("expected recipient name 'testuser', got '%s'", capturedName)
	}

	if capturedSubject != "Registration Approved - Recipe Book" {
		t.Errorf("expected subject 'Registration Approved - Recipe Book', got '%s'", capturedSubject)
	}

	if !strings.Contains(capturedContent, "testuser") {
		t.Error("expected content to contain username 'testuser'")
	}

	if !strings.Contains(capturedContent, "approved") {
		t.Error("expected content to mention approval")
	}
}

func TestSendRegistrationApprovedNotification_ReturnsErrorWhenSendFails(t *testing.T) {
	mockClient := &mocks.MockMailClient{
		SendEmailFunc: func(recipientEmail, recipientName, subject, plainContent string) error {
			return errors.New("send failed")
		},
	}

	err := SendRegistrationApprovedNotification(mockClient, "user@test.com", "testuser")

	if err == nil {
		t.Error("expected error, got nil")
	}
}
