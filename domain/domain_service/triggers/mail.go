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
	var content bytes.Buffer

	code := uuid.New().String()
	bodyToMap["code"] = code
	// SHOULD MAP AND APPLY CODE
	tmpl, err := template.New("email").Parse(tpl + "<br>" + signature)
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

func SendMail(from string, to string, mail utils.Record, id string, isValidButton bool) error {
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
	body.WriteString("Content-Type: text/html; charset=\"utf-8\"\r\n")
	body.WriteString("Content-Transfer-Encoding: 7bit\r\n\r\n")
	body.WriteString("<html>")
	body.WriteString("<body>")
	if isValidButton {
		body.Write([]byte(`
			<head>
				<meta charset="UTF-8">
				<style>
					.buttons {
					display: flex;
					width: 380px;
					gap: 10px;
					--b: 5px;   /* the border thickness */
					--h: 1.8em; /* the height */
					}

					.buttons button {
					--_c: #88C100;
					flex: calc(1.25 + var(--_s,0));
					min-width: 0;
					font-size: 40px;
					font-weight: bold;
					height: var(--h);
					cursor: pointer;
					color: var(--_c);
					border: var(--b) solid var(--_c);
					background: 
						conic-gradient(at calc(100% - 1.3*var(--b)) 0,var(--_c) 209deg, #0000 211deg) 
						border-box;
					clip-path: polygon(0 0,100% 0,calc(100% - 0.577*var(--h)) 100%,0 100%);
					padding: 0 calc(0.288*var(--h)) 0 0;
					margin: 0 calc(-0.288*var(--h)) 0 0;
					box-sizing: border-box;
					transition: flex .4s;
					}

					.buttons button + button {
					--_c: #FF003C;
					flex: calc(.75 + var(--_s,0));
					background: 
						conic-gradient(from -90deg at calc(1.3*var(--b)) 100%,var(--_c) 119deg, #0000 121deg) 
						border-box;
					clip-path: polygon(calc(0.577*var(--h)) 0,100% 0,100% 100%,0 100%);
					margin: 0 0 0 calc(-0.288*var(--h));
					padding: 0 0 0 calc(0.288*var(--h));
					}

					.buttons button:focus-visible {
					outline-offset: calc(-2*var(--b));
					outline: calc(var(--b)/2) solid #000;
					background: none;
					clip-path: none;
					margin: 0;
					padding: 0;
					}

					.buttons button:focus-visible + button {
					background: none;
					clip-path: none;
					margin: 0;
					padding: 0;
					}

					.buttons button:has(+ button:focus-visible) {
					background: none;
					clip-path: none;
					margin: 0;
					padding: 0;
					}

					button:hover,
					button:active:not(:focus-visible) {
					--_s: .75;
					}

					button:active {
					box-shadow: inset 0 0 0 100vmax var(--_c);
					color: #fff;
					}

					body {
					display: grid;
					place-content: center;
					margin: 0;
					height: 100vh;
					font-family: system-ui, sans-serif;
					}
				</style>
				</head>
			`))
	}
	body.Write([]byte(utils.GetString(mail, "content")))
	body.WriteString("</html>")
	body.WriteString("</body>")

	if isValidButton {
		host := os.Getenv("HOST")
		if host == "" {
			host = "http://capitalisation.irt-aese.local"
		}
		body.Write([]byte(fmt.Sprintf(`
			<br>
			<br>
			<div class="buttons">
				<form action="%s/v1/generic/dbemail_response?rows=all" method="POST">
      				<input type="hidden" name="%s" value="%s">
					<input type="hidden" name="got_response" value="true">
					<button type="submit">VALID</button>
				</form>
				<form action="%s/v1/generic/dbemail_response?rows=all" method="POST">
					<input type="hidden" name="action" value="confirm">
					<input type="hidden" name="%s" value="%s">
					<input type="hidden" name="got_response" value="false">
					<button type="submit">REFUSED</button>
				</form>
			</div>
			<br>
			<br>
		`, host, ds.EmailSendedDBField, id, host, ds.EmailSendedDBField, id)))
	}

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
		fmt.Println("EMAIL NOT SEND", err)
		return err
	}
	fmt.Println("EMAIL SEND")
	return nil
}
