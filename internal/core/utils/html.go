package utils

import (
	"context"
	"html/template"
)

type HtmlParser interface {
  ReadFile(ctx context.Context, path string) (*template.Template, error)
  ReplaceVars(ctx context.Context, data any, tmp *template.Template) ([]byte, error)
}

type HtmlUtil struct {}

func (u *HtmlUtil) ReadFile(ctx context.Context, path string) (*template.Template, error) {
  // TODO
  return &template.Template{}, nil
}

func (u *HtmlUtil) ReplaceVars(ctx context.Context, data any, tmp *template.Template) ([]byte, error) {
  // TODO
  return nil, nil
}
