package auth

import (
	"crypto/tls"
	"fmt"
	"net/smtp"
	"strings"

	"github.com/a5c-ai/hub/internal/config"
)

type SMTPEmailService struct {
	host     string
	port     string
	username string
	password string
	from     string
	useTLS   bool
}

func NewSMTPEmailService(cfg *config.Config) EmailService {
	return &SMTPEmailService{
		host:     cfg.SMTP.Host,
		port:     cfg.SMTP.Port,
		username: cfg.SMTP.Username,
		password: cfg.SMTP.Password,
		from:     cfg.SMTP.From,
		useTLS:   cfg.SMTP.UseTLS,
	}
}

func (s *SMTPEmailService) SendPasswordResetEmail(to, token string) error {
	subject := "Password Reset Request"
	resetURL := fmt.Sprintf("%s/reset-password?token=%s", getBaseURL(), token)
	
	body := fmt.Sprintf(`
		<html>
		<body>
			<h2>Password Reset Request</h2>
			<p>You have requested to reset your password. Click the link below to reset your password:</p>
			<p><a href="%s">Reset Password</a></p>
			<p>This link will expire in 1 hour.</p>
			<p>If you did not request this password reset, please ignore this email.</p>
		</body>
		</html>
	`, resetURL)
	
	return s.sendEmail(to, subject, body)
}

func (s *SMTPEmailService) SendEmailVerification(to, token string) error {
	subject := "Email Verification"
	verifyURL := fmt.Sprintf("%s/verify-email?token=%s", getBaseURL(), token)
	
	body := fmt.Sprintf(`
		<html>
		<body>
			<h2>Email Verification</h2>
			<p>Thank you for registering! Please verify your email address by clicking the link below:</p>
			<p><a href="%s">Verify Email</a></p>
			<p>This link will expire in 24 hours.</p>
		</body>
		</html>
	`, verifyURL)
	
	return s.sendEmail(to, subject, body)
}

