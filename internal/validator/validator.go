package validator

import (
	"strings"

	"github.com/kodaimura/ddlparse/internal/common"
)

type Validator interface {
	Validate(tokens []string) ([]string, error)
}

func NewValidator(rdbms common.Rdbms) Validator {
	if rdbms == common.PostgreSQL {
		return NewPostgreSQLValidator()
	} else if rdbms == common.MySQL {
		return NewMySQLValidator()
	}
	return NewSQLiteValidator()
}

type validator struct {
	tokens []string
	size int
	i int
	line int
	flg bool
	result []string
}


func (v *validator) init(tokens []string) {
	v.tokens = tokens
	v.size = len(v.tokens)
	v.i = 0
	v.line = 1
	v.flg = false
	v.result = []string{}
	if v.token() == "\n" {
		v.next()
	}
}


func (v *validator) token() string {
	if v.isOutOfRange() {
		return common.EOF
	}
	return v.tokens[v.i]
}


func (v *validator) flgOn() {
	v.flg = true
}


func (v *validator) flgOff() {
	v.flg = false
}


func (v *validator) isOutOfRange() bool {
	return v.i > v.size - 1
}


func (v *validator) next() string {
	if v.flg {
		v.result = append(v.result, v.token())
	}
	return v.nextAux()
}


func (v *validator) nextAux() string {
	if (v.isOutOfRange()) {
		return common.EOF
	}
	token := v.token()
	v.i += 1
	if (v.token() == "\n") {
		v.line += 1
		return v.nextAux()
	} else {
		return token
	}
}


func (v *validator) syntaxError() error {
	if v.isOutOfRange() {
		return common.NewValidateError(v.line, common.EOF)
	}
	return common.NewValidateError(v.line, v.tokens[v.i])
}


func (v *validator) matchToken(keywords ...string) bool {
	return common.Contains(
		append(
			common.MapSlice(keywords, strings.ToLower), 
			common.MapSlice(keywords, strings.ToUpper)...,
		), v.token())
}


func (v *validator) validateToken(keywords ...string) error {
	if (v.isOutOfRange()) {
		return v.syntaxError()
	}
	if v.matchToken(keywords...) {
		v.next()
		return nil
	}
	return v.syntaxError()
}


func (v *validator) validateBrackets() error {
	if err := v.validateToken("("); err != nil {
		return err
	}
	if err := v.validateBracketsAux(); err != nil {
		return err
	}
	if err := v.validateToken(")"); err != nil {
		return err
	}
	return nil
}


func (v *validator) validateBracketsAux() error {
	if v.matchToken(")") {
		return nil
	}
	if v.matchToken("(") {
		if err := v.validateBrackets(); err != nil {
			return err
		}
		return v.validateBracketsAux()
	}
	v.next()
	return v.validateBracketsAux()
}