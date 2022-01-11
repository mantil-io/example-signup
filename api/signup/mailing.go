package signup

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ses"
	"github.com/aws/aws-sdk-go-v2/service/ses/types"
)

const (
	appNameEnv    = "APP_NAME"
	sourceMailEnv = "SOURCE_MAIL"
)

func (r *Signup) sendActivationCode(email, name, activationCode string) error {
	toEmail := email
	sourceMail, ok := os.LookupEnv(sourceMailEnv)
	if !ok {
		return fmt.Errorf("no source email address provided")
	}
	subject := "Activation"
	body, err := activationMailBody(name, activationCode)
	if err != nil {
		log.Printf("failed to get mail body: %s", err)
		return err
	}

	return r.sendEmail(sourceMail, toEmail, subject, "", body)
}

func activationMailBody(name, activationCode string) (string, error) {
	data := struct {
		Name           string
		ActivationCode string
		AppName        string
	}{name, activationCode, os.Getenv(appNameEnv)}

	return renderTemplate(data, activationMailBodyTemplate)
}

const activationMailBodyTemplate = `Hi {{.Name}},

Your activation code is: {{.ActivationCode}}.

The {{.AppName}} Team`

func (r Signup) sendWelcomeMail(email, name string) error {
	toEmail := email
	sourceMail, ok := os.LookupEnv(sourceMailEnv)
	if !ok {
		return fmt.Errorf("no source email address provided")
	}
	subject := "Welcome"
	body, err := welcomeMailBody(name)
	if err != nil {
		log.Printf("failed to get mail body: %s", err)
		return err
	}
	return r.sendEmail(sourceMail, toEmail, subject, "", body)
}

func welcomeMailBody(name string) (string, error) {
	data := struct {
		Name    string
		AppName string
	}{name, os.Getenv(appNameEnv)}
	return renderTemplate(data, welcomeMailBodyTemplate)
}

const welcomeMailBodyTemplate = `Hi {{.Name}},

Welcome to {{.AppName}}, we are excited to have you in our early beta program!

Have fun!
{{.AppName}} Team
`

func renderTemplate(data interface{}, content string) (string, error) {
	tpl, err := template.New("").Parse(content)
	if err != nil {
		return "", err
	}
	buf := bytes.NewBuffer(nil)
	if err := tpl.Execute(buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func (r *Signup) sendEmail(fromEmail, toEmail, subject, body, htmlBody string) error {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		log.Printf("failed to load configuration: %s", err)
		return err
	}
	cli := ses.NewFromConfig(cfg)

	smi := &ses.SendEmailInput{
		Message: &types.Message{
			Body: &types.Body{},
			Subject: &types.Content{
				Data: aws.String(subject),
			},
		},
		Destination: &types.Destination{
			ToAddresses: []string{toEmail},
		},
		Source: aws.String(fromEmail),
	}

	if htmlBody == "" {
		smi.Message.Body.Text = &types.Content{Data: aws.String(body)}
	} else {
		smi.Message.Body.Html = &types.Content{Data: aws.String(htmlBody)}
	}

	if _, err := cli.SendEmail(context.Background(), smi); err != nil {
		log.Printf("send email error: %s", err)
		return err
	}
	return nil
}
