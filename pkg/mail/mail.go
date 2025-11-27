package mail

import (
	"fmt"
	"net/smtp"
)

// MailConfig 郵件配置
type MailConfig struct {
	SMTPHost     string
	SMTPPort     int
	SMTPUser     string
	SMTPPassword string
	From         string
}

// MailService 郵件服務
type MailService struct {
	config MailConfig
}

// NewMailService 建立郵件服務
func NewMailService(config MailConfig) *MailService {
	return &MailService{config: config}
}

// Send 發送郵件
func (s *MailService) Send(to, subject, body string) error {
	auth := smtp.PlainAuth("", s.config.SMTPUser, s.config.SMTPPassword, s.config.SMTPHost)

	msg := []byte(fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: %s\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n%s",
		s.config.From, to, subject, body,
	))

	addr := fmt.Sprintf("%s:%d", s.config.SMTPHost, s.config.SMTPPort)
	return smtp.SendMail(addr, auth, s.config.From, []string{to}, msg)
}

// SendPasswordReset 發送重設密碼郵件
func (s *MailService) SendPasswordReset(to, accessKey, baseURL string) error {
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

