package ddlparse

import (
	"github.com/kodaimura/ddlparse/internal/types"
	"github.com/kodaimura/ddlparse/internal/common"
	"github.com/kodaimura/ddlparse/internal/lexer"
	"github.com/kodaimura/ddlparse/internal/parser"
	"github.com/kodaimura/ddlparse/internal/validator"
)


type (
	Table = types.Table
	Column = types.Column
	DataType = types.DataType
	Constraint = types.Constraint
	Reference = types.Reference
	TableConstraint = types.TableConstraint
	PrimaryKey = types.PrimaryKey
	Unique = types.Unique
	Check = types.Check
	ForeignKey = types.ForeignKey
)

type (
	Rdbms = common.Rdbms
	ValidateError = common.ValidateError
)

const (
	PostgreSQL = common.PostgreSQL
	MySQL = common.MySQL
	SQLite = common.SQLite
)

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

func tokenize (ddl string, rdbms Rdbms) ([]string, error) {
	l := lexer.NewLexer(rdbms)
	return l.Lex(ddl)
}

func validate (tokens []string, rdbms Rdbms) ([]string, error) {
	v := validator.NewValidator(rdbms)
	return v.Validate(tokens)
}

func parse (tokens []string, rdbms Rdbms) []Table {
	p := parser.NewParser(rdbms, tokens)
	return p.Parse()
}
