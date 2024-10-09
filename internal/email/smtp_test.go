package email

import (
	"context"
	"errors"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ses"
	"gmf_message_processor/connection"
	"net/smtp"
	"os"
	"testing"
)

// MockSecretService es un mock de connection.SecretService.
type MockSecretService struct {
	GetSecretFunc func(secretName string) (*connection.SecretData, error)
}

func (m *MockSecretService) GetSecret(secretName string) (*connection.SecretData, error) {
	return m.GetSecretFunc(secretName)
}

// MockSMTP es un mock de smtp.SendMail
type MockSMTP struct {
	SendMailFunc func(addr string, a smtp.Auth, from string, to []string, msg []byte) error
}

func (m *MockSMTP) SendMail(addr string, a smtp.Auth, from string, to []string, msg []byte) error {
	return m.SendMailFunc(addr, a, from, to, msg)
}

// Test para inicializar SMTPEmailService con credenciales correctas
func TestNewSMTPEmailService_Success(t *testing.T) {
	secretData := &connection.SecretData{
		Username: "smtp_user",
		Password: "smtp_pass",
	}

	secretService := &MockSecretService{
		GetSecretFunc: func(secretName string) (*connection.SecretData, error) {
			return secretData, nil
		},
	}

	// Simulamos las variables de entorno necesarias
	os.Setenv("SMTP_SERVER", "smtp.example.com")
	os.Setenv("SMTP_PORT", "587")
	os.Setenv("SECRETS_SMTP", "smtp-secrets")

	smtpService, err := NewSMTPEmailService(secretService)
	if err != nil {
		t.Fatalf("No se esperaba error, pero se obtuvo: %v", err)
	}

	if smtpService.username != secretData.Username || smtpService.password != secretData.Password {
		t.Errorf("Las credenciales no coinciden. Esperado: %s/%s, Obtenido: %s/%s",
			secretData.Username, secretData.Password, smtpService.username, smtpService.password)
	}
}

// Test para error al obtener credenciales de Secrets Manager
func TestNewSMTPEmailService_Error(t *testing.T) {
	secretService := &MockSecretService{
		GetSecretFunc: func(secretName string) (*connection.SecretData, error) {
			return nil, errors.New("error obteniendo credenciales")
		},
	}

	_, err := NewSMTPEmailService(secretService)
	if err == nil {
		t.Error("Se esperaba un error, pero no se obtuvo ninguno")
	}
}

// Test para enviar correo usando SMTP con éxito
func TestSMTPEmailService_SendEmail_Success(t *testing.T) {
	mockSMTP := &MockSMTP{
		SendMailFunc: func(addr string, a smtp.Auth, from string, to []string, msg []byte) error {
			// Simulamos un envío exitoso
			return nil
		},
	}

	smtpService := &SMTPEmailService{
		server:   "smtp.example.com",
		port:     "587",
		username: "smtp_user",
		password: "smtp_pass",
		sendMail: mockSMTP.SendMail,
	}

	err := smtpService.SendEmail("test@example.com", "recipient@example.com", "Test Subject", "<h1>Test Body</h1>")
	if err != nil {
		t.Errorf("No se esperaba error, pero se obtuvo: %v", err)
	}
}

// Test para error al enviar correo usando SMTP
func TestSMTPEmailService_SendEmail_Error(t *testing.T) {
	mockSMTP := &MockSMTP{
		SendMailFunc: func(addr string, a smtp.Auth, from string, to []string, msg []byte) error {
			// Simulamos un error durante el envío
			return errors.New("error enviando correo")
		},
	}

	smtpService := &SMTPEmailService{
		server:   "smtp.example.com",
		port:     "587",
		username: "smtp_user",
		password: "smtp_pass",
		sendMail: mockSMTP.SendMail,
	}

	err := smtpService.SendEmail("test@example.com", "recipient@example.com", "Test Subject", "<h1>Test Body</h1>")
	if err == nil {
		t.Error("Se esperaba un error, pero no se obtuvo ninguno")
	}
}

/*
****************************************************************************************************************
 */

// MockSESClient es un mock de SESClientInterface
type MockSESClient struct {
	SendEmailFunc func(ctx context.Context, input *ses.SendEmailInput, opts ...func(*ses.Options)) (*ses.SendEmailOutput, error)
}

