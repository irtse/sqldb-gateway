package triggers

import (
	"bytes"
	"fmt"
	"net/smtp"
	"os"
	"strings"
	"text/template"
)

type EmailData struct {
	Name string
	Code string
}

func SendMail(from string, to string, subject string, tpl string, bodyToMap map[string]interface{}) error {
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")
	pwd := os.Getenv("SMTP_PASSWORD")
	// Charger le template HTML
	tmpl, err := template.New("email").Parse(tpl)
	if err != nil {
		return err
	}
	var body bytes.Buffer
	// En-têtes MIME
	body.WriteString("MIME-Version: 1.0\r\n")
	body.WriteString("Content-Type: text/html; charset=\"UTF-8\"\r\n")
	body.WriteString(fmt.Sprintf("To: "+to+"%s\r\n", to))
	body.WriteString("Subject: " + subject + "\r\n")
	body.WriteString("\r\n")

	// Appliquer le template
	if err := tmpl.Execute(&body, bodyToMap); err != nil {
		return err
	}
	// Auth SMTP
	auth := smtp.PlainAuth("", from, pwd, smtpHost)
	// Envoi
	err = smtp.SendMail(smtpHost+":"+smtpPort, auth, from, strings.Split(to, ","), body.Bytes())
	if err != nil {
		return err
	}
	return nil
}
