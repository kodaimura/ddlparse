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
	result []string
}


func (v *validator) init(tokens []string) {
	v.tokens = tokens
	v.size = len(v.tokens)
	v.i = 0
	v.line = 1
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


func (v *validator) set(token string) {
	v.result = append(v.result, token)
}


func (v *validator) isOutOfRange() bool {
	return v.i > v.size - 1
}


func (v *validator) next() string {
	if (v.isOutOfRange()) {
		return common.EOF
	}
	token := v.token()
	for true {
		v.i += 1
		if v.isOutOfRange() {
			break
		} else if (v.token() == "\n") {
			v.line += 1
			continue
		} else {
			break
		}
	}
	return token
}


func (v *validator) syntaxError() error {
	return common.NewValidateError(v.line, v.token())
}


func (v *validator) matchToken(keywords ...string) bool {
	return common.Contains(
		append(
			common.MapSlice(keywords, strings.ToLower), 
			common.MapSlice(keywords, strings.ToUpper)...,
		), v.token())
}


func (v *validator) matchTokenNext(set bool, keywords ...string) bool {
	if v.matchToken(keywords...) {
		if set {
			v.set(v.next())
		} else {
			v.next()
		}
		return true
	}
	return false
}


func (v *validator) validateToken(set bool, keywords ...string) error {
	if (v.isOutOfRange()) {
		return v.syntaxError()
	}
	if v.matchToken(keywords...) {
		if set {
			v.set(v.next())
		} else {
			v.next()
		}
		return nil
	}
	return v.syntaxError()
}


func (v *validator) validateBrackets(set bool) error {
	if err := v.validateToken(set, "("); err != nil {
		return err
	}
	if err := v.validateBracketsAux(set); err != nil {
		return err
	}
	if err := v.validateToken(set, ")"); err != nil {
		return err
	}
	return nil
}


func (v *validator) validateBracketsAux(set bool) error {
	if v.matchToken(")") {
		return nil
	}
	if v.matchToken("(") {
		if err := v.validateBrackets(set); err != nil {
			return err
		}
		return v.validateBracketsAux(set)
	}
	if set {
		v.set(v.next())
	} else {
		v.next()
	}
	return v.validateBracketsAux(set)
}