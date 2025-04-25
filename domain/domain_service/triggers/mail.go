package triggers

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"net/smtp"
	"os"
	ds "sqldb-ws/domain/schema/database_resources"
	"sqldb-ws/domain/utils"
	"strings"
	"text/template"

	"github.com/google/uuid"
)

type EmailData struct {
	Name string
	Code string
}

func ForgeMail(from utils.Record, to utils.Record, subject string, tpl string,
	bodyToMap map[string]interface{}, domain utils.DomainITF, tplID int64,
	bodySchema int64, destID int64, destOnResponse int64, fileAttached string) (utils.Record, error) {
	var content bytes.Buffer

	code := uuid.New().String()
	bodyToMap["code"] = code
	// SHOULD MAP AND APPLY CODE
	tmpl, err := template.New("email").Parse(tpl)
	if err != nil {
		return utils.Record{}, err
	}
	if err := tmpl.Execute(&content, bodyToMap); err != nil {
		return utils.Record{}, err
	}
	m := utils.Record{
		"from_email":            utils.GetString(from, "email"),
		"to_email":              utils.GetString(to, "email"),
		"subject":               subject,
		"content":               content.String(),
		"file_attached":         "",
		ds.EmailTemplateDBField: tplID,
	}
	if destOnResponse > -1 {
		m[ds.DestTableDBField+"_on_response"] = destOnResponse
	}
	if bodySchema > -1 {
		m["mapped_with"+ds.SchemaDBField] = bodySchema
	}
	if destID > -1 {
		m["mapped_with"+ds.DestTableDBField] = destID
	}
	m["code"] = code
	return m, nil
}

func SendMail(from string, to string, mail utils.Record) error {
	var body bytes.Buffer
	boundary := "MY-MIME-BOUNDARY"
	// En-têtes MIME
	body.WriteString(fmt.Sprintf("From: %s\r\n", from))
	body.WriteString(fmt.Sprintf("To: %s\r\n", to))
	body.WriteString("Subject: " + utils.GetString(mail, "subject") + "\r\n")
	body.WriteString("MIME-Version: 1.0\r\n")
	body.WriteString("Content-Type: multipart/mixed; boundary=" + boundary + "\r\n")
	body.WriteString("\r\n--" + boundary + "\r\n")

	// Partie texte
	body.WriteString("Content-Type: text/plain; charset=\"utf-8\"\r\n")
	body.WriteString("Content-Transfer-Encoding: 7bit\r\n\r\n")
	body.Write([]byte(utils.GetString(mail, "content")))
	body.WriteString("\r\n--" + boundary + "\r\n")

	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")
	pwd := os.Getenv("SMTP_PASSWORD")

	if file_attached := utils.GetString(mail, "file_attached"); file_attached != "" {
		splitted := strings.Split(file_attached, "/")
		fileName := splitted[len(splitted)-1]

		fileData, err := os.ReadFile(file_attached)
		if err == nil {
			fileBase64 := base64.StdEncoding.EncodeToString(fileData)
			body.WriteString("Content-Type: application/octet-stream\r\n")
			body.WriteString("Content-Transfer-Encoding: base64\r\n")
			body.WriteString("Content-Disposition: attachment; filename=\"" + fileName + "\"\r\n\r\n")
			// Diviser le base64 en lignes de 76 caractères (RFC)
			for i := 0; i < len(fileBase64); i += 76 {
				end := i + 76
				if end > len(fileBase64) {
					end = len(fileBase64)
				}
				body.WriteString(fileBase64[i:end] + "\r\n")
			}

			body.WriteString("--" + boundary + "--")
		}
	}
	// Charger le template HTML
	auth := smtp.PlainAuth("", from, pwd, smtpHost)

	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, from,
		[]string{
			from,
			to,
		}, body.Bytes())
	if err != nil {
		return err
	}
	fmt.Println("EMAIL SEND")
	return nil
}
