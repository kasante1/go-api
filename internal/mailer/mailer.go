package mailer

import (
	"bytes"
	"embed"
	"html/template"
	"log"

	"github.com/wneessen/go-mail"
)

//go:embed "templates"
var templateFS embed.FS

type Mailer struct {
	client *mail.Client
	sender string
}

func New(host string, port int, username, password, sender string) Mailer {

	c, err := mail.NewClient(host, mail.WithPort(port), mail.WithSMTPAuth(mail.SMTPAuthPlain),
		mail.WithUsername(username), mail.WithPassword(password))

	if err != nil {
		log.Fatal(err)
	}

	return Mailer{
		client: c,
		sender: sender,
	}
}

func (m Mailer) Send(recipient, templateFile string, data interface{}) error {

	tmpl, err := template.New("email").ParseFS(templateFS, "templates/"+templateFile)
	if err != nil {
		return err
	}

	subject := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(subject, "subject", data)
	if err != nil {
		return err
	}

	plainBody := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(plainBody, "plainBody", data)
	if err != nil {
		return err
	}

	htmlBody := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(htmlBody, "htmlBody", data)
	if err != nil {
		return err
	}

	msg := mail.NewMsg()
	msg.To(recipient)
	msg.From(m.sender)
	msg.SetGenHeader("Subject", subject.String())
	msg.SetBodyString("text/plain", plainBody.String())
	msg.AddAlternativeHTMLTemplate(tmpl, htmlBody.String())

	err = m.client.DialAndSend(msg)
	if err != nil {
		return err
	}
	return nil
}
