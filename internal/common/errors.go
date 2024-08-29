package common

import (
	"fmt"
)


type ValidateError struct {
	Line int
	Near string
}

func NewValidateError(line int, near string) error {
	return ValidateError{line, near}
}

func (e ValidateError) Error() string {
	return fmt.Sprintf("ValidateError: Syntax error: near '%s' at line %d.", e.Near, e.Line)
}