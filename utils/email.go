package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
)

type EmailRequest struct {
	Sender      map[string]string   `json:"sender"`
	To          []map[string]string `json:"to"`
	Subject     string              `json:"subject"`
	HTMLContent string              `json:"htmlContent"`
}

// sendEmail es una función genérica para enviar correos con Brevo
func sendEmail(to, subject, htmlContent string) error {
	apiKey := os.Getenv("BREVO_API_KEY")
	if apiKey == "" {
		return fmt.Errorf("BREVO_API_KEY no está configurada")
	}

	email := EmailRequest{
		Sender: map[string]string{
			"name":  "Cotizador Productos EML",
			"email": "pipe12.fm@gmail.com", // Usa un correo verificado en Brevo
		},
		To: []map[string]string{
			{"email": to, "name": "Usuario"},
		},
		Subject:     subject,
		HTMLContent: htmlContent,
	}

	data, err := json.Marshal(email)
	if err != nil {
		return fmt.Errorf("error al serializar el correo: %w", err)
	}

	req, err := http.NewRequest("POST", "https://api.brevo.com/v3/smtp/email", bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("error creando la solicitud: %w", err)
	}
	req.Header.Set("api-key", apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error enviando email: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("error en la respuesta de Brevo: %s", resp.Status)
	}

	fmt.Println("Email enviado con éxito:", resp.Status)
	return nil
}

// SendVerificationEmail envía un email de verificación de cuenta
func SendVerificationEmail(to, token string) error {
	frontendURL := os.Getenv("FRONTEND_URL")
	if to == "" || token == "" {
		return fmt.Errorf("parámetros inválidos: email o token vacío")
	}
	verifyURL := fmt.Sprintf("%s/confirmar-cuenta?token=%s", strings.TrimSuffix(frontendURL, "/"), token)

	htmlContent := fmt.Sprintf(`
	<!DOCTYPE html>
	<html>
	<body>
		<h1>Verifica tu cuenta</h1>
		<p>Haz clic en el siguiente enlace para completar la verificación:</p>
		<a href="%s" style="background-color: #007bff; color: white; padding: 10px 20px; text-decoration: none; border-radius: 5px;">
			Verificar cuenta
		</a>
		<p>Si no solicitaste este registro, ignora este mensaje.</p>
	</body>
	</html>
	`, verifyURL)

	return sendEmail(to, "Verifica tu cuenta", htmlContent)
}

// SendPasswordResetEmail envía un email para restablecer la contraseña
func SendPasswordResetEmail(to, token string) error {
	frontendURL := os.Getenv("FRONTEND_URL")
	if to == "" || token == "" {
		return fmt.Errorf("parámetros inválidos: email o token vacío")
	}
	resetURL := fmt.Sprintf("%s/reset-password?token=%s", strings.TrimSuffix(frontendURL, "/"), token)

	htmlContent := fmt.Sprintf(`
	<!DOCTYPE html>
	<html>
	<body>
		<h1>Restablece tu contraseña</h1>
		<p>Haz clic en el siguiente enlace para restablecer tu contraseña:</p>
		<a href="%s" style="background-color: #dc3545; color: white; padding: 10px 20px; text-decoration: none; border-radius: 5px;">
			Restablecer contraseña
		</a>
		<p>Si no solicitaste este cambio, ignora este mensaje.</p>
	</body>
	</html>
	`, resetURL)

	return sendEmail(to, "Restablece tu contraseña", htmlContent)
}
