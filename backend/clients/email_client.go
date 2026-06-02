package clients

import (
	"fmt"
	"os"
	"strconv"

	"gopkg.in/gomail.v2"
)

type EmailClient struct {
	dialer *gomail.Dialer
	from   string
}

func NewEmailClient() *EmailClient {
	host := os.Getenv("SMTP_HOST")
	portStr := os.Getenv("SMTP_PORT")
	user := os.Getenv("SMTP_USER")
	pass := os.Getenv("SMTP_PASS")

	port, err := strconv.Atoi(portStr)
	if err != nil {
		port = 587
	}

	dialer := gomail.NewDialer(host, port, user, pass)
	return &EmailClient{dialer: dialer, from: user}
}

func (e *EmailClient) SendPurchaseConfirmation(toEmail, toName, eventTitle, qrCode string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", e.from)
	m.SetHeader("To", toEmail)
	m.SetHeader("Subject", fmt.Sprintf("¡Compra exitosa! - %s", eventTitle))

	body := fmt.Sprintf(`
		<h2>¡Hola %s!</h2>
		<p>Tu compra para el evento <strong>%s</strong> fue procesada exitosamente.</p>
		<p>Tu código QR de acceso:</p>
		<img src="%s" alt="QR Code" style="width:200px;height:200px;" />
		<p>Presentá este QR en la entrada del evento.</p>
	`, toName, eventTitle, qrCode)

	m.SetBody("text/html", body)

	if err := e.dialer.DialAndSend(m); err != nil {
		return fmt.Errorf("error enviando email de compra: %w", err)
	}
	return nil
}

func (e *EmailClient) SendTransferNotification(toEmail, toName, fromName, eventTitle, qrCode string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", e.from)
	m.SetHeader("To", toEmail)
	m.SetHeader("Subject", fmt.Sprintf("Te transfirieron una entrada - %s", eventTitle))

	body := fmt.Sprintf(`
		<h2>¡Hola %s!</h2>
		<p><strong>%s</strong> te transfirió una entrada para el evento <strong>%s</strong>.</p>
		<p>Tu código QR de acceso:</p>
		<img src="%s" alt="QR Code" style="width:200px;height:200px;" />
		<p>Presentá este QR en la entrada del evento.</p>
	`, toName, fromName, eventTitle, qrCode)

	m.SetBody("text/html", body)

	if err := e.dialer.DialAndSend(m); err != nil {
		return fmt.Errorf("error enviando email de transferencia: %w", err)
	}
	return nil
}
