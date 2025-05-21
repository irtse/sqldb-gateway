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
	bodySchema int64, destID int64, destOnResponse int64, fileAttached string, signature string) (utils.Record, error) {
	var subj bytes.Buffer
	var content bytes.Buffer

	// SHOULD MAP AND APPLY CODE
	tmplSubj, err := template.New("email").Parse(subject)
	if err == nil {
		if err := tmplSubj.Execute(&subj, bodyToMap); err == nil {
			subject = subj.String()
		}
	}
	tmpl, err := template.New("email").Parse(tpl + "<br>" + signature)
	if err != nil {
		return utils.Record{}, err
	}

	if err := tmpl.Execute(&content, bodyToMap); err != nil {
		return utils.Record{}, err
	}
	m := utils.Record{
		"from_email":            utils.GetString(from, "id"),
		"to_email":              utils.GetString(to, "id"), // SHOULD BE ID
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
	if m["code"] == nil || m["code"] == "" {
		m["code"] = uuid.New()
	}
	return m, nil
}

func SendMail(from string, to string, mail utils.Record, isValidButton bool) error {
	var body bytes.Buffer
	boundary := "mixed-boundary"
	altboundary := "alt-boundary"
	// En-têtes MIME
	body.WriteString(fmt.Sprintf("From: %s\r\n", from))
	body.WriteString(fmt.Sprintf("To: %s\r\n", to))
	body.WriteString("Subject: " + utils.GetString(mail, "subject") + "\r\n")
	body.WriteString("MIME-Version: 1.0\r\n")
	body.WriteString("Content-Type: multipart/mixed; boundary=\"" + boundary + "\"\r\n")
	body.WriteString("\r\n--" + boundary + "\r\n")
	body.WriteString(fmt.Sprintf("Content-Type: multipart/alternative; boundary=\"%s\"\r\n", altboundary))
	body.WriteString("\r\n")
	// Partie texte
	body.WriteString(fmt.Sprintf("--%s\r\n", altboundary))
	body.WriteString("Content-Type: text/html; charset=\"utf-8\"\r\n")
	body.WriteString("Content-Transfer-Encoding: 7bit\r\n\r\n")
	body.WriteString("<html>")
	if isValidButton {
		body.WriteString(`
			<head>
				<meta charset="UTF-8">
			</head>
		`)
	}
	body.WriteString("<body style=\"margin:0; padding:0; font-family:Arial, sans-serif;\">")

	body.WriteString(utils.GetString(mail, "content"))

	code := utils.GetString(mail, "code")
	if isValidButton {
		host := os.Getenv("HOST")
		if host == "" {
			host = "http://capitalisation.irt-aese.local"
		}
		body.WriteString(fmt.Sprintf(`
			<div style="display:flex;justify-content:center;align-items: center;">
			<br>
				<table border="0" cellspacing="0" cellpadding="0" style="margin:0 10px 0 0">
					<tr>
						<td align="center" style="border-radius: 5px; background-color: #13aa52;">
							<a rel="noopener" target="_blank" rel="noopener" target="_blank" href="%s/v1/response/%s?got_response=true" target="_blank" style="font-size: 18px; font-family: Helvetica, Arial, sans-serif; color: #ffffff; font-weight: bold; text-decoration: none;border-radius: 5px; padding: 12px 18px; border: 1px solid #13aa52; display: inline-block;">✔</a>
						</td>
					</tr>
				</table>
				<table border="0" cellspacing="0" cellpadding="0">
					<tr>
						<td align="center" style="border-radius: 5px; background-color: #FF4742;">
							<a rel="noopener" target="_blank" rel="noopener" target="_blank" href="%s/v1/response/%s?got_response=false" target="_blank" style="font-size: 18px; font-family: Helvetica, Arial, sans-serif; color: #ffffff; font-weight: bold; text-decoration: none;border-radius: 5px; padding: 12px 18px; border: 1px solid #FF4742; display: inline-block;">✘</a>
						</td>
					</tr>
				</table>
				</div style="display:flex; ">
			<br>
		`, host, code, host, code))
	}
	body.WriteString("</body>")
	body.WriteString("</html>\n")
	body.WriteString("--" + altboundary + "--\n")
	body.WriteString("--" + boundary + "--\n")

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
		}
	}
	// Charger le template HTML
	var err error
	if pwd != "" {
		auth := smtp.PlainAuth("", from, pwd, smtpHost)
		err = smtp.SendMail(smtpHost+":"+smtpPort, auth, from,
			[]string{
				from,
				to,
			}, body.Bytes())
	} else {
		err = smtp.SendMail(smtpHost+":"+smtpPort, nil, from,
			[]string{
				from,
				to,
			}, body.Bytes())
	}
	if err != nil {
		fmt.Println("EMAIL NOT SEND", err)
		return err
	}
	fmt.Println("EMAIL SEND")
	return nil
}
