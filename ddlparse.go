package ddlparse

import (
	"fmt"
	"errors"
)

type Rdbms string

const (
	SQLite Rdbms = "SQLite"
	PostgreSQL Rdbms = "PostgreSQL"
	MySQL Rdbms = "MySQL"
)

type Table struct {
	Schema string
	Name string
	Columns []Column
}

type Column struct {
	Name string
	DataType string
	IsPK bool
	IsNotNull bool
	IsUnique bool
	IsAutoIncrement bool
	Default interface{}
}

func Parse(ddl string, rdbms Rdbms) ([]Table, error) {
	tokens, err := tokenize(ddl, rdbms)
	if err != nil {
		return []Table{}, err
	}
	validatedTokens, err := validate(ddl, rdbms)
	if err != nil {
		return []Table{}, err
	}
	tables, err := validate(ddl, rdbms)
	if err != nil {
		return []Table{}, err
	}
	return tables, nil
}

func ParseSQLite(ddl string) ([]Table, error) {
	return Parse(ddl, SQLite)
}

func ParsePostgreSQL(ddl string) ([]Table, error) {
	return Parse(ddl, PostgreSQL)
}

func ParseMySQL(ddl string) ([]Table, error) {
	return Parse(ddl, MySQL)
}

func ParseForce(ddl string) ([]Table, error) {
	ls := []Rdbms{SQLite, PostgreSQL, MySQL}
	var err error
	for _, rdbms := range ls {
		tables, err := Parse(ddl, rdbms)
		if err == nil {
			return tables, nil
		}
	}
	return []Table{}, err
}

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

type ParseError struct {
	message string
}

func NewParseError(message string) error {
	return ParseError{message}
}

func (e ParseError) Error() string {
	return fmt.Sprintf("ParseError: %s", e.message)
}