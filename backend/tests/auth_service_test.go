package tests

import (
	"testing"
	"ticketek-backend/domain"
	"ticketek-backend/utils"
)

func TestHashPassword(t *testing.T) {
	password := "miPassword123"
	hash := utils.HashPassword(password)

	if hash == password {
		t.Error("el hash no debe ser igual a la contraseña original")
	}
	if hash == "" {
		t.Error("el hash no debe estar vacío")
	}
}

func TestCheckPassword(t *testing.T) {
	password := "miPassword123"
	hash := utils.HashPassword(password)

	if !utils.CheckPassword(password, hash) {
		t.Error("CheckPassword debe retornar true para contraseña correcta")
	}
	if utils.CheckPassword("wrongPassword", hash) {
		t.Error("CheckPassword debe retornar false para contraseña incorrecta")
	}
}

func TestGenerateToken(t *testing.T) {
	token, err := utils.GenerateToken(1, "test@test.com", domain.RoleClient)
	if err != nil {
		t.Errorf("GenerateToken no debe retornar error: %v", err)
	}
	if token == "" {
		t.Error("el token no debe estar vacío")
	}
}

func TestValidateToken(t *testing.T) {
	token, _ := utils.GenerateToken(1, "test@test.com", domain.RoleClient)

	claims, err := utils.ValidateToken(token)
	if err != nil {
		t.Errorf("ValidateToken no debe retornar error: %v", err)
	}
	if claims.UserID != 1 {
		t.Errorf("UserID esperado 1, obtenido %d", claims.UserID)
	}
	if claims.Role != domain.RoleClient {
		t.Errorf("Role esperado %s, obtenido %s", domain.RoleClient, claims.Role)
	}
}

func TestValidateToken_Invalid(t *testing.T) {
	_, err := utils.ValidateToken("token.invalido.abc")
	if err == nil {
		t.Error("ValidateToken debe retornar error para token inválido")
	}
}

func TestGenerateQRCode(t *testing.T) {
	qr, err := utils.GenerateQRCode("TICKET-1-USER-1-EVENT-1")
	if err != nil {
		t.Errorf("GenerateQRCode no debe retornar error: %v", err)
	}
	if len(qr) == 0 {
		t.Error("el QR no debe estar vacío")
	}
}
