package utils

import (
  "bytes"
	"context"
	"html/template"
)

type HtmlParser interface {
  ReadFile(ctx context.Context, path string) (*template.Template, error)
  ReplaceVars(ctx context.Context, data any, tmp *template.Template) ([]byte, error)
}

type HtmlUtil struct {}

func (u *HtmlUtil) ReadFile(ctx context.Context, path string) (*template.Template, error) {
  return template.ParseFiles(path)
}

func (u *HtmlUtil) ReplaceVars(ctx context.Context, data any, tmp *template.Template) ([]byte, error) {
  var buf bytes.Buffer
  err := tmp.Execute(&buf, data)
  return buf.Bytes(), err
}
