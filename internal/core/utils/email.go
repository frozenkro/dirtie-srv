package utils

import (
	"context"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

type EmailSender interface {
  SendEmail(ctx context.Context, emailAddress string, subject string, body []byte) error
}

type EmailUtil struct {}

func (u *EmailUtil) SendEmail(ctx context.Context, emailAddress string, subject string, body []byte) error {
  from := mail.NewEmail("Dirtie Support", "dirtie.app@gmail.com")
  to := mail.NewEmail("", emailAddress)

  return nil
}
