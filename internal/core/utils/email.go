package utils

import (
	"context"
	"os"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

type EmailSender interface {
	SendEmail(ctx context.Context, emailAddress string, subject string, body string) error
}

type EmailUtil struct{}

func (u *EmailUtil) SendEmail(ctx context.Context, emailAddress string, subject string, body string) error {
	from := mail.NewEmail("Dirtie Support", "dirtie.app@gmail.com")
	to := mail.NewEmail("", emailAddress)

	message := mail.NewSingleEmail(from, subject, to, body, body)
	client := sendgrid.NewSendClient(os.Getenv("SENDGRID_API_KEY"))
	_, err := client.SendWithContext(ctx, message)
	return err
}
