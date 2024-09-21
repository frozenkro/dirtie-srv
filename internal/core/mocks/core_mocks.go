package mocks

import (
  "context"
  "html/template"

  "github.com/stretchr/testify/mock"
)

type MockHtmlParser struct {
  mock.Mock
}

type MockEmailSender struct {
  mock.Mock
}

func (m *MockHtmlParser) ReadFile(ctx context.Context, path string) (*template.Template, error) {
  args := m.Called(ctx, path)
  return args.Get(0).(*template.Template), args.Error(1)
}

func (m *MockHtmlParser) ReplaceVars(ctx context.Context, vars any, tmp *template.Template) ([]byte, error) {
  args := m.Called(ctx, vars, tmp)
  return args.Get(0).([]byte), args.Error(1)
}

func (m *MockEmailSender) SendEmail(ctx context.Context, emailAddress string, subject string, body string) error {
  args := m.Called(ctx, emailAddress, subject, body)
  return args.Error(0)
}
