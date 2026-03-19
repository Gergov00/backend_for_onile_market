package services

import (
	"crypto/rand"
	"crypto/tls"
	"fmt"
	"log"
	"math/big"
	"net/smtp"
	"os"

	"github.com/Notificore/notificore-go/src/notificorerest"
)

// SMSProvider интерфейс для отправки SMS
type SMSProvider interface {
	SendCode(phone, code string) error
}

// EmailProvider интерфейс для отправки Email
type EmailProvider interface {
	SendEmail(email, subject, message string) error
}

// MockSMS реализация для разработки (выводит код в консоль)
type MockSMS struct{}

func (m *MockSMS) SendCode(phone, code string) error {
	log.Printf("[MOCK SMS] Отправка на %s: Ваш код подтверждения: %s", phone, code)
	printSMSBox(phone, code)
	return nil
}

// MockEmail реализация для разработки
type MockEmail struct{}

func (m *MockEmail) SendEmail(email, subject, message string) error {
	log.Printf("[MOCK EMAIL] Отправка на %s | Тема: %s | Сообщение: %s", email, subject, message)
	fmt.Printf("\n"+
		"╔══════════════════════════════════════════╗\n"+
		"║          ПОДТВЕРЖДЕНИЕ EMAIL            ║\n"+
		"╠══════════════════════════════════════════╣\n"+
		"║  Email:   %-28s  ║\n"+
		"║  Тема:    %-28s  ║\n"+
		"║  Текст:   %-28s  ║\n"+
		"╚══════════════════════════════════════════╝\n\n",
		email, subject, message)
	return nil
}

// loginAuth реализует smtp.Auth для механизма LOGIN
type loginAuth struct {
	username, password string
}

func (a *loginAuth) Start(server *smtp.ServerInfo) (string, []byte, error) {
	return "LOGIN", []byte{}, nil
}

func (a *loginAuth) Next(fromServer []byte, more bool) ([]byte, error) {
	if more {
		switch string(fromServer) {
		case "Username:":
			return []byte(a.username), nil
		case "Password:":
			return []byte(a.password), nil
		default:
			return nil, fmt.Errorf("unexpected server challenge: %s", fromServer)
		}
	}
	return nil, nil
}

// SMTPEmail реализация email-провайдера через SMTP
type SMTPEmail struct {
	Host     string
	Port     string
	User     string
	Password string
	From     string
}

