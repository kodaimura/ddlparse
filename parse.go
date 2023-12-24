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
	Check func(interface{}) bool
}

type ValidateError struct {
	Line int
	Expected string
	Found string
}

func NewValidateError(line int, expected string, found string) error {
	return ValidateError{line, expected, found}
}

func (e ValidateError) Error() string {
	msg := fmt.Sprintf("ValidateError: (line:%d) ", e.Line)
	if (e.Expected == "" && e.Found != "") {
		msg += fmt.Sprintf("Unexpected characters '%s' found.", e.Found)

	} else if (e.Expected != "" && e.Found == "") {
		msg += fmt.Sprintf("Eexpected '%s', but not found.", e.Expected)

	} else if (e.Expected != "" && e.Found == "") {
		msg += fmt.Sprintf("Eexpected '%s', but found '%s'.", e.Expected, e.Found)

	} else {
		msg += "validate failed."
	}
	return fmt.Sprintf(msg)
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
	Parse() ([]Table, error)
}

func ParseSQLite(ddl string) ([]Table, error) {
	tokens := tokenize(ddl)
	parser := newSQLiteParser(tokens)

	if err := parser.Validate(); err != nil {
		return []Table{}, err
	}
	return parser.Parse()
}

func ParsePostgreSQL(ddl string) ([]Table, error) {
	tokens := tokenize(ddl)
	parser := newPostgreSQLParser(tokens)

	if err := parser.Validate(); err != nil {
		return []Table{}, err
	}
	return parser.Parse()
}

func ParseMySQL(ddl string) ([]Table, error) {
	tokens := tokenize(ddl)
	parser := newMySQLParser(tokens)

	if err := parser.Validate(); err != nil {
		return []Table{}, err
	}
	return parser.Parse()
}

func tokenize(ddl string) []string {
	ddl = strings.Replace(ddl, "(", " ( ", -1)
	ddl = strings.Replace(ddl, ")", " ) ", -1)
	ddl = strings.Replace(ddl, ";", " ; ", -1)
	ddl = strings.Replace(ddl, "\"", " \" ", -1)
	ddl = strings.Replace(ddl, "'", " ' ", -1)
	ddl = strings.Replace(ddl, "`", " ` ", -1)
	ddl = strings.Replace(ddl, ",", " , ", -1)
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

func filter(array []string, f func(string) bool) []string {
	var ret []string
	for _, s := range array {
		if f(s) {
			ret = append(ret, s)
		}
	}
	return ret
}