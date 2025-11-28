package mail

import (
	"fmt"
	"net/smtp"
)

// Config 郵件配置
type Config struct {
	SMTPHost     string
	SMTPPort     int
	SMTPUser     string
	SMTPPassword string
	From         string
}

// Service 郵件服務
type Service struct {
	config Config
}

// NewService 建立郵件服務
func NewService(config Config) *Service {
	return &Service{config: config}
}

// Send 發送郵件
func (s *Service) Send(to, subject, body string) error {
	auth := smtp.PlainAuth("", s.config.SMTPUser, s.config.SMTPPassword, s.config.SMTPHost)

	msg := []byte(fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: %s\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n%s",
		s.config.From, to, subject, body,
	))

	addr := fmt.Sprintf("%s:%d", s.config.SMTPHost, s.config.SMTPPort)
	return smtp.SendMail(addr, auth, s.config.From, []string{to}, msg)
}

// SendPasswordReset 發送重設密碼郵件
func (s *Service) SendPasswordReset(to, accessKey, baseURL string) error {
	resetURL := fmt.Sprintf("%s/ResetPassword/%s?email=%s", baseURL, accessKey, to)
	body := fmt.Sprintf(`
		<p>Hi,</p>
		<p>You received this email notification because you forgot password for HiggsTV account. 
		If you want to reset your password for HiggsTV account, please click the following link:</p>
		<p><a href="%s">%s</a></p>
		<p>If you don't reset your password, just delete this email.</p>
		<p>Cheers,<br>HiggsTV</p>
	`, resetURL, resetURL)

	return s.Send(to, "Reset Password for Your HiggsTV Account", body)
}

