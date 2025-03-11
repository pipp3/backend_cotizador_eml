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