func (s *SMTPEmailService) sendEmail(to, subject, body string) error {
	// If SMTP is not configured, log the email instead of using mock
	if s.host == "" {
		return s.logEmail(to, subject, body)
	}

	// Prepare message
	headers := make(map[string]string)
	headers["From"] = s.from
	headers["To"] = to
	headers["Subject"] = subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/html; charset=utf-8"

	message := ""
	for k, v := range headers {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + body

	// Connect to SMTP server
	addr := fmt.Sprintf("%s:%s", s.host, s.port)
	
	var auth smtp.Auth
	if s.username != "" && s.password != "" {
		auth = smtp.PlainAuth("", s.username, s.password, s.host)
	}

	var err error
	if s.useTLS {
		// TLS connection
		tlsConfig := &tls.Config{
			InsecureSkipVerify: false,
			ServerName:         s.host,
		}

		conn, err := tls.Dial("tcp", addr, tlsConfig)
		if err != nil {
			return fmt.Errorf("failed to connect to SMTP server: %w", err)
		}
		defer conn.Close()

		client, err := smtp.NewClient(conn, s.host)
		if err != nil {
			return fmt.Errorf("failed to create SMTP client: %w", err)
		}
		defer client.Quit()

		if auth != nil {
			if err = client.Auth(auth); err != nil {
				return fmt.Errorf("SMTP authentication failed: %w", err)
			}
		}

		if err = client.Mail(s.from); err != nil {
			return fmt.Errorf("failed to set sender: %w", err)
		}

		if err = client.Rcpt(to); err != nil {
			return fmt.Errorf("failed to set recipient: %w", err)
		}

		w, err := client.Data()
		if err != nil {
			return fmt.Errorf("failed to get data writer: %w", err)
		}

		_, err = w.Write([]byte(message))
		if err != nil {
			return fmt.Errorf("failed to write message: %w", err)
		}

		err = w.Close()
		if err != nil {
			return fmt.Errorf("failed to close data writer: %w", err)
		}
	} else {
		// Plain SMTP connection
		err = smtp.SendMail(addr, auth, s.from, []string{to}, []byte(message))
		if err != nil {
			return fmt.Errorf("failed to send email: %w", err)
		}
	}

	return nil
}

func getBaseURL() string {
	// TODO: Get this from configuration
	return "http://localhost:3000"
}

func extractTokenFromURL(body string) string {
	// Simple token extraction for mock fallback
	if idx := strings.Index(body, "token="); idx != -1 {
		start := idx + 6
		end := strings.Index(body[start:], "\"")
		if end != -1 {
			return body[start : start+end]
		}
	}
	return ""
}

// Enhanced email service with templates
type EmailTemplate struct {
	Subject  string
	HTMLBody string
	TextBody string
}

type TemplatedEmailService struct {
	smtpService EmailService
	templates   map[string]EmailTemplate
}

func NewTemplatedEmailService(smtpService EmailService) *TemplatedEmailService {
	templates := map[string]EmailTemplate{
		"password_reset": {
			Subject:  "Password Reset Request - A5C Hub",
			HTMLBody: getPasswordResetHTMLTemplate(),
			TextBody: getPasswordResetTextTemplate(),
		},
		"email_verification": {
			Subject:  "Verify Your Email - A5C Hub",
			HTMLBody: getEmailVerificationHTMLTemplate(),
			TextBody: getEmailVerificationTextTemplate(),
		},
		"mfa_setup": {
			Subject:  "Two-Factor Authentication Setup - A5C Hub",
			HTMLBody: getMFASetupHTMLTemplate(),
			TextBody: getMFASetupTextTemplate(),
		},
	}

	return &TemplatedEmailService{
		smtpService: smtpService,
		templates:   templates,
	}
}

func (s *TemplatedEmailService) SendPasswordResetEmail(to, token string) error {
	return s.smtpService.SendPasswordResetEmail(to, token)
}

func (s *TemplatedEmailService) SendEmailVerification(to, token string) error {
	return s.smtpService.SendEmailVerification(to, token)
}

// Email templates
func getPasswordResetHTMLTemplate() string {
	return `
<!DOCTYPE html>
<html>
<head>
	<meta charset="utf-8">
	<title>Password Reset Request</title>
	<style>
		body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
		.container { max-width: 600px; margin: 0 auto; padding: 20px; }
		.header { background-color: #007bff; color: white; padding: 20px; text-align: center; }
		.content { padding: 20px; background-color: #f8f9fa; }
		.button { display: inline-block; padding: 12px 24px; background-color: #007bff; color: white; text-decoration: none; border-radius: 4px; }
		.footer { padding: 20px; font-size: 14px; color: #666; text-align: center; }
	</style>
</head>
<body>
	<div class="container">
		<div class="header">
			<h1>A5C Hub</h1>
		</div>
		<div class="content">
			<h2>Password Reset Request</h2>
			<p>You have requested to reset your password for your A5C Hub account.</p>
			<p>Click the button below to reset your password:</p>
			<p><a href="{{.ResetURL}}" class="button">Reset Password</a></p>
			<p>This link will expire in 1 hour for security reasons.</p>
			<p>If you did not request this password reset, please ignore this email and your password will remain unchanged.</p>
		</div>
		<div class="footer">
			<p>This is an automated message from A5C Hub. Please do not reply to this email.</p>
		</div>
	</div>
</body>
</html>`
}

func getPasswordResetTextTemplate() string {
	return `A5C Hub - Password Reset Request

You have requested to reset your password for your A5C Hub account.

Reset your password: {{.ResetURL}}

This link will expire in 1 hour for security reasons.

If you did not request this password reset, please ignore this email and your password will remain unchanged.

This is an automated message from A5C Hub. Please do not reply to this email.`
}

func getEmailVerificationHTMLTemplate() string {
	return `
<!DOCTYPE html>
<html>
<head>
	<meta charset="utf-8">
	<title>Email Verification</title>
	<style>
		body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
		.container { max-width: 600px; margin: 0 auto; padding: 20px; }
		.header { background-color: #28a745; color: white; padding: 20px; text-align: center; }
		.content { padding: 20px; background-color: #f8f9fa; }
		.button { display: inline-block; padding: 12px 24px; background-color: #28a745; color: white; text-decoration: none; border-radius: 4px; }
		.footer { padding: 20px; font-size: 14px; color: #666; text-align: center; }
	</style>
</head>
<body>
	<div class="container">
		<div class="header">
			<h1>Welcome to A5C Hub!</h1>
		</div>
		<div class="content">
			<h2>Verify Your Email Address</h2>
			<p>Thank you for registering with A5C Hub! To complete your registration, please verify your email address.</p>
			<p>Click the button below to verify your email:</p>
			<p><a href="{{.VerifyURL}}" class="button">Verify Email</a></p>
			<p>This verification link will expire in 24 hours.</p>
		</div>
		<div class="footer">
			<p>This is an automated message from A5C Hub. Please do not reply to this email.</p>
		</div>
	</div>
</body>
</html>`
}

func getEmailVerificationTextTemplate() string {
	return `Welcome to A5C Hub!

Thank you for registering with A5C Hub! To complete your registration, please verify your email address.

Verify your email: {{.VerifyURL}}

This verification link will expire in 24 hours.

This is an automated message from A5C Hub. Please do not reply to this email.`
}

func getMFASetupHTMLTemplate() string {
	return `
<!DOCTYPE html>
<html>
<head>
	<meta charset="utf-8">
	<title>Two-Factor Authentication Setup</title>
	<style>
		body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
		.container { max-width: 600px; margin: 0 auto; padding: 20px; }
		.header { background-color: #ffc107; color: #333; padding: 20px; text-align: center; }
		.content { padding: 20px; background-color: #f8f9fa; }
		.footer { padding: 20px; font-size: 14px; color: #666; text-align: center; }
		.code { background-color: #e9ecef; padding: 10px; border-radius: 4px; font-family: monospace; text-align: center; font-size: 18px; }
	</style>
</head>
<body>
	<div class="container">
		<div class="header">
			<h1>Two-Factor Authentication Setup</h1>
		</div>
		<div class="content">
			<h2>Your Backup Codes</h2>
			<p>Please save these backup codes in a safe place. You can use them to access your account if you lose your primary authentication method.</p>
			<div class="code">
				{{range .BackupCodes}}
				{{.}}<br>
				{{end}}
			</div>
			<p><strong>Important:</strong> Each backup code can only be used once.</p>
		</div>
		<div class="footer">
			<p>This is an automated message from A5C Hub. Please do not reply to this email.</p>
		</div>
	</div>
</body>
</html>`
}

func getMFASetupTextTemplate() string {
	return `A5C Hub - Two-Factor Authentication Setup

Your Backup Codes:
{{range .BackupCodes}}
{{.}}
{{end}}

Please save these backup codes in a safe place. You can use them to access your account if you lose your primary authentication method.

Important: Each backup code can only be used once.

This is an automated message from A5C Hub. Please do not reply to this email.`
}

// logEmail logs email content when SMTP is not configured
func (s *SMTPEmailService) logEmail(to, subject, body string) error {
	// Log the email details to console/logs instead of sending
	fmt.Printf("=== EMAIL LOG (SMTP not configured) ===\n")
	fmt.Printf("To: %s\n", to)
	fmt.Printf("Subject: %s\n", subject)
	fmt.Printf("Body: %s\n", body)
	fmt.Printf("=====================================\n")
	
	// In production, you might want to:
	// 1. Log to a proper logging system
	// 2. Store in database for audit trail
	// 3. Send to a message queue for later processing
	// 4. Use a different notification method (SMS, in-app notification)
	
	return nil
}