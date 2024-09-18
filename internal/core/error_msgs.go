package core

import "fmt"

const (
	RequestParseError = "Error parsing request body"
)

func GetMissingParamError(paramName string) string {
	return fmt.Sprintf("Missing parameter: %v", paramName)
}
