// Package mail 提供邮件发送抽象。dev 模式打印日志，prod 模式走 SMTP。
package mail

import (
	"crypto/tls"
	"errors"
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
	msg := []byte(fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s",
		m.from, to, subject, body,
	))
	// 465 是隐式 TLS（SMTPS），标准库 net/smtp.SendMail 只支持明文+STARTTLS（587/25），
	// 直接用于 465 会在读取 SMTP 握手时与对端 TLS 握手死锁。故 465 走专门的 SMTPS 通道。
	if m.port == 465 {
		return sendMailSMTPS(addr, m.host, m.user, m.pass, m.from, []string{to}, msg)
	}
	auth := smtp.PlainAuth("", m.user, m.pass, m.host)
	return smtp.SendMail(addr, auth, m.from, []string{to}, msg)
}

// sendMailSMTPS 走隐式 TLS（465）：先 tls.Dial 建立 TLS 连接，再在其上跑 SMTP。
// 复用 net/smtp.Client，但因连接已加密，用自带 plainAuth（不做 client.tls 校验）做认证。
func sendMailSMTPS(addr, host, user, pass, from string, to []string, msg []byte) error {
	conn, err := tls.Dial("tcp", addr, &tls.Config{ServerName: host, MinVersion: tls.VersionTLS12})
	if err != nil {
		return fmt.Errorf("smtps dial %s: %w", addr, err)
	}

	c, err := smtp.NewClient(conn, host)
	if err != nil {
		conn.Close()
		return fmt.Errorf("smtps new client: %w", err)
	}
	defer c.Quit()

	if err := c.Auth(&plainAuth{username: user, password: pass}); err != nil {
		return fmt.Errorf("smtps auth: %w", err)
	}
	if err := c.Mail(from); err != nil {
		return fmt.Errorf("smtps mail from: %w", err)
	}
	for _, rcpt := range to {
		if err := c.Rcpt(rcpt); err != nil {
			return fmt.Errorf("smtps rcpt to %s: %w", rcpt, err)
		}
	}
	w, err := c.Data()
	if err != nil {
		return fmt.Errorf("smtps data: %w", err)
	}
	if _, err := w.Write(msg); err != nil {
		return fmt.Errorf("smtps write body: %w", err)
	}
	return w.Close()
}

// plainAuth 实现 AUTH PLAIN，但不做 net/smtp.PlainAuth 的「未加密连接拒绝」校验——
// sendMailSMTPS 已在 TLS 隧道内，连接必然加密；标准库 PlainAuth 因 smtp.Client.tls
// 标志未被设置（隐式 TLS 未走 STARTTLS 分支）会误判为未加密而拒绝，故自行实现。
type plainAuth struct {
	username, password string
}

func (a *plainAuth) Start(_ *smtp.ServerInfo) (string, []byte, error) {
	resp := []byte("\x00" + a.username + "\x00" + a.password)
	return "PLAIN", resp, nil
}

func (a *plainAuth) Next(_ []byte, more bool) ([]byte, error) {
	if more {
		return nil, errors.New("unexpected server challenge")
	}
	return nil, nil
}
