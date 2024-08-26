package validator

import (
	"errors"
	"strings"

	"github.com/kodaimura/ddlparse/internal/common"
)

type Validator interface {
	Validate() ([]string, error)
}

func NewValidator(rdbms common.Rdbms, tokens []string) Validator {
	if rdbms == common.PostgreSQL {
		return NewPostgreSQLValidator(tokens)
	} else if rdbms == common.MySQL {
		return NewMySQLValidator(tokens)
	}
	return NewSQLiteValidator(tokens)
}

type validator struct {
	tokens []string
	validatedTokens []string
	size int
	i int
	line int
	flg bool
}


func (v *validator) init() {
	v.validatedTokens = []string{}
	v.i = -1
	v.line = 1
	v.size = len(v.tokens)
	v.flg = false
	v.next()
}


func (v *validator) token() string {
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


func (v *validator) next() error {
	if v.flg {
		v.validatedTokens = append(v.validatedTokens, v.token())
	}
	return v.nextAux()
}


func (v *validator) nextAux() error {
	v.i += 1
	if (v.isOutOfRange()) {
		return errors.New("out of range")
	}
	if (v.token() == "\n") {
		v.line += 1
		return v.nextAux()
	} else {
		return nil
	}
}


func (v *validator) syntaxError() error {
	if v.isOutOfRange() {
		return common.NewValidateError(v.line, v.tokens[v.size - 1])
	}
	return common.NewValidateError(v.line, v.tokens[v.i])
}


func (v *validator) matchKeyword(keywords ...string) bool {
	return common.Contains(
		append(
			common.MapSlice(keywords, strings.ToLower), 
			common.MapSlice(keywords, strings.ToUpper)...,
		), v.token())
}


func (v *validator) matchSymbol(symbols ...string) bool {
	return common.Contains(symbols, v.token())
}


func (v *validator) validateKeyword(keywords ...string) error {
	if (v.isOutOfRange()) {
		return v.syntaxError()
	}
	if v.matchKeyword(keywords...) {
		v.next()
		return nil
	}
	return v.syntaxError()
}


func (v *validator) validateSymbol(symbols ...string) error {
	if (v.isOutOfRange()) {
		return v.syntaxError()
	}
	if v.matchSymbol(symbols...) {
		v.next()
		return nil
	}
	return v.syntaxError()
}