func (m *MockSESClient) SendEmail(ctx context.Context, input *ses.SendEmailInput, opts ...func(*ses.Options)) (*ses.SendEmailOutput, error) {
	return m.SendEmailFunc(ctx, input, opts...)
}

// MockAWSConfigLoader es un mock de AWSConfigLoader
type MockAWSConfigLoader struct {
	LoadConfigFunc func(ctx context.Context) (aws.Config, error)
}

func (m *MockAWSConfigLoader) LoadConfig(ctx context.Context) (aws.Config, error) {
	return m.LoadConfigFunc(ctx)
}

// Test para inicializar SESEmailService con una configuración de AWS válida
func TestNewSESEmailService_Success(t *testing.T) {
	mockConfigLoader := &MockAWSConfigLoader{
		LoadConfigFunc: func(ctx context.Context) (aws.Config, error) {
			return aws.Config{}, nil // Simulamos una configuración exitosa
		},
	}

	emailService := NewSESEmailService(mockConfigLoader)
	if emailService == nil {
		t.Error("Se esperaba una instancia de SESEmailService, pero se obtuvo nil")
	}
}

// Test para error al cargar la configuración de AWS
func TestNewSESEmailService_Error(t *testing.T) {
	mockConfigLoader := &MockAWSConfigLoader{
		LoadConfigFunc: func(ctx context.Context) (aws.Config, error) {
			return aws.Config{}, errors.New("error cargando configuración")
		},
	}

	emailService := NewSESEmailService(mockConfigLoader)
	if emailService != nil {
		t.Error("Se esperaba que NewSESEmailService devolviera nil debido a un error de configuración")
	}
}

// Test para envío exitoso de correo usando SES
func TestSESEmailService_SendEmail_Success(t *testing.T) {
	mockSESClient := &MockSESClient{
		SendEmailFunc: func(ctx context.Context, input *ses.SendEmailInput, opts ...func(*ses.Options)) (*ses.SendEmailOutput, error) {
			// Simulamos un envío exitoso
			return &ses.SendEmailOutput{}, nil
		},
	}

	emailService := &SESEmailService{
		Client: mockSESClient,
	}

	err := emailService.SendEmail("test@example.com", "recipient@example.com", "Test Subject", "<h1>Test Body</h1>")
	if err != nil {
		t.Errorf("No se esperaba error, pero se obtuvo: %v", err)
	}
}

// Test para error al enviar correo usando SES
func TestSESEmailService_SendEmail_Error(t *testing.T) {
	mockSESClient := &MockSESClient{
		SendEmailFunc: func(ctx context.Context, input *ses.SendEmailInput, opts ...func(*ses.Options)) (*ses.SendEmailOutput, error) {
			// Simulamos un error durante el envío
			return nil, errors.New("error enviando correo")
		},
	}

	emailService := &SESEmailService{
		Client: mockSESClient,
	}

	err := emailService.SendEmail("test@example.com", "recipient@example.com", "Test Subject", "<h1>Test Body</h1>")
	if err == nil {
		t.Error("Se esperaba un error, pero no se obtuvo ninguno")
	}
}

// Test para correos inválidos
func TestSESEmailService_InvalidEmail(t *testing.T) {
	emailService := &SESEmailService{
		Client: &MockSESClient{}, // No es necesario mockear SES aquí
	}

	// Test para remitente inválido
	err := emailService.SendEmail("invalid-email", "recipient@example.com", "Test Subject", "<h1>Test Body</h1>")
	if err == nil || err.Error() != "correo remitente inválido: mail: missing '@' or angle-addr" {
		t.Errorf("Se esperaba un error por correo remitente inválido, pero se obtuvo: %v", err)
	}

	// Test para destinatario inválido
	err = emailService.SendEmail("test@example.com", "invalid-email", "Test Subject", "<h1>Test Body</h1>")
	if err == nil || err.Error() != "correo destinatario inválido: mail: missing '@' or angle-addr" {
		t.Errorf("Se esperaba un error por correo destinatario inválido, pero se obtuvo: %v", err)
	}
}