// SendEmail отправляет письмо через SMTP с SSL (порт 465)
func (s *SMTPEmail) SendEmail(to, subject, message string) error {
	log.Printf("[SMTP] Отправка на %s | Тема: %s", to, subject)

	// HTML-версия письма для красивого отображения
	htmlBody := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head><meta charset="UTF-8"></head>
<body style="font-family: Arial, sans-serif; background: #f5f5f5; padding: 40px 0;">
  <div style="max-width: 480px; margin: 0 auto; background: white; border-radius: 12px; overflow: hidden; box-shadow: 0 4px 20px rgba(0,0,0,0.1);">
    <div style="background: linear-gradient(135deg, #C41E3A, #8B0000); padding: 30px; text-align: center;">
      <h1 style="color: white; margin: 0; font-size: 24px;">RUSSIAN-ARMO</h1>
      <p style="color: rgba(255,255,255,0.8); margin: 8px 0 0; font-size: 14px;">Военное снаряжение</p>
    </div>
    <div style="padding: 32px; text-align: center;">
      <h2 style="color: #333; margin: 0 0 16px;">%s</h2>
      <p style="color: #666; line-height: 1.6;">%s</p>
      <p style="color: #999; font-size: 12px; margin-top: 24px;">Если вы не запрашивали этот код, проигнорируйте письмо.</p>
    </div>
    <div style="background: #f9f9f9; padding: 16px; text-align: center; border-top: 1px solid #eee;">
      <p style="color: #aaa; font-size: 12px; margin: 0;">© RUSSIAN-ARMO Store | russian-armo.org</p>
    </div>
  </div>
</body>
</html>`, subject, message)

	// Формируем MIME-сообщение
	headers := fmt.Sprintf("From: RUSSIAN-ARMO <%s>\r\n"+
		"To: %s\r\n"+
		"Subject: %s\r\n"+
		"MIME-Version: 1.0\r\n"+
		"Content-Type: text/html; charset=UTF-8\r\n"+
		"\r\n", s.From, to, subject)

	msg := []byte(headers + htmlBody)

	// SSL-подключение (порт 465)
	addr := s.Host + ":" + s.Port

	tlsConfig := &tls.Config{
		ServerName: s.Host,
	}

	conn, err := tls.Dial("tcp", addr, tlsConfig)
	if err != nil {
		log.Printf("[SMTP] Ошибка TLS подключения: %v", err)
		return fmt.Errorf("ошибка подключения к SMTP: %v", err)
	}

	client, err := smtp.NewClient(conn, s.Host)
	if err != nil {
		log.Printf("[SMTP] Ошибка создания клиента: %v", err)
		return fmt.Errorf("ошибка SMTP клиента: %v", err)
	}
	defer client.Close()

	// Аутентификация (LOGIN — Beget не поддерживает PLAIN)
	auth := &loginAuth{username: s.User, password: s.Password}
	if err := client.Auth(auth); err != nil {
		log.Printf("[SMTP] Ошибка аутентификации: %v", err)
		return fmt.Errorf("ошибка аутентификации SMTP: %v", err)
	}

	// Отправка
	if err := client.Mail(s.From); err != nil {
		return fmt.Errorf("ошибка MAIL FROM: %v", err)
	}
	if err := client.Rcpt(to); err != nil {
		return fmt.Errorf("ошибка RCPT TO: %v", err)
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("ошибка DATA: %v", err)
	}
	_, err = w.Write(msg)
	if err != nil {
		return fmt.Errorf("ошибка записи сообщения: %v", err)
	}
	w.Close()

	client.Quit()

	log.Printf("[SMTP] Email отправлен на %s", to)
	return nil
}

// RealSMS реализация через Notificore
type RealSMS struct {
	APIKey string
}

// SendCode отправляет SMS через Notificore API и дублирует код в консоль
func (s *RealSMS) SendCode(phone, code string) error {
	msg := fmt.Sprintf("Ваш код подтверждения: %s", code)

	// Всегда выводим в консоль для отладки
	printSMSBox(phone, code)

	client := notificorerest.NewSmsClient("https://api.notificore.ru/rest", s.APIKey)

	// Генерируем случайный reference
	reference := "ext_" + GenerateCode() + GenerateCode()
	phoneObj := &notificorerest.SmsPhone{Msisdn: phone, Reference: reference}

	originator := os.Getenv("SMS_ORIGINATOR")
	if originator == "" {
		originator = "Info" // Замените на нужное имя отправителя
	}

	resp := client.CreateSms("1", nil, originator, msg, phoneObj)

	log.Printf("[NOTIFICORE] Ответ: %+v", resp)

	log.Printf("[NOTIFICORE] SMS отправлено на %s", phone)
	return nil
}

// printSMSBox выводит красивый блок с кодом верификации в консоль
func printSMSBox(phone, code string) {
	fmt.Printf("\n"+
		"╔══════════════════════════════════════════╗\n"+
		"║           ПОДТВЕРЖДЕНИЕ SMS             ║\n"+
		"╠══════════════════════════════════════════╣\n"+
		"║  Телефон: %-28s  ║\n"+
		"║  Код:     %-28s  ║\n"+
		"╚══════════════════════════════════════════╝\n\n",
		phone, code)
}

// GenerateCode генерирует случайный 6-значный код
func GenerateCode() string {
	max := big.NewInt(999999)
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "123456"
	}
	return fmt.Sprintf("%06d", n.Int64())
}

// GetSMSProvider возвращает провайдер SMS
// Если установлена переменная NOTIFICORE_API_KEY — использует Notificore
// Иначе — MockSMS (вывод в консоль)
func GetSMSProvider() SMSProvider {
	apiKey := os.Getenv("NOTIFICORE_API_KEY")
	if apiKey != "" {
		log.Println("SMS провайдер: Notificore (реальная отправка)")
		return &RealSMS{APIKey: apiKey}
	}
	log.Println("SMS провайдер: MockSMS (вывод в консоль)")
	return &MockSMS{}
}

// GetEmailProvider возвращает SMTP провайдер или MockEmail
func GetEmailProvider() EmailProvider {
	host := os.Getenv("SMTP_HOST")
	if host != "" {
		port := os.Getenv("SMTP_PORT")
		if port == "" {
			port = "465"
		}
		log.Printf("Email провайдер: SMTP (%s:%s)", host, port)
		return &SMTPEmail{
			Host:     host,
			Port:     port,
			User:     os.Getenv("SMTP_USER"),
			Password: os.Getenv("SMTP_PASSWORD"),
			From:     os.Getenv("SMTP_FROM"),
		}
	}
	log.Println("Email провайдер: MockEmail (вывод в консоль)")
	return &MockEmail{}
}
