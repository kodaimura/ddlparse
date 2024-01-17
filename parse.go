package ddlparse

import (
	"fmt"
	"errors"
	"strings"
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
	//Check func(interface{}) bool
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


func Parse(ddl string, rdbms Rdbms) ([]Table, error) {
	switch rdbms {
	case SQLite:
		return ParseSQLite(ddl)
	case PostgreSQL:
		return ParsePostgreSQL(ddl)
	case MySQL:
		return ParseMySQL(ddl)
	default:
		return []Table{}, errors.New("Not yet supported.")
	}
}


type parser interface {
	Validate() error
	Tokenize() error
	Parse() ([]Table, error)
}

func ParseSQLite(ddl string) ([]Table, error) {
	parser := newSQLiteParser(ddl)

	if err := parser.Validate(); err != nil {
		return []Table{}, err
	}
	return parser.Parse()
}

func ParsePostgreSQL(ddl string) ([]Table, error) {
	parser := newPostgreSQLParser(ddl)

	if err := parser.Validate(); err != nil {
		return []Table{}, err
	}
	return parser.Parse()
}

func ParseMySQL(ddl string) ([]Table, error) {
	parser := newMySQLParser(ddl)

	if err := parser.Validate(); err != nil {
		return []Table{}, err
	}
	return parser.Parse()
}


func Validate(ddl string, rdbms Rdbms) error {
	switch rdbms {
	case SQLite:
		return ValidateSQLite(ddl)
	case PostgreSQL:
		return ValidatePostgreSQL(ddl)
	case MySQL:
		return ValidateMySQL(ddl)
	default:
		return errors.New("Not yet supported.")
	}
}

func ValidateSQLite(ddl string) error {
	parser := newSQLiteParser(ddl)

	return parser.Validate()
}

func ValidatePostgreSQL(ddl string) error {
	parser := newPostgreSQLParser(ddl)

	return parser.Validate()
}

func ValidateMySQL(ddl string) error {
	parser := newMySQLParser(ddl)

	return parser.Validate()
}


func tokenize(ddl string) []string {
	ddl = strings.Replace(ddl, "(", " ( ", -1)
	ddl = strings.Replace(ddl, ")", " ) ", -1)
	ddl = strings.Replace(ddl, ";", " ; ", -1)
	ddl = strings.Replace(ddl, "\"", " \" ", -1)
	ddl = strings.Replace(ddl, "'", " ' ", -1)
	ddl = strings.Replace(ddl, "`", " ` ", -1)
	ddl = strings.Replace(ddl, ",", " , ", -1)
	ddl = strings.Replace(ddl, ".", " . ", -1)
	ddl = strings.Replace(ddl, "\n", " \n ", -1)
	ddl = strings.Replace(ddl, "\t", " ", -1)
	ddl = strings.Replace(ddl, "/*", " /* ", -1)
	ddl = strings.Replace(ddl, "*/", " */ ", -1)
	ddl = strings.Replace(ddl, "--", " -- ", -1)
	ddl = strings.Replace(ddl, "#", " # ", -1)

	return filter(
		strings.Split(ddl, " "), 
		func(s string) bool {return s != " " && s != ""},
	)
}