package code

import (
	"context"
	"crypto/rand"
	"crypto/tls"
	"errors"
	"fmt"
	"math/big"
	"net/smtp"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	mockCode = "123456"
	ttl      = 5 * time.Minute
)

// SMTPConfig holds outgoing mail server credentials.
// Port 465 uses implicit TLS; any other port (e.g. 587) uses STARTTLS.
type SMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
}

type Service struct {
	rdb  *redis.Client
	smtp *SMTPConfig // nil = dev mode, falls back to mockCode
}

func NewService(rdb *redis.Client, smtp *SMTPConfig) *Service {
	return &Service{rdb: rdb, smtp: smtp}
}

func key(target string) string {
	return fmt.Sprintf("im:vcode:%s", target)
}

// Send generates and delivers a verification code to target (email or phone).
// When SMTP is configured and target is an email address, a real random code is
// generated and emailed. Otherwise "123456" is stored for dev/phone use.
func (s *Service) Send(ctx context.Context, target string) error {
	code := mockCode
	if s.smtp != nil && strings.Contains(target, "@") {
		var err error
		code, err = randomCode()
		if err != nil {
			return fmt.Errorf("generate code: %w", err)
		}
		if err := s.sendEmail(target, code); err != nil {
			return fmt.Errorf("smtp: %w", err)
		}
	}
	return s.rdb.Set(ctx, key(target), code, ttl).Err()
}

// Verify checks the submitted code without consuming it.
// Call Consume after all dependent work succeeds to invalidate the code.
func (s *Service) Verify(ctx context.Context, target, submitted string) error {
	stored, err := s.rdb.Get(ctx, key(target)).Result()
	if errors.Is(err, redis.Nil) {
		return errors.New("code not found or expired — send code first")
	}
	if err != nil {
		return fmt.Errorf("redis: %w", err)
	}
	if stored != submitted {
		return errors.New("code mismatch")
	}
	return nil
}

// Consume deletes the code after successful use (one-time use).
func (s *Service) Consume(ctx context.Context, target string) {
	s.rdb.Del(ctx, key(target))
}

func randomCode() (string, error) {
	const digits = "0123456789"
	buf := make([]byte, 6)
	for i := range buf {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(digits))))
		if err != nil {
			return "", err
		}
		buf[i] = digits[n.Int64()]
	}
	return string(buf), nil
}

func (s *Service) sendEmail(to, code string) error {
	cfg := s.smtp
	subject := "验证码 / Verification Code"
	body := fmt.Sprintf("您的验证码是：%s，5 分钟内有效。\n\nYour verification code is: %s — valid for 5 minutes.", code, code)
	msg := []byte(fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: %s\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s",
		cfg.From, to, subject, body,
	))
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	auth := smtp.PlainAuth("", cfg.Username, cfg.Password, cfg.Host)

	if cfg.Port == 465 {
		// Implicit TLS (SSL) — used by some providers on port 465
		tlsCfg := &tls.Config{ServerName: cfg.Host}
		conn, err := tls.Dial("tcp", addr, tlsCfg)
		if err != nil {
			return err
		}
		defer conn.Close()
		c, err := smtp.NewClient(conn, cfg.Host)
		if err != nil {
			return err
		}
		defer c.Close()
		if err := c.Auth(auth); err != nil {
			return err
		}
		if err := c.Mail(cfg.From); err != nil {
			return err
		}
		if err := c.Rcpt(to); err != nil {
			return err
		}
		w, err := c.Data()
		if err != nil {
			return err
		}
		if _, err := w.Write(msg); err != nil {
			return err
		}
		return w.Close()
	}

	// STARTTLS — standard for port 587
	return smtp.SendMail(addr, auth, cfg.From, []string{to}, msg)
}
