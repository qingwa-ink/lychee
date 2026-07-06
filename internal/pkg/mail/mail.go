// Package mail 提供邮件发送抽象。dev 模式打印日志，prod 模式走 SMTP。
package mail

import (
	"fmt"
	"log"
	"net/smtp"
)

// Config 邮件配置。
type Config struct {
	Host string
	Port int
	User string
	Pass string
}

// Mailer 邮件发送接口。
type Mailer interface {
	Send(to, subject, body string) error
}

// NewMailer 根据配置选择实现：未配置 Host 时使用日志实现（便于本地开发联调）。
func NewMailer(cfg Config) Mailer {
	if cfg.Host == "" {
		return &logMailer{}
	}
	return &smtpMailer{host: cfg.Host, port: cfg.Port, user: cfg.User, pass: cfg.Pass, from: cfg.User}
}

// logMailer 仅打印日志，不真正发送。
type logMailer struct{}

func (m *logMailer) Send(to, subject, body string) error {
	log.Printf("[MAIL] to=%s | subject=%s | body=%s", to, subject, body)
	return nil
}

// smtpMailer 通过 SMTP 发送。
type smtpMailer struct {
	host string
	port int
	user string
	pass string
	from string
}

func (m *smtpMailer) Send(to, subject, body string) error {
	addr := fmt.Sprintf("%s:%d", m.host, m.port)
	auth := smtp.PlainAuth("", m.user, m.pass, m.host)
	msg := []byte(fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s",
		m.from, to, subject, body,
	))
	return smtp.SendMail(addr, auth, m.from, []string{to}, msg)
}
