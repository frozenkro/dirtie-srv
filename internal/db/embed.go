package db

import _ "embed"

//go:embed sqlc/schema.sql
var SchemaSql []byte
