package utils

import (
	"fmt"
	"os"
	"strings"

	"github.com/resend/resend-go/v2"
)

func SendVerificationEmail(to, token string) error {
	apiKey := os.Getenv("RESEND_API_KEY")
	if apiKey == "" {
		return fmt.Errorf("RESEND_API_KEY no está configurada")
	}
	frontendURL := os.Getenv("FRONTEND_URL")
	if to == "" || token == "" {
		return fmt.Errorf("parámetros inválidos: email o token vacío")
	}
	verifyURL := fmt.Sprintf("%s/verify?token=%s", strings.TrimSuffix(frontendURL, "/"), token)

	client := resend.NewClient(apiKey)

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

	params := &resend.SendEmailRequest{
		From:    "Cotizador Productos EML <onboarding@resend.dev>",
		To:      []string{to},
		Subject: "Verifica tu cuenta",
		Html:    htmlContent,
		Text:    fmt.Sprintf("Verifica tu cuenta visitando este enlace: %s", verifyURL),
	}
	sent, err := client.Emails.Send(params)
	if err != nil {
		panic(err)
	}
	fmt.Println(sent.Id)
	return nil
}

func SendPasswordResetEmail(to, token string) error {
	apiKey := os.Getenv("RESEND_API_KEY")
	if apiKey == "" {
		return fmt.Errorf("RESEND_API_KEY no está configurada")
	}
	frontendURL := os.Getenv("FRONTEND_URL")
	if to == "" || token == "" {
		return fmt.Errorf("parámetros inválidos: email o token vacío")
	}
	resetURL := fmt.Sprintf("%s/reset-password?token=%s", strings.TrimSuffix(frontendURL, "/"), token)

	client := resend.NewClient(apiKey)

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

	params := &resend.SendEmailRequest{
		From:    "Cotizador Productos EML <onboarding@resend.dev>",
		To:      []string{to},
		Subject: "Restablece tu contraseña",
		Html:    htmlContent,
		Text:    fmt.Sprintf("Para restablecer tu contraseña, visita este enlace: %s", resetURL),
	}
	sent, err := client.Emails.Send(params)
	if err != nil {
		return fmt.Errorf("error enviando email de recuperación: %w", err)
	}
	fmt.Println(sent.Id)
	return nil
}
