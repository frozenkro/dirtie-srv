package utils

import (
	"bytes"
	"context"
	"html/template"
	"net/http"

	"github.com/frozenkro/dirtie-srv/assets"
)

type HtmlUtil struct{}

func (u HtmlUtil) ReadFile(ctx context.Context, path string) (*template.Template, error) {
	return template.ParseFS(assets.AssetDir, path)
}

func (u HtmlUtil) ReplaceVars(ctx context.Context, data any, tmp *template.Template) ([]byte, error) {
	var buf bytes.Buffer
	err := tmp.Execute(&buf, data)
	return buf.Bytes(), err
}

func (u HtmlUtil) ReplaceAndWrite(ctx context.Context, data any, tmp *template.Template, w http.ResponseWriter) error {
	return tmp.Execute(w, data)
}
