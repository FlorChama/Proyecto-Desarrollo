package utils

import (
	"encoding/base64"
	"fmt"

	qrcode "github.com/skip2/go-qrcode"
)

// GenerateQRCode genera un código QR en base64 a partir de un string único.
// El string se guarda como propiedad del ticket en la base de datos.
func GenerateQRCode(data string) (string, error) {
	png, err := qrcode.Encode(data, qrcode.Medium, 256)
	if err != nil {
		return "", fmt.Errorf("error generando QR: %w", err)
	}
	encoded := base64.StdEncoding.EncodeToString(png)
	return "data:image/png;base64," + encoded, nil
}

// GenerateTicketQRData genera el string único para el QR de un ticket.
func GenerateTicketQRData(ticketID, userID, eventID uint) string {
	return fmt.Sprintf("TICKET-%d-USER-%d-EVENT-%d", ticketID, userID, eventID)
}
