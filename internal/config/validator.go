package config

import (
	"fmt"
)

// Validate 驗證配置
func (c *Config) Validate() error {
	if c.Server.Port == "" {
		return fmt.Errorf("server.port is required")
	}

	if c.Database.URI == "" {
		return fmt.Errorf("database.uri is required")
	}

	if c.Database.Database == "" {
		return fmt.Errorf("database.database is required")
	}

	if c.Session.Secret == "" || c.Session.Secret == "change-me-in-production" {
		return fmt.Errorf("session.secret must be set to a secure value")
	}

	return nil
}

// ValidateMail 驗證郵件配置（可選）
func (c *Config) ValidateMail() error {
	if c.Mail.SMTPHost == "" {
		return fmt.Errorf("mail.smtp_host is required if mail is enabled")
	}

	if c.Mail.SMTPPort == 0 {
		return fmt.Errorf("mail.smtp_port is required if mail is enabled")
	}

	if c.Mail.From == "" {
		return fmt.Errorf("mail.from is required if mail is enabled")
	}

	return nil
}

