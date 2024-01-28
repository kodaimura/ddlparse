package ddlparse

import (
	"fmt"
)

type Rdbms string

const (
	SQLite Rdbms = "SQLite"
	PostgreSQL Rdbms = "PostgreSQL"
	MySQL Rdbms = "MySQL"
)

type Table struct {
	Schema string `json:"schema"`
	Name string `json:"name"`
	IfNotExists bool `json:"if_not_exists"`
	Columns []Column `json:"columns"`
	Constraints TableConstraint `json:"constraints"`
}

type Column struct {
	Name string `json:"name"`
	DataType DataType `json:"data_type"`
	Constraint Constraint `json:"constraint"`
}

type DataType struct {
	Name string `json:"name"`
	DigitN int `json:"digit_n"`
	DigitM int `json:"digit_m"`
}

type Constraint struct {
	Name string `json:"name"`
	IsPrimaryKey bool `json:"is_primary_key"`
	IsUnique bool `json:"is_unique"`
	IsNotNull bool `json:"is_not_null"`
	IsAutoincrement bool `json:"is_autoincrement"`
	Default interface{} `json:"default"`
	Check string `json:"check"`
	Collate string `json:"collate"`
	References Reference `json:"references"`
}

type Reference struct {
	TableName string `json:"table_name"`
	ColumnNames []string `json:"column_names"`
}

type TableConstraint struct {
	PrimaryKey []PrimaryKey `json:"primary_key"`
	Unique []Unique `json:"unique"`
	Check []Check `json:"check"`
	ForeignKey []ForeignKey `json:"foreign_key"`
}

type PrimaryKey struct {
	Name string `json:"name"`
	ColumnNames []string `json:"column_names"`
}

type Unique struct {
	Name string `json:"name"`
	ColumnNames []string `json:"column_names"`
}

type Check struct {
	Name string `json:"name"`
	Expr string `json:"expr"`
}

type ForeignKey struct {
	Name string `json:"name"`
	ColumnNames []string `json:"column_names"`
	References Reference `json:"references"`
}


func Parse(ddl string, rdbms Rdbms) ([]Table, error) {
	tokens, err := tokenize(ddl, rdbms)
	if err != nil {
		return []Table{}, err
	}
	validatedTokens, err := validate(tokens, rdbms)
	if err != nil {
		return []Table{}, err
	}
	tables:= parse(validatedTokens, rdbms)
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