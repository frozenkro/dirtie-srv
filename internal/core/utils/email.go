package utils

import "context"

type EmailSender interface {
  SendEmail(ctx context.Context, emailAddress string, subject string, body []byte) error
}

type EmailUtil struct {}

func (u *EmailUtil) SendEmail(ctx context.Context, emailAddress string, subject string, body []byte) error {
  //TODO 
  return nil
}
