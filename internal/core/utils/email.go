package utils

import (
	"context"

	"github.com/frozenkro/dirtie-srv/internal/core"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

type EmailUtil struct{}

func (u EmailUtil) SendEmail(ctx context.Context, emailAddress string, subject string, body string) error {
	from := mail.NewEmail("Dirtie Support", "dirtie.app@gmail.com")
	to := mail.NewEmail("", emailAddress)

	message := mail.NewSingleEmail(from, subject, to, body, body)
	client := sendgrid.NewSendClient(core.SENDGRID_API_KEY)
	_, err := client.SendWithContext(ctx, message)
	return err
}
