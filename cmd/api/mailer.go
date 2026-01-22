package main

import (
	"bytes"
	"crypto/tls"
	"embed"
	"fmt"
	"html/template"
	"time"

	mail "github.com/xhit/go-simple-mail/v2"
)

//go:embed templates
var emailTemplateFS embed.FS

func (app *Application) SendEmail(from string, to string, subject string, tmpl string, data any) error {

	templateToRender := fmt.Sprintf("templates/%s.html.tmpl", tmpl)

	t, err := template.New("email-html").ParseFS(emailTemplateFS, templateToRender)
	if err != nil {
		app.ErrorLog.Println(err)
		return err
	}

	var tpl bytes.Buffer
	if err = t.ExecuteTemplate(&tpl, "body", data); err != nil {
		app.ErrorLog.Println(err)
		return err
	}

	formattedMsg := tpl.String()

	templateToRender = fmt.Sprintf("templates/%s.plain.tmpl", tmpl)
	t, err = template.New("email-plain").ParseFS(emailTemplateFS, templateToRender)
	if err = t.ExecuteTemplate(&tpl, "body", data); err != nil {
		app.ErrorLog.Println(err)
		return err
	}

	if err = t.ExecuteTemplate(&tpl, "body", data); err != nil {
		app.ErrorLog.Println(err)
		return err
	}

	plainMsg := tpl.String()

	// send email
	smtpHost := app.Config.SMTP.Host
	smtpPort := app.Config.SMTP.Port
	mailtrapUserName := app.Config.SMTP.Username
	mailtrapPassword := app.Config.SMTP.Password

	server := mail.NewSMTPClient()

	server.Host = smtpHost
	server.Port = smtpPort
	server.Username = mailtrapUserName
	server.Password = mailtrapPassword
	server.Encryption = mail.EncryptionSTARTTLS

	// Variable to keep alive connection
	server.KeepAlive = false

	// Timeout for connect to SMTP Server
	server.ConnectTimeout = 10 * time.Second

	// Timeout for send the data and wait respond
	server.SendTimeout = 10 * time.Second

	// Set TLSConfig to provide custom TLS configuration. For example,
	// to skip TLS verification (useful for testing):
	server.TLSConfig = &tls.Config{InsecureSkipVerify: true}

	// SMTP client
	smtpClient, err := server.Connect()

	if err != nil {
		app.ErrorLog.Println(err)
		return err
	}

	email := mail.NewMSG()
	email.SetFrom(from).AddTo(to).SetSubject(subject)
	email.SetBody(mail.TextHTML, formattedMsg)
	email.AddAlternative(mail.TextPlain, plainMsg)

	err = email.Send(smtpClient)
	if err != nil {
		app.ErrorLog.Println(err)
		return err
	}

	app.InfoLog.Println("Email sent!")

	return nil
}
